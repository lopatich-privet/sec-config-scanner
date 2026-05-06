package parser

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ParseDirectory recursively parses all JSON/YAML files in given directory
func ParseDirectory(ctx context.Context, dir string) ([]*Config, error) {
	var configs []*Config

	walker := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return parseFileIfConfig(path, d, &configs)
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

func parseFileIfConfig(path string, d os.DirEntry, configs *[]*Config) error {
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

	*configs = append(*configs, config)
	return nil
}

func isConfigFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"
}
