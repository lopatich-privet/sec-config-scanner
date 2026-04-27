package http

import (
	"config-analyzer/internal/analyzer"
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
	Format string          `json:"format"`
	Data   json.RawMessage `json:"data"`
}

type IssueResponse struct {
	Severity    string `json:"severity"`
	Field       string `json:"field"`
	Description string `json:"description"`
	Advice      string `json:"advice"`
}

type AnalyzeResponse struct {
	Success bool            `json:"success"`
	Issues  []IssueResponse `json:"issues"`
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to read request body"})
		return
	}

	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to parse JSON: %v", err)})
		return
	}

	issues := s.analyzer.Analyze(config)

	response := AnalyzeResponse{
		Success: len(issues) == 0,
		Issues:  make([]IssueResponse, len(issues)),
	}

	for i, issue := range issues {
		response.Issues[i] = IssueResponse{
			Severity:    string(issue.Severity),
			Field:       issue.Field,
			Description: issue.Description,
			Advice:      issue.Advice,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		fmt.Printf("error writing health response: %v\n", err)
	}
}
