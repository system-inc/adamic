package oracle

import "testing"

// A .ts literal cast sideways to a target it only partly writes, of fields held every way (N4 head B on
// #b5w3ycg): what it wrote reads back, and an optional field it didn't write reads undefined, natively, released
// and on JavaScript as on Node. Mutant: an optional unwritten field laid Uninitialized too; this program still
// agrees (an absent slot reads undefined whatever its readiness says), and lower's
// TestWidenedConstructionLaysUnwrittenFields is the one that goes red.
func TestConstructionPartialObjectAgreesWithNode(t *testing.T) {
	t.Parallel()
	program, path := castFixture(t, "construction_partial_object.ts")
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "partial\n1\nfirst\nboss\nundefined\nno size\nno extra\nwhole 2 boss whole\n" {
		t.Fatalf("Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	native, binary := natively(t, program)
	for backend, result := range map[string]run{"native": native, "native release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// The same construction's required object field, unwritten, stops at its read on both backends with the read's
// own words, where Node reads undefined and runs on (held-stops.json's row for this program).
func TestConstructionPartialUnsetReadStops(t *testing.T) {
	t.Parallel()
	program, path := castFixture(t, "construction_partial_unset_read_fails.ts")
	if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "partial\nno parent\n" {
		t.Fatalf("Node must run on past the unwritten read: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	want := run{exitCode: 70, stdout: []byte("partial\n"), stderr: []byte("adamic: panic: read before assignment: field 'parent' in partial.parent\n")}
	native, _ := natively(t, program)
	for backend, result := range map[string]run{"native": native, "native release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("%s: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
}
