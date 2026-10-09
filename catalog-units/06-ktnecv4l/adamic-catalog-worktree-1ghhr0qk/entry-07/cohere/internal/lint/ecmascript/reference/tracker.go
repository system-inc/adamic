package reference

import (
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The global reference tracker is a port of eslint-utils' ReferenceTracker.iterateGlobalReferences,
// at @eslint-community/eslint-utils 4.10.1, the version ESLint 10.8.1 installs (#jjfa7qb).
//
// Eight of ESLint's core rules find what they report through it: no-obj-calls, prefer-regex-literals,
// prefer-object-spread, no-misleading-character-class, no-useless-backreference,
// prefer-named-capture-group, prefer-exponentiation-operator and require-unicode-regexp. Each was ported
// with its own local approximation of the question, and two approximations of one question drift
// permanently, so the question lives here and each rule converges on it in its own unit.
//
// The question: where in this file is a global, or a member path off one, read, called or
// constructed, following it through the names it is copied into. Starting from every reference to the
// global, the tracker climbs to the parent and acts on what it finds:
//
//	a.b             a member named in the trace map extends the path
//	a()  new a()    a call or construction of a traced path, when the map asks for it
//	x = a           every read of x is followed too, and so is the assignment's own value
//	const x = a     every read of x, or of the names an object pattern takes from it
//	(x = a) => {}   a default value is followed into its binding
//
// on the way passing through what hands the value on unchanged: parentheses, either branch of a
// conditional, either side of `&&`, `||` and `??`, the last expression of a comma, and the type-only
// wrappers `as`, `satisfies`, `<T>x`, `x!` and an instantiation expression.
//
// The union is flow-insensitive, as ESLint's is. `let x = JSON; x = 1; x()` reports the call, because
// one of x's sources is JSON. A global that this file ever writes to is not followed at all, nor is a
// global object such as `globalThis` that it writes to.
//
// # Where it departs from eslint-utils
//
// A compound assignment other than the three logical ones is not followed: `x += JSON` makes x a
// string, so tracing x as JSON would be wrong. eslint-utils follows any assignment operator. The
// difference needs code that adds a namespace object to something, and is recorded rather than copied.
//
// # How "global" is decided without a scope manager
//
// A reference is to the global when the symbol it reads is declared in no source file of the program:
// declared only in a library's declaration files, made by the checker with no declaration at all
// (`globalThis`, `undefined`), or not resolved to anything. The last counts because an undeclared name
// can only ever be the runtime's global: `Temporal()` in a project whose lib predates Temporal still
// calls the global Temporal when there is one. ESLint knows its builtin globals from ecmaVersion and
// its configured globals, and a `/* global */` comment adds one, so ESLint's corpus writes those three
// ways for the same thing; all three reach the same answer here.

// Usage says how a tracked reference uses the value.
type Usage int

const (
	// Read is any read the trace map asks about, at the identifier, member or pattern property.
	Read Usage = iota
	// Call is a call of the traced value, reported at the call.
	Call
	// Construct is a `new` of the traced value, reported at the new expression.
	Construct
)

// TraceMap says which uses of a path are wanted and which members to follow further.
type TraceMap struct {
	Read      bool
	Call      bool
	Construct bool
	Members   map[string]*TraceMap
}

// Tracked is one use the tracker found: the node to report, the global path it holds, and the use.
type Tracked struct {
	Node  *ast.Node
	Path  []string
	Usage Usage
}

// DefaultGlobalObjectNames are the names eslint-utils treats as the global object itself, so that
// `globalThis.JSON` is JSON. Each counts only when it is a global, by the same test as any other.
var DefaultGlobalObjectNames = []string{"global", "globalThis", "self", "window"}

// Tracker answers for one file. Its index of references is built on first use and kept, so a rule
// asking twice pays for the walk once.
type Tracker struct {
	sourceFile        *ast.SourceFile
	typeChecker       *checker.Checker
	globalObjectNames []string

	// references holds every value-reference identifier in the file, by spelling. The checker is
	// asked about an identifier only when a trace reaches its spelling.
	references map[string][]*ast.Node

	// following holds the bindings currently being followed, which is eslint-utils' variableStack:
	// `let a = JSON; a = a;` would otherwise follow a's reads into a forever.
	following []bindingKey

	found []Tracked
}

// bindingKey names a binding being followed: a symbol, or a global by name, since an undeclared
// global has no symbol to hold.
type bindingKey struct {
	symbol *ast.Symbol
	global string
}

// NewTracker prepares a tracker over one file. A nil globalObjectNames takes the defaults.
func NewTracker(sourceFile *ast.SourceFile, typeChecker *checker.Checker, globalObjectNames []string) *Tracker {
	if globalObjectNames == nil {
		globalObjectNames = DefaultGlobalObjectNames
	}
	return &Tracker{sourceFile: sourceFile, typeChecker: typeChecker, globalObjectNames: globalObjectNames}
}

// GlobalReferences returns every use of the globals in traceMap that the map asks for, in source
// order. Two uses starting at one token keep the order they were found in when they share a path, as
// a read of `a.b` comes before the call `a.b()`, and are otherwise ordered by path: `(a ? JSON : Math)()`
// calls both, and comes back as JSON then Math.
func (tracker *Tracker) GlobalReferences(traceMap map[string]*TraceMap) []Tracked {
	tracker.found = nil

	names := make([]string, 0, len(traceMap))
	for name := range traceMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		references, unmodified := tracker.globalReferences(name)
		if !unmodified {
			continue
		}
		tracker.followReads(bindingKey{global: name}, references, []string{name}, traceMap[name], true)
	}

	root := &TraceMap{Members: traceMap}
	for _, name := range tracker.globalObjectNames {
		references, unmodified := tracker.globalReferences(name)
		if !unmodified {
			continue
		}
		tracker.followReads(bindingKey{global: name}, references, nil, root, false)
	}

	found := tracker.found
	tracker.found = nil
	sort.SliceStable(found, func(first, second int) bool {
		firstStart := rule.TokenRange(tracker.sourceFile, found[first].Node).Pos()
		secondStart := rule.TokenRange(tracker.sourceFile, found[second].Node).Pos()
		if firstStart != secondStart {
			return firstStart < secondStart
		}
		return strings.Join(found[first].Path, ".") < strings.Join(found[second].Path, ".")
	})
	return found
}

