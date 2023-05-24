// Package sub provides a subscriber for the MakeItKnown Slack app.
package sub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	v2 "github.com/cloudevents/sdk-go/v2"

	"github.com/burstsms/hack-making-it-known/openai"
	"github.com/burstsms/hack-making-it-known/types"
)

func init() {
	functions.CloudEvent("MakeItKnownSub", AskOpenAI)

	oaiClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

type OaiClient interface {
	CreateChatCompletion(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)
}

var oaiClient OaiClient

// AskOpenAI is triggered by an event it is subscribed to and sends the received message in the event to OpenAI and then sends the completion to slack\ .
func AskOpenAI(ctx context.Context, event v2.Event) error {

	// completion, err := oaiClient.CreateChatCompletion(context.Background(), &types.CompletionRequest{Message: event.Event.Text})
	// if err != nil {
	//     log.Printf("error calling OpenAI API: %s", err.Error())
	//     return
	// }
	// log.Println(completion.Message)
	// // send completion to Slack
	// err = slackClient.PostCompletionMessage(event, completion.Message)
	// if err != nil {
	//     log.Printf("error posting completion to Slack: %s", err.Error())
	// }

	return nil
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
