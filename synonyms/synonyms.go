package synonyms

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	//"github.com/davecgh/go-spew/spew"
	"gopkg.in/redis.v3"
)

// Public functions

// GetSynonyms from external APIs utilizing a redis cache for repeat requests
func GetSynonyms(word string, apiKeys map[string]string) []string {

	cachedSynonyms := getCachedSynonyms(word)

	if cachedSynonyms == "" {
		fmt.Printf("No cached synonyms found. Fetching from api. \n")
		fetchedSynonyms, fetchedSynonymsError := getSynonymsFromMultipleApis(word, apiKeys)
		if fetchedSynonymsError != nil {
			//do something
			return []string;
		}
		result := setCachedSynonyms(word, joinSynonymsSlice(fetchedSynonyms))

		fmt.Printf("Setting result: %t \n", result)

		return fetchedSynonyms
	}

	fmt.Printf("Cached synonyms found. \n")
	return splitSynonymsString(cachedSynonyms)

}

// Private functions

func getSynonymsFromMultipleApis(word string, apiKeys map[string]string) ([]string, error) {

	results := make(chan []string)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		result, err := getSynonymsFromDictionaryAPI(word, apiKeys["dictionaryapi"])
		if err == nil {
			results <- result
		}
	}()

	go func() {
		defer wg.Done()
		result, err := getSynonymsFromWordnik(word, apiKeys["wordnik"])
		if err == nil {
			results <- result
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var mergedResults []string

	for result := range results {
		fmt.Printf("after concurrent stuff result: %q \n", result)
		mergedResults = append(mergedResults, result...)
	}

	fmt.Printf("mergedResults: %q \n", mergedResults)

	sort.Strings(mergedResults)

	// removing dupes
	dedupedResults := []string{}
	seen := map[string]bool{}
	for _, val := range mergedResults {
		if !seen[val] {
			dedupedResults = append(dedupedResults, val)
			seen[val] = true
		}
	}

	return dedupedResults, nil

}

func getSynonymsFromDictionaryAPI(word string, apiKey string) ([]string, error) {

	url := fmt.Sprintf("http://www.dictionaryapi.com/api/v1/references/thesaurus/xml/%s?key=%s", word, apiKey)
	synonymsResult, synonymsError := http.Get(url)

	if synonymsError != nil {
		fmt.Printf("Request error \n")
		log.Fatal(synonymsError)
	}

	defer synonymsResult.Body.Close()

	if synonymsResult.StatusCode != 200 {
		return nil, fmt.Errorf(`{ "status": %q, "code": %d }`, synonymsResult.Status, synonymsResult.StatusCode)
	}

	xmlParser := xml.NewDecoder(synonymsResult.Body)

	shouldCaptureText := false
	var rawSynonyms bytes.Buffer

	for {

		token, _ := xmlParser.Token()

		if token == nil {
			break
		}

		//fmt.Printf("Handling token \n")
		switch se := token.(type) {
			case xml.StartElement:
				if se.Name.Local == "syn" || se.Name.Local == "rel" {
					shouldCaptureText = true
				}
				break
			case xml.EndElement:
				break
			default:
				if shouldCaptureText == true {
					rawSynonyms.WriteString(string([]byte(se.(xml.CharData))))
					rawSynonyms.WriteString(",")
					shouldCaptureText = false
				}
		}

	}

	regexRemovals := regexp.MustCompile(`(\s|\[\]|-)`)
	regexSemicolons := regexp.MustCompile(`([;])`)

	wordsAfterRemovals := regexRemovals.ReplaceAllLiteralString(rawSynonyms.String(), "")
	wordsAfterSemicolons := regexSemicolons.ReplaceAllLiteralString(wordsAfterRemovals, "")

	return splitSynonymsString(wordsAfterSemicolons), nil
}

//
type wordnikWordSet struct {
	RelationshipType string
	Words []string
}

type wordnikResult struct {
	Results []wordnikWordSet
}

// beginnings of wordnik api integration
func getSynonymsFromWordnik(word string, apiKey string) ([]string, error) {

	url := fmt.Sprintf("https://api.wordnik.com/v4/word.json/%s/relatedWords?useCanonical=false&limitPerRelationshipType=100&api_key=%s", word, apiKey)
	fmt.Printf("GET %s \n", url)

	synonymsResult, synonymsError := http.Get(url)

	if synonymsError != nil {
		fmt.Printf("Request error \n")
		log.Fatal(synonymsError)
	}

	defer synonymsResult.Body.Close()

	fmt.Printf("Response Status Code: %d \n", synonymsResult.StatusCode)
	fmt.Printf("Response Status: %s \n", synonymsResult.Status)

	if synonymsResult.StatusCode != 200 {
		return nil, fmt.Errorf(`{ "status": %q, "code": %d }`, synonymsResult.Status, synonymsResult.StatusCode)
	}

	var decodedBody []wordnikWordSet
	if decoderError := json.NewDecoder(synonymsResult.Body).Decode(&decodedBody); decoderError != nil {
		log.Fatal(decoderError)
		return nil, decoderError
	}

	var mergedResults []string
	for _, wordSet := range decodedBody {
		words := wordSet.Words

		switch wordSet.RelationshipType {
			case 
				"equivalent",
				"verb-form",
				"hypernym",
				"etymologically-related-term",
				"variant",
				"synonym",
				"same-context":
					mergedResults = append(mergedResults, words...)
		}
	}

	return mergedResults, nil

}

func getCachedSynonyms(word string) string {

	// Im not quite sure about the reference here. Just found in redis client examples.
	// original line: client := redis.NewClient(&redis.Options{
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	defer client.Close()

	result, err := client.Get(word).Result()
	if err != nil {
		fmt.Printf("getCachedSynonyms error: %q \n", err)
		return ""
	}
	fmt.Printf("getCachedSynonyms result: %q \n", result)

	return result

}

func setCachedSynonyms(word string, synonyms string) bool {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	defer client.Close()

	expiration := time.Duration(24*180) * time.Hour

	result, err := client.Set(word, synonyms, expiration).Result()
	if err != nil {
		fmt.Printf("settingCachedSynonyms error: %q \n", err.Error())
		return false
	}
	fmt.Printf("settingCachedSynonyms result: %q \n", result)

	return true

}

func splitSynonymsString(rawSynonyms string) []string {

	regexParens := regexp.MustCompile(`(\(.*\)|\()`)
	synonyms := strings.Split(rawSynonyms, ",")

	var cleanSynonyms []string

	for _, rawWord := range synonyms {
		if rawWord != "" {
			cleanSynonyms = append(cleanSynonyms, regexParens.ReplaceAllLiteralString(rawWord, ""))
		}
	}

	return cleanSynonyms

}
func joinSynonymsSlice(rawSynonyms []string) string {

	var synonymsString bytes.Buffer

	for _, word := range rawSynonyms {
		synonymsString.WriteString(word)
		synonymsString.WriteString(",")
	}

	return synonymsString.String()
}
