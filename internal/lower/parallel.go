package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const parallelMoveFix = "make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"

func (lowering *lowering) parallelMap(node *ast.Node) (ir.Expression, error) {
	if checkParallelMapError := lowering.checkParallelMap(node); checkParallelMapError != nil {
		return nil, checkParallelMapError
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	moved := lowering.parallelMoves(node)
	items, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, expressionError
	}
	work, expressionError := lowering.expression(arguments[1])
	if expressionError != nil {
		return nil, expressionError
	}
	if items.Type() != ir.Array || work.Type() != ir.Closure {
		return nil, lowering.notYet(node, "parallelMap operands without array and closure representations")
	}
	result, elementTypeError := lowering.elementType(node)
	if elementTypeError != nil {
		return nil, elementTypeError
	}
	// Globals are ordinarily direct loads, not closure cells. Carry every reference
	// the task's transitive call graph reads so runtime marking reaches them too.
	proof := &parallelProof{
		l:                  lowering,
		active:             map[*ast.Node]bool{},
		checked:            map[*ast.Node]bool{},
		shareableFunctions: map[*ast.Node]bool{},
	}
	body := proof.functionValue(arguments[1], map[*ast.Symbol]bool{})
	if !moved {
		if functionError := proof.function(body, nil); functionError != nil {
			return nil, functionError
		}
	}
	shared := []ir.Expression{}
	seen := map[*ast.Symbol]bool{}
	for _, global := range proof.globals {
		symbol := proof.symbol(global)
		if seen[symbol] {
			continue
		}
		seen[symbol] = true
		value, expressionError := lowering.expression(global)
		if expressionError != nil {
			return nil, expressionError
		}
		if !value.Type().IsReference() {
			continue
		}
		if read, ok := value.(ir.Read); ok {
			// Zero before initialization is safe to mark. A task's actual read still
			// checks the TDZ; pre-marking must not observe a branch the task never takes.
			read.Checked = false
			value = read
		}
		shared = append(shared, value)
	}
	lowering.hasMovedParallelMap = lowering.hasMovedParallelMap || moved
	lowering.moveSites = append(lowering.moveSites, node)
	return ir.ParallelMap{
		Site:   len(lowering.moveSites),
		Items:  items,
		Work:   work,
		Shared: shared,
		Result: result,
		Moved:  moved,
	}, nil
}

