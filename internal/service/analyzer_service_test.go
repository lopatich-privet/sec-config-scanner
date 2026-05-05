package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
)

func newTestService() *AnalyzerService {
	return NewAnalyzerService(rules.GetDefaultRules())
}

func TestAnalyzerService_Analyze_Success(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name       string
		data       []byte
		format     parser.Format
		wantIssues int
	}{
		{
			name: "valid JSON with TLS disabled",
			data: []byte(`{"tls": {"enabled": false}}`),
			format:     parser.FormatJSON,
			wantIssues: 1,
		},
		{
			name:       "valid JSON with no issues",
			data:       []byte(`{"server": {"host": "localhost"}}`),
			format:     parser.FormatJSON,
			wantIssues: 0,
		},
		{
			name:       "valid YAML with debug log",
			data:       []byte("log:\n  level: debug\n"),
			format:     parser.FormatYAML,
			wantIssues: 1,
		},
		{
			name:       "valid YAML with no issues",
			data:       []byte("server:\n  host: localhost\n"),
			format:     parser.FormatYAML,
			wantIssues: 0,
		},
		{
			name: "valid JSON with plaintext password",
			data: []byte(`{"database": {"password": "mysecret123"}}`),
			format:     parser.FormatJSON,
			wantIssues: 1,
		},
		{
			name: "valid JSON with bind all interfaces",
			data: []byte(`{"server": {"host": "0.0.0.0"}}`),
			format:     parser.FormatJSON,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, err := svc.Analyze(context.Background(), tt.data, tt.format)
			if err != nil {
				t.Fatalf("Analyze() returned unexpected error: %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Analyze() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestAnalyzerService_Analyze_EmptyData(t *testing.T) {
	svc := newTestService()

	_, err := svc.Analyze(context.Background(), []byte{}, parser.FormatJSON)
	if !errors.Is(err, ErrEmptyData) {
		t.Errorf("Analyze() error = %v, want ErrEmptyData", err)
	}
}

func TestAnalyzerService_Analyze_NilData(t *testing.T) {
	svc := newTestService()

	_, err := svc.Analyze(context.Background(), nil, parser.FormatJSON)
	if !errors.Is(err, ErrEmptyData) {
		t.Errorf("Analyze() error = %v, want ErrEmptyData", err)
	}
}

func TestAnalyzerService_Analyze_UnsupportedFormat(t *testing.T) {
	svc := newTestService()

	_, err := svc.Analyze(context.Background(), []byte(`{"a": 1}`), "toml")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Analyze() error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestAnalyzerService_Analyze_InvalidJSON(t *testing.T) {
	svc := newTestService()

	_, err := svc.Analyze(context.Background(), []byte(`{invalid`), parser.FormatJSON)
	if !errors.Is(err, ErrParseFailed) {
		t.Errorf("Analyze() error = %v, want ErrParseFailed", err)
	}
}

func TestAnalyzerService_Analyze_InvalidYAML(t *testing.T) {
	svc := newTestService()

	_, err := svc.Analyze(context.Background(), []byte(":\n  :\n  bad"), parser.FormatYAML)
	if !errors.Is(err, ErrParseFailed) {
		t.Errorf("Analyze() error = %v, want ErrParseFailed", err)
	}
}

func TestAnalyzerService_Analyze_CanceledContext(t *testing.T) {
	svc := newTestService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Analyze(ctx, []byte(`{"a": 1}`), parser.FormatJSON)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Analyze() error = %v, want context.Canceled", err)
	}
}
