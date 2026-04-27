package rules

import (
	"fmt"
	"strings"
)

type PlaintextPasswordRule struct{}

func (r *PlaintextPasswordRule) Name() string {
	return "plaintext_password"
}

func (r *PlaintextPasswordRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	passwordKeywords := []string{"password", "passwd", "pwd", "secret"}

	traverseAndCheck(cfg, "", func(path string, value any) bool {
		if value == nil {
			return false
		}

		str, ok := value.(string)
		if !ok {
			return false
		}

		if str == "" {
			return false
		}

		lowerPath := strings.ToLower(path)
		for _, keyword := range passwordKeywords {
			if strings.Contains(lowerPath, keyword) {
				if !isHash(str) {
					issues = append(issues, Issue{
						Severity:    HIGH,
						Field:       path,
						Description: fmt.Sprintf("пароль в открытом виде"),
						Advice:      "Используйте переменные окружения или vault для хранения секретов.",
					})
					return true
				}
			}
		}
		return false
	})

	return issues
}

func isHash(s string) bool {
	// Проверяем переменные окружения: $VAR или ${VAR}
	if strings.HasPrefix(s, "$") {
		return true
	}

	// Проверяем, выглядит ли строка как хеш
	if len(s) == 32 || len(s) == 40 || len(s) == 64 || len(s) == 128 {
		// MD5: 32 hex chars
		// SHA1: 40 hex chars
		// SHA256: 64 hex chars
		// SHA512: 128 hex chars
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	// Проверяем bcrypt (60 chars, с префиксом)
	if len(s) == 60 && (strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$")) {
		return true
	}

	return false
}

func NewPlaintextPasswordRule() Rule {
	return &PlaintextPasswordRule{}
}
