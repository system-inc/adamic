package doc

import "fmt"

// traverseExit marks, on the traversal stack, that the doc beneath it is being left rather than entered.
type traverseExit struct{}

func (traverseExit) isDoc() {}

// traverse is upstream's traverseDoc with shouldTraverseConditionalGroups set, which is the only way
// this package calls it. utilities/traverse-doc.js.
//
// It is iterative with an explicit stack, as upstream is, so the order children are entered matches
// exactly. IfBreak enters its break contents before its flat contents, because upstream pushes flat
// first onto a stack.
func traverse(root Doc, onEnter func(Doc) bool, onExit func(Doc)) {
	stack := []Doc{root}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, isExit := current.(traverseExit); isExit {
			leaving := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onExit(leaving)
			continue
		}

		stack = append(stack, current, traverseExit{})

		if !onEnter(current) {
			continue
		}

		switch document := current.(type) {
		case *Fill:
			for index := len(document.Parts) - 1; index >= 0; index-- {
				stack = append(stack, document.Parts[index])
			}
		case *IfBreak:
			stack = append(stack, document.FlatContents, document.BreakContents)
		case *Group:
			if document.ExpandedStates != nil {
				for index := len(document.ExpandedStates) - 1; index >= 0; index-- {
					stack = append(stack, document.ExpandedStates[index])
				}
			} else {
				stack = append(stack, document.Contents)
			}
		case *Align:
			stack = append(stack, document.Contents)
		case *Indent:
			stack = append(stack, document.Contents)
		case *IndentIfBreak:
			stack = append(stack, document.Contents)
		case *Label:
			stack = append(stack, document.Contents)
		case *LineSuffix:
			stack = append(stack, document.Contents)
		case Text, trimDoc, lineSuffixBoundaryDoc, *Line, breakParentDoc:
		default:
			parts, isArray := Parts(current)
			if !isArray {
				panic(fmt.Sprintf("invalid doc %T", current))
			}
			for index := len(parts) - 1; index >= 0; index-- {
				stack = append(stack, parts[index])
			}
		}
	}
}

// PropagateBreaks is upstream's propagateBreaks, utilities/index.js: a hard break anywhere inside a
// group breaks that group and every group around it, except across a conditional group, whose
// alternatives the printer chooses between explicitly.
//
// It writes Group.Break in place. A group shared by two parents is entered on both paths but its
// contents walked once, which is upstream's alreadyVisitedSet keyed on object identity. That is an
// optimization rather than behaviour: the first walk sets every break the second could, and a mutation
// that walked shared groups twice changed no output. Entering on both paths is what matters, because
// the exit is what carries a broken shared group's break up to its second parent.
func PropagateBreaks(root Doc) {
	visited := map[*Group]bool{}
	var groupStack []*Group

	breakParentGroup := func() {
		if len(groupStack) > 0 {
			parent := groupStack[len(groupStack)-1]
			if parent.ExpandedStates == nil && !parent.Break {
				parent.Break = true
			}
		}
	}

	traverse(root,
		func(document Doc) bool {
			if _, isBreakParent := document.(breakParentDoc); isBreakParent {
				breakParentGroup()
			}
			if group, isGroup := document.(*Group); isGroup {
				groupStack = append(groupStack, group)
				if visited[group] {
					return false
				}
				visited[group] = true
			}
			return true
		},
		func(document Doc) {
			if _, isGroup := document.(*Group); isGroup {
				group := groupStack[len(groupStack)-1]
				groupStack = groupStack[:len(groupStack)-1]
				if group.Break {
					breakParentGroup()
				}
			}
		},
	)
}
