package llm

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type geminiLLM struct {
	client *genai.Client
	model  *genai.GenerativeModel
	config Config
}

func newGemini(cfg Config) (LLM, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("error creando cliente Gemini: %w", err)
	}

	model := client.GenerativeModel(cfg.ModelName)
	model.SetTemperature(float32(cfg.Temperature))

	return &geminiLLM{
		client: client,
		model:  model,
		config: cfg,
	}, nil
}

func (g *geminiLLM) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generando respuesta de Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no se generaron candidatos")
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			result += string(textPart)
		}
	}

	if result == "" {
		return "", fmt.Errorf("no se generó contenido de texto")
	}

	return result, nil
}

func (g *geminiLLM) GenerateResponseAsync(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	respChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(respChan)
		defer close(errChan)

		resp, err := g.GenerateResponse(ctx, prompt)
		if err != nil {
			errChan <- err
			return
		}

		respChan <- resp
	}()

	return respChan, errChan
}
