package gateway

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/scenario-test-framework/stfw/assets"
)

// reportTemplates は同梱の HTML レポートテンプレート (html/template + inline CSS の自己完結)。
var reportTemplates = template.Must(template.ParseFS(assets.Report, "report/*.tmpl"))

// WriteHTML はテンプレートをレンダリングして path へ書き出す。
// 実行中の run を nginx 等が同時配信しても壊れたページを見せないよう、
// 一時ファイルへ書いてから rename で置き換える。
func WriteHTML(path, tmplName string, data any) error {
	var buf bytes.Buffer
	if err := reportTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return fmt.Errorf("render %s: %w", tmplName, err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	// CreateTemp は 0600 で作成するため、別ユーザーの配信プロセス (nginx 等)
	// が読めるよう 0644 に揃えてから公開する
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}
