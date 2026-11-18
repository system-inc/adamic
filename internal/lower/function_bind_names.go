package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// A name is usable only when every way of storing it is readable. Empty and unreadable are
// different: an absent optional field has no functions, while an unreadable write keeps the
// type-wide count. The allocation flow's known answer always takes precedence over this scan.
type boundNameFunctions struct {
	functions  []int
	unreadable bool
}

func boundMemberTarget(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	target := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if target.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	return target
}

// Build once per lowering, alongside boundValueReach. Writes of the same name elsewhere only
// add candidates, whether or not fresh proved them acyclic. Memory's writtenFunctionCells walk
// (ec73cab9) supplies the assignment reading; literal initializations and methods complete it.
func (l *lowering) boundNameReach() func(*ast.Node) ([]int, bool) {
	names := map[string]*boundNameFunctions{}
	for _, record := range l.closureRecords {
		if l.result.Functions[record.function].Source.Role == "bound_function" {
			if target := boundMemberTarget(record.node); target != nil {
				names[target.Name().Text()] = &boundNameFunctions{}
			}
		}
	}
	unknown := false
	for _, write := range fresh.ProveWrites(l.result) {
		if write.Kind == fresh.WriteUnknown {
			unknown = true
		}
		entry := names[write.Name]
		if write.Kind != fresh.WriteField || entry == nil {
			continue
		}
		if write.Site <= 0 || write.Site > len(l.writeSites) {
			entry.unreadable = true
			continue
		}
		functions, readable := l.boundAssignedFunction(l.writeSites[write.Site-1].node)
		entry.functions = append(entry.functions, functions...)
		entry.unreadable = entry.unreadable || !readable
	}
	visit := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok {
			for _, field := range literal.Fields {
				if entry := names[field.Name]; entry != nil {
					switch value := field.Value.(type) {
					case ir.MakeClosure:
						entry.functions = append(entry.functions, value.Function)
					case ir.BoundClosure:
						entry.functions = append(entry.functions, value.Function)
					case ir.Undefined, ir.Null:
					default:
						entry.unreadable = true
					}
				}
			}
		}
		return true
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	// Whole-object readers, including bind's receiver lookup, do not write a copy. Actual copies
	// retain the copied type even when their IR no longer spells out the inherited fields.
	for _, copied := range l.wholeObjectCopies {
		if copied.reader == "" {
			for name, entry := range names {
				if l.boundReceiverHolds(copied.copied, name) {
					entry.unreadable = true
				}
			}
		}
	}
	var scan ast.Visitor
	scan = func(node *ast.Node) bool {
		var receiver, key *ast.Node
		if node.Kind == ast.KindElementAccessExpression && ast.IsAssignmentTarget(node) {
			receiver, key = node.AsElementAccessExpression().Expression, node.AsElementAccessExpression().ArgumentExpression
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && call.Arguments != nil {
				host, member := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
				if len(call.Arguments.Nodes) >= 2 && ((member == "defineProperty" && l.isLibraryGlobal(host, "Object")) || (member == "set" && l.isLibraryGlobal(host, "Reflect"))) {
					receiver, key = call.Arguments.Nodes[0], call.Arguments.Nodes[1]
				}
				if member == "assign" && l.isLibraryGlobal(host, "Object") {
					for name, entry := range names {
						if l.boundReceiverHolds(l.checker.GetTypeAtLocation(node), name) {
							entry.unreadable = true
						}
					}
				}
			}
		}
		if receiver != nil {
			keyType := l.checker.GetTypeAtLocation(key)
			// Finite keys already produce named SetProperty writes (LogicalAssignment.KeyNames).
			// Other computed writes must be charged only to the shaped members they can hold.
			if !l.boundFiniteKey(keyType) {
				for name, entry := range names {
					if l.boundReceiverHolds(l.checker.GetTypeAtLocation(receiver), name) {
						entry.unreadable = true
					}
				}
			}
		}
		node.ForEachChild(scan)
		return false
	}
	// Every module the entry's lowering lowers, in moduleOrder's order, never load's Files(), which
	// holds only the roots: a defineProperty in an imported module writes the name as surely as one
	// in the entry (Lower's must on #x909e6q, the same shape as readsStdoutHandle's on #0bxc4jh).
	// Where moduleOrder refused, the program never lowers, and the answer fails closed.
	modules, err := l.moduleOrder(l.entry)
	if err != nil {
		unknown = true
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(scan)
	}
	for name, entry := range names {
		entry.functions = append(entry.functions, l.result.MethodsNamed(name)...)
	}
	return func(target *ast.Node) ([]int, bool) {
		if target == nil || unknown {
			return nil, false
		}
		entry := names[target.Name().Text()]
		if entry == nil || entry.unreadable || l.boundMemberUnreadable(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression), target.Name().Text()) {
			return nil, false
		}
		return entry.functions, true
	}
}

