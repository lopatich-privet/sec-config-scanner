package rules

import (
	"strings"
)

type DebugLogRule struct{}

func (r *DebugLogRule) Name() string {
	return "debug_log"
}

func (r *DebugLogRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	value, path := findKeyWithPath(cfg, []string{"log", "level"})
	if value != nil {
		if str, ok := value.(string); ok && strings.ToLower(str) == "debug" {
			issues = append(issues, Issue{
				Severity:    LOW,
				Field:       path,
				Description: "логирование в debug-режиме",
				Advice:      "Поменяйте режим на более избирательный (info+).",
			})
		}
	}

	return issues
}

func NewDebugLogRule() Rule {
	return &DebugLogRule{}
}
