package scaffold

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/scenario-test-framework/stfw/internal/repository"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestProject は ScaffoldFromSpec のガード条件 (stfw.yml + scenario/ の存在) を
// 満たす最小プロジェクトディレクトリを t.TempDir() 配下に作る。
func newTestProject(t *testing.T) string {
	t.Helper()
	projDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projDir, "stfw.yml"), []byte("stfw:\n  project_version: 0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projDir, "scenario"), 0o755); err != nil {
		t.Fatal(err)
	}
	return projDir
}

func sampleSpec() repository.ScenarioSpec {
	return repository.ScenarioSpec{
		Scenario:    "daily-balance",
		Description: "日次残高バッチのシナリオテスト",
		Bizdates: []repository.BizdateSpec{
			{
				Seq:         "10",
				Bizdate:     "20240101",
				Description: "Day1",
				Processes: []repository.ProcessSpec{
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
						Seq:   "30",
						Group: "act",
						Type:  "invokeRest",
					},
				},
			},
		},
	}
}

func TestScaffoldFromSpec(t *testing.T) {
	t.Run("ScaffoldFromSpec_正常なspecの場合_ディレクトリとconfigを生成すること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), false)

		// Assert
		if err != nil {
			t.Fatalf("ScaffoldFromSpec: %v", err)
		}

		scenarioDir := filepath.Join(projDir, "scenario", "daily-balance")
		meta, err := repository.ReadNodeMetadata(scenarioDir)
		if err != nil {
			t.Fatal(err)
		}
		if meta.Description != "日次残高バッチのシナリオテスト" {
			t.Errorf("scenario description = %q", meta.Description)
		}

		processDir := filepath.Join(scenarioDir, "_10_20240101", "_10_arrange_clearPostgres")
		if _, err := os.Stat(filepath.Join(processDir, "metadata.yml")); err != nil {
			t.Errorf("process metadata.yml not created: %v", err)
		}
		cfg, err := repository.ReadProcessConfigSubtree(processDir, "clearPostgres")
		if err != nil {
			t.Fatal(err)
		}
		wantCfg := map[string]any{"host_group": "db", "tables": []any{"transactions", "accounts"}}
		if !reflect.DeepEqual(cfg, wantCfg) {
			t.Errorf("process config = %#v, want %#v", cfg, wantCfg)
		}

		// data/scripts/expect 等の葉は生成しない (§0 の往復境界)。
		if _, err := os.Stat(filepath.Join(processDir, "scripts")); !os.IsNotExist(err) {
			t.Errorf("scaffold must not generate scripts/ (leaf): err=%v", err)
		}

		// config 未指定のプロセスは空スタブになる (プラグイン既定値での穴埋めはしない)。
		process2Dir := filepath.Join(scenarioDir, "_10_20240101", "_30_act_invokeRest")
		cfg2, err := repository.ReadProcessConfigSubtree(process2Dir, "invokeRest")
		if err != nil {
			t.Fatal(err)
		}
		if cfg2 != nil {
			t.Errorf("process2 config = %#v, want nil (empty stub)", cfg2)
		}
	})
}

// 既存シナリオディレクトリがあると既定 (sync=false) はエラーになる。
func TestScaffoldFromSpecExistingDirWithoutSync(t *testing.T) {
	t.Run("ScaffoldFromSpec_既存ディレクトリがありsyncfalse の場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer
		if err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), false); err != nil {
			t.Fatalf("1st ScaffoldFromSpec: %v", err)
		}

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), false)

		// Assert
		if err == nil {
			t.Fatal("2nd ScaffoldFromSpec (sync=false) should fail")
		}
	})
}

