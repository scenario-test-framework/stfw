package scenario

import "testing"

// v0.2 の checks.must_not_contains (`_` 禁止) と同じ規則であることを固定する。
func TestNewGroup(t *testing.T) {
	t.Run("NewGroup_許可された文字列の場合_成功すること", func(t *testing.T) {
		// Arrange
		valid := []string{"pre", "post", "web-ap", "group1"}
		for _, s := range valid {
			// Act
			g, err := NewGroup(s)
			// Assert
			if err != nil {
				t.Errorf("NewGroup(%q) = %v, want success", s, err)
				continue
			}
			if g.String() != s {
				t.Errorf("NewGroup(%q).String() = %q", s, g.String())
			}
		}
	})

	t.Run("NewGroup_禁止文字を含む場合_エラーであること", func(t *testing.T) {
		// Arrange
		invalid := map[string]string{
			"":          "空文字",
			"pre_group": "`_` を含む (ディレクトリ名パースの保護)",
			"_pre":      "`_` 始まり",
			"pre/sub":   "パス区切り文字 (v1.0 追加ガード)",
		}
		for s, reason := range invalid {
			// Act
			_, err := NewGroup(s)
			// Assert
			if err == nil {
				t.Errorf("NewGroup(%q) should fail (%s)", s, reason)
			}
		}
	})
}
