package scenario

import "testing"

// v0.2 の checks.must_be_number (数字のみ) と同じ規則であることを固定する。
func TestNewSeq(t *testing.T) {
	valid := []string{"10", "0", "010", "999999"}
	for _, s := range valid {
		seq, err := NewSeq(s)
		if err != nil {
			t.Errorf("NewSeq(%q) = %v, want success", s, err)
			continue
		}
		// 先頭ゼロを保持する (ディレクトリ名の一部になるため)
		if seq.String() != s {
			t.Errorf("NewSeq(%q).String() = %q", s, seq.String())
		}
	}

	invalid := []string{"", "1a", "a1", "-1", "1.5", "１０"}
	for _, s := range invalid {
		if _, err := NewSeq(s); err == nil {
			t.Errorf("NewSeq(%q) should fail", s)
		}
	}
}