// --sync は既存ディレクトリを更新し、spec にある process 配下に手動で追加した葉
// (data/scripts/expect) は温存する (spec に無い process のみを削除するため)。
func TestScaffoldFromSpecSyncPreservesLeavesOfKeptProcess(t *testing.T) {
	t.Run("ScaffoldFromSpec_syncで残るprocess配下に手動の葉がある場合_葉を温存し説明を更新すること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer
		if err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), false); err != nil {
			t.Fatalf("1st ScaffoldFromSpec: %v", err)
		}

		processDir := filepath.Join(projDir, "scenario", "daily-balance", "_10_20240101", "_10_arrange_clearPostgres")
		leaf := filepath.Join(processDir, "scripts", "100_manual")
		if err := os.MkdirAll(filepath.Dir(leaf), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(leaf, []byte("#!/bin/bash\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		spec := sampleSpec()
		spec.Description = "updated description"

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, spec, true)

		// Assert
		if err != nil {
			t.Fatalf("2nd ScaffoldFromSpec (sync=true): %v", err)
		}

		if _, err := os.Stat(leaf); err != nil {
			t.Errorf("spec にある process 配下の手動の葉が --sync で削除された: %v", err)
		}
		meta, err := repository.ReadNodeMetadata(filepath.Join(projDir, "scenario", "daily-balance"))
		if err != nil {
			t.Fatal(err)
		}
		if meta.Description != "updated description" {
			t.Errorf("description not updated: %q", meta.Description)
		}
	})
}

// 未知の VO 検証エラー (group に "_" を含む等) はディレクトリを一切作らずに失敗する。
func TestScaffoldFromSpecValidationError(t *testing.T) {
	t.Run("ScaffoldFromSpec_groupにアンダースコアを含む場合_ディレクトリを作らずエラーであること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer

		spec := sampleSpec()
		spec.Bizdates[0].Processes[0].Group = "pre_group"

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, spec, false)

		// Assert
		if err == nil {
			t.Fatal("expected validation error")
		}
		if _, statErr := os.Stat(filepath.Join(projDir, "scenario", "daily-balance")); !os.IsNotExist(statErr) {
			t.Errorf("scenario dir should not be created on validation error: err=%v", statErr)
		}
	})
}

// spec 内に同一 bizdate ディレクトリ名 (seq+bizdate が同じ) が 2 件あると、
// silent 上書きせずエラーにし、何も書き込まない (F1)。
func TestScaffoldFromSpecDuplicateBizdateDir(t *testing.T) {
	t.Run("ScaffoldFromSpec_bizdateディレクトリ名が重複する場合_何も書かずエラーであること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer

		spec := sampleSpec()
		spec.Bizdates = append(spec.Bizdates, repository.BizdateSpec{
			Seq:     "10",
			Bizdate: "20240101",
		})

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, spec, false)

		// Assert
		if err == nil {
			t.Fatal("expected duplicate bizdate directory error")
		}
		if !strings.Contains(err.Error(), "duplicate bizdate directory") {
			t.Errorf("error = %v, want message to contain %q", err, "duplicate bizdate directory")
		}
		if _, statErr := os.Stat(filepath.Join(projDir, "scenario", "daily-balance")); !os.IsNotExist(statErr) {
			t.Errorf("scenario dir should not be created on duplicate detection: err=%v", statErr)
		}
	})
}

// spec 内の 1 つの bizdate に同一 process ディレクトリ名 (seq+group+type が同じ) が
// 2 件あると、silent 上書きせずエラーにし、何も書き込まない (F1)。
func TestScaffoldFromSpecDuplicateProcessDir(t *testing.T) {
	t.Run("ScaffoldFromSpec_processディレクトリ名が重複する場合_何も書かずエラーであること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer

		spec := sampleSpec()
		spec.Bizdates[0].Processes = append(spec.Bizdates[0].Processes, repository.ProcessSpec{
			Seq:   "10",
			Group: "arrange",
			Type:  "clearPostgres",
		})

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, spec, false)

		// Assert
		if err == nil {
			t.Fatal("expected duplicate process directory error")
		}
		if !strings.Contains(err.Error(), "duplicate process directory") {
			t.Errorf("error = %v, want message to contain %q", err, "duplicate process directory")
		}
		if _, statErr := os.Stat(filepath.Join(projDir, "scenario", "daily-balance")); !os.IsNotExist(statErr) {
			t.Errorf("scenario dir should not be created on duplicate detection: err=%v", statErr)
		}
	})
}

