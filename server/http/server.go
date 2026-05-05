package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lopatich-privet/sec-config-scanner/internal/analyzer"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"

	"gopkg.in/yaml.v3"
)

type Server struct {
	analyzer *analyzer.Analyzer
	port     string
	server   *http.Server
	logger   *slog.Logger
}

type AnalyzeRequest struct {
	Data json.RawMessage `json:"data"`
}

type IssueResponse struct {
	Severity    string `json:"severity"`
	Field       string `json:"field"`
	Description string `json:"description"`
	Advice      string `json:"advice"`
}

type AnalyzeResponse struct {
	Success bool            `json:"success"`
	Issues  []IssueResponse `json:"issues,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(port string) *Server {
	return &Server{
		analyzer: analyzer.NewAnalyzer(rules.GetDefaultRules()),
		port:     port,
		logger:   slog.Default(),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", s.handleAnalyze)
	mux.HandleFunc("/health", s.handleHealth)

	handler := s.loggingMiddleware(s.recoveryMiddleware(mux))

	s.server = &http.Server{
		Addr:              ":" + s.port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		s.logger.Info("HTTP server started", "addr", fmt.Sprintf("http://0.0.0.0:%s", s.port))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	s.logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var config map[string]any

	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "yaml"), strings.Contains(contentType, "yml"):
		if err := yaml.Unmarshal(body, &config); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse YAML: %v", err))
			return
		}
	default:
		if err := json.Unmarshal(body, &config); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse JSON: %v", err))
			return
		}
	}

	issues := s.analyzer.Analyze(&parser.Config{Data: config})

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

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) writeJSONError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, ErrorResponse{Error: message})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).Round(time.Microsecond),
		)
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recovered", "error", rec, "path", r.URL.Path)
				s.writeJSONError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
