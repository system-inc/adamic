package next

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoLocationAssign = rule.Message{
	Id: "noLocationAssign",
	Description: "Navigating to an internal page through `location.assign` or by writing " +
		"`location.href` tears down the running application and reloads it from the server, so " +
		"client state, the router's history entry and any in-flight data are all discarded. Use " +
		"`redirect()` in the render phase, or `useRouter().push()` in a Client Component's event " +
		"handler.",
}

// absoluteDestination matches a destination that leaves the application.
//
// A scheme (`https:`, `mailto:`, `tel:`) or a protocol-relative prefix (`//host`) is treated as
// leaving; everything else is internal and reports. Copied from the original's `ABSOLUTE_URL_RE`
// including the case-insensitive flag and including the detail that a scheme must begin with a
// letter, so `1abc:/x` is relative rather than absolute.
var absoluteDestination = regexp.MustCompile(`(?i)^(?:[a-z][0-9+.a-z-]*:|//)`)

// locationRootPrefixes is the set of receivers the original accepts in front of `.location`.
//
// Four names, and the two easy ones to lose are `document` and `self`: the nearest analogue in this
// tree carries only `window` and `globalThis`, so a port that reuses that map drops half the set
// with nothing to notice. `top` and `parent` are real browser globals and are deliberately absent,
// because the original does not carry them either.
var locationRootPrefixes = map[string]bool{
	"window":     true,
	"globalThis": true,
	"document":   true,
	"self":       true,
}

// NoLocationAssignRelativeDestination flags a full page load to an internal destination.
//
//	valid:   location.assign('https://example.com')
//	valid:   router.push('/dashboard')
//	valid:   const target = buildUrl(); location.assign(target)
//	invalid: location.assign('/dashboard')
//	invalid: window.location.href = '/search?term=' + encodeURIComponent(term)
//
// Ported from `@next/next/no-location-assign-relative-destination`. There is no oxc implementation
// of this rule, so the original JavaScript is the entire specification and there is no imported
// corpus behind the fixtures; every case in the test file was established by running the original
// through ESLint's Linter API rather than by predicting its answer from the source.
//
// # What the original decides, and what it does not
//
// Two forms and only two. `location.assign(...)` as a call, and `location.href` as an assignment
// target. `location.replace(...)` is the same hazard and the original does not handle it, so
// neither does this; that gap is reproduced rather than closed, because closing it would report on
// code the framework's own rule accepts. Reading `location.href` is likewise untouched, since the
// original matches it only on the left of an assignment.
//
// The assignment operator is never read. `location.href += '/x'` reports exactly as `=` does, and so
// do `-=`, `||=` and `??=`. That looks like an oversight and it is preserved: measured against the
// original, all four report.
//
// # Resolving the destination
//
// The original asks `getStringIfConstant` first, which is a full static evaluator: it folds
// arithmetic, resolves constants through scope, and evaluates a whitelist of pure builtins, so
// `[”, 'x'].join('/')` and `String(1)` are both in its reach. That surface is deliberately not
// reproduced. What is reproduced is the shape that decides real code: a literal, the text before a
// template's first interpolation, the left side of a `+`, and an identifier resolved to its
// nearest binding.
//
// The consequence of declining the evaluator is a narrower rule rather than a wrong one, and it is
// narrow in one measurable direction. Measured against the original: `location.assign('htt' +
// 'ps://x')` is silent there because the evaluator folds the whole concatenation into an absolute
// url before the left-recursion is ever reached, while this rule sees only the relative left side
// and reports. Two constant strings concatenated into a scheme is not a shape that appears in real
// navigation code, and the alternative is carrying an evaluator to make one contrived input agree.
// Stated here rather than left for a reader to discover.
//
// A template opening with an interpolation has an empty first span, and the empty string is
// relative, so “ location.href = `${base}/path` “ reports whatever `base` actually holds. That is
// an upstream false-positive class and it is reproduced on purpose. Do not narrow it without
// changing the original first.
//
// # Deciding the receiver is the global
//
// The original looks the name up in the global scope's binding set and declines if anything
// declared it. Reproduced here as a walk outward from the use looking for a value binding of the
// name, which answers the same question on a file that does not type-check. The two differ on one
// measured shape: a block-scoped shadow that has already closed. Upstream reads the global scope
// rather than the lexical chain, so `{ const location = {}; } location.assign('/x')` reports there,
// and the walk below agrees because the closed block is not an enclosing scope of the use.
var NoLocationAssignRelativeDestination = rule.Rule{
	// Bare, with no family prefix. The config writes `nextjs/no-location-assign-relative-destination`
	// and the matcher strips the namespace on a `/` boundary, so a prefixed name here matches
	// nothing and the rule runs on no files while every fixture passes.
	Name: "@next/next/no-location-assign-relative-destination",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if !accessesNamedProperty(callee, "assign") {
					return
				}
				if !readsGlobalLocation(memberReceiver(callee)) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				// A spread carries no first argument to read. The original checks this explicitly
				// rather than letting the resolver decline, and the distinction matters: a spread
				// is unknowable, not merely unresolved.
				first := call.Arguments.Nodes[0]
				if first.Kind == ast.KindSpreadElement {
					return
				}
				reportIfRelative(ctx, node, first)
			},

			// Assignment is a binary expression in this tree rather than its own node kind, so the
			// operator has to be recognized rather than implied by the listener.
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if !assignmentOperators[binary.OperatorToken.Kind] {
					return
				}
				target := ast.SkipParentheses(binary.Left)
				if !accessesNamedProperty(target, "href") {
					return
				}
				if !readsGlobalLocation(memberReceiver(target)) {
					return
				}
				reportIfRelative(ctx, node, binary.Right)
			},
		}
	},
}

