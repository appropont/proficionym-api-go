package synonyms

import (
    //"encoding/json"
    //"builtin"
    "net/http"
    "fmt"
    "log"
    "bytes"
    "encoding/xml"
    "regexp"
    "errors"
    "strings"
    //"io/ioutil"

    "github.com/davecgh/go-spew/spew"
)

func GetSynonyms(word string, apiKey string) []string {
    synonyms, synonymsError := getSynonymsFromDictionaryApi(word, apiKey)
    if synonymsError != nil {
        //do something
    }

    return synonyms
}

func getSynonymsFromDictionaryApi (word string, apiKey string) ([]string, error) {

    url := fmt.Sprintf("http://www.dictionaryapi.com/api/v1/references/thesaurus/xml/%s?key=%s", word, apiKey)
    fmt.Printf("GET %s", url)
    synonymsResult, synonymsError := http.Get(url)

    if synonymsError != nil {
        fmt.Printf("Request error \n")
        log.Fatal(synonymsError)
    }

    defer synonymsResult.Body.Close()

    if synonymsResult.StatusCode != 200 {
        return nil, errors.New(fmt.Sprintf(`{ "status": %q, "code": %d }`, synonymsResult.Status, synonymsResult.StatusCode))
    }

    xmlParser := xml.NewDecoder(synonymsResult.Body)

    shouldCaptureText := false
    var rawSynonyms bytes.Buffer

    for {

        token, _ := xmlParser.Token();

        if(token == nil) {
            break
        }

        //fmt.Printf("Handling token \n")
        switch se := token.(type) { 
        case xml.StartElement: 
            if (se.Name.Local == "syn" || se.Name.Local == "rel") {
                shouldCaptureText = true
            }
            break
        case xml.EndElement:
            break
        default:
            if(shouldCaptureText == true) {
                spew.Dump(se)
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

func splitSynonymsString (rawSynonyms string) []string {

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