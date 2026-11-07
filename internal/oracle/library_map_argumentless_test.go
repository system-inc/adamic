package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An argumentless new Map() takes its key and value types from the slot it fills, through the
// prelude's MapConstructor overload (axiom C2), so a declared local, a branded key, ??= into an
// optional map, a class field, a return and an interface value each lower and match Node natively
// and on the JavaScript backend. Mutant: the prelude's MapConstructor removed, which types each one
// Map<any, any> and stops the program at "a Map whose keys aren't strings".
func TestArgumentlessMapTakesItsSlotsArguments(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/library_map_argumentless/admitted.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("an argumentless new Map() in a typed slot must lower: %v", loweredError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 {
		test.Fatalf("Node: %+v", node)
	}
	native, _ := natively(test, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(node, got); difference != "" {
			test.Fatalf("%s disagrees with Node: %s", backend, difference)
		}
	}
}

// With no slot to fill, new Map() is Map<unknown, unknown>, whose keys lowering can't hold yet.
func TestArgumentlessMapWithoutContextStops(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/library_map_argumentless/uncontextual.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	const what = "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"
	if !errors.As(loweredError, &notYet) || notYet.What != what || notYet.Where != path+":1:15" {
		test.Fatalf("want NotYet %q at :1:15, got %v", what, loweredError)
	}
}
