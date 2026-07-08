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

	if _, err := m.Write([]byte("plain text")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := buf.String(); got != "plain text" {
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
	if _, err := child.Write([]byte("using parentpw now")); err != nil {
		t.Fatalf("child Write: %v", err)
	}
	if strings.Contains(childBuf.String(), "parentpw") {
		t.Errorf("parent secret leaked in child: %q", childBuf.String())
	}

	// 子 (Wrap 結果) で登録 → 親の出力にも効く。
	if cm, ok := child.(*Masker); ok {
		cm.Register("childpw")
	} else {
		t.Fatalf("Wrap did not return *Masker")
	}
	if _, err := parent.Write([]byte("using childpw now")); err != nil {
		t.Fatalf("parent Write: %v", err)
	}
	if strings.Contains(parentBuf.String(), "childpw") {
		t.Errorf("child secret leaked in parent: %q", parentBuf.String())
	}
}
