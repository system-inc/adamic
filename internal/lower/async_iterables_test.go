package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestAsyncUserIteratorFunctionRemainsUnproven(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const source = { [Symbol.iterator]() { return { next() { return { value: 1, done: false }; } }; } };
async function f(): Promise<void> { for (const item of source) { await Promise.resolve(); break; } }
await f();
`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "async for-of over a user iterator") {
		t.Fatalf("want the explicit user iterator proof gap: %v", err)
	}
}

func TestAsyncUserIteratorModuleRemainsUnproven(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const source = { [Symbol.iterator]() { return { next() { return { value: 1, done: false }; } }; } };
for (const item of source) { await Promise.resolve(); break; }
`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "async for-of over a user iterator") {
		t.Fatalf("want the explicit user iterator proof gap: %v", err)
	}
}
