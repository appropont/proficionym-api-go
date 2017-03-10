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

	//cmd := exec.Command("whois", domain)
	//result, err := cmd.Output()

    result, err := whois.NewRequest(domain)

	if err != nil {
		log.Fatal(err)
	}

	availableResult, availableError := regexp.Match("No match for domain", result.Body)
	if availableError != nil {
		return outputStatus(domain, "error")
	}
	unavailableResult, unavailableError := regexp.Match("Domain Name:", result.Body)
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
