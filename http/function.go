// Package mik contains an HTTP Cloud Function.
package http

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"github.com/burstsms/hack-making-it-known/publisher"
	"github.com/burstsms/hack-making-it-known/slack"
	"github.com/burstsms/hack-making-it-known/types"
)

type CloudSlackEventPublisher interface {
	Publish(ctx context.Context, se types.SlackMessageEvent) (err error)
}

var pubclient CloudSlackEventPublisher

func init() {
	log.Println("init")
	functions.HTTP("MakeItKnown", Handler)

	pubclient = publisher.New()
}

// Handler sends the received message to OpenAI and returns the response.
func Handler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	// request body needs to be a byte array
	// so we can use it in multiple places
	body, err := io.ReadAll(r.Body)

	// is this a URL verification request?
	// returns true if it is and we finish here
	if slack.HandleURLValidation(w, r.Method, body) {
		return
	}

	// validate the Slack app HMAC signature
	if !slack.ValidateSignature(r.Header, body) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// parse the Slack message event
	var event types.SlackMessageEvent
	err = json.Unmarshal(body, &event)
	if err != nil {
		if err == io.EOF {
			log.Println("EOF")
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	log.Printf("received: %+v", event)

	// ignore events that are not message events
	if event.Type != "event_callback" || event.Event.Type != "message" {
		log.Printf("ignoring event type: %s", event.Type)
		return
	}
	log.Println("is a valid message event")

	// we pass off the message to OpenAI
	err = AskOpenAI(ctx, event)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

// AskOpenAI publishes the slack event to our make-it-known cloud function subscriber who will do the OpenAI asking
// and giving the completion back to slack.
func AskOpenAI(ctx context.Context, event types.SlackMessageEvent) error {
	log.Printf("asking: %s", event.Event.Text)

	return pubclient.Publish(ctx, event)
}
