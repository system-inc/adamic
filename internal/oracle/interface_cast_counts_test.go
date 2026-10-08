package oracle

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Default cast fixtures use the same counted runtime and table as ordinary fixtures.
func interfaceCastCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, path := range []string{"stage3/interface-downcasts/visitor.a", "stage3/interface-downcasts/wrong-kind.a"} {
		command := exec.Command("go", "run", "./cmd/adamic", "c", path)
		command.Dir = repository

		source, err := command.Output()
		if err != nil {
			t.Fatalf("default compiler for %s: %v", path, err)
		}
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(string(source), binary, native.Options{Count: true}); err != nil {
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
		rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	for _, name := range []string{"default-staged", "default-wrong-type", "default-wrong-boolean", "default-literal", "default-boxed-string", "default-boxed-write", "default-destructure", "default-read-before-set"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/"+name+".a"), false, nil, false, false))
	}
	for _, name := range []string{"objects-good", "objects-untagged-good", "objects-untagged-wrong", "objects-wrong-nested", "objects-missing-nested", "objects-uninitialized-nested"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane1/"+name+".a"), false, nil, false, false))
	}
	for _, name := range []string{"interfaces-good", "interfaces-missing-inherited", "interfaces-wrong-inherited", "interfaces-uninitialized-object"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane1/"+name+".a"), false, nil, false, false))
	}
	for _, name := range []string{"unions-objects-good", "unions-objects-wrong-payload", "unions-objects-wrong-tag", "unions-objects-missing-tag"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane1/"+name+".a"), false, nil, false, false))
	}
	// These rows count the source lowering and shared readiness helpers. The checked
	// view substitution in TestNarrowedFieldUsesSharedReadiness is a separate IR probe.
	for _, name := range []string{"identifier", "identifier-uninitialized", "number", "number-uninitialized"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/readiness-"+name+".a"), false, nil, false, false))
	}
	rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane4b/fixtures/fixes-graph.a"), false, nil, false, false))
	rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane4b/fixtures/fixes-tuple-unread.a"), false, nil, false, false))
	rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane4b/fixtures/fixes-tuple.a"), false, nil, false, false))
	rows = append(rows, viewCallableCounts(t)...)
	rows = append(rows, tupleOriginalCounts(t)...)
	rows = append(rows, viewRankedArrayCounts(t)...)
	return append(rows, viewIntersectionDeferredCounts(t)...)
}
