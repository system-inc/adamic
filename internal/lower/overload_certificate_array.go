package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Certificate A (Adamic's ruling on #5fws6q8): an overload whose result is narrower than its
// implementation's is served when the implementation returns a fresh array it built only by pushing,
// in one for-of over an iterable parameter, each element or the element mapped by an optional mapper
// parameter the bare `map ?` test selects. tsc's core.ts arrayFrom is the shape:
//
//	const result: (T | U)[] = [];
//	for (const value of iterator) {
//	    result.push(map ? map(value) : value);
//	}
//	return result;
//
// Under an overload with no mapper the array holds only elements, and under one whose mapper is required
// only mapped values. That reading takes the overload's arguments as exactly what the caller passed,
// which holds only at a direct call the checker resolved to that overload. So the certificate erases the
// result's check per call, never per declaration: overloadCertificateCall stops every other entry, and an
// array result can't be checked at a call, so each one is NotYet there. The certificate is a fact about
// values, never about how they're held: a direct call is erased only where the implementation's instance
// at that call stores its result as the call reads it (overloadCertificateStorage), since arrayFrom's
// instance at T = string, U = number builds a (string | number)[] of boxed unions the call would read as
// number[] (Lower's probe J-h37mw8d on #5fws6q8). Certificate B (overload_certificate_map.go) reads its own
// shape and shares this file's frame, confinement and gate.

// overloadArrayShape is what certificate A reads from an implementation's body.
type overloadArrayShape struct {
	// local is the fresh array the body returns, and element the for-of's const binding.
	local, element *ast.Symbol
	// iterable is the index of the parameter the loop reads, and mapper of the parameter the bare test
	// selects, or -1 when no push maps.
	iterable, mapper int
	// bare is whether some push is the element itself, and mapped whether some push is
	// map ? map(element) : element.
	bare, mapped bool
}

// overloadArrayShapeOf is the shape of implementation's body when it's certificate A's: a const local
// initialized to an empty array literal first, its bare return last, and between them exactly one
// for-of over a parameter, with const element binding, whose body is only pushes onto the local. Other
// statements between may run, since nothing they do can reach the local: its only references are the
// pushes' receivers and the return, nothing in the body writes a parameter, and no return but the last
// leaves.
func (l *lowering) overloadArrayShapeOf(implementation *ast.Node) (overloadArrayShape, bool) {
	shape := overloadArrayShape{mapper: -1}
	if !overloadCertificateFunction(implementation) {
		return shape, false
	}
	parameters := implementation.Parameters()
	for _, parameter := range parameters {
		declaration := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declaration.Initializer != nil || declaration.DotDotDotToken != nil {
			// A default runs where an omitted or undefined mapper is passed, so absence isn't the bare
			// test's false branch (array-default).
			return shape, false
		}
	}
	statements := implementation.Body().AsBlock().Statements.Nodes

	// The fresh array: a const, so nothing rebinds it, and an empty literal, so it starts holding nothing.
	name, last := l.overloadFreshLocal(statements, func(initializer *ast.Node) bool {
		return initializer.Kind == ast.KindArrayLiteralExpression && len(initializer.AsArrayLiteralExpression().Elements.Nodes) == 0
	})
	if name == nil {
		return shape, false
	}
	shape.local = l.symbol(name)

	// The one loop, over a parameter as the caller passed it: a const element binding nothing rebinds.
	loop := overloadOnlyLoop(statements, ast.KindForOfStatement)
	if loop == nil {
		return shape, false
	}
	iteration := loop.AsForInOrOfStatement()
	shape.iterable = l.overloadParameterIndex(parameters, iteration.Expression)
	if iteration.AwaitModifier != nil || shape.iterable < 0 || iteration.Initializer == nil || iteration.Initializer.Kind != ast.KindVariableDeclarationList {
		return shape, false
	}
	bindings := iteration.Initializer.AsVariableDeclarationList().Declarations.Nodes
	if iteration.Initializer.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst || len(bindings) != 1 || !ast.IsIdentifier(bindings[0].Name()) {
		return shape, false
	}
	shape.element = l.symbol(bindings[0].Name())
	element := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return ast.IsIdentifier(node) && l.symbol(node) == shape.element
	}

	// Each push: the Array prototype's own push onto the local, of the element, or of the element mapped
	// by the parameter its bare test selects. Nothing else is read, since a property read can run a
	// getter that answers differently each time (array-getter).
	pushes := []*ast.Node{iteration.Statement}
	if iteration.Statement.Kind == ast.KindBlock {
		pushes = iteration.Statement.AsBlock().Statements.Nodes
	}
	receivers := map[*ast.Node]bool{}
	for _, push := range pushes {
		receiver, arguments := l.overloadMutatorCall(push, shape.local, "Array", "push", 1)
		if receiver == nil {
			return shape, false
		}
		receivers[receiver] = true
		argument := ast.SkipParentheses(arguments[0])
		if element(argument) {
			shape.bare = true
			continue
		}
		if argument.Kind != ast.KindConditionalExpression {
			return shape, false
		}
		conditional := argument.AsConditionalExpression()
		mapper := l.overloadParameterIndex(parameters, conditional.Condition)
		if mapper < 0 || mapper == shape.iterable || shape.mapper >= 0 && mapper != shape.mapper || !element(conditional.WhenFalse) {
			return shape, false
		}
		if l.overloadParameterCall(parameters, conditional.WhenTrue, element) != mapper {
			return shape, false
		}
		shape.mapper, shape.mapped = mapper, true
	}

	// Every reference to the local is a push's receiver or the return, and no parameter is written.
	return shape, l.overloadLocalConfined(implementation, name, last, shape.local, receivers)
}

