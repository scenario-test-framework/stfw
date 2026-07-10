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
			pDir := filepath.Join(bDir, p.DirName)
			pMeta, err := ReadNodeMetadata(pDir)
			if err != nil {
				return scenario.DocData{}, err
			}
			cfg, err := ReadProcessConfigSubtree(pDir, p.ProcessType)
			if err != nil {
				return scenario.DocData{}, err
			}
			var cfgYAML string
			if len(cfg) > 0 {
				raw, err := yaml.Marshal(cfg)
				if err != nil {
					return scenario.DocData{}, err
				}
				cfgYAML = strings.TrimRight(string(raw), "\n")
			}

			docBizdate.Processes = append(docBizdate.Processes, scenario.DocProcess{
				SeqLabel:                  "_" + p.Seq,
				DirName:                   p.DirName,
				Group:                     p.Group,
				Type:                      p.ProcessType,
				Description:               firstLine(pMeta.Description),
				RequirementSpecifications: pMeta.RequirementSpecifications,
				ConfigYAML:                cfgYAML,
			})

			for _, req := range pMeta.RequirementSpecifications {
				trace[req] = append(trace[req], path.Join(b.DirName, p.DirName))
			}
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

// firstLine は s の先頭行 (前後の空白を除去) を返す。doc のテーブルセルは改行を
// 含められないため、複数行の description は先頭行のみを表示に使う。
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
