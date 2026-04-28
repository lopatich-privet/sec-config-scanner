package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gen "config-analyzer/api/gen"
)

func main() {
	// Подключение к gRPC серверу
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := gen.NewAnalyzerServiceClient(conn)

	// --- Тестирование bad.json ---
	log.Println("=== Testing bad.json ===")
	dataJsonBad, err := os.ReadFile("testdata/bad.json")
	if err != nil {
		log.Fatalf("failed to read bad.json: %v", err)
	}
	respJsonBad, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "json",
		Data:   dataJsonBad,
	})
	if err != nil {
		log.Printf("Analyze bad.json failed: %v", err)
	} else {
		log.Printf("bad.json - Success: %v, Issues found: %d\n", respJsonBad.Success, len(respJsonBad.Issues))
		for _, issue := range respJsonBad.Issues {
			log.Printf("  - %s: %s", issue.Severity, issue.Description)
		}
	}

	// --- Тестирование bad.yaml ---
	log.Println("\n=== Testing bad.yaml ===")
	dataYamlBad, err := os.ReadFile("testdata/bad.yaml")
	if err != nil {
		log.Fatalf("failed to read bad.yaml: %v", err)
	}
	respYamlBad, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "yaml",
		Data:   dataYamlBad,
	})
	if err != nil {
		log.Printf("Analyze bad.yaml failed: %v", err)
	} else {
		log.Printf("bad.yaml - Success: %v, Issues found: %d\n", respYamlBad.Success, len(respYamlBad.Issues))
		for _, issue := range respYamlBad.Issues {
			log.Printf("  - %s: %s", issue.Severity, issue.Description)
		}
	}

	// --- Тестирование safe.json ---
	log.Println("\n=== Testing safe.json ===")
	dataJsonSafe, err := os.ReadFile("testdata/safe.json")
	if err != nil {
		log.Fatalf("failed to read safe.json: %v", err)
	}
	respJsonSafe, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "json",
		Data:   dataJsonSafe,
	})
	if err != nil {
		log.Printf("Analyze safe.json failed: %v", err)
	} else {
		log.Printf("safe.json - Success: %v, Issues found: %d\n", respJsonSafe.Success, len(respJsonSafe.Issues))
	}

	// --- Тестирование safe.yaml ---
	log.Println("\n=== Testing safe.yaml ===")
	dataYamlSafe, err := os.ReadFile("testdata/safe.yaml")
	if err != nil {
		log.Fatalf("failed to read safe.yaml: %v", err)
	}
	respYamlSafe, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "yaml",
		Data:   dataYamlSafe,
	})
	if err != nil {
		log.Printf("Analyze safe.yaml failed: %v", err)
	} else {
		log.Printf("safe.yaml - Success: %v, Issues found: %d\n", respYamlSafe.Success, len(respYamlSafe.Issues))
	}

	log.Println("\n=== All gRPC tests completed! ===")
}
