package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

type Config struct {
	Data     map[string]any
	FilePath string
}

func ParseFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var format Format

	switch ext {
	case ".json":
		format = FormatJSON
	case ".yaml", ".yml":
		format = FormatYAML
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	result, err := Parse(data, format)
	if err != nil {
		return nil, err
	}

	result.FilePath = path
	return result, nil
}

func Parse(data []byte, format Format) (*Config, error) {
	var result map[string]any

	switch format {
	case FormatJSON:
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case FormatYAML:
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return &Config{Data: result}, nil
}

func ParseFromStdin() (*Config, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Пытаемся сначала JSON, потом YAML
	if config, err := Parse(data, FormatJSON); err == nil {
		return config, nil
	}

	return Parse(data, FormatYAML)
}
