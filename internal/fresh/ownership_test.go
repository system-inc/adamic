package fresh

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func TestOwnershipQueryShapes(t *testing.T) {
	cases := []struct {
		name    string
		objects []OwnershipObject
		edges   []OwnershipEdge
		roots   []OwnershipRoot
		want    Verdict
		members []int
	}{
		{"flat literal", []OwnershipObject{{ID: 1}}, nil, []OwnershipRoot{{Name: "item", Object: 1}}, Proven, []int{1}},
		{"nested object", []OwnershipObject{{ID: 1}, {ID: 2}}, []OwnershipEdge{{1, 2, ".child", false}}, []OwnershipRoot{{Name: "item", Object: 1}}, Proven, []int{1, 2}},
		{"array of objects", []OwnershipObject{{ID: 1}, {ID: 2}, {ID: 3}}, []OwnershipEdge{{1, 2, "[0]", false}, {1, 3, "[1]", false}}, []OwnershipRoot{{Name: "item", Object: 1}}, Proven, []int{1, 2, 3}},
		{"local ring", []OwnershipObject{{ID: 1, Region: 1}, {ID: 2, Region: 1}, {ID: 3, Region: 1}}, []OwnershipEdge{{1, 2, ".next", false}, {2, 3, ".next", false}, {3, 1, ".next", false}}, []OwnershipRoot{{Name: "item", Object: 1}}, Proven, []int{1, 2, 3}},
		{"ring with second outside reference", []OwnershipObject{{ID: 1, Region: 1}, {ID: 2, Region: 1}, {ID: 3}}, []OwnershipEdge{{1, 2, ".next", false}, {2, 1, ".next", false}, {3, 2, ".saved", false}}, []OwnershipRoot{{Name: "item", Object: 1}, {Name: "other", Object: 3}}, Refused, []int{1, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := QueryOwnership(OwnershipSnapshot{c.objects, c.edges, c.roots, true}, "item")
			if got.Verdict != c.want || !reflect.DeepEqual(got.Members, c.members) {
				t.Fatalf("%+v", got)
			}
			if len(got.Evidence) == 0 {
				t.Fatal("missing evidence")
			}
			t.Logf("%s members=%v witnesses=%+v evidence=%v", got.Verdict, got.Members, got.Witnesses, got.Evidence)
		})
	}
}
func TestOwnershipQuerySecondOwner(t *testing.T) {
	for _, after := range []bool{false, true} {
		s := OwnershipSnapshot{Objects: []OwnershipObject{{ID: 1}}, Roots: []OwnershipRoot{{Name: "item", Object: 1}, {Name: "alias", Object: 1, UsedAfter: after}}, Complete: true}
		got := QueryOwnership(s, "item")
		if got.Verdict != Refused {
			t.Fatalf("second surviving owner after=%v: %+v", after, got)
		}
	}
}
func TestOwnershipQueryUnknownAndWeak(t *testing.T) {
	for _, variant := range []string{"missing", "opaque", "incomplete", "weak root", "weak edge", "borrowed", "continuation", "disconnected region"} {
		t.Run(variant, func(t *testing.T) {
			s := OwnershipSnapshot{Objects: []OwnershipObject{{ID: 1}}, Roots: []OwnershipRoot{{Name: "item", Object: 1}}, Complete: true}
			want := Refused
			switch variant {
			case "missing":
				s.Objects = nil
				want = Unknown
			case "opaque":
				s.Objects = append(s.Objects, OwnershipObject{ID: 2, Unknown: true})
				s.Roots = append(s.Roots, OwnershipRoot{Name: "opaque", Object: 2})
				want = Unknown
			case "incomplete":
				s.Complete = false
				want = Unknown
			case "weak root":
				s.Roots = append(s.Roots, OwnershipRoot{Name: "observer", Object: 1, Weak: true})
			case "weak edge":
				s.Objects = append(s.Objects, OwnershipObject{ID: 2})
				s.Roots = append(s.Roots, OwnershipRoot{Name: "observer", Object: 2})
				s.Edges = append(s.Edges, OwnershipEdge{2, 1, ".weak", true})
			case "borrowed":
				s.Roots[0].Borrowed = true
			case "continuation":
				s.Roots[0].UsedAfter = true
			case "disconnected region":
				s.Objects[0].Region = 7
				s.Objects = append(s.Objects, OwnershipObject{ID: 2, Region: 7})
				s.Roots = append(s.Roots, OwnershipRoot{Name: "alias", Object: 2})
			}
			got := QueryOwnership(s, "item")
			if got.Verdict != want {
				t.Fatalf("%+v", got)
			}
		})
	}
}

// Exercise extraction and continuation facts, rather than supplying hand-built
// answers. In particular, fieldless descendants must remain in the inventory.
func TestOwnershipQueryExtractedRoots(t *testing.T) {
	for _, variant := range []string{"private", "dead alias", "ended alias", "later source", "weak alias", "opaque root"} {
		t.Run(variant, func(t *testing.T) {
			read := ir.Read{Local: 0, Of: ir.Array}
			program := &ir.Program{Locals: []ir.Local{{Name: "items", Type: ir.Array, Function: -1}, {Name: "alias", Type: ir.Array, Function: -1}, {Name: "work", Type: ir.Closure, Function: -1}}}
			program.Main = []ir.Statement{ir.Declare{Local: 0, Value: ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{}}}}}
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
			program.Main = append(program.Main, ir.Evaluate{Value: ir.ParallelMap{Moved: true, Items: read, Work: ir.Read{Local: 2, Of: ir.Closure}, Result: ir.Object}})
			if variant == "later source" {
				program.Main = append(program.Main, ir.Evaluate{Value: read})
				want = Refused
			}
			answers := OwnershipTransfers(program)
			if len(answers) != 1 || answers[0].Answer.Verdict != want {
				t.Fatalf("%+v", answers)
			}
			if answers[0].At == nil {
				t.Fatal("missing boundary source")
			}
		})
	}
}

func TestOwnershipQueryWholeRegionDescendants(t *testing.T) {
	s := OwnershipSnapshot{Complete: true, Objects: []OwnershipObject{{ID: 1, Region: 1}, {ID: 2, Region: 1}, {ID: 3}, {ID: 4, Region: 2}, {ID: 5, Region: 2}, {ID: 6}}, Edges: []OwnershipEdge{{2, 3, ".child", false}, {3, 4, ".region", false}, {5, 6, ".counted", false}}, Roots: []OwnershipRoot{{Name: "item", Object: 1}, {Name: "alias", Object: 6}}}
	got := QueryOwnership(s, "item")
	if got.Verdict != Refused || !reflect.DeepEqual(got.Members, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("%+v", got)
	}
}
