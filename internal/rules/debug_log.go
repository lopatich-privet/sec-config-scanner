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

func findKeyWithPath(cfg map[string]any, keys []string) (any, string) {
	if len(keys) == 0 {
		return nil, ""
	}

	current := cfg
	var path strings.Builder

	for i, key := range keys {
		path.WriteString(key)
		if i < len(keys)-1 {
			path.WriteString(".")
		}

		val, ok := current[key]
		if !ok {
			return nil, path.String()
		}

		if i == len(keys)-1 {
			return val, path.String()
		}

		next, ok := val.(map[string]any)
		if !ok {
			return nil, path.String()
		}

		current = next
	}

	return nil, path.String()
}
