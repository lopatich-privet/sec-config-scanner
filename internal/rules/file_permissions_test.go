package rules

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

func TestFilePermissionRule_CheckFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission tests are not applicable on Windows")
	}

	rule := &FilePermissionRule{}

	tests := []struct {
		name         string
		perm         os.FileMode
		wantIssues   int
		wantSeverity Severity
	}{
		{
			name:         "world-writable 0666 -> HIGH",
			perm:         0666,
			wantIssues:   1,
			wantSeverity: HIGH,
		},
		{
			name:         "world-writable 0777 -> HIGH",
			perm:         0777,
			wantIssues:   1,
			wantSeverity: HIGH,
		},
		{
			name:         "group-writable 0766 -> HIGH",
			perm:         0766,
			wantIssues:   1,
			wantSeverity: HIGH,
		},
		{
			name:         "world-readable 0644 -> MEDIUM",
			perm:         0644,
			wantIssues:   1,
			wantSeverity: MEDIUM,
		},
		{
			name:         "world-readable 0755 -> MEDIUM",
			perm:         0755,
			wantIssues:   1,
			wantSeverity: MEDIUM,
		},
		{
			name:         "group-readable 0640 -> MEDIUM",
			perm:         0640,
			wantIssues:   1,
			wantSeverity: MEDIUM,
		},
		{
			name:       "secure 0600 -> no issues",
			perm:       0600,
			wantIssues: 0,
		},
		{
			name:       "secure 0400 -> no issues",
			perm:       0400,
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.conf")

			if err := os.WriteFile(tmpFile, []byte("test"), tt.perm); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			issues := rule.CheckFilePermissions(tmpFile)
			if len(issues) != tt.wantIssues {
				t.Errorf("CheckFilePermissions() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
			if tt.wantIssues > 0 && issues[0].Severity != tt.wantSeverity {
				t.Errorf("expected severity %s, got %s", tt.wantSeverity, issues[0].Severity)
			}
		})
	}
}

func TestFilePermissionRule_CheckFilePermissions_NonExistent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission tests are not applicable on Windows")
	}

	rule := &FilePermissionRule{}

	issues := rule.CheckFilePermissions("/nonexistent/path/to/file.conf")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for non-existent file, got %d", len(issues))
	}
	if issues[0].Severity != LOW {
		t.Errorf("expected LOW severity, got %s", issues[0].Severity)
	}
}

func TestFilePermissionRule_Check_WithPathValues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission tests are not applicable on Windows")
	}

	rule := NewFilePermissionRule()

	tmpDir := t.TempDir()

	worldWritableFile := filepath.Join(tmpDir, "writable.conf")
	if err := os.WriteFile(worldWritableFile, []byte("test"), 0666); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	worldReadableFile := filepath.Join(tmpDir, "readable.conf")
	if err := os.WriteFile(worldReadableFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cfg := &parser.Config{
		Data: map[string]any{
			"ssl": map[string]any{
				"cert_file": worldWritableFile,
				"key_file":  worldReadableFile,
			},
		},
	}

	issues := rule.Check(cfg)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	foundHigh := false
	foundMedium := false
	for _, issue := range issues {
		if issue.Severity == HIGH {
			foundHigh = true
		}
		if issue.Severity == MEDIUM {
			foundMedium = true
		}
	}

	if !foundHigh {
		t.Error("expected HIGH severity issue for world-writable file")
	}
	if !foundMedium {
		t.Error("expected MEDIUM severity issue for world-readable file")
	}
}

func TestFilePermissionRule_Check_NoPathValues(t *testing.T) {
	rule := NewFilePermissionRule()

	cfg := &parser.Config{
		Data: map[string]any{
			"server": map[string]any{
				"host": "localhost",
			},
		},
	}

	issues := rule.Check(cfg)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestFilePermissionRule_Name(t *testing.T) {
	rule := NewFilePermissionRule()
	if rule.Name() != "file_permissions" {
		t.Errorf("Name() = %s, want file_permissions", rule.Name())
	}
}
