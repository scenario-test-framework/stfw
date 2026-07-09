package scenario

import (
	"strings"
	"testing"
)

// 正常な構成のシナリオ走査結果。
func validScenarioRaw(name string) RawDir {
	process := func(dirName string) RawDir {
		return RawDir{
			Name: dirName,
			Dirs: []RawDir{
				{Name: "config", Files: []string{"config.yml"}},
				{Name: "scripts", Files: []string{"100_1st_step", "200_2nd_step"}},
			},
			Files: []string{"metadata.yml"},
		}
	}
	return RawDir{
		Name: name,
		Dirs: []RawDir{
			// 意図的に降順で渡し、走査規則 (昇順) の適用を確認する
			{Name: "_20_99990102", Dirs: []RawDir{process("_10_pre_scripts")}, Files: []string{"metadata.yml"}},
			{Name: "_10_99990101", Dirs: []RawDir{process("_10_pre_scripts")}, Files: []string{"metadata.yml"}},
			// `_` 始まりでないディレクトリは走査対象外
			{Name: "note", Files: []string{"memo.txt"}},
		},
		Files: []string{"metadata.yml"},
	}
}

func TestScenarioTreeTraversalRule(t *testing.T) {
	tree := NewScenarioTree([]RawDir{validScenarioRaw("test2"), validScenarioRaw("test1")})

	// シナリオは名前昇順
	names := tree.Scenarios()
	if len(names) != 2 || names[0] != "test1" || names[1] != "test2" {
		t.Errorf("Scenarios() = %v, want [test1 test2]", names)
	}

	// プロセスタイプは重複なし
	types := tree.ProcessTypes()
	if len(types) != 1 || types[0] != "scripts" {
		t.Errorf("ProcessTypes() = %v, want [scripts]", types)
	}
}

func TestScenarioTreeValidatePass(t *testing.T) {
	tree := NewScenarioTree([]RawDir{validScenarioRaw("test")})
	vs := tree.Validate([]string{"scripts"})
	if len(vs) != 0 {
		t.Errorf("Validate() = %v, want no violations", vs)
	}
}

func TestScenarioTreeValidateViolations(t *testing.T) {
	raw := RawDir{
		Name: "broken",
		Dirs: []RawDir{
			// bizdate 命名規約違反 (seq が数字でない)
			{Name: "_1x_99990101"},
			// bizdate 命名規約違反 (実在しない日付)
			{Name: "_10_99990230"},
			{
				Name:  "_20_99990102",
				Files: []string{"bizdate.dig"}, // 残存 dig → 警告
				Dirs: []RawDir{
					// process 命名規約違反 (フィールド不足)
					{Name: "_10_scripts"},
					// config/config.yml 無し + 未インストールタイプ
					{Name: "_20_post_unknown", Files: []string{"metadata.yml"}},
				},
			},
		},
		Files: []string{"scenario.dig"}, // 残存 dig → 警告
	}
	tree := NewScenarioTree([]RawDir{raw})
	vs := tree.Validate([]string{"scripts"})

	if !vs.HasError() {
		t.Fatal("Validate() should have errors")
	}
	errors, warns := vs.Count()
	// エラー: bizdate 命名 x2 + process 命名 x1 + 未インストール x1 + config 無し x1
	if errors != 5 {
		t.Errorf("errors = %d, want 5: %v", errors, vs)
	}
	// 警告: scenario.dig + bizdate.dig
	if warns != 2 {
		t.Errorf("warns = %d, want 2: %v", warns, vs)
	}

	assertViolation(t, vs, ViolationError, "scenario/broken/_1x_99990101", "must be number")
	assertViolation(t, vs, ViolationError, "scenario/broken/_10_99990230", "not a valid date")
	assertViolation(t, vs, ViolationError, "scenario/broken/_20_99990102/_10_scripts", "format")
	assertViolation(t, vs, ViolationError, "scenario/broken/_20_99990102/_20_post_unknown", "is not installed")
	assertViolation(t, vs, ViolationError, "scenario/broken/_20_99990102/_20_post_unknown", "config/config.yml is not exist")
	assertViolation(t, vs, ViolationWarn, "scenario/broken/scenario.dig", "v1.0")
	assertViolation(t, vs, ViolationWarn, "scenario/broken/_20_99990102/bizdate.dig", "v1.0")
}

func assertViolation(t *testing.T, vs Violations, level ViolationLevel, path, msgPart string) {
	t.Helper()
	for _, v := range vs {
		if v.Level == level && v.Path == path && strings.Contains(v.Message, msgPart) {
			return
		}
	}
	t.Errorf("violation not found: level=%s path=%s msg~%q in %v", level, path, msgPart, vs)
}
