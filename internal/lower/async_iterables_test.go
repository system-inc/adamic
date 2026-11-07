package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestAsyncUserIteratorFunctionRemainsUnproven(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`const source = { [Symbol.iterator]() { return { next() { return { value: 1, done: false }; } }; } };
async function f(): Promise<void> { for (const item of source) { await Promise.resolve(); break; } }
await f();
`,
	)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "async for-of over a user iterator") {
		test.Fatalf("want the explicit user iterator proof gap: %v", err)
	}
}

func TestAsyncUserIteratorModuleRemainsUnproven(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`const source = { [Symbol.iterator]() { return { next() { return { value: 1, done: false }; } }; } };
for (const item of source) { await Promise.resolve(); break; }
`,
	)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "async for-of over a user iterator") {
		test.Fatalf("want the explicit user iterator proof gap: %v", err)
	}
}
