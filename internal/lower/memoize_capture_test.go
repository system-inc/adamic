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

// Assigning the returned closure into the cell it captures makes the finder's
// same-signature closure edge an actual runtime cycle, without any assertion.
func TestMemoizeSelfCaptureIsCycleCapable(t *testing.T) {
	t.Parallel()
	const source = "function tie(callback: () => string): () => string {\n    const result = () => callback();\n    callback = result;\n    return result;\n}\nconst get = tie(() => 'cwd');\nconsole.log('made');\n"
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "'callback', a variable a function value captures") || !strings.Contains(refused.Error(), "adamic/cycle-capable") {
		t.Fatalf("want the self-capturing callback cell refused, got %v", err)
	}
}
