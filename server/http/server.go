package http

import (
	"config-analyzer/internal/analyzer"
	"config-analyzer/internal/output"
	"config-analyzer/internal/rules"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	analyzer *analyzer.Analyzer
	port     string
}

type AnalyzeRequest struct {
	Format string          `json:"format"` // "json" or "yaml"
	Data   json.RawMessage `json:"data"`
}

type AnalyzeResponse struct {
	Success bool                   `json:"success"`
	Issues  []output.IssueResponse `json:"issues"`
}

func NewServer(port string) *Server {
	return &Server{
		analyzer: analyzer.NewAnalyzer(rules.GetDefaultRules()),
		port:     port,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/analyze", s.handleAnalyze)
	http.HandleFunc("/health", s.handleHealth)

	fmt.Printf("HTTP сервер запущен на http://0.0.0.0:%s\n", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Парсим JSON конфиг напрямую
	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Анализируем конфиг
	issues := s.analyzer.Analyze(config)

	// Формируем ответ
	response := AnalyzeResponse{
		Success: len(issues) == 0,
		Issues:  make([]output.IssueResponse, len(issues)),
	}

	for i, issue := range issues {
		response.Issues[i] = output.IssueResponse{
			Severity:    string(issue.Severity),
			Field:       issue.Field,
			Description: issue.Description,
			Advice:      issue.Advice,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
