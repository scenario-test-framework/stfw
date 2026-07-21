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
	seq       string
	bizdate   string
	parseErr  error
	raw       RawDir
	processes []processNode // 走査規則適用済み (昇順)
}

type processNode struct {
	dirName     string
	seq         string
	group       string
	processType string
	parseErr    error
	hasConfig   bool          // config/config.yml が存在するか
	children    []processNode // parallel タイプのみ保持する子プロセス (走査規則適用済み・昇順)
	raw         RawDir
}

// ParallelProcessType は子プロセスを並走させる組込みネイティブタイプ名。
// このタイプのプロセスディレクトリのみ、配下に子プロセスディレクトリ
// (`_{seq}_{group}_{type}`) を持てる (AS-BUILT §4.14)。
const ParallelProcessType = "parallel"

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
			if seq, bizdate, err := ParseBizdateDirName(b.Name); err != nil {
				bNode.parseErr = err
			} else {
				bNode.seq = seq.String()
				bNode.bizdate = bizdate.String()
			}
			for _, p := range runTargets(b.Dirs) {
				bNode.processes = append(bNode.processes, newProcessNode(p, true))
			}
			node.bizdates = append(node.bizdates, bNode)
		}
		tree.scenarios = append(tree.scenarios, node)
	}
	return tree
}

// newProcessNode は走査結果からプロセスノードを組み立てる。
// parallel タイプのみ、配下の子プロセスを 1 段だけ構築する (子の入れ子は
// 構築せず、Validate が子タイプ = parallel を入れ子禁止エラーにする)。
func newProcessNode(p RawDir, withChildren bool) processNode {
	pNode := processNode{dirName: p.Name, raw: p, hasConfig: hasConfigFile(p)}
	if seq, group, processType, err := ParseProcessDirName(p.Name); err != nil {
		pNode.parseErr = err
	} else {
		pNode.seq = seq.String()
		pNode.group = group.String()
		pNode.processType = processType
	}
	if withChildren && pNode.processType == ParallelProcessType {
		for _, c := range runTargets(p.Dirs) {
			pNode.children = append(pNode.children, newProcessNode(c, false))
		}
	}
	return pNode
}

// Scenarios は走査済みシナリオ名を昇順で返す。
func (t *ScenarioTree) Scenarios() []string {
	names := make([]string, 0, len(t.scenarios))
	for _, s := range t.scenarios {
		names = append(names, s.name)
	}
	return names
}

// ProcessTypes は走査結果に含まれるプロセスタイプ (parallel の子を含む) を
// 重複なし・昇順で返す。
func (t *ScenarioTree) ProcessTypes() []string {
	seen := map[string]bool{}
	var types []string
	collect := func(processType string) {
		if processType == "" || seen[processType] {
			return
		}
		seen[processType] = true
		types = append(types, processType)
	}
	for _, s := range t.scenarios {
		for _, b := range s.bizdates {
			for _, p := range b.processes {
				collect(p.processType)
				for _, c := range p.children {
					collect(c.processType)
				}
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
				vs = append(vs, processViolations(path.Join(bPath, p.dirName), p, installed)...)
			}
		}
		// 残存 *.dig の検出 (シナリオ配下を再帰的に走査)
		vs = append(vs, digViolations(path.Join(RootDirName), s.raw)...)
	}
	return vs
}

// processViolations はプロセス 1 件 (parallel の場合は子を含む) の規約違反を列挙する。
func processViolations(pPath string, p processNode, installed map[string]bool) Violations {
	var vs Violations
	if p.parseErr != nil {
		return append(vs, Violation{Path: pPath, Level: ViolationError, Message: p.parseErr.Error()})
	}
	if !installed[p.processType] {
		vs = append(vs, Violation{Path: pPath, Level: ViolationError,
			Message: fmt.Sprintf("process-plugin: %s is not installed", p.processType)})
	}
	if !p.hasConfig {
		vs = append(vs, Violation{Path: pPath, Level: ViolationError,
			Message: "config/config.yml is not exist"})
	}
	if p.processType != ParallelProcessType {
		return vs
	}

	// parallel 固有の検証: 子 0 件は定義不正、子の入れ子は禁止 (AS-BUILT §4.14)
	if len(p.children) == 0 {
		vs = append(vs, Violation{Path: pPath, Level: ViolationError,
			Message: "parallel process must have at least one child process"})
	}
	for _, c := range p.children {
		cPath := path.Join(pPath, c.dirName)
		if c.parseErr != nil {
			vs = append(vs, Violation{Path: cPath, Level: ViolationError, Message: c.parseErr.Error()})
			continue
		}
		if c.processType == ParallelProcessType {
			vs = append(vs, Violation{Path: cPath, Level: ViolationError,
				Message: "parallel process can not be nested"})
			continue
		}
		vs = append(vs, processViolations(cPath, c, installed)...)
	}
	return vs
}

// StructureViolations はディレクトリ名規約 (`_{seq}_{bizdate}` / `_{seq}_{group}_{type}`) の
// parse エラーのみを列挙する。Validate とは異なり、プラグイン解決可否・config.yml 存在・
// 残存 *.dig は対象にしない (scenario reverse の投影はプラグイン未インストールでも
// 行えるべきだが、seq/group/type が空のまま壊れた doc/spec を出すのは避けたいため)。
func (t *ScenarioTree) StructureViolations() Violations {
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
				for _, c := range p.children {
					if c.parseErr != nil {
						vs = append(vs, Violation{Path: path.Join(pPath, c.dirName), Level: ViolationError, Message: c.parseErr.Error()})
					}
				}
			}
		}
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

// ScenarioView / BizdateView / ProcessView は走査規則適用済みの実行計画ビュー。
// 実行順 (昇順) を保持する。Validate が通った木でのみ完全な値を持つ。
type ScenarioView struct {
	Name     string
	Bizdates []BizdateView
}

type BizdateView struct {
	DirName   string
	Seq       string
	Bizdate   string
	Processes []ProcessView
}

type ProcessView struct {
	DirName     string
	Seq         string
	Group       string
	ProcessType string
	Children    []ProcessView // parallel タイプのみ保持する子プロセス (昇順)
}

// ScenarioViews は実行順 (名前昇順) のシナリオビューを返す。
func (t *ScenarioTree) ScenarioViews() []ScenarioView {
	views := make([]ScenarioView, 0, len(t.scenarios))
	for _, s := range t.scenarios {
		views = append(views, scenarioView(s))
	}
	return views
}

// ScenarioView は名前指定でシナリオビューを返す。
func (t *ScenarioTree) ScenarioView(name string) (ScenarioView, bool) {
	for _, s := range t.scenarios {
		if s.name == name {
			return scenarioView(s), true
		}
	}
	return ScenarioView{}, false
}

func scenarioView(s scenarioNode) ScenarioView {
	view := ScenarioView{Name: s.name}
	for _, b := range s.bizdates {
		bView := BizdateView{DirName: b.dirName, Seq: b.seq, Bizdate: b.bizdate}
		for _, p := range b.processes {
			bView.Processes = append(bView.Processes, processView(p))
		}
		view.Bizdates = append(view.Bizdates, bView)
	}
	return view
}

func processView(p processNode) ProcessView {
	view := ProcessView{
		DirName:     p.dirName,
		Seq:         p.seq,
		Group:       p.group,
		ProcessType: p.processType,
	}
	for _, c := range p.children {
		view.Children = append(view.Children, processView(c))
	}
	return view
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
