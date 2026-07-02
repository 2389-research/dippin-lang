package ir

import "testing"

// eq builds an edge carrying a simple equality condition (pre-parsed).
func eq(from, to, variable, value string) *Edge {
	return &Edge{From: from, To: to, Condition: &Condition{
		Parsed: CondCompare{Variable: variable, Op: "=", Value: value},
	}}
}

func TestEdgesExhaustive(t *testing.T) {
	cases := []struct {
		name  string
		edges []*Edge
		want  bool
	}{
		{
			name:  "known success/fail set is exhaustive",
			edges: []*Edge{eq("A", "X", "ctx.outcome", "success"), eq("A", "Y", "ctx.outcome", "fail")},
			want:  true,
		},
		{
			name:  "single guard is non-exhaustive",
			edges: []*Edge{eq("A", "X", "ctx.flag", "yes")},
			want:  false,
		},
		{
			name:  "complete partition on one variable (2+ equality values)",
			edges: []*Edge{eq("A", "X", "ctx.tier", "gold"), eq("A", "Y", "ctx.tier", "silver")},
			want:  true,
		},
		{
			name: "complementary contains / not-contains pair is exhaustive",
			edges: []*Edge{
				{From: "A", To: "X", Condition: &Condition{Parsed: CondCompare{Variable: "ctx.body", Op: "contains", Value: "err"}}},
				{From: "A", To: "Y", Condition: &Condition{Parsed: CondNot{Inner: CondCompare{Variable: "ctx.body", Op: "contains", Value: "err"}}}},
			},
			want: true,
		},
		{
			name:  "two values across different variables is not a partition",
			edges: []*Edge{eq("A", "X", "ctx.a", "1"), eq("A", "Y", "ctx.b", "2")},
			want:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EdgesExhaustive(tc.edges); got != tc.want {
				t.Errorf("EdgesExhaustive = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractEqualityCondition(t *testing.T) {
	cmp, ok := ExtractEqualityCondition(eq("A", "B", "ctx.outcome", "success"))
	if !ok || cmp.Variable != "ctx.outcome" || cmp.Value != "success" {
		t.Fatalf("ExtractEqualityCondition = %+v, %v", cmp, ok)
	}
	if _, ok := ExtractEqualityCondition(&Edge{From: "A", To: "B"}); ok {
		t.Error("unconditional edge should not yield an equality condition")
	}
}
