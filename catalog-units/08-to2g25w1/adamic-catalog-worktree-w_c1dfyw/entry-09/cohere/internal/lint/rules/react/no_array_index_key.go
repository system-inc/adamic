package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoArrayIndexKey = rule.Message{
	Id: "noArrayIndex",
	Description: "This key is the array index the item happens to sit at, which is not an " +
		"identity. When the list is reordered, filtered or has an item inserted, the same key " +
		"lands on a different item, so React reuses the wrong element and whatever state that " +
		"element held goes with it: a checked box, a focused input, a half-typed value. Nothing " +
		"errors, so the bug surfaces later as state attached to the wrong row. Key on something " +
		"belonging to the item instead, usually its identifier.",
}

// noArrayIndexKeyIteratorPositions maps each iterator method to the parameter position its index
// argument occupies.
//
// This is upstream's `iteratorFunctionsToIndexParamPosition` verbatim. The two folding methods take
// an accumulator first, which pushes their index to position two, and getting that wrong is
// invisible in the common case: `foo.reduce((accumulator, item) => ...)` has an index at neither
// position, so a port using position one for everything is silent exactly where upstream is silent
// and wrong on the shape that matters. Both are measured in the fixtures.
var noArrayIndexKeyIteratorPositions = map[string]int{
	"every":       1,
	"filter":      1,
	"find":        1,
	"findIndex":   1,
	"flatMap":     1,
	"forEach":     1,
	"map":         1,
	"reduce":      2,
	"reduceRight": 2,
	"some":        1,
}

