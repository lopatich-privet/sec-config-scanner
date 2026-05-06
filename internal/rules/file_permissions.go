package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
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

func (r *FilePermissionRule) Check(cfg *parser.Config) []Issue {
	var issues []Issue

	// Проверка прав самого конфигурационного файла
	if cfg.FilePath != "" {
		issues = append(issues, r.CheckFilePermissions(cfg.FilePath)...)
	}

	// Проверка путей из значений конфига
	filePaths := extractFilePaths(cfg.Data)
	for _, filePath := range filePaths {
		issues = append(issues, r.CheckFilePermissions(filePath)...)
	}

	return issues
}

func (r *FilePermissionRule) CheckFilePermissions(filePath string) []Issue {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return []Issue{{
			Severity:    LOW,
			Field:       filePath,
			Description: fmt.Sprintf("не удалось проверить права файла: %v", err),
			Advice:      "Убедитесь, что путь существует и доступен для чтения.",
		}}
	}

	mode := fileInfo.Mode().Perm()
	permStr := fmt.Sprintf("%04o", mode)

	groupOtherWrite := mode & 0022
	groupOtherRead := mode & 0044

	if groupOtherWrite != 0 {
		return []Issue{{
			Severity:    HIGH,
			Field:       filePath,
			Description: fmt.Sprintf("файл доступен для записи всем: %s", permStr),
			Advice:      "Срочно ограничьте права (рекомендуется 0600). Злоумышленник может подменить конфигурацию.",
		}}
	}

	if groupOtherRead != 0 {
		return []Issue{{
			Severity:    MEDIUM,
			Field:       filePath,
			Description: fmt.Sprintf("файл доступен для чтения всем: %s", permStr),
			Advice:      "Ограничьте права доступа (рекомендуется 0600 или 0640).",
		}}
	}

	return nil
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
