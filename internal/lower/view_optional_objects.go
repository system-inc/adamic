package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An optional own data slot supplies either its object or undefined. Its shape
// retains presence; reading it never invents or removes an own property. The
// recursively registered child fields keep their checks through saved aliases.
func (l *lowering) optionalObjectViewRead(node *ast.Node, field *ast.Symbol) bool {
	if field.Flags&ast.SymbolFlagsOptional == 0 || accessorSymbol(field) || node.AsPropertyAccessExpression().QuestionDotToken != nil {
		return false
	}
	declared := l.checker.GetTypeOfSymbol(field)
	if declared.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	objects, undefined := 0, false
	for _, member := range declared.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			undefined = true
			continue
		}
		of, known := l.representation(member)
		if !known || of != ir.Object || member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) != 0 {
			return false
		}
		objects++
	}
	return undefined && objects == 1
}

// Lazy family failures can leave a placeholder without its field graph. Such a
// placeholder cannot certify an optional object alias. Refuse its demanded read
// rather than let the missing graph erase the child's checks.
func (l *lowering) optionalObjectViewBoundary(node *ast.Node, target *checker.Type) error {
	id := l.result.ViewContractTypes[int(target.Id())]
	if id <= 0 {
		return nil
	}
	root := l.result.ViewContracts[id-1]
	if root.Kind != ir.ViewUnknown && root.Unsupported == "" {
		return nil
	}
	family := root.Unsupported
	if family == "" {
		family = "object view"
	}
	names := map[string]bool{}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if field.Flags&ast.SymbolFlagsOptional != 0 {
			names[field.Name] = true
		}
	}
	if len(names) == 0 {
		return nil
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		if part.Kind == ast.KindPropertyAccessExpression && names[part.Name().Text()] {
			field := l.checker.GetSymbolAtLocation(part.Name())
			if field != nil && l.optionalObjectViewRead(part, field) {
				refused = &Refused{Where: l.program.Where(part), What: "an optional object field read " + sourceExpression(part) + " without a complete " + l.checker.TypeToString(target) + " view", Fix: "prove or implement the " + family + " contract before reading this field through its checked child view"}
				return true
			}
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if refused != nil {
			return refused
		}
	}
	return nil
}
