package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Only readonly consumers can use a representation checked at lookup. A read
// certificate never grants mutation of a fixed object's storage.
func (l *lowering) dictionaryReferenceConversion(node *ast.Node) bool {
	if node.Kind != ast.KindAsExpression {
		return false
	}
	as := node.AsAsExpression()
	source, target := l.concrete(l.checker.GetTypeAtLocation(as.Expression)), l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.primitiveDictionaryCandidate(source) || !l.stringDictionary(target) {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(target)
	return len(infos) == 1 && infos[0].IsReadonly() && len(l.checker.GetPropertiesOfType(target)) == 0
}

func (l *lowering) checkedDictionaryReferenceConversion(node *ast.Node) (ir.Expression, error) {
	value, err := l.expression(node.AsAsExpression().Expression)
	if err != nil {
		return nil, err
	}
	// Check the heap tag when this selected reference is consumed. Lookups
	// independently check fixed slots or record.c storage and element types.
	converted := ir.Narrow{Value: value, To: ir.Object, Dictionary: true, DictionaryWhere: l.program.Where(node), DictionaryExpression: sourceExpression(node.AsAsExpression().Expression), DictionaryType: l.checker.TypeToString(l.concrete(l.checker.GetTypeAtLocation(node)))}
	return l.view(node, converted, l.concrete(l.checker.GetTypeAtLocation(node)))
}

// Follow the same allocation graph as checked reads. Aliases, erasure and
// helper arguments cannot turn this read certificate into record storage.
func dictionaryConversionRecordOperations(program *ir.Program, graph *allocationFlowGraph) error {
	sites := map[int]bool{}
	unknown := false
	where := ""
	visit := func(value any) bool {
		if narrow, ok := value.(ir.Narrow); ok && narrow.Dictionary {
			if where == "" {
				where = narrow.DictionaryWhere
			}
			reaching := graph.ReachingAllocations(narrow.Value)
			unknown = unknown || reaching.Unknown
			for _, site := range reaching.Sites {
				sites[site] = true
			}
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	if where == "" {
		return nil
	}
	var refused error
	inspect := func(value any) bool {

		var receiver ir.Expression
		operation := ""
		switch value := value.(type) {
		case ir.RecordCall:
			if len(value.Arguments) == 0 || value.Method == "get" && value.DictionaryRead != nil {
				return true
			}
			receiver, operation = value.Arguments[0], "record "+value.Method
		case ir.RecordCoalesce:
			receiver, operation = value.Record, "record coalesce"
		case ir.ObjectCall:
			if len(value.Arguments) == 0 {
				return true
			}
			receiver, operation = value.Arguments[0], "Object."+value.Method
		case ir.ObjectKeys:
			receiver, operation = value.Object, "key enumeration"
		case ir.HasOwn:
			receiver, operation = value.Object, "own-key probe"
		case ir.SetProperty:
			receiver, operation = value.Object, "field set"
		case ir.ObjectLiteral:
			receiver, operation = value.Spread, "object spread"
		case ir.RecordLiteral:
			receiver, operation = value.Spread, "record spread"
		default:
			return true
		}
		if receiver == nil {
			return true
		}
		reaching := graph.ReachingAllocations(receiver)
		demanded := unknown || reaching.Unknown
		for _, site := range reaching.Sites {
			demanded = demanded || sites[site]
		}
		if demanded && refused == nil {
			refused = &Refused{Where: where, What: operation + " through a readonly dictionary view without a producer storage certificate", Fix: "prove the producer storage before this record operation; this conversion grants checked reads only"}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return refused
}
