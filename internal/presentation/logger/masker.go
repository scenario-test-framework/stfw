package logger

import (
	"io"
	"os"
	"strings"
	"sync"
)

// secretRegistry は複数の Masker が共有するマスク対象シークレットの集合。
// Wrap で派生した Masker はこのレジストリを共有するため、いずれかの Masker で
// Register したシークレットが全経路 (ログ・プラグイン stdout/stderr 等) で
// マスクされる。
type secretRegistry struct {
	mu      sync.RWMutex
	secrets []string
}

// register はマスク対象のシークレット値を追加する。空文字は無視する。
func (r *secretRegistry) register(secret string) {
	if secret == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.secrets = append(r.secrets, secret)
}

// mask は s 中の登録済みシークレットを [secret] へ置換する。
func (r *secretRegistry) mask(s string) string {
	r.mu.RLock()
	secrets := r.secrets
	r.mu.RUnlock()

	for _, sec := range secrets {
		s = strings.ReplaceAll(s, sec, "[secret]")
	}
	return s
}

// Masker は登録されたシークレット文字列を [secret] に置換する io.Writer。
// ログ・ジャーナル・スクリプト出力キャプチャの全経路をこの Writer 経由に
// することで、マスキング実装を 1 箇所に集約する (v0.2 の log.mask 互換)。
// Wrap で生成した派生 Masker は同一のシークレットレジストリを共有する。
type Masker struct {
	reg *secretRegistry
	out io.Writer
}

// NewMasker は out をラップする Masker を生成する。
// v0.2 互換として、環境変数 PASSWORD / TOKEN が設定されていれば
// その値を初期マスク対象に登録する。
func NewMasker(out io.Writer) *Masker {
	m := &Masker{reg: &secretRegistry{}, out: out}
	for _, name := range []string{"PASSWORD", "TOKEN"} {
		if v := os.Getenv(name); v != "" {
			m.Register(v)
		}
	}
	return m
}

// Register はマスク対象のシークレット値を追加する。空文字は無視する。
// 登録は共有レジストリに対して行われるため、Wrap で派生した Masker にも反映される。
func (m *Masker) Register(secret string) {
	m.reg.register(secret)
}

// Wrap は同一のシークレットレジストリを共有したまま別の出力先 w をラップする
// Masker を返す。ロガー (stderr) とプラグイン stdout/stderr を同じシークレット集合で
// マスクするために使う。
func (m *Masker) Wrap(w io.Writer) io.Writer {
	return &Masker{reg: m.reg, out: w}
}

// Write は p 中のシークレットを [secret] へ置換して出力する。
// 置換によりバイト数が変わっても、呼び出し元へは len(p) を返す
// (io.Writer 契約上、部分書き込みと誤認させないため)。
func (m *Masker) Write(p []byte) (int, error) {
	s := m.reg.mask(string(p))
	if _, err := io.WriteString(m.out, s); err != nil {
		return 0, err
	}
	return len(p), nil
}
