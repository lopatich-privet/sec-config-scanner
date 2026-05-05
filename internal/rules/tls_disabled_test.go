package rules

import (
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

func TestTLSDisabledRule_Check(t *testing.T) {
	rule := NewTLSDisabledRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
		wantFields []string
	}{
		{
			name: "insecure_skip_verify: true",
			cfg: map[string]any{
				"tls": map[string]any{
					"insecure_skip_verify": true,
				},
			},
			wantIssues: 1,
			wantFields: []string{"tls.insecure_skip_verify"},
		},
		{
			name: "tls.enabled: false",
			cfg: map[string]any{
				"tls": map[string]any{
					"enabled": false,
				},
			},
			wantIssues: 1,
			wantFields: []string{"tls.enabled"},
		},
		{
			name: "server.tls.enabled: false",
			cfg: map[string]any{
				"server": map[string]any{
					"tls": map[string]any{
						"enabled": false,
					},
				},
			},
			wantIssues: 1,
			wantFields: []string{"server.tls.enabled"},
		},
		{
			name: "insecure_skip_verify: true - nested",
			cfg: map[string]any{
				"database": map[string]any{
					"ssl": map[string]any{
						"insecure_skip_verify": true,
					},
				},
			},
			wantIssues: 1,
			wantFields: []string{"database.ssl.insecure_skip_verify"},
		},
		{
			name: "tls.enabled: true - no issue",
			cfg: map[string]any{
				"tls": map[string]any{
					"enabled": true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "insecure_skip_verify: false - no issue",
			cfg: map[string]any{
				"tls": map[string]any{
					"insecure_skip_verify": false,
				},
			},
			wantIssues: 0,
		},
		{
			name:       "no TLS config - no issue",
			cfg:        map[string]any{},
			wantIssues: 0,
		},
		{
			name: "non-boolean enabled field - no issue",
			cfg: map[string]any{
				"tls": map[string]any{
					"enabled": "true",
				},
			},
			wantIssues: 0,
		},
		{
			name: "non-boolean insecure_skip_verify field - no issue",
			cfg: map[string]any{
				"tls": map[string]any{
					"insecure_skip_verify": "yes",
				},
			},
			wantIssues: 0,
		},
		{
			name: "cache.enabled: false - no issue (no TLS context)",
			cfg: map[string]any{
				"cache": map[string]any{
					"enabled": false,
				},
			},
			wantIssues: 0,
		},
		{
			name: "feature.enabled: false - no issue (no TLS context)",
			cfg: map[string]any{
				"feature": map[string]any{
					"enabled": false,
				},
			},
			wantIssues: 0,
		},
		{
			name: "multiple issues",
			cfg: map[string]any{
				"tls": map[string]any{
					"enabled":              false,
					"insecure_skip_verify": true,
				},
			},
			wantIssues: 2,
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

func TestTLSDisabledRule_Name(t *testing.T) {
	rule := NewTLSDisabledRule()
	if rule.Name() != "tls_disabled" {
		t.Errorf("Name() = %s, want tls_disabled", rule.Name())
	}
}

func TestTLSDisabledRule_Severity(t *testing.T) {
	rule := NewTLSDisabledRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
	}{
		{
			name: "insecure_skip_verify - HIGH",
			cfg: map[string]any{
				"tls": map[string]any{"insecure_skip_verify": true},
			},
			wantIssues: 1,
		},
		{
			name: "enabled: false - HIGH",
			cfg: map[string]any{
				"tls": map[string]any{"enabled": false},
			},
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(&parser.Config{Data: tt.cfg})
			if len(issues) != tt.wantIssues {
				t.Fatalf("Expected %d issue(s), got %d", tt.wantIssues, len(issues))
			}
			if issues[0].Severity != HIGH {
				t.Errorf("Expected HIGH severity, got %s", issues[0].Severity)
			}
		})
	}
}

func TestTLSDisabledRule_CaseInsensitive(t *testing.T) {
	rule := NewTLSDisabledRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
	}{
		{
			name: "TLS.Enabled (mixed case)",
			cfg: map[string]any{
				"TLS": map[string]any{"Enabled": false},
			},
			wantIssues: 1,
		},
		{
			name: "TLS.INSECURE_SKIP_VERIFY (uppercase)",
			cfg: map[string]any{
				"TLS": map[string]any{"INSECURE_SKIP_VERIFY": true},
			},
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(&parser.Config{Data: tt.cfg})
			if len(issues) != tt.wantIssues {
				t.Errorf("Check() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}
