package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// fs.watch binds only the forms tsc's sys.ts uses (#j71fn59): fsWatchWorker's
// _fs.watch(path, supported ? {persistent: true, recursive: !!recursive} : {persistent: true}, callback)
// at :1779, presentWatcher.on("error", ...) at :1270, and close() through its FileWatcher. The
// result is an FSWatcher host object whose close and on are closures the runtime makes
// (node_fs_watch_impl.h). Every other form is a named NotYet.
const (
	nodeWatchName      = "node:fs.watch"
	nodeWatchCloseName = "node:fs.FSWatcher.close"
	nodeWatchOnName    = "node:fs.FSWatcher.on"
)

func (l *lowering) nodeWatch(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	if l.nodeLibraryMember(node) == nodeWatchName {
		value, err := l.nodeWatchCall(node)
		return value, true, err
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "on" && l.nodeHostOwnedMember(callee) == nodeWatchOnName {
		value, err := l.nodeWatchOn(node)
		return value, true, err
	}
	return nil, false, nil
}

func (l *lowering) nodeWatchCall(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, l.notYet(node, nodeWatchName+" through an optional call; use a direct call")
	}
	written := call.Arguments.Nodes
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(argument, nodeWatchName+" with spread arguments; write the arguments explicitly")
		}
	}
	if len(written) != 3 {
		return nil, l.notYet(node, nodeWatchName+" outside fs.watch(path, options, listener)")
	}
	path, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	if path.Type() != ir.String {
		return nil, l.notYet(written[0], nodeWatchName+" with a non-string path (Buffer and URL paths aren't bound yet)")
	}
	persistent, recursive, err := l.nodeWatchOptions(written[1])
	if err != nil {
		return nil, err
	}
	listener, err := l.nodeWatchCallback(written[2], nodeWatchName, 2)
	if err != nil {
		return nil, err
	}
	return ir.NodeFSFile{Operation: "host_watch", Arguments: []ir.Expression{path, persistent, recursive, listener}, Of: ir.Object}, nil
}

// nodeWatchOptions lowers tsc's options: a conditional of two literals chosen at run time from
// process.platform, or either literal alone. persistent is a boolean literal, the same in both arms,
// and recursive any boolean, false where an arm leaves it out. Only the chosen arm's recursive is
// evaluated, as JavaScript builds only the chosen literal.
func (l *lowering) nodeWatchOptions(node *ast.Node) (ir.Expression, ir.Expression, error) {
	options := ast.SkipParentheses(node)
	if options.Kind != ast.KindConditionalExpression {
		persistent, recursive, err := l.nodeWatchLiteral(options)
		return ir.BooleanConstant{Value: persistent}, recursive, err
	}
	conditional := options.AsConditionalExpression()
	condition, err := l.condition(conditional.Condition)
	if err != nil {
		return nil, nil, err
	}
	persistent, whenTrue, err := l.nodeWatchLiteral(conditional.WhenTrue)
	if err != nil {
		return nil, nil, err
	}
	otherwise, whenNot, err := l.nodeWatchLiteral(conditional.WhenFalse)
	if err != nil {
		return nil, nil, err
	}
	if persistent != otherwise {
		return nil, nil, l.notYet(options, nodeWatchName+" with options arms whose persistent differs")
	}
	return ir.BooleanConstant{Value: persistent}, ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot}, nil
}

func (l *lowering) nodeWatchLiteral(node *ast.Node) (bool, ir.Expression, error) {
	literal := ast.SkipParentheses(node)
	if literal.Kind != ast.KindObjectLiteralExpression {
		return false, nil, l.notYet(literal, nodeWatchName+" with options other than a {persistent} or {persistent, recursive} literal, or a conditional of two")
	}
	var persistent *bool
	var recursive ir.Expression
	for _, field := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment || field.Name() == nil || field.Name().Kind != ast.KindIdentifier {
			return false, nil, l.notYet(field, nodeWatchName+" with options other than plain persistent and recursive fields")
		}
		value := field.AsPropertyAssignment().Initializer
		switch name := field.Name().Text(); {
		case name == "persistent" && persistent == nil:
			given := ast.SkipParentheses(value)
			if given.Kind != ast.KindTrueKeyword && given.Kind != ast.KindFalseKeyword {
				return false, nil, l.notYet(value, nodeWatchName+" with a persistent other than true or false")
			}
			held := given.Kind == ast.KindTrueKeyword
			persistent = &held
		case name == "recursive" && recursive == nil:
			lowered, err := l.expression(value)
			if err != nil {
				return false, nil, err
			}
			if lowered.Type() != ir.Boolean {
				return false, nil, l.notYet(value, nodeWatchName+" with a non-boolean recursive")
			}
			recursive = lowered
		default:
			return false, nil, l.notYet(field, nodeWatchName+" with the option "+name+" (only persistent and recursive are bound)")
		}
	}
	if persistent == nil {
		return false, nil, l.notYet(literal, nodeWatchName+" without a literal persistent")
	}
	if recursive == nil {
		recursive = ir.BooleanConstant{Value: false}
	}
	return *persistent, recursive, nil
}

