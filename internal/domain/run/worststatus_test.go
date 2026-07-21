package run

import "testing"

func TestWorstStatus(t *testing.T) {
	cases := []struct {
		name string
		a, b NodeStatus
		want NodeStatus
	}{
		{"WorstStatus_両方Successの場合_Successであること", NodeSuccess, NodeSuccess, NodeSuccess},
		{"WorstStatus_SuccessとWarnの場合_Warnであること", NodeSuccess, NodeWarn, NodeWarn},
		{"WorstStatus_WarnとErrorの場合_Errorであること", NodeWarn, NodeError, NodeError},
		{"WorstStatus_ErrorとSuccessの場合_Errorであること", NodeError, NodeSuccess, NodeError},
		{"WorstStatus_両方Warnの場合_Warnであること", NodeWarn, NodeWarn, NodeWarn},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Arrange (ケース定義)
			// Act
			got := WorstStatus(c.a, c.b)
			// Assert
			if got != c.want {
				t.Errorf("WorstStatus(%s, %s) = %s, want %s", c.a, c.b, got, c.want)
			}
		})
	}
}
