package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gen "github.com/lopatich-privet/sec-config-scanner/api/gen"
)

const (
	serverAddr = "localhost:9090"
)

func main() {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := gen.NewAnalyzerServiceClient(conn)

	runTestFile(client, "bad.json", "json")
	runTestFile(client, "bad.yaml", "yaml")
	runTestFile(client, "safe.json", "json")
	runTestFile(client, "safe.yaml", "yaml")

	log.Println("\n=== All gRPC tests completed! ===")
}

func runTestFile(client gen.AnalyzerServiceClient, filename, format string) {
	log.Printf("=== Testing %s ===", filename)
	data, err := os.ReadFile("../../testdata/" + filename)
	if err != nil {
		log.Fatalf("failed to read %s: %v", filename, err)
	}

	resp, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: format,
		Data:   data,
	})
	if err != nil {
		log.Printf("Analyze %s failed: %v", filename, err)
		return
	}

	log.Printf("%s - Success: %v, Issues found: %d\n", filename, resp.Success, len(resp.Issues))
	for _, issue := range resp.Issues {
		log.Printf("  - %s: %s", issue.Severity, issue.Description)
	}
}
