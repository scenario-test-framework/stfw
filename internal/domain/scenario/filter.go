package scenario

import (
	"fmt"
	"strings"
)

// RunFilter は stfw run の部分実行 (--from / --only) を表す値オブジェクト。
// パスはシナリオ相対 `{bizdate_dirname}[/{process_dirname}]` (ディレクトリ名の完全一致)。
// from は「指定ノードより実行順で前をスキップ」、only は「指定サブツリーのみ実行」
// (AS-BUILT §3.4)。ゼロ値はフィルタなし (全ノード実行)。
type RunFilter struct {
	mode    filterMode
	bizdate string
	process string // 空なら bizdate 階層までの指定
}

type filterMode string

const (
	filterNone filterMode = ""
	filterFrom filterMode = "from"
	filterOnly filterMode = "only"
)

// NewRunFilter は --from / --only の指定から RunFilter を組み立てる。
// 両方空はフィルタなし。両方指定は排他エラー。
func NewRunFilter(from, only string) (RunFilter, error) {
	if from != "" && only != "" {
		return RunFilter{}, fmt.Errorf("--from and --only are mutually exclusive")
	}
	if from != "" {
		return parseFilterPath(filterFrom, from)
	}
	if only != "" {
		return parseFilterPath(filterOnly, only)
	}
	return RunFilter{}, nil
}

// parseFilterPath はフィルタパスを `{bizdate_dir}[/{process_dir}]` としてパースする。
func parseFilterPath(mode filterMode, path string) (RunFilter, error) {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return RunFilter{}, fmt.Errorf("--%s: %q is not {bizdate_dir}[/{process_dir}] format", mode, path)
	}
	for _, p := range parts {
		if p == "" {
			return RunFilter{}, fmt.Errorf("--%s: %q is not {bizdate_dir}[/{process_dir}] format", mode, path)
		}
	}
	f := RunFilter{mode: mode, bizdate: parts[0]}
	if len(parts) == 2 {
		f.process = parts[1]
	}
	return f, nil
}

// Active はフィルタが指定されているかを返す。
func (f RunFilter) Active() bool {
	return f.mode != filterNone
}

// Attr はジャーナル run ノードへ記録する属性キー (`from` / `only`) とパスを返す。
// 未指定時は両方空文字 (attrs へ記録しない)。
func (f RunFilter) Attr() (key, value string) {
	if !f.Active() {
		return "", ""
	}
	value = f.bizdate
	if f.process != "" {
		value += "/" + f.process
	}
	return string(f.mode), value
}

// Apply は view へフィルタを適用した実行計画を返す。
// 実行順は変えず実行対象のみ絞り込む。指定ノードが存在しない場合はエラー
// (呼び出し側で run 開始前の fail-fast に使う)。
func (f RunFilter) Apply(view ScenarioView) (ScenarioView, error) {
	if !f.Active() {
		return view, nil
	}
	bi := -1
	for i, b := range view.Bizdates {
		if b.DirName == f.bizdate {
			bi = i
			break
		}
	}
	if bi < 0 {
		return ScenarioView{}, fmt.Errorf("--%s: bizdate directory not found: %s/%s",
			f.mode, view.Name, f.bizdate)
	}

	if f.mode == filterOnly {
		bizdate := view.Bizdates[bi]
		if f.process != "" {
			pi, err := f.processIndex(view.Name, bizdate)
			if err != nil {
				return ScenarioView{}, err
			}
			bizdate.Processes = []ProcessView{bizdate.Processes[pi]}
		}
		return ScenarioView{Name: view.Name, Bizdates: []BizdateView{bizdate}}, nil
	}

	// from: 指定 bizdate 以降を実行。process まで指定された場合は
	// 先頭 bizdate 内の先行 process のみスキップし、後続 bizdate は全 process を実行する。
	bizdates := append([]BizdateView(nil), view.Bizdates[bi:]...)
	if f.process != "" {
		pi, err := f.processIndex(view.Name, bizdates[0])
		if err != nil {
			return ScenarioView{}, err
		}
		bizdates[0].Processes = bizdates[0].Processes[pi:]
	}
	return ScenarioView{Name: view.Name, Bizdates: bizdates}, nil
}

// processIndex は bizdate 内のプロセス位置をディレクトリ名の完全一致で解決する。
func (f RunFilter) processIndex(scenarioName string, bizdate BizdateView) (int, error) {
	for i, p := range bizdate.Processes {
		if p.DirName == f.process {
			return i, nil
		}
	}
	return -1, fmt.Errorf("--%s: process directory not found: %s/%s/%s",
		f.mode, scenarioName, f.bizdate, f.process)
}
