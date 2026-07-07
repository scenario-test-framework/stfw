package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// LoadScenarioTree は scenario/ 配下を走査して ScenarioTree を構築する。
// names を指定した場合は該当シナリオのみ、空の場合は全シナリオを対象とする。
func LoadScenarioTree(projDir string, names []string) (*scenario.ScenarioTree, error) {
	root := filepath.Join(projDir, scenario.RootDirName)
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("%s is not scenario-root-dir", root)
	}

	if len(names) == 0 {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
	}

	var raws []scenario.RawDir
	for _, name := range names {
		dir := filepath.Join(root, name)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("scenario: %s is not exist", name)
		}
		raw, err := scanDir(dir)
		if err != nil {
			return nil, err
		}
		raws = append(raws, raw)
	}
	return scenario.NewScenarioTree(raws), nil
}

// scanDir はディレクトリを再帰的に走査して RawDir を構築する。
func scanDir(dir string) (scenario.RawDir, error) {
	raw := scenario.RawDir{Name: filepath.Base(dir)}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return raw, err
	}
	for _, e := range entries {
		if e.IsDir() {
			child, err := scanDir(filepath.Join(dir, e.Name()))
			if err != nil {
				return raw, err
			}
			raw.Dirs = append(raw.Dirs, child)
			continue
		}
		raw.Files = append(raw.Files, e.Name())
	}
	return raw, nil
}
