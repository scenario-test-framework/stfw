package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadProcessConfigSubtree(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "stfw:\n  process:\n    clearPostgres:\n      host_group: db\n      tables:\n        - transactions\n        - accounts\n"
	if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	want := map[string]any{
		"host_group": "db",
		"tables":     []any{"transactions", "accounts"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadProcessConfigSubtree = %#v, want %#v", got, want)
	}

	// 別 type を指定すると見つからない (nil)。
	other, err := ReadProcessConfigSubtree(dir, "clearMysql")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree(other type): %v", err)
	}
	if other != nil {
		t.Errorf("ReadProcessConfigSubtree(other type) = %#v, want nil", other)
	}
}

// stfw.process.{type} が明示的な null (`clearPostgres:` のみ、値なし) の場合は
// 未設定と同一視して nil を返す (F3: null/欠落は従来どおり nil)。
func TestReadProcessConfigSubtreeExplicitNull(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "stfw:\n  process:\n    clearPostgres:\n"
	if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	if got != nil {
		t.Errorf("ReadProcessConfigSubtree(explicit null) = %#v, want nil", got)
	}
}

// stfw.process.{type} が mapping でない (list) 場合は silent drop せずエラーにする
// (F3: 往復忠実性のため fail-loud)。
func TestReadProcessConfigSubtreeNonMappingList(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "stfw:\n  process:\n    clearPostgres:\n      - a\n      - b\n"
	if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err == nil {
		t.Fatal("expected error for non-mapping stfw.process.{type} (list)")
	}
	if !strings.Contains(err.Error(), "must be a mapping") {
		t.Errorf("error = %v, want message to contain %q", err, "must be a mapping")
	}
}

// stfw.process.{type} が mapping でない (scalar/string) 場合も同様にエラーにする (F3)。
func TestReadProcessConfigSubtreeNonMappingScalar(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "stfw:\n  process:\n    clearPostgres: not-a-mapping\n"
	if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err == nil {
		t.Fatal("expected error for non-mapping stfw.process.{type} (scalar)")
	}
	if !strings.Contains(err.Error(), "must be a mapping") {
		t.Errorf("error = %v, want message to contain %q", err, "must be a mapping")
	}
}

func TestReadProcessConfigSubtreeMissing(t *testing.T) {
	dir := t.TempDir()
	got, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	if got != nil {
		t.Errorf("ReadProcessConfigSubtree(missing) = %#v, want nil", got)
	}
}

// 空スタブ (stfw.process.{type}: {}) は nil として読める (往復での同一性)。
func TestReadProcessConfigSubtreeEmptyStub(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "stfw:\n  process:\n    scripts: {}\n"
	if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadProcessConfigSubtree(dir, "scripts")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	if got != nil {
		t.Errorf("ReadProcessConfigSubtree(empty stub) = %#v, want nil", got)
	}
}

func TestWriteProcessConfigRoundtrip(t *testing.T) {
	dir := t.TempDir()
	cfg := map[string]any{
		"host_group": "db",
		"tables":     []any{"transactions", "accounts"},
	}

	if err := WriteProcessConfig(dir, "clearPostgres", cfg); err != nil {
		t.Fatalf("WriteProcessConfig: %v", err)
	}
	got, err := ReadProcessConfigSubtree(dir, "clearPostgres")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	if !reflect.DeepEqual(got, cfg) {
		t.Errorf("roundtrip = %#v, want %#v", got, cfg)
	}
}

// cfg が nil (未指定) の場合は空スタブを書き、読み戻すと nil になる (往復の決定性)。
func TestWriteProcessConfigEmpty(t *testing.T) {
	dir := t.TempDir()

	if err := WriteProcessConfig(dir, "scripts", nil); err != nil {
		t.Fatalf("WriteProcessConfig: %v", err)
	}
	got, err := ReadProcessConfigSubtree(dir, "scripts")
	if err != nil {
		t.Fatalf("ReadProcessConfigSubtree: %v", err)
	}
	if got != nil {
		t.Errorf("ReadProcessConfigSubtree(after empty write) = %#v, want nil", got)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "config", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) == "" {
		t.Error("config.yml should not be empty (expect stfw.process.scripts stub)")
	}
}

// yaml.v3 の map マーシャルはキー昇順で決定論的に出力する (spec / config の Marshal 全般が
// 依拠する前提を固定する)。
func TestWriteProcessConfigDeterministic(t *testing.T) {
	cfg := map[string]any{"zebra": 1, "apple": 2}
	dirA := t.TempDir()
	dirB := t.TempDir()
	if err := WriteProcessConfig(dirA, "sample", cfg); err != nil {
		t.Fatal(err)
	}
	if err := WriteProcessConfig(dirB, "sample", cfg); err != nil {
		t.Fatal(err)
	}
	rawA, err := os.ReadFile(filepath.Join(dirA, "config", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	rawB, err := os.ReadFile(filepath.Join(dirB, "config", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(rawA) != string(rawB) {
		t.Errorf("non-deterministic output:\nA=%q\nB=%q", rawA, rawB)
	}
}

// 祖先コンテナ (stfw / stfw.process 自体) が mapping でない場合も silent drop せず
// エラーにする (leaf だけでなく祖先の破損も fail-loud にする)。
func TestReadProcessConfigSubtreeNonMappingAncestor(t *testing.T) {
	cases := map[string]string{
		"stfw is list":           "stfw:\n  - a\n  - b\n",
		"stfw is scalar":         "stfw: not-a-mapping\n",
		"stfw.process is list":   "stfw:\n  process:\n    - a\n    - b\n",
		"stfw.process is scalar": "stfw:\n  process: not-a-mapping\n",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			confDir := filepath.Join(dir, "config")
			if err := os.MkdirAll(confDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(confDir, "config.yml"), []byte(raw), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := ReadProcessConfigSubtree(dir, "clearPostgres")
			if err == nil {
				t.Fatalf("expected error for %s", name)
			}
			if !strings.Contains(err.Error(), "must be a mapping") {
				t.Errorf("error = %v, want message to contain %q", err, "must be a mapping")
			}
		})
	}
}