// index builds the spelling index on first use.
func (tracker *Tracker) index() map[string][]*ast.Node {
	if tracker.references != nil {
		return tracker.references
	}
	tracker.references = map[string][]*ast.Node{}
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && IsValueReference(node) {
			tracker.references[node.Text()] = append(tracker.references[node.Text()], node)
		}
		node.ForEachChild(visit)
		return false
	}
	tracker.sourceFile.AsNode().ForEachChild(visit)
	return tracker.references
}

// globalReferences returns this file's references to the global named, and whether the file leaves
// it unmodified, which is eslint-utils' isModifiedGlobal: a global this file writes to anywhere is not
// followed at all.
func (tracker *Tracker) globalReferences(name string) ([]*ast.Node, bool) {
	var references []*ast.Node
	for _, identifier := range tracker.index()[name] {
		if rule.IsDeclaredInASourceFile(ReadSymbol(tracker.typeChecker, identifier)) {
			continue
		}
		if WritesToBinding(identifier) {
			return nil, false
		}
		references = append(references, identifier)
	}
	return references, len(references) > 0
}

// bindingReferences returns the references to a declared binding: the identifiers of its spelling
// that read its symbol.
func (tracker *Tracker) bindingReferences(name string, symbol *ast.Symbol) []*ast.Node {
	var references []*ast.Node
	for _, identifier := range tracker.index()[name] {
		if ReadSymbol(tracker.typeChecker, identifier) == symbol {
			references = append(references, identifier)
		}
	}
	return references
}

// followReads follows a binding's references, which is eslint-utils' _iterateVariableReferences.
//
// eslint-utils skips a reference that only writes. Nothing here needs to: a write target is never the
// object of a member, the callee of a call, or the value side of an assignment or a declaration, so
// following one finds nothing, and a global with a write among its references is not followed at all.
func (tracker *Tracker) followReads(key bindingKey, references []*ast.Node, path []string, traceMap *TraceMap, reportReads bool) {
	for _, following := range tracker.following {
		if following == key {
			return
		}
	}
	tracker.following = append(tracker.following, key)
	defer func() { tracker.following = tracker.following[:len(tracker.following)-1] }()

	for _, identifier := range references {
		if reportReads && traceMap.Read {
			tracker.add(identifier, path, Read)
		}
		tracker.followValue(identifier, path, traceMap)
	}
}

