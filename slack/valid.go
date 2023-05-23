package slack

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

func HandleURLValidation(w http.ResponseWriter, method string, body []byte) bool {
	if method == http.MethodPost {
		var urlVerification map[string]interface{}
		err := json.Unmarshal(body, &urlVerification)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return true
		}
		if urlVerification["type"] == "url_verification" {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(urlVerification["challenge"].(string)))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return true
			}
			return true
		}
	}
	return false
}

// ValidateSignature validate the signature value as a HMAC signature
// https://api.slack.com/authentication/verifying-requests-from-slack#about
func ValidateSignature(header http.Header, body []byte) bool {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if signingSecret == "" {
		log.Println("SLACK_SIGNING_SECRET is not set")
		return true
	}
	// if the timestamp is more than five minutes from local time, reject the request
	// https://api.slack.com/authentication/verifying-requests-from-slack#additional_verification_steps
	verifier, err := slack.NewSecretsVerifier(header, signingSecret)
	if err != nil {
		log.Printf("slack.NewSecretsVerifier: %v", err)
		return false
	}
	_, err = verifier.Write(body)
	if err != nil {
		log.Printf("verifier.Write: %v", err)
		return false
	}
	err = verifier.Ensure()
	if err != nil {
		log.Printf("verifier.Ensure: %v", err)
		return false
	}
	return true
}
