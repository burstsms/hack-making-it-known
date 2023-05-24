// Package mik contains an HTTP Cloud Function.
package http

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/pubsub"
	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"github.com/burstsms/hack-making-it-known/slack"
	"github.com/burstsms/hack-making-it-known/types"
)

type SlackClient interface {
	PostCompletionMessage(event types.SlackMessageEvent, message string) error
}
type OaiClient interface {
	CreateChatCompletion(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)
}

const topicName = "hack-slack-bridge"
const projectID = "tmp-hack-no-team-name"

func init() {
	log.Println("init")
	functions.HTTP("MakeItKnown", Handler)
}

// Handler sends the received message to OpenAI and returns the response.
func Handler(w http.ResponseWriter, r *http.Request) {

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
		return
	}

	// ignore events that are not message events
	if event.Type != "event_callback" || event.Event.Type != "message" {
		log.Printf("ignoring event type: %s", event.Type)
		return
	}

	// we pass off the message to OpenAI
	err = AskOpenAI(event)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

// AskOpenAI publishes the slack event to our make-it-known cloud function subscriber who will do the OpenAI asking
// and giving the completion back to slack.
func AskOpenAI(event types.SlackMessageEvent) error {
	log.Printf("asking: %s", event.Event.Text)

	// Set up the Google Cloud Pub/Sub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
		return err
	}

	// Get a reference to the topic
	topic := client.Topic(topicName)

	// encode the Slack event message into json
	eventJson, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return err
	}

	// Publish the Slack event message to the topic
	result := topic.Publish(ctx, &pubsub.Message{Data: eventJson})

	// Get the server-generated message ID
	msgID, err := result.Get(ctx)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}

	log.Printf("Published a message; msg ID: %v\n", msgID)

	return nil
}
