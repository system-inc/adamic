package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A record's required declared property can't be deleted out from under its type (#p9jj9jg). tsc
// refuses delete options.required by name (TS2790), but checks delete options[key] against the index
// signature alone when key isn't one literal, so interface Options { target?: number; required:
// number; [option: string]: number | undefined } lets a string key that holds "required" remove it,
// and a later options.required is typed number while it holds nothing: Node prints NaN for
// options.required + 1. The key is judged by the names it can be (writtenNames): any name, or a
// union's or a folded constant's, any of which a required declared property of the receiver takes.
// In .a it's refused; in .ts, which runs the programs tsc accepts, it's not yet, since the read after
// it would stop where Node goes on. A literal key is tsc's to judge, and a key whose names are all
// optional or undeclared removes nothing the type promises.

// declaredDeleteFix is what the refusal offers.
const declaredDeleteFix = "delete an optional property by its name, or give the record no required keys"

// recordDeclaredDeleteRefusal refuses, in .a, and stops not yet, in .ts, a delete through a key that
// isn't one literal whose receiver, or a member of its union, declares a required property under a
// name the key can be. It runs after deleteRefusal (1300), which refuses a delete of anything but a
// record entry, and before the delete is lowered as a record's.
func (lowering *lowering) recordDeclaredDeleteRefusal(node *ast.Node) error {
	if node.Kind != ast.KindDeleteExpression {
		return nil
	}
	operand := ast.SkipParentheses(node.AsDeleteExpression().Expression)
	if operand.Kind != ast.KindElementAccessExpression {
		return nil
	}
	access := operand.AsElementAccessExpression()
	names, single := lowering.writtenNames(access.ArgumentExpression)
	if single {
		return nil
	}
	receiver := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression))
	for _, member := range lowering.definedMembers(receiver) {
		for _, property := range lowering.checker.GetPropertiesOfType(member) {
			if property.Flags&ast.SymbolFlagsOptional != 0 || names != nil && !names[property.Name] {
				continue
			}
			what := "a delete from a record through a key that is not one literal, which can remove a required declared property (" +
				sourceExpression(node) +
				" on " +
				lowering.checker.TypeToString(member) +
				" by " +
				sourceExpression(access.ArgumentExpression) +
				": " +
				property.Name +
				" is declared " +
				lowering.checker.TypeToString(lowering.checker.GetTypeOfSymbol(property)) +
				", not optional)"
			if !strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
				return lowering.notYet(node, what)
			}
			return &Refused{Where: lowering.program.Where(node), What: what, Fix: declaredDeleteFix}
		}
	}
	return nil
}

func init() {
	refusalRules.register(1350, "recordDeclaredDeleteRefusal", (*lowering).recordDeclaredDeleteRefusal)
}
