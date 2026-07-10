package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestMaskerRegisterAndWrite(t *testing.T) {
	t.Run("Masker_登録済みシークレットを含む完了行の場合_マスクして出力すること", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		m := &Masker{reg: &secretRegistry{}, out: &buf}
		m.Register("s3cr3t")
		// 改行で終わる完了行はそのまま置換して出力される。
		input := "password is s3cr3t here\n"

		// Act
		n, err := m.Write([]byte(input))

		// Assert
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		// 置換でバイト数が変わっても呼び出し元へは len(p) を返す契約。
		if n != len(input) {
			t.Errorf("Write returned %d, want %d", n, len(input))
		}
		got := buf.String()
		if strings.Contains(got, "s3cr3t") {
			t.Errorf("secret leaked: %q", got)
		}
		if !strings.Contains(got, "[secret]") {
			t.Errorf("secret not masked: %q", got)
		}
	})
}

func TestMaskerRegisterEmptyIgnored(t *testing.T) {
	t.Run("Masker_空文字を登録した場合_何もマスクせず出力すること", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		m := &Masker{reg: &secretRegistry{}, out: &buf}
		m.Register("")

		// Act
		_, err := m.Write([]byte("plain text\n"))

		// Assert
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if got := buf.String(); got != "plain text\n" {
			t.Errorf("unexpected output: %q", got)
		}
	})
}

func TestMaskerLineBufferAndFlush(t *testing.T) {
	t.Run("Masker_改行が来るまで保持しFlushする場合_未改行の残りをマスクして出すこと", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		m := &Masker{reg: &secretRegistry{}, out: &buf}
		m.Register("pw")

		// Act
		// 改行が無いので保持される (まだ出力されない)。
		_, err := m.Write([]byte("auth pw"))

		// Assert
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("expected buffered (no output yet), got %q", buf.String())
		}
		// Flush で未改行の残りをマスクして出力する。
		if err := m.Flush(); err != nil {
			t.Fatalf("Flush: %v", err)
		}
		if got := buf.String(); got != "auth [secret]" {
			t.Errorf("unexpected flushed output: %q", got)
		}
	})
}

func TestMaskerChunkSplitSecret(t *testing.T) {
	t.Run("Masker_シークレットが複数Writeに分割される場合_改行までバッファし取りこぼさないこと", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		m := &Masker{reg: &secretRegistry{}, out: &buf}
		m.Register("secretpw")

		// Act
		// "secretpw" が "secret" と "pw\n" に分割されて届くケース。
		if _, err := m.Write([]byte("using secret")); err != nil {
			t.Fatalf("Write1: %v", err)
		}
		_, err := m.Write([]byte("pw now\n"))

		// Assert
		if err != nil {
			t.Fatalf("Write2: %v", err)
		}
		got := buf.String()
		if strings.Contains(got, "secretpw") {
			t.Errorf("chunk-split secret leaked: %q", got)
		}
		if got != "using [secret] now\n" {
			t.Errorf("unexpected output: %q", got)
		}
	})
}

func TestMaskerOverlappingSecrets(t *testing.T) {
	t.Run("Masker_包含関係のシークレットの場合_長い方を先に置換し部分漏洩しないこと", func(t *testing.T) {
		// Arrange
		// 包含関係にあるシークレット (abc ⊂ abc123)。
		var buf bytes.Buffer
		m := &Masker{reg: &secretRegistry{}, out: &buf}
		m.Register("abc")
		m.Register("abc123")

		// Act
		_, err := m.Write([]byte("token=abc123\n"))

		// Assert
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		got := buf.String()
		if strings.Contains(got, "123") {
			t.Errorf("longer secret partially leaked: %q", got)
		}
		if got != "token=[secret]\n" {
			t.Errorf("unexpected output: %q", got)
		}
	})
}

func TestMaskerWrapSharesRegistry(t *testing.T) {
	t.Run("Masker_Wrapで派生した場合_親子どちらの登録も両方の出力でマスクされること", func(t *testing.T) {
		// Arrange
		// Wrap で派生した Masker は同一レジストリを共有する。
		var parentBuf, childBuf bytes.Buffer
		parent := &Masker{reg: &secretRegistry{}, out: &parentBuf}
		child := parent.Wrap(&childBuf)

		// Act & Assert
		// 親で登録 → 子の出力にも効く。
		parent.Register("parentpw")
		if _, err := child.Write([]byte("using parentpw now\n")); err != nil {
			t.Fatalf("child Write: %v", err)
		}
		if strings.Contains(childBuf.String(), "parentpw") {
			t.Errorf("parent secret leaked in child: %q", childBuf.String())
		}

		// 子 (Wrap 結果) で登録 → 親の出力にも効く。
		child.Register("childpw")
		if _, err := parent.Write([]byte("using childpw now\n")); err != nil {
			t.Fatalf("parent Write: %v", err)
		}
		if strings.Contains(parentBuf.String(), "childpw") {
			t.Errorf("child secret leaked in parent: %q", parentBuf.String())
		}
	})
}
