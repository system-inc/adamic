package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// External types and bodies are normalized; declaration forms are retained, including blockers.
func TestTscNamespaceDeclarationShapes(test *testing.T) {
	test.Parallel()
	// Namespace overloads must follow the module-level implementation policy.
	// Integration admits checked overloads; this branch still reports bodyless signatures.
	_, overloadError := lowerSource(
		test,
		"function read():number; function read():number; function read():number{return 0;} console.log(`${read()}`);",
	)
	overloadReason := ""
	if overloadError != nil {
		var notYet *NotYet
		if !errors.As(overloadError, &notYet) || !strings.Contains(overloadError.Error(), "function without a body") {
			test.Fatalf("unexpected module overload outcome: %v", overloadError)
		}
		overloadReason = "function without a body"
	}
	for _, testCase := range []struct{ name, reason string }{
		{"BuilderState", ""}, {"JsxNames", ""}, {"ReactNames", ""}, {"BinaryExpressionState", ""},
		{"Parser.JSDocParser", ""},
		{"Debug", ""},
		{"Debug.log", ""},
		{"Parser", overloadReason},
		{"IncrementalParser", overloadReason},
		{"tracingEnabled", "no runtime container is emitted"},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join("../../stage3/namespaces/shapes", testCase.name+".a")
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			_, lowerSourceError := lowerSource(test, string(source))
			if testCase.reason == "" {
				if lowerSourceError != nil {
					test.Fatal(lowerSourceError)
				}
				test.Log("declaration shape lowers")
				return
			}
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), testCase.reason) {
				test.Fatalf("got %v, want NotYet %s", lowerSourceError, testCase.reason)
			}
			test.Log(lowerSourceError)
		})
	}
}
