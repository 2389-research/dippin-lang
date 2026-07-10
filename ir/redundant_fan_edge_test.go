package ir

import "testing"

func TestIsRedundantFanEdge(t *testing.T) {
	w := &Workflow{
		Nodes: []*Node{
			{ID: "Fan", Config: ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "A"}, {ID: "B"},
			{ID: "Join", Config: FanInConfig{Sources: []string{"A", "B"}}},
		},
	}
	cases := []struct {
		name string
		e    *Edge
		want bool
	}{
		{"parallel fork match", &Edge{From: "Fan", To: "A"}, true},
		{"fan_in source match", &Edge{From: "A", To: "Join"}, true},
		{"not a fan edge", &Edge{From: "A", To: "B"}, false},
		{"conditional not redundant", &Edge{From: "Fan", To: "A", Condition: &Condition{Raw: "ctx.x = 1"}}, false},
		{"labeled not redundant", &Edge{From: "Fan", To: "A", Label: "left"}, false},
		{"choice not redundant", &Edge{From: "Fan", To: "A", Choice: "left"}, false},
		{"weighted not redundant", &Edge{From: "Fan", To: "A", Weight: 2}, false},
		{"override not redundant", &Edge{From: "Fan", To: "A", Override: true}, false},
		{"restart not redundant", &Edge{From: "Fan", To: "A", Restart: true}, false},
		{"comment not redundant", &Edge{From: "Fan", To: "A", Comment: "note"}, false},
		{"parallel target not listed", &Edge{From: "Fan", To: "C"}, false},
		{"unknown nodes", &Edge{From: "X", To: "Y"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRedundantFanEdge(w, tc.e); got != tc.want {
				t.Errorf("IsRedundantFanEdge = %v, want %v", got, tc.want)
			}
		})
	}
}
