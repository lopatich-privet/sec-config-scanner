package output

import (
	"config-analyzer/internal/rules"
	"fmt"
	"sort"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

type Output struct {
	issues []rules.Issue
}

func NewOutput(issues []rules.Issue) *Output {
	return &Output{issues: issues}
}

func (o *Output) Print() {
	if len(o.issues) == 0 {
		fmt.Println("✓ Конфигурация безопасна")
		return
	}

	// Сортируем по severity: HIGH → MEDIUM → LOW
	sort.Slice(o.issues, func(i, j int) bool {
		severityOrder := map[rules.Severity]int{
			rules.HIGH:   3,
			rules.MEDIUM: 2,
			rules.LOW:    1,
		}
		return severityOrder[o.issues[i].Severity] > severityOrder[o.issues[j].Severity]
	})

	for _, issue := range o.issues {
		fmt.Println(issue.String())
	}
}

func (o *Output) GetIssues() []rules.Issue {
	return o.issues
}

func (o *Output) HasIssues() bool {
	return len(o.issues) > 0
}
