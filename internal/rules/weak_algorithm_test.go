package rules

import "testing"

func TestWeakAlgorithmRule_Check(t *testing.T) {
	rule := NewWeakAlgorithmRule()

	tests := []struct {
		name       string
		cfg        map[string]any
		wantIssues int
		wantFields []string
	}{
		{
			name: "MD5 algorithm",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "md5",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.algorithm"},
		},
		{
			name: "SHA1 algorithm",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algo": "sha1",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.algo"},
		},
		{
			name: "DES cipher",
			cfg: map[string]any{
				"encryption": map[string]any{
					"cipher": "des",
				},
			},
			wantIssues: 1,
			wantFields: []string{"encryption.cipher"},
		},
		{
			name: "RC4 algorithm",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "rc4",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.algorithm"},
		},
		{
			name: "3DES cipher",
			cfg: map[string]any{
				"crypto": map[string]any{
					"cipher": "3des",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.cipher"},
		},
		{
			name: "NULL cipher",
			cfg: map[string]any{
				"crypto": map[string]any{
					"cipher": "null",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.cipher"},
		},
		{
			name: "SHA-1 (with dash)",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "sha-1",
				},
			},
			wantIssues: 1,
			wantFields: []string{"crypto.algorithm"},
		},
		{
			name: "SHA256 - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "sha256",
				},
			},
			wantIssues: 0,
		},
		{
			name: "SHA512 - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "sha512",
				},
			},
			wantIssues: 0,
		},
		{
			name: "AES-256 - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"cipher": "aes-256-gcm",
				},
			},
			wantIssues: 0,
		},
		{
			name:       "empty string - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "",
				},
			},
			wantIssues: 0,
		},
		{
			name: "case insensitive detection",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "MD5",
				},
			},
			wantIssues: 1,
		},
		{
			name: "non-string value - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": 123,
				},
			},
			wantIssues: 0,
		},
		{
			name: "algorithm value without keyword - no issue",
			cfg: map[string]any{
				"value": "md5",
			},
			wantIssues: 0,
		},
		{
			name: "keyword without matching algorithm - no issue",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "unknown",
				},
			},
			wantIssues: 0,
		},
		{
			name: "trailing whitespace detection",
			cfg: map[string]any{
				"crypto": map[string]any{
					"algorithm": "  md5  ",
				},
			},
			wantIssues: 1,
		},
		{
			name: "multiple weak algorithms",
			cfg: map[string]any{
				"crypto1": map[string]any{"algorithm": "md5"},
				"crypto2": map[string]any{"algorithm": "sha1"},
			},
			wantIssues: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(tt.cfg)
			if len(issues) != tt.wantIssues {
				t.Errorf("Check() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
			if tt.wantFields != nil && len(issues) > 0 {
				for i, wantField := range tt.wantFields {
					if i >= len(issues) {
						break
					}
					if issues[i].Field != wantField {
						t.Errorf("Check() issue %d field = %s, want %s", i, issues[i].Field, wantField)
					}
				}
			}
		})
	}
}

func TestWeakAlgorithmRule_Name(t *testing.T) {
	rule := NewWeakAlgorithmRule()
	if rule.Name() != "weak_algorithm" {
		t.Errorf("Name() = %s, want weak_algorithm", rule.Name())
	}
}

func TestWeakAlgorithmRule_Severity(t *testing.T) {
	rule := NewWeakAlgorithmRule()
	cfg := map[string]any{"crypto": map[string]any{"algorithm": "md5"}}
	issues := rule.Check(cfg)
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != HIGH {
		t.Errorf("Expected HIGH severity, got %s", issues[0].Severity)
	}
}

func TestWeakAlgorithmRule_AllWeakAlgorithms(t *testing.T) {
	weakAlgorithms := []string{"md5", "md4", "md2", "sha1", "sha-1", "des", "3des", "rc4", "null"}
	rule := NewWeakAlgorithmRule()

	for _, algo := range weakAlgorithms {
		t.Run(algo, func(t *testing.T) {
			cfg := map[string]any{
				"crypto": map[string]any{
					"algorithm": algo,
				},
			}
			issues := rule.Check(cfg)
			if len(issues) != 1 {
				t.Errorf("Algorithm %s: expected 1 issue, got %d", algo, len(issues))
			}
		})
	}
}

func TestWeakAlgorithmRule_AllKeywords(t *testing.T) {
	keywords := []string{"algorithm", "algo", "cipher", "digest", "hash", "encryption"}
	rule := NewWeakAlgorithmRule()

	for _, keyword := range keywords {
		t.Run(keyword, func(t *testing.T) {
			cfg := map[string]any{
				"crypto": map[string]any{
					keyword: "md5",
				},
			}
			issues := rule.Check(cfg)
			if len(issues) != 1 {
				t.Errorf("Keyword %s: expected 1 issue, got %d", keyword, len(issues))
			}
		})
	}
}
