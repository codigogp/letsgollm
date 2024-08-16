package vector_storage

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gonum.org/v1/gonum/mat"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// SerializationFormat define el formato de serialización para la base de datos
type SerializationFormat int

const (
	Binary SerializationFormat = iota
	JSON
)

// VectorDatabase representa una base de datos de vectores
type VectorDatabase struct {
	DBFolder               string
	vectors                *mat.Dense
	metadata               []map[string]interface{}
	useSemanticConnections bool
	mu                     sync.RWMutex
}

// SimilarityResult representa el resultado de una búsqueda de similitud
type SimilarityResult struct {
	Metadata   map[string]interface{}
	Similarity float64
}

// NewVectorDatabase crea una nueva instancia de VectorDatabase
func NewVectorDatabase(dbFolder string, useSemanticConnections bool) *VectorDatabase {
	if _, err := os.Stat(dbFolder); os.IsNotExist(err) {
		os.MkdirAll(dbFolder, 0755)
	}
	return &VectorDatabase{
		DBFolder:               dbFolder,
		vectors:                nil, // Inicializamos como nil
		metadata:               []map[string]interface{}{},
		useSemanticConnections: useSemanticConnections,
	}
}

// LoadFromDisk carga la base de datos desde el disco
func (vdb *VectorDatabase) LoadFromDisk(collectionName string, format SerializationFormat) error {
	vdb.mu.Lock()
	defer vdb.mu.Unlock()

	filePath := filepath.Join(vdb.DBFolder, fmt.Sprintf("%s.svdb", collectionName))
	var err error

	switch format {
	case Binary:
		err = vdb.loadPickle(filePath)
	case JSON:
		err = vdb.loadJSON(filePath)
	default:
		return fmt.Errorf("unsupported serialization format")
	}

	if err != nil {
		return fmt.Errorf("error loading from disk: %v", err)
	}

	return nil
}

// SaveToDisk guarda la base de datos en el disco
func (vdb *VectorDatabase) SaveToDisk(collectionName string, format SerializationFormat) error {
	vdb.mu.RLock()
	defer vdb.mu.RUnlock()

	filePath := filepath.Join(vdb.DBFolder, fmt.Sprintf("%s.svdb", collectionName))
	var err error

	switch format {
	case Binary:
		err = vdb.savePickle(filePath)
	case JSON:
		err = vdb.saveJSON(filePath)
	default:
		return fmt.Errorf("unsupported serialization format")
	}

	if err != nil {
		return fmt.Errorf("error saving to disk: %v", err)
	}

	return nil
}

func (vdb *VectorDatabase) loadPickle(filePath string) error {
	// Implementación de carga de archivo binario
	// Nota: Go no tiene un equivalente directo de pickle, así que esto es una simplificación
	return fmt.Errorf("pickle loading not implemented in Go")
}

func (vdb *VectorDatabase) savePickle(filePath string) error {
	// Implementación de guardado de archivo binario
	return fmt.Errorf("pickle saving not implemented in Go")
}

func (vdb *VectorDatabase) loadJSON(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data struct {
		Vectors  [][]float64              `json:"vectors"`
		Metadata []map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	vdb.vectors = mat.NewDense(len(data.Vectors), len(data.Vectors[0]), nil)
	for i, vec := range data.Vectors {
		vdb.vectors.SetRow(i, vec)
	}
	vdb.metadata = data.Metadata

	return nil
}

func (vdb *VectorDatabase) saveJSON(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	rows, _ := vdb.vectors.Dims()
	vectors := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		vectors[i] = vdb.vectors.RawRowView(i)
	}

	data := struct {
		Vectors  [][]float64              `json:"vectors"`
		Metadata []map[string]interface{} `json:"metadata"`
	}{
		Vectors:  vectors,
		Metadata: vdb.metadata,
	}

	return json.NewEncoder(file).Encode(data)
}

