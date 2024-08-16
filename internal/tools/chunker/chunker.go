package chunker

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/jdkato/prose/v2"
)

// ChunkInfo representa la información de un chunk de texto
type ChunkInfo struct {
	Text          string
	NumCharacters int
	NumWords      int
}

// TextChunks representa una colección de chunks de texto
type TextChunks struct {
	NumChunks int
	ChunkList []ChunkInfo
}

// ChunkByMaxChunkSize divide el texto en chunks basados en el tamaño máximo de chunk
func ChunkByMaxChunkSize(text string, maxChunkSize int, preserveSentenceStructure bool) TextChunks {
	var chunks []ChunkInfo

	if preserveSentenceStructure {
		sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
		currentChunk := ""

		for _, sentence := range sentences {
			if len(currentChunk)+len(sentence) <= maxChunkSize {
				currentChunk += sentence + " "
			} else {
				if currentChunk != "" {
					chunks = append(chunks, createChunkInfo(strings.TrimSpace(currentChunk)))
				}
				if len(sentence) > maxChunkSize {
					chunks = append(chunks, createChunkInfo(sentence))
					currentChunk = ""
				} else {
					currentChunk = sentence + " "
				}
			}
		}

		if currentChunk != "" {
			chunks = append(chunks, createChunkInfo(strings.TrimSpace(currentChunk)))
		}
	} else {
		for i := 0; i < len(text); i += maxChunkSize {
			end := i + maxChunkSize
			if end > len(text) {
				end = len(text)
			}
			chunks = append(chunks, createChunkInfo(text[i:end]))
		}
	}

	return TextChunks{
		NumChunks: len(chunks),
		ChunkList: chunks,
	}
}

// ChunkBySentences divide el texto en chunks basados en oraciones
func ChunkBySentences(text string) TextChunks {
	doc, _ := prose.NewDocument(text)
	sentences := doc.Sentences()
	chunks := make([]ChunkInfo, len(sentences))

	for i, sentence := range sentences {
		chunks[i] = createChunkInfo(strings.TrimSpace(sentence.Text))
	}

	return TextChunks{
		NumChunks: len(chunks),
		ChunkList: chunks,
	}
}

// ChunkByParagraphs divide el texto en chunks basados en párrafos
func ChunkByParagraphs(text string) TextChunks {
	paragraphs := regexp.MustCompile(`\n\s*\n`).Split(text, -1)
	chunks := make([]ChunkInfo, 0, len(paragraphs))

	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed != "" {
			chunks = append(chunks, createChunkInfo(trimmed))
		}
	}

	return TextChunks{
		NumChunks: len(chunks),
		ChunkList: chunks,
	}
}

// ChunkBySemantics divide el texto en chunks basados en semántica
func ChunkBySemantics(text string, thresholdPercentage float64) TextChunks {
	sentences := splitSentences(text)
	combinedSentences := combineSentences(sentences)
	embeddings := convertToVector(combinedSentences)
	distances := calculateCosineSimilarities(embeddings)

	// Ordenar las distancias para el cálculo del umbral
	sortedDistances := make([]float64, len(distances))
	copy(sortedDistances, distances)
	sort.Float64s(sortedDistances)

	// Calcular el umbral usando el percentil
	thresholdIndex := int(float64(len(sortedDistances)-1) * thresholdPercentage / 100)
	breakpointDistanceThreshold := sortedDistances[thresholdIndex]

	var chunks []ChunkInfo
	startIndex := 0

	for i, distance := range distances {
		if distance > breakpointDistanceThreshold {
			chunk := strings.Join(sentences[startIndex:i+1], " ")
			chunks = append(chunks, createChunkInfo(chunk))
			startIndex = i + 1
		}
	}

	if startIndex < len(sentences) {
		chunk := strings.Join(sentences[startIndex:], " ")
		chunks = append(chunks, createChunkInfo(chunk))
	}

	return TextChunks{
		NumChunks: len(chunks),
		ChunkList: chunks,
	}
}

func createChunkInfo(text string) ChunkInfo {
	return ChunkInfo{
		Text:          text,
		NumCharacters: len(text),
		NumWords:      len(strings.Fields(text)),
	}
}

func splitSentences(text string) []string {
	return regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
}

func combineSentences(sentences []string) []string {
	combined := make([]string, len(sentences))
	for i, sentence := range sentences {
		if i > 0 {
			combined[i] = sentences[i-1] + " " + sentence
		}
		if i < len(sentences)-1 {
			combined[i] += " " + sentences[i+1]
		}
	}
	return combined
}

func convertToVector(sentences []string) [][]float64 {
	// Esta es una implementación simplificada. En la práctica, necesitarías usar
	// un modelo de embedding real, como Word2Vec o BERT.
	vectors := make([][]float64, len(sentences))
	for i, sentence := range sentences {
		vector := make([]float64, 10) // Usar 10 dimensiones como ejemplo
		for j, word := range strings.Fields(sentence) {
			vector[j%10] += float64(len(word))
		}
		vectors[i] = vector
	}
	return vectors
}

func calculateCosineSimilarities(vectors [][]float64) []float64 {
	similarities := make([]float64, len(vectors)-1)
	for i := 0; i < len(vectors)-1; i++ {
		similarities[i] = cosineSimilarity(vectors[i], vectors[i+1])
	}
	return similarities
}

func cosineSimilarity(a, b []float64) float64 {
	dot := 0.0
	normA := 0.0
	normB := 0.0
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
