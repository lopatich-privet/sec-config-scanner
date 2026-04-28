package analyzer

import (
	"testing"

	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
)

func TestNewAnalyzer(t *testing.T) {
	rulesList := []rules.Rule{
		rules.NewDebugLogRule(),
		rules.NewPlaintextPasswordRule(),
	}

	analyzer := NewAnalyzer(rulesList)

	if analyzer == nil {
		t.Fatal("NewAnalyzer() returned nil")
	}

	if len(analyzer.rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(analyzer.rules))
	}
}

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name       string
		rulesList  []rules.Rule
		cfg        map[string]any
		wantIssues int
	}{
		{
			name: "no rules, no issues",
			cfg: map[string]any{
				"log": map[string]any{"level": "debug"},
			},
			rulesList:  []rules.Rule{},
			wantIssues: 0,
		},
		{
			name: "single rule, no issue",
			cfg: map[string]any{
				"log": map[string]any{"level": "info"},
			},
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
			},
			wantIssues: 0,
		},
		{
			name: "single rule, one issue",
			cfg: map[string]any{
				"log": map[string]any{"level": "debug"},
			},
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
			},
			wantIssues: 1,
		},
		{
			name: "multiple rules, multiple issues",
			cfg: map[string]any{
				"log":    map[string]any{"level": "debug"},
				"server": map[string]any{"host": "0.0.0.0"},
			},
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
				rules.NewBindAllRule(),
			},
			wantIssues: 2,
		},
		{
			name: "all rules, multiple issues",
			cfg: map[string]any{
				"log":    map[string]any{"level": "debug"},
				"server": map[string]any{"host": "0.0.0.0"},
				"tls":    map[string]any{"enabled": false},
				"db":     map[string]any{"password": "secret"},
			},
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
				rules.NewBindAllRule(),
				rules.NewTLSDisabledRule(),
				rules.NewPlaintextPasswordRule(),
			},
			wantIssues: 4,
		},
		{
			name: "empty config",
			cfg:  map[string]any{},
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
			},
			wantIssues: 0,
		},
		{
			name: "nil config",
			cfg:  nil,
			rulesList: []rules.Rule{
				rules.NewDebugLogRule(),
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(tt.rulesList)
			issues := analyzer.Analyze(tt.cfg)

			if len(issues) != tt.wantIssues {
				t.Errorf("Analyze() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestAddRule(t *testing.T) {
	analyzer := NewAnalyzer([]rules.Rule{})

	if len(analyzer.rules) != 0 {
		t.Errorf("Expected 0 rules initially, got %d", len(analyzer.rules))
	}

	analyzer.AddRule(rules.NewDebugLogRule())

	if len(analyzer.rules) != 1 {
		t.Errorf("Expected 1 rule after AddRule, got %d", len(analyzer.rules))
	}

	analyzer.AddRule(rules.NewBindAllRule())

	if len(analyzer.rules) != 2 {
		t.Errorf("Expected 2 rules after second AddRule, got %d", len(analyzer.rules))
	}
}

func TestGetRules(t *testing.T) {
	rulesList := []rules.Rule{
		rules.NewDebugLogRule(),
		rules.NewBindAllRule(),
	}

	analyzer := NewAnalyzer(rulesList)

	gotRules := analyzer.GetRules()

	if len(gotRules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(gotRules))
	}

	if gotRules[0].Name() != "debug_log" {
		t.Errorf("Expected first rule to be debug_log, got %s", gotRules[0].Name())
	}

	if gotRules[1].Name() != "bind_all" {
		t.Errorf("Expected second rule to be bind_all, got %s", gotRules[1].Name())
	}
}

func TestAnalyze_IssueAggregation(t *testing.T) {
	analyzer := NewAnalyzer([]rules.Rule{
		rules.NewDebugLogRule(),
		rules.NewPlaintextPasswordRule(),
	})

	cfg := map[string]any{
		"log": map[string]any{
			"level": "debug",
		},
		"database": map[string]any{
			"password": "secret123",
		},
	}

	issues := analyzer.Analyze(cfg)

	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}

	foundDebug := false
	foundPassword := false

	for _, issue := range issues {
		if issue.Severity == rules.LOW && issue.Field == "log.level" {
			foundDebug = true
		}
		if issue.Severity == rules.HIGH && issue.Field == "database.password" {
			foundPassword = true
		}
	}

	if !foundDebug {
		t.Error("Expected to find debug log issue")
	}

	if !foundPassword {
		t.Error("Expected to find plaintext password issue")
	}
}

func TestAnalyze_NilSafety(t *testing.T) {
	rule := rules.NewDebugLogRule()
	analyzer := NewAnalyzer([]rules.Rule{rule})

	issues := analyzer.Analyze(nil)

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for nil config, got %d", len(issues))
	}
}

func TestAnalyzer_InitialState(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	if analyzer == nil {
		t.Fatal("NewAnalyzer with nil returned nil")
	}

	if len(analyzer.rules) != 0 {
		t.Errorf("Expected 0 rules for nil input, got %d", len(analyzer.rules))
	}
}
