package oracle

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/oracle/countproducts"
)

// TestProduct_Adamic declares the stage 0 compiler interfaceCastCountCases runs, so this package's units are handed it;
// internal/native's TestProduct_WASICompiler builds it (buildcache.AdamicOwner), and this one finds it.
func TestProduct_Adamic(t *testing.T) {
	t.Parallel()
	buildcache.AdamicBuiltElsewhere(t)
}

// Default cast fixtures use the same counted runtime and table as ordinary fixtures, each in its own committed count file.
func interfaceCastCountCases() []countCase {
	var cases []countCase
	for _, name := range []string{"visitor", "wrong-kind"} {
		path := interfaceFixturePath(name)
		cases = append(cases, countCase{path, func(t *testing.T) (string, func() string) { return interfaceCastCounted(t, path) }})
	}
	var paths []string
	for _, name := range []string{"default-staged", "default-wrong-type", "default-wrong-boolean", "default-literal", "default-boxed-string", "default-boxed-write", "default-destructure", "default-read-before-set"} {
		paths = append(paths, interfaceFixturePath(name))
	}
	// These rows count the source lowering and shared readiness helpers. The checked
	// view substitution in TestNarrowedFieldUsesSharedReadiness is a separate IR probe.
	for _, name := range []string{"identifier", "identifier-uninitialized", "number", "number-uninitialized"} {
		paths = append(paths, "stage3/interface-downcasts/readiness-"+name+".ts")
	}
	for _, name := range []string{
		"interfaces-good", "interfaces-missing-inherited", "interfaces-uninitialized-object", "interfaces-wrong-inherited",
		"objects-good", "objects-missing-nested", "objects-uninitialized-nested", "objects-untagged-good", "objects-untagged-wrong", "objects-wrong-nested", "optional-read-control", "flag-downcast-wrong",
	} {
		paths = append(paths, interfaceFixturePath("lane1/"+name))
	}
	for _, path := range paths {
		cases = append(cases, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
	return cases
}

// interfaceCastCounted is a cast fixture compiled by the stage 0 compiler's default, the C its row runs, and measure,
// which builds that C counted, runs it and returns the row, as its shard's product does (countRows).
func interfaceCastCounted(t *testing.T, path string) (string, func() string) {
	t.Helper()
	// The stage 0 compiler is a product built ahead, run here as go run ran it (#5qykzj5): its owner builds it
	// (buildcache.AdamicOwner), so a count product measuring this row finds it rather than building it again.
	command := exec.Command(buildcache.AdamicBuiltElsewhere(t), "c", path)
	command.Dir = repository

	source, err := command.Output()
	if err != nil {
		t.Fatalf("default compiler for %s: %v", path, err)
	}
	// wrong-kind.ts's cast fails, as it's meant to, with its counts written as it stops.
	exit := 0
	if strings.HasSuffix(path, "wrong-kind.ts") {
		exit = 70
	}
	row := countproducts.Row{Key: path, C: string(source), Run: countproducts.Plain, Checked: true, Exit: exit}
	countRows.Store(path, row)
	return countPortable(t, string(source), ""), func() string { return countMeasuredHere(t, row) }
}
