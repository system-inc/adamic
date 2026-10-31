package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// An object intersection whose fields hold objects, arrays and boxed unions (Compiler's ruling on
// #qqvq6n7): a spread copy dropped before its original and after it, an object field the copy
// shares and that is written through one and read through the other, a boxed union field and an
// object that may be undefined, a union with undefined and null that keeps the two apart, and
// Object.assign onto a target with an object field, alone and in a program with graph types. Each
// agrees with Node natively, sanitized and on the JavaScript backend, and lets go of everything it
// made.
func TestObjectIntersectionFieldsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/object_intersection_fields"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			objectIntersectionAgrees(t, path)
		})
	}
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}

// An intersection whose field holds the intersection itself, a chain walked and summed. Main's
// objectIntersection recursed without end on it, a goroutine stack overflow, so the admission
// delta's base can't classify it: it is written here rather than kept in the corpus.
func TestObjectIntersectionHoldingItselfAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "cyclic.a")
	source := `type Chain = { readonly value: number } & { readonly next: Chain | undefined };
function build(count: number): Chain | undefined {
    let chain: Chain | undefined = undefined;
    for (let index = 1; index <= count; index++) {
        const link: Chain = { value: index, next: chain };
        chain = link;
    }
    return chain;
}
let total = 0;
let at = build(4);
while (at !== undefined) {
    total += at.value;
    at = at.next;
}
console.log(` + "`${total}`" + `);
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	// Readonly links each made with the one before close no cycle, so no graph region holds them.
	if got := objectIntersectionGraph(t, objectIntersectionAgrees(t, path)); got != "" {
		t.Errorf("graph counts %q, want none", got)
	}
}

// An intersection whose mutable field can hold the intersection again, written so that it does: a
// link written to itself, two written to each other, and a link closed through a plain object it
// holds. The cycle finder sees an intersection's members as the object types they are (cycles.go's
// use), so the write seeds a graph region rather than leaking a cycle reference counting can't
// free. Each is made three times and let go of: a lone member is collected each time, and two are
// one region made by a merge. Each agrees with Node natively, sanitized and on JavaScript, and lets
// go of everything under the leak check.
func TestObjectIntersectionCycleIsHeldByAGraphRegion(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, graph string }{
		{"itself", `type Link = { readonly value: number } & { next: Link | undefined };
function run(seed: number): number {
    const link: Link = { value: seed, next: undefined };
    link.next = link;
    return link.next.value;
}
let total = 0;
for (let index = 1; index <= 3; index++) {
    total += run(index);
}
console.log(` + "`${total}`" + `);
`, "live 1, live 1, live 1"},
		{"each other", `type Link = { readonly value: number } & { next: Link | undefined };
function run(seed: number): number {
    const one: Link = { value: seed, next: undefined };
    const other: Link = { value: seed * 10, next: undefined };
    one.next = other;
    other.next = one;
    return one.next.value + other.next.value;
}
let total = 0;
for (let index = 1; index <= 3; index++) {
    total += run(index);
}
console.log(` + "`${total}`" + `);
`, "live 2, live 2, live 2, regions 3 merges 3"},
		{"through a field", `interface Holder {
    back: Link | undefined;
}
type Link = { readonly value: number } & { readonly holder: Holder };
function run(seed: number): number {
    const holder: Holder = { back: undefined };
    const link: Link = { value: seed, holder: holder };
    holder.back = link;
    return link.holder.back === undefined ? -1 : link.holder.back.value;
}
let total = 0;
for (let index = 1; index <= 3; index++) {
    total += run(index);
}
console.log(` + "`${total}`" + `);
`, "live 2, live 2, live 2, regions 3 merges 3"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "cycle.a")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := objectIntersectionGraph(t, objectIntersectionAgrees(t, path)); got != probe.graph {
				t.Errorf("graph counts %q, want %q", got, probe.graph)
			}
		})
	}
}

var objectIntersectionGraphCounts = regexp.MustCompile(`(?m)^adamic: graph counts: regions (\d+) merges (\d+)\n`)

// objectIntersectionGraph runs a program's counted build and says what its graph regions held: each
// region's live members as it was let go of, then how many regions merges made, or "" with none.
func objectIntersectionGraph(t *testing.T, program *ir.Program) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, arguments := pinnedStack(binary)
	counted := execute(t, name, arguments...)
	if counted.exitCode != 0 || countsLine.FindSubmatch(counted.stderr) == nil {
		t.Fatalf("counts %d %s", counted.exitCode, counted.stderr)
	}
	held := []string{}
	for _, region := range graphRegionLine.FindAllSubmatch(counted.stderr, -1) {
		held = append(held, "live "+string(region[1]))
	}
	if merged := objectIntersectionGraphCounts.FindSubmatch(counted.stderr); merged != nil {
		held = append(held, fmt.Sprintf("regions %s merges %s", merged[1], merged[2]))
	}
	return strings.Join(held, ", ")
}

// objectIntersectionAgrees holds a program to Node natively, sanitized and on JavaScript, and to the
// leak check, and returns it lowered.
func objectIntersectionAgrees(t *testing.T, path string) *ir.Program {
	t.Helper()
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Errorf("leaks:\n%s", report)
	}
	return program
}
