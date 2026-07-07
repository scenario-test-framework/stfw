// Package assets は stfw バイナリに同梱する静的資産を保持する。
package assets

import "embed"

// Template は stfw init で展開するプロジェクトテンプレート。
//
//go:embed all:template
var Template embed.FS

// DefaultConfig はデフォルト設定 (プロジェクト stfw.yml で上書きされる)。
//
//go:embed config/stfw.yml
var DefaultConfig []byte
