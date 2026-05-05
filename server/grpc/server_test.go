package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/api/gen"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
	"github.com/lopatich-privet/sec-config-scanner/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type mockAnalyzer struct {
	issues []rules.Issue
	err    error
}

func (m *mockAnalyzer) Analyze(_ context.Context, _ []byte, _ parser.Format) ([]rules.Issue, error) {
	return m.issues, m.err
}

func setupGRPCClient(t *testing.T, mock service.ConfigAnalyzer) (gen.AnalyzerServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)

	srv := &Server{
		service: mock,
		port:    "0",
	}

	grpcSrv := grpc.NewServer()
	gen.RegisterAnalyzerServiceServer(grpcSrv, srv)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("grpc server error: %v", err)
		}
	}()

	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	client := gen.NewAnalyzerServiceClient(conn)
	cleanup := func() {
		conn.Close()
		grpcSrv.Stop()
	}

	return client, cleanup
}

func TestGRPC_Analyze_Success(t *testing.T) {
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

	client, cleanup := setupGRPCClient(t, mock)
	defer cleanup()

	tests := []struct {
		name       string
		format     string
		data       []byte
		wantIssues int
	}{
		{
			name:       "valid JSON",
			format:     "json",
			data:       []byte(`{"tls": {"enabled": false}}`),
			wantIssues: 1,
		},
		{
			name:       "valid YAML",
			format:     "yaml",
			data:       []byte("tls:\n  enabled: false\n"),
			wantIssues: 1,
		},
		{
			name:       "no issues found",
			format:     "json",
			data:       []byte(`{"server": {"host": "localhost"}}`),
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
				Format: tt.format,
				Data:   tt.data,
			})
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}
			if len(resp.Issues) != tt.wantIssues {
				t.Errorf("got %d issues, want %d", len(resp.Issues), tt.wantIssues)
			}
		})
	}
}

func TestGRPC_Analyze_EmptyData(t *testing.T) {
	mock := &mockAnalyzer{}
	client, cleanup := setupGRPCClient(t, mock)
	defer cleanup()

	_, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "json",
		Data:   []byte{},
	})

	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}

func TestGRPC_Analyze_InvalidFormat(t *testing.T) {
	mock := &mockAnalyzer{}
	client, cleanup := setupGRPCClient(t, mock)
	defer cleanup()

	_, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "toml",
		Data:   []byte(`{"a": 1}`),
	})

	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}

func TestGRPC_Analyze_ServiceErrors(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
	}{
		{
			name:    "ErrParseFailed",
			mockErr: service.ErrParseFailed,
		},
		{
			name:    "ErrEmptyData",
			mockErr: service.ErrEmptyData,
		},
		{
			name:    "ErrUnsupportedFormat",
			mockErr: service.ErrUnsupportedFormat,
		},
		{
			name:    "context.Canceled",
			mockErr: context.Canceled,
		},
		{
			name:    "unknown error",
			mockErr: errors.New("something unexpected"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAnalyzer{err: tt.mockErr}
			client, cleanup := setupGRPCClient(t, mock)
			defer cleanup()

			_, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
				Format: "json",
				Data:   []byte(`{"test": true}`),
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestGRPC_Analyze_EmptyResult(t *testing.T) {
	mock := &mockAnalyzer{issues: nil}
	client, cleanup := setupGRPCClient(t, mock)
	defer cleanup()

	resp, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "json",
		Data:   []byte(`{"server": {"host": "localhost"}}`),
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if len(resp.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(resp.Issues))
	}
}
