package scenario

import "testing"

func mustSeq(t *testing.T, s string) Seq {
	t.Helper()
	seq, err := NewSeq(s)
	if err != nil {
		t.Fatal(err)
	}
	return seq
}

func mustBizdate(t *testing.T, s string) Bizdate {
	t.Helper()
	b, err := NewBizdate(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustGroup(t *testing.T, s string) Group {
	t.Helper()
	g, err := NewGroup(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// v0.2 の bizdate_spec.dirname (`_${seq}_${bizdate}`) と同じ形式であることを固定する。
func TestBizdateDirName(t *testing.T) {
	got := BizdateDirName(mustSeq(t, "10"), mustBizdate(t, "99990101"))
	if got != "_10_99990101" {
		t.Errorf("BizdateDirName = %q, want %q", got, "_10_99990101")
	}
}

func TestParseBizdateDirName(t *testing.T) {
	seq, bizdate, err := ParseBizdateDirName("_10_99990101")
	if err != nil {
		t.Fatal(err)
	}
	if seq.String() != "10" || bizdate.String() != "99990101" {
		t.Errorf("parsed = (%q, %q), want (10, 99990101)", seq.String(), bizdate.String())
	}

	invalid := map[string]string{
		"10_99990101":    "`_` 始まりでない",
		"_99990101":      "フィールド不足",
		"_10_99990101_x": "フィールド過多",
		"_1a_99990101":   "seq が数字でない",
		"_10_9999010":    "bizdate が 8 桁でない",
		"_10_99990230":   "bizdate が実在しない日付",
		"_10_pre":        "bizdate が日付形式でない (process 命名の紛れ込み)",
	}
	for name, reason := range invalid {
		if _, _, err := ParseBizdateDirName(name); err == nil {
			t.Errorf("ParseBizdateDirName(%q) should fail (%s)", name, reason)
		}
	}
}

// v0.2 の process_spec.dirname (`_${seq}_${group}_${type}`) と同じ形式であることを固定する。
func TestProcessDirName(t *testing.T) {
	got := ProcessDirName(mustSeq(t, "10"), mustGroup(t, "pre"), "scripts")
	if got != "_10_pre_scripts" {
		t.Errorf("ProcessDirName = %q, want %q", got, "_10_pre_scripts")
	}
}

func TestParseProcessDirName(t *testing.T) {
	seq, group, processType, err := ParseProcessDirName("_10_pre_scripts")
	if err != nil {
		t.Fatal(err)
	}
	if seq.String() != "10" || group.String() != "pre" || processType != "scripts" {
		t.Errorf("parsed = (%q, %q, %q), want (10, pre, scripts)",
			seq.String(), group.String(), processType)
	}

	invalid := map[string]string{
		"10_pre_scripts":     "`_` 始まりでない",
		"_10_pre":            "フィールド不足 (bizdate 命名)",
		"_10_pre_sub_script": "フィールド過多 (group か type に `_`)",
		"_1a_pre_scripts":    "seq が数字でない",
		"_10__scripts":       "group が空",
		"_10_pre_":           "type が空",
	}
	for name, reason := range invalid {
		if _, _, _, err := ParseProcessDirName(name); err == nil {
			t.Errorf("ParseProcessDirName(%q) should fail (%s)", name, reason)
		}
	}
}
