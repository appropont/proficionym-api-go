package main

import (
    //"encoding/json"
    "net/http"
    "fmt"

    "proficionym/domains"

    "github.com/gorilla/mux"
)

func main() {    
    r := mux.NewRouter()
    r.HandleFunc("/whois/{domain}", func(response http.ResponseWriter, request *http.Request) {
        lookupResult := domains.WhoisLookup(mux.Vars(request)["domain"])
        response.Header().Set("Content-Type", "application/json")
        response.Write([]byte(lookupResult))
    }).Methods("GET")

    r.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
        response.Header().Set("Content-Type", "application/json")
        response.Write([]byte(`{ "name": "Proficionym API", "version": "0.0.1" }`))
    }).Methods("GET")

    fmt.Printf("Starting Server on port 8080...")

    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}