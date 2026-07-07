package scenario

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// RawDir は repository が走査したディレクトリ 1 つ分の生情報。
// domain は I/O を持たないため、走査結果をこの形で受け取って ScenarioTree を組み立てる。
type RawDir struct {
	Name  string   // ディレクトリ名
	Dirs  []RawDir // 直下のサブディレクトリ
	Files []string // 直下のファイル名
}

// ScenarioTree は scenario/ 配下の走査結果を表すファーストクラスコレクション。
// 「`_` 始まりのディレクトリのみを実行対象とし、名前昇順に実行順を決定する」
// 走査規則 (v0.2 の dig_repository の grep "^_" + sort と同じ) を内包する。
type ScenarioTree struct {
	scenarios []scenarioNode
}

type scenarioNode struct {
	name     string
	raw      RawDir
	bizdates []bizdateNode // 走査規則適用済み (昇順)
}

type bizdateNode struct {
	dirName   string
	parseErr  error
	raw       RawDir
	processes []processNode // 走査規則適用済み (昇順)
}

type processNode struct {
	dirName     string
	processType string
	parseErr    error
	hasConfig   bool // config/config.yml が存在するか
	raw         RawDir
}

// NewScenarioTree は走査結果から ScenarioTree を組み立てる。
// 各階層に走査規則 (`_` 始まりのみ・名前昇順) を適用する。
func NewScenarioTree(scenarios []RawDir) *ScenarioTree {
	tree := &ScenarioTree{}
	sorted := append([]RawDir(nil), scenarios...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, s := range sorted {
		node := scenarioNode{name: s.Name, raw: s}
		for _, b := range runTargets(s.Dirs) {
			bNode := bizdateNode{dirName: b.Name, raw: b}
			if _, _, err := ParseBizdateDirName(b.Name); err != nil {
				bNode.parseErr = err
			}
			for _, p := range runTargets(b.Dirs) {
				pNode := processNode{dirName: p.Name, raw: p, hasConfig: hasConfigFile(p)}
				if _, _, processType, err := ParseProcessDirName(p.Name); err != nil {
					pNode.parseErr = err
				} else {
					pNode.processType = processType
				}
				bNode.processes = append(bNode.processes, pNode)
			}
			node.bizdates = append(node.bizdates, bNode)
		}
		tree.scenarios = append(tree.scenarios, node)
	}
	return tree
}

// Scenarios は走査済みシナリオ名を昇順で返す。
func (t *ScenarioTree) Scenarios() []string {
	names := make([]string, 0, len(t.scenarios))
	for _, s := range t.scenarios {
		names = append(names, s.name)
	}
	return names
}

// ProcessTypes は走査結果に含まれるプロセスタイプを重複なし・昇順で返す。
func (t *ScenarioTree) ProcessTypes() []string {
	seen := map[string]bool{}
	var types []string
	for _, s := range t.scenarios {
		for _, b := range s.bizdates {
			for _, p := range b.processes {
				if p.processType == "" || seen[p.processType] {
					continue
				}
				seen[p.processType] = true
				types = append(types, p.processType)
			}
		}
	}
	sort.Strings(types)
	return types
}

// Validate はディレクトリ規約の違反を列挙する。
// installedTypes は解決可能な (インストール済みの) プロセスタイプ一覧。
func (t *ScenarioTree) Validate(installedTypes []string) Violations {
	installed := map[string]bool{}
	for _, name := range installedTypes {
		installed[name] = true
	}

	var vs Violations
	for _, s := range t.scenarios {
		sPath := path.Join(RootDirName, s.name)
		for _, b := range s.bizdates {
			bPath := path.Join(sPath, b.dirName)
			if b.parseErr != nil {
				vs = append(vs, Violation{Path: bPath, Level: ViolationError, Message: b.parseErr.Error()})
				continue
			}
			for _, p := range b.processes {
				pPath := path.Join(bPath, p.dirName)
				if p.parseErr != nil {
					vs = append(vs, Violation{Path: pPath, Level: ViolationError, Message: p.parseErr.Error()})
					continue
				}
				if !installed[p.processType] {
					vs = append(vs, Violation{Path: pPath, Level: ViolationError,
						Message: fmt.Sprintf("process-plugin: %s is not installed", p.processType)})
				}
				if !p.hasConfig {
					vs = append(vs, Violation{Path: pPath, Level: ViolationError,
						Message: "config/config.yml is not exist"})
				}
			}
		}
		// 残存 *.dig の検出 (シナリオ配下を再帰的に走査)
		vs = append(vs, digViolations(path.Join(RootDirName), s.raw)...)
	}
	return vs
}

// digViolations は dir 配下に残存する *.dig ファイルを警告として列挙する。
// dig 生成は v1.0 で廃止された (validate への静的検証昇格)。
func digViolations(parent string, dir RawDir) Violations {
	cur := path.Join(parent, dir.Name)
	var vs Violations
	for _, f := range dir.Files {
		if strings.HasSuffix(f, ".dig") {
			vs = append(vs, Violation{Path: path.Join(cur, f), Level: ViolationWarn,
				Message: "*.dig は v1.0 では不要です (削除を推奨)"})
		}
	}
	for _, child := range dir.Dirs {
		vs = append(vs, digViolations(cur, child)...)
	}
	return vs
}

// runTargets は走査規則を適用する: `_` 始まりのディレクトリのみを名前昇順で返す。
func runTargets(dirs []RawDir) []RawDir {
	var targets []RawDir
	for _, d := range dirs {
		if strings.HasPrefix(d.Name, "_") {
			targets = append(targets, d)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets
}

// hasConfigFile はプロセスディレクトリ直下の config/config.yml の存在を判定する。
func hasConfigFile(process RawDir) bool {
	for _, d := range process.Dirs {
		if d.Name != "config" {
			continue
		}
		for _, f := range d.Files {
			if f == "config.yml" {
				return true
			}
		}
	}
	return false
}
