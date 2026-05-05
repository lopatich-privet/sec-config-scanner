package rules

import (
	"fmt"
	"strings"
)

func traverseAndCheck(node any, path string, checker func(path string, value any) bool) {
	switch v := node.(type) {
	case map[string]any:
		for k, val := range v {
			traverseAndCheck(val, joinPath(path, k), checker)
		}
	case map[any]any:
		for k, val := range v {
			traverseAndCheck(val, joinPath(path, fmt.Sprintf("%v", k)), checker)
		}
	case []any:
		for i, item := range v {
			traverseAndCheck(item, fmt.Sprintf("%s[%d]", path, i), checker)
		}
	default:
		checker(path, node)
	}
}

func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func findKeyWithPath(cfg map[string]any, keys []string) (any, string) {
	if len(keys) == 0 {
		return nil, ""
	}

	current := cfg
	var path strings.Builder

	for i, key := range keys {
		if i > 0 {
			path.WriteString(".")
		}
		path.WriteString(key)

		val, ok := current[key]
		if !ok {
			return nil, path.String()
		}

		if i == len(keys)-1 {
			return val, path.String()
		}

		if m, ok := val.(map[string]any); ok {
			current = m
			continue
		}

		return trySearchInArray(val, keys[i+1:], path.String())
	}

	return nil, path.String()
}

func trySearchInArray(val any, remainingKeys []string, basePath string) (any, string) {
	arr, ok := val.([]any)
	if !ok {
		return nil, basePath
	}

	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		result, p := findKeyWithPath(m, remainingKeys)
		if result != nil {
			return result, basePath + "." + p
		}
	}

	return nil, basePath
}