// overloadCertificateFunction is whether implementation is a function declaration with a body whose call
// runs that body to its result: not async, whose result is a promise, and not a generator, whose result is
// an iterator.
func overloadCertificateFunction(implementation *ast.Node) bool {
	return implementation != nil && implementation.Kind == ast.KindFunctionDeclaration && implementation.Body() != nil && !ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) && !generatorDeclaration(implementation)
}

// overloadParameterIndex is the index in parameters of the one node names, past parentheses, or -1 when
// node isn't a name of one of them.
func (l *lowering) overloadParameterIndex(parameters []*ast.Node, node *ast.Node) int {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return -1
	}
	for index, parameter := range parameters {
		if l.symbol(node) == l.symbol(parameter.Name()) {
			return index
		}
	}
	return -1
}

// overloadParameterCall is the index in parameters of the one node calls directly, past parentheses, with
// no ?., no type arguments and one argument argument accepts, or -1 when node isn't such a call.
func (l *lowering) overloadParameterCall(parameters []*ast.Node, node *ast.Node, argument func(*ast.Node) bool) int {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return -1
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || call.TypeArguments != nil || len(call.Arguments.Nodes) != 1 || !argument(call.Arguments.Nodes[0]) {
		return -1
	}
	return l.overloadParameterIndex(parameters, call.Expression)
}

// overloadConstDeclaration is statement's one declaration when statement is a const declaring one name,
// or nil.
func overloadConstDeclaration(statement *ast.Node) *ast.Node {
	if statement.Kind != ast.KindVariableStatement {
		return nil
	}
	list := statement.AsVariableStatement().DeclarationList
	declarations := list.AsVariableDeclarationList().Declarations.Nodes
	if list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst || len(declarations) != 1 || !ast.IsIdentifier(declarations[0].Name()) {
		return nil
	}
	return declarations[0]
}

