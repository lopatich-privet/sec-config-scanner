package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var fileKeywords = []string{
	"file", "path", "config", "key", "cert", "pem",
	"private_key", "public_key", "certificate", "crt",
}

func isPathLikeField(path string) bool {
	lowerPath := strings.ToLower(path)
	for _, keyword := range fileKeywords {
		if strings.Contains(lowerPath, keyword) {
			return true
		}
	}
	return false
}

type FilePermissionRule struct{}

func (r *FilePermissionRule) Name() string {
	return "file_permissions"
}

func (r *FilePermissionRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	filePaths := extractFilePaths(cfg)

	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		mode := fileInfo.Mode().Perm()
		if mode&0077 != 0 {
			permStr := fmt.Sprintf("%04o", mode)
			issues = append(issues, Issue{
				Severity:    MEDIUM,
				Field:       filePath,
				Description: fmt.Sprintf("файл имеет избыточные права доступа: %s", permStr),
				Advice:      "Ограничьте права доступа (рекомендуется 0600 или 0640).",
			})
		}
	}

	return issues
}

func extractFilePaths(cfg map[string]any) []string {
	var filePaths []string

	traverseAndCheck(cfg, "", func(path string, value any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}

		if !isPathLikeField(path) {
			return false
		}

		if filepath.IsAbs(str) {
			filePaths = append(filePaths, str)
		}
		return true
	})

	return filePaths
}

func NewFilePermissionRule() Rule {
	return &FilePermissionRule{}
}