// assignmentOperators is every operator that writes to its left side.
//
// The original never reads the operator at all, so matching only `=` would drop the other four
// while passing every `=` fixture. Listed rather than derived so the set is readable at the line.
var assignmentOperators = map[ast.Kind]bool{
	ast.KindEqualsToken:                   true,
	ast.KindPlusEqualsToken:               true,
	ast.KindMinusEqualsToken:              true,
	ast.KindAsteriskEqualsToken:           true,
	ast.KindAsteriskAsteriskEqualsToken:   true,
	ast.KindSlashEqualsToken:              true,
	ast.KindPercentEqualsToken:            true,
	ast.KindLessThanLessThanEqualsToken:   true,
	ast.KindAmpersandEqualsToken:          true,
	ast.KindBarEqualsToken:                true,
	ast.KindCaretEqualsToken:              true,
	ast.KindBarBarEqualsToken:             true,
	ast.KindAmpersandAmpersandEqualsToken: true,
	ast.KindQuestionQuestionEqualsToken:   true,
}

// reportIfRelative reports against reported when value resolves to an internal destination.
func reportIfRelative(ctx rule.Context, reported *ast.Node, value *ast.Node) {
	text, ok := staticStringPrefix(value, map[*ast.Node]bool{})
	if !ok || absoluteDestination.MatchString(text) {
		return
	}
	ctx.ReportNode(reported, messageNoLocationAssign)
}

// accessesNamedProperty reports whether node reads the property `name` off something.
//
// Both spellings count, `location.assign` and `location['assign']`, which is what the original's
// `isMemberExprWithNamedProperty` accepts. A computed key must be a string literal: the original
// tests for an ESTree `Literal`, and a template is not one, so “ location[`assign`] “ is silent
// there. Reproduced by refusing to read a template key even though this tree could.
func accessesNamedProperty(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression().Name()
		return property != nil && property.Kind == ast.KindIdentifier && property.Text() == name
	case ast.KindElementAccessExpression:
		// The subscript is read without skipping parentheses, matching the original, which
		// destructures `expr.property` directly rather than through an accessor that sees through
		// them.
		argument := node.AsElementAccessExpression().ArgumentExpression
		return argument != nil && argument.Kind == ast.KindStringLiteral && argument.Text() == name
	}
	return false
}

