package main

import (
	"fmt"
	"time"

	"github.com/codigogp/letsgollm/internal/tools/vector_storage"
)

func main() {
	fmt.Println("Iniciando programa...")

	// Crear una nueva instancia de VectorDatabase
	vdb := vector_storage.NewVectorDatabase("./db", true)
	fmt.Println("Base de datos creada.")

	// Añadir algunos vectores
	fmt.Println("Añadiendo vectores...")
	start := time.Now()
	id1 := vdb.AddVector("Ejemplo 1", []float64{1, 2, 3}, map[string]interface{}{"tag": "test"}, true)
	duration := time.Since(start)
	fmt.Printf("Primer vector añadido con ID: %s en %v\n", id1, duration)

	start = time.Now()
	id2 := vdb.AddVector("Ejemplo 2", []float64{4, 5, 6}, map[string]interface{}{"tag": "test"}, true)
	duration = time.Since(start)
	fmt.Printf("Segundo vector añadido con ID: %s en %v\n", id2, duration)

	// Realizar una búsqueda de similitud
	fmt.Println("Realizando búsqueda de similitud...")
	start = time.Now()
	results := vdb.TopCosineSimilarity([]float64{1, 2, 3}, 2)
	duration = time.Since(start)
	fmt.Printf("Búsqueda completada en %v\n", duration)

	if results != nil {
		for _, result := range results {
			fmt.Printf("ID: %s, Similitud: %f\n", result.Metadata["id"], result.Similarity)
		}
	} else {
		fmt.Println("No se encontraron resultados.")
	}

	fmt.Println("Programa finalizado.")
}
