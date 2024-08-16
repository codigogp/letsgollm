package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LoadJSONFile carga un archivo JSON y lo decodifica en la estructura proporcionada
func LoadJSONFile(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error al leer el archivo JSON: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("error al decodificar el JSON: %w", err)
	}

	return nil
}

// SaveJSONFile guarda una estructura como archivo JSON
func SaveJSONFile(filePath string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("error al codificar el JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo JSON: %w", err)
	}

	return nil
}

// TruncateText trunca un texto a una longitud máxima, añadiendo "..." si es necesario
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// ExtractDomain extrae el dominio de una URL
func ExtractDomain(url string) string {
	re := regexp.MustCompile(`^(?:https?:\/\/)?(?:[^@\n]+@)?(?:www\.)?([^:\/\n]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// RemoveHTMLTags elimina las etiquetas HTML de un string
func RemoveHTMLTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(html, "")
}

// Slugify convierte un string en un slug (URL-friendly)
func Slugify(text string) string {
	// Convertir a minúsculas
	text = strings.ToLower(text)

	// Reemplazar espacios con guiones
	text = strings.ReplaceAll(text, " ", "-")

	// Eliminar caracteres no alfanuméricos
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	text = re.ReplaceAllString(text, "")

	// Eliminar guiones múltiples
	re = regexp.MustCompile(`-+`)
	text = re.ReplaceAllString(text, "-")

	// Eliminar guiones al principio y al final
	return strings.Trim(text, "-")
}

// ContainsAny verifica si un slice contiene alguno de los elementos dados
func ContainsAny(slice []string, elements ...string) bool {
	for _, s := range slice {
		for _, e := range elements {
			if s == e {
				return true
			}
		}
	}
	return false
}