// A plain assignment's site is its receiver. Neither an alias holding a closure nor a call's
// result is read as a function here: losing either's possibilities would fail open.
func (l *lowering) boundAssignedFunction(receiver *ast.Node) ([]int, bool) {
	if receiver == nil || receiver.Parent == nil {
		return nil, false
	}
	access := receiver.Parent
	switch access.Kind {
	case ast.KindPropertyAccessExpression:
		if access.AsPropertyAccessExpression().Expression != receiver {
			return nil, false
		}
	case ast.KindElementAccessExpression:
		if access.AsElementAccessExpression().Expression != receiver {
			return nil, false
		}
	default:
		return nil, false
	}
	assignment := access.Parent
	if assignment == nil || assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken || assignment.AsBinaryExpression().Left != access {
		return nil, false
	}
	written := ast.SkipParentheses(assignment.AsBinaryExpression().Right)
	if written.Kind == ast.KindNullKeyword {
		return nil, true
	}
	if ast.IsIdentifier(written) && written.Text() == "undefined" {
		symbol := l.symbol(written)
		if symbol == nil || len(symbol.Declarations) == 0 || l.isLibraryGlobal(written, "undefined") {
			return nil, true
		}
	}
	target := written
	switch written.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
	case ast.KindIdentifier:
		symbol := l.symbol(written)
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
			return nil, false
		}
		target = symbol.Declarations[0]
	default:
		return nil, false
	}
	functions := []int{}
	for _, closure := range l.closureRecords {
		if closure.node == target {
			functions = append(functions, closure.function)
		}
	}
	return functions, len(functions) != 0
}

func (l *lowering) boundFiniteKey(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.boundFiniteKey(l.checker.GetBaseConstraintOfType(proven))
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.boundFiniteKey(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) != 0
}

// An index-signature view writes record storage, never a shaped object's field. A union may
// contain a shaped receiver as well; a type parameter inherits its constraint, or any without one.
func (l *lowering) boundReceiverHolds(proven *checker.Type, name string) bool {
	if proven == nil || proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.boundReceiverHolds(l.checker.GetBaseConstraintOfType(proven), name)
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.boundReceiverHolds(member, name) {
				return true
			}
		}
		return false
	}
	return len(l.checker.GetIndexInfosOfType(proven)) == 0 && l.checker.GetPropertyOfType(proven, name) != nil
}

// Record literals and indexed writes do not enter the shaped-field name table. A bind through
// that storage keeps its type-wide candidates, as does a member the host can fill from a .d.ts.
func (l *lowering) boundMemberUnreadable(proven *checker.Type, name string) bool {
	if proven == nil || proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.boundMemberUnreadable(l.checker.GetBaseConstraintOfType(proven), name)
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.boundMemberUnreadable(member, name) {
				return true
			}
		}
		return false
	}
	if len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return true
	}
	property := l.checker.GetPropertyOfType(proven, name)
	if property == nil {
		return true
	}
	for _, declaration := range property.Declarations {
		if ast.GetSourceFileOfNode(declaration).IsDeclarationFile {
			return true
		}
	}
	return false
}