// memberReceiver returns the object a member access reads from.
func memberReceiver(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// readsGlobalLocation reports whether node denotes the global `location` object.
//
// Two accepted shapes, matching the original's `getLocationRootIdentifier`: a bare `location`, or
// `<prefix>.location` where the prefix is one of four global names. In both cases the identifier
// that has to be the real global is the outermost one, so a user binding of `window` silences a
// `window.location.href` write just as a binding of `location` silences a bare one.
func readsGlobalLocation(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return node.Text() == "location" && isUnshadowedGlobal(node)
	}
	if !accessesNamedProperty(node, "location") {
		return false
	}
	receiver := ast.SkipParentheses(memberReceiver(node))
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return false
	}
	return locationRootPrefixes[receiver.Text()] && isUnshadowedGlobal(receiver)
}

// isUnshadowedGlobal reports whether an identifier still refers to the ambient global of its name.
//
// The original resolves this through ESLint's scope manager, asking whether the global scope holds
// a binding of the name with no definitions. Reproduced here as a walk outward from the use rather
// than as a type query, because the interesting inputs are files that are mid-edit or that do not
// type-check, and a rule that goes quiet on those is a rule that never fires while someone is
// actually writing the bug.
//
// What counts as shadowing is any value binding of the name in an enclosing scope: a variable, a
// parameter, a function or class declaration, an import, or a catch parameter. A property named
// `location` is not a binding and must not count, which is why this is only ever asked about an
// identifier in receiver position.
func isUnshadowedGlobal(identifier *ast.Node) bool {
	name := identifier.Text()
	for scope := identifier.Parent; scope != nil; scope = scope.Parent {
		if scopeBindsName(scope, name) {
			return false
		}
	}
	return true
}

// scopeBindsName reports whether a single node introduces a value binding called name.
//
// Only the node's own children are examined, so the caller's walk decides the scope chain. A
// declaration nested inside a block that has already closed is therefore invisible here, which is
// the behavior the original has for a different reason: it reads the global scope's binding set
// rather than the lexical chain, and a closed block never contributed to it. Measured to agree.
func scopeBindsName(scope *ast.Node, name string) bool {
	bound := false
	scope.ForEachChild(func(child *ast.Node) bool {
		if bindingIntroducesName(child, name) {
			bound = true
			return true
		}
		return false
	})
	if bound {
		return true
	}
	return declaresParameterNamed(scope, name)
}

// bindingIntroducesName reports whether a statement declares name as a value.
func bindingIntroducesName(node *ast.Node, name string) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		list := node.AsVariableStatement().DeclarationList
		if list == nil {
			return false
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			if bindsIdentifierNamed(declaration.AsVariableDeclaration().Name(), name) {
				return true
			}
		}
	case ast.KindVariableDeclarationList:
		for _, declaration := range node.AsVariableDeclarationList().Declarations.Nodes {
			if bindsIdentifierNamed(declaration.AsVariableDeclaration().Name(), name) {
				return true
			}
		}
	case ast.KindFunctionDeclaration:
		return bindsIdentifierNamed(node.AsFunctionDeclaration().Name(), name)
	case ast.KindClassDeclaration:
		return bindsIdentifierNamed(node.AsClassDeclaration().Name(), name)
	case ast.KindImportDeclaration:
		return importBindsName(node, name)
	}
	return false
}

// bindsIdentifierNamed reports whether a binding name is exactly this identifier.
//
// A destructuring pattern is declined rather than searched, matching the original, which only ever
// compares a plain name. A pattern that happens to bind `location` through a rename is a shape no
// measured input produces and guessing at it would be this rule inventing a decision.
func bindsIdentifierNamed(binding *ast.Node, name string) bool {
	return binding != nil && binding.Kind == ast.KindIdentifier && binding.Text() == name
}

