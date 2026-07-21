package runscenario

import (
	"bytes"
	"testing"
)

func TestParseMaxParallel(t *testing.T) {
	t.Run("parseMaxParallel_未設定の場合_0であること", func(t *testing.T) {
		// Arrange
		v := ""
		// Act
		got, err := parseMaxParallel(v)
		// Assert
		if err != nil || got != 0 {
			t.Errorf("parseMaxParallel(%q) = (%d, %v), want (0, nil)", v, got, err)
		}
	})

	t.Run("parseMaxParallel_0の場合_上限なしの0であること", func(t *testing.T) {
		// Arrange
		v := "0"
		// Act
		got, err := parseMaxParallel(v)
		// Assert
		if err != nil || got != 0 {
			t.Errorf("parseMaxParallel(%q) = (%d, %v), want (0, nil)", v, got, err)
		}
	})

	t.Run("parseMaxParallel_正整数の場合_その値であること", func(t *testing.T) {
		// Arrange
		v := "2"
		// Act
		got, err := parseMaxParallel(v)
		// Assert
		if err != nil || got != 2 {
			t.Errorf("parseMaxParallel(%q) = (%d, %v), want (2, nil)", v, got, err)
		}
	})

	t.Run("parseMaxParallel_負数の場合_エラーであること", func(t *testing.T) {
		// Arrange
		v := "-1"
		// Act
		_, err := parseMaxParallel(v)
		// Assert
		if err == nil {
			t.Errorf("parseMaxParallel(%q) = nil, want error", v)
		}
	})

	t.Run("parseMaxParallel_非整数の場合_エラーであること", func(t *testing.T) {
		// Arrange
		v := "abc"
		// Act
		_, err := parseMaxParallel(v)
		// Assert
		if err == nil {
			t.Errorf("parseMaxParallel(%q) = nil, want error", v)
		}
	})
}

func TestLineWriter(t *testing.T) {
	t.Run("Write_改行を含まない場合_下位Writerへ転送しないこと", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		w := &lineWriter{w: &buf}
		// Act
		n, err := w.Write([]byte("frag"))
		// Assert
		if err != nil || n != 4 {
			t.Fatalf("Write() = (%d, %v), want (4, nil)", n, err)
		}
		if buf.String() != "" {
			t.Errorf("underlying = %q, want empty (buffered until newline)", buf.String())
		}
	})

	t.Run("Write_改行を含む場合_完了行のみ転送すること", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		w := &lineWriter{w: &buf}
		// Act
		_, _ = w.Write([]byte("ab"))
		_, _ = w.Write([]byte("c\nde"))
		// Assert
		if buf.String() != "abc\n" {
			t.Errorf("underlying = %q, want %q (incomplete line stays buffered)", buf.String(), "abc\n")
		}
	})

	t.Run("Flush_未改行の残りがある場合_改行を補って転送すること", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		w := &lineWriter{w: &buf}
		_, _ = w.Write([]byte("abc\nde"))
		// Act
		err := w.Flush()
		// Assert
		if err != nil {
			t.Fatalf("Flush() = %v, want nil", err)
		}
		if buf.String() != "abc\nde\n" {
			t.Errorf("underlying = %q, want %q (newline appended on flush)", buf.String(), "abc\nde\n")
		}
	})

	t.Run("Flush_残りが無い場合_何も転送しないこと", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		w := &lineWriter{w: &buf}
		_, _ = w.Write([]byte("abc\n"))
		before := buf.String()
		// Act
		err := w.Flush()
		// Assert
		if err != nil || buf.String() != before {
			t.Errorf("Flush() = %v, underlying = %q, want no additional output", err, buf.String())
		}
	})
}
