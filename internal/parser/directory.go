package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseDirectory recursively parses all JSON/YAML files in the given directory
func ParseDirectory(dir string) ([]*Config, error) {
	var configs []*Config

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only parse JSON/YAML files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Parse the file
		config, err := ParseFile(path)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		configs = append(configs, config)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no valid config files found in directory: %s", dir)
	}

	return configs, nil
}
