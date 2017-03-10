package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"proficionym/domains"
	"proficionym/synonyms"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKeys := make(map[string]string)
	apiKeys["dictionaryapi"] = os.Getenv("DICTIONARYAPI_API_KEY")
	apiKeys["wordnik"] = os.Getenv("WORDNIK_API_KEY")

	r := mux.NewRouter()
	r.HandleFunc("/whois/{domain}", func(response http.ResponseWriter, request *http.Request) {
		whoisResult := domains.WhoisLookup(mux.Vars(request)["domain"])
		response.Header().Set("Content-Type", "application/json")
		response.Write([]byte(whoisResult))
	}).Methods("GET")

	r.HandleFunc("/synonyms/{word}", func(response http.ResponseWriter, request *http.Request) {
		synonymsResult := synonyms.GetSynonyms(mux.Vars(request)["word"], apiKeys)
		jsonSynonyms, _ := json.Marshal(synonymsResult)
		response.Header().Set("Content-Type", "application/json")
		response.Write([]byte(fmt.Sprintf(`{ "synonyms": %s }`, jsonSynonyms)))
	}).Methods("GET")

	r.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.Write([]byte(`{ "name": "Proficionym API", "version": "0.0.1" }`))
	}).Methods("GET")

	fmt.Printf("Starting Server on port " + os.Getenv("PORT") + "... \n")

	http.Handle("/", r)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
