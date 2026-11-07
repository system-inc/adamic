package main

import (
	"reflect"
	"testing"
	"time"
)

func TestClassification(test *testing.T) {
	test.Parallel()
	for _, o := range []observation{{Exit: -1, Error: "timeout"}, {Exit: 1, Stderr: "panic: broken"}} {
		if compileClass(o) != "error" {
			test.Fatal("compiler failure treated as refusal")
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
			test.Fatalf("classify = %s, want %s", got, c.want)
		}
	}
}

func TestOutputsAgree(test *testing.T) {
	test.Parallel()
	good := observation{Exit: 0, Stdout: "42\n"}
	if !outputsAgree(good, good, good) {
		test.Fatal("equal outputs rejected")
	}
	for _, bad := range []observation{
		{Exit: 0, Stdout: "41\n"},
		{Exit: 1, Stdout: "42\n"},
		{Exit: 0, Stdout: "42\n", Error: "timeout"},
		{Exit: -1, Stdout: "42\n"},
	} {
		if outputsAgree(good, good, bad) || outputsAgree(good, bad, good) || outputsAgree(bad, good, good) {
			test.Fatalf("mismatch accepted: %+v", bad)
		}
	}
}

func TestSampling(test *testing.T) {
	test.Parallel()
	programs := []entry{{program: program{Path: "w"}, Corpus: "witnesses", Class: "newly-accepted"}}
	for _, path := range []string{"d", "a", "c", "b"} {
		programs = append(programs, entry{program: program{Path: path}, Corpus: "fixtures", Class: "newly-accepted"})
	}
	seed, samplingSeedError := samplingSeed("head-sha", "")
	if samplingSeedError != nil {
		test.Fatal(samplingSeedError)
	}
	selected, omitted := sample(programs, seed, 15, time.Second)
	if len(selected) != 3 || omitted != 2 || selected[0] != 0 {
		test.Fatalf("sample: %v, omitted %d", selected, omitted)
	}
	again, _ := sample(programs, seed, 15, time.Second)
	if !reflect.DeepEqual(selected, again) {
		test.Fatal("nondeterministic sample")
	}
	selected, omitted = sample(programs, seed, 1, time.Second)
	if len(selected) != 1 || selected[0] != 0 || omitted != 4 {
		test.Fatal("witness lost under budget")
	}
	selected, omitted = sample(programs, seed, 0, time.Second)
	if len(selected) != 5 || omitted != 0 {
		test.Fatal("unlimited budget omitted programs")
	}
}
