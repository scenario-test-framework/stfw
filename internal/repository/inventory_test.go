package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeInventory(t *testing.T, projDir, fileName, content string) {
	t.Helper()
	dir := filepath.Join(projDir, "config", "inventory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadInventory(t *testing.T) {
	projDir := t.TempDir()
	writeInventory(t, projDir, "staging.yml", `stfw_inventory:
  - web:
    - 127.0.0.1
    - localhost
  - ap:
    - 127.0.0.1
  - db:
    - localhost
`)

	groups, err := LoadInventory(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventory: %v", err)
	}
	want := map[string][]string{
		"web": {"127.0.0.1", "localhost"},
		"ap":  {"127.0.0.1"},
		"db":  {"localhost"},
	}
	if !reflect.DeepEqual(groups, want) {
		t.Errorf("LoadInventory = %v, want %v", groups, want)
	}
}

func TestLoadInventoryMergesDuplicateGroups(t *testing.T) {
	projDir := t.TempDir()
	writeInventory(t, projDir, "staging.yml", `stfw_inventory:
  - web:
    - host1
  - web:
    - host2
`)

	groups, err := LoadInventory(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventory: %v", err)
	}
	if !reflect.DeepEqual(groups["web"], []string{"host1", "host2"}) {
		t.Errorf("groups[web] = %v, want [host1 host2]", groups["web"])
	}
}

func TestLoadInventoryFileNotFound(t *testing.T) {
	projDir := t.TempDir()
	if _, err := LoadInventory(projDir, "missing.yml"); err == nil {
		t.Error("LoadInventory(missing) = nil, want error")
	}
}

// TestLoadInventoryStructuredEntries は文字列形式と構造化形式 (host+arch) の
// 混在を後方互換で受理し、LoadInventory はホスト名、LoadInventoryHostArch は
// arch を返すことを固定する。
func TestLoadInventoryStructuredEntries(t *testing.T) {
	projDir := t.TempDir()
	writeInventory(t, projDir, "staging.yml", `stfw_inventory:
  - web:
    - 127.0.0.1
  - db:
    - host: db1.example
      arch: linux_amd64
    - host: db2.example
      arch: linux_arm64
    - plain.example
`)

	groups, err := LoadInventory(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventory: %v", err)
	}
	if !reflect.DeepEqual(groups["web"], []string{"127.0.0.1"}) {
		t.Errorf("groups[web] = %v, want [127.0.0.1]", groups["web"])
	}
	if !reflect.DeepEqual(groups["db"], []string{"db1.example", "db2.example", "plain.example"}) {
		t.Errorf("groups[db] = %v", groups["db"])
	}

	hostArch, err := LoadInventoryHostArch(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventoryHostArch: %v", err)
	}
	want := map[string]string{"db1.example": "linux_amd64", "db2.example": "linux_arm64"}
	if !reflect.DeepEqual(hostArch, want) {
		t.Errorf("LoadInventoryHostArch = %v, want %v", hostArch, want)
	}
}
