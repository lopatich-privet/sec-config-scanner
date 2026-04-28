package output

import (
	"config-analyzer/internal/rules"
	"testing"
)

func TestNewOutput(t *testing.T) {
	issues := []rules.Issue{
		{
			Severity:    rules.HIGH,
			Field:       "test.field",
			Description: "test description",
			Advice:      "test advice",
		},
	}

	output := NewOutput(issues)

	if output == nil {
		t.Fatal("NewOutput() returned nil")
	}

	if len(output.issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(output.issues))
	}
}

func TestOutput_GetIssues(t *testing.T) {
	issues := []rules.Issue{
		{Severity: rules.HIGH, Field: "field1"},
		{Severity: rules.LOW, Field: "field2"},
	}

	output := NewOutput(issues)

	gotIssues := output.GetIssues()

	if len(gotIssues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(gotIssues))
	}

	if gotIssues[0].Field != "field1" {
		t.Errorf("Expected field1, got %s", gotIssues[0].Field)
	}
}

func TestOutput_HasIssues(t *testing.T) {
	tests := []struct {
		name       string
		issues     []rules.Issue
		wantHas    bool
	}{
		{
			name:    "has issues",
			issues:  []rules.Issue{{Severity: rules.HIGH, Field: "test"}},
			wantHas: true,
		},
		{
			name:    "empty issues",
			issues:  []rules.Issue{},
			wantHas: false,
		},
		{
			name:    "nil issues",
			issues:  nil,
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(tt.issues)
			if got := output.HasIssues(); got != tt.wantHas {
				t.Errorf("HasIssues() = %v, want %v", got, tt.wantHas)
			}
		})
	}
}

func TestOutput_Print(t *testing.T) {
	tests := []struct {
		name   string
		issues []rules.Issue
	}{
		{
			name:   "no issues",
			issues: []rules.Issue{},
		},
		{
			name: "single issue",
			issues: []rules.Issue{
				{
					Severity:    rules.HIGH,
					Field:       "test.field",
					Description: "test description",
					Advice:      "test advice",
				},
			},
		},
		{
			name: "multiple issues",
			issues: []rules.Issue{
				{Severity: rules.LOW, Field: "low.field", Description: "low", Advice: "fix low"},
				{Severity: rules.MEDIUM, Field: "medium.field", Description: "medium", Advice: "fix medium"},
				{Severity: rules.HIGH, Field: "high.field", Description: "high", Advice: "fix high"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(tt.issues)
			// Just verify Print doesn't panic
			output.Print()
		})
	}
}

func TestOutput_IssueString(t *testing.T) {
	issue := rules.Issue{
		Severity:    rules.HIGH,
		Field:       "test.field",
		Description: "test description",
		Advice:      "test advice",
	}

	expected := "HIGH: test description. test advice"
	if got := issue.String(); got != expected {
		t.Errorf("Issue.String() = %s, want %s", got, expected)
	}
}

func TestOutput_Sorting(t *testing.T) {
	issues := []rules.Issue{
		{Severity: rules.LOW, Field: "low", Description: "low issue"},
		{Severity: rules.MEDIUM, Field: "medium", Description: "medium issue"},
		{Severity: rules.HIGH, Field: "high", Description: "high issue"},
		{Severity: rules.LOW, Field: "low2", Description: "low issue 2"},
	}

	output := NewOutput(issues)

	// Print triggers sorting
	output.Print()

	gotIssues := output.GetIssues()

	// After sorting: HIGH, MEDIUM, LOW, LOW
	expectedOrder := []rules.Severity{rules.HIGH, rules.MEDIUM, rules.LOW, rules.LOW}

	if len(gotIssues) != len(expectedOrder) {
		t.Fatalf("Expected %d issues, got %d", len(expectedOrder), len(gotIssues))
	}

	for i, expectedSeverity := range expectedOrder {
		if gotIssues[i].Severity != expectedSeverity {
			t.Errorf("Issue %d: expected severity %s, got %s", i, expectedSeverity, gotIssues[i].Severity)
		}
	}
}

func TestOutput_AllSeverities(t *testing.T) {
	tests := []struct {
		name     string
		severity rules.Severity
	}{
		{"LOW severity", rules.LOW},
		{"MEDIUM severity", rules.MEDIUM},
		{"HIGH severity", rules.HIGH},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := []rules.Issue{
				{
					Severity:    tt.severity,
					Field:       "test",
					Description: "test",
					Advice:      "advice",
				},
			}

			output := NewOutput(issues)

			if !output.HasIssues() {
				t.Error("Expected HasIssues() to return true")
			}

			gotIssues := output.GetIssues()
			if gotIssues[0].Severity != tt.severity {
				t.Errorf("Expected severity %s, got %s", tt.severity, gotIssues[0].Severity)
			}
		})
	}
}

func TestOutput_NilSafety(t *testing.T) {
	output := NewOutput(nil)

	if output.HasIssues() {
		t.Error("Expected HasIssues() to return false for nil")
	}

	if output.GetIssues() != nil {
		t.Error("Expected GetIssues() to return nil")
	}
}

func TestOutput_EmptyIssuesSlice(t *testing.T) {
	output := NewOutput([]rules.Issue{})

	if output.HasIssues() {
		t.Error("Expected HasIssues() to return false for empty slice")
	}

	if len(output.GetIssues()) != 0 {
		t.Error("Expected GetIssues() to return empty slice")
	}
}

func TestOutput_MultipleFields(t *testing.T) {
	issues := []rules.Issue{
		{Severity: rules.HIGH, Field: "field1", Description: "desc1", Advice: "adv1"},
		{Severity: rules.MEDIUM, Field: "field2", Description: "desc2", Advice: "adv2"},
		{Severity: rules.LOW, Field: "field3", Description: "desc3", Advice: "adv3"},
	}

	output := NewOutput(issues)

	gotIssues := output.GetIssues()

	if len(gotIssues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(gotIssues))
	}

	fields := []string{"field1", "field2", "field3"}
	for i, expectedField := range fields {
		if gotIssues[i].Field != expectedField {
			t.Errorf("Issue %d: expected field %s, got %s", i, expectedField, gotIssues[i].Field)
		}
	}
}
