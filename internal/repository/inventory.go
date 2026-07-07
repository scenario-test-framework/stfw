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

// LoadInventory はインベントリ定義を読み込み、グループ名 → ホスト一覧の
// マップを返す。ファイル形式は v0.2 互換:
//
//	stfw_inventory:
//	  - <group-name>:
//	    - <ip | hostname>
//
// 同名グループが複数定義された場合はホストをマージする
// (v0.2 の flatten + grep と同じ挙動)。
func LoadInventory(projDir, fileName string) (map[string][]string, error) {
	path := InventoryPath(projDir, fileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}

	var doc struct {
		Inventory []map[string][]string `yaml:"stfw_inventory"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("inventory: %s: %w", path, err)
	}

	groups := map[string][]string{}
	for _, entry := range doc.Inventory {
		for group, hosts := range entry {
			groups[group] = append(groups[group], hosts...)
		}
	}
	return groups, nil
}
