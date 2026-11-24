package main

import (
	"strings"
	"testing"
)

// A program whose runs meet a first-link trap is classed apart, never agreeing nor a checked stop, even where Node
// and both backends write the same trap and stop alike, and it fails the run, named on its own line. The program
// beside it that agrees is still counted agreeing. Mutant: the first-link class dropped, and the trapped program
// counts agreeing, so the run passes.
func TestAgreementClassesAFirstLinkTrapApart(t *testing.T) {
	t.Parallel()
	const trap = "adamic: first-link trap: NotYet: Math.random at main.a:5:9"
	trapped := "stop: " + trap + "\nnative: echo '" + trap + "' >&2\nnative: exit 70\n"
	dir, sha, head := agreementFixture(t, []agreementSource{{"corpus/agree.a", agreeingSource}, {"corpus/trapped.a", trapped}})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.Verdict != "fail" || r.TrappedPrograms != 1 || r.Agreeing != 1 || r.CheckedStops != 0 || r.Compared != 1 {
		t.Fatalf("exit %d, %+v, want exit 1 with the trapped program apart and the other agreeing: %s", o.Exit, r, o.Stderr)
	}
	judgments := map[string]string{}
	for _, judged := range r.Programs {
		judgments[judged.Path] = judged.Judgment
	}
	if judgments["corpus/trapped.a"] != "first-link trap" || judgments["corpus/agree.a"] != "agreeing" {
		t.Fatalf("judgments %v, want corpus/trapped.a a first-link trap and corpus/agree.a agreeing", judgments)
	}
	if line := stderrLine(o.Stderr, "first-link trap corpus/trapped.a:"); !strings.Contains(line, firstLinkTrapPrefix) {
		t.Fatalf("the trapped program's line: %q in %s", line, o.Stderr)
	}
}
