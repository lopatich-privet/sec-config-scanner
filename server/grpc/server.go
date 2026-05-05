package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lopatich-privet/sec-config-scanner/api/gen"
	"github.com/lopatich-privet/sec-config-scanner/internal/analyzer"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	var config map[string]any

	switch req.Format {
	case "yaml", "yml":
		if err := yaml.Unmarshal(req.Data, &config); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse YAML: %v", err)
		}
	default:
		if err := json.Unmarshal(req.Data, &config); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse JSON: %v", err)
		}
	}

	issues := s.analyzer.Analyze(&parser.Config{Data: config})

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
		Issues: pbIssues,
	}, nil
}
