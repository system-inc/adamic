package latent

import (
	"strings"
	"testing"
)

func TestLatentPrepassesDiscardedMutant(t *testing.T) {
	t.Parallel()
	path := "stage3/census/latent/units.go.txt"
	source := latentPrepassRead(t, "../../../"+path)
	mutant := strings.Replace(source, "orderErr = base.collectViewedElements(modules)", "base.collectViewedElements(modules)", 1)
	if mutant == source {
		t.Fatal("mutant did not discard collectViewedElements error")
	}
	methods, err := latentErrorPrepasses("../../..", latentPrepassRead(t, "../../../internal/lower/lower.go"))
	if err != nil {
		t.Fatal(err)
	}
	gaps := latentDiscardedPrepassErrors(methods, map[string]string{path: mutant})
	if !strings.Contains(strings.Join(gaps, "\n"), path+": production prepass collectViewedElements error discarded") {
		t.Fatalf("mutant survived: %v", gaps)
	}
	t.Logf("guard red as required: %s", strings.Join(gaps, "\n"))
}
