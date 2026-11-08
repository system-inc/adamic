package latent

import (
	"strings"
	"testing"
)

func TestLatentPrepassesOmissionMutantSygfnmc(t *testing.T) {
	t.Parallel()
	path := "stage3/census/latent/lower.go.txt"
	source := latentPrepassRead(t, "../../../"+path)
	mutant := strings.Replace(source, "l.noteOmittedOptionals(modules)", "", 1)
	if mutant == source {
		t.Fatal("mutant did not delete noteOmittedOptionals")
	}
	gaps := latentPrepassGaps(latentPrepassRead(t, "../../../internal/lower/lower.go"), map[string]string{path: mutant})
	if len(gaps) == 0 || !strings.Contains(strings.Join(gaps, "\n"), path+": production prepass noteOmittedOptionals") {
		t.Fatalf("mutant survived: %v", gaps)
	}
	t.Logf("guard red as required: %s", strings.Join(gaps, "\n"))
}