// NoArrayIndexKey flags a React key derived from an array index.
//
//	valid:   foo.map((bar) => <Foo key={bar.id} />)
//	valid:   foo.map((bar, i) => <Foo key={bar[i]} />)
//	valid:   <Foo key={i} />                              (not inside an iterator callback)
//	valid:   foo.map((bar, i) => <Foo key="fixed" />)
//	invalid: foo.map((bar, i) => <Foo key={i} />)
//	invalid: foo.map((bar, i) => <Foo key={`x-${i}`} />)
//	invalid: foo.map((bar, i) => <Foo key={String(i)} />)
//	invalid: foo.map((bar, i) => React.createElement("div", { key: i }))
//
// Ported from `react/no-array-index-key` in `eslint-plugin-react`, read from the clone at
// `lib/rules/no-array-index-key.js`. `schema: []` means no options; one message; no fixer. Every
// behaviour below was measured by driving the installed build (7.37.5) through the ESLint Linter
// API rather than reasoned from the source.
//
// # There is no file-suffix gate
//
// Measured under `.tsx`, `.jsx` and `.js`: all three report. Three rules in this package used to
// carry such a gate as oxc residue and it was removed; nothing here reintroduces one.
//
// # The shape is a stack, and this tree has no exit listener
//
// Upstream pushes an index parameter name on entering an iterator call and pops it on exiting,
// then asks whether a key expression names anything currently on the stack. Our walk is pre-order
// with no exit hook, so the whole rule runs inside one `KindSourceFile` listener that walks the
// tree itself and maintains the stack across the recursion. That is the shape
// `no_multi_comp.go` and `no_class_assign.go` both use.
//
// The stack rather than a single name is load-bearing: nested iterators each contribute a name, and
// `foo.map((a, i) => bar.map((b, j) => <Foo key={i} />))` reports on the OUTER index from inside the
// inner callback. Measured.
//
// # Which calls push, and the two-argument form
//
// The callee must be a property access whose property names one of the ten iterator methods above,
// and the callback must be a function expression or arrow with enough parameters to have an index
// at that method's position. Measured silent: `foo.each(...)` and `foo.sort(...)` are not in the
// set, `foo.map((bar) => ...)` has too few parameters, `foo.map(cb)` passes a reference rather than
// a literal function, and `foo["map"](...)` is a computed access whose property is not an
// identifier.
//
// `React.Children.map` and `Children.map` take the collection first and the callback SECOND, so the
// callback argument index shifts. Upstream detects this by asking whether the callee's object is
// named `Children`, or is itself a member expression whose own object is the React pragma. Measured:
// `Children.map(c, (child, i) => ...)` and `React.Children.map(...)` both report, while
// `Other.map(c, (child, i) => ...)` does not, because for any other receiver the callback is read
// from argument zero and argument zero is the collection.
//
// # What counts as using the index, and the four shapes upstream accepts
//
//	key={i}                 the identifier itself
//	key={`x-${i}`}          any template expression naming it
//	key={'x-' + i}          any identifier anywhere in a binary expression tree
//	key={i.toString()}      that exact method name on the index
//	key={String(i)}         that exact function name with the index first
//
// Everything else is silent, and the near misses are the interesting part. Measured:
// `key={Number(i)}` and `key={i.toFixed()}` are both clean, because upstream hardcodes `String` and
// `toString` rather than reasoning about conversion; `key={bar.toString()}` is clean because the
// receiver is not the index; and `key={bar[i]}` is clean because an element access is not one of
// the five shapes, even though the index is plainly in it. A port that generalized any of these
// would report on code upstream leaves alone.
//
// # A template or binary expression reports once PER index reference
//
// Upstream filters the expression list and reports inside a `forEach`, so a template naming the
// index twice reports twice, both times on the whole template. Measured, and no corpus case has
// more than one reference.
//
// # createElement and cloneElement
//
// A call to `React.createElement` or `React.cloneElement` with at least two arguments has its
// second argument read as a props object, and a `key` property in it is checked the same way. This
// only applies inside an iterator callback: the same call at the top level is silent, because the
// stack is empty. Measured, including that a spread-only props object is silent and a computed
// `["key"]` property is silent.
//
// Upstream also accepts a bare `cloneElement` imported from `react`, resolved through its own
// variable index. The checker answers the same question: a bare identifier qualifies when it
// resolves to an import whose module specifier is exactly `react`. Measured both ways, with an
// import from another module silent.
//
// # Where the finding points
//
// Four of the five shapes report the whole key expression; `String(i)` reports the ARGUMENT rather
// than the call, which is upstream passing `node.arguments[0]` at that one site and the whole node
// at the other four. A message-id fixture cannot see the difference and the corpus does not assert
// it, so the span cases below do.
var NoArrayIndexKey = rule.Rule{
	Name:             "react/no-array-index-key",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			// One listener over the whole file, because the rule is a stack and there is no exit
			// hook to pop it with. `KindSourceFile` fires before its children, so the walk below is
			// the rule's own rather than the harness's.
			ast.KindSourceFile: func(node *ast.Node) {
				walker := &noArrayIndexKeyWalker{ctx: ctx}
				walker.walk(node)
			},
		}
	},
}

// noArrayIndexKeyWalker carries the index-name stack across the recursion.
type noArrayIndexKeyWalker struct {
	ctx rule.Context

	// indexNames is every index parameter name currently in scope, outermost first. A name is
	// pushed on entering an iterator call whose callback declares one and popped on leaving it.
	indexNames []string
}

// walk visits one node, pushing and popping the stack around an iterator call.
func (w *noArrayIndexKeyWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	pushed := false
	switch node.Kind {
	case ast.KindCallExpression:
		if w.checkCreateElementCall(node) {
			// Upstream returns early after handling a createElement call, so such a call never
			// also pushes an index name. Reproduced: the two arms are exclusive.
			break
		}
		if name, ok := w.indexParameterName(node); ok {
			w.indexNames = append(w.indexNames, name)
			pushed = true
		}

	case ast.KindJsxAttribute:
		w.checkAttribute(node)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})

	if pushed {
		w.indexNames = w.indexNames[:len(w.indexNames)-1]
	}
}

// isArrayIndex reports whether a node is an identifier naming an index parameter in scope.
func (w *noArrayIndexKeyWalker) isArrayIndex(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	text := node.Text()
	for _, name := range w.indexNames {
		if name == text {
			return true
		}
	}
	return false
}

