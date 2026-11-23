package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// dynamicRecordReadGap stops a named dynamic read (ir.DynamicProperty, adamic_dynamic_property) of a value
// that may be a record holding numbers or booleans (#859ksab). The runtime's record branch returns
// adamic_retain(slot->reference) whatever the record's value storage (union.c:332 to :337), so a number's
// unboxed bits were retained as a pointer: `'k' in o` then `typeof o.k`, with o a Record<string, number>
// passed as object, faulted natively under ASan where Node and the JavaScript backend print number
// (Language's J-t5mxr1v and J-dwhn493 on main 0c9b1201). A record of references reads right, so a program
// with no scalar record keeps lowering. A read whose receiver can't be found fails closed. A run-time key read (adamic_dynamic_property_keyed) boxes a record's
// value by its kind (record.c), so it isn't this road. Library's #3tjjcjs fixes the branch, which lifts this.
func (l *lowering) dynamicRecordReadGap(node *ast.Node, name string) error {
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
	}
	if !l.scalarRecordsInProgram() {
		return nil
	}
	if receiver != nil && !l.viewMayHoldRecord(l.checker.GetTypeAtLocation(receiver), 0) {
		return nil
	}
	return l.notYet(node, "a read of '"+name+"' through a value that may be a record of numbers or booleans, which a named dynamic read doesn't box yet")
}

// viewMayHoldRecord reports whether a value seen as proven may be a record at run time: proven is object,
// unknown, any or a type parameter, or a record itself, through any union or intersection part. The
// library's Record<K, unknown> an in test intersects in adds a key, not storage, so it's skipped
// (inNarrowingAdds' member); a program object type pins fixed storage, which no record reaches.
func (l *lowering) viewMayHoldRecord(proven *checker.Type, depth int) bool {
	if proven == nil || depth > 16 {
		return true
	}
	proven = l.concrete(proven)
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, part := range proven.Types() {
			if part.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) == 0 && l.viewMayHoldRecord(part, depth+1) {
				return true
			}
		}
		return false
	}
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive|checker.TypeFlagsTypeParameter) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if alias := proven.Alias(); alias != nil && alias.Symbol().Name == "Record" && l.librarySymbol(alias.Symbol()) && len(alias.TypeArguments()) == 2 && alias.TypeArguments()[1].Flags()&checker.TypeFlagsUnknown != 0 {
		return false
	}
	return l.recordElement(proven) != nil
}

// scalarRecordsInProgram reports whether any type the program writes is a record whose values are held as
// numbers or booleans, the storage the named dynamic read misreads. It reads every type node of the
// program's modules once. When the modules can't be read, it answers true.
func (l *lowering) scalarRecordsInProgram() bool {
	if l.scalarRecordsKnown {
		return l.scalarRecords
	}
	l.scalarRecordsKnown = true
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		l.scalarRecords = true
		return true
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if l.scalarRecords {
			return true
		}
		if ast.IsTypeNodeKind(node.Kind) {
			if element := l.recordElement(l.checker.GetTypeFromTypeNode(node)); element != nil {
				if held, known := l.representation(l.checker.GetNonNullableType(element)); !known || !held.IsReference() && held != ir.Union {
					l.scalarRecords = true
					return true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if l.scalarRecords {
			break
		}
	}
	return l.scalarRecords
}
