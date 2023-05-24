// Package sub provides a subscriber for the MakeItKnown Slack app.
package sub

import (
	"context"
	"encoding/json"
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
var modelEnv string

func init() {
	oaiClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	modelEnv = os.Getenv("OPENAI_MODEL")
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

	payload := string(msg.Message.Data) // Automatically decoded from base64.
	log.Printf("Payload: %s", payload)
	slackEvent := types.SlackMessageEvent{}
	json.Unmarshal(msg.Message.Data, &slackEvent)

	askOpenAI(slackEvent)
	return nil
}

func askOpenAI(event types.SlackMessageEvent) {
	log.Printf("message: %s", event.Event.Text)
	model := "gpt-4"
	if modelEnv != "" {
		model = modelEnv
	}
	completion, err := oaiClient.CreateChatCompletion(context.Background(), &types.CompletionRequest{Message: event.Event.Text, Model: model})
	if err != nil {
		log.Printf("error calling OpenAI API: %s", err.Error())
		return
	}
	log.Println(completion.Message)
	// send completion to Slack
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	_, _, err = api.PostMessage(event.Event.Channel, slack.MsgOptionText(completion.Message, false), slack.MsgOptionTS(event.Event.Ts))
	if err != nil {
		log.Printf("error sending message to Slack: %s", err.Error())
		return
	}
}
