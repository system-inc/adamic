package oracle

import (
	"strings"
	"testing"
)

// A cast from a union of plain objects no literal tag tells apart is a checked view of its target with no test at
// the cast (#axqgdck, rulings.md:18): a value of another member that lacks a field the target reads stops at that
// read, with the view's own message, on both backends alike, never at the cast and never reading through, while
// Node runs on and prints undefined. Mutant: untaggedUnionCast returning the value without its view reads the
// missing slot raw, and native no longer stops in the view's words.
func TestUntaggedUnionViewStopsAtTheRead(t *testing.T) {
	t.Parallel()
	program, path := castFixture(t, "cast_untagged_union_lie.ts")
	native, _ := natively(t, program)
	javaScript := onJavaScriptBackend(t, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": javaScript} {
		if result.exitCode != 70 || string(result.stdout) != "casting\n" || !strings.HasPrefix(string(result.stderr), "adamic: panic: field read failed:") || strings.Contains(string(result.stderr), "cast failed") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the read, never at the cast", backend, result.exitCode, result.stdout, result.stderr)
		}
	}
	if string(native.stderr) != string(javaScript.stderr) {
		t.Errorf("the backends stop differently: native %q, JavaScript %q", native.stderr, javaScript.stderr)
	}
	// The stop's words, for its negative witness (cloud/admission-corpus/negative-witnesses.json).
	t.Logf("stderr %q", native.stderr)
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "casting\nundefined\n" {
		t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
	}
}

func init() {
	// cast_untagged_union.ts agrees with Node on every backend; the lie stops, so it's held to the JavaScript backend.
	for _, fixture := range []struct {
		file    string
		checked bool
	}{{"cast_untagged_union.ts", false}, {"cast_untagged_union_lie.ts", true}} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + fixture.file, true, fixture.checked})
	}
}