// nodeWatchCallback admits any function value Node can call as a watcher's listener: a parameter,
// an arrow or a declaration. The runtime hands it Node's own arguments, handed of them (the event
// and the file name for a change, the Error for 'error'), and undefined past them, so each
// parameter must hold what it's handed as a reference: a string or union where a string goes, an
// object or union where the Error goes, and undefined as NULL past Node's count.
func (l *lowering) nodeWatchCallback(node *ast.Node, member string, handed int) (ir.Expression, error) {
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, l.notYet(node, member+" with a listener that isn't one plain function")
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.ThisParameter() != nil || signature.HasRestParameter() {
		return nil, l.notYet(node, member+" with a generic, this-typed or rest-parameter listener")
	}
	parameters := signature.Parameters()
	if len(parameters) > 3 {
		return nil, l.notYet(node, member+" with a listener of more than three parameters")
	}
	for index, parameter := range parameters {
		of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		admitted := known && of.IsReference() && of != ir.Weak && of != ir.Record
		switch {
		case index >= handed:
		case handed == 2:
			admitted = admitted && (of == ir.String || of == ir.Union)
		default:
			admitted = admitted && (of == ir.Object || of == ir.Union)
		}
		if !admitted {
			return nil, l.notYet(node, member+" with a listener whose parameter "+parameter.Name+" can't hold what Node hands it")
		}
	}
	listener, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if listener.Type() != ir.Closure {
		return nil, l.notYet(node, member+" with a non-closure listener")
	}
	return listener, nil
}

// nodeWatchOn lowers watcher.on("error", listener) on an FSWatcher, its receiver's own type, which is
// how it's told from the same inherited declaration on every other host type (Library's must on
// #j71fn59). Node returns the watcher, so the call gives its receiver back.
func (l *lowering) nodeWatchOn(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, l.notYet(node, nodeWatchOnName+" through an optional chain; call it directly")
	}
	written := call.Arguments.Nodes
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(argument, nodeWatchOnName+" with spread arguments; write the arguments explicitly")
		}
	}
	if len(written) != 2 {
		return nil, l.notYet(node, nodeWatchOnName+" outside on(event, listener)")
	}
	event := ast.SkipParentheses(written[0])
	if event.Kind != ast.KindStringLiteral && event.Kind != ast.KindNoSubstitutionTemplateLiteral || event.Text() != "error" {
		return nil, l.notYet(written[0], nodeWatchOnName+` with an event other than the literal "error"`)
	}
	receiver, err := l.expression(ast.SkipParentheses(call.Expression).AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	if receiver.Type() != ir.Object {
		return nil, l.notYet(node, nodeWatchOnName+" on a receiver that isn't held as an object")
	}
	listener, err := l.nodeWatchCallback(written[1], nodeWatchOnName, 1)
	if err != nil {
		return nil, err
	}
	return ir.NodeFSFile{Operation: "host_watch_on", Arguments: []ir.Expression{receiver, listener}, Of: ir.Object}, nil
}

// nodeWatchUseRefusal is the refusal visitor's view of an FSWatcher's close and on: called through
// the watcher, the call, its property access and the property's name are admitted, and anywhere
// else they're NotYet. Node's are prototype methods, so a detached one runs without its watcher.
func (l *lowering) nodeWatchUseRefusal(node *ast.Node) (bool, error) {
	access := ast.SkipParentheses(node)
	if access.Kind == ast.KindCallExpression {
		access = ast.SkipParentheses(access.AsCallExpression().Expression)
	}
	if access.Parent != nil && access.Parent.Kind == ast.KindPropertyAccessExpression && access.Parent.Name() == access {
		access = access.Parent
	}
	// Most nodes are no read of a close or an on; only those ask the checker for a receiver's type.
	if access.Kind != ast.KindPropertyAccessExpression || access.Name().Text() != "close" && access.Name().Text() != "on" {
		return false, nil
	}
	member := l.nodeHostOwnedMember(access)
	if member != nodeWatchCloseName && member != nodeWatchOnName {
		return false, nil
	}
	if node.Kind == ast.KindCallExpression {
		return true, nil
	}
	outer := access
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindCallExpression && ast.SkipParentheses(outer.Parent.AsCallExpression().Expression) == access {
		return true, nil
	}
	return true, l.notYet(access, member+" used other than called on its watcher")
}

func init() { RegisterNodeLibraryMembers(nodeWatchName, nodeWatchCloseName, nodeWatchOnName) }
