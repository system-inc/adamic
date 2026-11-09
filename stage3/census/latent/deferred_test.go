package latent_test

import (
	"os"
	"strings"
	"testing"
)

func TestLatentDeferredDriverHooksZb0q0zd(t *testing.T) {
	t.Parallel()
	for file, hooks := range map[string][]string{
		"lower.go.txt":                     {"fresh().latentScan(file)", "if l.latentDeferredUnit(node)", "latent.DeferredStop(ordinary, probe.refusal.deferred)", `latentFindingPhase = "deferred"`, "if len(latentFindings) != first"},
		"units.go.txt":                     {"fresh().latentScan(file)", "if l.latentDeferredUnit(node)"},
		"speculative.go.txt":               {"l.latentScan(file)", "copy.latentDeferredUnit(node)"},
		"../predicates/tools/lower.go.txt": {"fresh().latentScan(file)", "if l.latentDeferredUnit(node)", "latent.DeferredStop(ordinary, probe.refusal.deferred)"},
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, hook := range hooks {
			if strings.Count(string(source), hook) != 1 {
				t.Errorf("%s: missing or repeated deferred hook %q", file, hook)
			}
		}
	}
}
