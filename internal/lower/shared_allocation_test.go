package lower

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// sharedAllocationSource declares one value of each type mayShareAllocation's pins read. Far is the plain type the
// fail-closed rows share no names with: each of them holds a required name Far lacks and lacks Far's zone, so the
// name test would cut it from Far were it plain.
const sharedAllocationSource = `interface Plain { kind: number; name: string }
interface Sub extends Plain { extra: number }
interface Other { size: number; weight: number }
interface Far { zone: boolean }
interface Optional { note?: string; extra?: number }
interface AlsoOptional { label?: string }
interface Printable { toString(): string }
interface Indexed { [key: string]: number; count: number }
interface Callable { (): number; size: number }
interface Constructable { new (): Plain; size: number }
interface Box<T> { inner: T }
type Mapped = { readonly [K in keyof Other]: Other[K] };
declare const plain: Plain;
declare const sub: Sub;
declare const other: Other;
declare const far: Far;
declare const optional: Optional;
declare const alsoOptional: AlsoOptional;
declare const printable: Printable;
declare const either: Plain | Other;
declare const both: Plain & Other;
declare const indexed: Indexed;
declare const callable: Callable;
declare const constructable: Constructable;
declare const box: Box<number>;
declare const mapped: Mapped;
declare const map: Map<string, number>;
declare const error: Error;
`

// The finder's own relation for selection fails closed on every type it can't read plainly, and cuts only a pair
// the checker never relates (#nxb3xh1 head 6, Compiler's condition 1): each fail-closed row is related to Far, a
// plain type it shares no names with, beside a plain pair it cuts and plain pairs it relates by their names. Every
// row it cuts is checked against the checker both ways, which must refuse it too, and the Printable row is one the
// checker relates by the global Object type's toString, which no listed property shows. The relation itself never
// asks the checker's relation; the test does, to hold it to that.
func TestMayShareAllocationFailsClosed(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, sharedAllocationSource)
	for _, row := range []struct {
		name, one, other string
		want             bool
	}{
		{"a union", "either", "far", true},
		{"an intersection", "both", "far", true},
		{"an index signature", "indexed", "far", true},
		{"a callable", "callable", "far", true},
		{"a constructable", "constructable", "far", true},
		{"a generic reference", "box", "far", true},
		{"a mapped type", "mapped", "far", true},
		{"a lib type, a Map", "map", "far", true},
		{"a lib type, an Error", "error", "far", true},
		{"a plain pair, each holding a required name the other lacks", "plain", "other", false},
		{"a plain pair, the far one", "plain", "far", false},
		{"a plain pair with only optional names", "optional", "alsoOptional", true},
		{"a plain type with only optional names beside one with required names", "optional", "plain", true},
		{"a plain pair, one holding every required name of the other", "sub", "plain", true},
		{"a required name every object holds, Object's toString", "printable", "far", true},
		{"the same type", "far", "far", true},
	} {
		t.Run(row.name, func(t *testing.T) {
			one, other := types[row.one], types[row.other]
			if one == nil || other == nil {
				t.Fatalf("the source declares no %s or no %s", row.one, row.other)
			}
			if got := f.mayShareAllocation(one, other); got != row.want {
				t.Fatalf("mayShareAllocation(%s, %s) = %t, want %t", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other), got, row.want)
			}
			if got := f.mayShareAllocation(other, one); got != row.want {
				t.Fatalf("mayShareAllocation(%s, %s) = %t, want %t, both ways", f.l.checker.TypeToString(other), f.l.checker.TypeToString(one), got, row.want)
			}
			if !row.want && (f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)) {
				t.Fatalf("mayShareAllocation cut %s from %s, which the checker relates", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other))
			}
		})
	}
	if !f.l.checker.IsTypeAssignableTo(types["far"], types["printable"]) {
		t.Fatal("the checker doesn't relate Far to Printable, so the toString row pins nothing")
	}
	// The one-way form graphLinks' shapes road asks: a Sub may be seen as a Plain, and a Plain never as a Sub,
	// which requires extra.
	if !f.mayBeSeenAs(types["sub"], types["plain"]) || f.mayBeSeenAs(types["plain"], types["sub"]) {
		t.Fatalf("mayBeSeenAs: Sub as Plain %t, Plain as Sub %t, want true and false", f.mayBeSeenAs(types["sub"], types["plain"]), f.mayBeSeenAs(types["plain"], types["sub"]))
	}
}

// The index answers what the relation does: for every node, the sweep's views are exactly the candidates
// mayShareAllocation relates it to, and the shapes road's shapes exactly those that mayBeSeenAs it, never a scan
// (Compiler's condition 2). Every declared value's type is a candidate and a shape here, the fail-closed ones
// among them, so each arm of the index meets each kind of group.
func TestMayShareAllocationIndexMatchesTheRelation(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, sharedAllocationSource)
	names := []string{}
	for name := range types {
		names = append(names, name)
	}
	slices.Sort(names)
	candidates := []cycleNode{}
	for _, name := range names {
		candidates = append(candidates, cycleNode{proven: types[name]})
		if types[name].Flags()&checker.TypeFlagsObject != 0 {
			f.shapes = append(f.shapes, types[name])
		}
	}
	f.allocation = f.allocationIndexOf(candidates)
	for _, name := range names {
		proven := types[name]
		if !f.allocationView(proven) {
			continue
		}
		want := []cycleNode{}
		for _, view := range candidates {
			if f.allocationView(view.proven) && f.mayShareAllocation(proven, view.proven) {
				want = append(want, view)
			}
		}
		if got := f.allocationViews(proven); !sameAllocationNodes(got, want) {
			t.Errorf("views of %s: index %d, relation %d", f.l.checker.TypeToString(proven), len(got), len(want))
		}
		wantShapes := []*checker.Type{}
		for _, shape := range f.shapes {
			if shape != proven && f.mayBeSeenAs(shape, proven) {
				wantShapes = append(wantShapes, shape)
			}
		}
		if got := f.allocationShapesSeenAs(proven); !sameAllocationTypes(got, wantShapes) {
			t.Errorf("shapes seen as %s: index %d, relation %d", f.l.checker.TypeToString(proven), len(got), len(wantShapes))
		}
	}
	if f.allocationTests == 0 {
		t.Fatal("the index made no name test, so no plain node was asked")
	}
}

// sameAllocationNodes is whether two lists hold the same nodes, each as often.
func sameAllocationNodes(one, other []cycleNode) bool {
	counts := map[cycleNode]int{}
	for _, node := range one {
		counts[node]++
	}
	for _, node := range other {
		counts[node]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

// sameAllocationTypes is whether two lists hold the same types, each as often.
func sameAllocationTypes(one, other []*checker.Type) bool {
	nodes := func(types []*checker.Type) []cycleNode {
		each := []cycleNode{}
		for _, proven := range types {
			each = append(each, cycleNode{proven: proven})
		}
		return each
	}
	return sameAllocationNodes(nodes(one), nodes(other))
}
