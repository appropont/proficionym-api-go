package main

import (
    "encoding/json"
    "net/http"
    "fmt"
    "log"
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

    synonymsApiKey := os.Getenv("DICTIONARYAPI_API_KEY")
    //synonymsApiKey := os.Getenv("WORDNIK_API_KEY")

    r := mux.NewRouter()
    r.HandleFunc("/whois/{domain}", func(response http.ResponseWriter, request *http.Request) {
        whoisResult := domains.WhoisLookup(mux.Vars(request)["domain"])
        response.Header().Set("Content-Type", "application/json")
        response.Write([]byte(whoisResult))
    }).Methods("GET")

    r.HandleFunc("/synonyms/{word}", func(response http.ResponseWriter, request *http.Request) {
        synonymsResult := synonyms.GetSynonyms(mux.Vars(request)["word"], synonymsApiKey)
        jsonSynonyms, _ := json.Marshal(synonymsResult)
        response.Header().Set("Content-Type", "application/json")
        response.Write([]byte(fmt.Sprintf(`{ "synonyms": %s }`, jsonSynonyms)))
    }).Methods("GET")

    r.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
        response.Header().Set("Content-Type", "application/json")
        response.Write([]byte(`{ "name": "Proficionym API", "version": "0.0.1" }`))
    }).Methods("GET")

    fmt.Printf("Starting Server on port 8080... \n")

    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}