// checkAttribute handles a `key={...}` attribute on a JSX element.
func (w *noArrayIndexKeyWalker) checkAttribute(node *ast.Node) {
	if len(w.indexNames) == 0 {
		// Not inside any iterator callback, so nothing can be an index. Upstream checks this first
		// too, and it is what makes `<Foo key={i} />` at the top level clean.
		//
		// The check is SUBSUMED here and is kept as a cost guard rather than as a behavioural one.
		// A mutation forcing it always-false survived the whole fixture set, correctly: every one
		// of the five report sites in `checkKeyExpression` is gated on `isArrayIndex`, whose loop
		// over an empty slice answers false, and the one non-nil return in
		// `binaryIdentifiersNamingAnIndex` is gated on the same call. Enumerated by grep over both
		// functions rather than by reading, and no input can distinguish the two versions.
		//
		// It stays because most JSX attributes in a real tree sit outside any iterator, and this
		// returns before reading a name or an initializer for all of them. The verdict names the
		// report sites that existed when it was taken; a sixth report site not routed through
		// `isArrayIndex` would void it.
		return
	}
	attribute := node.AsJsxAttribute()
	name := attribute.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "key" {
		return
	}
	initializer := attribute.Initializer
	if initializer == nil || initializer.Kind != ast.KindJsxExpression {
		// `key="foo"` or a bare `key`. Both silent upstream.
		return
	}
	w.checkKeyExpression(initializer.AsJsxExpression().Expression)
}

// checkKeyExpression is upstream's `checkPropValue`, the five shapes it accepts and nothing else.
func (w *noArrayIndexKeyWalker) checkKeyExpression(node *ast.Node) {
	if node == nil {
		return
	}

	// `key={i}`.
	if w.isArrayIndex(node) {
		w.ctx.ReportNode(node, messageNoArrayIndexKey)
		return
	}

	switch node.Kind {
	case ast.KindTemplateExpression:
		// `key={`foo-${bar}`}`. Upstream reports once per index reference, always on the whole
		// template rather than on the reference.
		template := node.AsTemplateExpression()
		if template.TemplateSpans == nil {
			return
		}
		for _, span := range template.TemplateSpans.Nodes {
			if span.Kind != ast.KindTemplateSpan {
				continue
			}
			if w.isArrayIndex(span.AsTemplateSpan().Expression) {
				w.ctx.ReportNode(node, messageNoArrayIndexKey)
			}
		}
		return

	case ast.KindBinaryExpression:
		// `key={'foo' + bar}`. Upstream collects every identifier in the binary tree and reports
		// once per one that names an index, always on the whole expression.
		//
		// A LOGICAL operator is excluded, and this is a parser difference that costs findings in
		// the reporting direction rather than the silent one. In ESTree `||`, `&&` and `??` are a
		// `LogicalExpression`, a node type distinct from `BinaryExpression`, and upstream's
		// `checkPropValue` has an arm only for the latter. Our parser gives all of them
		// `KindBinaryExpression`, so without this test `key={item.id ?? index}` reports here and is
		// silent upstream.
		//
		// Found on the real tree rather than in a fixture: the dry run produced four findings
		// ESLint does not, all of this shape, and `key={item.id ?? index}` is a completely ordinary
		// thing to write. Measured against the installed build: `??`, `||` and `&&` in a key are
		// all silent, and `+` reports.
		//
		// Upstream's own `getIdentifiersFromBinaryExpression` does carry a recursion arm for
		// logical expressions, which reads as though they were meant to be covered. They are not
		// reachable from the dispatch above, so that arm is dead in upstream. Reproduced as
		// upstream behaves rather than as upstream appears to intend.
		// This decline and the one inside `binaryIdentifiersNamingAnIndex` are a PAIR, and each
		// survives mutation alone. Removing either one leaves the other declining the same inputs:
		// the dispatch's decline stops a logical key before the walk, and the walk's decline stops
		// it at the top of the tree, so both routes end in silence. Mutating both at once is caught
		// by seven lines. Kept as a pair rather than collapsed, because they answer at different
		// depths and a nested shape reaches only the second.
		if noArrayIndexKeyIsLogicalOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			return
		}
		for range w.binaryIdentifiersNamingAnIndex(node) {
			w.ctx.ReportNode(node, messageNoArrayIndexKey)
		}
		return

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := call.Expression
		if callee == nil {
			return
		}

		// `key={bar.toString()}`.
		if callee.Kind == ast.KindPropertyAccessExpression {
			access := callee.AsPropertyAccessExpression()
			methodName := access.Name()
			if w.isArrayIndex(access.Expression) &&
				methodName != nil && methodName.Kind == ast.KindIdentifier &&
				methodName.Text() == "toString" {
				w.ctx.ReportNode(node, messageNoArrayIndexKey)
			}
			return
		}

		// `key={String(bar)}`. This is the one site that reports the ARGUMENT rather than the
		// whole call, which the span fixtures pin.
		if callee.Kind == ast.KindIdentifier && callee.Text() == "String" &&
			call.Arguments != nil && len(call.Arguments.Nodes) > 0 &&
			w.isArrayIndex(call.Arguments.Nodes[0]) {
			w.ctx.ReportNode(call.Arguments.Nodes[0], messageNoArrayIndexKey)
		}
	}
}