// Run before general refusals, so async work gets its concurrency diagnostic even
// when its declaration appears before the call. No safety check is bypassed here.
func (lowering *lowering) parallelPreflight(module *ast.SourceFile) error {
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(node) {
			return false
		}
		if node.Kind == ast.KindCallExpression &&
			lowering.isPreludeFunction(node.AsCallExpression().Expression, "parallelMap") {
			found = lowering.checkParallelMap(node)
			if found != nil {
				return false
			}
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

func (lowering *lowering) checkParallelMap(node *ast.Node) error {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 2 {
		return lowering.notYet(node, "parallelMap with other than items and work")
	}
	items := lowering.concrete(lowering.checker.GetTypeAtLocation(arguments[0]))
	if !lowering.isArrayType(items) {
		return &Refused{
			Where: lowering.program.Where(arguments[0]),
			What:  "parallelMap items must be an array",
			Fix:   parallelMoveFix,
		}
	}
	if lowering.parallelMoves(node) {
		return lowering.checkMove(node)
	}
	// Even primitive elements need an immutable array: every task can capture the array itself.
	if !lowering.isLibraryType(items, "ReadonlyArray") {
		return &Refused{
			Where: lowering.program.Where(arguments[0]),
			What:  "parallelMap items are not shareable: items is a mutable " + lowering.checker.TypeToString(items),
			Fix:   parallelMoveFix,
		}
	}
	for _, element := range lowering.checker.GetTypeArguments(items) {
		path := lowering.checker.TypeToString(element)
		if why := lowering.shareable(element, path, arguments[0]); why != "" {
			return &Refused{
				Where: lowering.program.Where(arguments[0]),
				What:  "parallelMap items are not shareable: " + why,
				Fix:   parallelMoveFix,
			}
		}
	}
	proof := &parallelProof{
		l:                  lowering,
		active:             map[*ast.Node]bool{},
		checked:            map[*ast.Node]bool{},
		shareableFunctions: map[*ast.Node]bool{},
	}
	work := proof.functionValue(arguments[1], map[*ast.Symbol]bool{})
	if work == nil {
		return proof.effect(arguments[1], []string{"work"}, "calls a closure whose body isn't known")
	}
	if ast.IsIdentifier(ast.SkipParentheses(arguments[1])) {
		declaration := proof.declaration(ast.SkipParentheses(arguments[1]))
		if declaration != nil && !proof.immutable(declaration) {
			return proof.capture(arguments[1], nil, nil)
		}
	}
	return proof.function(work, nil)
}

type parallelProof struct {
	l                  *lowering
	active, checked    map[*ast.Node]bool
	shareableFunctions map[*ast.Node]bool
	globals            []*ast.Node
}

func (proof *parallelProof) effect(where *ast.Node, chain []string, what string) error {
	return &Refused{
		Where: proof.l.program.Where(where),
		What:  "parallelMap work writes state another task can reach: " + strings.Join(chain, " -> ") + " " + what,
		Fix:   "keep work and everything it calls pure; write only task locals and fresh objects",
	}
}

func parallelInside(declaration, function *ast.Node) bool {
	for node := declaration; node != nil; node = node.Parent {
		if node == function {
			return true
		}
	}
	return false
}

func (proof *parallelProof) declaration(node *ast.Node) *ast.Node {
	if symbol := proof.symbol(node); symbol != nil && len(symbol.Declarations) == 1 {
		return symbol.Declarations[0]
	}
	return nil
}

func (proof *parallelProof) functionValue(node *ast.Node, visited map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipParentheses(node)
	if ast.IsFunctionLike(node) && node.Body() != nil {
		return node
	}
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := proof.symbol(node)
	if symbol == nil || visited[symbol] {
		return nil
	}
	visited[symbol] = true
	declaration := proof.declaration(node)
	if declaration == nil {
		return nil
	}
	if ast.IsFunctionLike(declaration) && declaration.Body() != nil {
		return declaration
	}
	if declaration.Kind == ast.KindVariableDeclaration && proof.immutable(declaration) {
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			return proof.functionValue(initializer, visited)
		}
	}
	return nil
}

// A parameter can be captured only when its entire source function never rebinds it.
// We deliberately include writes in nested closures and after the parallel call.
func (proof *parallelProof) immutable(declaration *ast.Node) bool {
	if declaration.Kind == ast.KindFunctionDeclaration {
		return true
	}
	if declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindBindingElement {
		for node := declaration.Parent; node != nil; node = node.Parent {
			if node.Kind == ast.KindVariableDeclarationList {
				return node.Flags&ast.NodeFlagsConstant != 0
			}
		}
		return false
	}
	return declaration.Kind == ast.KindParameter && !proof.reassigned(declaration)
}

func parallelWrite(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		return node.AsBinaryExpression().Left
	}
	if node.Kind == ast.KindPrefixUnaryExpression {
		u := node.AsPrefixUnaryExpression()
		if u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken {
			return u.Operand
		}
	}
	if node.Kind == ast.KindPostfixUnaryExpression {
		return node.AsPostfixUnaryExpression().Operand
	}
	if node.Kind == ast.KindDeleteExpression {
		return node.AsDeleteExpression().Expression
	}
	return nil
}

func (proof *parallelProof) reassigned(declaration *ast.Node) bool {
	symbol := proof.l.symbol(declaration.Name())
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found || ast.IsPartOfTypeNode(node) {
			return false
		}
		if target := parallelWrite(node); target != nil {
			var names ast.Visitor
			names = func(name *ast.Node) bool {
				if ast.IsIdentifier(name) && proof.l.symbol(name) == symbol {
					found = true
				}
				// A field/element write doesn't rebind its receiver.
				if name.Kind == ast.KindPropertyAccessExpression || name.Kind == ast.KindElementAccessExpression {
					return false
				}
				return name.ForEachChild(names)
			}
			names(ast.SkipParentheses(target))
		}
		return node.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return found
}

