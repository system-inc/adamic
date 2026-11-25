package oracle

import (
	"strings"
	"testing"
)

// A run that met a first-link trap disagrees with every run, one just like it included, so no comparison the
// oracle makes passes a program lowered in trap mode. Mutant: disagreement's first-link case dropped, and two
// runs trapping alike agree.
func TestFirstLinkTrapRunNeverAgrees(t *testing.T) {
	t.Parallel()
	trapped := run{stdout: []byte("first ran\n"), stderr: []byte("adamic: first-link trap: NotYet: Math.random at main.a:5:9\n"), exitCode: 75}
	for name, pair := range map[string][2]run{
		"both trapped alike": {trapped, trapped},
		"the backend":        {{stdout: []byte("first ran\n")}, trapped},
		"the oracle":         {trapped, {stdout: []byte("first ran\n")}},
	} {
		if got := disagreement(pair[0], pair[1]); !strings.Contains(got, "first-link trap") {
			t.Errorf("%s: disagreement %q, want the first-link trap named", name, got)
		}
	}
	same := run{stdout: []byte("first ran\n"), stderr: []byte("adamic: panic: why\n"), exitCode: 70}
	if got := disagreement(same, same); got != "" {
		t.Errorf("two runs alike with no trap: disagreement %q, want none", got)
	}
}
