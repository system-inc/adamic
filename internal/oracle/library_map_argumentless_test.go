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
func TestArgumentlessMapTakesItsSlotsArguments(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_map_argumentless/admitted.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("an argumentless new Map() in a typed slot must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 {
		t.Fatalf("Node: %+v", node)
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("%s disagrees with Node: %s", backend, difference)
		}
	}
}

// With no slot to fill, new Map() is Map<unknown, unknown>, whose keys lowering can't hold yet.
func TestArgumentlessMapWithoutContextStops(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_map_argumentless/uncontextual.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	const what = "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"
	if !errors.As(err, &notYet) || notYet.What != what || notYet.Where != path+":1:15" {
		t.Fatalf("want NotYet %q at :1:15, got %v", what, err)
	}
}
