package analyzer

import (
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
)

type Analyzer struct {
	rules []rules.Rule
}

func NewAnalyzer(rulesList []rules.Rule) *Analyzer {
	return &Analyzer{
		rules: rulesList,
	}
}

func (a *Analyzer) Analyze(cfg map[string]any) []rules.Issue {
	var issues []rules.Issue

	for _, rule := range a.rules {
		ruleIssues := rule.Check(cfg)
		issues = append(issues, ruleIssues...)
	}

	return issues
}
