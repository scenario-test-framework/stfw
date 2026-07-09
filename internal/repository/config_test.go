package repository

import "testing"

// v0.2 export_yaml のサンプル (bash_utils のコメント) と同じ規則であることを固定する。
func TestFlattenYAML(t *testing.T) {
	t.Setenv("EXPAND_VAR_FOR_TEST", "http://env.example/endpoint")

	raw := []byte(`
map:
  key: value1
  list:
  - list_value1
  - list_value2
expand: ${EXPAND_VAR_FOR_TEST}
missing: ${UNDEFINED_VAR_FOR_TEST}
`)
	got := map[string]string{}
	if err := flattenYAML(raw, got); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"map_key":    "value1",
		"map_list_0": "list_value1",
		"map_list_1": "list_value2",
		"expand":     "http://env.example/endpoint",
		"missing":    "", // bash の source と同じく未定義は空文字
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("flat[%q] = %q, want %q", k, got[k], w)
		}
	}
}

func TestFlattenYAMLOverride(t *testing.T) {
	dst := map[string]string{}
	if err := flattenYAML([]byte("stfw:\n  loglevel: info\n  timezone: Asia/Tokyo\n"), dst); err != nil {
		t.Fatal(err)
	}
	if err := flattenYAML([]byte("stfw:\n  loglevel: debug\n"), dst); err != nil {
		t.Fatal(err)
	}
	if dst["stfw_loglevel"] != "debug" {
		t.Errorf("project 設定でデフォルトが上書きされていない: %q", dst["stfw_loglevel"])
	}
	if dst["stfw_timezone"] != "Asia/Tokyo" {
		t.Errorf("上書きされていないキーはデフォルトを維持すべき: %q", dst["stfw_timezone"])
	}
}
