package scenario

import "testing"

// v0.2 の checks.must_be_date_format (8 桁数字) + v1.0 で追加した実在日付検証を固定する。
func TestNewBizdate(t *testing.T) {
	t.Run("NewBizdate_8桁実在日の場合_成功すること", func(t *testing.T) {
		// Arrange
		valid := []string{"99990101", "20260228", "20240229", "00010101"}
		for _, s := range valid {
			// Act
			b, err := NewBizdate(s)
			// Assert
			if err != nil {
				t.Errorf("NewBizdate(%q) = %v, want success", s, err)
				continue
			}
			if b.String() != s {
				t.Errorf("NewBizdate(%q).String() = %q", s, b.String())
			}
		}
	})

	t.Run("NewBizdate_不正な書式や実在しない日付の場合_エラーであること", func(t *testing.T) {
		// Arrange
		invalid := map[string]string{
			"2025011":   "7 桁",
			"202501011": "9 桁",
			"2025010a":  "数字以外を含む",
			"":          "空文字",
			"あいうえおかきく":  "全角 8 文字",
			"20250230":  "実在しない日付 (2/30)",
			"20251301":  "実在しない月 (13 月)",
			"20250100":  "実在しない日 (0 日)",
		}
		for s, reason := range invalid {
			// Act
			_, err := NewBizdate(s)
			// Assert
			if err == nil {
				t.Errorf("NewBizdate(%q) should fail (%s)", s, reason)
			}
		}
	})
}
