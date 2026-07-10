package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// processConfigFileName はプロセスディレクトリ直下の設定ファイル名。
const processConfigFileName = "config.yml"

// ReadProcessConfigSubtree は config/config.yml から `stfw.process.{type}` サブツリーを
// 生の map として読む。既存の config リポジトリ (config.go) はプラグイン実行時の env 用に
// フラット化してしまうため、spec / doc の投影にはそのままの構造を読む経路が別途必要になる。
// ファイル不在・キー欠落・明示的な null (`stfw.process.{type}:` を空のまま) は
// 空 (nil) を返す (spec export は歯抜けの config も許容する)。一方 `stfw.process.{type}` が
// 存在し値も入っているのに mapping でない (list/scalar/string) 場合は、config を silent
// drop せずエラーにする (往復忠実性: 中身を黙って消すと spec export が壊れた出力になるため)。
func ReadProcessConfigSubtree(processDir, processType string) (map[string]any, error) {
	path := filepath.Join(processDir, "config", processConfigFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	// stfw / stfw.process の祖先コンテナも、非 nil かつ非 map なら silent drop せずエラーに
	// する (leaf だけでなく祖先が破損していても spec export を壊れた空設定にしないため)。
	stfw, err := mapField(root, "stfw", "stfw", path)
	if err != nil {
		return nil, err
	}
	process, err := mapField(stfw, "process", "stfw.process", path)
	if err != nil {
		return nil, err
	}

	val, exists := process[processType]
	if !exists || val == nil {
		return nil, nil
	}
	sub, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: stfw.process.%s must be a mapping (got %T)", path, processType, val)
	}
	if len(sub) == 0 {
		// 空 map と未設定 (nil) を同一視する (往復での spec.Config の同一性を保つため)。
		return nil, nil
	}
	return sub, nil
}

// mapField は m[key] を map として取り出す。欠落・null は (nil, nil)、非 nil かつ非 map は
// エラーにする (config の破損を silent drop しないため)。m 自体が nil でも安全 (nil map の
// インデックスはゼロ値を返す)。label はエラーメッセージ用のドット区切りキー。
func mapField(m map[string]any, key, label, path string) (map[string]any, error) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, nil
	}
	sub, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: %s must be a mapping (got %T)", path, label, v)
	}
	return sub, nil
}

// WriteProcessConfig は processDir/config/config.yml へ `stfw.process.{type}` サブツリーを
// 書き出す (`scenario scaffold` の入口)。cfg が空の場合は空スタブ (`{}`) を書く
// (往復の決定性を優先する判断。プラグイン既定値での穴埋めはしない。詳細は AS-BUILT.md 参照)。
func WriteProcessConfig(processDir, processType string, cfg map[string]any) error {
	dir := filepath.Join(processDir, "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if cfg == nil {
		cfg = map[string]any{}
	}
	doc := map[string]any{
		"stfw": map[string]any{
			"process": map[string]any{
				processType: cfg,
			},
		},
	}
	raw, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, processConfigFileName), raw, 0o644)
}
