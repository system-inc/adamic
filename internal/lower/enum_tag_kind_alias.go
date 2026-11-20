package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// #s8gy89r U7 applies #00rd54t to tsc's checkExpressionWorker, which saves `const kind = node.kind;`
// and then switches on kind. The checker narrows node by that switch as it does by
// `switch (node.kind)`: getCandidateDiscriminantPropertyAccess takes a const's initializer as the
// access. The walk reads such a switch as the field the alias saved, judged by the same hazard and
// accessor checks as a direct switch on it, only where the saved value is still the reference's own
// field when the switch runs. Rule 3350 still judges every read, and an if over the alias is unchanged.

// switchField is the field a switch statement's expression reads off the reference: directField's,
// or a const alias's (kindAliasField).
func (w *enumTagWalk) switchField(statement *ast.Node) string {
	if field := w.directField(statement.AsSwitchStatement().Expression); field != "" {
		return field
	}
	return w.kindAliasField(statement)
}

// kindAliasField is f when statement switches on a name that is a const alias of the reference's
// ref.f, and "" otherwise. Every condition must hold:
//   - the name resolves to one const declaration (not let, var or using), unannotated and naming one
//     identifier, as the checker asks, whose initializer is a field read directly off the reference
//     that fieldRead accepts, so no optional chain, element access or accessor;
//   - its statement runs first, in the same function (kindAliasRunsFirst);
//   - the reference is a plain name its function never writes from the alias's end to the switch's
//     start (kindAliasRewritten).
//
// kindAliasRewritten guards the reference's binding, never the saved field: a write to the tag through
// another name between the alias and the switch would leave the alias stale where a direct switch rereads
// it. The road stands on a refusal elsewhere: each member's tag is one literal, so a write that changes it
// needs a writable view whose tag is wider: over a readonly tag that's a readonly field made writable,
// refused at the call (family 4), and over a mutable one a wider view that can write, family 3's. If
// either rule is ever narrowed, this road opens, and
// TestKindAliasTagWriteThroughWiderViewIsRefused is where it shows first.
func (w *enumTagWalk) kindAliasField(statement *ast.Node) string {
	name := ast.SkipParentheses(statement.AsSwitchStatement().Expression)
	if name.Kind != ast.KindIdentifier || w.reference.Kind != ast.KindIdentifier {
		return ""
	}
	symbol := w.l.symbol(name)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return ""
	}
	declaration := symbol.Declarations[0]
	list := declaration.Parent
	if declaration.Kind != ast.KindVariableDeclaration || list == nil || list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst {
		return ""
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Name().Kind != ast.KindIdentifier || variable.Type != nil || variable.Initializer == nil || ast.SkipParentheses(variable.Initializer).Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	if list.Parent == nil || list.Parent.Kind != ast.KindVariableStatement || !kindAliasRunsFirst(list.Parent, statement) {
		return ""
	}
	if w.kindAliasRewritten(list.Parent.Parent, declaration.End(), statement.Pos()) {
		return ""
	}
	return w.fieldRead(variable.Initializer)
}

// kindAliasRunsFirst is whether alias, a variable statement, is an earlier statement of a list on the
// switch's own road out to its function, with only blocks, ifs, switches, case clauses, labels, try
// and catch between: it has run, once, whenever the switch runs. A function or a class between runs
// the switch at another time, and a loop could run it again after a write later in its body.
func kindAliasRunsFirst(alias, statement *ast.Node) bool {
	if alias.End() > statement.Pos() {
		return false
	}
	for node := statement.Parent; node != nil; node = node.Parent {
		if node == alias.Parent {
			return true
		}
		switch node.Kind {
		case ast.KindBlock, ast.KindIfStatement, ast.KindSwitchStatement, ast.KindCaseBlock, ast.KindCaseClause, ast.KindDefaultClause, ast.KindLabeledStatement, ast.KindTryStatement, ast.KindCatchClause:
		default:
			return false
		}
	}
	return false
}

// kindAliasRewritten is whether the reference's binding may be written in container from from to to.
// Each name there that is the binding must be a plain read: never an assignment's target (plain,
// compound or destructuring), a ++ or --, or a declaration. A nested function there counts too, and
// arguments or eval, which can write a parameter in sloppy code, fail closed.
func (w *enumTagWalk) kindAliasRewritten(container *ast.Node, from, to int) bool {
	binding := w.l.valueSymbol(w.reference)
	if container == nil || binding == nil {
		return true
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if written || node.End() <= from || node.Pos() >= to {
			return written
		}
		if node.Kind == ast.KindIdentifier {
			if node.Text() == "arguments" || node.Text() == "eval" {
				written = true
				return true
			}
			shorthand := node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment
			if w.l.valueSymbol(node) == binding && (ast.IsAssignmentTarget(node) || (ast.IsDeclarationName(node) && !shorthand)) {
				written = true
				return true
			}
		}
		node.ForEachChild(visit)
		return written
	}
	container.ForEachChild(visit)
	return written
}
