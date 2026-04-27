package analyzer

import (
	"config-analyzer/internal/rules"
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

func (a *Analyzer) AddRule(rule rules.Rule) {
	a.rules = append(a.rules, rule)
}

func (a *Analyzer) GetRules() []rules.Rule {
	return a.rules
}




