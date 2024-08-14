package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type anthropicLLM struct {
	apiKey string
	config Config
}

func newAnthropic(cfg Config) (LLM, error) {
	return &anthropicLLM{
		apiKey: cfg.APIKey,
		config: cfg,
	}, nil
}

func (a *anthropicLLM) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/complete"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":                a.config.ModelName,
		"prompt":               fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		"max_tokens_to_sample": a.config.MaxTokens,
		"temperature":          a.config.Temperature,
	})
	if err != nil {
		return "", fmt.Errorf("error al crear el cuerpo de la solicitud: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error al crear la solicitud: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error al hacer la solicitud: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error al leer la respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error en la respuesta de Anthropic: %s", string(body))
	}

	var result struct {
		Completion string `json:"completion"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error al decodificar la respuesta: %w", err)
	}

	return result.Completion, nil
}

func (a *anthropicLLM) GenerateResponseAsync(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	respChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(respChan)
		defer close(errChan)

		resp, err := a.GenerateResponse(ctx, prompt)
		if err != nil {
			errChan <- err
			return
		}

		respChan <- resp
	}()

	return respChan, errChan
}
