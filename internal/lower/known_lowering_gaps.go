package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Stop known wrong answers before either backend receives an IR program. These
// are implementation gaps, not new language exclusions.
func (l *lowering) knownLoweringGap(node *ast.Node) error {
	switch node.Kind {
	case ast.KindSpreadAssignment:
		source := node.AsSpreadAssignment().Expression
		if l.errorStorage(l.checker.GetTypeAtLocation(source), map[*checker.Type]bool{}) {
			return l.notYet(node, "object spread of Error storage with non-enumerable properties; copy the desired fields explicitly")
		}
	case ast.KindSetAccessor:
		parameters := node.AsSetAccessorDeclaration().Parameters.Nodes
		if len(parameters) == 1 {
			if of, known := l.representation(l.checker.GetTypeAtLocation(parameters[0])); known && of == ir.MaybeNumber {
				return l.notYet(node, "a setter with an optional numeric input; use a plain optional numeric field until accessor thunks support two-word arguments")
			}
		}
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindPropertyDeclaration, ast.KindPropertySignature:
		if len(plainPropertyName(node)) > 4095 {
			return l.notYet(node.Name(), "a property name longer than 4095 bytes; shorten the name until native metadata supports long strings")
		}
	case ast.KindClassDeclaration:
		for _, clause := range nodesOf(node.AsClassDeclaration().HeritageClauses) {
			if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
				continue
			}
			for _, member := range node.Members() {
				if member.Name() != nil && member.Name().Kind == ast.KindComputedPropertyName {
					return l.notYet(member.Name(), "a computed member name in a derived class; use a named member until override lookup supports computed names")
				}
			}
		}
	case ast.KindThisKeyword:
		var arrow *ast.Node
		for parent := node.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindArrowFunction {
				arrow = parent
				continue
			}
			if parent.Kind == ast.KindConstructor {
				if arrow == nil {
					return nil
				}
				// The existing readiness boundary counts direct unconditional assignments.
				// A closure created earlier captures this even when its only use is a field read.
				for _, member := range parent.Parent.Members() {
					if member.Kind == ast.KindPropertyDeclaration && member.Name().Kind == ast.KindComputedPropertyName {
						return l.notYet(member.Name(), "constructor readiness with a computed field; use a named field")
					}
				}
				if arrow.Pos() < lastFieldAssignment(parent.Parent, parent) {
					return &Refused{Where: l.program.Where(node), What: "this escaping a constructor through a closure before every field is set", Fix: "assign every field first, then create the closure that reads this"}
				}
				return nil
			}
			if ast.IsFunctionLike(parent) {
				return nil
			}
		}
	}
	return nil
}

// Error and its subclasses store non-enumerable own fields. Their structural
// name/message types do not prove that copying runtime storage is a JS spread.
func (l *lowering) errorStorage(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	if l.isLibraryType(proven, "Error", "EvalError", "RangeError", "ReferenceError", "SyntaxError", "TypeError", "URIError", "AggregateError") {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.errorStorage(member, seen) {
				return true
			}
		}
	}
	if isClassInstance(proven) {
		for _, base := range l.classBases(proven) {
			if l.errorStorage(base, seen) {
				return true
			}
		}
	}
	return false
}
