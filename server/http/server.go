package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/service"
)

const maxBodySize = 10 << 20 // 10 MB

const (
	errMethodNotAllowed   = "method not allowed"
	errBodyEmpty          = "request body is empty"
	errContentTypeMissing = "Content-Type header is required"
)

type Server struct {
	service service.ConfigAnalyzer
	port    string
	server  *http.Server
	logger  *slog.Logger
}

type IssueResponse struct {
	Severity    string `json:"severity"`
	Field       string `json:"field"`
	Description string `json:"description"`
	Advice      string `json:"advice"`
}

type AnalyzeResponse struct {
	Issues []IssueResponse `json:"issues,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(port string, svc service.ConfigAnalyzer) *Server {
	return &Server{
		service: svc,
		port:    port,
		logger:  slog.Default(),
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

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("HTTP server started", "addr", fmt.Sprintf("http://0.0.0.0:%s", s.port))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		s.logger.Info("received signal, shutting down...", "signal", sig)
	}

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
		s.writeJSONError(w, http.StatusMethodNotAllowed, errMethodNotAllowed)
		return
	}

	format, ok := s.validateRequest(w, r)
	if !ok {
		return
	}

	body, ok := s.readBody(w, r)
	if !ok {
		return
	}

	issues, err := s.service.Analyze(r.Context(), body, format)
	if err != nil {
		s.writeJSONError(w, mapHTTPCode(err), err.Error())
		return
	}

	response := AnalyzeResponse{
		Issues: make([]IssueResponse, len(issues)),
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

func (s *Server) validateRequest(w http.ResponseWriter, r *http.Request) (parser.Format, bool) {
	if r.ContentLength == 0 {
		s.writeJSONError(w, http.StatusBadRequest, errBodyEmpty)
		return "", false
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		s.writeJSONError(w, http.StatusBadRequest, errContentTypeMissing)
		return "", false
	}

	format, ok := parser.FormatFromContentType(contentType)
	if !ok {
		s.writeJSONError(w, http.StatusUnsupportedMediaType,
			fmt.Sprintf("unsupported Content-Type: %q", contentType))
		return "", false
	}

	return format, true
}

func (s *Server) readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	reader := http.MaxBytesReader(w, r.Body, maxBodySize)
	body, err := io.ReadAll(reader)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			s.writeJSONError(w, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("request body exceeds %d bytes limit", maxBodySize))
			return nil, false
		}
		s.writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return nil, false
	}
	defer r.Body.Close()

	if len(body) == 0 {
		s.writeJSONError(w, http.StatusBadRequest, errBodyEmpty)
		return nil, false
	}

	return body, true
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSONError(w, http.StatusMethodNotAllowed, errMethodNotAllowed)
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

func mapHTTPCode(err error) int {
	if errors.Is(err, service.ErrUnsupportedFormat) ||
		errors.Is(err, service.ErrParseFailed) ||
		errors.Is(err, service.ErrEmptyData) {
		return http.StatusBadRequest
	}
	if errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout
	}
	return http.StatusInternalServerError
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
