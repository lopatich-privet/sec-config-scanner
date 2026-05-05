package rules

import (
	"strings"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

type DebugLogRule struct{}

func (r *DebugLogRule) Name() string {
	return "debug_log"
}

func (r *DebugLogRule) Check(cfg *parser.Config) []Issue {
	var issues []Issue

	traverseAndCheck(cfg.Data, "", func(path string, value any) bool {
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