func (proof *parallelProof) capture(node, owner *ast.Node, chain []string) error {
	declaration := proof.declaration(node)
	if declaration != nil && parallelInside(declaration, owner) {
		return nil
	}
	if declaration == nil {
		return proof.effect(node, chain, "reads a binding whose declaration isn't known")
	}
	name := node.Text()
	proof.global(node)
	if !proof.immutable(declaration) {
		return &Refused{
			Where: proof.l.program.Where(node),
			What:  "task capture '" + name + "' is not shareable: " + name + " is not an immutable binding",
			Fix:   parallelMoveFix,
		}
	}
	if work := proof.functionValue(node, map[*ast.Symbol]bool{}); work != nil {
		return proof.function(work, chain)
	}
	if why := proof.l.shareableWith(
		proof.l.concrete(proof.l.checker.GetTypeOfSymbol(proof.symbol(node))),
		name,
		node,
		proof.shareableFunctions,
	); why != "" {
		return &Refused{
			Where: proof.l.program.Where(node),
			What:  "task capture '" + name + "' is not shareable: " + why,
			Fix:   parallelMoveFix,
		}
	}
	return nil
}

func (proof *parallelProof) function(node *ast.Node, chain []string) error {
	if proof.active[node] || proof.checked[node] {
		return nil
	}
	proof.active[node] = true
	defer delete(proof.active, node)
	name := "work"
	if node.Name() != nil {
		name = node.Name().Text()
	}
	chain = append(append([]string{}, chain...), name)
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return &Refused{
			Where: proof.l.program.Where(node),
			What:  "async work in parallelMap is part 3 (#p286ycm)",
			Fix:   "use synchronous work until concurrency part 3",
		}
	}
	var found error
	var captures []*ast.Node
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(child) {
			return false
		}
		if child.Kind == ast.KindAwaitExpression {
			found = &Refused{
				Where: proof.l.program.Where(child),
				What:  "async work in parallelMap is part 3 (#p286ycm)",
				Fix:   "use synchronous work until concurrency part 3",
			}
			return false
		}
		if target := parallelWrite(child); target != nil {
			found = proof.write(target, node, chain)
			if found != nil {
				return false
			}
		}
		if child.Kind == ast.KindThisKeyword {
			found = proof.effect(child, chain, "reaches this, whose task ownership isn't proven")
			return false
		}
		if child.Kind == ast.KindCallExpression || child.Kind == ast.KindNewExpression {
			found = proof.call(child, node, chain)
			return false
		}
		if ast.IsFunctionLike(child) {
			found = proof.function(child, chain)
			return false
		}
		if parallelReference(child) {
			declaration := proof.declaration(child)
			if declaration == nil || !parallelInside(declaration, node) {
				captures = append(captures, child)
			}
		}
		return child.ForEachChild(visit)
	}
	// Parameter defaults run in the task too.
	for _, parameter := range node.Parameters() {
		if initializer := parameter.AsParameterDeclaration().Initializer; initializer != nil {
			visit(initializer)
		}
	}
	visit(node.Body())
	if found != nil {
		return found
	}
	for _, capture := range captures {
		if proof.l.checker.GetTypeAtLocation(capture).Flags()&checker.TypeFlagsUndefined != 0 &&
			capture.Text() == "undefined" {
			continue
		}
		if captureError := proof.capture(capture, node, chain); captureError != nil {
			return captureError
		}
	}
	proof.checked[node] = true
	return nil
}

