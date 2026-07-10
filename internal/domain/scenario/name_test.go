package scenario

import "testing"

func TestNewScenarioName(t *testing.T) {
	t.Run("NewScenarioName_許可された文字列の場合_成功すること", func(t *testing.T) {
		// Arrange
		valid := []string{"test", "sample", "test_case_1", "テストシナリオ"}
		for _, s := range valid {
			// Act
			n, err := NewScenarioName(s)
			// Assert
			if err != nil {
				t.Errorf("NewScenarioName(%q) = %v, want success", s, err)
				continue
			}
			if n.String() != s {
				t.Errorf("NewScenarioName(%q).String() = %q", s, n.String())
			}
		}
	})

	t.Run("NewScenarioName_空文字やパス区切りを含む場合_エラーであること", func(t *testing.T) {
		// Arrange
		invalid := map[string]string{
			"":    "空文字 (v0.2 の must_not_null と同じ)",
			".":   "カレントディレクトリ (v1.0 追加ガード)",
			"..":  "親ディレクトリ (v1.0 追加ガード)",
			"a/b": "パス区切り文字 (v1.0 追加ガード)",
			`a\b`: "パス区切り文字 (v1.0 追加ガード)",
		}
		for s, reason := range invalid {
			// Act
			_, err := NewScenarioName(s)
			// Assert
			if err == nil {
				t.Errorf("NewScenarioName(%q) should fail (%s)", s, reason)
			}
		}
	})
}