// AddVector añade un vector a la base de datos
func (vdb *VectorDatabase) AddVector(chunkText string, embedding []float64, metadata map[string]interface{}, normalize bool) string {
	fmt.Println("Iniciando AddVector...")
	vdb.mu.Lock()
	fmt.Println("Lock adquirido en AddVector")

	if normalize {
		fmt.Println("Normalizando vector...")
		embedding = normalizeVector(embedding)
	}

	rows := 0
	if vdb.vectors != nil {
		rows, _ = vdb.vectors.Dims()
	}
	fmt.Printf("Número actual de vectores: %d\n", rows)

	if vdb.vectors == nil {
		fmt.Println("Creando nueva matriz de vectores...")
		vdb.vectors = mat.NewDense(1, len(embedding), embedding)
	} else {
		fmt.Println("Añadiendo vector a la matriz existente...")
		newVectors := mat.NewDense(rows+1, len(embedding), nil)
		newVectors.Slice(0, rows, 0, len(embedding)).(*mat.Dense).Copy(vdb.vectors)
		newVectors.SetRow(rows, embedding)
		vdb.vectors = newVectors
	}

	uniqueID := uuid.New().String()
	fmt.Printf("Generado ID único: %s\n", uniqueID)

	record := map[string]interface{}{
		"id":          uniqueID,
		"chunk_text":  chunkText,
		"embedding":   embedding,
		"metadata":    metadata,
		"connections": []interface{}{},
	}

	vdb.metadata = append(vdb.metadata, record)
	fmt.Printf("Metadata actualizado. Número total de registros: %d\n", len(vdb.metadata))

	vdb.mu.Unlock()
	fmt.Println("Lock liberado en AddVector")

	if vdb.useSemanticConnections && len(vdb.metadata) > 1 {
		fmt.Println("Actualizando conexiones semánticas...")
		vdb.updateConnections(len(vdb.metadata) - 1)
	} else {
		fmt.Println("No se actualizan conexiones semánticas para el primer vector.")
	}

	fmt.Println("AddVector completado")
	return uniqueID
}

// AddVectorsBatch añade un lote de vectores a la base de datos
func (vdb *VectorDatabase) AddVectorsBatch(records []map[string]interface{}, normalize bool) {
	vdb.mu.Lock()
	defer vdb.mu.Unlock()

	for _, record := range records {
		embedding := record["embedding"].([]float64)
		if normalize {
			embedding = normalizeVector(embedding)
		}

		rows, cols := vdb.vectors.Dims()
		if cols == 0 {
			vdb.vectors = mat.NewDense(1, len(embedding), embedding)
		} else {
			newVectors := mat.NewDense(rows+1, cols, nil)
			newVectors.Slice(0, rows, 0, cols).(*mat.Dense).Copy(vdb.vectors)
			newVectors.SetRow(rows, embedding)
			vdb.vectors = newVectors
		}

		metadata := map[string]interface{}{}
		for k, v := range record {
			if k != "embedding" {
				metadata[k] = v
			}
		}
		metadata["connections"] = []interface{}{}
		vdb.metadata = append(vdb.metadata, metadata)
	}

	if vdb.useSemanticConnections {
		for i := len(vdb.metadata) - len(records); i < len(vdb.metadata); i++ {
			vdb.updateConnections(i)
		}
	}
}

// UpdateVector actualiza un vector existente en la base de datos
func (vdb *VectorDatabase) UpdateVector(id string, newEmbedding []float64, newMetadata map[string]interface{}, normalize bool) error {
	vdb.mu.Lock()
	defer vdb.mu.Unlock()

	index := -1
	for i, meta := range vdb.metadata {
		if meta["id"] == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("vector with id %s not found", id)
	}

	if normalize {
		newEmbedding = normalizeVector(newEmbedding)
	}

	vdb.vectors.SetRow(index, newEmbedding)

	for k, v := range newMetadata {
		vdb.metadata[index][k] = v
	}

	if vdb.useSemanticConnections {
		vdb.updateConnections(index)
	}

	return nil
}

// DeleteVector elimina un vector de la base de datos
func (vdb *VectorDatabase) DeleteVector(id string) error {
	vdb.mu.Lock()
	defer vdb.mu.Unlock()

	index := -1
	for i, meta := range vdb.metadata {
		if meta["id"] == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("vector with id %s not found", id)
	}

	rows, cols := vdb.vectors.Dims()
	newVectors := mat.NewDense(rows-1, cols, nil)
	newVectors.Slice(0, index, 0, cols).(*mat.Dense).Copy(vdb.vectors.Slice(0, index, 0, cols))
	newVectors.Slice(index, rows-1, 0, cols).(*mat.Dense).Copy(vdb.vectors.Slice(index+1, rows, 0, cols))
	vdb.vectors = newVectors

	vdb.metadata = append(vdb.metadata[:index], vdb.metadata[index+1:]...)

	if vdb.useSemanticConnections {
		for i := range vdb.metadata {
			vdb.updateConnections(i)
		}
	}

	return nil
}

