package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// wholeObjectCopy is a value copied whole into another object: an object spread's source, or an
// Object.assign source. A rest binding copies whole too, and destructuring refuses it before it
// lowers (destructureFrom).
type wholeObjectCopy struct {
	node   *ast.Node
	source ir.Expression
	copied *checker.Type
}

// A checked view's fields are checked where they're read, under the view's receiver type
// (CheckedFieldKey). A whole-object copy makes an object whose fields are no view's, so it would
// move a lying layout's fields as the types the view claims, unchecked. Until a copy checks each
// field it moves, a copy of a value of a view's receiver type is NotYet. Views register as their
// casts lower, so this is decided once lowering is done: a copy lowered before its view's cast is
// judged by the same registry. A copy through another type that names a checked field is NotYet
// too when the value may be a viewed allocation, as checkLazyViewReads' field-name fallback is.
func (l *lowering) checkViewCopies() error {
	if len(l.wholeObjectCopies) == 0 {
		return nil
	}
	receivers := map[int]bool{}
	names := map[string]bool{}
	for key := range l.result.CheckedFields {
		// A record's field is checked under its bare name, and is no view's.
		receiver, name, keyed := strings.Cut(key, ":")
		id, err := strconv.Atoi(receiver)
		if !keyed || err != nil {
			continue
		}
		receivers[id] = true
		names[name] = true
	}
	if len(receivers) == 0 {
		return nil
	}
	var graph *allocationFlowGraph
	var viewed map[int]bool
	unknown := len(l.result.ViewOrigins) == 0
	for _, copied := range l.wholeObjectCopies {
		if view := l.checkedViewType(copied.copied, receivers); view != nil {
			return l.notYet(copied.node, "a whole-object copy of a value of checked view type "+l.checker.TypeToString(view))
		}
		named := false
		for _, property := range l.checker.GetPropertiesOfType(l.checker.GetNonNullableType(copied.copied)) {
			named = named || names[property.Name]
		}
		if !named {
			continue
		}
		if graph == nil {
			var reachUnknown bool
			graph, viewed, reachUnknown = viewedAllocations(l.result)
			unknown = unknown || reachUnknown
		}
		reaches := graph.ReachingAllocations(copied.source)
		demanded := unknown || reaches.Unknown
		for _, site := range reaches.Sites {
			demanded = demanded || viewed[site]
		}
		if demanded {
			return l.notYet(copied.node, "a whole-object copy of a value of type "+l.checker.TypeToString(copied.copied)+" that may hold a checked view's value")
		}
	}
	return nil
}

// checkedViewType is the view receiver type a copied value's type is, or holds as a union or
// intersection member, or nil.
func (l *lowering) checkedViewType(copied *checker.Type, receivers map[int]bool) *checker.Type {
	if receivers[l.viewReceiverTypeID(copied)] {
		return copied
	}
	copied = l.checker.GetNonNullableType(l.concrete(copied))
	if copied.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range copied.Types() {
			if view := l.checkedViewType(member, receivers); view != nil {
				return view
			}
		}
	}
	return nil
}
