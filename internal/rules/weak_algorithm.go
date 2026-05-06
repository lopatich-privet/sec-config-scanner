package rules

import (
	"fmt"
	"strings"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
)

var weakAlgorithms = map[string]struct{}{
	"md5": {}, "md4": {}, "md2": {}, "sha1": {}, "sha-1": {},
	"des": {}, "3des": {}, "rc4": {}, "null": {},
}

var algorithmKeywords = []string{"algorithm", "algo", "cipher", "digest", "hash", "encryption"}

func isWeakAlgorithm(value string) bool {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	_, ok := weakAlgorithms[lowerValue]
	return ok
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
	traverseAndCheck(cfg.Data, "", r.makeChecker(&issues))
	return issues
}

func (r *WeakAlgorithmRule) makeChecker(issues *[]Issue) func(path string, value any) bool {
	return func(path string, value any) bool {
		str, ok := value.(string)
		if !ok || str == "" {
			return false
		}
		if !containsAlgorithmKeyword(path) {
			return false
		}
		if isWeakAlgorithm(str) {
			*issues = append(*issues, Issue{
				Severity:    HIGH,
				Field:       path,
				Description: fmt.Sprintf("слишком слабый алгоритм - %s", str),
				Advice:      "Замените его на более безопасный.",
			})
			return true
		}
		return false
	}
}

func NewWeakAlgorithmRule() Rule {
	return &WeakAlgorithmRule{}
}
