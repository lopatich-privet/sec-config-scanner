package rules

import (
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
		if issue := checkPlaintextPassword(path, value, passwordKeywords); issue != nil {
			issues = append(issues, *issue)
			return true
		}
		return false
	})

	return issues
}

func checkPlaintextPassword(path string, value any, keywords []string) *Issue {
	if !isValidStringValue(value) {
		return nil
	}

	str := value.(string)
	if isHash(str) {
		return nil
	}

	if !containsPasswordKeyword(path, keywords) {
		return nil
	}

	return &Issue{
		Severity:    HIGH,
		Field:       path,
		Description: "пароль в открытом виде",
		Advice:      "Используйте переменные окружения или vault для хранения секретов.",
	}
}

func isValidStringValue(value any) bool {
	str, ok := value.(string)
	return ok && str != ""
}

func containsPasswordKeyword(path string, keywords []string) bool {
	lowerPath := strings.ToLower(path)
	for _, keyword := range keywords {
		if strings.Contains(lowerPath, keyword) {
			return true
		}
	}
	return false
}

func isHash(s string) bool {
	if strings.HasPrefix(s, "$") {
		return true
	}

	if isHexString(s) {
		return true
	}

	if isBcryptHash(s) {
		return true
	}

	return false
}

func isHexString(s string) bool {
	if len(s) != 32 && len(s) != 40 && len(s) != 64 && len(s) != 128 {
		return false
	}

	return isAllHexChars(s)
}

func isAllHexChars(s string) bool {
	for _, c := range s {
		if !isHexChar(c) {
			return false
		}
	}
	return true
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isBcryptHash(s string) bool {
	if len(s) != 60 {
		return false
	}
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$")
}

func NewPlaintextPasswordRule() Rule {
	return &PlaintextPasswordRule{}
}