// overloadFreshLocal reads a certified body's frame, at least three statements: the first a const
// declaring one name, whose initializer fresh accepts, and the last a bare return of that name. It answers
// the declared name and the return, or nils when statements aren't that frame.
func (l *lowering) overloadFreshLocal(statements []*ast.Node, fresh func(initializer *ast.Node) bool) (name, last *ast.Node) {
	if len(statements) < 3 {
		return nil, nil
	}
	declaration := overloadConstDeclaration(statements[0])
	if declaration == nil || declaration.Initializer() == nil || !fresh(declaration.Initializer()) {
		return nil, nil
	}
	last = statements[len(statements)-1]
	if last.Kind != ast.KindReturnStatement || last.AsReturnStatement().Expression == nil {
		return nil, nil
	}
	returned := ast.SkipParentheses(last.AsReturnStatement().Expression)
	if !ast.IsIdentifier(returned) || l.symbol(returned) != l.symbol(declaration.Name()) {
		return nil, nil
	}
	return declaration.Name(), last
}

// overloadOnlyLoop is the one statement of kind among a frame's statements between its first and last, or
// nil when there are none or several.
func overloadOnlyLoop(statements []*ast.Node, kind ast.Kind) *ast.Node {
	var loop *ast.Node
	for _, statement := range statements[1 : len(statements)-1] {
		if statement.Kind != kind {
			continue
		}
		if loop != nil {
			return nil
		}
		loop = statement
	}
	return loop
}

// overloadMutatorCall reads statement as a call of the library's own owner.method on local, by name, with
// arity arguments: an expression statement whose call has no ?. and whose callee is a property access on
// local's name. It answers the receiver and the arguments, or a nil receiver when statement is anything
// else.
func (l *lowering) overloadMutatorCall(statement *ast.Node, local *ast.Symbol, owner, method string, arity int) (*ast.Node, []*ast.Node) {
	if statement.Kind != ast.KindExpressionStatement {
		return nil, nil
	}
	call := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
	if call.Kind != ast.KindCallExpression {
		return nil, nil
	}
	invocation := call.AsCallExpression()
	callee := ast.SkipParentheses(invocation.Expression)
	if invocation.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression || len(invocation.Arguments.Nodes) != arity {
		return nil, nil
	}
	receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil || !ast.IsIdentifier(receiver) || l.symbol(receiver) != local || !l.overloadIntrinsicMethod(callee, owner, method) {
		return nil, nil
	}
	return receiver, invocation.Arguments.Nodes
}

// overloadLocalConfined is whether local, the fresh container declared at name and returned by last, is
// written only through uses, the receivers of the library's own mutator the certificate judged: every
// reference to it is one of uses or the return, no closure captures it, nothing in implementation's body
// writes one of its parameters, and no return but last leaves.
func (l *lowering) overloadLocalConfined(implementation, name, last *ast.Node, local *ast.Symbol, uses map[*ast.Node]bool) bool {
	parameters := implementation.Parameters()
	returned := ast.SkipParentheses(last.AsReturnStatement().Expression)
	type reference struct {
		node     *ast.Node
		captured bool
	}
	references := []reference{}
	written, leaves := false, false
	var visit func(node *ast.Node, captured bool)
	visit = func(node *ast.Node, captured bool) {
		if ast.IsIdentifier(node) && node != name && l.valueSymbol(node) == local {
			references = append(references, reference{node: node, captured: captured})
		}
		if ast.IsIdentifier(node) && ast.IsAssignmentTarget(node) {
			for _, parameter := range parameters {
				if l.valueSymbol(node) == l.symbol(parameter.Name()) {
					written = true
				}
			}
		}
		if node.Kind == ast.KindReturnStatement && node != last && !captured {
			leaves = true
		}
		inner := captured || ast.IsFunctionLike(node)
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, inner)
			return false
		})
	}
	implementation.Body().ForEachChild(func(child *ast.Node) bool {
		visit(child, false)
		return false
	})
	if written {
		return false // A written parameter isn't what the caller passed, the premise the overload's types are read under.
	}
	if leaves {
		return false // Another return hands back something other than the local.
	}
	for _, found := range references {
		if found.captured {
			return false // A closure can push, set or write after the loop, or hand the container out (array-capture, map-capture).
		}
		if uses[found.node] || found.node == returned {
			continue
		}
		if overloadArrayWritten(found.node) {
			return false // An index, length or whole write puts a value no push proved (array-index-write).
		}
		return false // Any other read is an alias or a mutator no certificate judged, writing what nothing proved (array-alias, map-alias, map-wrong-set).
	}
	return true
}

