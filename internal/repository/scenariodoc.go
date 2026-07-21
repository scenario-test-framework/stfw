package repository

import (
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// BuildDocFromTree は走査済みシナリオ (ScenarioView) + metadata.yml + config/config.yml から
// `stfw scenario reverse` のレンダリング用データ (scenario.DocData) を組み立てる (tree → doc の投影)。
func BuildDocFromTree(projDir string, view scenario.ScenarioView) (scenario.DocData, error) {
	scenarioDir := filepath.Join(projDir, scenario.RootDirName, view.Name)
	meta, err := ReadNodeMetadata(scenarioDir)
	if err != nil {
		return scenario.DocData{}, err
	}

	doc := scenario.DocData{Name: view.Name, Description: strings.TrimSpace(meta.Description)}

	// requirement_specifications → 検証する process のパス一覧 (要求トレーサビリティ表用)。
	trace := map[string][]string{}

	for _, b := range view.Bizdates {
		bDir := filepath.Join(scenarioDir, b.DirName)
		bMeta, err := ReadNodeMetadata(bDir)
		if err != nil {
			return scenario.DocData{}, err
		}
		title := b.DirName
		if d := firstLine(bMeta.Description); d != "" {
			title = b.DirName + " — " + d
		}
		docBizdate := scenario.DocBizdate{DirName: b.DirName, Title: title}

		for _, p := range b.Processes {
			docProcess, err := buildDocProcess(bDir, b.DirName, "", p, trace)
			if err != nil {
				return scenario.DocData{}, err
			}
			docBizdate.Processes = append(docBizdate.Processes, docProcess)
		}
		doc.Bizdates = append(doc.Bizdates, docBizdate)
	}

	reqs := make([]string, 0, len(trace))
	for req := range trace {
		reqs = append(reqs, req)
	}
	sort.Strings(reqs)
	for _, req := range reqs {
		doc.Traceability = append(doc.Traceability, scenario.TraceRow{
			RequirementSpecification: req,
			ProcessPaths:             strings.Join(trace[req], ", "),
		})
	}

	return doc, nil
}

// buildDocProcess はプロセス 1 件 (parallel の場合は子を含む) の doc データを組み立てる。
// parentRel は bizdate ディレクトリからの親プロセス相対パス (トップレベルは空)。
// 子プロセスの表示名は "{親 dir}/{子 dir}" の連結にする。
func buildDocProcess(bDir, bizdateDirName, parentRel string, p scenario.ProcessView, trace map[string][]string) (scenario.DocProcess, error) {
	rel := p.DirName
	if parentRel != "" {
		rel = path.Join(parentRel, p.DirName)
	}
	pDir := filepath.Join(bDir, filepath.FromSlash(rel))
	pMeta, err := ReadNodeMetadata(pDir)
	if err != nil {
		return scenario.DocProcess{}, err
	}
	cfg, err := ReadProcessConfigSubtree(pDir, p.ProcessType)
	if err != nil {
		return scenario.DocProcess{}, err
	}
	var cfgYAML string
	if len(cfg) > 0 {
		raw, err := yaml.Marshal(cfg)
		if err != nil {
			return scenario.DocProcess{}, err
		}
		cfgYAML = strings.TrimRight(string(raw), "\n")
	}

	docProcess := scenario.DocProcess{
		SeqLabel:                  "_" + p.Seq,
		DirName:                   rel,
		Group:                     p.Group,
		Type:                      p.ProcessType,
		Description:               firstLine(pMeta.Description),
		RequirementSpecifications: pMeta.RequirementSpecifications,
		ConfigYAML:                cfgYAML,
	}
	for _, req := range pMeta.RequirementSpecifications {
		trace[req] = append(trace[req], path.Join(bizdateDirName, rel))
	}

	for _, c := range p.Children {
		docChild, err := buildDocProcess(bDir, bizdateDirName, rel, c, trace)
		if err != nil {
			return scenario.DocProcess{}, err
		}
		docProcess.Children = append(docProcess.Children, docChild)
	}
	return docProcess, nil
}

// firstLine は s の先頭行 (前後の空白を除去) を返す。doc のテーブルセルは改行を
// 含められないため、複数行の description は先頭行のみを表示に使う。
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
