package project

import (
	"reflect"
	"testing"
)

func testGroups() map[string][]string {
	return map[string][]string{
		"web":   {"127.0.0.1", "localhost"},
		"ap":    {"127.0.0.1"},
		"db":    {"localhost"},
		"empty": {},
	}
}

func TestSelectInventoryHosts(t *testing.T) {
	tests := []struct {
		name  string
		group string
		want  []string
	}{
		{"単一グループ", "ap", []string{"127.0.0.1"}},
		{"複数ホストは昇順", "web", []string{"127.0.0.1", "localhost"}},
		{"all は全グループ横断 + 重複排除 (v0.2 予約値)", "all", []string{"127.0.0.1", "localhost"}},
		{"未定義グループは空", "NOTEXIST", []string{}},
		{"空グループは空", "empty", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectInventoryHosts(testGroups(), tt.group)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectInventoryHosts(%q) = %v, want %v", tt.group, got, tt.want)
			}
		})
	}
}

func TestInventoryGroupExists(t *testing.T) {
	tests := []struct {
		name  string
		group string
		want  bool
	}{
		{"定義済みグループ", "ap", true},
		{"all はホストがあれば true", "all", true},
		{"未定義グループ", "NOTEXIST", false},
		{"空グループはホスト取得結果の有無で false (v0.2 互換)", "empty", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InventoryGroupExists(testGroups(), tt.group); got != tt.want {
				t.Errorf("InventoryGroupExists(%q) = %v, want %v", tt.group, got, tt.want)
			}
		})
	}
}
