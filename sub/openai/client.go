package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"

	"github.com/burstsms/hack-making-it-known/sub/types"
)

type Client struct {
	client *openai.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c Client) CreateChatCompletion(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: req.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.Message,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &types.CompletionResponse{Message: resp.Choices[0].Message.Content}, nil
}
