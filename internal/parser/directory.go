package parser

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ParseDirectory recursively parses all JSON/YAML files in the given directory
func ParseDirectory(dir string) ([]*Config, error) {
	var configs []*Config

	walker := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !isConfigFile(path) {
			return nil
		}

		config, err := ParseFile(path)
		if err != nil {
			slog.Warn("failed to parse file", "path", path, "error", err)
			return nil
		}

		configs = append(configs, config)
		return nil
	}

	err := filepath.WalkDir(dir, walker)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no valid config files found in directory: %s", dir)
	}

	return configs, nil
}

func isConfigFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"
}
