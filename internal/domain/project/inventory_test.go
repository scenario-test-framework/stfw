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
		{"SelectInventoryHosts_単一グループの場合_そのグループのホストを返すこと", "ap", []string{"127.0.0.1"}},
		{"SelectInventoryHosts_複数ホストの場合_昇順で返すこと", "web", []string{"127.0.0.1", "localhost"}},
		{"SelectInventoryHosts_allの場合_全グループ横断で重複排除して返すこと", "all", []string{"127.0.0.1", "localhost"}},
		{"SelectInventoryHosts_未定義グループの場合_空であること", "NOTEXIST", []string{}},
		{"SelectInventoryHosts_空グループの場合_空であること", "empty", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			groups := testGroups()
			// Act
			got := SelectInventoryHosts(groups, tt.group)
			// Assert
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
		{"InventoryGroupExists_定義済みグループの場合_trueであること", "ap", true},
		{"InventoryGroupExists_allでホストがある場合_trueであること", "all", true},
		{"InventoryGroupExists_未定義グループの場合_falseであること", "NOTEXIST", false},
		{"InventoryGroupExists_空グループの場合_falseであること", "empty", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			groups := testGroups()
			// Act
			got := InventoryGroupExists(groups, tt.group)
			// Assert
			if got != tt.want {
				t.Errorf("InventoryGroupExists(%q) = %v, want %v", tt.group, got, tt.want)
			}
		})
	}
}
