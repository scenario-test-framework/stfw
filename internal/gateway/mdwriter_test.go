package gateway

import (
	"testing"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// RenderScenarioDoc の golden テスト: 固定フィクスチャ (要求トレーサビリティ・
// 複数プロセス・設定あり/なし混在) → 期待 Markdown が完全一致することを固定する。
func TestRenderScenarioDoc(t *testing.T) {
	doc := scenario.DocData{
		Name:        "daily-balance",
		Description: "日次残高バッチのシナリオテスト。",
		Traceability: []scenario.TraceRow{
			{RequirementSpecification: "SPEC-013-01", ProcessPaths: "_10_20240101/_10_arrange_clearPostgres"},
			{RequirementSpecification: "SPEC-015-01", ProcessPaths: "_10_20240101/_30_act_invokeRest, _20_20240102/_10_act_invokeRest"},
		},
		Bizdates: []scenario.DocBizdate{
			{
				DirName: "_10_20240101",
				Title:   "_10_20240101 — Day1",
				Processes: []scenario.DocProcess{
					{
						SeqLabel:                  "_10",
						DirName:                   "_10_arrange_clearPostgres",
						Phase:                     "Arrange",
						Type:                      "clearPostgres",
						Description:               "truncate",
						RequirementSpecifications: []string{"SPEC-013-01"},
						ConfigYAML:                "host_group: db\ntables:\n    - transactions\n    - accounts",
					},
					{
						SeqLabel:    "_30",
						DirName:     "_30_act_invokeRest",
						Phase:       "Act",
						Type:        "invokeRest",
						Description: "取引 POST",
					},
				},
			},
		},
	}

	want := `# シナリオ: daily-balance

日次残高バッチのシナリオテスト。

## 要求トレーサビリティ

| 要求仕様 | 検証する process |
|---|---|
| SPEC-013-01 | _10_20240101/_10_arrange_clearPostgres |
| SPEC-015-01 | _10_20240101/_30_act_invokeRest, _20_20240102/_10_act_invokeRest |

## _10_20240101 — Day1

| # | process | フェーズ(推定) | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_arrange_clearPostgres | Arrange | clearPostgres | truncate |
| _30 | _30_act_invokeRest | Act | invokeRest | 取引 POST |

### _10_arrange_clearPostgres

- フェーズ(推定): Arrange
- 要求仕様: SPEC-013-01
- 設定:

    ` + "```yaml" + `
    host_group: db
    tables:
        - transactions
        - accounts
    ` + "```" + `

### _30_act_invokeRest

- フェーズ(推定): Act
- 要求仕様: -
`

	got, err := RenderScenarioDoc(doc)
	if err != nil {
		t.Fatalf("RenderScenarioDoc: %v", err)
	}
	if got != want {
		t.Errorf("RenderScenarioDoc mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// 要求トレーサビリティ・説明のいずれも無い最小構成では、対応する節を丸ごと省略する。
func TestRenderScenarioDocMinimal(t *testing.T) {
	doc := scenario.DocData{
		Name: "empty",
		Bizdates: []scenario.DocBizdate{
			{
				DirName: "_10_20240101",
				Title:   "_10_20240101",
				Processes: []scenario.DocProcess{
					{SeqLabel: "_10", DirName: "_10_pre_scripts", Phase: "-", Type: "scripts"},
				},
			},
		},
	}

	want := `# シナリオ: empty

## _10_20240101

| # | process | フェーズ(推定) | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_pre_scripts | - | scripts |  |

### _10_pre_scripts

- フェーズ(推定): -
- 要求仕様: -
`

	got, err := RenderScenarioDoc(doc)
	if err != nil {
		t.Fatalf("RenderScenarioDoc: %v", err)
	}
	if got != want {
		t.Errorf("RenderScenarioDoc mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
