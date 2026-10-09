package oracle

import (
	"path/filepath"
	"testing"
)

func TestProgramRegionHost14AgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host/14_getCurrentDirectory.a"))
	if err != nil {
		t.Fatal(err)
	}
	counts := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	if counts[5] != 4 {
		t.Fatalf("host 14 adopts %d members, want its two cells and two closures", counts[5])
	}
	t.Log("host 14 matches Node in both backends with sanitizers and leak checks; it contains no parser Node or NodeArray")
}
