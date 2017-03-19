package domains

import (
	"fmt"
	"log"
	"regexp"

  "github.com/domainr/whois"
)

func outputStatus(domain string, status string) string {
	return fmt.Sprintf(`{ "domain": %q, "status": %q }`, domain, status)
}

func WhoisLookup(domain string) string {

  request, err := whois.NewRequest(domain)
	if err != nil {
		log.Fatal(err)
	}

	response, err := whois.DefaultClient.Fetch(request)
	if err != nil {
		log.Fatal(err)
	} 

	availableResult, availableError := regexp.Match("No match for", response.Body)
	if availableError != nil {
		return outputStatus(domain, "error")
	}
	unavailableResult, unavailableError := regexp.Match("Domain Name:", response.Body)
	if unavailableError != nil {
		return outputStatus(domain, "error")
	}

	if availableResult {
		return outputStatus(domain, "available")
	} else if unavailableResult {
		return outputStatus(domain, "registered")
	} else {
		return outputStatus(domain, "error")
	}

}