// binaryIdentifiersNamingAnIndex returns every identifier in a binary expression tree that names an
// index parameter.
//
// Upstream's `getIdentifiersFromBinaryExpression` recurses only through binary expressions and
// returns null for anything else, so `'x-' + foo(i)` contributes nothing: the call is not an
// identifier and not a binary expression. Reproduced rather than widened.
func (w *noArrayIndexKeyWalker) binaryIdentifiersNamingAnIndex(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindIdentifier {
		if w.isArrayIndex(node) {
			return []*ast.Node{node}
		}
		return nil
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := node.AsBinaryExpression()
	if noArrayIndexKeyIsLogicalOperator(binary.OperatorToken.Kind) {
		// A logical operator is a different node type upstream, so its subtree is not walked. See
		// the dispatch above for the measurement.
		return nil
	}
	found := w.binaryIdentifiersNamingAnIndex(binary.Left)
	return append(found, w.binaryIdentifiersNamingAnIndex(binary.Right)...)
}

// indexParameterName returns the name of the index parameter an iterator call's callback declares.
//
// Nil-safe at every step, and the steps are where a port panics rather than where it mis-decides.
// A callee can be a call expression (`getFoo().map(...)`), an element access (`foo["map"](...)`) or
// a bare identifier, and `Text()` panics on the first two, so the KIND is checked before the text
// is read every time. Measured: the computed form is silent upstream and the call-as-callee form
// reports, so neither may crash and they must answer differently.
func (w *noArrayIndexKeyWalker) indexParameterName(node *ast.Node) (string, bool) {
	call := node.AsCallExpression()
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}

	access := callee.AsPropertyAccessExpression()
	methodName := access.Name()
	if methodName == nil || methodName.Kind != ast.KindIdentifier {
		// The kind half of this is EQUIVALENT rather than load-bearing, and it is kept because the
		// reason is not obvious from the line.
		//
		// A property access's name can be a private identifier here, so `foo.#map(...)` is a real
		// parse rather than a hypothetical. It is not a panic hazard: probed, `Text()` on a private
		// identifier returns `#map` rather than crashing, unlike `Text()` on a property access or a
		// call expression, which is why the CALLEE kind test above is a genuine crash guard and this
		// one is not.
		//
		// No input can distinguish the two versions, and the argument is one sentence: every key in
		// the iterator table is a bare method name, and a private identifier's text always carries a
		// leading hash, so the lookup misses either way. A mutation removing the kind test survived
		// the fixtures, correctly.
		return "", false
	}
	position, known := noArrayIndexKeyIteratorPositions[methodName.Text()]
	if !known {
		return "", false
	}

	// `React.Children.map(collection, callback)` and `Children.map(...)` take the callback second.
	callbackIndex := 0
	if noArrayIndexKeyIsReactChildrenCall(access) {
		callbackIndex = 1
	}

	if call.Arguments == nil || len(call.Arguments.Nodes) <= callbackIndex {
		return "", false
	}
	callback := call.Arguments.Nodes[callbackIndex]
	if callback == nil {
		return "", false
	}

	var parameters *ast.NodeList
	switch callback.Kind {
	case ast.KindArrowFunction:
		parameters = callback.AsArrowFunction().Parameters
	case ast.KindFunctionExpression:
		parameters = callback.AsFunctionExpression().Parameters
	default:
		// A reference rather than a literal function. `foo.map(cb)` is silent upstream.
		return "", false
	}

	if parameters == nil || len(parameters.Nodes) < position+1 {
		return "", false
	}

	parameter := parameters.Nodes[position].AsParameterDeclaration()
	if parameter.Initializer != nil {
		// A defaulted parameter, and this is a PARSER difference rather than a rule decision.
		//
		// Upstream reads `params[n].name`, and ESTree wraps a defaulted parameter in an
		// `AssignmentPattern` whose `.name` is undefined, so `foo.map((bar, i = 0) => ...)` is
		// silent there. Measured on the installed build. Our parser keeps the identifier as the
		// parameter's name and hangs the default off `Initializer`, so the name is readable and a
		// port that only checked the name kind would report where upstream does not.
		//
		// Probed rather than assumed: `(bar, i = 0)` gives `name=KindIdentifier(i) init=yes`, while
		// `(bar, {i})` gives `name=KindObjectBindingPattern`. Two different shapes upstream
		// collapses into one absent name.
		return "", false
	}
	if parameter.DotDotDotToken != nil {
		// A rest parameter is an ESTree `RestElement`, also without a `.name`. It cannot be an
		// index in any case, since it collects every remaining argument.
		return "", false
	}

	parameterName := parameter.Name()
	if parameterName == nil || parameterName.Kind != ast.KindIdentifier {
		// A destructuring pattern has no name upstream either, and `params[n].name` is undefined
		// there. Measured: `foo.map((bar, {i}) => <Foo key={i} />)` is silent.
		return "", false
	}
	return parameterName.Text(), true
}

