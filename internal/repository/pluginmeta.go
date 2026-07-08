package repository

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/scenario-test-framework/stfw/assets"
	"github.com/scenario-test-framework/stfw/internal/gateway"
)

// pluginMetaFileName はプラグインのランタイム依存を宣言するメタデータファイル名。
const pluginMetaFileName = "plugin.yml"

// PluginMeta はプラグインメタデータ (plugin.yml) の内容。
//
//	requires:          # このプラグインが前提とするコマンド (PATH 上の存在を検証する)
//	  - mysql
//	  - mysqldump
type PluginMeta struct {
	Requires []string `yaml:"requires"`
}

// LoadPluginMeta はプラグインの plugin.yml を読み込む。
// メタデータファイルが存在しない場合は空の PluginMeta を返す (依存宣言なし)。
func LoadPluginMeta(loc PluginLocation) (PluginMeta, error) {
	var raw []byte
	var err error
	if loc.Embedded {
		raw, err = fs.ReadFile(assets.Plugins, path.Join(loc.EmbedPath, pluginMetaFileName))
	} else {
		raw, err = os.ReadFile(filepath.Join(loc.Dir, pluginMetaFileName))
	}
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return PluginMeta{}, nil
		}
		return PluginMeta{}, fmt.Errorf("plugin meta: %w", err)
	}

	var meta PluginMeta
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return PluginMeta{}, fmt.Errorf("plugin meta: %s: %w", pluginMetaFileName, err)
	}
	return meta, nil
}

// MissingRequire は満たされていないランタイム依存 1 件。
type MissingRequire struct {
	ProcessType string
	Command     string
}

// CheckPluginRequires はシナリオで使用する各プロセスタイプのプラグインについて、
// plugin.yml の requires が PATH 上に存在するかを検証し、欠落を返す
// (条件「プラグインのランタイム依存宣言と存在チェック」)。
// プラグインを解決できないプロセスタイプはスキップする (構造検証側で検出済み)。
func CheckPluginRequires(projDir string, processTypes []string) ([]MissingRequire, error) {
	var missing []MissingRequire
	for _, pt := range processTypes {
		loc, err := ResolveProcessPlugin(projDir, pt)
		if err != nil {
			continue
		}
		meta, err := LoadPluginMeta(loc)
		if err != nil {
			return nil, err
		}
		for _, cmd := range meta.Requires {
			if !gateway.CommandExists(cmd) {
				missing = append(missing, MissingRequire{ProcessType: pt, Command: cmd})
			}
		}
	}
	return missing, nil
}
