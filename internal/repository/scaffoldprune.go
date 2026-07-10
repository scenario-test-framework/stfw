package repository

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// PruneScenarioTree はシナリオ配下から、spec に含まれない bizdate / process
// ディレクトリを削除する (spec との差分同期。`scenario scaffold --prune` の削除相当)。
//
//   - keptBizdates: 残す bizdate ディレクトリ名の集合
//   - keptProcesses[bizdateDir]: 各 bizdate 配下で残す process ディレクトリ名の集合
//
// bizdate 規約 (ParseBizdateDirName) / process 規約 (ParseProcessDirName) に合致する
// ディレクトリのみを対象にし、規約外のファイル・ディレクトリ (README・metadata.yml 等)
// には触れない。spec に無い bizdate はサブツリーごと削除し、残す bizdate 配下では
// spec に無い process のみを削除する (実装済みの葉 data/scripts/expect も巻き込んで消える)。
// 削除したディレクトリの絶対パス一覧 (昇順) を返す。
func PruneScenarioTree(scenarioDir string, keptBizdates map[string]bool, keptProcesses map[string]map[string]bool) ([]string, error) {
	var removed []string

	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// bizdate 規約に合致しないディレクトリは同期対象外 (触れない)。
		if _, _, err := scenario.ParseBizdateDirName(e.Name()); err != nil {
			continue
		}
		bDir := filepath.Join(scenarioDir, e.Name())

		// spec に無い bizdate はサブツリーごと削除。
		if !keptBizdates[e.Name()] {
			if err := os.RemoveAll(bDir); err != nil {
				return nil, err
			}
			removed = append(removed, bDir)
			continue
		}

		// 残す bizdate 配下は、spec に無い process のみを削除。
		pEntries, err := os.ReadDir(bDir)
		if err != nil {
			return nil, err
		}
		kept := keptProcesses[e.Name()]
		for _, pe := range pEntries {
			if !pe.IsDir() {
				continue
			}
			if _, _, _, err := scenario.ParseProcessDirName(pe.Name()); err != nil {
				continue
			}
			if !kept[pe.Name()] {
				pDir := filepath.Join(bDir, pe.Name())
				if err := os.RemoveAll(pDir); err != nil {
					return nil, err
				}
				removed = append(removed, pDir)
			}
		}
	}

	sort.Strings(removed)
	return removed, nil
}
