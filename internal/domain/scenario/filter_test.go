package scenario

import (
	"reflect"
	"testing"
)

// filterTestView は 2 bizdate x 2 process の実行計画ビューを組み立てる。
func filterTestView() ScenarioView {
	return ScenarioView{
		Name: "demo",
		Bizdates: []BizdateView{
			{
				DirName: "_10_20240101", Seq: "10", Bizdate: "20240101",
				Processes: []ProcessView{
					{DirName: "_10_pre_scripts", Seq: "10", Group: "pre", ProcessType: "scripts"},
					{DirName: "_20_main_scripts", Seq: "20", Group: "main", ProcessType: "scripts"},
				},
			},
			{
				DirName: "_20_20240102", Seq: "20", Bizdate: "20240102",
				Processes: []ProcessView{
					{DirName: "_10_pre_scripts", Seq: "10", Group: "pre", ProcessType: "scripts"},
					{DirName: "_20_main_scripts", Seq: "20", Group: "main", ProcessType: "scripts"},
				},
			},
		},
	}
}

func TestNewRunFilter(t *testing.T) {
	t.Run("NewRunFilter_両方空の場合_フィルタなしであること", func(t *testing.T) {
		// Arrange / Act
		f, err := NewRunFilter("", "")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Active() {
			t.Errorf("Active() = true, want false")
		}
		if key, value := f.Attr(); key != "" || value != "" {
			t.Errorf("Attr() = (%q, %q), want empty", key, value)
		}
	})

	t.Run("NewRunFilter_両方指定の場合_エラーであること", func(t *testing.T) {
		// Arrange / Act
		_, err := NewRunFilter("_10_20240101", "_20_20240102")

		// Assert
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("NewRunFilter_3セグメントの場合_エラーであること", func(t *testing.T) {
		// Arrange / Act
		_, err := NewRunFilter("_10_20240101/_10_pre_scripts/extra", "")

		// Assert
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("NewRunFilter_空セグメントの場合_エラーであること", func(t *testing.T) {
		// Arrange / Act
		_, err := NewRunFilter("", "_10_20240101/")

		// Assert
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("NewRunFilter_fromを指定する場合_attrがfromキーで記録されること", func(t *testing.T) {
		// Arrange / Act
		f, err := NewRunFilter("_10_20240101/_20_main_scripts", "")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		key, value := f.Attr()
		if key != "from" || value != "_10_20240101/_20_main_scripts" {
			t.Errorf("Attr() = (%q, %q), want (from, _10_20240101/_20_main_scripts)", key, value)
		}
	})

	t.Run("NewRunFilter_onlyを指定する場合_attrがonlyキーで記録されること", func(t *testing.T) {
		// Arrange / Act
		f, err := NewRunFilter("", "_20_20240102")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		key, value := f.Attr()
		if key != "only" || value != "_20_20240102" {
			t.Errorf("Attr() = (%q, %q), want (only, _20_20240102)", key, value)
		}
	})
}

func TestRunFilterApply(t *testing.T) {
	t.Run("Apply_フィルタなしの場合_viewが変わらないこと", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("", "")

		// Act
		got, err := f.Apply(view)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, view) {
			t.Errorf("Apply() = %+v, want %+v", got, view)
		}
	})

	t.Run("Apply_fromでbizdateを指定する場合_先行bizdateがスキップされること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("_20_20240102", "")

		// Act
		got, err := f.Apply(view)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := ScenarioView{Name: "demo", Bizdates: []BizdateView{view.Bizdates[1]}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Apply() = %+v, want %+v", got, want)
		}
	})

	t.Run("Apply_fromでprocessまで指定する場合_同一bizdate内の先行processのみスキップされること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("_10_20240101/_20_main_scripts", "")

		// Act
		got, err := f.Apply(view)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := ScenarioView{Name: "demo", Bizdates: []BizdateView{
			{
				DirName: "_10_20240101", Seq: "10", Bizdate: "20240101",
				Processes: []ProcessView{view.Bizdates[0].Processes[1]},
			},
			view.Bizdates[1], // 後続 bizdate は全 process を実行する
		}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Apply() = %+v, want %+v", got, want)
		}
	})

	t.Run("Apply_fromを適用する場合_元のviewを破壊しないこと", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("_10_20240101/_20_main_scripts", "")

		// Act
		if _, err := f.Apply(view); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Assert
		if !reflect.DeepEqual(view, filterTestView()) {
			t.Errorf("view mutated: %+v", view)
		}
	})

	t.Run("Apply_onlyでbizdateを指定する場合_該当bizdateのみ実行されること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("", "_10_20240101")

		// Act
		got, err := f.Apply(view)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := ScenarioView{Name: "demo", Bizdates: []BizdateView{view.Bizdates[0]}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Apply() = %+v, want %+v", got, want)
		}
	})

	t.Run("Apply_onlyでprocessまで指定する場合_該当processのみ実行されること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("", "_20_20240102/_20_main_scripts")

		// Act
		got, err := f.Apply(view)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := ScenarioView{Name: "demo", Bizdates: []BizdateView{
			{
				DirName: "_20_20240102", Seq: "20", Bizdate: "20240102",
				Processes: []ProcessView{view.Bizdates[1].Processes[1]},
			},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Apply() = %+v, want %+v", got, want)
		}
	})

	t.Run("Apply_存在しないbizdateの場合_エラーであること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("_99_99990101", "")

		// Act
		_, err := f.Apply(view)

		// Assert
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Apply_存在しないprocessの場合_エラーであること", func(t *testing.T) {
		// Arrange
		view := filterTestView()
		f, _ := NewRunFilter("", "_10_20240101/_99_missing_scripts")

		// Act
		_, err := f.Apply(view)

		// Assert
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
