package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FilePermissionRule struct{}

func (r *FilePermissionRule) Name() string {
	return "file_permissions"
}

func (r *FilePermissionRule) Check(cfg map[string]any) []Issue {
	var issues []Issue

	// Check if config contains file paths
	filePaths := extractFilePaths(cfg)

	for _, filePath := range filePaths {
		// Get file info to check permissions
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			// If file doesn't exist, skip it
			continue
		}

		// Check if file permissions allow others to read/write/execute
		// 0077 is the permission bits for "others" (world)
		// If any of these are set, the file is world-accessible
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

		// Check if the field name suggests a file path
		lowerPath := strings.ToLower(path)
		fileKeywords := []string{
			"file", "path", "config", "key", "cert", "pem",
			"private_key", "public_key", "certificate", "crt",
		}

		for _, keyword := range fileKeywords {
			if strings.Contains(lowerPath, keyword) {
				// Resolve absolute path if needed
				if filepath.IsAbs(str) {
					filePaths = append(filePaths, str)
				}
				return true
			}
		}

		return false
	})

	return filePaths
}

func NewFilePermissionRule() Rule {
	return &FilePermissionRule{}
}
