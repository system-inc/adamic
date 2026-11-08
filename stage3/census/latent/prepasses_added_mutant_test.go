package latent

import (
	"strings"
	"testing"
)

func TestLatentPrepassesAddedProductionMutantSygfnmc(t *testing.T) {
	t.Parallel()
	source := latentPrepassRead(t, "../../../internal/lower/lower.go")
	mutant := strings.Replace(source, "lowering.noteOmittedOptionals(modules)", "lowering.noteOmittedOptionals(modules)\n lowering.fakeCensusPrepass(modules)", 1)
	if mutant == source {
		t.Fatal("mutant did not add fakeCensusPrepass")
	}
	templates := map[string]string{}
	for _, path := range latentPrepassTemplates {
		templates[path] = latentPrepassRead(t, "../../../"+path)
	}
	gaps := latentPrepassGaps(mutant, templates)
	if len(gaps) != len(templates) {
		t.Fatalf("mutant must fail every template: %v", gaps)
	}
	for _, gap := range gaps {
		if !strings.Contains(gap, "fakeCensusPrepass") {
			t.Fatal(gap)
		}
		t.Log("guard red as required: " + gap)
	}
}