// overloadArrayWritten is whether reference, an identifier, is written itself or is the receiver of an
// access that's written.
func overloadArrayWritten(reference *ast.Node) bool {
	if ast.IsAssignmentTarget(reference) {
		return true
	}
	child, parent := reference, reference.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression && parent.Kind != ast.KindElementAccessExpression || parent.Expression() != child {
		return false
	}
	return ast.IsAssignmentTarget(parent)
}

// overloadIntrinsicMethod is whether callee, a property access, is the library's own owner.method, Array's
// push or Map's set: every root declaration of its member is that method of that interface in the
// library, never a program method of the same name. A rewrite of the prototype's method at run time is
// the library's to stop (array-prototype).
func (l *lowering) overloadIntrinsicMethod(callee *ast.Node, owner, method string) bool {
	symbol := l.memberSymbol(callee)
	if symbol == nil {
		return false
	}
	roots := l.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			if declaration.Kind != ast.KindMethodSignature || !ast.IsIdentifier(declaration.Name()) || declaration.Name().Text() != method || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
			parent := declaration.Parent
			if parent == nil || parent.Kind != ast.KindInterfaceDeclaration || parent.Name() == nil || parent.Name().Text() != owner {
				return false
			}
		}
	}
	return true
}

// overloadArrayCertificate is whether certificate A serves overload's result: implementation's body has
// its shape, and each value a push can put under this overload's arguments relates to the element of the
// array the overload promises. Absent from the overload, the mapper is undefined and the bare test picks
// the element; required by it, the mapper is a function and the test picks its result. A mapper the
// overload makes optional could be either, so it fails. censusOverload asks it with the implementation's
// binders mapped to the overload's. That the mapper takes what the loop hands it isn't judged here: the
// mapper's parameter is related to the iterable's element by censusOverload's parameter census, which
// runs first and refuses an overload whose iterable and mapper disagree (Language's q02 on #5fws6q8).
func (l *lowering) overloadArrayCertificate(implementation, overload *ast.Node) bool {
	shape, shaped := l.overloadArrayShapeOf(implementation)
	if !shaped {
		return false
	}
	declared := overload.Parameters()
	if len(declared) > len(implementation.Parameters()) || shape.iterable >= len(declared) {
		return false
	}
	for _, parameter := range declared {
		if parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return false // A rest parameter takes any count of arguments, so a position says nothing.
		}
	}
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)))
	if !l.isArrayType(promised) {
		return false
	}
	held := l.elementTypeOfArray(promised)
	element := l.checker.GetTypeOfSymbol(shape.element)
	present := shape.mapped && shape.mapper < len(declared)
	if shape.bare || shape.mapped && !present {
		if !l.censusRelated(element, held) {
			return false
		}
	}
	if !present {
		return true
	}
	parameter := declared[shape.mapper]
	if ast.HasQuestionToken(parameter) || !ast.IsIdentifier(parameter.Name()) {
		return false
	}
	// Not callable alone is undefined or a union with it, either of which the bare test can send to the
	// element.
	result, callable := l.overloadCallableResult(l.concrete(l.censusCallableParameterType(l.symbol(parameter.Name()))))
	return callable && l.censusRelated(result, held)
}

// overloadCallableResult is what a function of type callable returns, when callable is an object type with
// exactly one call signature.
func (l *lowering) overloadCallableResult(callable *checker.Type) (*checker.Type, bool) {
	if callable.Flags()&checker.TypeFlagsObject == 0 {
		return nil, false
	}
	signatures := l.checker.GetSignaturesOfType(callable, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, false
	}
	return l.checker.GetReturnTypeOfSignature(signatures[0]), true
}

