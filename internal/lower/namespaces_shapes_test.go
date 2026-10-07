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
	for _, test := range []struct{ name, reason string }{
		{"BuilderState", ""}, {"JsxNames", ""}, {"ReactNames", ""}, {"BinaryExpressionState", ""},
		{"Parser.JSDocParser", ""},
		{"Debug", "class inside a namespace"},
		{"Debug.log", "callable object properties"},
		{"Parser", "function without a body"},
		{"IncrementalParser", "function without a body"},
		{"tracingEnabled", "no runtime container is emitted"},
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