// Only direct slots of a never-rebound local allocation qualify. Following a field
// of a fresh object could reach an item or capture stored there, so is conservative.
func (proof *parallelProof) fresh(node, owner *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression {
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	declaration := proof.declaration(node)
	if declaration == nil ||
		!parallelInside(declaration, owner) ||
		declaration.Kind != ast.KindVariableDeclaration ||
		proof.reassigned(declaration) {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return false
	}
	initializer = ast.SkipParentheses(initializer)
	return initializer.Kind == ast.KindObjectLiteralExpression || initializer.Kind == ast.KindArrayLiteralExpression ||
		(initializer.Kind == ast.KindNewExpression && (proof.l.isLibraryGlobal(
			initializer.AsNewExpression().Expression,
			"Map",
		) || proof.l.isLibraryGlobal(initializer.AsNewExpression().Expression, "Set")))
}

func (proof *parallelProof) write(target, owner *ast.Node, chain []string) error {
	target = ast.SkipParentheses(target)
	if _, environment := proof.l.nodeEnvironmentAccess(target); environment {
		return proof.effect(target, chain, "writes the process environment")
	}
	if ast.IsIdentifier(target) {
		declaration := proof.declaration(target)
		if declaration != nil && parallelInside(declaration, owner) {
			return nil
		}
		kind := "capture"
		if declaration != nil {
			kind = "global"
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if ast.IsFunctionLike(parent) {
					kind = "capture"
					break
				}
			}
		}
		return proof.effect(target, chain, "writes the "+kind+" '"+target.Text()+"'")
	}
	if target.Kind == ast.KindPropertyAccessExpression &&
		proof.fresh(target.AsPropertyAccessExpression().Expression, owner) {
		return nil
	}
	if target.Kind == ast.KindElementAccessExpression &&
		proof.fresh(target.AsElementAccessExpression().Expression, owner) {
		return nil
	}
	return proof.effect(target, chain, "writes an item, capture, or object not proven allocated by the task")
}

func (proof *parallelProof) call(node, owner *ast.Node, chain []string) error {
	var callee *ast.Node
	var arguments []*ast.Node
	if node.Kind == ast.KindCallExpression {
		callee = ast.SkipParentheses(node.AsCallExpression().Expression)
		arguments = nodesOf(node.AsCallExpression().Arguments)
	} else {
		callee = ast.SkipParentheses(node.AsNewExpression().Expression)
		arguments = nodesOf(node.AsNewExpression().Arguments)
	}
	// Every argument's evaluation and every callback body must be proved too.
	for _, argument := range arguments {
		if len(
			proof.l.checker.GetSignaturesOfType(proof.l.checker.GetTypeAtLocation(argument), checker.SignatureKindCall),
		) > 0 && proof.functionValue(argument, map[*ast.Symbol]bool{}) == nil {
			return proof.effect(node, chain, "calls a closure whose body isn't known")
		}
		if expressionError := proof.expression(argument, owner, chain); expressionError != nil {
			return expressionError
		}
	}
	if proof.l.isPreludeFunction(callee, "parallelMap") {
		return proof.l.checkParallelMap(node)
	}
	for _, name := range []string{"readTextFile", "writeTextFile", "readDirectory", "fileStatus", "programArguments"} {
		if proof.l.isPreludeFunction(callee, name) {
			return proof.effect(node, chain, "calls effectful "+name)
		}
	}
	for _, name := range []string{"panic", "utf8At", "utf8Length"} {
		if proof.l.isPreludeFunction(callee, name) {
			return nil
		}
	}
	if proof.l.isLibraryGlobal(callee, "Date") {
		return proof.effect(node, chain, "calls effectful Date")
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		receiver, method := access.Expression, access.Name().Text()
		if proof.l.isLibraryGlobal(receiver, "console") || proof.l.isConsole(callee) {
			return proof.effect(node, chain, "calls effectful console."+method)
		}
		if proof.l.isLibraryGlobal(receiver, "Math") && method == "random" {
			return proof.effect(node, chain, "calls effectful Math.random")
		}
		if proof.l.isLibraryGlobal(receiver, "Date") {
			return proof.effect(node, chain, "calls effectful Date."+method)
		}
		if proof.l.isLibraryGlobal(receiver, "Math") ||
			proof.l.isLibraryGlobal(receiver, "Number") ||
			proof.l.isLibraryGlobal(receiver, "String") {
			if proof.l.libraryMember(callee) {
				return nil
			}
		}
		if expressionError := proof.expression(receiver, owner, chain); expressionError != nil {
			return expressionError
		}
		if symbol := proof.l.symbol(callee); symbol != nil &&
			len(symbol.Declarations) > 0 &&
			load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			t := proof.l.checker.GetNonNullableType(proof.l.checker.GetTypeAtLocation(receiver))
			if t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike) != 0 {
				return nil
			}
			if proof.l.isArrayType(t) || proof.l.isLibraryType(t, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
				switch method {
				case "set", "add", "delete", "clear", "push", "pop", "shift", "unshift", "splice", "fill", "sort", "reverse", "copyWithin":
					if proof.fresh(receiver, owner) {
						return nil
					}
					return proof.effect(
						node,
						chain,
						"calls mutating "+method+" on an object not proven allocated by the task",
					)
				case "get", "has", "at", "join", "slice", "concat", "includes", "indexOf", "lastIndexOf", "map", "filter", "find", "findIndex", "every", "some", "forEach", "reduce", "keys", "values", "entries":
					// A callback argument must have a body we can inspect, not merely a Shareable type.
					for _, argument := range arguments {
						if len(proof.l.checker.GetSignaturesOfType(
							proof.l.checker.GetTypeAtLocation(argument),
							checker.SignatureKindCall,
						)) > 0 && proof.functionValue(argument, map[*ast.Symbol]bool{}) == nil {
							return proof.effect(node, chain, "calls a closure whose body isn't known")
						}
					}
					return nil
				}
			}
		}
		return proof.effect(node, chain, "calls a method whose complete dispatch hierarchy isn't proven")
	}
	for _, name := range []string{"String", "Number", "Boolean", "Error", "Map", "Set"} {
		if proof.l.isLibraryGlobal(callee, name) {
			return nil
		}
	}
	if work := proof.functionValue(callee, map[*ast.Symbol]bool{}); work != nil {
		proof.global(callee)
		if ast.IsIdentifier(callee) {
			declaration := proof.declaration(callee)
			if declaration != nil && declaration.Kind != ast.KindFunctionDeclaration && !proof.immutable(declaration) {
				return proof.capture(callee, owner, chain)
			}
		}
		return proof.function(work, chain)
	}
	return proof.effect(node, chain, "calls a closure whose body isn't known")
}

