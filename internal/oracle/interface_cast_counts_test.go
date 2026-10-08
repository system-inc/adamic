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
		rows = append(rows, counted(t, interfaceSource(name), false, nil, false, false))
	}
	// These rows count the source lowering and shared readiness helpers. The checked
	// view substitution in TestNarrowedFieldUsesSharedReadiness is a separate IR probe.
	for _, name := range []string{"identifier", "identifier-uninitialized", "number", "number-uninitialized"} {
		rows = append(rows, counted(t, interfaceSource("readiness-"+name), false, nil, false, false))
	}
	for _, name := range []string{
		"lane1/objects-good", "lane1/objects-untagged-good", "lane1/objects-untagged-wrong",
		"lane1/objects-wrong-nested", "lane1/objects-missing-nested", "lane1/objects-uninitialized-nested",
		"lane1/interfaces-good", "lane1/interfaces-missing-inherited", "lane1/interfaces-wrong-inherited",
		"lane1/interfaces-uninitialized-object", "lazy/unread", "lazy/unread-untagged", "lazy/ordinary-disjoint", "lazy/optional-unread",
	} {
		rows = append(rows, counted(t, interfaceSource(name), false, nil, false, false))
	}
	for _, path := range []string{
		"stage3/fixtures/assertions/11_parenthesized_kind.a",
		"stage3/fixtures/assertions/13_flag_downcast.a",
		"stage3/fixtures/assertions/20_type_flag_mask.a",
	} {
		rows = append(rows, counted(t, path, false, nil, false, false))
	}
	return rows
}
