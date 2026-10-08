package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only represented scalar source arrays enter the first storage bridge. Target
// descendants remain complete and lazy. Existing tuple writes still refuse.
func (l *lowering) arrayTupleCastCandidate(source, target *checker.Type) bool {
	return l.checker.IsArrayType(source) && l.isLibraryType(source, "Array") && fixedViewTuple(target) && interfaceScalar(l.concrete(l.viewArrayElementType(source)))
}
func (l *lowering) viewArrayTupleCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Array || !l.arrayTupleCastCandidate(source, target) {
		return nil, nil
	}
	if _, err := l.viewContract(node, source); err != nil {
		return nil, err
	}
	alias := ir.Narrow{Value: value, To: ir.Object, Tuple: true}
	return l.view(node, alias, target)
}

// Keep consumers that read raw object layouts closed only when this array view
// can reach them. Queries include stored descendants and joined helper flows.
func (l *lowering) checkArrayTupleConsumers(graph *allocationFlowGraph) error {
	if !ir.HasArrayTupleViews(l.result) {
		return nil
	}
	sources := map[int]bool{}
	sourceUnknown := false
	for _, origin := range l.result.ViewOrigins {
		marker, ok := origin.(ir.Narrow)
		if !ok || !marker.Tuple || marker.To != ir.Object || marker.Value.Type() != ir.Array {
			continue
		}
		set := graph.ReachingAllocations(marker.Value)
		sourceUnknown = sourceUnknown || set.Unknown
		for _, site := range set.Sites {
			sources[site] = true
		}
	}
	index := graph.projectionIndex()
	var reaches func(ir.Expression, bool, map[int]bool) bool
	reaches = func(value ir.Expression, descendants bool, seen map[int]bool) bool {
		if value == nil || !viewAggregate(value) {
			return false
		}
		set := graph.ReachingAllocations(value)
		if sourceUnknown || set.Unknown {
			return true
		}
		for _, site := range set.Sites {
			if sources[site] {
				return true
			}
			if !descendants || seen[site] {
				continue
			}
			seen[site] = true
			if record, ok := index.records[site]; ok {
				for _, field := range record.Fields {
					if reaches(field.Value, true, seen) {
						return true
					}
				}
				if reaches(record.Spread, true, seen) {
					return true
				}
			}
			if array, ok := index.arrays[site]; ok {
				if index.arrayMutation {
					return true
				}
				for _, element := range array.Elements {
					if reaches(element, true, seen) {
						return true
					}
				}
			}
			for _, values := range index.stores[site] {
				for _, stored := range values {
					if reaches(stored, true, seen) {
						return true
					}
				}
			}
		}
		return false
	}
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		var value ir.Expression
		family := ""
		descendants := false
		switch read := node.(type) {
		case ir.Property:
			if read.View == "" || ir.ArrayTupleReceiver(l.result, read) == 0 {
				value, family = read.Object, "positional read without a represented tuple receiver"
			}
		case ir.JSONStringify:
			value, family, descendants = read.Value, "JSON serialization", true
		case ir.ObjectLiteral:
			value, family = read.Spread, "object spread"
		case ir.ObjectCall:
			if len(read.Arguments) > 0 {
				value, family = read.Arguments[0], "Object consumer"
			}
		case ir.ObjectKeys:
			value, family = read.Object, "Object.keys"
		case ir.HasOwn:
			value, family = read.Object, "own-field query"
		case ir.MapNew:
			for _, entry := range read.Entries {
				for _, item := range entry {
					if reaches(item, true, map[int]bool{}) {
						value, family, descendants = item, "Map storage", true
					}
				}
			}
			if reaches(read.Pairs, true, map[int]bool{}) {
				value, family, descendants = read.Pairs, "Map pair storage", true
			}
		case ir.MapSet:
			for _, item := range []ir.Expression{read.Key, read.Value} {
				if reaches(item, true, map[int]bool{}) {
					value, family, descendants = item, "Map storage", true
				}
			}
		case ir.SetNew:
			value, family, descendants = read.Values, "Set storage", true
		case ir.SetAdd:
			value, family, descendants = read.Value, "Set storage", true
		}
		if family != "" && reaches(value, descendants, map[int]bool{}) {
			refused = &Refused{What: "checked array-to-tuple view consumer requiring " + family + " storage adaptation", Fix: "use checked positional or length reads until this consumer preserves the source array storage"}
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	return refused
}
