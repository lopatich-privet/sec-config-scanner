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

	traverseAndCheck(cfg, "", func(path string, value any) bool {
		str, ok := value.(string)
		if !ok || strings.ToLower(str) != "debug" {
			return false
		}
		lowerPath := strings.ToLower(path)
		if strings.HasSuffix(lowerPath, "level") &&
			(strings.Contains(lowerPath, "log") || strings.Contains(lowerPath, "logging")) {
			issues = append(issues, Issue{
				Severity:    LOW,
				Field:       path,
				Description: "логирование в debug-режиме",
				Advice:      "Поменяйте режим на более избирательный (info+).",
			})
			return true
		}
		return false
	})

	return issues
}

func NewDebugLogRule() Rule {
	return &DebugLogRule{}
}
