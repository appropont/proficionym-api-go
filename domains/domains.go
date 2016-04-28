package domains

import (
    //"encoding/json"
    //"net/http"
    "fmt"
    "log"
    "os/exec"
    "regexp"
)

func outputStatus (domain string, status string) string {
    return fmt.Sprintf(`{ "domain": %q, "status": %q }`, domain, status)
}

func WhoisLookup (domain string) string {

    cmd := exec.Command("whois", domain)
    result, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }

    availableResult, availableError := regexp.Match("No match for domain", result)
    if availableError != nil {
        return outputStatus(domain, "error")
    }
    unavailableResult, unavailableError := regexp.Match("Domain Name:", result)
    if unavailableError != nil {
        return outputStatus(domain, "error")
    }


    if availableResult {
        return outputStatus(domain, "available")
    } else if unavailableResult {
        return outputStatus(domain, "unavailable")
    } else {
        return outputStatus(domain, "error")
    }

}