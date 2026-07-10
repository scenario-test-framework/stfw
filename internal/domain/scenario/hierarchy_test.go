package scenario

import "testing"

// v0.2 の is_*-dir 判定 (scenario ルートからの深さ) と同じ規則であることを固定する。
func TestHierarchy(t *testing.T) {
	tests := []struct {
		name     string
		rel      string
		root     bool
		scenario bool
		bizdate  bool
		process  bool
	}{
		{"IsXxxDir_scenarioルートの場合_rootのみtrueであること", "scenario", true, false, false, false},
		{"IsXxxDir_シナリオディレクトリの場合_scenarioのみtrueであること", "scenario/test", false, true, false, false},
		{"IsXxxDir_bizdateディレクトリの場合_bizdateのみtrueであること", "scenario/test/_10_99990101", false, false, true, false},
		{"IsXxxDir_processディレクトリの場合_processのみtrueであること", "scenario/test/_10_99990101/_10_pre_scripts", false, false, false, true},
		{"IsXxxDir_カレントの場合_すべてfalseであること", ".", false, false, false, false},
		{"IsXxxDir_空文字の場合_すべてfalseであること", "", false, false, false, false},
		{"IsXxxDir_configディレクトリの場合_すべてfalseであること", "config", false, false, false, false},
		{"IsXxxDir_config配下の場合_すべてfalseであること", "config/inventory", false, false, false, false},
		{"IsXxxDir_process配下の深い階層の場合_すべてfalseであること", "scenario/test/_10_99990101/_10_pre_scripts/scripts", false, false, false, false},
		{"IsXxxDir_scenarioルート外の場合_すべてfalseであること", "../outside/scenario/test", false, false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act & Assert
			if got := IsScenarioRootDir(tt.rel); got != tt.root {
				t.Errorf("IsScenarioRootDir(%q) = %v, want %v", tt.rel, got, tt.root)
			}
			if got := IsScenarioDir(tt.rel); got != tt.scenario {
				t.Errorf("IsScenarioDir(%q) = %v, want %v", tt.rel, got, tt.scenario)
			}
			if got := IsBizdateDir(tt.rel); got != tt.bizdate {
				t.Errorf("IsBizdateDir(%q) = %v, want %v", tt.rel, got, tt.bizdate)
			}
			if got := IsProcessDir(tt.rel); got != tt.process {
				t.Errorf("IsProcessDir(%q) = %v, want %v", tt.rel, got, tt.process)
			}
		})
	}
}
