package llm

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

type openAILLM struct {
	client *openai.Client
	config Config
}

func newOpenAI(cfg Config) (LLM, error) {
	client := openai.NewClient(cfg.APIKey)
	return &openAILLM{
		client: client,
		config: cfg,
	}, nil
}

func (o *openAILLM) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: o.config.ModelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   o.config.MaxTokens,
			Temperature: float32(o.config.Temperature),
		},
	)

	if err != nil {
		return "", fmt.Errorf("error generando respuesta de OpenAI: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *openAILLM) GenerateResponseAsync(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	respChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(respChan)
		defer close(errChan)

		resp, err := o.GenerateResponse(ctx, prompt)
		if err != nil {
			errChan <- err
			return
		}

		respChan <- resp
	}()

	return respChan, errChan
}
