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

		issue := r.checkTLSConfig(path, boolVal)
		if issue == nil {
			return false
		}

		issue.Field = path
		issues = append(issues, *issue)
		return true
	})

	return issues
}

func (r *TLSDisabledRule) checkTLSConfig(path string, enabled bool) *Issue {
	lowerPath := strings.ToLower(path)

	if strings.Contains(lowerPath, "insecure_skip_verify") && enabled {
		return &Issue{
			Severity:    HIGH,
			Description: "TLS проверка отключена",
			Advice:      "Включите TLS в продакшн-окружении.",
		}
	}

	if strings.Contains(lowerPath, "enabled") && !enabled {
		return &Issue{
			Severity:    HIGH,
			Description: "TLS отключён",
			Advice:      "Включите TLS в продакшн-окружении.",
		}
	}

	return nil
}

func NewTLSDisabledRule() Rule {
	return &TLSDisabledRule{}
}
