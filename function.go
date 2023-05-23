// Package mik contains an HTTP Cloud Function.
package mik

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/slack-go/slack"

	"github.com/burstsms/hack-making-it-known/openai"
)

var client openai.OaiClient

func init() {
	log.Println("init")
	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

type SlackMessageEvent struct {
	Token    string `json:"token"`
	TeamId   string `json:"team_id"`
	ApiAppId string `json:"api_app_id"`
	Event    struct {
		Type        string `json:"type"`
		EventTs     string `json:"event_ts"`
		User        string `json:"user"`
		Text        string `json:"text"`
		Ts          string `json:"ts"`
		Channel     string `json:"channel"`
		ChannelType string `json:"channel_type"`
	} `json:"event"`
	Type        string   `json:"type"`
	EventId     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
}

// Handler sends the received message to OpenAI and returns the response.
func Handler(w http.ResponseWriter, r *http.Request) {

	// is this a URL verification request?
	// returns true if it is and we finish here
	if HandleURLValidation(w, r) {
		return
	}

	// validate the Slack app HMAC signature
	if !ValidateSignature(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// parse the Slack message event
	var event SlackMessageEvent
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

	// we end the response here so that Slack doesn't think we're timing out
	w.WriteHeader(http.StatusOK)

	// we pass off the message for the OpenAI API to a go routine here
	go AskOpenAI(event)
}

func AskOpenAI(event SlackMessageEvent) {
	log.Printf("message: %s", event.Event.Text)
	completion, err := client.CreateChatCompletion(context.Background(), &openai.CompletionRequest{Message: event.Event.Text})
	if err != nil {
		log.Printf("error calling OpenAI API: %s", err.Error())
		return
	}
	log.Println(completion.Message)
	// send completion to Slack
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	_, _, err = api.PostMessage(event.Event.Channel, slack.MsgOptionText(completion.Message, false))
	if err != nil {
		log.Printf("error sending message to Slack: %s", err.Error())
		return
	}
}

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

// validate the signature value as a HMAC signature
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

// SendToCloudTopic Publish function example
func SendToCloudTopic(w http.ResponseWriter, r *http.Request) {
	// Set up the Google Cloud Pub/Sub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "transmit-non-prod")
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}

	// topic name
	topicName := "hack-slack-bridge"

	// Get a reference to the topic
	topic := client.Topic(topicName)

	// Publish a message to the topic
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("First Test!"),
	})

	// Get the server-generated message ID
	msgID, err := result.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Printf("Published message with ID: %s\n", msgID)
}

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// HelloPubSub consumes a Pub/Sub message.
// This is a format sample
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	return nil
}
