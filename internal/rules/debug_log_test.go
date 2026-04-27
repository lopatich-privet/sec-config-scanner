package rules

import "testing"

func TestDebugLogRule_Check(t *testing.T) {
	rule := NewDebugLogRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
		wantFields []string
	}{
		{
			name:       "debug mode - lowercase",
			cfg:        map[string]any{"log": map[string]any{"level": "debug"}},
			wantIssues: 1,
			wantFields: []string{"log.level"},
		},
		{
			name:       "debug mode - mixed case",
			cfg:        map[string]any{"log": map[string]any{"level": "DeBuG"}},
			wantIssues: 1,
			wantFields: []string{"log.level"},
		},
		{
			name:       "info mode - no issue",
			cfg:        map[string]any{"log": map[string]any{"level": "info"}},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "warn mode - no issue",
			cfg:        map[string]any{"log": map[string]any{"level": "warn"}},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "error mode - no issue",
			cfg:        map[string]any{"log": map[string]any{"level": "error"}},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "no log level field",
			cfg:        map[string]any{"log": map[string]any{"format": "json"}},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "top-level log level",
			cfg:        map[string]any{"level": "debug"},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "empty config",
			cfg:        map[string]any{},
			wantIssues: 0,
			wantFields: nil,
		},
		{
			name:       "nil config",
			cfg:        nil,
			wantIssues: 0,
			wantFields: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(tt.cfg)
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

func TestDebugLogRule_Name(t *testing.T) {
	rule := NewDebugLogRule()
	if rule.Name() != "debug_log" {
		t.Errorf("Name() = %s, want debug_log", rule.Name())
	}
}

func TestDebugLogRule_Severity(t *testing.T) {
	rule := NewDebugLogRule()
	cfg := map[string]any{"log": map[string]any{"level": "debug"}}
	issues := rule.Check(cfg)
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != LOW {
		t.Errorf("Expected LOW severity, got %s", issues[0].Severity)
	}
}
