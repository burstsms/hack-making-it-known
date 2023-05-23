// Package mik contains an HTTP Cloud Function.
package mik

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	"github.com/burstsms/hack-making-it-known/openai"
	"github.com/burstsms/hack-making-it-known/slack"
	"github.com/burstsms/hack-making-it-known/types"
)

type OaiClient interface {
	CreateChatCompletion(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)
}

var oaiClient OaiClient

type SlackClient interface {
	PostCompletionMessage(event types.SlackMessageEvent, message string) error
}

var slackClient SlackClient

func init() {
	log.Println("init")
	oaiClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	slackClient = slack.New(os.Getenv("SLACK_BOT_TOKEN"))
}

// Handler sends the received message to OpenAI and returns the response.
func Handler(w http.ResponseWriter, r *http.Request) {

	// is this a URL verification request?
	// returns true if it is and we finish here
	if slack.HandleURLValidation(w, r) {
		return
	}

	// validate the Slack app HMAC signature
	if !slack.ValidateSignature(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// parse the Slack message event
	var event types.SlackMessageEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		if err == io.EOF {
			log.Println("EOF")
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// ignore events that are not message events
	if event.Type != "event_callback" || event.Event.Type != "message" {
		log.Printf("ignoring event type: %s", event.Type)
		return
	}

	// we pass off the message for the OpenAI API to a go routine here
	AskOpenAI(event)

}

func AskOpenAI(event types.SlackMessageEvent) {
	log.Printf("message: %s", event.Event.Text)
	completion, err := oaiClient.CreateChatCompletion(context.Background(), &types.CompletionRequest{Message: event.Event.Text})
	if err != nil {
		log.Printf("error calling OpenAI API: %s", err.Error())
		return
	}
	log.Println(completion.Message)
	// send completion to Slack
	err = slackClient.PostCompletionMessage(event, completion.Message)
	if err != nil {
		log.Printf("error posting completion to Slack: %s", err.Error())
	}
}