// noArrayIndexKeyIsReactChildrenCall reports whether a callee is the React children iterator, whose
// callback is the second argument rather than the first.
//
// Upstream restricts this to `map` and `forEach`, which is narrower than the ten-method set the
// caller has already matched, so `Children.filter(...)` takes the ordinary argument order.
func noArrayIndexKeyIsReactChildrenCall(access *ast.PropertyAccessExpression) bool {
	methodName := access.Name()
	if methodName == nil || methodName.Kind != ast.KindIdentifier {
		return false
	}
	if methodName.Text() != "map" && methodName.Text() != "forEach" {
		return false
	}

	receiver := access.Expression
	if receiver == nil {
		return false
	}

	// `Children.map(...)`.
	if receiver.Kind == ast.KindIdentifier {
		return receiver.Text() == "Children"
	}

	// `React.Children.map(...)`: the receiver is itself a property access whose own object is the
	// React pragma. Upstream never checks that the middle name is `Children` here, only that the
	// outermost object is React, and that looseness is reproduced.
	if receiver.Kind == ast.KindPropertyAccessExpression {
		inner := receiver.AsPropertyAccessExpression().Expression
		return inner != nil && inner.Kind == ast.KindIdentifier && inner.Text() == "React"
	}
	return false
}

// checkCreateElementCall handles `React.createElement(type, props)` and its clone counterpart, and
// reports whether the call was one.
//
// The boolean is what makes the two arms exclusive: upstream returns early after handling such a
// call, so a `createElement` call never also pushes an index name.
func (w *noArrayIndexKeyWalker) checkCreateElementCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if !w.isCreateOrCloneElement(call.Expression) {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) <= 1 {
		return false
	}
	if len(w.indexNames) == 0 {
		// Upstream checks the stack before reading the props, so a createElement call outside any
		// iterator is silent AND still counts as handled. Measured silent.
		return true
	}

	properties := call.Arguments.Nodes[1]
	if properties == nil || properties.Kind != ast.KindObjectLiteralExpression {
		return true
	}

	list := properties.AsObjectLiteralExpression().Properties
	if list == nil {
		return true
	}
	for _, property := range list.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			// A spread element carries no key, which is upstream's `{ ...foo }` comment.
			continue
		}
		assignment := property.AsPropertyAssignment()
		name := assignment.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "key" {
			// A computed `["key"]` is a KindComputedPropertyName here and has no identifier text,
			// which matches upstream reading `prop.key.name` off a non-Identifier as undefined.
			// Measured silent.
			continue
		}
		w.checkKeyExpression(assignment.Initializer)
	}
	return true
}

