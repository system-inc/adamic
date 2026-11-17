package lower

import (
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
// array result can't be checked at a call, so each one is NotYet there.

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
	if implementation == nil || implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil || ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) || generatorDeclaration(implementation) {
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
	parameterOf := func(node *ast.Node) int {
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
	statements := implementation.Body().AsBlock().Statements.Nodes
	if len(statements) < 3 {
		return shape, false
	}

	// The fresh array: a const, so nothing rebinds it, and an empty literal, so it starts holding nothing.
	first := statements[0]
	if first.Kind != ast.KindVariableStatement {
		return shape, false
	}
	list := first.AsVariableStatement().DeclarationList
	declarations := list.AsVariableDeclarationList().Declarations.Nodes
	if list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst || len(declarations) != 1 || !ast.IsIdentifier(declarations[0].Name()) {
		return shape, false
	}
	initializer := declarations[0].Initializer()
	if initializer == nil || initializer.Kind != ast.KindArrayLiteralExpression || len(initializer.AsArrayLiteralExpression().Elements.Nodes) != 0 {
		return shape, false
	}
	shape.local = l.symbol(declarations[0].Name())

	last := statements[len(statements)-1]
	if last.Kind != ast.KindReturnStatement || last.AsReturnStatement().Expression == nil {
		return shape, false
	}
	returned := ast.SkipParentheses(last.AsReturnStatement().Expression)
	if !ast.IsIdentifier(returned) || l.symbol(returned) != shape.local {
		return shape, false
	}

	// The one loop, over a parameter as the caller passed it: a const element binding nothing rebinds.
	var loop *ast.Node
	for _, statement := range statements[1 : len(statements)-1] {
		if statement.Kind != ast.KindForOfStatement {
			continue
		}
		if loop != nil {
			return shape, false
		}
		loop = statement
	}
	if loop == nil {
		return shape, false
	}
	iteration := loop.AsForInOrOfStatement()
	shape.iterable = parameterOf(iteration.Expression)
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
		if push.Kind != ast.KindExpressionStatement {
			return shape, false
		}
		call := ast.SkipParentheses(push.AsExpressionStatement().Expression)
		if call.Kind != ast.KindCallExpression {
			return shape, false
		}
		invocation := call.AsCallExpression()
		callee := ast.SkipParentheses(invocation.Expression)
		if invocation.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression || len(invocation.Arguments.Nodes) != 1 {
			return shape, false
		}
		receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if callee.AsPropertyAccessExpression().QuestionDotToken != nil || !ast.IsIdentifier(receiver) || l.symbol(receiver) != shape.local || !l.overloadArrayIntrinsicPush(callee) {
			return shape, false
		}
		receivers[receiver] = true
		argument := ast.SkipParentheses(invocation.Arguments.Nodes[0])
		if element(argument) {
			shape.bare = true
			continue
		}
		if argument.Kind != ast.KindConditionalExpression {
			return shape, false
		}
		conditional := argument.AsConditionalExpression()
		mapper := parameterOf(conditional.Condition)
		if mapper < 0 || mapper == shape.iterable || shape.mapper >= 0 && mapper != shape.mapper || !element(conditional.WhenFalse) {
			return shape, false
		}
		mapped := ast.SkipParentheses(conditional.WhenTrue)
		if mapped.Kind != ast.KindCallExpression {
			return shape, false
		}
		mapping := mapped.AsCallExpression()
		if mapping.QuestionDotToken != nil || mapping.TypeArguments != nil || parameterOf(mapping.Expression) != mapper || len(mapping.Arguments.Nodes) != 1 || !element(mapping.Arguments.Nodes[0]) {
			return shape, false
		}
		shape.mapper, shape.mapped = mapper, true
	}

	// Every reference to the local is a push's receiver or the return, and no parameter is written.
	type reference struct {
		node     *ast.Node
		captured bool
	}
	references := []reference{}
	written, leaves := false, false
	var visit func(node *ast.Node, captured bool)
	visit = func(node *ast.Node, captured bool) {
		if ast.IsIdentifier(node) && node != declarations[0].Name() && l.symbol(node) == shape.local {
			references = append(references, reference{node: node, captured: captured})
		}
		if ast.IsIdentifier(node) && ast.IsAssignmentTarget(node) && parameterOf(node) >= 0 {
			written = true
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
		return shape, false // A written parameter isn't what the caller passed, the premise the overload's types are read under.
	}
	if leaves {
		return shape, false // Another return hands back something other than the local.
	}
	for _, found := range references {
		if found.captured {
			return shape, false // A closure can push or write after the loop, or hand the array out (array-capture).
		}
		if receivers[found.node] || found.node == returned {
			continue
		}
		if overloadArrayWritten(found.node) {
			return shape, false // An index, length or whole write puts a value no push proved (array-index-write).
		}
		return shape, false // Any other read is an alias, a second view that can write what no push proved (array-alias).
	}
	return shape, true
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

// overloadArrayIntrinsicPush is whether callee, a property access, is the Array prototype's own push:
// every root declaration of its member is Array's push in the library, never a program method named
// push. A rewrite of the prototype's push at run time is the library's to stop (array-prototype).
func (l *lowering) overloadArrayIntrinsicPush(callee *ast.Node) bool {
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
			if declaration.Kind != ast.KindMethodSignature || !ast.IsIdentifier(declaration.Name()) || declaration.Name().Text() != "push" || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
			owner := declaration.Parent
			if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil || owner.Name().Text() != "Array" {
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
	mapper := l.concrete(l.censusCallableParameterType(l.symbol(parameter.Name())))
	if mapper.Flags()&checker.TypeFlagsObject == 0 {
		return false // Undefined or a union with it, either of which the bare test can send to the element.
	}
	signatures := l.checker.GetSignaturesOfType(mapper, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return false
	}
	return l.censusRelated(l.checker.GetReturnTypeOfSignature(signatures[0]), held)
}

// overloadArrayCertified is whether censusOverload accepted overload by certificate A, judged once per
// overload. A call can be lowered before its set's declaration is, an instance of a generic set among
// them, so an overload not yet judged is judged here by running the set's census, which records the
// certificate's acceptance (census_small.go). Only a body with certificate A's shape is asked, so every
// other set's census runs where it always did.
func (l *lowering) overloadArrayCertified(overload *ast.Node) (bool, error) {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil || overload.Name() == nil {
		return false, nil
	}
	if certified, judged := l.overloadCertificates[overload]; judged {
		return certified, nil
	}
	certified := false
	if implementation := l.censusImplementation(overload); implementation != nil {
		if _, shaped := l.overloadArrayShapeOf(implementation); shaped {
			if err := l.censusOverloads(implementation); err != nil {
				return false, err
			}
			certified = l.overloadCertificates[overload]
		}
	}
	l.noteOverloadCertificate(overload, certified)
	return certified, nil
}

// noteOverloadCertificate records whether census accepted overload by certificate A.
func (l *lowering) noteOverloadCertificate(overload *ast.Node, certified bool) {
	if l.overloadCertificates == nil {
		l.overloadCertificates = map[*ast.Node]bool{}
	}
	l.overloadCertificates[overload] = certified
}

// overloadCertificateEntry is whether node enters overload as certificate A's premise reads it: a call
// whose callee, past parentheses, is an identifier or a property access whose symbol, through an import
// alias, is overload's set, with no spread argument, so the overload's parameters hold exactly the
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
// census accepted by certificate A or by the node-test proof and enters it other than directly
// (overloadCertificateEntry), or enters the node-test proof's directly and overloadNodeTestCall refuses
// it. An array result can't be checked at the call, and the node-test proof's premise is the caller's
// arguments, so a non-direct entry is NotYet, naming its form. Each road a call enters lowering by asks it
// beside noteOverloadCall, before the callee or an argument is lowered.
func (l *lowering) overloadCertificateCall(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression {
		return nil
	}
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil {
		return nil
	}
	overload := resolved.Declaration()
	certified, err := l.overloadArrayCertified(overload)
	if err != nil {
		return err
	}
	tested, err := l.overloadNodeTestCertified(overload)
	if err != nil {
		return err
	}
	if !certified && !tested {
		return nil
	}
	if l.overloadCertificateEntry(node, overload) {
		if tested {
			return l.overloadNodeTestCall(node, resolved)
		}
		return nil
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

// overloadCertificateSpread is whether call, a call expression, passes a spread argument.
func overloadCertificateSpread(call *ast.Node) bool {
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}
