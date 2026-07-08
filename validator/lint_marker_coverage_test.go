package validator

import (
	"reflect"
	"sort"
	"testing"
)

func TestEnumerateMarkers(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		ok   bool
	}{
		{"^(tests-ok|tests-failed)$", []string{"tests-failed", "tests-ok"}, true},
		{"^(a|b|c)$", []string{"a", "b", "c"}, true},
		{"(a|b)", []string{"a", "b"}, true},
		{"tests_pass", []string{"tests_pass"}, true},
		{"^pass$", []string{"pass"}, true},
		{"^(a|a|b)$", []string{"a", "b"}, true}, // dedup
		// non-enumerable → (nil,false):
		{"^(a|b|)$", nil, false},  // empty branch
		{"(a|b)|c", nil, false},   // group not full-span
		{"(a|b)?", nil, false},    // quantifier after group
		{"^(a.b|c)$", nil, false}, // metachar in branch
		{"^\\d+$", nil, false},    // metachars
		{".*fail.*", nil, false},  // metachars
		{"(?i)(a|b)", nil, false}, // flags: inner group has metachars, not full-span
	}
	for _, tc := range cases {
		got, ok := enumerateMarkers(tc.in)
		if ok != tc.ok {
			t.Errorf("%q: ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok {
			sort.Strings(got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%q: markers = %v, want %v", tc.in, got, tc.want)
			}
		}
	}
}