// overloadCertified is whether censusOverload accepted overload by certificate A, B or D, judged once per
// overload. A call can be lowered before its set's declaration is, an instance of a generic set among
// them, so an overload not yet judged is judged here by running the set's census, which records the
// certificate's acceptance (census_small.go). Only a body with certificate A's, B's or D's shape is asked,
// so every other set's census runs where it always did.
func (l *lowering) overloadCertified(overload *ast.Node) (bool, error) {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil || overload.Name() == nil {
		return false, nil
	}
	if certified, judged := l.overloadCertificates[overload]; judged {
		return certified, nil
	}
	certified := false
	if implementation := l.censusImplementation(overload); implementation != nil {
		_, arrayShaped := l.overloadArrayShapeOf(implementation)
		_, mapShaped := l.overloadMapShapeOf(implementation)
		_, filterShaped := l.overloadFilterShapeOf(implementation)
		if arrayShaped || mapShaped || filterShaped {
			// Its own verdict, never its set's: an unserved sibling stops only the calls that resolve to it
			// (#4czdn4q). overloadCertificateCall asks overloadCallVerdict first, so a call reaching here with
			// this overload unserved has already stopped there.
			if err := l.overloadVerdict(overload); err != nil {
				return false, err
			}
			certified = l.overloadCertificates[overload]
		}
	}
	l.noteOverloadCertificate(overload, certified)
	return certified, nil
}

// noteOverloadCertificate records whether census accepted overload by certificate A, B or D.
func (l *lowering) noteOverloadCertificate(overload *ast.Node, certified bool) {
	if l.overloadCertificates == nil {
		l.overloadCertificates = map[*ast.Node]bool{}
	}
	l.overloadCertificates[overload] = certified
}

// overloadCertificateEntry is whether node enters overload as certificates A and B read their premise: a
// call whose callee, past parentheses, is an identifier or a property access whose symbol, through an
// import alias, is overload's set, with no spread argument, so the overload's parameters hold exactly the
// arguments the caller passed. arrayFrom?.(it) is direct, its question token on the call, and so is
// ts.arrayFrom(it). Anything else answers false, a node that isn't a call among them, so the answer never
// leans on a value use of the set being stopped elsewhere (overloadSetValueStop).
func (l *lowering) overloadCertificateEntry(node *ast.Node, overload *ast.Node) (erase bool) {
	if node == nil || node.Kind != ast.KindCallExpression || overload == nil || overload.Name() == nil {
		return false
	}
	call := node.AsCallExpression()
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
	}
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsIdentifier(callee) && callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	set := l.symbol(overload.Name())
	return set != nil && l.symbol(callee) == set
}

