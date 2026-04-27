package grpc

import (
	"config-analyzer/api/gen"
	"config-analyzer/internal/analyzer"
	"config-analyzer/internal/rules"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type Server struct {
	gen.UnimplementedAnalyzerServiceServer
	analyzer *analyzer.Analyzer
	logger   *slog.Logger
	port     string
}

func NewServer(port string) *Server {
	return &Server{
		analyzer: analyzer.NewAnalyzer(rules.GetDefaultRules()),
		logger:   slog.Default(),
		port:     port,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterAnalyzerServiceServer(grpcServer, s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		s.logger.Info("gRPC server started", "addr", fmt.Sprintf("0.0.0.0:%s", s.port))
		if err := grpcServer.Serve(listener); err != nil {
			s.logger.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	s.logger.Info("shutting down gRPC server...")
	grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")

	return nil
}

func (s *Server) Analyze(ctx context.Context, req *gen.AnalyzeRequest) (*gen.AnalyzeResponse, error) {
	var config map[string]any

	switch req.Format {
	case "yaml", "yml":
		if err := yaml.Unmarshal(req.Data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		if err := json.Unmarshal(req.Data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	issues := s.analyzer.Analyze(config)

	pbIssues := make([]*gen.Issue, len(issues))
	for i, issue := range issues {
		pbIssues[i] = &gen.Issue{
			Severity:    string(issue.Severity),
			Field:       issue.Field,
			Description: issue.Description,
			Advice:      issue.Advice,
		}
	}

	return &gen.AnalyzeResponse{
		Success: len(issues) == 0,
		Issues:  pbIssues,
	}, nil
}
