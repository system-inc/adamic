package main

import "testing"

func TestClassification(t *testing.T) {
	t.Parallel()
	accepted := observation{Exit: 0}
	refused := observation{Exit: 1, Stderr: "adamic: x: stage 0 can't lower view yet"}
	cases := []struct {
		a, b observation
		want string
	}{
		{accepted, accepted, "accepted-by-both"}, {refused, refused, "refused-by-both"},
		{refused, accepted, "newly-accepted"}, {accepted, refused, "newly-refused"},
		{observation{Exit: 1, Stderr: "panic: broken"}, accepted, "compiler-error"},
		{observation{Exit: -1, Error: "timeout"}, refused, "compiler-error"},
		{observation{Exit: 1, Stderr: "missing input"}, accepted, "compiler-error"},
	}
	for _, c := range cases {
		if got := classify(c.a, c.b); got != c.want {
			t.Fatalf("classify = %s, want %s", got, c.want)
		}
	}
}
