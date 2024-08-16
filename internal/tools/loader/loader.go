package loader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
	"github.com/unidoc/unioffice/document"
)

// TextDocument representa un documento de texto cargado
type TextDocument struct {
	FileSize       int64
	WordCount      int
	CharacterCount int
	Content        string
	Title          string
	URLOrPath      string
}

// LoadContent carga contenido de una ruta de archivo o URL dada
func LoadContent(inputPathOrURL string) (*TextDocument, error) {
	if isURL(inputPathOrURL) {
		if isYouTubeURL(inputPathOrURL) {
			return readYouTubeVideo(inputPathOrURL)
		}
		return readBlogFromURL(inputPathOrURL)
	}

	ext := strings.ToLower(filepath.Ext(inputPathOrURL))
	switch ext {
	case ".txt", ".csv":
		return readTextFile(inputPathOrURL)
	case ".docx":
		return readDocxFile(inputPathOrURL)
	case ".pdf":
		return readPDFFile(inputPathOrURL)
	default:
		// Intentar leer como archivo de texto por defecto
		return readTextFile(inputPathOrURL)
	}
}

func isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

func isYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

func readTextFile(filePath string) (*TextDocument, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)
	return &TextDocument{
		FileSize:       fileInfo.Size(),
		WordCount:      len(strings.Fields(text)),
		CharacterCount: len(text),
		Content:        text,
		URLOrPath:      filePath,
	}, nil
}

func readDocxFile(filePath string) (*TextDocument, error) {
	doc, err := document.Open(filePath)
	if err != nil {
		return nil, err
	}

	var content strings.Builder
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			content.WriteString(run.Text())
		}
		content.WriteString("\n")
	}

	text := content.String()
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &TextDocument{
		FileSize:       fileInfo.Size(),
		WordCount:      len(strings.Fields(text)),
		CharacterCount: len(text),
		Content:        text,
		URLOrPath:      filePath,
	}, nil
}

func readPDFFile(filePath string) (*TextDocument, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var content strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			return nil, err
		}
		content.WriteString(text)
	}

	text := content.String()
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &TextDocument{
		FileSize:       fileInfo.Size(),
		WordCount:      len(strings.Fields(text)),
		CharacterCount: len(text),
		Content:        text,
		URLOrPath:      filePath,
	}, nil
}

func readBlogFromURL(url string) (*TextDocument, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al parsear el HTML: %v", err)
	}

	title := doc.Find("title").Text()
	var content strings.Builder

	// Intentar extraer el contenido principal
	doc.Find("article, .content, .post-content, .entry-content").Each(func(i int, s *goquery.Selection) {
		content.WriteString(s.Text())
	})

	// Si no se encontró contenido, intentar con el body
	if content.Len() == 0 {
		content.WriteString(doc.Find("body").Text())
	}

	text := strings.TrimSpace(content.String())

	return &TextDocument{
		FileSize:       int64(len(text)),
		WordCount:      len(strings.Fields(text)),
		CharacterCount: len(text),
		Content:        text,
		Title:          title,
		URLOrPath:      url,
	}, nil
}

func readYouTubeVideo(videoURL string) (*TextDocument, error) {
	// Implementación simplificada. En la práctica, necesitarías usar
	// la API de YouTube o una biblioteca específica para Go.
	return nil, fmt.Errorf("YouTube video transcription not implemented")
}
