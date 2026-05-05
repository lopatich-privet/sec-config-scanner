package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
	"github.com/lopatich-privet/sec-config-scanner/internal/service"
)

type mockAnalyzer struct {
	issues []rules.Issue
	err    error
}

func (m *mockAnalyzer) Analyze(_ context.Context, _ []byte, _ parser.Format) ([]rules.Issue, error) {
	return m.issues, m.err
}

func newTestServer(mock service.ConfigAnalyzer) *Server {
	return &Server{
		service: mock,
		port:    "0",
	}
}

func TestHandleAnalyze_Success(t *testing.T) {
	mock := &mockAnalyzer{
		issues: []rules.Issue{
			{
				Severity:    rules.HIGH,
				Field:       "tls.enabled",
				Description: "TLS disabled",
				Advice:      "Enable TLS.",
			},
		},
	}
	srv := newTestServer(mock)

	tests := []struct {
		name         string
		method       string
		contentType  string
		body         string
		wantStatus   int
		wantInBody   string
	}{
		{
			name:        "valid JSON request",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"tls": {"enabled": false}}`,
			wantStatus:  http.StatusOK,
			wantInBody:  `"issues"`,
		},
		{
			name:        "valid YAML request",
			method:      http.MethodPost,
			contentType: "application/yaml",
			body:        "tls:\n  enabled: false\n",
			wantStatus:  http.StatusOK,
			wantInBody:  `"issues"`,
		},
		{
			name:        "valid JSON with charset",
			method:      http.MethodPost,
			contentType: "application/json; charset=UTF-8",
			body:        `{"tls": {"enabled": false}}`,
			wantStatus:  http.StatusOK,
			wantInBody:  `"issues"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/analyze", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.ContentLength = int64(len(tt.body))

			w := httptest.NewRecorder()
			srv.handleAnalyze(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			body, _ := io.ReadAll(resp.Body)
			if !strings.Contains(string(body), tt.wantInBody) {
				t.Errorf("body = %s, want to contain %q", string(body), tt.wantInBody)
			}
		})
	}
}

func TestHandleAnalyze_MethodNotAllowed(t *testing.T) {
	mock := &mockAnalyzer{}
	srv := newTestServer(mock)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/analyze", nil)
			w := httptest.NewRecorder()
			srv.handleAnalyze(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHandleAnalyze_EmptyBody(t *testing.T) {
	mock := &mockAnalyzer{}
	srv := newTestServer(mock)

	req := httptest.NewRequest(http.MethodPost, "/analyze", nil)
	req.ContentLength = 0
	w := httptest.NewRecorder()
	srv.handleAnalyze(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleAnalyze_MissingContentType(t *testing.T) {
	mock := &mockAnalyzer{}
	srv := newTestServer(mock)

	body := `{"test": true}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	srv.handleAnalyze(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	respBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), "Content-Type") {
		t.Errorf("body = %s, want to contain 'Content-Type'", string(respBody))
	}
}

func TestHandleAnalyze_UnsupportedContentType(t *testing.T) {
	mock := &mockAnalyzer{}
	srv := newTestServer(mock)

	body := `{"test": true}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	srv.handleAnalyze(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestHandleAnalyze_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "ErrParseFailed -> 400",
			mockErr:    service.ErrParseFailed,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrEmptyData -> 400",
			mockErr:    service.ErrEmptyData,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrUnsupportedFormat -> 400",
			mockErr:    service.ErrUnsupportedFormat,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "context.Canceled -> 408",
			mockErr:    context.Canceled,
			wantStatus: http.StatusRequestTimeout,
		},
		{
			name:       "unknown error -> 500",
			mockErr:    errors.New("something unexpected"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAnalyzer{err: tt.mockErr}
			srv := newTestServer(mock)

			body := `{"test": true}`
			req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.ContentLength = int64(len(body))
			w := httptest.NewRecorder()
			srv.handleAnalyze(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHandleAnalyze_EmptyResult(t *testing.T) {
	mock := &mockAnalyzer{issues: nil}
	srv := newTestServer(mock)

	body := `{"server": {"host": "localhost"}}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	srv.handleAnalyze(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var respBody AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(respBody.Issues) != 0 {
		t.Errorf("expected empty issues, got %d", len(respBody.Issues))
	}
}

func TestHandleHealth(t *testing.T) {
	mock := &mockAnalyzer{}
	srv := newTestServer(mock)

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{
			name:       "GET /health",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST /health -> method not allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()
			srv.handleHealth(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
