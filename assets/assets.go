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

// Plugins は同梱プラグイン (v0.2 の src/plugins から移植)。
// プラグイン解決順はプロジェクト plugins/ → 同梱の順 (v0.2 互換)。
//
//go:embed all:plugins
var Plugins embed.FS

// Report は HTML レポートのテンプレート (html/template + inline CSS の自己完結)。
//
//go:embed report
var Report embed.FS
