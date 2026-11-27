package lower

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// sharedAllocationSource declares one value of each type mayShareAllocation's pins read. Far is the plain type most
// rows are held against: it shares no name with any of them, so each holds a required name Far lacks and lacks Far's
// zone.
const sharedAllocationSource = `interface Plain { kind: number; name: string }
interface Sub extends Plain { extra: number }
interface Other { size: number; weight: number }
interface Far { zone: boolean }
interface Lengthy { length: number }
interface NodeLike { kind: number; parent: NodeLike | undefined }
interface Optional { note?: string; extra?: number }
interface AlsoOptional { label?: string }
interface Printable { toString(): string }
interface Indexed { [key: string]: number; count: number }
interface Callable { (): number; size: number }
interface Constructable { new (): Plain; size: number }
interface Box<T> { inner: T }
type Mapped = { readonly [K in keyof Other]: Other[K] };
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
function wrap<T>(value: T): T {
	let inner!: { readonly [K in keyof T]: T[K] };
	console.log(typeof inner);
	return value;
}
declare const plain: Plain;
declare const sub: Sub;
declare const other: Other;
declare const far: Far;
declare const lengthy: Lengthy;
declare const nodeLike: NodeLike;
declare const nodes: NodeLike[];
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
declare const mutable: Mutable<Plain>;
declare const partial: Partial<Plain>;
declare const picked: Pick<Plain, "kind">;
declare const frozen: Readonly<Plain>;
declare const map: Map<string, number>;
declare const error: Error;
declare const text: string;
declare const unit: "unit";
`

// sharedAllocationChecked is the checker's own answer for a pair, as deriveRelated reads one (each by its
// relatedShape, assignable either way), asked of IsTypeAssignableTo so a function's type is read too.
func sharedAllocationChecked(f *cycleFinder, one, other *checker.Type) bool {
	one, other = f.l.relatedShape(one), f.l.relatedShape(other)
	return f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)
}

// The finder's own relation for selection fails closed on every type it can't read by its names, and cuts only a
// pair the checker never relates (#nxb3xh1 head 6, Compiler's condition 1 and her ruling at 14:16). Each row holds
// mayShareAllocation's answer, both ways, beside the checker's own: a row it cuts is one the checker refuses both
// ways and deriveRelated refuses, and a row it relates names whether the checker does too, so the pins read the
// checker, not our reading of it. Primitives and literals are related, failing closed, never judged by an empty
// property list; unions and intersections are related, never judged as one object; Printable, whose one required
// name is the global Object type's toString, is related to every type here, as the checker relates them. The
// relation never asks the checker's relation; the test does, to hold it to the checker.
func TestMayShareAllocationFailsClosed(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, sharedAllocationSource)
	for _, row := range []struct {
		name, one, other string
		want, checked    bool
	}{
		{"a union", "either", "far", true, false},
		{"an intersection", "both", "far", true, false},
		{"a callable", "callable", "far", true, false},
		{"a constructable", "constructable", "far", true, false},
		{"a generic mapped type, inside a generic function", "inner", "far", true, false},
		{"a mapped type that isn't generic, cut", "mapped", "far", false, false},
		{"Mutable<Plain>, cut", "mutable", "far", false, false},
		{"Partial<Plain>, all optional, related", "partial", "far", true, false},
		{"Partial<Plain> beside Plain", "partial", "plain", true, true},
		{"Pick<Plain, kind>, cut where kind is missing", "picked", "other", false, false},
		{"Readonly<Plain> beside Plain", "frozen", "plain", true, true},
		{"a primitive", "text", "lengthy", true, true},
		{"a primitive against a type it shares no name with", "text", "far", true, false},
		{"a literal", "unit", "lengthy", true, true},
		{"an index signature, cut", "indexed", "far", false, false},
		{"a concrete reference, Node[] against a plain Node-like, cut", "nodes", "nodeLike", false, false},
		{"a generic reference instantiated, cut", "box", "far", false, false},
		{"a lib type, an Error, cut", "error", "far", false, false},
		{"a lib type, a Map, cut", "map", "far", false, false},
		{"a plain pair, each holding a required name the other lacks, cut", "plain", "other", false, false},
		{"a plain pair, the far one, cut", "plain", "far", false, false},
		{"a plain pair with only optional names", "optional", "alsoOptional", true, false},
		{"a plain type with only optional names beside one with required names", "optional", "plain", true, false},
		{"a plain pair, one holding every required name of the other", "sub", "plain", true, true},
		{"the same type", "far", "far", true, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			one, other := types[row.one], types[row.other]
			if one == nil || other == nil {
				t.Fatalf("the source declares no %s or no %s", row.one, row.other)
			}
			if got, back := f.mayShareAllocation(one, other), f.mayShareAllocation(other, one); got != row.want || back != row.want {
				t.Fatalf("mayShareAllocation(%s, %s) = %t, and the other way %t, want %t", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other), got, back, row.want)
			}
			if checked := sharedAllocationChecked(f, one, other); checked != row.checked {
				t.Fatalf("the checker relates %s and %s: %t, the row says %t", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other), checked, row.checked)
			}
			if !row.want && f.deriveRelated(one, other) {
				t.Fatalf("mayShareAllocation cut %s from %s, which deriveRelated relates", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other))
			}
		})
	}
	t.Run("Object's toString, related to everything", func(t *testing.T) {
		printable := types["printable"]
		for name, proven := range types {
			// A generic mapped type is related to everything already, and the checker reads it through its
			// constraint, so it holds nothing here.
			if allocationGeneric(f.l.checker, proven) {
				continue
			}
			if !f.mayShareAllocation(printable, proven) || !f.mayBeSeenAs(proven, printable) {
				t.Errorf("Printable isn't related to %s (%s)", name, f.l.checker.TypeToString(proven))
			}
			if !f.l.checker.IsTypeAssignableTo(f.l.relatedShape(proven), printable) {
				t.Errorf("the checker doesn't take %s (%s) as a Printable, so the row pins nothing there", name, f.l.checker.TypeToString(proven))
			}
		}
	})
	// The one-way form graphLinks' shapes road asks: a Sub may be seen as a Plain, and a Plain never as a Sub, which
	// requires extra, as the checker has it.
	t.Run("one way", func(t *testing.T) {
		if !f.mayBeSeenAs(types["sub"], types["plain"]) || f.mayBeSeenAs(types["plain"], types["sub"]) {
			t.Fatalf("mayBeSeenAs: Sub as Plain %t, Plain as Sub %t, want true and false", f.mayBeSeenAs(types["sub"], types["plain"]), f.mayBeSeenAs(types["plain"], types["sub"]))
		}
		if f.l.checker.IsTypeAssignableTo(types["plain"], types["sub"]) {
			t.Fatal("the checker takes a Plain as a Sub, so the cut pins nothing")
		}
	})
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

