package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lopatich-privet/sec-config-scanner/api/gen"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
	"github.com/lopatich-privet/sec-config-scanner/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	gen.UnimplementedAnalyzerServiceServer
	service service.ConfigAnalyzer
	logger  *slog.Logger
	port    string
}

func NewServer(port string) *Server {
	return &Server{
		service: service.NewAnalyzerService(rules.GetDefaultRules()),
		logger:  slog.Default(),
		port:    port,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterAnalyzerServiceServer(grpcServer, s)

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("gRPC server started", "addr", "0.0.0.0:"+s.port)
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
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

	s.logger.Info("shutting down gRPC server...")
	grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")

	return nil
}

func (s *Server) Analyze(ctx context.Context, req *gen.AnalyzeRequest) (*gen.AnalyzeResponse, error) {
	issues, err := s.service.Analyze(ctx, req.Data, parser.Format(req.Format))
	if err != nil {
		return nil, status.Errorf(mapToGRPCCode(err), "%v", err)
	}

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

func mapToGRPCCode(err error) codes.Code {
	if errors.Is(err, service.ErrUnsupportedFormat) ||
		errors.Is(err, service.ErrParseFailed) ||
		errors.Is(err, service.ErrEmptyData) {
		return codes.InvalidArgument
	}
	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	}
	return codes.Internal
}
