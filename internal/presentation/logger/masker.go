package logger

import (
	"io"
	"os"
	"strings"
	"sync"
)

// Masker は登録されたシークレット文字列を [secret] に置換する io.Writer。
// ログ・ジャーナル・スクリプト出力キャプチャの全経路をこの Writer 経由に
// することで、マスキング実装を 1 箇所に集約する (v0.2 の log.mask 互換)。
type Masker struct {
	mu      sync.RWMutex
	secrets []string
	out     io.Writer
}

// NewMasker は out をラップする Masker を生成する。
// v0.2 互換として、環境変数 PASSWORD / TOKEN が設定されていれば
// その値を初期マスク対象に登録する。
func NewMasker(out io.Writer) *Masker {
	m := &Masker{out: out}
	for _, name := range []string{"PASSWORD", "TOKEN"} {
		if v := os.Getenv(name); v != "" {
			m.Register(v)
		}
	}
	return m
}

// Register はマスク対象のシークレット値を追加する。空文字は無視する。
func (m *Masker) Register(secret string) {
	if secret == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets = append(m.secrets, secret)
}

// Write は p 中のシークレットを [secret] へ置換して出力する。
// 置換によりバイト数が変わっても、呼び出し元へは len(p) を返す
// (io.Writer 契約上、部分書き込みと誤認させないため)。
func (m *Masker) Write(p []byte) (int, error) {
	m.mu.RLock()
	secrets := m.secrets
	m.mu.RUnlock()

	s := string(p)
	for _, sec := range secrets {
		s = strings.ReplaceAll(s, sec, "[secret]")
	}
	if _, err := io.WriteString(m.out, s); err != nil {
		return 0, err
	}
	return len(p), nil
}
