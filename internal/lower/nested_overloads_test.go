package lower

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// Overload signatures in a function body have no body. tsc's moduleNameResolver.ts
// createCacheWithRedirects declares getOrCreateMap this way; visiting a signature's
// missing body once stopped lowering with a Go nil dereference.
const nestedOverloadSource = `function measure(create: boolean): number {
    const maps = new Map<string, Map<string, number>>();
    function lookup(key: string, create: true): Map<string, number>;
    function lookup(key: string, create: false): Map<string, number> | undefined;
    function lookup(key: string, create: boolean): Map<string, number> | undefined {
        let result = maps.get(key);
        if (create && result === undefined) {
            result = new Map<string, number>();
            maps.set(key, result);
        }
        return result;
    }
    const made = lookup("abc", true);
    const found = lookup(create ? "abc" : "missing", false);
    return made.size + (found === undefined ? 10 : found.size + 100);
}
console.log(String(measure(true)));
`

func TestNestedOverloadSignaturesBindOnlyTheImplementation(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, nestedOverloadSource)
	if err != nil {
		t.Fatal(err)
	}
	implementations := 0
	for _, function := range program.Functions {
		if function.Name == "lookup" {
			implementations++
		}
	}
	if implementations != 1 {
		t.Fatalf("want one runtime lookup for three declarations, got %d", implementations)
	}
	// The first overload promises a Map the implementation need not return.
	if !slices.ContainsFunc(program.Strings, func(text string) bool {
		return strings.HasPrefix(text, "overload 1 of lookup result: expected Map<string, number>, got undefined")
	}) {
		t.Fatalf("the narrower overload result is not checked: %q", program.Strings)
	}
}

// The meter's own site is generic. Its outer type parameters in a nested body
// are an existing stage 0 gap, which must be named rather than crash.
func TestNestedGenericOverloadSignaturesAreNamed(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function cache<K, V>(): number {
    const maps = new Map<string, Map<K, V>>();
    function lookup(key: string, create: true): Map<K, V>;
    function lookup(key: string, create: false): Map<K, V> | undefined;
    function lookup(key: string, create: boolean): Map<K, V> | undefined {
        let result = maps.get(key);
        if (create && result === undefined) {
            result = new Map<K, V>();
            maps.set(key, result);
        }
        return result;
    }
    return lookup("a", true).size;
}
console.log(String(cache<string, number>()));
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want a named NotYet, got %v", err)
	}
}

// A value would be called through one closure signature, not the resolved overload.
func TestNestedOverloadedValueIsNamed(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function run(): number {
    function pick(value: string): number;
    function pick(value: string): number { return value.length; }
    const saved: (value: string) => number = pick;
    return saved("abc");
}
console.log(String(run()));
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "an overloaded function read as a value") {
		t.Fatalf("want NotYet for an overloaded value, got %v", err)
	}
}
