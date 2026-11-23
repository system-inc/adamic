package oracle

import (
	"strings"
	"testing"
)

// A cast from a union of plain objects no literal tag tells apart is a checked view of its target with no test at
// the cast (#axqgdck, rulings.md:18). A lie stops where it's read or written, with the view's own message, on both
// backends alike, never at the cast and never reading or writing through, while Node runs on: a field the value
// lacks read through the view, a field read in a callee compiled for the target after an honest call agrees
// (Language's y04: the view's checks travel with the value), and a write to a field the value lacks (Language's
// y13). Mutant: untaggedUnionCast returning the value without its view, and the first lie reads the missing slot
// raw, native no longer stopping in the view's words.
func TestUntaggedUnionViewStopsAtTheRead(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout, node string }{
		{"cast_untagged_union_lie.ts", "casting\n", "casting\nundefined\n"},
		{"cast_untagged_union_callee_lie.ts", "string\n", "string\nnumber\n"},
		{"cast_untagged_union_missing_write.ts", "", "nb\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			program, path := castFixture(t, probe.file)
			native, _ := natively(t, program)
			javaScript := onJavaScriptBackend(t, program)
			for backend, result := range map[string]run{"native": native, "JavaScript": javaScript} {
				if result.exitCode != 70 || string(result.stdout) != probe.stdout || !strings.HasPrefix(string(result.stderr), "adamic: panic: field read failed:") || strings.Contains(string(result.stderr), "cast failed") {
					t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the read or write, never at the cast", backend, result.exitCode, result.stdout, result.stderr)
				}
			}
			if string(native.stderr) != string(javaScript.stderr) {
				t.Errorf("the backends stop differently: native %q, JavaScript %q", native.stderr, javaScript.stderr)
			}
			// The stop's words, for its negative witness (cloud/admission-corpus/negative-witnesses.json).
			t.Logf("stderr %q", native.stderr)
			if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
			}
		})
	}
}

func init() {
	// The honest fixtures agree with Node on every backend, the counted build freeing all they made; the lies stop,
	// so each is held to the JavaScript backend.
	for _, fixture := range []struct {
		file    string
		checked bool
	}{
		{"cast_untagged_union.ts", false},
		{"cast_untagged_union_writes.ts", false},
		{"cast_untagged_union_cycle.ts", false},
		{"cast_untagged_union_branded.ts", false},
		{"cast_untagged_union_lie.ts", true},
		{"cast_untagged_union_callee_lie.ts", true},
		{"cast_untagged_union_missing_write.ts", true},
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + fixture.file, true, fixture.checked})
	}
}
