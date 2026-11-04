package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A class binding is in its temporal dead zone until its declaration runs, as a let's is, and Node
// throws "Cannot access 'Late' before initialization" on a new Late() that runs before then (#aw65psm).
// Stage 0 gives a class no readiness: its constructor is static code, so new Late() inside a function
// works whenever the function runs, and a call that runs before class Late { } is reached prints what
// Node refuses. tsc rejects a use before the declaration at the module's own top level (TS2449), so
// the road is a function, a method or an arrow holding the reference that runs before the declaration.
// It can only run then if the code before the declaration can reach it, so a reference is not yet
// exactly when classReferenceMayRunEarly says the top-level statement holding it is reachable from
// the statements before the class. A reference inside the class's own body can't run before its
// declaration, since nothing can make an instance or reach a static before it. A reference from
// another module is judged where the modules are ordered (import cycles' own checks). Readiness for
// classes lifts this.
func (l *lowering) classDeadZoneGap(node *ast.Node) error {
	if node.Kind != ast.KindIdentifier || ast.IsPartOfTypeNode(node) {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsClass == 0 {
		return nil
	}
	declaration := symbol.ValueDeclaration
	if declaration == nil || declaration.Kind != ast.KindClassDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindSourceFile {
		return nil
	}
	if node == declaration.Name() || ast.GetSourceFileOfNode(node) != ast.GetSourceFileOfNode(declaration) {
		return nil
	}
	inFunction := false
	var top *ast.Node
	for at := node; at.Parent != nil; at = at.Parent {
		if at == declaration {
			return nil
		}
		if ast.IsFunctionLike(at) {
			inFunction = true
		}
		if at.Parent.Kind == ast.KindSourceFile {
			top = at
			break
		}
	}
	if !inFunction || top == nil || top == declaration || !classReferenceMayRunEarly(declaration, top) {
		return nil
	}
	return l.notYet(node, "a class used from a function that may run before the class's declaration (its temporal dead zone)")
}

func init() { refusalRules.register(2050, "classDeadZoneGap", (*lowering).classDeadZoneGap) }

// classReferenceMayRunEarly is whether the function in top, the top-level statement holding a
// reference to declaration's class, may run before declaration does. A statement that isn't a
// declaration and comes before the class may run any function it holds at once, so it may. One after
// the class runs after it. A function declaration, class or variable holding the function runs it only
// when its name is called or reached, and a hoisted function declaration can be called from before
// the class wherever it sits, so it may run early exactly when the code before the class can reach one
// of its names (reachedBefore). An exported one can also be reached through another module, so it may
// whenever the code before the class makes any call at all.
func classReferenceMayRunEarly(declaration, top *ast.Node) bool {
	statements := declaration.Parent.AsSourceFile().Statements.Nodes
	position, at := -1, -1
	for index, statement := range statements {
		if statement == declaration {
			position = index
		}
		if statement == top {
			at = index
		}
	}
	if position < 0 || at < 0 {
		return true
	}
	names := []string{}
	switch top.Kind {
	case ast.KindFunctionDeclaration:
		if name := top.Name(); name != nil {
			names = append(names, name.Text())
		}
	case ast.KindClassDeclaration, ast.KindVariableStatement:
		if at > position {
			// A class or a variable declared after this one doesn't exist before it.
			return false
		}
		names = declaredNames(top)
	default:
		return at < position
	}
	if len(names) == 0 {
		return true
	}
	running := []*ast.Node{}
	for _, statement := range statements[:position] {
		running = append(running, runningParts(statement)...)
	}
	if ast.HasSyntacticModifier(top, ast.ModifierFlagsExport) && anyCall(running) {
		return true
	}
	reached := reachedBefore(statements, running)
	for _, name := range names {
		if reached[name] {
			return true
		}
	}
	return false
}

// declaredNames are the names a top-level function, class or variable statement binds, and none for
// a variable bound by a pattern, which the caller then treats as reachable.
func declaredNames(statement *ast.Node) []string {
	switch statement.Kind {
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		if name := statement.Name(); name != nil {
			return []string{name.Text()}
		}
	case ast.KindVariableStatement:
		names := []string{}
		for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if !ast.IsIdentifier(declaration.Name()) {
				return nil
			}
			names = append(names, declaration.Name().Text())
		}
		return names
	}
	return nil
}

// runningParts are the parts of a top-level statement that run when it's reached: a function
// declaration and an interface, type or import run nothing; a variable statement runs its
// initializers; a class runs its heritage, static initializers, static blocks and computed names; any
// other statement runs whole.
func runningParts(statement *ast.Node) []*ast.Node {
	switch statement.Kind {
	case ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindImportDeclaration, ast.KindImportEqualsDeclaration, ast.KindEmptyStatement:
		return nil
	case ast.KindVariableStatement:
		parts := []*ast.Node{}
		for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
				parts = append(parts, initializer)
			}
		}
		return parts
	case ast.KindClassDeclaration:
		parts := []*ast.Node{}
		for _, clause := range nodesOf(statement.AsClassDeclaration().HeritageClauses) {
			parts = append(parts, clause)
		}
		for _, member := range statement.Members() {
			if member.Name() != nil && member.Name().Kind == ast.KindComputedPropertyName {
				parts = append(parts, member.Name())
			}
			if member.Kind == ast.KindClassStaticBlockDeclaration {
				parts = append(parts, member)
			}
			if member.Kind == ast.KindPropertyDeclaration && ast.HasStaticModifier(member) && member.AsPropertyDeclaration().Initializer != nil {
				parts = append(parts, member.AsPropertyDeclaration().Initializer)
			}
		}
		return parts
	}
	return []*ast.Node{statement}
}

// reachedBefore is every name the code before the class can reach: each identifier its running parts
// mention, and then, until nothing new is found, each identifier in a top-level function,
// class or variable statement anywhere in the module whose name is already reached. Names are
// compared as text, so shadowing only adds names, which only stops more.
func reachedBefore(statements, running []*ast.Node) map[string]bool {
	reached := map[string]bool{}
	for _, part := range running {
		mentionedNames(part, reached)
	}
	expanded := map[*ast.Node]bool{}
	for grew := true; grew; {
		grew = false
		for _, statement := range statements {
			if expanded[statement] {
				continue
			}
			for _, name := range declaredNames(statement) {
				if reached[name] {
					expanded[statement] = true
					mentionedNames(statement, reached)
					grew = true
					break
				}
			}
		}
	}
	return reached
}

func mentionedNames(node *ast.Node, into map[string]bool) {
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if inner.Kind == ast.KindIdentifier {
			into[inner.Text()] = true
		}
		return inner.ForEachChild(visit)
	}
	visit(node)
}

// anyCall is whether any of the running parts makes a call, a new or a tagged template.
func anyCall(statements []*ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if inner.Kind == ast.KindCallExpression || inner.Kind == ast.KindNewExpression || inner.Kind == ast.KindTaggedTemplateExpression {
			found = true
			return true
		}
		return inner.ForEachChild(visit)
	}
	for _, statement := range statements {
		if visit(statement) {
			break
		}
	}
	return found
}