// importBindsName reports whether an import declaration introduces name locally.
//
// All three shapes have to be read, because any of them can bind `location` or `window`, and a
// version reading only one looks complete while silencing half the inputs. `imports.BindingsOf`
// already separates them, and reaching past it into the clause was caught by the accessor guard
// rather than by any fixture: the inline version and the shelf version agree on every measured
// input, so nothing else could have told them apart.
//
// What matters for shadowing is the local name, which for a renamed specifier is the name after
// `as`. `element.Name()` answers that for both `{ location }` and `{ anything as location }`.
func importBindsName(node *ast.Node, name string) bool {
	bindings := imports.BindingsOf(node)
	if bindsIdentifierNamed(bindings.Default, name) {
		return true
	}
	// `Namespace` holds the `NamespaceImport` node rather than the identifier inside it, despite the
	// field's doc comment saying "the local name". Measured: comparing the field directly is always
	// false, so `import * as window from 'm'` would stop shadowing. `no_import_assign.go` reads it
	// the same way, through `.Name()`, which is what settled which of the two the field means.
	if bindings.Namespace != nil && bindsIdentifierNamed(bindings.Namespace.Name(), name) {
		return true
	}
	for _, element := range bindings.Named {
		if bindsIdentifierNamed(element.Name(), name) {
			return true
		}
	}
	return false
}

// declaresParameterNamed reports whether a function-like node or a catch clause binds name.
//
// A parameter is not a child the way a statement is, so it needs its own read. This is the half a
// walk over statements alone silently misses, and a parameter named `window` is the shape that
// shows up in real code: a component that takes a mock global for a test.
func declaresParameterNamed(scope *ast.Node, name string) bool {
	if scope.Kind == ast.KindCatchClause {
		// The catch binding is a `VariableDeclaration` wrapping the name rather than the identifier
		// itself, so comparing the clause's node directly answers false for every input. That is a
		// silent always-false rather than a crash, and the only thing that found it was a fixture
		// written because the mutation sweep said this branch was invisible.
		declaration := scope.AsCatchClause().VariableDeclaration
		if declaration == nil {
			return false
		}
		return bindsIdentifierNamed(declaration.AsVariableDeclaration().Name(), name)
	}
	// `ParameterList` reaches through `FunctionLikeData`, which is nil on every node that is not
	// function-like, so calling it unguarded panics on the first expression statement the walk
	// meets. Asking the shim's own predicate rather than enumerating the function kinds here keeps
	// the two from drifting.
	if !ast.IsFunctionLike(scope) {
		return false
	}
	parameters := scope.ParameterList()
	if parameters == nil {
		return false
	}
	for _, parameter := range parameters.Nodes {
		if bindsIdentifierNamed(parameter.AsParameterDeclaration().Name(), name) {
			return true
		}
	}
	return false
}

// staticStringPrefix resolves the leading text of a destination, when the syntax settles it.
//
// Four branches, mirroring the original minus its constant evaluator: a string or template with no
// interpolation answers its own text, a template with interpolations answers the text before the
// first one, a `+` answers whatever its left side answers, and an identifier answers through its
// nearest binding.
//
// The visited set breaks the cycle in `const a = b; const b = a`. That is not valid code, but it is
// reachable while someone is mid-edit, and the original has no such guard: measured, ESLint's own
// resolver overflows its stack and takes the whole lint run down on exactly that input. Diverging
// here is deliberate, because a linter that crashes on a half-written file is worse than one that
// declines to answer for it.
func staticStringPrefix(node *ast.Node, visited map[*ast.Node]bool) (string, bool) {
	node = ast.SkipParentheses(node)
	if node == nil || visited[node] {
		return "", false
	}
	visited[node] = true

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true

	case ast.KindTemplateExpression:
		// The text before the first interpolation, cooked rather than raw, matching the original's
		// `quasis[0].value.cooked`. An empty head is the common and load-bearing case: empty is
		// relative, so a template opening with an interpolation reports.
		head := node.AsTemplateExpression().Head
		if head == nil {
			return "", false
		}
		return head.Text(), true

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindPlusToken {
			return "", false
		}
		return staticStringPrefix(binary.Left, visited)

	case ast.KindIdentifier:
		initializer := nearestBindingInitializer(node, node.Text())
		if initializer == nil {
			return "", false
		}
		return staticStringPrefix(initializer, visited)
	}

	return "", false
}

