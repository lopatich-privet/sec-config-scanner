package rules

import (
	"fmt"
	"strings"
)

type WeakAlgorithmRule struct{}

func (r *WeakAlgorithmRule) Name() string {
	return "weak_algorithm"
}

func (r *WeakAlgorithmRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	weakAlgorithms := map[string]bool{
		"md5":  true,
		"md4":  true,
		"md2":  true,
		"sha1": true,
		"sha-1": true,
		"des":  true,
		"3des": true,
		"rc4":  true,
		"null": true,
	}

	algorithmKeywords := []string{"algorithm", "algo", "cipher", "digest", "hash", "encryption"}

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
		lowerValue := strings.ToLower(strings.TrimSpace(str))

		// Проверяем, содержит ли путь одно из ключевых слов алгоритма
		for _, keyword := range algorithmKeywords {
			if strings.Contains(lowerPath, keyword) {
				if weakAlgorithms[lowerValue] {
					issues = append(issues, Issue{
						Severity:    HIGH,
						Field:       path,
						Description: fmt.Sprintf("слабый алгоритм — %s", str),
						Advice:      "Используйте SHA-256 или выше.",
					})
					return true
				}
			}
		}

		return false
	})

	return issues
}

func NewWeakAlgorithmRule() Rule {
	return &WeakAlgorithmRule{}
}
