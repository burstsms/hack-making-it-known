package slack

import (
	"os"

	"github.com/slack-go/slack"

	"github.com/burstsms/hack-making-it-known/sub/types"
)

var slackAPI *slack.Client

func init() {
	slackAPI = slack.New(os.Getenv("SLACK_BOT_TOKEN"))
}

type Client struct {
	botToken string
}

func New(botToken string) *Client {
	return &Client{botToken: botToken}
}

func (s Client) PostCompletionMessage(event types.SlackMessageEvent, message string) error {
	_, _, err := slackAPI.PostMessage(
		event.Event.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(event.Event.Ts),
	)
	if err != nil {
		return err
	}
	return nil
}
