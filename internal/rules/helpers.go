package rules

import (
	"fmt"
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
