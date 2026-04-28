package rules

import (
	"strings"
)

type TLSDisabledRule struct{}

func (r *TLSDisabledRule) Name() string {
	return "tls_disabled"
}

func (r *TLSDisabledRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	traverseAndCheck(cfg, "", func(path string, value any) bool {
		boolVal, ok := value.(bool)
		if !ok {
			return false
		}

		lowerPath := strings.ToLower(path)

		if strings.Contains(lowerPath, "insecure_skip_verify") && boolVal {
			issues = append(issues, Issue{
				Severity:    HIGH,
				Field:       path,
				Description: "TLS проверка отключена",
				Advice:      "Включите TLS в продакшн-окружении.",
			})
			return true
		}

		if strings.Contains(lowerPath, "enabled") && !boolVal {
			issues = append(issues, Issue{
				Severity:    HIGH,
				Field:       path,
				Description: "TLS отключён",
				Advice:      "Включите TLS в продакшн-окружении.",
			})
			return true
		}

		return false
	})

	return issues
}

func NewTLSDisabledRule() Rule {
	return &TLSDisabledRule{}
}
