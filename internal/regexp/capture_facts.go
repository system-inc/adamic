package regexp

import "strconv"

// CaptureParticipation describes captures in source order. Indices includes the
// whole match at index zero. A true entry proves presence on every successful
// match; false is conservative and must retain string | undefined. Names merges
// duplicate names across alternatives, independently of numeric positions.
// It does not prove reachability or use constraints imposed by other assertions.
type CaptureParticipation struct {
	Indices []bool
	Names   map[string]bool
}

// CaptureFacts parses a literal pattern and proves capture participation using
// ECMA-262 CompileSubpattern and RepeatMatcher's capture-reset semantics.
// Runtime-built patterns must not use these literal-specific proofs.
func CaptureFacts(pattern, flags string) (CaptureParticipation, error) {
	tree, err := Parse(pattern, flags)
	if err != nil {
		return CaptureParticipation{}, err
	}
	ids := map[*Group]int{}
	names := map[string]bool{}
	var enumerate func(Node)
	enumerate = func(n Node) {
		switch n := n.(type) {
		case *Disjunction:
			for _, a := range n.Alternatives {
				for _, t := range a.Terms {
					enumerate(t)
				}
			}
		case *Quantifier:
			enumerate(n.Atom)
		case *Group:
			if n.Kind == Capturing {
				ids[n] = len(ids) + 1
				if n.Name != "" {
					names[n.Name] = false
				}
			}
			enumerate(n.Body)
		}
	}
	enumerate(tree.Body)
	type present map[string]bool
	var visit func(Node) present
	visit = func(n Node) present {
		result := present{}
		switch n := n.(type) {
		case *Disjunction:
			for i, a := range n.Alternatives {
				branch := present{}
				for _, t := range a.Terms {
					for key := range visit(t) {
						branch[key] = true
					}
				}
				if i == 0 {
					result = branch
				} else {
					for key := range result {
						if !branch[key] {
							delete(result, key)
						}
					}
				}
			}
		case *Quantifier:
			// Every iteration resets its enclosed captures. Only guarantees from the
			// final iteration survive, and a zero minimum permits no iteration at all.
			if n.Min.Sign() > 0 {
				result = visit(n.Atom)
			}
		case *Group:
			// Captures in a successful negative assertion are always undefined.
			if n.Kind != NegativeLookahead && n.Kind != NegativeLookbehind {
				result = visit(n.Body)
				if n.Kind == Capturing {
					result[indexKey(ids[n])] = true
					if n.Name != "" {
						result["name:"+n.Name] = true
					}
				}
			}
			// An unmatched (including forward) backreference can match the empty
			// string. It therefore supplies no participation proof for its target.
		}
		return result
	}
	proven := visit(tree.Body)
	facts := CaptureParticipation{Indices: make([]bool, len(ids)+1), Names: names}
	facts.Indices[0] = true
	for i := 1; i < len(facts.Indices); i++ {
		facts.Indices[i] = proven[indexKey(i)]
	}
	for name := range names {
		names[name] = proven["name:"+name]
	}
	return facts, nil
}

func indexKey(index int) string { return "index:" + strconv.Itoa(index) }