// Analyze an argument with the same rules as a body, retaining its lexical owner.
func (proof *parallelProof) expression(node, owner *ast.Node, chain []string) error {
	if ast.IsPartOfTypeNode(node) {
		return nil
	}
	if ast.IsFunctionLike(node) {
		return proof.function(node, chain)
	}
	if target := parallelWrite(node); target != nil {
		if writeError := proof.write(target, owner, chain); writeError != nil {
			return writeError
		}
	}
	if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {
		return proof.call(node, owner, chain)
	}
	if parallelReference(node) {
		if node.Text() == "undefined" &&
			proof.l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			return nil
		}
		return proof.capture(node, owner, chain)
	}
	var found error
	node.ForEachChild(func(child *ast.Node) bool {
		if found == nil {
			found = proof.expression(child, owner, chain)
		}
		return found != nil
	})
	return found
}

// Property names are not bindings. Shorthand properties are reads of bindings,
// even though the syntax tree also classifies their names as declaration names.
func parallelReference(node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}
	if parent := node.Parent; parent != nil {
		if parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == node {
			return false
		}
		if parent.Kind == ast.KindShorthandPropertyAssignment {
			return true
		}
	}
	return !ast.IsDeclarationName(node) && ast.IsExpressionNode(node)
}

func (proof *parallelProof) symbol(node *ast.Node) *ast.Symbol {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
		return proof.l.checker.GetShorthandAssignmentValueSymbol(parent)
	}
	return proof.l.symbol(node)
}

// Reference-valued globals need runtime marking even though ordinary lowering
// does not put them in a closure's environment. Their bindings are proven const.
func (proof *parallelProof) global(node *ast.Node) {
	declaration := proof.declaration(node)
	if declaration == nil ||
		(declaration.Kind != ast.KindVariableDeclaration && declaration.Kind != ast.KindBindingElement) {
		return
	}
	for parent := declaration.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLike(parent) {
			return
		}
	}
	proof.globals = append(proof.globals, node)
}
