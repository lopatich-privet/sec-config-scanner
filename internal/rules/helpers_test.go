package rules

import (
	"fmt"
	"strings"
	"testing"
)

func TestTraverseAndCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name: "flat map",
			input: map[string]any{
				"password": "admin123",
				"host":     "localhost",
			},
			expected: []string{"password=admin123", "host=localhost"},
		},
		{
			name: "nested map",
			input: map[string]any{
				"database": map[string]any{
					"password": "secret",
				},
			},
			expected: []string{"database.password=secret"},
		},
		{
			name: "array of maps",
			input: map[string]any{
				"users": []any{
					map[string]any{"password": "admin123"},
					map[string]any{"password": "secret"},
				},
			},
			expected: []string{
				"users[0].password=admin123",
				"users[1].password=secret",
			},
		},
		{
			name: "nested arrays",
			input: map[string]any{
				"groups": []any{
					[]any{
						map[string]any{"name": "root"},
					},
				},
			},
			expected: []string{"groups[0][0].name=root"},
		},
		{
			name: "array of scalars",
			input: map[string]any{
				"ports": []any{80, 443, "8080"},
			},
			expected: []string{
				"ports[0]=80",
				"ports[1]=443",
				"ports[2]=8080",
			},
		},
		{
			name: "map[any]any legacy YAML",
			input: map[any]any{
				"database": map[any]any{
					"password": "legacy_secret",
				},
			},
			expected: []string{"database.password=legacy_secret"},
		},
		{
			name: "mixed map types",
			input: map[string]any{
				"users": []any{
					map[any]any{
						"name":     "admin",
						"password": "from_legacy",
					},
				},
			},
			expected: []string{
				"users[0].name=admin",
				"users[0].password=from_legacy",
			},
		},
		{
			name: "root array",
			input: []any{
				map[string]any{"password": "root_array_pwd"},
			},
			expected: []string{"[0].password=root_array_pwd"},
		},
		{
			name: "bool and nil values",
			input: map[string]any{
				"enabled": true,
				"timeout": nil,
			},
			expected: []string{"enabled=true"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var visited []string
			traverseAndCheck(tc.input, "", func(path string, value any) bool {
				visited = append(visited, path+"="+toString(value))
				return false
			})

			for _, exp := range tc.expected {
				found := false
				for _, v := range visited {
					if v == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in visited paths, got %v", exp, visited)
				}
			}
		})
	}
}

func TestTraverseAndCheck_KeywordInArrayPath(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"users": []any{
			map[string]any{"password": "admin123"},
		},
	}

	var found bool
	traverseAndCheck(cfg, "", func(path string, value any) bool {
		if strings.Contains(path, "password") && value == "admin123" {
			found = true
		}
		return false
	})

	if !found {
		t.Error("password inside array not found by keyword match")
	}
}

func TestFindKeyWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       map[string]any
		keys      []string
		wantVal   any
		wantFound bool
	}{
		{
			name: "simple key",
			cfg: map[string]any{
				"level": "debug",
			},
			keys:      []string{"level"},
			wantVal:   "debug",
			wantFound: true,
		},
		{
			name: "nested key",
			cfg: map[string]any{
				"log": map[string]any{
					"level": "debug",
				},
			},
			keys:      []string{"log", "level"},
			wantVal:   "debug",
			wantFound: true,
		},
		{
			name: "key inside array",
			cfg: map[string]any{
				"users": []any{
					map[string]any{"name": "admin"},
					map[string]any{"name": "guest"},
				},
			},
			keys:      []string{"users", "name"},
			wantVal:   "admin",
			wantFound: true,
		},
		{
			name: "missing key",
			cfg: map[string]any{
				"other": "value",
			},
			keys:      []string{"log", "level"},
			wantFound: false,
		},
		{
			name: "empty keys",
			cfg:   map[string]any{"x": "y"},
			keys:  []string{},
			wantFound: false,
		},
		{
			name: "key not in array elements",
			cfg: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			keys:      []string{"items", "name"},
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			val, _ := findKeyWithPath(tc.cfg, tc.keys)
			if tc.wantFound {
				if val != tc.wantVal {
					t.Errorf("findKeyWithPath() value = %v, want %v", val, tc.wantVal)
				}
			} else if val != nil {
				t.Errorf("findKeyWithPath() expected nil, got %v", val)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		parent, key, want string
	}{
		{"", "password", "password"},
		{"database", "password", "database.password"},
		{"servers[0]", "host", "servers[0].host"},
	}

	for _, tc := range tests {
		got := joinPath(tc.parent, tc.key)
		if got != tc.want {
			t.Errorf("joinPath(%q, %q) = %q, want %q", tc.parent, tc.key, got, tc.want)
		}
	}
}

func toString(v any) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}
