package rules

import (
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

type BindAllRule struct{}

func (r *BindAllRule) Name() string {
	return "bind_all"
}

func (r *BindAllRule) Check(cfg *parser.Config) []Issue {
	var issues []Issue

	traverseAndCheck(cfg.Data, "", func(path string, value any) bool {
		if value == nil {
			return false
		}

		str, ok := value.(string)
		if !ok {
			return false
		}

		if str == "0.0.0.0" {
			issues = append(issues, Issue{
				Severity:    MEDIUM,
				Field:       path,
				Description: "сервис слушает на 0.0.0.0 без ограничений",
				Advice:      "Ограничьте bind конкретным интерфейсом или внутренним IP.",
			})
			return true
		}

		return false
	})

	return issues
}

func NewBindAllRule() Rule {
	return &BindAllRule{}
}
