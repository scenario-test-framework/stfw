package scenario

import "testing"

// v0.2 の is_*-dir 判定 (scenario ルートからの深さ) と同じ規則であることを固定する。
func TestHierarchy(t *testing.T) {
	tests := []struct {
		rel      string
		root     bool
		scenario bool
		bizdate  bool
		process  bool
	}{
		{"scenario", true, false, false, false},
		{"scenario/test", false, true, false, false},
		{"scenario/test/_10_99990101", false, false, true, false},
		{"scenario/test/_10_99990101/_10_pre_scripts", false, false, false, true},
		{".", false, false, false, false},
		{"", false, false, false, false},
		{"config", false, false, false, false},
		{"config/inventory", false, false, false, false},
		{"scenario/test/_10_99990101/_10_pre_scripts/scripts", false, false, false, false},
		{"../outside/scenario/test", false, false, false, false},
	}
	for _, tt := range tests {
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
	}
}
