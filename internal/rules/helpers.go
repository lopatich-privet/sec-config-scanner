package rules

import (
	"fmt"
)

func traverseAndCheck(node any, path string, checker func(path string, value any) bool) {
	switch v := node.(type) {
	case map[string]any:
		traverseMap(v, path, checker)
	case map[any]any:
		traverseAnyMap(v, path, checker)
	case []any:
		traverseSlice(v, path, checker)
	default:
		checker(path, node)
	}
}

func traverseMap(v map[string]any, path string, checker func(path string, value any) bool) {
	for k, val := range v {
		traverseAndCheck(val, joinPath(path, k), checker)
	}
}

func traverseAnyMap(v map[any]any, path string, checker func(path string, value any) bool) {
	for k, val := range v {
		traverseAndCheck(val, joinPath(path, fmt.Sprintf("%v", k)), checker)
	}
}

func traverseSlice(v []any, path string, checker func(path string, value any) bool) {
	for i, item := range v {
		traverseAndCheck(item, fmt.Sprintf("%s[%d]", path, i), checker)
	}
}

func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
