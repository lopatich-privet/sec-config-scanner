package rules

import (
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

func TestBindAllRule_Check(t *testing.T) {
	rule := NewBindAllRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
		wantFields []string
	}{
		{
			name: "bind to 0.0.0.0 - exact match",
			cfg: map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
				},
			},
			wantIssues: 1,
			wantFields: []string{"server.host"},
		},
		{
			name: "bind to 0.0.0.0 - nested config",
			cfg: map[string]any{
				"network": map[string]any{
					"bind": "0.0.0.0",
				},
			},
			wantIssues: 1,
			wantFields: []string{"network.bind"},
		},
		{
			name: "bind to 127.0.0.1 - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": "127.0.0.1",
				},
			},
			wantIssues: 0,
		},
		{
			name: "bind to localhost - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
			wantIssues: 0,
		},
		{
			name: "bind to 192.168.1.1 - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": "192.168.1.1",
				},
			},
			wantIssues: 0,
		},
		{
			name: "empty string - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": "",
				},
			},
			wantIssues: 0,
		},
		{
			name: "multiple fields, one with 0.0.0.0",
			cfg: map[string]any{
				"server": map[string]any{
					"host1": "127.0.0.1",
					"host2": "0.0.0.0",
				},
			},
			wantIssues: 1,
			wantFields: []string{"server.host2"},
		},
		{
			name: "non-string value - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": 8080,
				},
			},
			wantIssues: 0,
		},
		{
			name: "nil value - no issue",
			cfg: map[string]any{
				"server": map[string]any{
					"host": nil,
				},
			},
			wantIssues: 0,
		},
		{
			name:       "empty config - no issue",
			cfg:        map[string]any{},
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

func TestBindAllRule_Name(t *testing.T) {
	rule := NewBindAllRule()
	if rule.Name() != "bind_all" {
		t.Errorf("Name() = %s, want bind_all", rule.Name())
	}
}

func TestBindAllRule_Severity(t *testing.T) {
	rule := NewBindAllRule()
	cfg := map[string]any{"host": "0.0.0.0"}
	issues := rule.Check(&parser.Config{Data: cfg})
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != MEDIUM {
		t.Errorf("Expected MEDIUM severity, got %s", issues[0].Severity)
	}
}

func TestBindAllRule_IssueDetails(t *testing.T) {
	rule := NewBindAllRule()
	cfg := map[string]any{"host": "0.0.0.0"}
	issues := rule.Check(&parser.Config{Data: cfg})
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if issue.Field != "host" {
		t.Errorf("Expected Field 'host', got '%s'", issue.Field)
	}
	if issue.Description != "сервис слушает на 0.0.0.0 без ограничений" {
		t.Errorf("Unexpected Description: %s", issue.Description)
	}
	expectedAdvice := "Ограничьте bind конкретным интерфейсом или внутренним IP."
	if issue.Advice != expectedAdvice {
		t.Errorf("Expected Advice '%s', got '%s'", expectedAdvice, issue.Advice)
	}
}
