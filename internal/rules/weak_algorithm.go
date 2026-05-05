package rules

import (
	"fmt"
	"strings"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

var weakAlgorithms = map[string]bool{
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

var algorithmKeywords = []string{"algorithm", "algo", "cipher", "digest", "hash", "encryption"}

func isWeakAlgorithm(value string) bool {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	return weakAlgorithms[lowerValue]
}

func containsAlgorithmKeyword(path string) bool {
	lowerPath := strings.ToLower(path)
	for _, keyword := range algorithmKeywords {
		if strings.Contains(lowerPath, keyword) {
			return true
		}
	}
	return false
}

type WeakAlgorithmRule struct{}

func (r *WeakAlgorithmRule) Name() string {
	return "weak_algorithm"
}

func (r *WeakAlgorithmRule) Check(cfg *parser.Config) []Issue {
	var issues []Issue

	traverseAndCheck(cfg.Data, "", func(path string, value any) bool {
		if value == nil {
			return false
		}

		str, ok := value.(string)
		if !ok || str == "" {
			return false
		}

		if !containsAlgorithmKeyword(path) {
			return false
		}

		if isWeakAlgorithm(str) {
			issues = append(issues, Issue{
				Severity:    HIGH,
				Field:       path,
				Description: fmt.Sprintf("слишком слабый алгоритм - %s", str),
				Advice:      "Замените его на более безопасный.",
			})
			return true
		}

		return false
	})

	return issues
}

func NewWeakAlgorithmRule() Rule {
	return &WeakAlgorithmRule{}
}
