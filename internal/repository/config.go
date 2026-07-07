package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/scenario-test-framework/stfw/assets"
	"github.com/scenario-test-framework/stfw/internal/domain/project"
)

// Config はデフォルト設定にプロジェクト設定を上書きした結果。
// Flatten した KEY=VALUE がプラグイン実行契約の env として公開される
// (v0.2 の export_yaml と同じ規則: キーを `_` で連結、リストは添字)。
type Config struct {
	flat map[string]string
}

// LoadConfig はデフォルト設定 → プロジェクト stfw.yml の順で読み込む。
// プロジェクト設定が無い場合はデフォルトのみで動作する。
func LoadConfig(projDir string) (*Config, []string, error) {
	flat := map[string]string{}
	if err := flattenYAML(assets.DefaultConfig, flat); err != nil {
		return nil, nil, fmt.Errorf("default config: %w", err)
	}

	projPath := filepath.Join(projDir, project.ConfigFileName)
	if raw, err := os.ReadFile(projPath); err == nil {
		if err := flattenYAML(raw, flat); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", projPath, err)
		}
	}

	var warns []string
	for k := range flat {
		if strings.HasPrefix(k, "stfw_server_") {
			warns = append(warns, "stfw.server.* は v1.0 で廃止されました (実行エンジン内包化により digdag server は不要です)")
			break
		}
	}
	return &Config{flat: flat}, warns, nil
}

// Get はフラット化済みキー (例: stfw_loglevel) の値を返す。
func (c *Config) Get(key string) string { return c.flat[key] }

// LogLevel は設定されたログレベル文字列を返す。
func (c *Config) LogLevel() string { return c.flat["stfw_loglevel"] }

// Environ はプラグインへ公開する KEY=VALUE リストをキー昇順で返す。
func (c *Config) Environ() []string {
	keys := make([]string, 0, len(c.flat))
	for k := range c.flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, k := range keys {
		env = append(env, k+"="+c.flat[k])
	}
	return env
}

// flattenYAML は YAML を v0.2 の export_yaml 互換規則でフラット化し dst に上書きする。
//   - map はキーを `_` で連結 (stfw.loglevel → stfw_loglevel)
//   - list は添字を付与 (stfw.webhooks.urls[0] → stfw_webhooks_urls_0)
//   - 値中の ${VAR} は環境変数で展開 (未定義は空文字。bash の source と同挙動)
func flattenYAML(raw []byte, dst map[string]string) error {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return err
	}
	flattenValue("", root, dst)
	return nil
}

func flattenValue(prefix string, v any, dst map[string]string) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			flattenValue(joinKey(prefix, k), child, dst)
		}
	case []any:
		for i, child := range val {
			flattenValue(joinKey(prefix, fmt.Sprintf("%d", i)), child, dst)
		}
	case nil:
		dst[prefix] = ""
	default:
		dst[prefix] = os.Expand(fmt.Sprintf("%v", val), func(name string) string {
			return os.Getenv(name)
		})
	}
}

func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "_" + key
}
