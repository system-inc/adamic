package lower

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
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
function make(): Held { return { kind: 'w', w: 1 }; }
const held: Held = make();
console.log(held.kind);
`)
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source.String(), extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":74:20" {
				t.Fatalf("want fail-closed depth NotYet at 74:20, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}

// The Weak is hidden only in T here. A direct T witness supplies the mapper;
// main does not infer T from the union alias alone.
func TestUnionKeepingMappedWeakArgument(t *testing.T) {
	const source = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Held<T> = { readonly kind: 'v'; readonly v: T } | { readonly kind: 'w'; readonly w: number };
function copy<T>(x: Held<T>, witness: T): Held<T> {
 const seen: Held<T> = x;
 return seen;
}
const got = copy<Weak<Box>>({ kind: 'v', v: { label: 'new'.repeat(2) } }, undefined);
if (got.kind === 'v' && got.v !== undefined) console.log(got.v.label);
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			program, err := lowerSourceAs(t, source, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || program != nil || filepath.Base(stop.Where) != "main"+extension+":5:24" {
				t.Fatalf("want mapped Weak argument NotYet at 5:24, got program=%v error=%v", program != nil, err)
			}
			t.Log(err)
		})
	}
}
