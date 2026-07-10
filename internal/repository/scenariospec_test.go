package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMarshalUnmarshalSpecRoundtrip(t *testing.T) {
	newSpec := func() ScenarioSpec {
		return ScenarioSpec{
			Scenario:                  "daily-balance",
			Description:               "日次残高バッチのシナリオテスト\n",
			RequirementSpecifications: []string{"SPEC-000"},
			Bizdates: []BizdateSpec{
				{
					Seq:         "10",
					Bizdate:     "20240101",
					Description: "Day1",
					Processes: []ProcessSpec{
						{
							Seq:                       "10",
							Group:                     "arrange",
							Type:                      "clearPostgres",
							Description:               "truncate",
							RequirementSpecifications: []string{"SPEC-013-01"},
							Config: map[string]any{
								"host_group": "db",
								"database":   "appdb",
								"tables":     []any{"transactions", "accounts"},
							},
						},
					},
				},
			},
		}
	}

	t.Run("MarshalSpec_Marshalして再Unmarshalする場合_同一値になること", func(t *testing.T) {
		// Arrange
		spec := newSpec()

		// Act
		raw, err := MarshalSpec(spec)
		if err != nil {
			t.Fatalf("MarshalSpec: %v", err)
		}
		got, err := UnmarshalSpec(raw)

		// Assert
		if err != nil {
			t.Fatalf("UnmarshalSpec: %v", err)
		}
		if !reflect.DeepEqual(got, spec) {
			t.Errorf("roundtrip = %#v,\nwant %#v", got, spec)
		}
	})

	t.Run("MarshalSpec_同一specを2回Marshalする場合_同一バイト列になること", func(t *testing.T) {
		// Arrange
		spec := newSpec()

		// Act
		raw, err := MarshalSpec(spec)
		if err != nil {
			t.Fatalf("MarshalSpec: %v", err)
		}
		// 2 回目の Marshal でも同一バイト列になる (決定論的)。
		raw2, err := MarshalSpec(spec)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != string(raw2) {
			t.Errorf("MarshalSpec is not deterministic:\n1st=%q\n2nd=%q", raw, raw2)
		}
	})
}

// seq の先頭ゼロは NewSeq が保持する規則 (dirname.go) と同じく、spec でも失われてはならない。
func TestMarshalSpecPreservesLeadingZeroSeq(t *testing.T) {
	t.Run("MarshalSpec_seqに先頭ゼロがある場合_往復で保持されること", func(t *testing.T) {
		// Arrange
		spec := ScenarioSpec{
			Scenario: "test",
			Bizdates: []BizdateSpec{{Seq: "010", Bizdate: "20240101"}},
		}

		// Act
		raw, err := MarshalSpec(spec)
		if err != nil {
			t.Fatal(err)
		}
		got, err := UnmarshalSpec(raw)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got.Bizdates[0].Seq != "010" {
			t.Errorf("Seq = %q, want %q (leading zero must be preserved)", got.Bizdates[0].Seq, "010")
		}
	})
}

// writeFixtureFile はディレクトリを作成してからファイルを書き出すテストヘルパ。
func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildFixtureProject は BuildSpecFromTree / BuildDocFromTree のテストで共有する
// 小さなシナリオ (bizdate 1 件・process 2 件) をディスク上に組み立てる。
func buildFixtureProject(t *testing.T) string {
	t.Helper()
	projDir := t.TempDir()
	base := filepath.Join(projDir, "scenario", "sample")

	writeFixtureFile(t, filepath.Join(base, "metadata.yml"),
		"description: |\n  サンプルシナリオ\n\nrequirement_specifications:\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "metadata.yml"),
		"description: Day1\n\nrequirement_specifications:\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_10_arrange_clearPostgres", "metadata.yml"),
		"description: truncate\n\nrequirement_specifications:\n  - SPEC-013-01\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_10_arrange_clearPostgres", "config", "config.yml"),
		"stfw:\n  process:\n    clearPostgres:\n      host_group: db\n      tables:\n        - transactions\n        - accounts\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_30_act_invokeRest", "metadata.yml"),
		"description: 取引 POST\n\nrequirement_specifications:\n  - SPEC-015-01\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_30_act_invokeRest", "config", "config.yml"),
		"stfw:\n  process:\n    invokeRest: {}\n")

	return projDir
}

func TestBuildSpecFromTree(t *testing.T) {
	t.Run("BuildSpecFromTree_サンプルシナリオツリーの場合_ScenarioSpecを組み立てること", func(t *testing.T) {
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
		spec, err := BuildSpecFromTree(projDir, view)

		// Assert
		if err != nil {
			t.Fatalf("BuildSpecFromTree: %v", err)
		}
		want := ScenarioSpec{
			Scenario:    "sample",
			Description: "サンプルシナリオ\n",
			Bizdates: []BizdateSpec{
				{
					Seq:         "10",
					Bizdate:     "20240101",
					Description: "Day1",
					Processes: []ProcessSpec{
						{
							Seq:                       "10",
							Group:                     "arrange",
							Type:                      "clearPostgres",
							Description:               "truncate",
							RequirementSpecifications: []string{"SPEC-013-01"},
							Config: map[string]any{
								"host_group": "db",
								"tables":     []any{"transactions", "accounts"},
							},
						},
						{
							Seq:                       "30",
							Group:                     "act",
							Type:                      "invokeRest",
							Description:               "取引 POST",
							RequirementSpecifications: []string{"SPEC-015-01"},
						},
					},
				},
			},
		}
		if !reflect.DeepEqual(spec, want) {
			t.Errorf("BuildSpecFromTree =\n%#v,\nwant\n%#v", spec, want)
		}
	})
}

// tree → spec → tree' → spec' で spec == spec' (正規化 YAML 一致) を固定する
// (往復セマンティクスの DoD。usecase/scaffold.ScaffoldFromSpec と組み合わせた
// 完全な往復は usecase 側の TestScaffoldFromSpecRoundtrip でも検証する)。
func TestBuildSpecFromTreeRoundtripStable(t *testing.T) {
	t.Run("BuildSpecFromTree_同一treeを別ディレクトリで再構築した場合_spec正規化が一致すること", func(t *testing.T) {
		// Arrange
		projDir := buildFixtureProject(t)
		tree, err := LoadScenarioTree(projDir, []string{"sample"})
		if err != nil {
			t.Fatal(err)
		}
		view, _ := tree.ScenarioView("sample")
		spec, err := BuildSpecFromTree(projDir, view)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := MarshalSpec(spec)
		if err != nil {
			t.Fatal(err)
		}

		// Act
		// 再構築 (別プロジェクトディレクトリへ同じ tree を再現) しても同じ spec が組み立つ。
		projDir2 := buildFixtureProject(t)
		tree2, err := LoadScenarioTree(projDir2, []string{"sample"})
		if err != nil {
			t.Fatal(err)
		}
		view2, _ := tree2.ScenarioView("sample")
		spec2, err := BuildSpecFromTree(projDir2, view2)
		if err != nil {
			t.Fatal(err)
		}
		raw2, err := MarshalSpec(spec2)
		if err != nil {
			t.Fatal(err)
		}

		// Assert
		if string(raw) != string(raw2) {
			t.Errorf("spec is not stable across rebuilds:\n1st=%s\n2nd=%s", raw, raw2)
		}
	})
}
