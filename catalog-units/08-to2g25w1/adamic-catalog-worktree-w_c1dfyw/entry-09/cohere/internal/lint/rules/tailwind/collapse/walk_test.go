package tailwind

import (
	"strings"
	"testing"
)

// walkFixture is a tree with a node at every kind and enough depth that a traversal bug shows up as
// a different sequence rather than a missing leaf.
//
// Named nodes so the assertions read as sequences rather than as structure. Declarations carry
// their name as the property; rules carry it as the selector; at-rules as params; comments and
// contexts as their own marker.
func walkFixture() []*Node {
	return []*Node{
		Declaration("a", "1"),
		StyleRule(".b",
			Declaration("b1", "1"),
			AtRule("@media", "b2",
				Declaration("b2a", "1"),
				Declaration("b2b", "1"),
			),
			Declaration("b3", "1"),
		),
		AtRoot(
			AtRule("@property", "c1",
				Declaration("c1a", "1"),
			),
		),
		Context(map[string]string{"theme": "true"},
			Declaration("d1", "1"),
			StyleRule(".d2",
				Declaration("d2a", "1"),
			),
		),
		Comment("e"),
		Declaration("f", "1"),
	}
}

// label renders a node as the single token the walk assertions compare on.
func label(node *Node) string {
	switch node.Kind {
	case KindDeclaration:
		return node.Property
	case KindComment:
		return "comment:" + node.Value
	case KindRule:
		return "rule" + node.Selector
	case KindAtRule:
		return "atrule:" + node.Params
	case KindAtRoot:
		return "atroot"
	case KindContext:
		return "context"
	default:
		return "unknown"
	}
}

// record walks the fixture with the given visitor and returns the visit sequence.
func record(nodes []*Node, decide func(node *Node) WalkAction) string {
	var visited []string
	Walk(nodes, func(node *Node) WalkAction {
		visited = append(visited, label(node))
		if decide == nil {
			return WalkContinue
		}
		return decide(node)
	})
	return strings.Join(visited, " ")
}

// TestWalkVisitsDepthFirstInDocumentOrder pins the full sequence.
//
// Written out in full rather than checked for length or membership, because the failure modes that
// matter here are reorderings and omissions, and both survive a weaker assertion. Every container
// kind is descended into, including at-root and context, which is upstream's behavior and
// deliberately unlike PropertySort's.
func TestWalkVisitsDepthFirstInDocumentOrder(t *testing.T) {
	t.Parallel()
	want := "a rule.b b1 atrule:b2 b2a b2b b3 atroot atrule:c1 c1a context d1 rule.d2 d2a comment:e f"

	if got := record(walkFixture(), nil); got != want {
		t.Errorf("visit order\n got: %s\nwant: %s", got, want)
	}
}

// TestWalkSkipPrunesTheSubtreeAndContinues states what Skip does and, just as importantly, what it
// does not do: it prunes children, it does not end the walk and it does not skip the next sibling.
func TestWalkSkipPrunesTheSubtreeAndContinues(t *testing.T) {
	t.Parallel()
	t.Run("prunes a rule's children and continues to the next sibling", func(t *testing.T) {
		t.Parallel()
		want := "a rule.b atroot atrule:c1 c1a context d1 rule.d2 d2a comment:e f"
		got := record(walkFixture(), func(node *Node) WalkAction {
			if node.Kind == KindRule && node.Selector == ".b" {
				return WalkSkip
			}
			return WalkContinue
		})
		if got != want {
			t.Errorf("visit order\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("prunes only the node it is returned for", func(t *testing.T) {
		t.Parallel()
		want := "a rule.b b1 atrule:b2 b3 atroot atrule:c1 c1a context d1 rule.d2 d2a comment:e f"
		got := record(walkFixture(), func(node *Node) WalkAction {
			if node.Kind == KindAtRule && node.Params == "b2" {
				return WalkSkip
			}
			return WalkContinue
		})
		if got != want {
			t.Errorf("visit order\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("skipping a leaf is indistinguishable from continuing", func(t *testing.T) {
		t.Parallel()
		skipped := record(walkFixture(), func(node *Node) WalkAction {
			if node.Kind == KindDeclaration {
				return WalkSkip
			}
			return WalkContinue
		})
		if plain := record(walkFixture(), nil); skipped != plain {
			t.Errorf("skipping leaves changed the walk\n got: %s\nwant: %s", skipped, plain)
		}
	})
}

// TestWalkStopEndsTheWalkImmediately states that Stop unwinds every level rather than only the
// current one. A walker that returns from the innermost recursion without propagating would visit
// the stopped node's later siblings and its ancestors' later siblings, which the deep case catches.
func TestWalkStopEndsTheWalkImmediately(t *testing.T) {
	t.Parallel()
	t.Run("at the top level", func(t *testing.T) {
		t.Parallel()
		want := "a rule.b b1 atrule:b2 b2a b2b b3 atroot"
		got := record(walkFixture(), func(node *Node) WalkAction {
			if node.Kind == KindAtRoot {
				return WalkStop
			}
			return WalkContinue
		})
		if got != want {
			t.Errorf("visit order\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("from the deepest node, unwinding every level", func(t *testing.T) {
		t.Parallel()
		want := "a rule.b b1 atrule:b2 b2a"
		got := record(walkFixture(), func(node *Node) WalkAction {
			if node.Kind == KindDeclaration && node.Property == "b2a" {
				return WalkStop
			}
			return WalkContinue
		})
		if got != want {
			t.Errorf("visit order\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("on the very first node", func(t *testing.T) {
		t.Parallel()
		want := "a"
		got := record(walkFixture(), func(node *Node) WalkAction {
			return WalkStop
		})
		if got != want {
			t.Errorf("visit order\n got: %s\nwant: %s", got, want)
		}
	})
}

// TestWalkDescendsIntoWrappersUnlikePropertySort states the one place the two traversals in this
// package deliberately disagree, so that a later change unifying them fails here with the reason
// attached.
func TestWalkDescendsIntoWrappersUnlikePropertySort(t *testing.T) {
	t.Parallel()
	buried := Declaration("display", "flex")

	for _, wrapper := range []struct {
		name string
		node *Node
	}{
		{"at-root", AtRoot(buried)},
		{"context", Context(map[string]string{"theme": "true"}, buried)},
	} {
		if got := record([]*Node{wrapper.node}, nil); !strings.Contains(got, "display") {
			t.Errorf("%s: Walk did not descend into it (%s); upstream's walk descends on the presence of children", wrapper.name, got)
		}
		if got := PropertySort([]*Node{wrapper.node}); got.Count != 0 {
			t.Errorf("%s: PropertySort descended into it; the two traversals must differ here", wrapper.name)
		}
	}
}

// TestWalkOnEmptyInput covers the edges that a recursive walker gets wrong silently.
func TestWalkOnEmptyInput(t *testing.T) {
	t.Parallel()
	Walk(nil, func(node *Node) WalkAction {
		t.Error("visited a node in a nil tree")
		return WalkContinue
	})

	visits := 0
	Walk([]*Node{StyleRule(".empty")}, func(node *Node) WalkAction {
		visits++
		return WalkContinue
	})
	if visits != 1 {
		t.Errorf("a childless rule was visited %d times, want 1", visits)
	}
}
