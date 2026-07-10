package repository

import (
	"os"
	"path/filepath"
	"sort"
)

// CreateNodeScaffold は scenario / bizdate 階層の scaffold を生成する。
// ディレクトリが無ければ作成し、metadata.yml を (再) 生成する
// (v0.2 の scenario/bizdate initialize と同じ冪等な挙動。dig は生成しない)。
// 作成したファイルの絶対パス一覧を返す。
func CreateNodeScaffold(parentDir, dirName string) ([]string, error) {
	dir := filepath.Join(parentDir, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	metaPath := filepath.Join(dir, metadataFileName)
	if err := os.WriteFile(metaPath, []byte(metadataContent), 0o644); err != nil {
		return nil, err
	}
	return []string{metaPath}, nil
}

// CreateProcessScaffold はプロセス階層の scaffold を生成する。
// 既存ディレクトリは削除して作り直し、雛形と metadata.yml を生成する
// (v0.2 の process initialize と同じ挙動)。
// template/ を持つプラグイン (組込みでは scripts) はその内容を展開する。
// template/ を持たない config 駆動プラグインは、プラグインのデフォルト
// config.yml を config/config.yml として配置する。
// 作成したファイルの絶対パス一覧 (昇順) を返す。
func CreateProcessScaffold(loc PluginLocation, parentDir, dirName string) ([]string, error) {
	dir := filepath.Join(parentDir, dirName)
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	var created []string
	var err error
	if PluginHasTemplate(loc) {
		created, err = CopyPluginTemplate(loc, dir)
	} else {
		created, err = CopyPluginDefaultConfig(loc, dir)
	}
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(dir, metadataFileName)
	if err := os.WriteFile(metaPath, []byte(metadataContent), 0o644); err != nil {
		return nil, err
	}
	created = append(created, metaPath)
	sort.Strings(created)
	return created, nil
}
