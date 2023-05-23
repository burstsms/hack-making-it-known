package slack

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

func HandleURLValidation(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		var urlVerification map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&urlVerification)
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
func ValidateSignature(r *http.Request) bool {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	// if the timestamp is more than five minutes from local time, reject the request
	// https://api.slack.com/authentication/verifying-requests-from-slack#additional_verification_steps
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		log.Printf("slack.NewSecretsVerifier: %v", err)
		return false
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		return false
	}
	_, err = verifier.Write(b)
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
