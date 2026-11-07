package lower

import "testing"

func TestProgramRegionSCCOwningEdges(test *testing.T) {
	test.Parallel()
	edges := map[string][]string{
		"left":  {"right"},
		"right": {"left", "leaf"},
		"leaf":  {},
		"view1": {"view2"},
		"view2": {"view1"},
	}
	owning := map[string][]string{"left": {"right"}, "right": {"leaf"}}
	selected := ProgramRegionSCC(edges, owning)
	if !selected["left"] || !selected["right"] || selected["leaf"] || selected["view1"] || selected["view2"] {
		test.Fatalf("wrong owning SCC: %v", selected)
	}
}
func TestProgramRegionShapeKeySeparators(test *testing.T) {
	test.Parallel()
	if programShapeKey([]string{"a\x00b"}, nil) == programShapeKey([]string{"a", "b"}, nil) {
		test.Fatal("distinct property shapes collide")
	}
	if programShapeKey([]string{"a\x01b"}, nil) == programShapeKey([]string{"a"}, []string{"b\x01"}) {
		test.Fatal("required and available shape boundary collides")
	}
}
