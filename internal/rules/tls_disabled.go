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

	tlsKeys := map[string]bool{
		"tls.enabled":            true,
		"tls.insecure_skip_verify": true,
		"server.tls":            true,
		"server.ssl":            true,
	}

	traverseAndCheck(cfg, "", func(path string, value any) bool {
		if value == nil {
			return false
		}

		boolVal, ok := value.(bool)
		if !ok {
			return false
		}

		// Проверяем ключи из tlsKeys
		lowerPath := strings.ToLower(path)
		for key := range tlsKeys {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerPath, lowerKey) {
				if strings.Contains(lowerPath, "insecure_skip_verify") {
					if boolVal {
						issues = append(issues, Issue{
							Severity:    HIGH,
							Field:       path,
							Description: "TLS проверка отключена (insecure_skip_verify)",
							Advice:      "Включите TLS в продакшн-окружении.",
						})
						return true
					}
				} else if strings.Contains(lowerPath, "enabled") {
					if !boolVal {
						issues = append(issues, Issue{
							Severity:    HIGH,
							Field:       path,
							Description: "TLS отключён",
							Advice:      "Включите TLS в продакшн-окружении.",
						})
						return true
					}
				}
			}
		}

		return false
	})

	return issues
}

func NewTLSDisabledRule() Rule {
	return &TLSDisabledRule{}
}
