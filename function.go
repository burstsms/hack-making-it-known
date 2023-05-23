// Package mik contains an HTTP Cloud Function.
package mik

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"mik/openai"
)

var client openai.OaiClient

func init() {
	log.Println("init")
	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

// Handler sends the received message to OpenAI and returns the response.
func Handler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Message string `json:"message"`
	}

	log.Println("received request")

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	log.Printf("message: %s", d.Message)

	completion, err := client.CreateChatCompletion(context.Background(), &openai.CompletionRequest{Message: d.Message})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Println(completion.Message)

	_, err = fmt.Fprint(w, html.EscapeString(completion.Message))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
