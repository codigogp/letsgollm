# letsgollm

⚡ Tu puerta de entrada fácil a la IA avanzada en Go ⚡

## 🤔 ¿Qué es letsgollm?

letsgollm es una librería de código abierto en Go diseñada para simplificar las interacciones con Grandes Modelos de Lenguaje (LLMs) para investigadores y principiantes. Ofrece una interfaz unificada para diferentes proveedores de LLM y un conjunto de herramientas para mejorar las capacidades de los modelos de lenguaje y facilitar el desarrollo de herramientas y aplicaciones basadas en IA.

## Características

- **Interfaz LLM unificada**: Define una instancia de LLM en una línea para proveedores como OpenAI y Google Gemini.
- **Cargador de texto genérico**: Carga texto de diversas fuentes como archivos DOCX, PDF, TXT, scripts de YouTube o publicaciones de blog.
- **Conector RapidAPI**: Conéctate con servicios de IA en RapidAPI.
- **Integración SERP**: Realiza búsquedas utilizando diferentes motores de búsqueda.
- **Generador de plantillas de prompts**: Crea y gestiona fácilmente plantillas de prompts.
- **Chunking de texto**: Divide textos largos en fragmentos manejables basados en diferentes criterios.

## Instalación

```bash
go get github.com/codigogp/letsgollm
```

## Configuración

Para usar esta librería, necesitas configurar varias claves de API en tu entorno. Crea un archivo `.env` en el directorio raíz de tu proyecto y añade tus claves de API allí.

Ejemplo de `.env`:

```
OPENAI_API_KEY="tu_clave_api_openai"
GEMINI_API_KEY="tu_clave_api_gemini"
ANTHROPIC_API_KEY="tu_clave_api_anthropic"
RAPIDAPI_API_KEY="tu_clave_api_rapidapi"
SERPER_API_KEY="tu_clave_api_serper"
```

## Uso

### Creación de una instancia LLM

```go
import (
    "github.com/codigogp/letsgollm/llm"
)

func main() {
    llmInstance, err := llm.New(llm.Config{
        Provider:   llm.OpenAI,
        ModelName:  "gpt-3.5-turbo",
        APIKey:     os.Getenv("OPENAI_API_KEY"),
        MaxTokens:  100,
        Temperature: 0.7,
    })
    if err != nil {
        log.Fatalf("Error creando la instancia de LLM: %v", err)
    }

    response, err := llmInstance.GenerateResponse(context.Background(), "Genera una frase de 5 palabras sobre IA")
    if err != nil {
        log.Fatalf("Error generando respuesta: %v", err)
    }
    fmt.Println(response)
}
```

### Uso de herramientas

```go
import (
    "github.com/codigogp/letsgollm/tools/serp"
    "github.com/codigogp/letsgollm/tools/loader"
)

// Búsqueda SERP
results, err := serp.SearchWithSerperAPI(context.Background(), "Go programming language", os.Getenv("SERPER_API_KEY"), 3)

// Carga de contenido
content, err := loader.LoadContent("archivo.pdf")
```

## Contribución

Las contribuciones son bienvenidas! Por favor, lee las directrices de contribución antes de enviar un pull request.

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.
