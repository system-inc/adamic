package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// External types and bodies are normalized; declaration forms are retained, including blockers.
func TestTscNamespaceDeclarationShapes(t *testing.T) {
	t.Parallel()
	// Namespace overloads must follow the module-level implementation policy.
	// Integration admits checked overloads; this branch still reports bodyless signatures.
	_, overloadError := lowerSource(t, "function read():number; function read():number; function read():number{return 0;} console.log(`${read()}`);")
	overloadReason := ""
	if overloadError != nil {
		var notYet *NotYet
		if !errors.As(overloadError, &notYet) || !strings.Contains(overloadError.Error(), "function without a body") {
			t.Fatalf("unexpected module overload outcome: %v", overloadError)
		}
		overloadReason = "function without a body"
	}
	for _, test := range []struct{ name, reason string }{
		{"BuilderState", ""}, {"JsxNames", ""}, {"ReactNames", ""}, {"BinaryExpressionState", ""},
		{"Parser.JSDocParser", ""},
		{"Debug", ""},
		{"Debug.log", ""},
		{"Parser", overloadReason},
		{"IncrementalParser", overloadReason},
		{"tracingEnabled", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join("../../stage3/namespaces/shapes", test.name+".a")
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			if test.reason == "" {
				if err != nil {
					t.Fatal(err)
				}
				t.Log("declaration shape lowers")
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("got %v, want NotYet %s", err, test.reason)
			}
			t.Log(err)
		})
	}
}
