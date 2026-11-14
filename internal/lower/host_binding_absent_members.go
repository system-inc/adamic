package lower

import (
	"maps"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// hostBindingProcessSyntax is the cheap gate for rule 70: only an identifier, or
// globalThis.process / globalThis["process"], can need a host lookup. It proves no binding identity.
func hostBindingProcessSyntax(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	if ast.IsIdentifier(node) {
		return node
	}
	var global *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if node.Name().Text() != "process" {
			return nil
		}
		global = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if (key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral) || key.Text() != "process" {
			return nil
		}
		global = node.AsElementAccessExpression().Expression
	default:
		return nil
	}
	global = ast.SkipParentheses(global)
	if ast.IsIdentifier(global) && global.Text() == "globalThis" {
		return node
	}
	return nil
}

// hostBindingProcessReceiver proves the binding after the syntactic gate. globalThis's process is
// the host's too; keep these writes closed independently of the namespace-value refusal (#zqz01vh).
func (l *lowering) hostBindingProcessReceiver(node *ast.Node) bool {
	node = hostBindingProcessSyntax(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		return l.processPath(node) == "process"
	}
	return l.globalObject(ast.SkipParentheses(node.Expression()))
}

// hostBindingProcessAbsent says whether a name is one of the members Node's process lacks
// (hostBindingAbsentMembers), each of which reads as undefined on every target and is never installed.
func hostBindingProcessAbsent(name string) bool {
	return hostBindingAbsentMembers["node:process.process"][name]
}

// hostBindingProcessAbsentNames spells every absent member of process, sorted, as "process.browser or
// process.recordreplay", for a stop that could reach any of them.
func hostBindingProcessAbsentNames() string {
	names := []string{}
	for _, name := range slices.Sorted(maps.Keys(hostBindingAbsentMembers["node:process.process"])) {
		names = append(names, "process."+name)
	}
	return strings.Join(names, " or ")
}

// hostBindingProcessAbsentViewed says whether a view of process names any member it lacks.
func (l *lowering) hostBindingProcessAbsentViewed(view *checker.Type) bool {
	for name := range hostBindingAbsentMembers["node:process.process"] {
		if l.checker.GetPropertyOfType(l.checker.GetNonNullableType(view), name) != nil {
			return true
		}
	}
	return false
}

// hostBindingAbsentKey says which absent member a key names, or whether it could name one at run time.
func hostBindingAbsentKey(node *ast.Node) (member string, runtime bool) {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if hostBindingProcessAbsent(node.Text()) {
			return node.Text(), false
		}
		return "", false
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return "", false
	}
	return "", true
}

// hostBindingAbsentAccess extracts only an access that names an absent member or has a run-time key.
func hostBindingAbsentAccess(node *ast.Node) (receiver *ast.Node, member string, runtime bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression && hostBindingProcessAbsent(node.Name().Text()) {
		return node.AsPropertyAccessExpression().Expression, node.Name().Text(), false
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if member, runtime := hostBindingAbsentKey(access.ArgumentExpression); member != "" || runtime {
			return access.Expression, member, runtime
		}
	}
	return nil, "", false
}

// hostBindingAbsentArgument is the absent member a reflective call's later arguments spell, as a key
// ('recordreplay') or an object literal's property ({ recordreplay: 1 }), so the call's stop names it.
// It's "" when they spell none, and the stop names every absent member.
func hostBindingAbsentArgument(arguments []*ast.Node) string {
	for _, argument := range arguments {
		argument = ast.SkipParentheses(argument)
		if member, _ := hostBindingAbsentKey(argument); member != "" {
			return member
		}
		if argument.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			if name := property.Name(); name != nil && (name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral) && hostBindingProcessAbsent(name.Text()) {
				return name.Text()
			}
		}
	}
	return ""
}

