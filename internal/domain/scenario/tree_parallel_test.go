package scenario

import (
	"strings"
	"testing"
)

// childRaw は parallel 配下の子プロセス走査結果を組み立てる。
func childRaw(dirName string) RawDir {
	return RawDir{
		Name: dirName,
		Dirs: []RawDir{
			{Name: "config", Files: []string{"config.yml"}},
		},
		Files: []string{"metadata.yml"},
	}
}

// parallelScenarioRaw は parallel プロセスを含む正常なシナリオ走査結果。
func parallelScenarioRaw(name string) RawDir {
	parallel := RawDir{
		Name: "_10_export_parallel",
		Dirs: []RawDir{
			{Name: "config", Files: []string{"config.yml"}},
			// 意図的に降順で渡し、走査規則 (昇順) の適用を確認する
			childRaw("_20_db_exportB"),
			childRaw("_10_db_exportA"),
			// `_` 始まりでないディレクトリは走査対象外
			{Name: "note", Files: []string{"memo.txt"}},
		},
		Files: []string{"metadata.yml"},
	}
	return RawDir{
		Name: name,
		Dirs: []RawDir{
			{Name: "_10_99990101", Dirs: []RawDir{parallel}, Files: []string{"metadata.yml"}},
		},
		Files: []string{"metadata.yml"},
	}
}

func TestScenarioTreeParallelChildren(t *testing.T) {
	t.Run("NewScenarioTree_parallel配下に子がある場合_子ビューが昇順で構築されること", func(t *testing.T) {
		// Arrange
		tree := NewScenarioTree([]RawDir{parallelScenarioRaw("test")})

		// Act
		view, ok := tree.ScenarioView("test")

		// Assert
		if !ok {
			t.Fatal("ScenarioView(test) not found")
		}
		p := view.Bizdates[0].Processes[0]
		if p.ProcessType != ParallelProcessType {
			t.Fatalf("ProcessType = %s, want %s", p.ProcessType, ParallelProcessType)
		}
		if len(p.Children) != 2 {
			t.Fatalf("len(Children) = %d, want 2", len(p.Children))
		}
		if p.Children[0].DirName != "_10_db_exportA" || p.Children[1].DirName != "_20_db_exportB" {
			t.Errorf("Children order = [%s %s], want ascending", p.Children[0].DirName, p.Children[1].DirName)
		}
		if p.Children[0].Seq != "10" || p.Children[0].Group != "db" || p.Children[0].ProcessType != "exportA" {
			t.Errorf("Children[0] = %+v, want seq=10 group=db type=exportA", p.Children[0])
		}
	})

	t.Run("NewScenarioTree_parallel以外のプロセスの場合_配下ディレクトリを子として構築しないこと", func(t *testing.T) {
		// Arrange
		tree := NewScenarioTree([]RawDir{validScenarioRaw("test")})

		// Act
		view, _ := tree.ScenarioView("test")

		// Assert
		for _, b := range view.Bizdates {
			for _, p := range b.Processes {
				if len(p.Children) != 0 {
					t.Errorf("process %s has children %v, want none", p.DirName, p.Children)
				}
			}
		}
	})

	t.Run("ProcessTypes_parallelの子がある場合_子タイプを含むこと", func(t *testing.T) {
		// Arrange
		tree := NewScenarioTree([]RawDir{parallelScenarioRaw("test")})

		// Act
		types := tree.ProcessTypes()

		// Assert
		want := []string{"exportA", "exportB", "parallel"}
		if len(types) != len(want) {
			t.Fatalf("ProcessTypes() = %v, want %v", types, want)
		}
		for i := range want {
			if types[i] != want[i] {
				t.Errorf("ProcessTypes()[%d] = %s, want %s", i, types[i], want[i])
			}
		}
	})
}

func TestScenarioTreeParallelValidate(t *testing.T) {
	installed := []string{"parallel", "exportA", "exportB", "scripts"}

	t.Run("Validate_正常なparallel構成の場合_違反なしであること", func(t *testing.T) {
		// Arrange
		tree := NewScenarioTree([]RawDir{parallelScenarioRaw("test")})

		// Act
		vs := tree.Validate(installed)

		// Assert
		if len(vs) != 0 {
			t.Errorf("Validate() = %v, want no violations", vs)
		}
	})

	t.Run("Validate_parallelに子が無い場合_エラーであること", func(t *testing.T) {
		// Arrange
		raw := RawDir{
			Name: "test",
			Dirs: []RawDir{
				{Name: "_10_99990101", Dirs: []RawDir{{
					Name: "_10_export_parallel",
					Dirs: []RawDir{{Name: "config", Files: []string{"config.yml"}}},
				}}},
			},
		}
		tree := NewScenarioTree([]RawDir{raw})

		// Act
		vs := tree.Validate(installed)

		// Assert
		if len(vs) != 1 || !strings.Contains(vs[0].Message, "must have at least one child process") {
			t.Errorf("Validate() = %v, want 1 violation (no child process)", vs)
		}
	})

	t.Run("Validate_parallelの入れ子の場合_エラーであること", func(t *testing.T) {
		// Arrange
		nested := RawDir{
			Name: "_10_nest_parallel",
			Dirs: []RawDir{
				{Name: "config", Files: []string{"config.yml"}},
				childRaw("_10_db_exportA"),
			},
		}
		raw := RawDir{
			Name: "test",
			Dirs: []RawDir{
				{Name: "_10_99990101", Dirs: []RawDir{{
					Name: "_10_export_parallel",
					Dirs: []RawDir{
						{Name: "config", Files: []string{"config.yml"}},
						nested,
					},
				}}},
			},
		}
		tree := NewScenarioTree([]RawDir{raw})

		// Act
		vs := tree.Validate(installed)

		// Assert
		if len(vs) != 1 || !strings.Contains(vs[0].Message, "can not be nested") {
			t.Errorf("Validate() = %v, want 1 violation (nested parallel)", vs)
		}
		if len(vs) == 1 && !strings.Contains(vs[0].Path, "_10_export_parallel/_10_nest_parallel") {
			t.Errorf("Violation path = %s, want child path", vs[0].Path)
		}
	})

	t.Run("Validate_子の命名違反と設定欠落と未インストールの場合_エラー3件であること", func(t *testing.T) {
		// Arrange
		raw := RawDir{
			Name: "test",
			Dirs: []RawDir{
				{Name: "_10_99990101", Dirs: []RawDir{{
					Name: "_10_export_parallel",
					Dirs: []RawDir{
						{Name: "config", Files: []string{"config.yml"}},
						// 命名違反 (要素不足)
						{Name: "_10_broken"},
						// config/config.yml 欠落
						{Name: "_20_db_exportA", Files: []string{"metadata.yml"}},
						// プラグイン未解決
						childRaw("_30_db_unknown"),
					},
				}}},
			},
		}
		tree := NewScenarioTree([]RawDir{raw})

		// Act
		vs := tree.Validate(installed)

		// Assert
		if len(vs) != 3 {
			t.Fatalf("Validate() = %v, want 3 violations", vs)
		}
		if !strings.Contains(vs[0].Message, "format") {
			t.Errorf("vs[0] = %v, want dirname format violation", vs[0])
		}
		if !strings.Contains(vs[1].Message, "config/config.yml is not exist") {
			t.Errorf("vs[1] = %v, want config missing violation", vs[1])
		}
		if !strings.Contains(vs[2].Message, "unknown is not installed") {
			t.Errorf("vs[2] = %v, want plugin not installed violation", vs[2])
		}
	})
}
