// Package sub provides a subscriber for the MakeItKnown Slack app.
package sub

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/slack-go/slack"

	"github.com/burstsms/hack-making-it-known/sub/openai"
	"github.com/burstsms/hack-making-it-known/sub/types"
)

type OaiClient interface {
	CreateChatCompletion(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)
}

var oaiClient OaiClient

func init() {
	oaiClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	functions.CloudEvent("HelloPubSub", handler)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// handler consumes a CloudEvent message and extracts the Pub/Sub message.
func handler(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	name := string(msg.Message.Data) // Automatically decoded from base64.
	if name == "" {
		name = "World"
	}
	log.Printf("Hello, %s!", name)
	event := types.SlackMessageEvent{}
	askOpenAI(event)
	return nil
}

// AskOpenAI is triggered by an event it is subscribed to and sends the received message in the event to OpenAI and then sends the completion to slack\ .
// func AskOpenAI(ctx context.Context, event v2.Event) error {

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

//return nil
//}
// 	return nil
// }

func askOpenAI(event types.SlackMessageEvent) {
	log.Printf("message: %s", event.Event.Text)
	completion, err := oaiClient.CreateChatCompletion(context.Background(), &openai.CompletionRequest{Message: event.Event.Text})
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
