package scenario

import "testing"

func TestNewScenarioName(t *testing.T) {
	valid := []string{"test", "sample", "test_case_1", "テストシナリオ"}
	for _, s := range valid {
		n, err := NewScenarioName(s)
		if err != nil {
			t.Errorf("NewScenarioName(%q) = %v, want success", s, err)
			continue
		}
		if n.String() != s {
			t.Errorf("NewScenarioName(%q).String() = %q", s, n.String())
		}
	}

	invalid := map[string]string{
		"":    "空文字 (v0.2 の must_not_null と同じ)",
		".":   "カレントディレクトリ (v1.0 追加ガード)",
		"..":  "親ディレクトリ (v1.0 追加ガード)",
		"a/b": "パス区切り文字 (v1.0 追加ガード)",
		`a\b`: "パス区切り文字 (v1.0 追加ガード)",
	}
	for s, reason := range invalid {
		if _, err := NewScenarioName(s); err == nil {
			t.Errorf("NewScenarioName(%q) should fail (%s)", s, reason)
		}
	}
}
