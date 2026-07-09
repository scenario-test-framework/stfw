package repository

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/assets"
)

const templateRoot = "template"

// MaterializeTemplate は同梱テンプレートを projDir へ展開し、
// 作成したファイルの相対パス一覧 (昇順) を返す。
// 埋め込み (go:embed) はファイルモードを保持しないため、shebang (#!) で始まる
// ファイルには実行権限を付与する。
func MaterializeTemplate(projDir string) ([]string, error) {
	var created []string
	err := fs.WalkDir(assets.Template, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(projDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		raw, err := assets.Template.ReadFile(path)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if bytes.HasPrefix(raw, []byte("#!")) {
			mode = 0o755
		}
		if err := os.WriteFile(dest, raw, mode); err != nil {
			return err
		}
		created = append(created, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(created)
	return created, nil
}

// ProjectConfigExists はプロジェクト設定ファイルの存在を確認する。
func ProjectConfigExists(projDir, filename string) bool {
	_, err := os.Stat(filepath.Join(projDir, filename))
	return err == nil
}

// DirExists はディレクトリの存在を確認する。
func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
