package rules

import "strings"

func traverseAndCheck(cfg map[string]any, path string, checker func(path string, value any) bool) {
	for key, value := range cfg {
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			traverseAndCheck(v, currentPath, checker)
		default:
			if checker(currentPath, value) {
				continue
			}
		}
	}
}

func findKeyWithPath(cfg map[string]any, keys []string) (any, string) {
	if len(keys) == 0 {
		return nil, ""
	}

	current := cfg
	var path strings.Builder

	for i, key := range keys {
		path.WriteString(key)
		if i < len(keys)-1 {
			path.WriteString(".")
		}

		val, ok := current[key]
		if !ok {
			return nil, path.String()
		}

		if i == len(keys)-1 {
			return val, path.String()
		}

		next, ok := val.(map[string]any)
		if !ok {
			return nil, path.String()
		}

		current = next
	}

	return nil, path.String()
}
