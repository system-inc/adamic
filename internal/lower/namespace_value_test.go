package lower

import (
	"errors"
	"strings"
	"testing"
)

// namespaceValueSource is a namespace exporting functions alone, the form whose name read as a value
// has an object (#67jk7wq), followed by each row's use of it.
const namespaceValueSource = "namespace N {\n\texport function f(): number { return 1; }\n}\n"

// A namespace read as a value stops at compile time wherever the narrow object can't be what Node's is,
// in its own words. Node's answer stands beside each row, and the row holds Node to it, so each stop
// is a program Node runs and Adamic doesn't lower yet, never one the checker refused. The merged row
// reads between the two blocks, where Node's object has no g yet, so an object built whole from both
// blocks would disagree with it.
func TestNamespaceValueStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source, node, reason string
		refused                    bool
	}{
		{"merged blocks", "namespace N {\n\texport function f(): number { return 1; }\n}\nconst early: typeof N = N;\nconsole.log(typeof early.g);\nnamespace N {\n\texport function g(): number { return 2; }\n}\nconsole.log(typeof early.g);\n", "undefined\nfunction\n", "a namespace merged across blocks used as a value", false},
		{"exported let", "namespace N {\n\texport let count = 0;\n\texport function f(): number { return count; }\n}\nconst held = N;\nconsole.log(String(held.f()));\n", "0\n", "an exported variable in a namespace used as a value", false},
		{"Object.keys", namespaceValueSource + "console.log(Object.keys(N).join(','));\n", "f\n", "a namespace value seen as an argument to Object.keys", false},
		{"spread", namespaceValueSource + "const copy = { ...N };\nconsole.log(typeof copy.f);\n", "function\n", "a namespace value seen as a spread source", false},
		{"computed name", namespaceValueSource + "const key: 'f' = 'f';\nconsole.log(String(N[key]()));\n", "1\n", "a namespace value seen as a receiver read by a computed name", false},
		{"structural view", namespaceValueSource + "const view: { f(): number } = N;\nconsole.log(String(view.f()));\n", "1\n", "a namespace value seen as {", false},
		{"unknown", namespaceValueSource + "const opaque: unknown = N;\nconsole.log(typeof opaque);\n", "object\n", "a namespace value seen as unknown", false},
		{"generic argument", namespaceValueSource + "function same<T>(value: T): T { return value; }\nconsole.log(typeof same(N));\n", "object\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"logical value", namespaceValueSource + "function pick(flag: boolean): void {\n\tconst either = flag || N;\n\tconsole.log(typeof either);\n}\npick(false);\n", "object\n", "a namespace value seen as true", false},
		{"generic class parameter", namespaceValueSource + "class Holder<T extends { f: () => number }> {\n\tset(x: T): void { x.f = (): number => 2; }\n}\nnew Holder<typeof N>().set(N);\nconsole.log(String(N.f()));\n", "2\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"generic interface parameter", namespaceValueSource + "interface Sink<T> { put(x: T): void; }\nclass Swap<T extends { f: () => number }> implements Sink<T> {\n\tput(x: T): void { x.f = (): number => 2; }\n}\nconst sink: Sink<typeof N> = new Swap<typeof N>();\nsink.put(N);\nconsole.log(String(N.f()));\n", "2\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"overload", namespaceValueSource + "function g(x: typeof N): void;\nfunction g(x: { f: () => number }): void { x.f = (): number => 2; }\ng(N);\nconsole.log(String(N.f()));\n", "2\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"function value", namespaceValueSource + "const set: (x: typeof N) => void = (x: { f: () => number }): void => { x.f = (): number => 2; };\nset(N);\nconsole.log(String(N.f()));\n", "2\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"wider override", namespaceValueSource + "class Base {\n\tset(x: typeof N): void { console.log(String(x.f())); }\n}\nclass Wide extends Base {\n\toverride set(x: { f: () => number }): void { x.f = (): number => 2; }\n}\nconst s: Base = new Wide();\ns.set(N);\nconsole.log(String(N.f()));\n", "2\n", "a namespace value seen as an argument to a callee whose body can't be named", false},
		{"member write", namespaceValueSource + "const held = N;\nheld.f = (): number => 2;\nconsole.log(String(N.f()));\n", "2\n", "a write to a namespace member through the namespace's value", true},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, "main.a", row.source)
			if node := runAgreementNode(t, path); node.code != 0 || string(node.stdout) != row.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q; the row says %q", node.code, node.stdout, node.stderr, row.node)
			}
			_, err := lowerSource(t, row.source)
			if row.refused {
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), row.reason) {
					t.Fatalf("got %v, want Refused %q", err, row.reason)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), row.reason) {
				t.Fatalf("got %v, want NotYet %q", err, row.reason)
			}
		})
	}
}

// The name read as a value is one object wherever it's read, and its fields are the functions' own
// values, so it agrees with Node on ===.
func TestNamespaceValueIdentityAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, namespaceValueSource+"const a = N;\nconst b = N;\nfunction again(): typeof N { return N; }\nconsole.log(`${a === b} ${a === again()} ${a.f === N.f} ${b.f() + N.f()}`);\n")
}

// A namespace value passed into a parameter declared as its own type, or as its own type or undefined,
// is the same object inside the callee, read and called through ?. there.
func TestNamespaceValueParameterAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, namespaceValueSource+"function call(x: typeof N | undefined): number { return x?.f() ?? 0; }\nfunction same(x: typeof N): boolean { return x === N; }\nconsole.log(`${call(N)} ${call(undefined)} ${same(N)}`);\n")
}
