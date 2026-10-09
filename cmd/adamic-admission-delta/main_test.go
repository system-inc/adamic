package main

import (
	"reflect"
	"testing"
	"time"
)

func TestClassification(t *testing.T) {
	t.Parallel()
	for _, o := range []observation{{Exit: -1, Error: "timeout"}, {Exit: 1, Stderr: "panic: broken"}} {
		if compileClass(o) != "error" {
			t.Fatal("compiler failure treated as refusal")
		}
	}
	accepted := observation{Exit: 0}
	refused := observation{Exit: 1, Stderr: "adamic: x: stage 0 can't lower view yet"}
	cases := []struct {
		a, b observation
		want string
	}{
		{accepted, accepted, "accepted-by-both"}, {refused, refused, "refused-by-both"},
		{refused, accepted, "newly-accepted"}, {accepted, refused, "newly-refused"},
		{observation{Exit: 1, Stderr: "panic: broken"}, accepted, "compiler-crash"},
		{observation{Exit: -1, Error: "timeout"}, refused, "compiler-timeout"},
		{observation{Exit: 1, Stderr: "missing input"}, accepted, "compiler-error"},
	}
	for _, c := range cases {
		if got := classify(c.a, c.b); got != c.want {
			t.Fatalf("classify = %s, want %s", got, c.want)
		}
	}
}

func TestOutputsAgree(t *testing.T) {
	t.Parallel()
	good := observation{Exit: 0, Stdout: "42\n"}
	if !outputsAgree(good, good, good) {
		t.Fatal("equal outputs rejected")
	}
	for _, bad := range []observation{{Exit: 0, Stdout: "41\n"}, {Exit: 1, Stdout: "42\n"}, {Exit: 0, Stdout: "42\n", Error: "timeout"}, {Exit: -1, Stdout: "42\n"}} {
		if outputsAgree(good, good, bad) || outputsAgree(good, bad, good) || outputsAgree(bad, good, good) {
			t.Fatalf("mismatch accepted: %+v", bad)
		}
	}
}

func TestSampling(t *testing.T) {
	t.Parallel()
	programs := []entry{{program: program{Path: "w"}, Corpus: "witnesses", Class: "newly-accepted"}}
	for _, path := range []string{"d", "a", "c", "b"} {
		programs = append(programs, entry{program: program{Path: path}, Corpus: "fixtures", Class: "newly-accepted"})
	}
	selected, omitted := sample(programs, "head-sha", 15, time.Second)
	if len(selected) != 3 || omitted != 2 || selected[0] != 0 {
		t.Fatalf("sample: %v, omitted %d", selected, omitted)
	}
	again, _ := sample(programs, "head-sha", 15, time.Second)
	if !reflect.DeepEqual(selected, again) {
		t.Fatal("nondeterministic sample")
	}
	selected, omitted = sample(programs, "head-sha", 1, time.Second)
	if len(selected) != 1 || selected[0] != 0 || omitted != 4 {
		t.Fatal("witness lost under budget")
	}
	selected, omitted = sample(programs, "head-sha", 0, time.Second)
	if len(selected) != 5 || omitted != 0 {
		t.Fatal("unlimited budget omitted programs")
	}
}
