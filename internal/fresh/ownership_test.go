package fresh

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func ownershipQueryShapes(test *testing.T, name string) {
	test.Helper()
	cases := []struct {
		name    string
		objects []OwnershipObject
		edges   []OwnershipEdge
		roots   []OwnershipRoot
		want    Verdict
		members []int
	}{
		{"flat literal", []OwnershipObject{{ID: 1}}, nil, []OwnershipRoot{{Name: "item", Object: 1}}, Proven, []int{1}},
		{
			"nested object",
			[]OwnershipObject{{ID: 1}, {ID: 2}},
			[]OwnershipEdge{{1, 2, ".child", false}},
			[]OwnershipRoot{{Name: "item", Object: 1}},
			Proven,
			[]int{1, 2},
		},
		{
			"array of objects",
			[]OwnershipObject{{ID: 1}, {ID: 2}, {ID: 3}},
			[]OwnershipEdge{{1, 2, "[0]", false}, {1, 3, "[1]", false}},
			[]OwnershipRoot{{Name: "item", Object: 1}},
			Proven,
			[]int{1, 2, 3},
		},
		{
			"local ring",
			[]OwnershipObject{{ID: 1, Region: 1}, {ID: 2, Region: 1}, {ID: 3, Region: 1}},
			[]OwnershipEdge{{1, 2, ".next", false}, {2, 3, ".next", false}, {3, 1, ".next", false}},
			[]OwnershipRoot{{Name: "item", Object: 1}},
			Proven,
			[]int{1, 2, 3},
		},
		{
			"ring with second outside reference",
			[]OwnershipObject{{ID: 1, Region: 1}, {ID: 2, Region: 1}, {ID: 3}},
			[]OwnershipEdge{{1, 2, ".next", false}, {2, 1, ".next", false}, {3, 2, ".saved", false}},
			[]OwnershipRoot{{Name: "item", Object: 1}, {Name: "other", Object: 3}},
			Refused,
			[]int{1, 2},
		},
	}
	for _, testCase := range cases {
		if testCase.name != name {
			continue
		}
		got := QueryOwnership(OwnershipSnapshot{testCase.objects, testCase.edges, testCase.roots, true}, "item")
		if got.Verdict != testCase.want || !reflect.DeepEqual(got.Members, testCase.members) {
			test.Fatalf("%+v", got)
		}
		if len(got.Evidence) == 0 {
			test.Fatal("missing evidence")
		}
		test.Logf("%s members=%v witnesses=%+v evidence=%v", got.Verdict, got.Members, got.Witnesses, got.Evidence)
	}
}
func TestOwnershipQueryShapesFlatLiteral(test *testing.T) {
	test.Parallel()
	ownershipQueryShapes(test, "flat literal")
}
func TestOwnershipQueryShapesNestedObject(test *testing.T) {
	test.Parallel()
	ownershipQueryShapes(test, "nested object")
}
func TestOwnershipQueryShapesArrayOfObjects(test *testing.T) {
	test.Parallel()
	ownershipQueryShapes(test, "array of objects")
}
func TestOwnershipQueryShapesLocalRing(test *testing.T) {
	test.Parallel()
	ownershipQueryShapes(test, "local ring")
}
func TestOwnershipQueryShapesOutsideRing(test *testing.T) {
	test.Parallel()
	ownershipQueryShapes(test, "ring with second outside reference")
}

