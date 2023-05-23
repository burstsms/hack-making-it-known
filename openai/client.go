package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type CompletionRequest struct {
	Message string
}
type CompletionResponse struct {
	Message string
}

type OaiClient interface {
	CreateChatCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

type Client struct {
	client *openai.Client
}

func NewClient(apiKey string) OaiClient {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c Client) CreateChatCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
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
	return &CompletionResponse{Message: resp.Choices[0].Message.Content}, nil
}
