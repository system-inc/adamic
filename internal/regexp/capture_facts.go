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
	tree, parseError := Parse(pattern, flags)
	if parseError != nil {
		return CaptureParticipation{}, parseError
	}
	ids := map[*Group]int{}
	names := map[string]bool{}
	var enumerate func(Node)
	enumerate = func(node Node) {
		switch typed := node.(type) {
		case *Disjunction:
			for _, alternative := range typed.Alternatives {
				for _, term := range alternative.Terms {
					enumerate(term)
				}
			}
		case *Quantifier:
			enumerate(typed.Atom)
		case *Group:
			if typed.Kind == Capturing {
				ids[typed] = len(ids) + 1
				if typed.Name != "" {
					names[typed.Name] = false
				}
			}
			enumerate(typed.Body)
		}
	}
	enumerate(tree.Body)
	type present map[string]bool
	var visit func(Node) present
	visit = func(node Node) present {
		result := present{}
		switch typed := node.(type) {
		case *Disjunction:
			for index, alternative := range typed.Alternatives {
				branch := present{}
				for _, term := range alternative.Terms {
					for key := range visit(term) {
						branch[key] = true
					}
				}
				if index == 0 {
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
			if typed.Min.Sign() > 0 {
				result = visit(typed.Atom)
			}
		case *Group:
			// Captures in a successful negative assertion are always undefined.
			if typed.Kind != NegativeLookahead && typed.Kind != NegativeLookbehind {
				result = visit(typed.Body)
				if typed.Kind == Capturing {
					result[indexKey(ids[typed])] = true
					if typed.Name != "" {
						result["name:"+typed.Name] = true
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
	for index := 1; index < len(facts.Indices); index++ {
		facts.Indices[index] = proven[indexKey(index)]
	}
	for name := range names {
		names[name] = proven["name:"+name]
	}
	return facts, nil
}

func indexKey(index int) string { return "index:" + strconv.Itoa(index) }
