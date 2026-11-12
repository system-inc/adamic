package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"strings"
	"testing"
)

// Mutant run and restored: accepting array bases only lets q06.ts lower again.
// q06.a retains the older overload refusal once the presence guard is removed.
func TestPhantomObjectPresenceLanguage(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, where string }{
		{"q01_never_object", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nif (!('__tagged' in r)) {\n\tconst lie: { n: number } = r;\n\tconsole.log(String(lie.n + 1));\n}\n", "7:7"},
		{"q02_never_two_fields", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nif (!('__tagged' in r)) {\n\tconst lie: { a: number; b: number; c: number } = r;\n\tconsole.log(String(lie.c));\n}\n", "7:7"},
		{"q03_never_array", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nif (!('__tagged' in r)) {\n\tconst lie: number[] = r;\n\tconsole.log(String(lie.length));\n}\n", "7:7"},
		{"q04_never_argument", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nfunction read(value: { label: string }): string { return value.label; }\nif (!('__tagged' in r)) {\n\tconsole.log(String(read(r)));\n}\n", "8:7"},
		{"q05_never_write", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nconst keep: Node = r;\nif (!('__tagged' in r)) {\n\tconst lie: { n: number; m: number } = r;\n\tlie.m = 7;\n\tconsole.log(String(lie.m));\n\tconsole.log(String(keep.kind));\n}\n", "8:7"},
		{"q06_never_string_method", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nif (!('__tagged' in r)) {\n\tconst lie: { kind: string } = r;\n\tconsole.log(lie.kind.toUpperCase());\n}\n", "7:7"},
		{"q07_predicate_never_object", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction isTagged(node: Node): node is TaggedNode { return node.kind === 1; }\nconst n: Node = { kind: 1 };\nif (isTagged(n)) {\n\tif (!('__tagged' in n)) {\n\t\tconst lie: { n: number } = n;\n\t\tconsole.log(String(lie.n + 1));\n\t}\n}\n", "6:8"},
		{"p02_never_string", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction tag(node: Node, sure: true): TaggedNode;\nfunction tag(node: Node, sure: boolean): Node;\nfunction tag(node: Node, sure: boolean): Node { return node; }\nconst r = tag({ kind: 1 }, true);\nif ('__tagged' in r) {\n\tconsole.log('in');\n} else {\n\tconst lie: string = r;\n\tconsole.log(typeof lie);\n}\n", "7:5"},
		{"p11_predicate_never", "interface Node { readonly kind: number }\ninterface TaggedNode extends Node { __tagged: undefined }\nfunction isTagged(node: Node): node is TaggedNode { return node.kind === 1; }\nconst n: Node = { kind: 1 };\nif (isTagged(n)) {\n\tif ('__tagged' in n) {\n\t\tconsole.log('in');\n\t} else {\n\t\tconst lie: string = n;\n\t\tconsole.log(typeof lie);\n\t}\n}\n", "6:6"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, row.source, extension)
				want := "main" + extension + ":" + row.where + ": Adamic 0.1 refuses an observation of TaggedNode's phantom member __tagged (an `in` test); " + phantomObjectPresenceFix
				if extension == ".a" && (row.name == "p11_predicate_never" || row.name == "q07_predicate_never_object") {
					// .a already refuses these predicates before reaching their observation.
					prior := "main.a:3:53: Adamic 0.1 refuses a type predicate whose return is not proven (true return narrows to Node, not TaggedNode); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)"
					phantomObjectPresenceWant(t, err, prior)
					path := writeAgreementSource(t, "main.a", row.source)
					program, loadErr := load.Load([]string{path})
					if loadErr != nil {
						t.Fatal(loadErr)
					}
					checked, release := program.Checker(context.Background(), program.Files()[0])
					defer release()
					l := lowering{program: program, checker: checked}
					var in *ast.Node
					var visit ast.Visitor
					visit = func(n *ast.Node) bool {
						if n.Kind == ast.KindBinaryExpression && n.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword {
							in = n
						}
						return n.ForEachChild(visit)
					}
					program.Files()[0].AsNode().ForEachChild(visit)
					if in == nil {
						t.Fatal("missing in test")
					}
					err = l.phantomPresenceRefusal(in)
				}
				phantomObjectPresenceWant(t, err, want)
			})
		}
	}
}
func TestPhantomObjectPresenceRealMember(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		lowersAndAgreesWithNodeAs(t, "interface Plain { readonly kind: number; readonly tag: number }\nconst p: Plain = { kind: 1, tag: 2 };\nif (!('tag' in p)) {\n\tconst lie: { n: number } = p;\n\tconsole.log(String(lie.n + 1));\n} else {\n\tconsole.log('present');\n}\n", extension)
	}
}

const phantomObjectPresenceFix = "a phantom brand is never present at run time, so test the value's real fields, or test it as its base type (rulings.md:31)"

func phantomObjectPresenceWant(t *testing.T, err error, want string) {
	t.Helper()
	var refused *Refused
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
}

func TestPhantomObjectPresenceObservations(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, where, road string }{
		{"hasOwn", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.hasOwn(value, '__brand')));\n", "7:20", "Object.hasOwn"},
		{"spread", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconst copy = { ...value };\nconsole.log(String(copy.kind));\n", "7:16", "an object spread"},
		{"supertype", "interface Base { readonly kind: number }\ninterface Branded { readonly kind: number; __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String('__brand' in value));\n", "7:20", "an `in` test"},
		{"generic", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nfunction present<T extends Branded | {other: number}>(x: T): boolean { return '__brand' in x; }\nconsole.log(String(present(value)));\n", "7:79", "an `in` test"},
		{"keys", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.keys(value)));\n", "7:20", "Object.keys"},
		{"values", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.values(value)));\n", "7:20", "Object.values"},
		{"entries", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.entries(value)));\n", "7:20", "Object.entries"},
		{"names", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.getOwnPropertyNames(value)));\n", "7:20", "Object.getOwnPropertyNames"},
		{"assign", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(Object.assign({}, value)));\n", "7:20", "Object.assign"},
		{"json", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(JSON.stringify(value)));\n", "7:20", "JSON.stringify"},
		{"ownMethod", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(value.hasOwnProperty('__brand')));\n", "7:20", "hasOwnProperty"},
		{"enumerable", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String(value.propertyIsEnumerable('__brand')));\n", "7:20", "propertyIsEnumerable"},
		{"write", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nvalue.__brand = undefined;\n", "7:1", "a write"},
		{"elementWrite", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nvalue['__brand'] = undefined;\n", "7:1", "a write"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, row.source, extension)
				want := "main" + extension + ":" + row.where + ": Adamic 0.1 refuses an observation of Branded's phantom member __brand (" + row.road + "); " + phantomObjectPresenceFix
				phantomObjectPresenceWant(t, err, want)
			})
		}
	}
}

