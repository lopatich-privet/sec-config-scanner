package rules

import (
	"strings"
)

const enableTLSAdvice = "Включите TLS в продакшн-окружении."

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

func (r *TLSDisabledRule) hasTLSContext(path string) bool {
	tlsContextKeywords := []string{"tls", "ssl", "https", "secure"}
	lowerPath := strings.ToLower(path)
	for _, kw := range tlsContextKeywords {
		if strings.Contains(lowerPath, kw) {
			return true
		}
	}
	return false
}

func (r *TLSDisabledRule) checkTLSConfig(path string, enabled bool) *Issue {
	lowerPath := strings.ToLower(path)

	if strings.Contains(lowerPath, "insecure_skip_verify") && enabled {
		return &Issue{
			Severity:    HIGH,
			Description: "TLS проверка отключена",
			Advice:      enableTLSAdvice,
		}
	}

	if strings.Contains(lowerPath, "enabled") && !enabled && r.hasTLSContext(path) {
		return &Issue{
			Severity:    HIGH,
			Description: "TLS отключён",
			Advice:      enableTLSAdvice,
		}
	}

	return nil
}

func NewTLSDisabledRule() Rule {
	return &TLSDisabledRule{}
}
