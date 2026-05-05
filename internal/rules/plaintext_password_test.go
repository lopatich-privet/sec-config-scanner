package rules

import (
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

func TestPlaintextPasswordRule_Check(t *testing.T) {
	rule := NewPlaintextPasswordRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
		wantFields []string
	}{
		{
			name: "plaintext password - keyword password",
			cfg: map[string]any{
				"database": map[string]any{
					"password": "secret123",
				},
			},
			wantIssues: 1,
			wantFields: []string{"database.password"},
		},
		{
			name: "plaintext password - keyword passwd",
			cfg: map[string]any{
				"db": map[string]any{
					"passwd": "mypass",
				},
			},
			wantIssues: 1,
			wantFields: []string{"db.passwd"},
		},
		{
			name: "plaintext password - keyword pwd",
			cfg: map[string]any{
				"auth": map[string]any{
					"pwd": "admin",
				},
			},
			wantIssues: 1,
			wantFields: []string{"auth.pwd"},
		},
		{
			name: "plaintext password - keyword secret",
			cfg: map[string]any{
				"api": map[string]any{
					"secret": "apikey123",
				},
			},
			wantIssues: 1,
			wantFields: []string{"api.secret"},
		},
		{
			name: "hash - MD5 (32 hex chars)",
			cfg: map[string]any{
				"password": "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
			},
			wantIssues: 0,
		},
		{
			name: "hash - SHA1 (40 hex chars)",
			cfg: map[string]any{
				"password": "356a192b7913b04c54574d18c28d46e6395428ab",
			},
			wantIssues: 0,
		},
		{
			name: "hash - SHA256 (64 hex chars)",
			cfg: map[string]any{
				"password": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			},
			wantIssues: 0,
		},
		{
			name: "hash - SHA512 (128 hex chars)",
			cfg: map[string]any{
				"password": "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
			},
			wantIssues: 0,
		},
		{
			name: "hash - bcrypt",
			cfg: map[string]any{
				"password": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			},
			wantIssues: 0,
		},
		{
			name: "env var - dollar sign prefix",
			cfg: map[string]any{
				"password": "$DB_PASSWORD",
			},
			wantIssues: 0,
		},
		{
			name: "empty password - no issue",
			cfg: map[string]any{
				"password": "",
			},
			wantIssues: 0,
		},
		{
			name: "non-string password - no issue",
			cfg: map[string]any{
				"password": 12345,
			},
			wantIssues: 0,
		},
		{
			name: "hash with invalid chars",
			cfg: map[string]any{
				"password": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaag",
			},
			wantIssues: 1,
		},
		{
			name: "non-hash string without keyword",
			cfg: map[string]any{
				"value": "somevalue",
			},
			wantIssues: 0,
		},
		{
			name: "env var with braces",
			cfg: map[string]any{
				"secret": "${API_SECRET}",
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(&parser.Config{Data: tt.cfg})
			if len(issues) != tt.wantIssues {
				t.Errorf("Check() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
			if tt.wantFields != nil && len(issues) > 0 {
				for i, wantField := range tt.wantFields {
					if i >= len(issues) {
						break
					}
					if issues[i].Field != wantField {
						t.Errorf("Check() issue %d field = %s, want %s", i, issues[i].Field, wantField)
					}
				}
			}
		})
	}
}

func TestPlaintextPasswordRule_Name(t *testing.T) {
	rule := NewPlaintextPasswordRule()
	if rule.Name() != "plaintext_password" {
		t.Errorf("Name() = %s, want plaintext_password", rule.Name())
	}
}

func TestPlaintextPasswordRule_Severity(t *testing.T) {
	rule := NewPlaintextPasswordRule()
	cfg := map[string]any{"password": "plaintext"}
	issues := rule.Check(&parser.Config{Data: cfg})
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != HIGH {
		t.Errorf("Expected HIGH severity, got %s", issues[0].Severity)
	}
}

func TestIsHash(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8", true},
		{"356a192b7913b04c54574d18c28d46e6395428ab", true},
		{"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", true},
		{"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", true},
		{"$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", true},
		{"$2b$12$XWN3YrHz6ZBZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQZQ", true},
		{"$ENV_VAR", true},
		{"${ENV_VAR}", true},
		{"$MY_PASSWORD_123", true},
		{"plaintext", false},
		{"", false},
		{"password123", false},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaag", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isHash(tt.input); got != tt.want {
				t.Errorf("isHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
