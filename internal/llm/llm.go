package llm

import (
	"context"
	"fmt"
)

// Provider representa los diferentes proveedores de LLM soportados
type Provider int

const (
	OpenAI Provider = iota
	Gemini
	Anthropic
)

// LLM define la interfaz para interactuar con modelos de lenguaje
type LLM interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
	GenerateResponseAsync(ctx context.Context, prompt string) (<-chan string, <-chan error)
}

// Config contiene la configuración para crear una instancia de LLM
type Config struct {
	Provider    Provider
	ModelName   string
	APIKey      string
	MaxTokens   int
	Temperature float64
}

// New crea una nueva instancia de LLM basada en la configuración proporcionada
func New(cfg Config) (LLM, error) {
	switch cfg.Provider {
	case OpenAI:
		return newOpenAI(cfg)
	case Gemini:
		return newGemini(cfg)
	case Anthropic:
		return newAnthropic(cfg)
	default:
		return nil, fmt.Errorf("proveedor LLM no soportado: %v", cfg.Provider)
	}
}