func TestOwnershipQuerySecondOwner(test *testing.T) {
	test.Parallel()
	for _, after := range []bool{false, true} {
		snapshot := OwnershipSnapshot{
			Objects:  []OwnershipObject{{ID: 1}},
			Roots:    []OwnershipRoot{{Name: "item", Object: 1}, {Name: "alias", Object: 1, UsedAfter: after}},
			Complete: true,
		}
		got := QueryOwnership(snapshot, "item")
		if got.Verdict != Refused {
			test.Fatalf("second surviving owner after=%v: %+v", after, got)
		}
	}
}
func ownershipQueryUnknownandweak(test *testing.T, name string) {
	test.Helper()
	for _, variant := range []string{
		"missing",
		"opaque",
		"incomplete",
		"weak root",
		"weak edge",
		"borrowed",
		"continuation",
		"disconnected region",
	} {
		if variant != name {
			continue
		}
		snapshot := OwnershipSnapshot{
			Objects:  []OwnershipObject{{ID: 1}},
			Roots:    []OwnershipRoot{{Name: "item", Object: 1}},
			Complete: true,
		}
		want := Refused
		switch variant {
		case "missing":
			snapshot.Objects = nil
			want = Unknown
		case "opaque":
			snapshot.Objects = append(snapshot.Objects, OwnershipObject{ID: 2, Unknown: true})
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "opaque", Object: 2})
			want = Unknown
		case "incomplete":
			snapshot.Complete = false
			want = Unknown
		case "weak root":
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "observer", Object: 1, Weak: true})
		case "weak edge":
			snapshot.Objects = append(snapshot.Objects, OwnershipObject{ID: 2})
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "observer", Object: 2})
			snapshot.Edges = append(snapshot.Edges, OwnershipEdge{2, 1, ".weak", true})
		case "borrowed":
			snapshot.Roots[0].Borrowed = true
		case "continuation":
			snapshot.Roots[0].UsedAfter = true
		case "disconnected region":
			snapshot.Objects[0].Region = 7
			snapshot.Objects = append(snapshot.Objects, OwnershipObject{ID: 2, Region: 7})
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "alias", Object: 2})
		}
		got := QueryOwnership(snapshot, "item")
		if got.Verdict != want {
			test.Fatalf("%+v", got)
		}
	}
}

// Exercise extraction and continuation facts, rather than supplying hand-built
// answers. In particular, fieldless descendants must remain in the inventory.
func TestOwnershipQueryUnknownAndWeakMissing(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "missing")
}
func TestOwnershipQueryUnknownAndWeakOpaque(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "opaque")
}
func TestOwnershipQueryUnknownAndWeakIncomplete(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "incomplete")
}
func TestOwnershipQueryUnknownAndWeakWeakRoot(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "weak root")
}
func TestOwnershipQueryUnknownAndWeakWeakEdge(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "weak edge")
}
func TestOwnershipQueryUnknownAndWeakBorrowed(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "borrowed")
}
func TestOwnershipQueryUnknownAndWeakContinuation(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "continuation")
}
func TestOwnershipQueryUnknownAndWeakDisconnectedRegion(test *testing.T) {
	test.Parallel()
	ownershipQueryUnknownandweak(test, "disconnected region")
}

func ownershipQueryExtractedroots(test *testing.T, name string) {
	test.Helper()
	for _, variant := range []string{
		"private",
		"dead alias",
		"ended alias",
		"later source",
		"weak alias",
		"opaque root",
	} {
		if variant != name {
			continue
		}
		read := ir.Read{Local: 0, Of: ir.Array}
		program := &ir.Program{Locals: []ir.Local{
			{Name: "items", Type: ir.Array, Function: -1},
			{Name: "alias", Type: ir.Array, Function: -1},
			{Name: "work", Type: ir.Closure, Function: -1},
		}}
		program.Main = []ir.Statement{ir.Declare{
			Local: 0,
			Value: ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{}}},
		}}
		want := Proven
		if variant == "dead alias" || variant == "ended alias" {
			program.Main = append(program.Main, ir.Declare{Local: 1, Value: read})
			want = Refused
		}
		if variant == "ended alias" {
			program.Main = append(program.Main, ir.Assign{Local: 1, Value: ir.Undefined{Of: ir.Array}})
			want = Proven
		}
		if variant == "weak alias" {
			program.Locals[1].Type = ir.Weak
			program.Main = append(program.Main, ir.Declare{Local: 1, Value: ir.WeakOf{Value: read}})
			want = Refused
		}
		if variant == "opaque root" {
			program.Main = append(program.Main, ir.Declare{Local: 1, Value: ir.Read{Local: 2, Of: ir.Closure}})
			want = Unknown
		}
		program.Main = append(program.Main, ir.Evaluate{
			Value: ir.ParallelMap{Moved: true, Items: read, Work: ir.Read{Local: 2, Of: ir.Closure}, Result: ir.Object},
		})
		if variant == "later source" {
			program.Main = append(program.Main, ir.Evaluate{Value: read})
			want = Refused
		}
		answers := OwnershipTransfers(program)
		if len(answers) != 1 || answers[0].Answer.Verdict != want {
			test.Fatalf("%+v", answers)
		}
		if answers[0].At == nil {
			test.Fatal("missing boundary source")
		}
	}
}

