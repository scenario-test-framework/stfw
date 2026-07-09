package run

import "fmt"

// Steps はステップ実行結果のファーストクラスコレクション。
// 「昇順逐次実行・エラー時 Blocked」の状態モデル (Pending → Success/Error/Blocked)
// の遷移検証を内包する (v0.2 の scripts プラグイン bulk_exec_scripts の記録規則)。
type Steps struct {
	order  []string
	status map[string]StepStatus
}

// NewSteps はステップ名リスト (実行順) から全件 Pending で初期化する。
func NewSteps(names []string) (*Steps, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("steps must not be empty")
	}
	s := &Steps{status: map[string]StepStatus{}}
	for _, name := range names {
		if name == "" {
			return nil, fmt.Errorf("step name must not null")
		}
		if _, ok := s.status[name]; ok {
			return nil, fmt.Errorf("step %s is duplicated", name)
		}
		s.order = append(s.order, name)
		s.status[name] = StepPending
	}
	return s, nil
}

// MarkEnd はステップの終了状態 (Success | Error | Blocked) への遷移を検証して記録する。
func (s *Steps) MarkEnd(name string, to StepStatus) error {
	cur, ok := s.status[name]
	if !ok {
		return fmt.Errorf("step %s is not enumerated", name)
	}
	if err := cur.Transition(to); err != nil {
		return fmt.Errorf("step %s: %w", name, err)
	}
	s.status[name] = to
	return nil
}

// StepView はステップの表示用スナップショット。
type StepView struct {
	Name   string
	Status StepStatus
}

// Views は実行順のステップビューを返す。
func (s *Steps) Views() []StepView {
	views := make([]StepView, 0, len(s.order))
	for _, name := range s.order {
		views = append(views, StepView{Name: name, Status: s.status[name]})
	}
	return views
}
