package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// keptCheckDowncastRefusal refuses in .a a downcast whose checks no proof erases (item 8 stage 2,
// docs/0.1.md): a tag test against a union member, an interface view's tag test, or an untagged view's
// checked reads. Each is a check the program asserts and the compiler can't prove, so .a writes it as a
// narrowing that panics, as it writes !, and .ts keeps it a checked cast. A construction, a node array, an
// upcast and a phantom brand are proofs castProof already erases, and a class downcast keeps its nominal
// identity test. castRefusal (1900) runs first, so a cast no check can hold, and stage 1's vacuous tag
// test, keep their own refusals. The angle-bracket form <Circle>shape is NotYet in stage 0 (conformance's
// refusals/angle_bracket_downcast.a); whoever teaches stage 0 to lower it takes it in here too.
func (l *lowering) keptCheckDowncastRefusal(node *ast.Node) error {
	if node.Kind != ast.KindAsExpression || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") || l.processPath(node.AsAsExpression().Expression) != "" {
		return nil
	}
	proof, err := l.castProof(node)
	if err != nil {
		return err
	}
	if len(proof.classes) > 0 || !proof.interfaceView && !proof.structuralView && len(proof.allowed) == 0 {
		return nil
	}
	source := l.checker.TypeToString(l.concrete(l.checker.GetTypeAtLocation(node.AsAsExpression().Expression)))
	target := l.checker.TypeToString(l.concrete(l.checker.GetTypeAtLocation(node)))
	refused := &Refused{Where: l.program.Where(node), What: "a downcast from " + source + " to " + target + " in .a whose checks no proof erases"}
	const ts = "or move the file to .ts, where the cast stays a checked cast"
	literal := l.fieldLiteral(l.concrete(l.checker.GetTypeAtLocation(node)), proof.field)
	switch {
	case len(proof.allowed) > 0:
		refused.Fix = "narrow it, as .a writes !: " + l.tagNarrowing(node, proof.field, proof.allowed) + ", " + ts
	case proof.interfaceView && literal != nil:
		refused.Fix = "declare " + source + " as a union with " + target + " as a member and narrow it, as .a writes !: " + l.tagNarrowing(node, proof.field, []*checker.Type{literal}) + ", " + ts
	default:
		refused.Fix = "give " + target + " a tag of its own in a discriminated union and narrow on it, as .a writes !, or build the value as a " + target + " so the compiler can prove it, " + ts
	}
	return refused
}

func init() {
	refusalRules.register(1950, "keptCheckDowncastRefusal", (*lowering).keptCheckDowncastRefusal)
}

// tagNarrowing is the statement that narrows a cast's operand to the members whose field holds one of
// allowed, written with the operand's own name when it's a reference a narrowing can name.
func (l *lowering) tagNarrowing(node *ast.Node, field string, allowed []*checker.Type) string {
	operand := ast.SkipParentheses(node.AsAsExpression().Expression)
	reference := ast.IsIdentifier(operand) || operand.Kind == ast.KindPropertyAccessExpression || operand.Kind == ast.KindThisKeyword
	held := "value"
	if reference {
		held = sourceExpression(operand)
	}
	tests := []string{}
	for _, literal := range allowed {
		tests = append(tests, held+"."+field+" !== "+l.checker.TypeToString(literal))
	}
	narrowing := "if (" + strings.Join(tests, " && ") + ") panic('why')"
	if !reference {
		return "hold it as const value, then " + narrowing
	}
	return narrowing
}