// sharedAllocationPoint is a program whose class requires toString, which every object holds by the global Object
// type, so a Located is a Point to the checker, and a Far, which shares no name with either, is cut from both: the
// property test's row for Object's members.
const sharedAllocationPoint = "class Point {\n\tconstructor(readonly x: number) {}\n\ttoString(): string {\n\t\treturn `${this.x}`;\n\t}\n}\ninterface Located {\n\treadonly x: number;\n}\ninterface Far {\n\treadonly zone: boolean;\n}\nconst located: Located = { x: 2 };\nconst far: Far = { zone: true };\nconst point = new Point(1);\nconsole.log(`${point.toString()} ${located.x} ${far.zone}`);\n"

// Over real programs, every pair the sweep cuts and every shape the shapes road cuts is one the checker refuses
// (Lower's ask on #nxb3xh1 head 6): each program is lowered, its candidates gathered as graphTypes gathers them,
// and every candidate pair the index doesn't relate is asked of deriveRelated, every shape it doesn't take as a node
// of IsTypeAssignableTo. A cut the checker relates fails, naming both types. The programs are graph-region programs
// from the oracle and Point, whose toString only the global Object type gives a Located, beside the pins' own
// declared types, every group among them: index signatures, references, library and mapped types.
func TestMayShareAllocationCutsOnlyWhatTheCheckerRefuses(t *testing.T) {
	t.Parallel()
	paths := []string{}
	for _, fixture := range []string{"graph_regions_list.a", "graph_regions_cache.a", "graph_regions_structural_literal.a", "graph_regions_static.a", "graph_regions_entries.a", "graph_regions_literals.a", "graph_regions_coverage_maps.a"} {
		path, err := filepath.Abs(filepath.Join("../oracle/testdata", fixture))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	paths = append(paths, writeAgreementSource(t, "point.a", sharedAllocationPoint))
	cut := 0
	// The pins' declared types, mapped ones among them, every pair of them asked as the sweep and the shapes road
	// ask a pair.
	t.Run("the declared types", func(t *testing.T) {
		f, types := cycleIndexTypes(t, sharedAllocationSource)
		for _, one := range types {
			for _, other := range types {
				if !f.allocationView(one) || !f.allocationView(other) {
					continue
				}
				if !f.mayShareAllocation(one, other) {
					cut++
					if f.deriveRelated(one, other) {
						t.Errorf("mayShareAllocation cut %s from %s, which deriveRelated relates", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other))
					}
				}
				if one != other && !f.mayBeSeenAs(one, other) {
					cut++
					if f.l.checker.IsTypeAssignableTo(f.l.relatedShape(one), other) {
						t.Errorf("mayBeSeenAs cut %s from %s, which the checker takes as one", f.l.checker.TypeToString(one), f.l.checker.TypeToString(other))
					}
				}
			}
		}
	})
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			entry := loaded.Files()[0]
			check, release := loaded.Checker(context.Background(), entry)
			t.Cleanup(release)
			lowered, err := lowerChecked(loaded, check, entry, nil, false, false)
			if err != nil {
				t.Fatal(err)
			}
			f := lowered.cycleDiscovery(loaded.Files())
			candidates := f.graphCandidates(loaded.Files())
			f.allocation = f.allocationIndexOf(candidates)
			pairs, shapes := 0, 0
			for _, node := range f.allocation.views {
				related := map[cycleNode]bool{}
				for _, view := range f.allocationViews(node.proven) {
					related[view] = true
				}
				for _, view := range f.allocation.views {
					if related[view] {
						continue
					}
					pairs++
					if f.deriveRelated(node.proven, view.proven) {
						t.Errorf("the sweep cut %s from %s, which deriveRelated relates", check.TypeToString(node.proven), check.TypeToString(view.proven))
					}
				}
				seen := map[*checker.Type]bool{}
				for _, shape := range f.allocationShapesSeenAs(node.proven) {
					seen[shape] = true
				}
				for _, shape := range f.shapes {
					if shape == node.proven || seen[shape] {
						continue
					}
					shapes++
					if check.IsTypeAssignableTo(lowered.relatedShape(shape), node.proven) {
						t.Errorf("the shapes road cut %s from %s, which the checker takes as one", check.TypeToString(shape), check.TypeToString(node.proven))
					}
				}
			}
			t.Logf("%d candidates, %d views, %d cut pairs and %d cut shapes asked of the checker", len(candidates), len(f.allocation.views), pairs, shapes)
			cut += pairs + shapes
		})
	}
	if cut == 0 {
		t.Fatal("no program had a cut pair or shape, so the test pins nothing")
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
