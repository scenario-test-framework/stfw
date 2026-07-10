package gateway

import (
	"testing"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// RenderScenarioDoc の golden テスト: 固定フィクスチャ (要求トレーサビリティ・
// 複数プロセス・設定あり/なし混在) → 期待 Markdown が完全一致することを固定する。
func TestRenderScenarioDoc(t *testing.T) {
	t.Run("RenderScenarioDoc_トレーサビリティと複数プロセスがある場合_期待Markdownと完全一致すること", func(t *testing.T) {
		// Arrange
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
							Group:                     "arrange",
							Type:                      "clearPostgres",
							Description:               "truncate",
							RequirementSpecifications: []string{"SPEC-013-01"},
							ConfigYAML:                "host_group: db\ntables:\n    - transactions\n    - accounts",
						},
						{
							SeqLabel:    "_30",
							DirName:     "_30_act_invokeRest",
							Group:       "act",
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

| # | process | グループ | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_arrange_clearPostgres | arrange | clearPostgres | truncate |
| _30 | _30_act_invokeRest | act | invokeRest | 取引 POST |

### _10_arrange_clearPostgres

- グループ: arrange
- 要求仕様: SPEC-013-01
- 設定:

    ` + "```yaml" + `
    host_group: db
    tables:
        - transactions
        - accounts
    ` + "```" + `

### _30_act_invokeRest

- グループ: act
- 要求仕様: -
`

		// Act
		got, err := RenderScenarioDoc(doc)

		// Assert
		if err != nil {
			t.Fatalf("RenderScenarioDoc: %v", err)
		}
		if got != want {
			t.Errorf("RenderScenarioDoc mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
		}
	})
}

// 要求トレーサビリティ・説明のいずれも無い最小構成では、対応する節を丸ごと省略する。
func TestRenderScenarioDocMinimal(t *testing.T) {
	t.Run("RenderScenarioDoc_トレーサビリティも説明も無い最小構成の場合_該当節を省略すること", func(t *testing.T) {
		// Arrange
		doc := scenario.DocData{
			Name: "empty",
			Bizdates: []scenario.DocBizdate{
				{
					DirName: "_10_20240101",
					Title:   "_10_20240101",
					Processes: []scenario.DocProcess{
						{SeqLabel: "_10", DirName: "_10_pre_scripts", Group: "pre", Type: "scripts"},
					},
				},
			},
		}

		want := `# シナリオ: empty

## _10_20240101

| # | process | グループ | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_pre_scripts | pre | scripts |  |

### _10_pre_scripts

- グループ: pre
- 要求仕様: -
`

		// Act
		got, err := RenderScenarioDoc(doc)

		// Assert
		if err != nil {
			t.Fatalf("RenderScenarioDoc: %v", err)
		}
		if got != want {
			t.Errorf("RenderScenarioDoc mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
		}
	})
}
