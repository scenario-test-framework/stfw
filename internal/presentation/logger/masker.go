package logger

import (
	"bytes"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

// secretRegistry は複数の Masker が共有するマスク対象シークレットの集合。
// Wrap で派生した Masker はこのレジストリを共有するため、いずれかの Masker で
// Register したシークレットが全経路 (ログ・プラグイン stdout/stderr 等) で
// マスクされる。secrets は長さの降順で保持し、包含関係にある短いシークレットが
// 先に置換されて長いシークレットが部分的に漏れる事故を防ぐ。
type secretRegistry struct {
	mu      sync.RWMutex
	secrets []string
}

// register はマスク対象のシークレット値を追加する。空文字は無視する。
// 追加後、長さの降順に並べ替える (長いシークレットを先に置換するため)。
func (r *secretRegistry) register(secret string) {
	if secret == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.secrets = append(r.secrets, secret)
	sort.SliceStable(r.secrets, func(i, j int) bool {
		return len(r.secrets[i]) > len(r.secrets[j])
	})
}

// mask は s 中の登録済みシークレットを [secret] へ置換する。
// 長い順に置換するため、包含関係にあるシークレットも取りこぼさない。
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
//
// シークレットが複数回の Write に分割されて届いても取りこぼさないよう、
// 改行までを 1 単位としてバッファリングしてから置換する (シークレット値は
// 改行を含まない前提)。未改行の残りは Flush で出力する。
type Masker struct {
	reg *secretRegistry
	out io.Writer

	mu  sync.Mutex
	buf []byte
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
// マスクするために使う。派生 Masker は独自のバッファを持つため、利用終了時に
// Flush を呼んで未改行の残りを出力する必要がある。
func (m *Masker) Wrap(w io.Writer) *Masker {
	return &Masker{reg: m.reg, out: w}
}

// Write は p を蓄積し、改行までの完了行に含まれるシークレットを [secret] へ
// 置換して出力する。改行を跨いで届いたシークレットも取りこぼさない。
// 置換によりバイト数が変わっても、呼び出し元へは len(p) を返す
// (io.Writer 契約上、部分書き込みと誤認させないため)。
func (m *Masker) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buf = append(m.buf, p...)
	idx := bytes.LastIndexByte(m.buf, '\n')
	if idx < 0 {
		// 完了行が無い。改行が来る (または Flush される) まで保持する。
		return len(p), nil
	}
	masked := m.reg.mask(string(m.buf[:idx+1]))
	// 未改行の残りをバッファ先頭へ詰め直す。
	m.buf = append(m.buf[:0], m.buf[idx+1:]...)
	if _, err := io.WriteString(m.out, masked); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Flush はバッファに残った未改行のデータをマスクして出力する。
// プラグイン stdout/stderr のように末尾が改行で終わらない出力を
// 取りこぼさないため、利用終了時に呼ぶ。
func (m *Masker) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.buf) == 0 {
		return nil
	}
	masked := m.reg.mask(string(m.buf))
	m.buf = m.buf[:0]
	_, err := io.WriteString(m.out, masked)
	return err
}
