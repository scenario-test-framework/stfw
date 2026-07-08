package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// InventoryPath はインベントリファイル (config/inventory/{fileName}) のパスを返す。
func InventoryPath(projDir, fileName string) string {
	return filepath.Join(projDir, "config", "inventory", fileName)
}

// invHost はインベントリのホストエントリ 1 件。
// 後方互換のため、YAML では文字列 (ホスト名のみ) と
// マップ (host + arch 等の接続メタデータ) の両形式を受理する。
//
//	stfw_inventory:
//	  - web:
//	    - 127.0.0.1              # 文字列形式 (arch 未指定)
//	  - db:
//	    - host: db1.example      # マップ形式
//	      arch: linux_amd64      # 収集系プラグインのバイナリ送り分け用
type invHost struct {
	Name string
	Arch string
}

// UnmarshalYAML は文字列・マップ両形式のホストエントリを受理する。
func (h *invHost) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&h.Name)
	}
	var m struct {
		Host string `yaml:"host"`
		Arch string `yaml:"arch"`
	}
	if err := value.Decode(&m); err != nil {
		return err
	}
	h.Name = m.Host
	h.Arch = m.Arch
	return nil
}

// loadInventoryEntries はインベントリ定義をグループ名 → ホストエントリ一覧へ
// パースする。同名グループが複数定義された場合はエントリをマージする。
func loadInventoryEntries(projDir, fileName string) (map[string][]invHost, error) {
	path := InventoryPath(projDir, fileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}

	var doc struct {
		Inventory []map[string][]invHost `yaml:"stfw_inventory"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("inventory: %s: %w", path, err)
	}

	groups := map[string][]invHost{}
	for _, entry := range doc.Inventory {
		for group, hosts := range entry {
			groups[group] = append(groups[group], hosts...)
		}
	}
	return groups, nil
}

// LoadInventory はインベントリ定義を読み込み、グループ名 → ホスト名一覧の
// マップを返す (v0.2 互換。arch 等のメタデータは含まない)。
func LoadInventory(projDir, fileName string) (map[string][]string, error) {
	entries, err := loadInventoryEntries(projDir, fileName)
	if err != nil {
		return nil, err
	}
	groups := make(map[string][]string, len(entries))
	for group, hosts := range entries {
		names := make([]string, 0, len(hosts))
		for _, h := range hosts {
			names = append(names, h.Name)
		}
		groups[group] = names
	}
	return groups, nil
}

// LoadInventoryHostArch はホスト名 → arch のマップを返す。
// arch 未指定のホストは含めない。同一ホストが複数グループに属していても
// arch が一致していれば許容するが、異なる arch が定義された場合は設定の
// 不整合としてエラーにする (map 走査順に依存しない決定的な結果にするため)。
func LoadInventoryHostArch(projDir, fileName string) (map[string]string, error) {
	entries, err := loadInventoryEntries(projDir, fileName)
	if err != nil {
		return nil, err
	}
	hostArch := map[string]string{}
	for _, hosts := range entries {
		for _, h := range hosts {
			if h.Name == "" || h.Arch == "" {
				continue
			}
			if prev, ok := hostArch[h.Name]; ok && prev != h.Arch {
				return nil, fmt.Errorf("inventory: host %q has conflicting arch: %q vs %q", h.Name, prev, h.Arch)
			}
			hostArch[h.Name] = h.Arch
		}
	}
	return hostArch, nil
}
