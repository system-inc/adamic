package lower

import (
	"errors"
	"strings"
	"testing"
)

// The main base stops at the assertion, before it can make a cycle claim.
func TestMemoizeCaptureStopsAtAssertion(t *testing.T) {
	t.Parallel()
	const source = "function memoize<T>(callback: () => T): () => T { let value: T; return () => { if (callback) { value = callback(); callback = undefined!; } return value; }; } const get = memoize(() => 'cwd'); console.log(get());\n"
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.HasSuffix(refused.Error(), "main.a:1:127: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case") {
		t.Fatalf("want the pinned main assertion stop, got %v", err)
	}
}

// A store of the returned closure into its captured cell really closes a cycle.
// The synchronous environment and closure must be graph members together.
func TestMemoizeSelfCaptureUsesGraph(t *testing.T) {
	t.Parallel()
	const source = "function tie(callback: () => string): () => string {\n    const result = () => callback();\n    callback = result;\n    return result;\n}\nconst get = tie(() => 'cwd');\nconsole.log('made');\n"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, local := range program.Locals {
		if local.Name == "callback" && local.Captured {
			found = true
			if !local.GraphCell {
				t.Fatal("cyclic callback is not a graph member")
			}
		}
	}
	if !found {
		t.Fatal("no captured callback")
	}
	for _, function := range program.Functions {
		if len(function.Environment) > 0 && !function.GraphClosure {
			t.Fatalf("closure %s is not a graph member", function.Name)
		}
	}
}
