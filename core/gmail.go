package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Credentials struct {
	Web struct {
		ClientID     string   `json:"client_id"`
		ProjectID    string   `json:"project_id"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		CertURL      string   `json:"auth_provider_x509_cert_url"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"web"`
}

func ListTodaysEmails() {
	credentialsFile := os.Getenv("EACHDIGITAL_CREDENTIALS_FILE")
	if credentialsFile == "" {
		log.Fatal("EACHDIGITAL_CREDENTIALS_FILE environment variable not set")
	}

	config := getOAuthConfig(credentialsFile)

	client, expiresIn := getClient(config)

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	today := time.Now().Format("2006/01/02")

	call := srv.Users.Messages.List("me").Q("after:" + today)

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

func getOAuthConfig(credentialsFile string) *oauth2.Config {
	credentialsData, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	var credentials Credentials
	err = json.Unmarshal(credentialsData, &credentials)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	config := &oauth2.Config{
		ClientID:     credentials.Web.ClientID,
		ClientSecret: credentials.Web.ClientSecret,
		RedirectURL:  credentials.Web.RedirectURIs[0],
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	return config
}

func getClient(config *oauth2.Config) (*http.Client, time.Duration) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	expiresIn := time.Until(tok.Expiry)
	if expiresIn < 0 {
		fmt.Println("Token has expired. Refreshing...")
		newToken, err := config.TokenSource(context.Background(), tok).Token()
		if err != nil {
			log.Fatalf("Unable to refresh token: %v", err)
		}
		tok = newToken
		saveToken(tokFile, tok)
		expiresIn = time.Until(tok.Expiry)
	}

	client := config.Client(context.Background(), tok)

	return client, expiresIn
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	redirectURL := "http://localhost:8080/oauth2callback"
	fmt.Printf("Redirect URL: %s\n", redirectURL)

	codeCh := make(chan string)
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeCh <- code
		fmt.Fprintf(w, "Authorization code received. You can now close this browser window.")
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start local server: %v", err)
		}
	}()

	authCode := <-codeCh

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Unable to encode oauth token: %v", err)
	}
}
