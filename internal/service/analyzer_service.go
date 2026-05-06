package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lopatich-privet/sec-config-scanner/internal/analyzer"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported configuration format")
	ErrParseFailed       = errors.New("failed to parse configuration")
	ErrEmptyData         = errors.New("configuration data is empty")
)

type ConfigAnalyzer interface {
	Analyze(ctx context.Context, data []byte, format parser.Format) ([]rules.Issue, error)
}

type AnalyzerService struct {
	analyzer *analyzer.Analyzer
}

func NewAnalyzerService(rulesList []rules.Rule) *AnalyzerService {
	return &AnalyzerService{
		analyzer: analyzer.NewAnalyzer(rulesList),
	}
}

func (s *AnalyzerService) Analyze(ctx context.Context, data []byte, format parser.Format) ([]rules.Issue, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	if format != parser.FormatJSON && format != parser.FormatYAML {
		return nil, ErrUnsupportedFormat
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, err := parser.Parse(data, format)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	issues := s.analyzer.Analyze(result)

	return issues, nil
}
