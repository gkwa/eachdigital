package core

import (
	"fmt"
	"log"
	"time"

	"google.golang.org/api/gmail/v1"
)

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

	for _, message := range response.Messages {
		msg, err := srv.Users.Messages.Get("me", message.Id).Format("metadata").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", message.Id, err)
		}
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				fmt.Println(header.Value)
				break
			}
		}
	}

	fmt.Printf("Token valid for: %v\n", expiresIn.Round(time.Second))
}
