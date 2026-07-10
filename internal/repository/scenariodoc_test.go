package repository

import (
	"reflect"
	"testing"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

func TestBuildDocFromTree(t *testing.T) {
	t.Run("BuildDocFromTree_サンプルシナリオツリーの場合_DocDataを組み立てること", func(t *testing.T) {
		// Arrange
		projDir := buildFixtureProject(t)
		tree, err := LoadScenarioTree(projDir, []string{"sample"})
		if err != nil {
			t.Fatalf("LoadScenarioTree: %v", err)
		}
		view, ok := tree.ScenarioView("sample")
		if !ok {
			t.Fatal("ScenarioView(sample) not found")
		}

		// Act
		doc, err := BuildDocFromTree(projDir, view)

		// Assert
		if err != nil {
			t.Fatalf("BuildDocFromTree: %v", err)
		}
		want := scenario.DocData{
			Name:        "sample",
			Description: "サンプルシナリオ",
			Traceability: []scenario.TraceRow{
				{RequirementSpecification: "SPEC-013-01", ProcessPaths: "_10_20240101/_10_arrange_clearPostgres"},
				{RequirementSpecification: "SPEC-015-01", ProcessPaths: "_10_20240101/_30_act_invokeRest"},
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
							SeqLabel:                  "_30",
							DirName:                   "_30_act_invokeRest",
							Group:                     "act",
							Type:                      "invokeRest",
							Description:               "取引 POST",
							RequirementSpecifications: []string{"SPEC-015-01"},
						},
					},
				},
			},
		}
		if !reflect.DeepEqual(doc, want) {
			t.Errorf("BuildDocFromTree =\n%#v,\nwant\n%#v", doc, want)
		}
	})
}