// overloadCertificateCall is nil unless node, a call, a new or a tagged template, resolves to an overload
// census accepted by certificate A or B, by the node-test proof or by the guarded-array certificate, and
// enters it other than directly (overloadCertificateEntry), or enters it directly and is stopped there: a
// certificate's call whose instance stores its result otherwise than the call reads it
// (overloadCertificateStorage), a node-test call overloadNodeTestCall refuses, or a guarded-array call
// whose premises overloadGuardedArrayCall finds unkept. An array or Map result can't be checked at the call, and the node-test
// proof's premise is the caller's arguments, so a non-direct entry is NotYet, naming why. Before any of
// that, the call is judged by its overload's census verdict, or its set's (overloadCallVerdict). Each road
// a call enters lowering by asks it beside noteOverloadCall, before the callee or an argument is lowered.
func (l *lowering) overloadCertificateCall(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression {
		return nil
	}
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil {
		return nil
	}
	overload := resolved.Declaration()
	if err := l.overloadCallVerdict(node, overload); err != nil {
		return err
	}
	certified, err := l.overloadCertified(overload)
	if err != nil {
		return err
	}
	tested, err := l.overloadNodeTestCertified(overload)
	if err != nil {
		return err
	}
	guarded, err := l.overloadGuardedArrayCertified(overload)
	if err != nil {
		return err
	}
	if !certified && !tested && !guarded {
		return nil
	}
	if l.overloadCertificateEntry(node, overload) {
		if tested {
			// Its own road, never the storage condition's: the node-test proof's result is a node or
			// undefined, held at the top level, which censusOverloadFit fits to the call's representation or
			// stops. overloadCertificateStorage guards an array's or Map's elements, which no fit reaches.
			return l.overloadNodeTestCall(node, resolved)
		}
		if guarded {
			// The guarded-array certificate's premises first, then its NodeArray's storage, which
			// overloadCertificateSameHeld reads through the NodeArray's own layout.
			if err := l.overloadGuardedArrayCall(node, overload, resolved); err != nil {
				return err
			}
		}
		return l.overloadCertificateStorage(node, overload, resolved)
	}
	form := "a callee other than the overload set's name"
	switch {
	case node.Kind == ast.KindNewExpression:
		form = "new"
	case node.Kind == ast.KindTaggedTemplateExpression:
		form = "a tagged template"
	case overloadCertificateSpread(node):
		form = "a spread argument"
	case ast.SkipParentheses(node.AsCallExpression().Expression).Kind == ast.KindBinaryExpression && ast.SkipParentheses(node.AsCallExpression().Expression).AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken:
		form = "a comma expression callee"
	}
	return l.notYet(node, "an overload certified by its body, entered other than by a direct call: "+form)
}

// overloadCallVerdict judges node, a call, a new or a tagged template resolved to overload, by Adamic's
// amendment of rulings.md:131 (#4czdn4q). A direct call (overloadCertificateEntry) is judged by overload
// alone, so a call to a served overload lowers whatever its siblings' verdicts, and a call to an unserved
// one stops here in that overload's own words. Any other entry, a spread, a callee other than the set's
// own name, a new or a tagged template, can't say which overload's premise holds, so it's judged by the
// whole set (overloadSetVerdict). Either stop is said at node. Only a function declaration's set is judged
// (censusImplementation), never a method's, whose neighbors stay "an overloaded override" and "a function
// without a body" (censusOverloads). A call it lets through is recorded in overloadVerdictCalls, which
// checkOverloadVerdictCalls holds every call in the program to once lowering is done, so a road that
// misses this gate, or a caller that drops its stop and lowers the call another way, can't lower a call
// to an unserved overload silently.
func (l *lowering) overloadCallVerdict(node, overload *ast.Node) error {
	judged, err := l.overloadCallJudgment(node, overload)
	if err != nil || !judged {
		return err
	}
	if l.overloadVerdictCalls == nil {
		l.overloadVerdictCalls = map[*ast.Node]bool{}
	}
	l.overloadVerdictCalls[node] = true
	if l.overloadCertificateEntry(node, overload) {
		l.activateCallbackReads(overload)
	} else {
		l.activateCallbackSetReads(l.censusImplementation(overload))
	}
	return nil
}

// overloadCallJudgment is overloadCallVerdict's answer for node, recording nothing: whether overload is
// an overload of a function declaration's set, by which the call is judged, and its stop when it's stopped.
func (l *lowering) overloadCallJudgment(node, overload *ast.Node) (bool, error) {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
		return false, nil
	}
	implementation := l.censusImplementation(overload)
	if implementation == nil {
		return false, nil
	}
	if l.overloadCertificateEntry(node, overload) {
		return true, overloadVerdictAt(l.overloadVerdict(overload), l.program.Where(node))
	}
	return true, overloadVerdictAt(l.overloadSetVerdict(implementation), l.program.Where(node))
}

