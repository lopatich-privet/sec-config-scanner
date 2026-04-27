package rules

import "fmt"

type Severity string

const (
	LOW    Severity = "LOW"
	MEDIUM Severity = "MEDIUM"
	HIGH   Severity = "HIGH"
)

type Issue struct {
	Severity    Severity
	Field       string
	Description string
	Advice      string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: %s (%s). %s", i.Severity, i.Description, i.Field, i.Advice)
}

type Rule interface {
	Name() string
	Check(cfg map[string]any) []Issue
}
