package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// stdoutHandleMember is the one member the runtime's stdout handle holds (node_process.c's
// {setBlocking} shape). @types declares no _handle, so the type a program reads it as is always the
// program's own cast, and that cast's other members aren't on the object: a cast naming readonly fd
// read fd as undefined natively where Node gives the descriptor (#0asf98m).
const stdoutHandleMember = "setBlocking"

// stdoutHandleView judges every type process.stdout._handle is seen as on its way out of the
// expression that reads it: the read's own type, each enclosing as, satisfies or ! result, and the
// declared type of a variable it initializes. Each must be undefined, null, or an object type whose
// only member is setBlocking, with no signature or index. any passes only where a later cast in the
// same chain names the type (node_process_terminal.a's `as any` then `as { setBlocking?: ... }`), so
// a handle that leaves the chain as any stops. Past the chain, every later view is a relation from
// that type, which the hiding rule judges. It's "" when every view holds, or the stop otherwise.
func (l *lowering) stdoutHandleView(node *ast.Node) string {
	seen := []*checker.Type{l.checker.GetTypeAtLocation(node)}
	outer := node
	for outer.Parent != nil {
		parent := outer.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
			seen = append(seen, l.checker.GetTypeAtLocation(parent))
		default:
			parent = nil
		}
		if parent == nil {
			break
		}
		outer = parent
	}
	if parent := outer.Parent; parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == outer && parent.AsVariableDeclaration().Type != nil {
		seen = append(seen, l.checker.GetTypeAtLocation(parent.Name()))
	}
	for index, view := range seen {
		if view.Flags()&checker.TypeFlagsAny != 0 {
			if index < len(seen)-1 {
				continue
			}
			return "process.stdout._handle seen as any; cast it to a type whose only member is setBlocking"
		}
		if extra := l.stdoutHandleExtra(view); extra != "" {
			return "process.stdout._handle seen as " + l.checker.TypeToString(view) + ", which names " + extra + " where the runtime's stdout handle holds only setBlocking"
		}
	}
	return ""
}

// stdoutHandleExtra names what view declares beyond the stdout handle's one member, or "" when it
// declares nothing more. A type it can't read member by member is named whole, which fails closed.
func (l *lowering) stdoutHandleExtra(view *checker.Type) string {
	if view.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range view.Types() {
			if extra := l.stdoutHandleExtra(member); extra != "" {
				return extra
			}
		}
		return ""
	}
	if view.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
		return ""
	}
	if view.Flags()&checker.TypeFlagsObject == 0 {
		return l.checker.TypeToString(view)
	}
	if len(l.checker.GetSignaturesOfType(view, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(view, checker.SignatureKindConstruct)) != 0 {
		return "a call or construct signature"
	}
	if len(l.checker.GetIndexInfosOfType(view)) != 0 {
		return "an index signature"
	}
	for _, property := range l.checker.GetPropertiesOfType(view) {
		if property.Name != stdoutHandleMember {
			return property.Name
		}
	}
	return ""
}