func TestPhantomObjectPresenceControls(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"real field", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nconst value: Branded = {kind: 1, __brand: undefined};\nconsole.log(String('kind' in value));\n"},
		{"own method", "interface Base { readonly kind: number; hasOwnProperty(key: string): boolean }\ninterface Branded extends Base { __brand: undefined }\nconst value: Branded = {kind: 1, __brand: undefined, hasOwnProperty: (_key: string): boolean => true};\nconsole.log(String(value.hasOwnProperty('__brand')));\n"},
		{"real class", "class Real { kind = 1; __brand: undefined = undefined; }\nconst value = new Real();\nconsole.log(String('__brand' in value));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, row.source, extension)
			})
		}
	}
}

func TestPhantomObjectPresenceEdges(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, where, brand, member, road string }{
		{"other union arm", "interface Base { readonly kind: number }\ninterface Branded extends Base { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\ninterface Other extends Base { __other: undefined }\nfunction see(x: Branded | Other): boolean { return '__other' in x; }\nconsole.log(String(see(value)));\n", "8:52", "Other", "__other", "an `in` test"},
		{"numeric write", "interface Base {kind: number; [key: number]: undefined}\ninterface Branded extends Base {0: undefined}\nfunction set(value: Branded, key: number): void { value[key] = undefined; }\nset({kind: 1, 0: undefined}, 1);\n", "3:51", "Branded", "0", "a write"},
		{"inherited supertype", "interface Base { readonly kind: number }\ninterface Super { readonly kind: number; __super: undefined }\ninterface Branded extends Super { __brand: undefined }\nfunction claim(x: Base, sure: true): Branded;\nfunction claim(x: Base, sure: boolean): Base;\nfunction claim(x: Base, sure: boolean): Base { return x; }\nconst value = claim({kind: 1}, true);\nconsole.log(String('__super' in value));\n", "8:20", "Branded", "__super", "an `in` test"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, row.source, extension)
				if row.name == "numeric write" {
					if err == nil || !strings.HasSuffix(err.Error(), "main"+extension+":1:31: stage 0 can't lower a non-mutable unrestricted string signature, or numeric, symbol, patterned or callable dictionary storage yet") {
						t.Fatalf("numeric dictionary changed its earlier stop: %v", err)
					}
					err = phantomObjectPresenceRule(t, row.source, extension)
				}
				want := "main" + extension + ":" + row.where + ": Adamic 0.1 refuses an observation of " + row.brand + "'s phantom member " + row.member + " (" + row.road + "); " + phantomObjectPresenceFix
				phantomObjectPresenceWant(t, err, want)
			})
		}
	}
}

// Judge only rule 450 when an existing, earlier stop owns the whole program.
func phantomObjectPresenceRule(t *testing.T, source, extension string) error {
	t.Helper()
	path := writeAgreementSource(t, "main"+extension, source)
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := lowering{program: program, checker: checked}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if err != nil {
			return true
		}
		err = l.phantomPresenceRefusal(n)
		if err != nil {
			return true
		}
		return n.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	return err
}
