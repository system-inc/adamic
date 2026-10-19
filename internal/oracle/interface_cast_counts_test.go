package oracle

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
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
	for _, path := range []string{"stage3/interface-downcasts/visitor.a", "stage3/interface-downcasts/wrong-kind.a"} {
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
		paths = append(paths, "stage3/interface-downcasts/lane1/"+name+".a")
	}
	for _, path := range paths {
		cases = append(cases, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
	return cases
}

// interfaceCastCounted is a cast fixture compiled by the stage 0 compiler's default, the C its row runs, and measure,
// which builds that C counted, runs it and returns the row.
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
	return countPortable(t, string(source), ""), func() string { return interfaceCastRun(t, path, string(source)) }
}

// interfaceCastRun builds a cast fixture's C counted, runs it and returns its row.
func interfaceCastRun(t *testing.T, path string, source string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, arguments := pinnedStack(binary)
	result := execute(t, name, arguments...)
	wantedExit := 0
	if strings.HasSuffix(path, "wrong-kind.a") {
		wantedExit = 70
	}
	if result.exitCode != wantedExit {
		t.Fatalf("%s counted exit %d stderr %q", path, result.exitCode, result.stderr)
	}
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		t.Fatalf("%s has no runtime counts: %q", path, result.stderr)
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6])
}
