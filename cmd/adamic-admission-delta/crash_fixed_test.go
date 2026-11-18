package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// A crash this change fixes is judged by head alone (Compiler's rule on #ppgk12h); a crash at head, and a
// base failure that may be the box's, still fail as compiler crashes or errors.
func TestCrashFixedClassification(t *testing.T) {
	t.Parallel()
	panicked := observation{Exit: 2, Stderr: "panic: Unhandled case in Node.Text: *ast.ComputedPropertyName\n"}
	compilerBug := observation{Exit: 4, Stderr: "adamic: compiler bug: index out of range\n"}
	accepted := observation{Exit: 0}
	refused := observation{Exit: 1, Stderr: "adamic: x: stage 0 can't lower view yet"}
	for _, c := range []struct {
		name       string
		base, head observation
		class      string
		fixed      bool
	}{
		{"a panic fixed and accepted", panicked, accepted, "newly-accepted", true},
		{"a compiler bug fixed and refused", compilerBug, refused, "newly-refused", true},
		{"a panic at both", panicked, panicked, "compiler-crash", false},
		{"a panic at head", accepted, panicked, "compiler-crash", false},
		{"a runtime fatal error at base, which may be the box's", observation{Exit: 2, Stderr: "fatal error: runtime: out of memory\npanic: x\n"}, accepted, "compiler-crash", false},
		{"a signal at base", observation{Exit: -1, Error: "crash"}, accepted, "compiler-crash", false},
		{"exit 4 with no compiler bug line", observation{Exit: 4, Stderr: "missing input\n"}, accepted, "compiler-error", false},
		{"a panic fixed into an unrecognized error", panicked, observation{Exit: 1, Stderr: "missing input\n"}, "compiler-error", false},
		{"a panic fixed into a timeout", panicked, observation{Exit: -1, Error: "timeout"}, "compiler-timeout", false},
	} {
		if got := classify(c.base, c.head); got != c.class {
			t.Errorf("%s: classify = %s, want %s", c.name, got, c.class)
		}
		if got := crashFixed(c.base, c.head) && strings.HasPrefix(classify(c.base, c.head), "newly-"); got != c.fixed {
			t.Errorf("%s: crash fixed = %v, want %v", c.name, got, c.fixed)
		}
	}
}

// A fixed crash newly accepted is held to Node like any admission: commandFixture's head prints 41 where Node
// prints 42, so the gate fails on the mismatch, and the fix is still said on its own line.
func TestCrashFixedIsHeldToNode(t *testing.T) {
	t.Parallel()
	dir, sha, _, head := commandFixture(t)
	base := filepath.Join(dir, "panicking")
	writeExecutable(t, base, "#!/bin/sh\necho 'panic: Unhandled case in Node.Text: *ast.ComputedPropertyName' >&2\nexit 2\n")
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json", "--timeout", "2s", "--budget", "1")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	if o.Exit != 1 || r.Verdict != "fail" || r.Admitted != 1 || r.Programs[0].Class != "newly-accepted" || !r.Programs[0].CrashFixed {
		t.Fatalf("a fixed crash wasn't held to Node: exit %d %+v %s", o.Exit, r, o.Stderr)
	}
	if p := r.Programs[0]; p.Agree == nil || *p.Agree {
		t.Fatalf("the fixed crash's mismatch wasn't judged: %+v", p)
	}
	if len(r.CrashFixedPaths) != 1 || r.CrashFixedPaths[0] != "corpus/input.a" || !strings.Contains(o.Stderr, "crashes fixed: 1\n") || !strings.Contains(o.Stderr, "crash fixed corpus/input.a (witnesses): newly-accepted at head\n") {
		t.Fatalf("the fix isn't said on its own line: %v\n%s", r.CrashFixedPaths, o.Stderr)
	}
}

// A fixed crash head refuses passes, as any new refusal does, and is still counted.
func TestCrashFixedIntoARefusalPasses(t *testing.T) {
	t.Parallel()
	dir, sha, refusing, _ := commandFixture(t)
	base := filepath.Join(dir, "panicking")
	writeExecutable(t, base, "#!/bin/sh\necho 'panic: Unhandled case in Node.Text: *ast.ComputedPropertyName' >&2\nexit 2\n")
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", refusing, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	if o.Exit != 0 || r.Verdict != "pass" || r.Admitted != 0 || r.Programs[0].Class != "newly-refused" || len(r.CrashFixedPaths) != 1 || !strings.Contains(o.Stderr, "crashes fixed: 1\n") {
		t.Fatalf("a crash fixed into a refusal: exit %d %+v %s", o.Exit, r, o.Stderr)
	}
}

// A crash at head still fails, whatever base did, and is no fix.
func TestCrashAtHeadStillFails(t *testing.T) {
	t.Parallel()
	dir, sha, base, _ := commandFixture(t)
	head := filepath.Join(dir, "panicking")
	writeExecutable(t, head, "#!/bin/sh\necho 'panic: broken' >&2\nexit 2\n")
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	if o.Exit != 1 || r.Verdict != "fail" || r.Programs[0].Class != "compiler-crash" || len(r.CrashFixedPaths) != 0 || !strings.Contains(o.Stderr, "crashes fixed: 0\n") {
		t.Fatalf("a crash at head: exit %d %+v %s", o.Exit, r, o.Stderr)
	}
}
