package parser

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirectory_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "JSON config",
			filename: "app.json",
			content:  `{"server": {"host": "0.0.0.0"}}`,
		},
		{
			name:     "YAML config",
			filename: "app.yaml",
			content:  "server:\n  host: 0.0.0.0\n",
		},
		{
			name:     "YML config",
			filename: "app.yml",
			content:  "server:\n  host: 0.0.0.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			configs, err := ParseDirectory(context.Background(), tmpDir)
			if err != nil {
				t.Fatalf("ParseDirectory() error: %v", err)
			}

			found := false
			for _, cfg := range configs {
				if cfg.FilePath == filePath {
					found = true
					if cfg.Data == nil {
						t.Error("Config.Data is nil")
					}
					break
				}
			}
			if !found {
				t.Errorf("config file %s not found in results", tt.filename)
			}
		})
	}
}

func TestParseDirectory_NestedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	nestedDir := filepath.Join(tmpDir, "subdir", "deep")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	jsonContent := `{"tls": {"enabled": false}}`
	yamlContent := "tls:\n  enabled: false\n"

	if err := os.WriteFile(filepath.Join(tmpDir, "root.json"), []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "nested.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	configs, err := ParseDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ParseDirectory() error: %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}
}

func TestParseDirectory_SkipsNonConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a config"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	_, err := ParseDirectory(context.Background(), tmpDir)
	if err == nil {
		t.Fatal("expected error for directory with no config files")
	}
}

func TestParseDirectory_InvalidConfigSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	invalidYAML := []byte(":\n  :\n  bad: [")
	validJSON := []byte(`{"server": {"host": "localhost"}}`)

	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), invalidYAML, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.json"), validJSON, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	configs, err := ParseDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ParseDirectory() error: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("expected 1 valid config, got %d", len(configs))
	}
}

func TestParseDirectory_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := ParseDirectory(context.Background(), tmpDir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func TestParseDirectory_NonExistentDirectory(t *testing.T) {
	_, err := ParseDirectory(context.Background(), "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestParseDirectory_CanceledContext(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(`{"a": 1}`), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ParseDirectory(ctx, tmpDir)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestParseDirectory_FilePathSet(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(filePath, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	configs, err := ParseDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ParseDirectory() error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].FilePath != filePath {
		t.Errorf("FilePath = %s, want %s", configs[0].FilePath, filePath)
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"app.json", true},
		{"app.yaml", true},
		{"app.yml", true},
		{"app.JSON", true},
		{"app.YAML", true},
		{"app.YML", true},
		{"app.txt", false},
		{"app.conf", false},
		{"app.toml", false},
		{"Makefile", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isConfigFile(tt.path)
			if got != tt.want {
				t.Errorf("isConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
