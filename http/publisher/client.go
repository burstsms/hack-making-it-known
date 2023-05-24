package publisher

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"cloud.google.com/go/pubsub"

	"github.com/burstsms/hack-making-it-known/types"
)

type Client struct {
	topic *pubsub.Topic
}

func (c Client) Publish(ctx context.Context, se types.SlackMessageEvent) (err error) {
	// encode the Slack event message into json
	eventJson, err := json.Marshal(se)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return err
	}

	// Publish the Slack event message to the topic
	result := c.topic.Publish(ctx, &pubsub.Message{Data: eventJson})

	// Get the server-generated message ID
	msgID, err := result.Get(ctx)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}

	log.Printf("Published a message; msg ID: %v\n", msgID)
	return nil

}

func New() *Client {
	topicName := os.Getenv("CLOUD_PUB_SUB_TOPIC")
	if topicName == "" {
		topicName = "hack-slack-bridge"
		log.Printf("CLOUD_PUB_SUB_TOPIC environment variable not set, using %s", topicName)
	}
	projectID := os.Getenv("CLOUD_PROJECT_ID")
	if projectID == "" {
		projectID = "tmp-hack-no-team-name"
		log.Printf("CLOUD_PROJECT_ID environment variable not set, using %s", projectID)
	}

	// Set up the Google Cloud Pub/Sub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}

	// Get a reference to the topic
	return &Client{
		topic: client.Topic(topicName),
	}
}