// nearestBindingInitializer finds the value most recently written to name before this read.
//
// The original walks the variable's references, taking the last write whose position precedes the
// read and falling back to the declaration's initializer. Reproduced by walking outward from the
// use: within each enclosing scope, the latest assignment to the bare name that starts before the
// read wins, and the declaration's initializer is the floor. That is the same answer on every
// measured input, including the two that distinguish the direction: a write before the read is
// taken, and a write after it is ignored.
//
// Only a variable declaration is a source of a value. A parameter or an import has nothing to read
// and the original declines them explicitly, which is a deliberate decline rather than a gap.
func nearestBindingInitializer(reference *ast.Node, name string) *ast.Node {
	readPosition := reference.Pos()
	for scope := reference.Parent; scope != nil; scope = scope.Parent {
		var declared *ast.Node
		var latestWrite *ast.Node
		latestWritePosition := -1

		scope.ForEachChild(func(child *ast.Node) bool {
			if initializer, ok := declarationInitializerFor(child, name); ok {
				declared = initializer
				return false
			}
			if written, position, ok := assignmentToName(child, name); ok &&
				position < readPosition && position > latestWritePosition {
				latestWrite = written
				latestWritePosition = position
			}
			return false
		})

		if latestWrite != nil {
			return latestWrite
		}
		if declared != nil {
			return declared
		}
		// A binding of the name with no initializer still stops the walk. Continuing outward would
		// read an outer variable that this one shadows, which is a different value.
		if scopeBindsName(scope, name) {
			return nil
		}
	}
	return nil
}

// declarationInitializerFor returns the initializer of a declaration of name in this statement.
//
// The second result separates "declares the name with nothing to read" from "does not declare it",
// which the caller needs in order to stop walking outward rather than read a shadowed outer value.
func declarationInitializerFor(node *ast.Node, name string) (*ast.Node, bool) {
	if node.Kind != ast.KindVariableStatement {
		return nil, false
	}
	list := node.AsVariableStatement().DeclarationList
	if list == nil {
		return nil, false
	}
	for _, declarationNode := range list.AsVariableDeclarationList().Declarations.Nodes {
		declaration := declarationNode.AsVariableDeclaration()
		if bindsIdentifierNamed(declaration.Name(), name) {
			return declaration.Initializer, declaration.Initializer != nil
		}
	}
	return nil, false
}

// assignmentToName returns the value written to a bare identifier by an expression statement.
//
// Only a plain `name = value` counts. A compound assignment reads the old value as well, so what it
// writes is not the right side alone and the original's `writeExpr` would not be a destination this
// rule could resolve.
func assignmentToName(node *ast.Node, name string) (*ast.Node, int, bool) {
	if node.Kind != ast.KindExpressionStatement {
		return nil, 0, false
	}
	expression := ast.SkipParentheses(node.AsExpressionStatement().Expression)
	if expression == nil || expression.Kind != ast.KindBinaryExpression {
		return nil, 0, false
	}
	binary := expression.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsToken {
		return nil, 0, false
	}
	target := ast.SkipParentheses(binary.Left)
	if target == nil || target.Kind != ast.KindIdentifier || target.Text() != name {
		return nil, 0, false
	}
	return binary.Right, target.Pos(), true
}