// followValue climbs from an expression holding a traced value to what uses it, which is
// eslint-utils' _iteratePropertyReferences.
func (tracker *Tracker) followValue(node *ast.Node, path []string, traceMap *TraceMap) {
	for isPassThrough(node) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if parent.Expression() != node {
			return
		}
		key, isStatic := memberKey(parent)
		if !isStatic {
			return
		}
		next := traceMap.Members[key]
		if next == nil {
			return
		}
		nextPath := extend(path, key)
		if next.Read {
			tracker.add(parent, nextPath, Read)
		}
		tracker.followValue(parent, nextPath, next)

	case ast.KindCallExpression:
		if parent.Expression() == node && traceMap.Call {
			tracker.add(parent, path, Call)
		}

	case ast.KindNewExpression:
		if parent.Expression() == node && traceMap.Construct {
			tracker.add(parent, path, Construct)
		}

	case ast.KindBinaryExpression:
		assignment := parent.AsBinaryExpression()
		if assignment.Right != node || !holdsTheAssignedValue(assignment.OperatorToken.Kind) {
			return
		}
		tracker.followTarget(assignment.Left, path, traceMap)
		tracker.followValue(parent, path, traceMap)

	case ast.KindVariableDeclaration:
		if parent.Initializer() == node {
			tracker.followTarget(parent.Name(), path, traceMap)
		}

	// The three spellings of a default value, which ESTree calls one AssignmentPattern: a parameter's,
	// a binding element's, and a shorthand property's in a destructuring assignment.
	case ast.KindParameter, ast.KindBindingElement:
		if parent.Initializer() == node {
			tracker.followTarget(parent.Name(), path, traceMap)
		}
	case ast.KindShorthandPropertyAssignment:
		if parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == node {
			tracker.followTarget(parent.Name(), path, traceMap)
		}
	}
}

// holdsTheAssignedValue reports whether an assignment operator leaves the right side as the target's
// value: plain `=`, and the logical forms when they assign. See the package doc for `+=`.
func holdsTheAssignedValue(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// followTarget follows a traced value into what it is assigned to, which is eslint-utils'
// _iterateLhsReferences: a name, whose reads are followed, or an object pattern, whose properties
// named in the map are. An array pattern is not followed, as eslint-utils does not follow one.
func (tracker *Tracker) followTarget(target *ast.Node, path []string, traceMap *TraceMap) {
	target = ast.SkipParentheses(target)
	switch target.Kind {
	case ast.KindIdentifier:
		symbol := ReadSymbol(tracker.typeChecker, target)
		if symbol == nil {
			// An assignment to an undeclared name creates a global nobody declared, which
			// eslint-utils' findVariable does not find either.
			return
		}
		tracker.followReads(bindingKey{symbol: symbol}, tracker.bindingReferences(target.Text(), symbol), path, traceMap, false)

	case ast.KindObjectBindingPattern:
		for _, element := range target.AsBindingPattern().Elements.Nodes {
			binding := element.AsBindingElement()
			if binding.DotDotDotToken != nil {
				continue
			}
			keyName := binding.PropertyName
			if keyName == nil {
				keyName = binding.Name()
			}
			tracker.followPatternProperty(element, keyName, binding.Name(), path, traceMap)
		}

	case ast.KindObjectLiteralExpression:
		// A destructuring assignment's target, which the parser keeps as an object literal.
		for _, member := range target.AsObjectLiteralExpression().Properties.Nodes {
			switch member.Kind {
			case ast.KindPropertyAssignment:
				tracker.followPatternProperty(member, member.Name(), member.Initializer(), path, traceMap)
			case ast.KindShorthandPropertyAssignment:
				tracker.followPatternProperty(member, member.Name(), member.Name(), path, traceMap)
			}
		}

	case ast.KindBinaryExpression:
		// A default inside a destructuring assignment's target, `{ JSON: j = x } = globalThis`,
		// which ESTree calls an AssignmentPattern and the parser keeps as an assignment.
		if target.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			tracker.followTarget(target.AsBinaryExpression().Left, path, traceMap)
		}
	}
}

// followPatternProperty follows one property of an object pattern when the map names its key.
func (tracker *Tracker) followPatternProperty(property *ast.Node, keyName *ast.Node, value *ast.Node, path []string, traceMap *TraceMap) {
	key, isStatic := patternKey(keyName)
	if !isStatic {
		return
	}
	next := traceMap.Members[key]
	if next == nil {
		return
	}
	nextPath := extend(path, key)
	if next.Read {
		tracker.add(property, nextPath, Read)
	}
	tracker.followTarget(value, nextPath, next)
}

