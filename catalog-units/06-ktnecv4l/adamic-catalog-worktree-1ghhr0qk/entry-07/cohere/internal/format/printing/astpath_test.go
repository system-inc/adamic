package printing

import (
	"math/rand"
	"testing"
)

// TestFindAncestorAgreesWithAncestors walks every node of generated toy programs, through children slices
// and the label property both, and checks that FindAncestor answers as a search of Ancestors does
// (#vbjv3d6): for a predicate that matches several ancestors, so the nearest must win, one that matches
// only the root, and one that matches nothing.
func TestFindAncestorAgreesWithAncestors(t *testing.T) {
	t.Parallel()
	predicates := map[string]func(*toyNode) bool{
		"list":    func(node *toyNode) bool { return node.Kind == "List" },
		"program": func(node *toyNode) bool { return node.Kind == "Program" },
		"none":    func(*toyNode) bool { return false },
	}
	matched := map[string]int{}
	checked := 0

	var visit func(path *AstPath[*toyNode])
	visit = func(path *AstPath[*toyNode]) {
		for name, predicate := range predicates {
			expected, expectedFound := searchAncestors(path, predicate)
			actual, actualFound := path.FindAncestor(predicate)
			if actual != expected || actualFound != expectedFound {
				t.Fatalf("%s: FindAncestor found %v (%t), a search of Ancestors %v (%t)", name, actual, actualFound, expected, expectedFound)
			}
			if expectedFound {
				matched[name]++
			}
		}
		checked++
		if node, _ := path.Node(); node.Label != nil {
			Call(path, func(path *AstPath[*toyNode]) struct{} {
				visit(path)
				return struct{}{}
			}, "label")
		}
		path.Each(func(path *AstPath[*toyNode], _ int, _ any) { visit(path) }, "children")
	}
	for seed := int64(1); seed <= 40; seed++ {
		root, _, _ := toyProgram(rand.New(rand.NewSource(seed)))
		visit(NewAstPath(root))
	}

	if matched["list"] == 0 || matched["program"] == 0 || matched["none"] != 0 {
		t.Fatalf("over %d nodes the predicates matched %v; the walk did not reach both a match and none", checked, matched)
	}
}

// searchAncestors is FindAncestor as it was written before #vbjv3d6: a search of the Ancestors list.
func searchAncestors(path *AstPath[*toyNode], predicate func(*toyNode) bool) (*toyNode, bool) {
	for _, ancestor := range path.Ancestors() {
		if predicate(ancestor) {
			return ancestor, true
		}
	}
	return nil, false
}
