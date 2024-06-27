package core

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

type EmailInfo struct {
	Subject string
	Name    string
	Email   string
}

func ListTodaysEmails() {
	srv, expiresIn := setupGmailService()
	call := getTodaysEmailsCall(srv)
	processEmailCall(srv, call, expiresIn)
}

func ListNonSubscriptionEmails() {
	srv, expiresIn := setupGmailService()
	call := getNonSubscriptionEmailsLastMonthCall(srv)
	processEmailCall(srv, call, expiresIn)
}

func processEmailCall(srv *gmail.Service, call *gmail.UsersMessagesListCall, expiresIn time.Duration) {
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	domainMap := make(map[string]map[string][]EmailInfo)

	for _, message := range response.Messages {
		msg, err := srv.Users.Messages.Get("me", message.Id).Format("metadata").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", message.Id, err)
		}

		var subject, from string
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "Subject":
				subject = header.Value
			case "From":
				from = header.Value
			}
		}

		name, email, domain := parseFromField(from)
		info := EmailInfo{
			Subject: subject,
			Name:    name,
			Email:   email,
		}

		if _, ok := domainMap[domain]; !ok {
			domainMap[domain] = make(map[string][]EmailInfo)
		}
		domainMap[domain][from] = append(domainMap[domain][from], info)
	}

	printGroupedResults(domainMap)

	fmt.Printf("\nToken valid for: %v\n", expiresIn.Round(time.Second))
}

func parseFromField(from string) (name, email, domain string) {
	parts := strings.SplitN(from, "<", 2)
	if len(parts) == 2 {
		name = strings.TrimSpace(parts[0])
		email = strings.TrimSuffix(parts[1], ">")
	} else {
		email = from
	}

	atParts := strings.SplitN(email, "@", 2)
	if len(atParts) == 2 {
		domain = atParts[1]
	}

	return
}

func printGroupedResults(domainMap map[string]map[string][]EmailInfo) {
	domains := make([]string, 0, len(domainMap))
	for domain := range domainMap {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	for _, domain := range domains {
		fmt.Printf("Domain: %s\n", domain)

		froms := make([]string, 0, len(domainMap[domain]))
		for from := range domainMap[domain] {
			froms = append(froms, from)
		}
		sort.Strings(froms)

		for _, from := range froms {
			fmt.Printf("  From: %s\n", from)
			for _, info := range domainMap[domain][from] {
				fmt.Printf("    Subject: %s\n", info.Subject)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