// hostBindingAbsentMemberRefusal keeps the constant reads of process's absent members exact
// (hostBindingAbsentMembers: browser and recordreplay): no program may install one, and run-time key
// operations never read an ordinary object slot. A structural view naming one is still unregistered
// (nodeHostViewReach), before its reads or writes lower. The undefined fold rests on two stops: this
// rule judges writes naming process, and the process-as-a-value stop blocks handing it elsewhere (a
// parameter, generic or array element). Lowering process as a value must bring this judgment to the
// receiving side, or the fold is no longer exact.
func (l *lowering) hostBindingAbsentMemberRefusal(node *ast.Node) error {
	// member names the absent member the stop is about, or is "" where the stop could reach any of them.
	stop := func(member string, runtime bool) error {
		if runtime {
			return l.notYet(node, "a run-time key on process, which could name "+hostBindingProcessAbsentNames())
		}
		if member == "" {
			return l.notYet(node, hostBindingProcessAbsentNames()+" outside a plain property read")
		}
		return l.notYet(node, "process."+member+" outside a plain property read")
	}
	if node.Kind == ast.KindDeleteExpression {
		if receiver, member, runtime := hostBindingAbsentAccess(node.AsDeleteExpression().Expression); l.hostBindingProcessReceiver(receiver) {
			return stop(member, runtime)
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression && hostBindingProcessAbsent(node.Name().Text()) {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if ast.IsAssignmentTarget(outer) || outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
			if l.hostBindingProcessReceiver(node.AsPropertyAccessExpression().Expression) {
				return stop(node.Name().Text(), false)
			}
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		if receiver, member, runtime := hostBindingAbsentAccess(node); l.hostBindingProcessReceiver(receiver) {
			return stop(member, runtime)
		}
	}
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword {
		binary := node.AsBinaryExpression()
		if member, runtime := hostBindingAbsentKey(binary.Left); (member != "" || runtime) && l.hostBindingProcessReceiver(binary.Right) {
			return stop(member, runtime)
		}
	}
	// A spread copies every member, so its stop names them all.
	if node.Kind == ast.KindSpreadAssignment && l.hostBindingProcessReceiver(node.AsSpreadAssignment().Expression) {
		return stop("", false)
	}
	if node.Kind == ast.KindVariableDeclaration && node.Initializer() != nil && l.hostBindingProcessReceiver(node.Initializer()) {
		if node.Name().Kind == ast.KindIdentifier {
			// Judge a later view before the generic process-as-a-value stop hides its member name.
			own := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.Initializer()))
			if row, _ := l.hostBindingOf(node.Initializer()); row != nil && l.hostBindingRouted(*row) {
				if model := l.hostBindingModel(*row); model != nil {
					own = model
				}
			}
			view := l.checker.GetTypeAtLocation(node.Name())
			if l.hostBindingProcessAbsentViewed(view) {
				if reached := l.nodeHostViewAt(own, view); reached != "" {
					return l.nodeHostViewRefusal(node.Initializer(), own, view, reached)
				}
			}
		}
		if node.Name().Kind == ast.KindObjectBindingPattern {
			for _, binding := range node.Name().AsBindingPattern().Elements.Nodes {
				name := binding.AsBindingElement().PropertyName
				if name == nil {
					name = binding.Name()
				}
				// Computed and nested bindings aren't plain reads of an absent member either, and could
				// reach any of them.
				if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral {
					return stop("", false)
				}
				if hostBindingProcessAbsent(name.Text()) {
					return stop(name.Text(), false)
				}
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || len(call.Arguments.Nodes) == 0 || !l.hostBindingProcessReceiver(call.Arguments.Nodes[0]) {
			return nil
		}
		access := callee.AsPropertyAccessExpression()
		if l.isLibraryGlobal(access.Expression, "Object") {
			switch access.Name().Text() {
			case "assign", "defineProperty", "defineProperties", "setPrototypeOf":
				return stop(hostBindingAbsentArgument(call.Arguments.Nodes[1:]), false)
			}
		}
		if l.isLibraryGlobal(access.Expression, "Reflect") {
			switch access.Name().Text() {
			case "get", "set", "defineProperty", "deleteProperty", "setPrototypeOf":
				return stop(hostBindingAbsentArgument(call.Arguments.Nodes[1:]), false)
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(70, "hostBindingAbsentMemberRefusal", (*lowering).hostBindingAbsentMemberRefusal)
}
