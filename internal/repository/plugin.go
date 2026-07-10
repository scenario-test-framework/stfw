package repository

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/scenario-test-framework/stfw/assets"
	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/gateway"
)

// embeddedPluginRoot は同梱プラグインの embed FS 上のルート。
const embeddedPluginRoot = "plugins"

// PluginCacheDir はプラグインが provisioning した資産 (install でダウンロード
// したバイナリ等) を置く永続キャッシュディレクトリを返す。
// MaterializePlugin が毎回ワイプする .stfw/plugins/ とは別に、実行をまたいで
// 保持する必要がある資産のための場所 (例: collectLog の logfilter バイナリ)。
func PluginCacheDir(projDir, processType string) string {
	return filepath.Join(projDir, project.DataDirName, "cache", "plugins", processType)
}

// PluginLocation は解決済みプラグインの所在。
type PluginLocation struct {
	// Dir はプロジェクトプラグインのディスクパス (Embedded=false のとき有効)。
	Dir string
	// EmbedPath は同梱プラグインの embed FS 上のパス (Embedded=true のとき有効)。
	EmbedPath string
	// Embedded は同梱プラグインかどうか。
	Embedded bool
}

// ResolveProcessPlugin はプロセスプラグインを解決する。
// 解決順はプロジェクト plugins/ → 同梱 assets/plugins/ (v0.2 の
// stfw.get_installed_plugin_path と同じ順序)。
func ResolveProcessPlugin(projDir, processType string) (PluginLocation, error) {
	// 空のプロセスタイプは filepath.Join で plugins/process ディレクトリ自体に
	// 解決してしまうため明示的に弾く (ディレクトリ名 parse error のプロセスが
	// processType="" のまま渡ってくるケースへの防御)。
	if processType == "" {
		return PluginLocation{}, fmt.Errorf("process-plugin: empty process type")
	}
	projPlugin := filepath.Join(projDir, "plugins", "process", processType)
	if info, err := os.Stat(projPlugin); err == nil && info.IsDir() {
		return PluginLocation{Dir: projPlugin}, nil
	}

	embedPath := path.Join(embeddedPluginRoot, "process", processType)
	if info, err := fs.Stat(assets.Plugins, embedPath); err == nil && info.IsDir() {
		return PluginLocation{EmbedPath: embedPath, Embedded: true}, nil
	}

	return PluginLocation{}, fmt.Errorf("process-plugin: %s is not installed", processType)
}

