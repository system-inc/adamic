package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/adamic/internal/load"
)

func TestUnionKeepingAssignments(t *testing.T) {
	for _, probe := range []struct {
		name, source, what string
		line, column       int
	}{
		{"p5_before", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } | undefined = make(box);
  const second: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number } = first;
  const seen = second;
  if (seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a { readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; } seen as a { readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; } (one keeps something weakly that the other keeps strongly)`, 7, 110},
		{"p5_after", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } | undefined = make(box);
  const second: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number } = first;
  first = undefined;
  box = undefined;
  const seen = second;
  if (seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a { readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; } seen as a { readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; } (one keeps something weakly that the other keeps strongly)`, 7, 110},
		{"p6_before", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: { readonly inner: { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } } | undefined = { inner: make(box) };
  const second: { readonly inner: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number } } = first;
  const seen = second.inner;
  if (seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a { readonly inner: { readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; }; } seen as a { readonly inner: { readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; }; } (one keeps something weakly that the other keeps strongly)`, 7, 130},
		{"p6_after", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: { readonly inner: { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } } | undefined = { inner: make(box) };
  const second: { readonly inner: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number } } = first;
  first = undefined;
  box = undefined;
  const seen = second.inner;
  if (seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a { readonly inner: { readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; }; } seen as a { readonly inner: { readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; }; } (one keeps something weakly that the other keeps strongly)`, 7, 130},
		{"p7", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function first(x: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number }): string {
  if (x.kind === 'v' && x.v !== undefined) return x.v.label;
  return 'number';
}
const second: (x: { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number }) => string = first;
console.log(second({ kind: 'v', v: { label: 'new'.repeat(2) } }));
`, `a (x: { readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; }) => string seen as a (x: { readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; }) => string (one keeps something weakly that the other keeps strongly)`, 7, 117},
		{"p8_before", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: readonly ({ readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number })[] | undefined = [make(box)];
  const second: readonly ({ readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number })[] = first;
  const seen = second[0];
  if (seen !== undefined && seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a readonly ({ readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; })[] seen as a readonly ({ readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; })[] (one keeps something weakly that the other keeps strongly)`, 7, 123},
		{"p8_after", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function make(box: Box): { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number } { return { kind: 'v', v: box }; }
function run(): void {
  let box: Box | undefined = { label: 'new'.repeat(2) };
  let first: readonly ({ readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number })[] | undefined = [make(box)];
  const second: readonly ({ readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number })[] = first;
  first = undefined;
  box = undefined;
  const seen = second[0];
  if (seen !== undefined && seen.kind === 'v' && seen.v !== undefined) console.log(seen.v.label);
}
run();
`, `a readonly ({ readonly kind: "v"; readonly v: Box; } | { readonly kind: "w"; readonly w: number; })[] seen as a readonly ({ readonly kind: "v"; readonly v: Weak<Box>; } | { readonly kind: "w"; readonly w: number; })[] (one keeps something weakly that the other keeps strongly)`, 7, 123},
		{"p9", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
const first: { readonly v: Box } = { v: { label: 'new'.repeat(2) } };
const second: { readonly v: Weak<Box> } = first;
if (second.v !== undefined) console.log(second.v.label);
`, `a { readonly v: Box; } seen as a { readonly v: Weak<Box>; } (one keeps something weakly that the other keeps strongly)`, 4, 43},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				program, err := lowerSourceAs(t, probe.source, extension)
				var stop *NotYet
				if !errors.As(err, &stop) || program != nil {
					t.Fatalf("want assignment NotYet, got program=%v error=%v", program != nil, err)
				}
				want := fmt.Sprintf("main%s:%d:%d", extension, probe.line, probe.column)
				if filepath.Base(stop.Where) != want {
					t.Fatalf("want assignment at %s, got %v", want, err)
				}
				if stop.What != probe.what {
					t.Fatalf("want %q, got %v", probe.what, err)
				}
				t.Log(err)
			})
		}
	}
}