// isCreateOrCloneElement reports whether a callee names React's element factory.
//
// Two forms, matching upstream. A property access whose object is the React pragma and whose
// property is `createElement` or `cloneElement`; or a bare identifier that resolves to an import
// from the module `react` specifically. The second is upstream's variable lookup, answered here
// through the checker, and measured both ways: an import of `cloneElement` from `react` reports and
// the same name imported from another module does not.
func (w *noArrayIndexKeyWalker) isCreateOrCloneElement(callee *ast.Node) bool {
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		receiver := access.Expression
		if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return false
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.Text() == "createElement" || name.Text() == "cloneElement"

	case ast.KindIdentifier:
		return w.resolvesToReactImport(callee)
	}
	return false
}

// resolvesToReactImport reports whether an identifier was imported from the module `react`.
//
// Upstream asks its own variable index for an `ImportSpecifier` and reads the parent import's
// source. The checker resolves the same binding; the declaration is an import specifier whose
// import declaration carries the module specifier.
//
// Note that upstream checks the SOURCE and not the imported name, so any identifier imported from
// `react` qualifies here, not only `createElement` and `cloneElement`. That is upstream being loose
// rather than upstream being right, and it is reproduced: the name has already been used as a
// callee with a props object, so the shapes that reach this are narrow in practice.
func (w *noArrayIndexKeyWalker) resolvesToReactImport(identifier *ast.Node) bool {
	symbol := w.ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindImportSpecifier {
			continue
		}
		specifier := noArrayIndexKeyImportModuleSpecifier(declaration)
		if specifier != nil && specifier.Kind == ast.KindStringLiteral &&
			specifier.Text() == "react" {
			return true
		}
	}
	return false
}

// noArrayIndexKeyImportModuleSpecifier walks from an import specifier up to its declaration's module
// string, or nil.
//
// The chain is specifier to named-bindings to import-clause to import-declaration, and every step
// is checked rather than assumed, because a malformed import recovers into shapes well-formed source
// never produces.
func noArrayIndexKeyImportModuleSpecifier(specifier *ast.Node) *ast.Node {
	for current := specifier; current != nil; current = current.Parent {
		if current.Kind == ast.KindImportDeclaration {
			return current.AsImportDeclaration().ModuleSpecifier
		}
	}
	return nil
}

// noArrayIndexKeyIsLogicalOperator reports whether an operator makes this an ESTree
// `LogicalExpression` rather than a `BinaryExpression`.
//
// The distinction has no meaning in our tree, where every one of these is a `KindBinaryExpression`,
// and it decides the rule's verdict because upstream dispatches on the node type. Measured against
// the installed build: a key built with `+` reports and one built with `??`, `||` or `&&` does not.
func noArrayIndexKeyIsLogicalOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindAmpersandAmpersandToken,
		ast.KindBarBarToken,
		ast.KindQuestionQuestionToken:
		return true
	}
	return false
}
