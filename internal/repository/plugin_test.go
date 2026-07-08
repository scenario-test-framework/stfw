package repository

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveProcessPluginEmptyType は空のプロセスタイプが plugins/process
// ディレクトリ自体に解決されないこと (parse error プロセスへの防御) を固定する。
func TestResolveProcessPluginEmptyType(t *testing.T) {
	projDir := t.TempDir()
	// plugins/process ディレクトリを実在させる (空タイプの filepath.Join が
	// ここに解決してしまう退行を検出するため)。
	if err := os.MkdirAll(filepath.Join(projDir, "plugins", "process"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveProcessPlugin(projDir, ""); err == nil {
		t.Fatal("空のプロセスタイプは解決エラーになるべき (plugins/process への誤解決を防ぐ)")
	}
}