// ListProcessPlugins はプロセスプラグイン名を重複なし・昇順で返す。
// 同梱 + プロジェクトの和集合から `_` 始まり (共通処理) を除外する
// (v0.2 の process_repository.list と同じ規則)。
func ListProcessPlugins(projDir string) ([]string, error) {
	seen := map[string]bool{}

	entries, err := fs.ReadDir(assets.Plugins, path.Join(embeddedPluginRoot, "process"))
	if err != nil {
		return nil, fmt.Errorf("embedded plugins: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			seen[e.Name()] = true
		}
	}

	projRoot := filepath.Join(projDir, "plugins", "process")
	if projEntries, err := os.ReadDir(projRoot); err == nil {
		for _, e := range projEntries {
			if e.IsDir() {
				seen[e.Name()] = true
			}
		}
	}

	var names []string
	for name := range seen {
		if strings.HasPrefix(name, "_") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// MaterializePlugin はプラグインの実体をディスク上に確保し、そのパスを返す。
// プロジェクトプラグインはそのままのパス、同梱プラグインは
// .stfw/plugins/ 配下へ展開したパスを返す (スクリプト実行に必要)。
func MaterializePlugin(projDir string, loc PluginLocation) (string, error) {
	if !loc.Embedded {
		return loc.Dir, nil
	}
	dest := filepath.Join(projDir, project.DataDirName, loc.EmbedPath)
	if err := os.RemoveAll(dest); err != nil {
		return "", err
	}
	if _, err := copyFSTree(assets.Plugins, loc.EmbedPath, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// CopyPluginTemplate はプラグインの template/ 配下を destDir へコピーし、
// 作成したファイルの絶対パス一覧 (昇順) を返す。
func CopyPluginTemplate(loc PluginLocation, destDir string) ([]string, error) {
	if loc.Embedded {
		src := path.Join(loc.EmbedPath, "template")
		if _, err := fs.Stat(assets.Plugins, src); err != nil {
			return nil, fmt.Errorf("plugin template: %s is not exist", src)
		}
		return copyFSTree(assets.Plugins, src, destDir)
	}

	src := filepath.Join(loc.Dir, "template")
	if info, err := os.Stat(src); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("plugin template: %s is not exist", src)
	}
	return copyFSTree(os.DirFS(src), ".", destDir)
}

// PluginHasTemplate はプラグインが template/ ディレクトリを持つかを返す。
// 組込みプラグインでは scripts のみが template/ を持ち、config 駆動の
// プラグイン (clearPostgres 等) は持たない。
func PluginHasTemplate(loc PluginLocation) bool {
	if loc.Embedded {
		info, err := fs.Stat(assets.Plugins, path.Join(loc.EmbedPath, "template"))
		return err == nil && info.IsDir()
	}
	info, err := os.Stat(filepath.Join(loc.Dir, "template"))
	return err == nil && info.IsDir()
}

// CopyPluginDefaultConfig はプラグインのデフォルト config.yml を
// destDir/config/config.yml へコピーし、作成したファイルの絶対パス一覧を返す。
// template/ を持たない config 駆動プラグインの雛形として使う
// (プロセスの config/config.yml があれば validate の上書き対象になる)。
// プラグインに config.yml が無い場合は何も作らず (nil, nil) を返す。
func CopyPluginDefaultConfig(loc PluginLocation, destDir string) ([]string, error) {
	var raw []byte
	var err error
	if loc.Embedded {
		raw, err = fs.ReadFile(assets.Plugins, path.Join(loc.EmbedPath, "config.yml"))
	} else {
		raw, err = os.ReadFile(filepath.Join(loc.Dir, "config.yml"))
	}
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("plugin config: %w", err)
	}

	confDir := filepath.Join(destDir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		return nil, err
	}
	dest := filepath.Join(confDir, "config.yml")
	if err := os.WriteFile(dest, raw, 0o644); err != nil {
		return nil, err
	}
	return []string{dest}, nil
}

// ProcessConfigEnv はプロセス実行時に注入するプラグイン設定の env を返す。
// 優先順 (後勝ち) はプラグイン config.yml → プロジェクト
// config/plugins/process/{type}/config.yml → プロセス config/config.yml
// (v0.2 の process_service.private.export_config + scripts プラグイン execute の
// export_yaml と同じ上書きチェーン)。
func ProcessConfigEnv(projDir string, loc PluginLocation, processType, processDir string) (map[string]string, error) {
	flat, err := PluginConfigEnv(projDir, loc, processType)
	if err != nil {
		return nil, err
	}
	// プロセス設定
	if err := flattenYAMLFile(filepath.Join(processDir, "config", "config.yml"), flat); err != nil {
		return nil, err
	}
	return flat, nil
}

// PluginConfigEnv はプロセス非依存のプラグイン設定 env を返す。
// プラグイン config.yml → プロジェクト config/plugins/process/{type}/config.yml の
// 上書きチェーン (プロセスの config/config.yml は含まない)。
// install / is_installed のプロビジョニングは特定プロセスに紐づかないため、
// この段までの設定 (例: collectLog の logfilter_version / logfilter_arches) を使う。
func PluginConfigEnv(projDir string, loc PluginLocation, processType string) (map[string]string, error) {
	flat := map[string]string{}

	// プラグイン設定
	if loc.Embedded {
		if raw, err := fs.ReadFile(assets.Plugins, path.Join(loc.EmbedPath, "config.yml")); err == nil {
			if err := flattenYAML(raw, flat); err != nil {
				return nil, fmt.Errorf("%s/config.yml: %w", loc.EmbedPath, err)
			}
		}
	} else {
		if err := flattenYAMLFile(filepath.Join(loc.Dir, "config.yml"), flat); err != nil {
			return nil, err
		}
	}

	// プロジェクト上書き
	projConf := filepath.Join(projDir, "config", "plugins", "process", processType, "config.yml")
	if err := flattenYAMLFile(projConf, flat); err != nil {
		return nil, err
	}
	return flat, nil
}

// IsPluginInstalled はプラグインの bin/install/is_installed を実行して
// インストール済みかを判定する。標準出力が "true" の場合のみ true
// (v0.2 の process_spec.is_installed と同じ判定)。
// スクリプトが無い・実行できない場合は未インストール扱いとする。
func IsPluginInstalled(pluginDir string, env []string, errOut io.Writer) bool {
	script := filepath.Join(pluginDir, "bin", "install", "is_installed")
	var stdout bytes.Buffer
	code, err := gateway.RunScript(pluginDir, script, env, &stdout, errOut)
	if err != nil || code != 0 {
		return false
	}
	return strings.TrimSpace(stdout.String()) == "true"
}

// RunPluginInstall はプラグインの bin/install/install を実行し、終了コードを返す。
func RunPluginInstall(pluginDir string, env []string, out, errOut io.Writer) (int, error) {
	script := filepath.Join(pluginDir, "bin", "install", "install")
	return gateway.RunScript(pluginDir, script, env, out, errOut)
}

// copyFSTree は fsys の root 配下を destDir へコピーし、
// 作成したファイルの絶対パス一覧 (昇順) を返す。
// 埋め込み (go:embed) はファイルモードを保持しないため、shebang (#!) で始まる
// ファイルには実行権限を付与する。
func copyFSTree(fsys fs.FS, root, destDir string) ([]string, error) {
	var created []string
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if bytes.HasPrefix(raw, []byte("#!")) {
			mode = 0o755
		}
		if info, err := d.Info(); err == nil && info.Mode()&0o111 != 0 {
			mode = 0o755
		}
		if err := os.WriteFile(dest, raw, mode); err != nil {
			return err
		}
		created = append(created, dest)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(created)
	return created, nil
}
