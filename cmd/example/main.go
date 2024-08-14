package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/codigogp/letsgollm/llm"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

	// Obtener la clave API de Anthropic desde las variables de entorno
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("La clave API de Anthropic no está configurada en el archivo .env")
	}

	// Crear una instancia de LLM para Anthropic
	anthropicLLM, err := llm.New(llm.Config{
		Provider:    llm.Anthropic,
		ModelName:   "claude-2", // Asegúrate de usar el nombre correcto del modelo
		APIKey:      apiKey,
		MaxTokens:   100,
		Temperature: 0.7,
	})
	if err != nil {
		log.Fatalf("Error creando la instancia de Anthropic LLM: %v", err)
	}

	// Definir un prompt de ejemplo
	prompt := "Explica el concepto de inteligencia artificial en 3 frases cortas."

	// Generar una respuesta
	ctx := context.Background()
	response, err := anthropicLLM.GenerateResponse(ctx, prompt)
	if err != nil {
		log.Fatalf("Error generando respuesta: %v", err)
	}

	// Imprimir la respuesta
	fmt.Printf("Prompt: %s\n\nRespuesta:\n%s\n", prompt, response)

	// Ejemplo de uso asíncrono
	fmt.Println("\nGenerando respuesta de forma asíncrona...")
	respChan, errChan := anthropicLLM.GenerateResponseAsync(ctx, "¿Cuál es el futuro de la IA?")

	select {
	case resp := <-respChan:
		fmt.Printf("\nRespuesta asíncrona:\n%s\n", resp)
	case err := <-errChan:
		log.Fatalf("Error en la generación asíncrona: %v", err)
	}
}
