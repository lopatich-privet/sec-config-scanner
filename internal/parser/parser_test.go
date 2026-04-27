package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		format      Format
		wantNil     bool
		wantDataLen int
		wantErr     bool
	}{
		{
			name:        "valid JSON",
			data:        []byte(`{"key":"value","nested":{"bool":true}}`),
			format:      FormatJSON,
			wantNil:     false,
			wantDataLen: 2,
			wantErr:     false,
		},
		{
			name:        "valid YAML",
			data:        []byte("key: value\nnested:\n  bool: true"),
			format:      FormatYAML,
			wantNil:     false,
			wantDataLen: 2,
			wantErr:     false,
		},
		{
			name:        "invalid JSON",
			data:        []byte(`{invalid json`),
			format:      FormatJSON,
			wantNil:     true,
			wantDataLen: 0,
			wantErr:     true,
		},
		{
			name:        "invalid YAML",
			data:        []byte(`invalid: yaml: unclosed`),
			format:      FormatYAML,
			wantNil:     true,
			wantDataLen: 0,
			wantErr:     true,
		},
		{
			name:        "unsupported format",
			data:        []byte(`{}`),
			format:      Format("xml"),
			wantNil:     true,
			wantDataLen: 0,
			wantErr:     true,
		},
		{
			name:        "empty JSON object",
			data:        []byte(`{}`),
			format:      FormatJSON,
			wantNil:     false,
			wantDataLen: 0,
			wantErr:     false,
		},
		{
			name:        "empty YAML object",
			data:        []byte(`{}`),
			format:      FormatYAML,
			wantNil:     false,
			wantDataLen: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.data, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("Parse() got = %v, wantNil %v", got, tt.wantNil)
				return
			}
			if got != nil && len(got.Data) != tt.wantDataLen {
				t.Errorf("Parse() got.Data length = %v, want %v", len(got.Data), tt.wantDataLen)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name        string
		ext         string
		content     string
		wantNil     bool
		wantErr     bool
	}{
		{
			name:    "valid JSON file",
			ext:     ".json",
			content: `{"test":"value"}`,
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "valid YAML file (.yaml)",
			ext:     ".yaml",
			content: `test: value`,
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "valid YAML file (.yml)",
			ext:     ".yml",
			content: `test: value`,
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "unsupported extension",
			ext:     ".xml",
			content: `<test></test>`,
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "file does not exist",
			ext:     ".json",
			content: "",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != "" {
				tmpDir := t.TempDir()
				path = filepath.Join(tmpDir, "test"+tt.ext)
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			} else {
				path = "nonexistent_file.json"
			}

			got, err := ParseFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("ParseFile() got = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

func TestParseFromStdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantNil     bool
		wantErr     bool
		wantDataKey string
	}{
		{
			name:        "JSON from stdin",
			input:       `{"stdin_key":"stdin_value"}`,
			wantNil:     false,
			wantErr:     false,
			wantDataKey: "stdin_key",
		},
		{
			name:        "YAML from stdin",
			input:       `yaml_key: yaml_value`,
			wantNil:     false,
			wantErr:     false,
			wantDataKey: "yaml_key",
		},
		{
			name:        "empty input",
			input:       "",
			wantNil:     false,
			wantErr:     false,
			wantDataKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			defer r.Close()
			defer w.Close()

			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			os.Stdin = r

			go func() {
				w.WriteString(tt.input)
				w.Close()
			}()

			got, err := ParseFromStdin()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromStdin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("ParseFromStdin() got = %v, wantNil %v", got, tt.wantNil)
				return
			}
			if tt.wantDataKey != "" {
				if _, ok := got.Data[tt.wantDataKey]; !ok {
					t.Errorf("ParseFromStdin() expected key %s not found", tt.wantDataKey)
				}
			}
		})
	}
}

func TestFormatDetection(t *testing.T) {
	tests := []struct {
		ext     string
		wantFmt Format
	}{
		{".json", FormatJSON},
		{".JSON", FormatJSON},
		{".yaml", FormatYAML},
		{".YAML", FormatYAML},
		{".yml", FormatYAML},
		{".YML", FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test"+tt.ext)
			os.WriteFile(path, []byte(`{"key":"value"}`), 0644)

			_, err := ParseFile(path)
			if err != nil && strings.Contains(err.Error(), "unsupported") {
				t.Errorf("Extension %s not recognized", tt.ext)
			}
		})
	}
}
