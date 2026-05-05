package rules

import (
	"fmt"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

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
	return fmt.Sprintf("%s: %s. %s", i.Severity, i.Description, i.Advice)
}

type Rule interface {
	Name() string
	Check(cfg *parser.Config) []Issue
}