func TestUnionKeepingControls(t *testing.T) {
	for _, extension := range []string{".a", ".ts"} {
		t.Run("same"+extension, func(t *testing.T) {
			lowersAndAgreesWithNodeAs(t, `interface Box { readonly label: string }
type Held = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
function make(box: Box): Held { return { kind: 'v', v: box }; }
const first: Held = make({ label: 'new'.repeat(2) });
const second: Held = first;
if (second.kind === 'v') console.log(second.v.label);
`, extension)
		})
		t.Run("wider"+extension, func(t *testing.T) {
			lowersAndAgreesWithNodeAs(t, `type Narrow = { readonly kind: 'v'; readonly n: number } | { readonly kind: 'w'; readonly n: number };
type Wide = { readonly kind: 'v' | 'x'; readonly n: number } | { readonly kind: 'w'; readonly n: number };
function make(): Narrow { return { kind: 'v', n: 1 }; }
const first: Narrow = make();
const second: Wide = first;
console.log(String(second.n));
`, extension)
		})
	}
}

func TestUnionKeepingNarrowingControls(t *testing.T) {
	for _, extension := range []string{".a", ".ts"} {
		t.Run("optional"+extension, func(t *testing.T) {
			lowersAndAgreesWithNodeAs(t, `const boxes: readonly ({ readonly name: string } | undefined)[] = [{ name: 'ada' }, undefined];
for (const box of boxes) {
 console.log(box?.name ?? 'none');
}
`, extension)
		})
	}
	// .ts permits the existing runtime-checked downcast; .a has its own cast policy.
	t.Run("checked_downcast.ts", func(t *testing.T) {
		lowersAndAgreesWithNodeAs(t, `type Shape = { readonly kind: 'Circle'; readonly radius: number } | { readonly kind: 'Square'; readonly side: number };
const shapes: readonly Shape[] = [{ kind: 'Circle', radius: 2 }];
for (const shape of shapes) {
 const circle = shape as Extract<Shape, { kind: 'Circle' }>;
 console.log(String(circle.radius));
}
`, ".ts")
	})
}