// spec → scaffold → tree → spec' の往復が安定する (plan §7 の往復セマンティクス)。
func TestScaffoldFromSpecRoundtrip(t *testing.T) {
	t.Run("ScaffoldFromSpec_scaffold後にtreeからspecを再構築する場合_元のspecと一致すること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer
		spec := sampleSpec()

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, spec, false)

		// Assert
		if err != nil {
			t.Fatalf("ScaffoldFromSpec: %v", err)
		}

		tree, err := repository.LoadScenarioTree(projDir, []string{"daily-balance"})
		if err != nil {
			t.Fatal(err)
		}
		view, ok := tree.ScenarioView("daily-balance")
		if !ok {
			t.Fatal("ScenarioView not found")
		}
		got, err := repository.BuildSpecFromTree(projDir, view)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, spec) {
			t.Errorf("roundtrip spec mismatch:\ngot  = %#v\nwant = %#v", got, spec)
		}
	})
}

// --sync は spec との差分同期: spec に無い bizdate/process ディレクトリを
// (実装済みの葉ごと) 削除し、spec にある process 配下の葉は温存する。
func TestScaffoldFromSpecSyncPrunesTree(t *testing.T) {
	t.Run("ScaffoldFromSpec_syncでspecに無いディレクトリがある場合_余分を削除し葉と規約外を温存すること", func(t *testing.T) {
		// Arrange
		projDir := newTestProject(t)
		var out bytes.Buffer
		if err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), false); err != nil {
			t.Fatalf("1st ScaffoldFromSpec: %v", err)
		}

		scenarioDir := filepath.Join(projDir, "scenario", "daily-balance")

		// (a) spec に無い余分な bizdate を葉つきで追加
		extraBizdate := filepath.Join(scenarioDir, "_20_20240102", "_10_act_invokeRest", "script.js")
		if err := os.MkdirAll(filepath.Dir(extraBizdate), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(extraBizdate, []byte("// old\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// (b) 残る bizdate 配下に spec に無い余分な process を葉つきで追加
		extraProcess := filepath.Join(scenarioDir, "_10_20240101", "_99_extra_scripts", "scripts", "100_old")
		if err := os.MkdirAll(filepath.Dir(extraProcess), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(extraProcess, []byte("#!/bin/bash\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		// (c) spec にある process 配下の葉 (温存されるべき)
		keptLeaf := filepath.Join(scenarioDir, "_10_20240101", "_10_arrange_clearPostgres", "data", "appdb", "accounts.csv")
		if err := os.MkdirAll(filepath.Dir(keptLeaf), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keptLeaf, []byte("id,balance\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// (d) 規約外のディレクトリ (触れられるべきでない)
		unmanaged := filepath.Join(scenarioDir, "notes", "memo.md")
		if err := os.MkdirAll(filepath.Dir(unmanaged), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(unmanaged, []byte("keep me\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		out.Reset()

		// Act
		err := ScaffoldFromSpec(testLogger(), &out, projDir, sampleSpec(), true)

		// Assert
		if err != nil {
			t.Fatalf("sync ScaffoldFromSpec: %v", err)
		}

		// spec に無い bizdate/process は削除される
		if _, err := os.Stat(filepath.Join(scenarioDir, "_20_20240102")); !os.IsNotExist(err) {
			t.Error("spec に無い bizdate _20_20240102 は削除されるべき")
		}
		if _, err := os.Stat(filepath.Join(scenarioDir, "_10_20240101", "_99_extra_scripts")); !os.IsNotExist(err) {
			t.Error("spec に無い process _99_extra_scripts は削除されるべき")
		}
		// spec にある process の葉は温存
		if _, err := os.Stat(keptLeaf); err != nil {
			t.Errorf("spec にある process の葉 (accounts.csv) は温存されるべき: %v", err)
		}
		// spec にある process 本体も残る
		if _, err := os.Stat(filepath.Join(scenarioDir, "_10_20240101", "_30_act_invokeRest")); err != nil {
			t.Errorf("spec にある process _30_act_invokeRest は残るべき: %v", err)
		}
		// 規約外ディレクトリは触れない
		if _, err := os.Stat(unmanaged); err != nil {
			t.Errorf("規約外ディレクトリ notes/ は温存されるべき: %v", err)
		}
		// 出力に削除行が含まれる
		if !strings.Contains(out.String(), "removed: scenario/daily-balance/_20_20240102") {
			t.Errorf("prune の出力に削除行が無い:\n%s", out.String())
		}
	})
}
