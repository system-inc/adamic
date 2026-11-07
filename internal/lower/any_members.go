package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A member a program declares as any holds whatever was stored in it, so every read of it through a
// typed view is unchecked (#tmdhz1d). #bw3xg7c rules any refused, written and flowing inferred alike, so a
// property the program declares (an interface's or a type literal's property signature, a class's field)
// whose type is any is refused where it's declared, written (text: any, or an alias of any) or inferred
// from its initializer (x = JSON.parse(source)). A member of a lib or prelude declaration isn't the
// program's, so it's left to the relation rule (refuseAnyRelation) and the axiom audit. A member whose
// any is nested (items: any[]) is a relation's to judge where it's read as a typed member, which
// refuseAnyRelation does.
func (lowering *lowering) refuseAnyMember(node *ast.Node) error {
	if node.Kind != ast.KindPropertySignature && node.Kind != ast.KindPropertyDeclaration {
		return nil
	}
	symbol := lowering.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return nil
	}
	if lowering.checker.GetTypeOfSymbol(symbol).Flags()&checker.TypeFlagsAny == 0 {
		return nil
	}
	written := "declared"
	if node.Type() == nil {
		written = "inferred from its initializer"
	}
	return &Refused{
		Where: lowering.program.Where(node),
		What: "a member typed any (" +
			scannedMemberName(node) +
			", " +
			written +
			"): any has no proven type, so every read of it through a typed view is unchecked",
		Fix: "type it unknown and narrow it where it's read, or give it its real type",
	}
}

// scannedMemberName is a member's name as written.
func scannedMemberName(node *ast.Node) string {
	if name := node.Name(); name != nil {
		if ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name) {
			return name.Text()
		}
	}
	return "a computed name"
}

func init() { refusalRules.register(3280, "refuseAnyMember", (*lowering).refuseAnyMember) }
