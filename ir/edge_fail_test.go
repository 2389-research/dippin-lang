package ir

import "testing"

func failEdge(variable, value string) *Edge {
	return &Edge{From: "A", To: "B", Condition: &Condition{
		Parsed: CondCompare{Variable: variable, Op: "=", Value: value},
	}}
}

func TestEdgeRoutesOnFail(t *testing.T) {
	cases := []struct {
		name string
		edge *Edge
		want bool
	}{
		{"ctx.outcome = fail", failEdge("ctx.outcome", "fail"), true},
		{"ctx.outcome = failure", failEdge("ctx.outcome", "failure"), true},
		{"bare outcome = fail", failEdge("outcome", "fail"), true},
		{"ctx.outcome = success", failEdge("ctx.outcome", "success"), false},
		{"other variable = fail", failEdge("ctx.tool_marker", "fail"), false},
		{"unconditional edge", &Edge{From: "A", To: "B"}, false},
		{"unparsed condition", &Edge{From: "A", To: "B", Condition: &Condition{Raw: "ctx.outcome = fail"}}, false},
	}
	for _, tc := range cases {
		if got := EdgeRoutesOnFail(tc.edge); got != tc.want {
			t.Errorf("%s: EdgeRoutesOnFail = %v, want %v", tc.name, got, tc.want)
		}
	}
}