// GetConnectedChunks obtiene los chunks conectados a un vector dado
func (vdb *VectorDatabase) GetConnectedChunks(id string, depth int) ([]map[string]interface{}, error) {
	vdb.mu.RLock()
	defer vdb.mu.RUnlock()

	if !vdb.useSemanticConnections {
		return nil, fmt.Errorf("semantic connections are not enabled")
	}

	index := -1
	for i, meta := range vdb.metadata {
		if meta["id"] == id {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("vector with id %s not found", id)
	}

	visited := make(map[string]bool)
	result := []map[string]interface{}{}

	var dfs func(string, int)
	dfs = func(currentID string, currentDepth int) {
		if currentDepth > depth || visited[currentID] {
			return
		}
		visited[currentID] = true

		for _, meta := range vdb.metadata {
			if meta["id"] == currentID {
				result = append(result, meta)
				connections := meta["connections"].([]interface{})
				for _, conn := range connections {
					connMap := conn.(map[string]interface{})
					dfs(connMap["id"].(string), currentDepth+1)
				}
				break
			}
		}
	}

	dfs(id, 0)

	return result, nil
}

// SemanticSearch realiza una búsqueda semántica en la base de datos
func (vdb *VectorDatabase) SemanticSearch(queryEmbedding []float64, topK int, depth int) ([]map[string]interface{}, error) {
	vdb.mu.RLock()
	defer vdb.mu.RUnlock()

	if !vdb.useSemanticConnections {
		return nil, fmt.Errorf("semantic connections are not enabled")
	}

	initialResults := vdb.TopCosineSimilarity(queryEmbedding, topK)

	expandedResults := make(map[string]map[string]interface{})
	for _, result := range initialResults {
		id := result.Metadata["id"].(string)
		expandedResults[id] = result.Metadata

		connectedChunks, err := vdb.GetConnectedChunks(id, depth)
		if err != nil {
			return nil, err
		}

		for _, chunk := range connectedChunks {
			chunkID := chunk["id"].(string)
			if _, exists := expandedResults[chunkID]; !exists {
				expandedResults[chunkID] = chunk
			}
		}
	}

	finalResults := make([]map[string]interface{}, 0, len(expandedResults))
	for _, result := range expandedResults {
		finalResults = append(finalResults, result)
	}

	return finalResults, nil
}

// TopCosineSimilarity encuentra los vectores más similares

func (vdb *VectorDatabase) TopCosineSimilarity(targetVector []float64, topN int) []SimilarityResult {
	fmt.Println("Iniciando TopCosineSimilarity...")

	fmt.Println("Intentando adquirir RLock en TopCosineSimilarity...")
	vdb.mu.RLock()
	fmt.Println("RLock adquirido en TopCosineSimilarity")
	defer func() {
		vdb.mu.RUnlock()
		fmt.Println("RLock liberado en TopCosineSimilarity")
	}()

	if vdb.vectors == nil || vdb.vectors.IsEmpty() {
		fmt.Println("La base de datos está vacía.")
		return nil
	}

	rows, cols := vdb.vectors.Dims()
	fmt.Printf("Dimensiones de la matriz de vectores: %d x %d\n", rows, cols)
	fmt.Printf("Longitud del vector objetivo: %d\n", len(targetVector))

	if topN > rows {
		topN = rows
		fmt.Printf("Ajustando topN a %d\n", topN)
	}

	fmt.Printf("Vector objetivo: %v\n", targetVector)

	similarities := make([]SimilarityResult, 0, rows)
	for i := 0; i < rows; i++ {
		fmt.Printf("Procesando vector %d\n", i)
		vector := vdb.vectors.RawRowView(i)
		fmt.Printf("Vector %d: %v\n", i, vector)

		similarity, err := cosineSimilarity(targetVector, vector)
		if err != nil {
			fmt.Printf("Error al calcular similitud para vector %d: %v\n", i, err)
			continue
		}

		fmt.Printf("Similitud con vector %d: %f\n", i, similarity)
		similarities = append(similarities, SimilarityResult{
			Metadata:   vdb.metadata[i],
			Similarity: similarity,
		})
	}

	fmt.Println("Ordenando resultados...")
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	if topN > len(similarities) {
		topN = len(similarities)
		fmt.Printf("Ajustando topN final a %d\n", topN)
	}

	fmt.Println("TopCosineSimilarity completado")
	return similarities[:topN]
}

func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("los vectores tienen longitudes diferentes: %d vs %d", len(a), len(b))
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("uno de los vectores es un vector cero")
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

func (vdb *VectorDatabase) updateConnections(index int, topK ...int) {
	fmt.Println("Iniciando updateConnections...")
	k := 5
	if len(topK) > 0 {
		k = topK[0]
	}
	fmt.Printf("Actualizando conexiones para el índice %d con k=%d\n", index, k)

	fmt.Println("Intentando adquirir RLock en updateConnections...")
	vdb.mu.RLock()
	fmt.Println("RLock adquirido en updateConnections")
	defer func() {
		vdb.mu.RUnlock()
		fmt.Println("RLock liberado en updateConnections")
	}()

	if vdb.vectors == nil {
		fmt.Println("La matriz de vectores es nil, no se pueden actualizar las conexiones")
		return
	}

	rows, cols := vdb.vectors.Dims()
	fmt.Printf("Dimensiones de la matriz de vectores: %d x %d\n", rows, cols)

	if rows < 2 {
		fmt.Println("No hay suficientes vectores para actualizar conexiones. Se necesitan al menos 2 vectores.")
		return
	}

	fmt.Printf("Obteniendo vector para el índice %d\n", index)
	vector := vdb.vectors.RawRowView(index)
	if vector == nil {
		fmt.Printf("Error: No se pudo obtener el vector para el índice %d\n", index)
		return
	}
	fmt.Printf("Vector para el índice %d: %v\n", index, vector)

	fmt.Println("Calculando similitudes...")
	similarities := vdb.TopCosineSimilarity(vector, rows)
	if similarities == nil {
		fmt.Println("No se pudieron calcular las similitudes")
		return
	}
	fmt.Printf("Encontradas %d similitudes\n", len(similarities))

	fmt.Println("Actualizando conexiones para el vector actual...")
	connections := make([]map[string]interface{}, 0, k)
	for _, sim := range similarities {
		if sim.Metadata["id"] != vdb.metadata[index]["id"] {
			connections = append(connections, map[string]interface{}{
				"id":    sim.Metadata["id"],
				"score": sim.Similarity,
			})
			if len(connections) >= k {
				break
			}
		}
	}

	fmt.Println("Actualizando metadata...")
	vdb.mu.RUnlock()
	vdb.mu.Lock()
	vdb.metadata[index]["connections"] = connections
	vdb.mu.Unlock()
	vdb.mu.RLock()
	fmt.Printf("Actualizadas %d conexiones para el vector actual\n", len(connections))

	fmt.Println("Actualizando conexiones de otros vectores...")
	for i, meta := range vdb.metadata {
		if i == index {
			continue
		}
		fmt.Printf("Actualizando conexiones para el vector %d\n", i)
		otherConnections, ok := meta["connections"].([]map[string]interface{})
		if !ok {
			otherConnections = []map[string]interface{}{}
		}
		for _, sim := range similarities {
			if sim.Metadata["id"] == meta["id"] {
				otherConnections = append(otherConnections, map[string]interface{}{
					"id":    vdb.metadata[index]["id"],
					"score": sim.Similarity,
				})
				break
			}
		}
		sort.Slice(otherConnections, func(i, j int) bool {
			return otherConnections[i]["score"].(float64) > otherConnections[j]["score"].(float64)
		})
		if len(otherConnections) > k {
			otherConnections = otherConnections[:k]
		}

		vdb.mu.RUnlock()
		vdb.mu.Lock()
		meta["connections"] = otherConnections
		vdb.mu.Unlock()
		vdb.mu.RLock()

		fmt.Printf("Actualizadas conexiones para el vector %d\n", i)
	}
	fmt.Println("updateConnections completado")
}

// normalizeVector normaliza un vector
func normalizeVector(vector []float64) []float64 {
	fmt.Printf("Normalizando vector: %v\n", vector)
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		fmt.Println("Advertencia: El vector es un vector cero, no se puede normalizar")
		return vector
	}

	normalized := make([]float64, len(vector))
	for i, v := range vector {
		normalized[i] = v / norm
	}

	fmt.Printf("Vector normalizado: %v\n", normalized)
	return normalized
}

// cosineSimilarity calcula la similitud del coseno entre dos vectores
