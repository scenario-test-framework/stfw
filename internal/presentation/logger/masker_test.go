package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestMaskerRegisterAndWrite(t *testing.T) {
	var buf bytes.Buffer
	m := &Masker{reg: &secretRegistry{}, out: &buf}
	m.Register("s3cr3t")

	// 改行で終わる完了行はそのまま置換して出力される。
	input := "password is s3cr3t here\n"
	n, err := m.Write([]byte(input))
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
}

func TestMaskerRegisterEmptyIgnored(t *testing.T) {
	var buf bytes.Buffer
	m := &Masker{reg: &secretRegistry{}, out: &buf}
	m.Register("")

	if _, err := m.Write([]byte("plain text\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := buf.String(); got != "plain text\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

// 改行が来るまでは出力を保持し、Flush で未改行の残りを出す。
func TestMaskerLineBufferAndFlush(t *testing.T) {
	var buf bytes.Buffer
	m := &Masker{reg: &secretRegistry{}, out: &buf}
	m.Register("pw")

	// 改行が無いので保持される (まだ出力されない)。
	if _, err := m.Write([]byte("auth pw")); err != nil {
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
}

// シークレットが複数回の Write に分割されて届いても、改行までバッファされ取りこぼさない。
func TestMaskerChunkSplitSecret(t *testing.T) {
	var buf bytes.Buffer
	m := &Masker{reg: &secretRegistry{}, out: &buf}
	m.Register("secretpw")

	// "secretpw" が "secret" と "pw\n" に分割されて届くケース。
	if _, err := m.Write([]byte("using secret")); err != nil {
		t.Fatalf("Write1: %v", err)
	}
	if _, err := m.Write([]byte("pw now\n")); err != nil {
		t.Fatalf("Write2: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "secretpw") {
		t.Errorf("chunk-split secret leaked: %q", got)
	}
	if got != "using [secret] now\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

// 包含関係にあるシークレット (abc ⊂ abc123) でも、長い方を先に置換し部分漏洩しない。
func TestMaskerOverlappingSecrets(t *testing.T) {
	var buf bytes.Buffer
	m := &Masker{reg: &secretRegistry{}, out: &buf}
	m.Register("abc")
	m.Register("abc123")

	if _, err := m.Write([]byte("token=abc123\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "123") {
		t.Errorf("longer secret partially leaked: %q", got)
	}
	if got != "token=[secret]\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

// Wrap で派生した Masker は同一レジストリを共有するため、
// 親でも子でも Register したシークレットが両方の出力でマスクされる。
func TestMaskerWrapSharesRegistry(t *testing.T) {
	var parentBuf, childBuf bytes.Buffer
	parent := &Masker{reg: &secretRegistry{}, out: &parentBuf}
	child := parent.Wrap(&childBuf)

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
}
