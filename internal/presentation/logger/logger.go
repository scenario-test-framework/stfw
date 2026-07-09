// Package logger は slog のセットアップとシークレットマスキングを提供する。
package logger

import (
	"io"
	"log/slog"
	"strings"
)

// LevelTrace は slog 標準に無い trace レベル (v0.2 のログレベル互換)。
const LevelTrace = slog.LevelDebug - 4

// ParseLevel は設定文字列 (trace|debug|info|warn|error) を slog.Level へ変換する。
// 不明な値は info にフォールバックする。
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New はマスキング付きロガーと、シークレット登録用の Masker を返す。
func New(out io.Writer, level slog.Level) (*slog.Logger, *Masker) {
	masker := NewMasker(out)
	handler := slog.NewTextHandler(masker, &slog.HandlerOptions{Level: level})
	return slog.New(handler), masker
}
