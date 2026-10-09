package lower

import "testing"

func TestProgramRegionSCCOwningEdges(t *testing.T) {
	t.Parallel()
	edges := map[string][]string{"left": {"right"}, "right": {"left", "leaf"}, "leaf": {}, "view1": {"view2"}, "view2": {"view1"}}
	owning := map[string][]string{"left": {"right"}, "right": {"leaf"}}
	selected := ProgramRegionSCC(edges, owning)
	if !selected["left"] || !selected["right"] || selected["leaf"] || selected["view1"] || selected["view2"] {
		t.Fatalf("wrong owning SCC: %v", selected)
	}
}
func TestProgramRegionShapeKeySeparators(t *testing.T) {
	t.Parallel()
	if programShapeKey([]string{"a\x00b"}, nil) == programShapeKey([]string{"a", "b"}, nil) {
		t.Fatal("distinct property shapes collide")
	}
	if programShapeKey([]string{"a\x01b"}, nil) == programShapeKey([]string{"a"}, []string{"b\x01"}) {
		t.Fatal("required and available shape boundary collides")
	}
}