func TestOwnershipQueryExtractedRootsPrivate(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "private")
}
func TestOwnershipQueryExtractedRootsDeadAlias(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "dead alias")
}
func TestOwnershipQueryExtractedRootsEndedAlias(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "ended alias")
}
func TestOwnershipQueryExtractedRootsLaterSource(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "later source")
}
func TestOwnershipQueryExtractedRootsWeakAlias(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "weak alias")
}
func TestOwnershipQueryExtractedRootsOpaqueRoot(test *testing.T) {
	test.Parallel()
	ownershipQueryExtractedroots(test, "opaque root")
}

func TestOwnershipQueryWholeRegionDescendants(test *testing.T) {
	test.Parallel()
	snapshot := OwnershipSnapshot{
		Complete: true,
		Objects: []OwnershipObject{
			{ID: 1, Region: 1},
			{ID: 2, Region: 1},
			{ID: 3},
			{ID: 4, Region: 2},
			{ID: 5, Region: 2},
			{ID: 6},
		},
		Edges: []OwnershipEdge{{2, 3, ".child", false}, {3, 4, ".region", false}, {5, 6, ".counted", false}},
		Roots: []OwnershipRoot{{Name: "item", Object: 1}, {Name: "alias", Object: 6}},
	}
	got := QueryOwnership(snapshot, "item")
	if got.Verdict != Refused || !reflect.DeepEqual(got.Members, []int{1, 2, 3, 4, 5, 6}) {
		test.Fatalf("%+v", got)
	}
}

func TestOwnershipQueryProgramAllocationIsUnknown(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Name: "items", Type: ir.Array, Function: -1}}}
	program.Main = []ir.Statement{ir.Declare{
		Local: 0,
		Value: ir.ArrayLiteral{ProgramRegion: true, Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{}}},
	}, ir.Evaluate{
		Value: ir.ParallelMap{Moved: true, Items: ir.Read{Local: 0, Of: ir.Array}, Work: ir.Undefined{Of: ir.Closure}},
	}}
	answers := OwnershipTransfers(program)
	if len(answers) != 1 || answers[0].Answer.Verdict != Unknown {
		test.Fatalf("Program membership is not transfer permission: %+v", answers)
	}
}

// A type identity cannot stand in for the missing dynamic graph inventory.
func TestOwnershipQueryGraphAllocationIsUnknown(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		GraphTypes: map[int]bool{1: true},
		Locals:     []ir.Local{{Name: "items", Type: ir.Array, Function: -1}},
	}
	program.Main = []ir.Statement{ir.Declare{
		Local: 0,
		Value: ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{GraphTypes: []int{1}}}},
	}, ir.Evaluate{
		Value: ir.ParallelMap{Moved: true, Items: ir.Read{Local: 0, Of: ir.Array}, Work: ir.Undefined{Of: ir.Closure}},
	}}
	answers := OwnershipTransfers(program)
	if len(answers) != 1 || answers[0].Answer.Verdict != Unknown {
		test.Fatalf("graph membership is not transfer permission: %+v", answers)
	}
}