// checkOverloadVerdictCalls holds every call, new and tagged template in modules to overloadCallVerdict's
// judgment, once lowering is done: one the checker resolved to an overload of a function declaration's set,
// whose judgment is a stop, and which the gate never let through (overloadVerdictCalls), is a road that
// missed the gate, or a stop a caller dropped, and stops here in the gate's own words, the first in module
// order. It fails closed: it walks every module whole, not walkLoweredRegions, since a call in a region no
// note records (a generic class's initializer) would otherwise go unjudged, and the only calls it skips are
// those inside the body of a generic function declaration no instance was lowered from, which no code is
// built for (overloadVerdictUninstantiated). A call it stops in a region lowering never built, a method of
// a class never constructed among them, is a call to a set with an unserved overload: a module-level set's
// declaration stopped that program before #4czdn4q anyway, and a set nested in an unbuilt body, which was
// never judged before, is the one stop this adds, erring closed.
func (l *lowering) checkOverloadVerdictCalls(modules []*ast.SourceFile) error {
	var stop error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if stop != nil || ast.IsPartOfTypeNode(node) {
			return stop != nil
		}
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression {
			if resolved := l.checker.GetResolvedSignature(node); resolved != nil && !l.overloadVerdictCalls[node] {
				if judged, err := l.overloadCallJudgment(node, resolved.Declaration()); judged && err != nil && !l.overloadVerdictUninstantiated(node) {
					stop = err
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		if module.AsNode().ForEachChild(visit) {
			break
		}
	}
	return stop
}

// overloadVerdictUninstantiated is whether node lies inside the body of a generic function declaration
// no instance was lowered from: instantiateFunctionSignature keys each instance in genericInstances by the
// declaration's Where, alone or followed by a comma and what the instance is of, and a declaration with no
// such key was never lowered, so nothing in it was built.
func (l *lowering) overloadVerdictUninstantiated(node *ast.Node) bool {
	for enclosing := node.Parent; enclosing != nil; enclosing = enclosing.Parent {
		if enclosing.Kind != ast.KindFunctionDeclaration || enclosing.Body() == nil || len(enclosing.TypeParameters()) == 0 {
			continue
		}
		where := l.program.Where(enclosing)
		instantiated := false
		for key := range l.genericInstances {
			instantiated = instantiated || key == where || strings.HasPrefix(key, where+",")
		}
		if !instantiated {
			return true
		}
	}
	return false
}

// overloadCertificateStorage is nil when the implementation's instance a direct call to overload lowers
// stores its result as the call reads overload's promise, resolved, and NotYet at the call otherwise.
// The instance's result is the implementation's own, instantiated as instantiateFunctionSignature
// instantiates it for this call (instanceTypes), never by a second inference; a non-generic
// implementation's is its declared result. The certificates prove which values the result holds, not how
// it holds them: at arrayFrom's T = string, U = number the instance builds (string | number)[], each
// element a boxed union, and the call reading number[] read each box as a number (Lower's probe J-h37mw8d
// on #5fws6q8, native 1.0,1.0 where Node prints 4.0,2.0). Fails closed on anything it can't read.
func (l *lowering) overloadCertificateStorage(node, overload *ast.Node, resolved *checker.Signature) error {
	implementation := l.censusImplementation(overload)
	if implementation == nil {
		return l.notYet(node, "a certified overload whose implementation isn't found")
	}
	signature := l.checker.GetSignatureFromDeclaration(implementation)
	stored := l.concrete(l.checker.GetReturnTypeOfSignature(signature))
	if declaration, isGeneric := l.generics[l.symbol(implementation.Name())]; isGeneric {
		// The declaration the call lowers its instance of, as call does, read as instantiateFunctionSignature reads it.
		target := l.checker.GetSignatureFromDeclaration(declaration)
		if mapper := l.instanceMapper(declaration, l.instanceTypes(declaration, target, resolved)); mapper != nil {
			stored = instantiateType(l.checker, l.checker.GetReturnTypeOfSignature(target), mapper)
		}
	}
	read := l.concrete(l.checker.GetReturnTypeOfSignature(resolved))
	if !l.overloadCertificateSameStorage(stored, read) {
		return l.notYet(node, "a certified overload whose implementation stores "+l.checker.TypeToString(stored)+" where the call reads "+l.checker.TypeToString(read))
	}
	return nil
}

// overloadCertificateSameStorage is whether a value of type stored is held as one of type read is, decided
// only by the ir.Type each side lowers to, level by level, through the functions lowering and the emitter
// hold a checker type by: kept for the result, an array's element and a Map's value, representation for a
// Map's key and a Set's element, as mapTypes and the array roads read them. The same at the top and, through
// an array, Map or Set, the same all the way down. A NodeArray is read through its own layout
// (nodeArrayLayoutOf), never isArrayType, which never takes one. Never checker identity or assignability, since two types
// tsc calls one can lower differently, a branded arm or a literal widened in one place (Compiler's
// condition on #5fws6q8). Any other type with type arguments, a tuple among them, isn't read: false.
func (l *lowering) overloadCertificateSameStorage(stored, read *checker.Type) bool {
	return l.overloadCertificateSameHeld(stored, read, false)
}

// overloadCertificateSameHeld is overloadCertificateSameStorage at one level, keyed when that level is a
// Map's key or a Set's element.
func (l *lowering) overloadCertificateSameHeld(stored, read *checker.Type, keyed bool) bool {
	held := l.kept
	if keyed {
		held = l.representation
	}
	storedHeld, storedKnown := held(stored)
	readHeld, readKnown := held(read)
	if !storedKnown || !readKnown || storedHeld != readHeld {
		return false
	}
	if storedHeld.IsReference() {
		// A reference that may be missing is held as the reference itself, missing as null (ir.Maybe), so
		// once both sides are held alike, a union of one type with undefined is read through that type:
		// filter's instance stores readonly string[] | undefined where overload 2's call reads string[]
		// (certificate D, #4z7y6pq), and each array's element is still compared below.
		if defined := l.definedMembers(stored); len(defined) == 1 {
			stored = defined[0]
		}
		if defined := l.definedMembers(read); len(defined) == 1 {
			read = defined[0]
		}
	}
	// A NodeArray is held as its layout says, the fixed metadata fields after the array header and its
	// elements as its element's representation, so two are held alike when their layouts are and their
	// elements are, all the way down: the guarded-array certificate's NodeArray<Node> read as NodeArray<TOut>
	// (Lower's condition 5 on #ytagqs5, certificate A's lesson that a value fact is no storage fact).
	storedLayout, storedElement, storedNodeArray := l.nodeArrayLayoutOf(stored)
	readLayout, readElement, readNodeArray := l.nodeArrayLayoutOf(read)
	if storedNodeArray || readNodeArray {
		return storedNodeArray && readNodeArray && slices.Equal(storedLayout.Fields, readLayout.Fields) && l.overloadCertificateSameHeld(storedElement, readElement, false)
	}
	storedArguments, readArguments := l.typeArguments(stored), l.typeArguments(read)
	if len(storedArguments) == 0 && len(readArguments) == 0 {
		return true
	}
	collection := func(proven *checker.Type) string {
		switch {
		case l.isArrayType(proven):
			return "Array"
		case l.isLibraryType(proven, "Map", "ReadonlyMap"):
			return "Map"
		case l.isLibraryType(proven, "Set", "ReadonlySet"):
			return "Set"
		}
		return ""
	}
	kind := collection(stored)
	if kind == "" || kind != collection(read) || len(storedArguments) != len(readArguments) {
		return false
	}
	for index := range storedArguments {
		if !l.overloadCertificateSameHeld(storedArguments[index], readArguments[index], kind != "Array" && index == 0) {
			return false
		}
	}
	return true
}

// overloadCertificateSpread is whether call, a call expression, passes a spread argument.
func overloadCertificateSpread(call *ast.Node) bool {
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}
