package scenario

import "testing"

// v0.2 の checks.must_be_number (数字のみ) と同じ規則であることを固定する。
func TestNewSeq(t *testing.T) {
	t.Run("NewSeq_数字のみの場合_先頭ゼロを保持して成功すること", func(t *testing.T) {
		// Arrange
		valid := []string{"10", "0", "010", "999999"}
		for _, s := range valid {
			// Act
			seq, err := NewSeq(s)
			// Assert
			if err != nil {
				t.Errorf("NewSeq(%q) = %v, want success", s, err)
				continue
			}
			// 先頭ゼロを保持する (ディレクトリ名の一部になるため)
			if seq.String() != s {
				t.Errorf("NewSeq(%q).String() = %q", s, seq.String())
			}
		}
	})

	t.Run("NewSeq_数字以外を含む場合_エラーであること", func(t *testing.T) {
		// Arrange
		invalid := []string{"", "1a", "a1", "-1", "1.5", "１０"}
		for _, s := range invalid {
			// Act
			_, err := NewSeq(s)
			// Assert
			if err == nil {
				t.Errorf("NewSeq(%q) should fail", s)
			}
		}
	})
}