func TestUnionKeepingNestedWeak(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: { readonly inner: Box } } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: { readonly inner: Weak<Box> } } | { readonly kind: 'w'; readonly w: number };
function make(box: Box): Strong { return { kind: 'v', v: { inner: box } }; }
const first: Strong = make({ label: 'new'.repeat(2) });
const second: Weakly = first;
if (second.kind === 'v' && second.v.inner !== undefined) console.log(second.v.inner.label);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":7:24" || stop.What != "a Strong seen as a Weakly (one keeps something weakly that the other keeps strongly)" {
				t.Fatalf("want nested Weak assignment NotYet at 7:24, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}

func TestUnionKeepingGenericWeakInstance(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
function view<T extends object>(x: { readonly kind: 'v'; readonly v: T } | { readonly kind: 'w'; readonly w: number }): { readonly kind: 'v'; readonly v: Weak<T> } | { readonly kind: 'w'; readonly w: number } {
 return x;
}
const got = view<Box>({ kind: 'v', v: { label: 'new'.repeat(2) } });
if (got.kind === 'v' && got.v !== undefined) console.log(got.v.label);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":4:9" {
				t.Fatalf("want Box instance return NotYet at 4:9, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}

func TestUnionKeepingDepthLimit(t *testing.T) {
	var source strings.Builder
	source.WriteString("type Level0 = { readonly n: number };\n")
	for depth := 1; depth <= 70; depth++ {
		fmt.Fprintf(&source, "type Level%d = { readonly next: Level%d };\n", depth, depth-1)
	}
	source.WriteString(`type Held = { readonly kind: 'v'; readonly v: Level70 } | { readonly kind: 'w'; readonly w: number };
type Twin = { readonly kind: 'v'; readonly v: Level70 } | { readonly kind: 'w'; readonly w: number };
function make(): Held { return { kind: 'w', w: 1 }; }
const held: Held = make();
const twin: Twin = held;
console.log(twin.kind);
`)
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source.String(), extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":76:20" {
				t.Fatalf("want fail-closed depth NotYet at 76:20, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}

// An identical mapped Weak union is an identity view, so it stays admitted.
func TestUnionKeepingMappedWeakArgument(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Held<T> = { readonly kind: 'v'; readonly v: T } | { readonly kind: 'w'; readonly w: number };
function copy<T>(x: Held<T>, witness: T): Held<T> { const seen: Held<T> = x; return seen; }
const box: Box = { label: 'new'.repeat(2) };
const got = copy<Weak<Box>>({ kind: 'v', v: box }, box);
if (got.kind === 'v' && got.v !== undefined) console.log(got.v.label);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) { lowersAndAgreesWithNodeAs(t, source, extension) })
	}
}

func TestUnionKeepingRegExpLibrary(t *testing.T) {
	const source = `const p: string | RegExp = "[[/]/";
const regex = new RegExp(p, 'u');
console.log(regex.source + '|' + regex.flags + '|' + String(regex.test('//')));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) { lowersAndAgreesWithNodeAs(t, source, extension) })
	}
}

func TestUnionKeepingRecursiveNoWeak(t *testing.T) {
	const source = `type N = { readonly parent?: N; readonly kind: 'a' } | { readonly parent?: N; readonly kind: 'b' };
type Twin = { readonly parent?: Twin; readonly kind: 'a' } | { readonly parent?: Twin; readonly kind: 'b' };
function make(): N { return { kind: 'a' }; }
const first: N = make();
const same: N = first;
const twin: Twin = same;
console.log(twin.kind);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) { lowersAndAgreesWithNodeAs(t, source, extension) })
	}
}

func TestUnionKeepingLibraryWeakArgument(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Weakly = { readonly kind: 'v'; readonly v: readonly Weak<Box>[] } | { readonly kind: 'w'; readonly w: number };
type Strong = { readonly kind: 'v'; readonly v: readonly (Box | undefined)[] } | { readonly kind: 'w'; readonly w: number };
function make(box: Box): Weakly { return { kind: 'v', v: [box] }; }
const first: Weakly = make({ label: 'new'.repeat(2) });
const second: Strong = first;
console.log(second.kind);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":7:24" {
				t.Fatalf("want library Weak argument NotYet at 7:24, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}

func TestUnionKeepingPreludeDoesNotHideWeak(t *testing.T) {
	source, err := os.ReadFile("../load/prelude.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{}, string(source), core.ScriptKindTS)
	weakAliases := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeAliasDeclaration && node.Name().Text() == "Weak" {
			weakAliases++
			return false
		}
		if ast.IsIdentifier(node) && (node.Text() == "Weak" || node.Text() == "WeakBrand") && !ast.IsDeclarationName(node) && ast.IsPartOfTypeNode(node) {
			t.Errorf("prelude type outside Weak names %s", node.Text())
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if weakAliases != 1 {
		t.Fatalf("want exactly one allowed Weak alias, got %d", weakAliases)
	}
}

// The loader checks roots as modules, so this Array declaration stays local.
// No accepted program merge is available to mutate; keep the all-declarations
// production condition even though this frontend boundary prevents the merge.
func TestUnionKeepingLocalArrayDoesNotMerge(t *testing.T) {
	const source = `interface Box { readonly label: string }
type Weak<T extends object> = import('adamic').Weak<T>;
interface Array<T> { held?: Weak<Box> }
const values: number[] = [];
console.log(String(values.length));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := program.Checker(context.Background(), program.Files()[0])
			defer release()
			var array, local *checker.Type
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindArrayType {
					array = checked.GetTypeAtLocation(node)
				}
				if node.Kind == ast.KindInterfaceDeclaration && node.Name().Text() == "Array" {
					local = checked.GetTypeAtLocation(node.Name())
				}
				return node.ForEachChild(visit)
			}
			program.Files()[0].AsNode().ForEachChild(visit)
			if array == nil || local == nil {
				t.Fatal("missing local or library Array type")
			}
			if checked.GetPropertyOfType(array, "held") != nil {
				t.Fatal("module-local Array merged held into library Array")
			}
			if checked.GetPropertyOfType(local, "held") == nil {
				t.Fatal("local Array declaration lost held")
			}
		})
	}
}

// Library mapped/anonymous objects can retain a program's Weak through their
// instantiation mapper even when they expose no alias or reference arguments.
func TestUnionKeepingInstantiatedLibraryObjects(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
const x: { readonly v: Weak<Box> } = { v: undefined };
const descriptors = Object.getOwnPropertyDescriptors(x);
const values: Weak<Box>[] = [];
const push = values.push;
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := program.Checker(context.Background(), program.Files()[0])
			defer release()
			types := map[string]*checker.Type{}
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindVariableDeclaration && node.Name() != nil && ast.IsIdentifier(node.Name()) {
					types[node.Name().Text()] = checked.GetTypeAtLocation(node.Name())
				}
				return node.ForEachChild(visit)
			}
			program.Files()[0].AsNode().ForEachChild(visit)
			analysis := &lowering{program: program, checker: checked}
			descriptors, push := types["descriptors"], types["push"]
			if descriptors == nil || push == nil {
				t.Fatal("missing instantiated library types")
			}
			mapped := false
			for _, member := range descriptors.Types() {
				mapped = mapped || member.Flags()&checker.TypeFlagsObject != 0 && member.ObjectFlags()&checker.ObjectFlagsMapped != 0
			}
			if !mapped {
				t.Fatalf("descriptors must contain an instantiated mapped type: %s", checked.TypeToString(descriptors))
			}
			if !analysis.mentionsWeak(descriptors, map[*checker.Type]bool{}) {
				t.Fatal("instantiated library mapped type hid Weak")
			}
			if push.Flags()&checker.TypeFlagsObject == 0 || push.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 || len(push.Alias().TypeArguments()) != 0 || len(analysis.typeArguments(push)) != 0 {
				t.Fatalf("push must be anonymous with no exposed type arguments: %s", checked.TypeToString(push))
			}
			symbol := push.Symbol()
			if symbol == nil || len(symbol.Declarations) == 0 {
				t.Fatal("anonymous push type must retain its library symbol")
			}
			for _, declaration := range symbol.Declarations {
				if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
					t.Fatal("push declaration is not from the library")
				}
			}
			if !analysis.mentionsWeak(push, map[*checker.Type]bool{}) {
				t.Fatal("instantiated library anonymous type hid Weak")
			}
		})
	}
}

func TestUnionKeepingLibraryAliasAndTupleControls(t *testing.T) {
	for _, probe := range []struct{ name, source, where, what string }{
		{"record", `class Box { v = 1; }
let x: Box | Record<string, number> = new Box();
function swap(): void { const d: Record<string, number> = {}; d["v"] = 5; x = d; }
if (x instanceof Box) { swap(); console.log(String(x.v)); }
`, ":4:5", "object tag cannot be checked with typeof"},
		{"tuple", `function pick(flag: boolean): [string, number] | [number, string] {
 return flag ? ["a", 1] : [1, "a"];
}
const [a, b]: [string, number] | [number, string] = pick(true);
console.log(String(a) + String(b));
`, ":4:7", "destructuring anything but a tuple into [names]"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				program, err := lowerSourceAs(t, probe.source, extension)
				var stop *NotYet
				if !errors.As(err, &stop) || program != nil || !strings.HasSuffix(stop.Where, probe.where) || !strings.Contains(stop.What, probe.what) {
					t.Fatalf("want NotYet %s at %s, got %v", probe.what, probe.where, err)
				}
				t.Log(err)
			})
		}
	}
}

func TestUnionKeepingLibraryAliasAndTupleWeakArguments(t *testing.T) {
	for _, probe := range []struct{ name, weak, strong string }{
		{"record", "Record<string, Weak<Box>> | { readonly w: number }", "Record<string, Box | undefined> | { readonly w: number }"},
		{"tuple", "[Weak<Box>, number] | [number, string]", "[Box | undefined, number] | [number, string]"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				source := `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Weakly = ` + probe.weak + `;
type Strong = ` + probe.strong + `;
function view(first: Weakly): void {
 const second: Strong = first;
 console.log("done");
}
`
				program, err := lowerSourceAs(t, source, extension)
				var stop *NotYet
				if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":6:25" || !strings.Contains(stop.What, "one keeps something weakly that the other keeps strongly") {
					t.Fatalf("want Weak argument assignment NotYet at 6:25, got %v", err)
				}
				t.Log(err)
			})
		}
	}
}
