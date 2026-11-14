package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These run the real node under --timeout, beside fakes that sleep past it, so the limit leaves node room on a
// loaded box: at 1s, node itself timed out on Lunatic and its timeout joined the reasons (J-tax3dmj). The fakes sleep
// 10s against 4s, so a sleep can't pass for a run that finished.

// A runtime build that times out under the gate's load is retried alone before its timeout counts, so a clean head
// reads as a pass, with the timeout and the retry said (#pajbaba, seen on #ny0c3gb's proof at --workers 4). The head's
// first native build sleeps past the limit and every later one doesn't. Mutant
// (timeout-retry.mutants.md, (a)): observeAlone returning the first run's timeouts as still timed out, and the run is
// void.
func TestRuntimeTimeoutRetriedAloneIsNotCounted(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo \"console.log('42');\";;\nbuild) slept=\"$(dirname \"$0\")/slept\"; if [ ! -e \"$slept\" ]; then : > \"$slept\"; sleep 10; fi; printf '#!/bin/sh\\necho 42\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json", "--timeout", "4s")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	if o.Exit != 0 || r.Void || r.Verdict != "pass" || r.TimedOutRuns != 1 || r.RetriedRuns != 1 || r.StillTimedOutRuns != 0 {
		t.Fatalf("exit %d, void %v, verdict %s, timed out %d, retried %d, still %d, want a pass with one retried timeout: %s", o.Exit, r.Void, r.Verdict, r.TimedOutRuns, r.RetriedRuns, r.StillTimedOutRuns, o.Stderr)
	}
	if record := r.Programs[0]; !record.Retried || !reflect.DeepEqual(record.TimedOut, []string{"native compile"}) || record.Agree == nil || !*record.Agree {
		t.Fatalf("the program's record: retried %v, timed out %q, agree %v", record.Retried, record.TimedOut, record.Agree)
	}
	if !strings.Contains(o.Stderr, "runtime timeouts: 1 runs timed out, 1 retried alone, 0 timed out again") {
		t.Fatalf("the summary doesn't say the timeout and its retry: %s", o.Stderr)
	}
}

// A program whose commands time out again alone is void, and every command that timed out is its reason, not just the
// first. Mutant (b): timedOut naming only the first command, and the reason loses native compile.
func TestEveryTimeoutIsItsReason(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) sleep 10;;\nbuild) sleep 10;;\nesac\n")
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json", "--timeout", "4s")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	const reason = "corpus/input.a: javascript compile, native compile timeout"
	if o.Exit != 3 || !r.Void || r.Verdict != "void" || r.VoidReason != reason || r.StillTimedOutRuns != 1 {
		t.Fatalf("exit %d, void %v, verdict %s, reason %q, still %d, want exit 3 and %q: %s", o.Exit, r.Void, r.Verdict, r.VoidReason, r.StillTimedOutRuns, reason, o.Stderr)
	}
}

// A checked-boundary witness whose repair's build times out, alone too, is unjudged: the program is void with the
// repair's timeout as its reason, never a red. That was the one fail on #ny0c3gb's proof (views_v3/values-lie.ts,
// repair_agree false under load). Mutant (c): the repair's timeouts left out of the reasons, and the run fails.
func TestRepairTimeoutVoidsNotFails(t *testing.T) {
	t.Parallel()
	o, r := checkedBoundaryCommandWith(t, true, "input.a", "sleep 10;", "--timeout", "4s")
	if o.Exit != 3 || !r.Void || r.Verdict != "void" || !strings.HasSuffix(r.VoidReason, "repair native compile timeout") || r.Programs[0].NegativeWitness {
		t.Fatalf("exit %d, void %v, verdict %s, reason %q, witness %v, want exit 3 and the repair's timeout: %s", o.Exit, r.Void, r.Verdict, r.VoidReason, r.Programs[0].NegativeWitness, o.Stderr)
	}
	if !reflect.DeepEqual(r.Programs[0].TimedOut, []string{"repair native compile"}) || !r.Programs[0].Retried {
		t.Fatalf("the program's record: timed out %q, retried %v", r.Programs[0].TimedOut, r.Programs[0].Retried)
	}
}