// isPassThrough reports whether a node's parent hands its value on unchanged, which is eslint-utils'
// isPassThrough with parentheses added, since ESTree has no node for them.
func isPassThrough(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
		ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindExpressionWithTypeArguments:
		return true
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == node || conditional.WhenFalse == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return true
		case ast.KindCommaToken:
			return binary.Right == node
		}
	}
	return false
}

// memberKey returns the property a member access names, folding a constant computed key as
// eslint-utils' getStringIfConstant does: `globalThis["JS" + "ON"]` names JSON. A private name keeps
// its `#`, so it never matches a property.
func memberKey(access *ast.Node) (string, bool) {
	if access.Kind == ast.KindPropertyAccessExpression {
		return access.AsPropertyAccessExpression().Name().Text(), true
	}
	return ConstantString(access.AsElementAccessExpression().ArgumentExpression)
}

// patternKey returns the key an object pattern property names: an identifier, a string or number
// literal, or a computed key whose expression is constant.
func patternKey(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}
	if name.Kind == ast.KindComputedPropertyName {
		return ConstantString(name.AsComputedPropertyName().Expression)
	}
	return property.Name(name, property.Named|property.Quoted|property.Numeric)
}

// ConstantString evaluates an expression to the string JavaScript would make of it, where the syntax
// alone settles it, as eslint-utils' getStringIfConstant does without a scope: string, template,
// number and keyword literals, and `+` and template substitutions over those. Anything that needs a
// binding resolved answers false. Numbers are taken as written by the parser, which normalizes them,
// and adding two numbers declines rather than reproduce JavaScript's number formatting.
func ConstantString(expression *ast.Node) (string, bool) {
	text, _, isConstant := constantValue(expression, nil)
	return text, isConstant
}

// constantValue folds an expression, saying also whether the value is a string, which decides what
// `+` does. A nil resolve is the scope-less reading; ConstantStringIn passes one that reads a name's
// binding and a regex literal's text.
func constantValue(expression *ast.Node, resolve func(*ast.Node) (string, bool, bool)) (text string, isString bool, isConstant bool) {
	if expression == nil {
		return "", false, false
	}
	expression = ast.SkipParentheses(expression)
	if resolve != nil && (expression.Kind == ast.KindIdentifier || expression.Kind == ast.KindRegularExpressionLiteral) {
		return resolve(expression)
	}
	switch expression.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return expression.Text(), true, true
	case ast.KindNumericLiteral:
		return expression.Text(), false, true
	case ast.KindTrueKeyword:
		return "true", false, true
	case ast.KindFalseKeyword:
		return "false", false, true
	case ast.KindNullKeyword:
		return "null", false, true
	case ast.KindTemplateExpression:
		template := expression.AsTemplateExpression()
		var builder strings.Builder
		builder.WriteString(template.Head.Text())
		for _, span := range template.TemplateSpans.Nodes {
			text, _, isConstant := constantValue(span.AsTemplateSpan().Expression, resolve)
			if !isConstant {
				return "", false, false
			}
			builder.WriteString(text)
			builder.WriteString(span.AsTemplateSpan().Literal.Text())
		}
		return builder.String(), true, true
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindPlusToken {
			return "", false, false
		}
		left, leftIsString, leftIsConstant := constantValue(binary.Left, resolve)
		right, rightIsString, rightIsConstant := constantValue(binary.Right, resolve)
		if !leftIsConstant || !rightIsConstant || (!leftIsString && !rightIsString) {
			return "", false, false
		}
		return left + right, true, true
	}
	return "", false, false
}

// extend returns path with key appended, never sharing the backing array of a path another branch
// of the walk still holds.
func extend(path []string, key string) []string {
	next := make([]string, len(path), len(path)+1)
	copy(next, path)
	return append(next, key)
}

// add records one use the trace map asked for.
func (tracker *Tracker) add(node *ast.Node, path []string, usage Usage) {
	tracker.found = append(tracker.found, Tracked{Node: node, Path: path, Usage: usage})
}

// String renders a use for test failures and debugging: the path and how it is used.
func (tracked Tracked) String() string {
	usage := map[Usage]string{Read: "Read", Call: "Call", Construct: "Construct"}[tracked.Usage]
	return usage + " " + strings.Join(tracked.Path, ".") + " at " + strconv.Itoa(tracked.Node.Pos())
}
