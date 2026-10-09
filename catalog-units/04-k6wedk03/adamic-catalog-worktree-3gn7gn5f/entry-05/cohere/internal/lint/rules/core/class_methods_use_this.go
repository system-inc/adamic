package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ClassMethodsUseThisOptions configures which members are exempt.
//
// Upstream's schema is a single object with four properties. Our config layer unwraps the severity
// tuple before dispatch, so the decoder receives that object directly rather than upstream's
// one-element array.
//
// `EnforceForClassFields` is a POINTER because its default is TRUE. A plain bool decodes an absent
// option to false, which silently narrows the rule to methods only and drops every class-field
// finding; the generic `rule.DecodeOptionsInto` would do exactly that.
type ClassMethodsUseThisOptions struct {
	// ExceptMethods names members that may ignore `this`. A private member is written with its
	// hash, as `#foo`, matching upstream's `hashIfNeeded + name`.
	ExceptMethods []string `json:"exceptMethods"`

	// EnforceForClassFields extends the rule to class fields holding a function. Absent means true.
	EnforceForClassFields *bool `json:"enforceForClassFields"`

	// IgnoreOverrideMethods exempts a member carrying TypeScript's `override` modifier, whose body
	// is constrained by the base class rather than by its own needs.
	IgnoreOverrideMethods bool `json:"ignoreOverrideMethods"`

	// IgnoreClassesWithImplements exempts members of a class that implements an interface, where
	// the signature is likewise imposed from outside. Upstream's enum is `all` or `public-fields`.
	IgnoreClassesWithImplements string `json:"ignoreClassesWithImplements"`
}

// classMethodsUseThisMissingThis builds the finding, naming the member the way upstream does.
//
// The Id is fixed and only the Description moves, so `ExpectFindings` can count these while the
// rendered text carries the computed name. The first sentence is upstream's `Expected 'this' to be
// used by class {{name}}.` verbatim, because that is the half a reader compares against upstream's
// own output.
func classMethodsUseThisMissingThis(name string) rule.Message {
	return rule.Message{
		Id: "missingThis",
		Description: fmt.Sprintf(
			"Expected `this` to be used by class %s. A method that never touches the instance is not "+
				"really a method: it is a plain function that has been given privileged access it "+
				"does not use, and every caller now needs an instance to reach it. Either use the "+
				"instance, make it `static`, or move it out of the class entirely so its independence "+
				"is visible at the call site.",
			name,
		),
	}
}

// ClassMethodsUseThis flags a class method whose body never reads `this`.
//
//	valid:   class A { foo() { this.bar(); } }
//	valid:   class A { static foo() {} }
//	valid:   class A { constructor() {} }
//	valid:   class A { foo() { super.bar(); } }
//	invalid: class A { foo() {} }
//	invalid: class A { foo = () => {}; }
//
// Ported from `class-methods-use-this` in ESLint, read from the clone at
// `lib/rules/class-methods-use-this.js`. Four options, one message, no fixer.
//
// The whole 45-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written.
//
// # Upstream is a stack machine and this tree has no exit hooks, so the shape had to change
//
// Upstream registers `FunctionDeclaration`, `FunctionExpression` and their `:exit` twins, pushing a
// flag on enter and popping it on exit, with `ThisExpression` and `Super` setting the top of the
// stack. There is no `rule.OnExit` here and the walk is pre-order, so that structure is not
// available.
//
// What replaces it is a single `KindSourceFile` listener that walks the file itself. The listener
// fires before its children, so the walk below is the rule's own recursion rather than the engine's,
// and enter/exit become the two sides of one recursive call. That is the shipped pattern for a rule
// that must gather before it can judge.
//
// The stack semantics are preserved exactly, and two of them are load-bearing:
//
//   - A nested function gets its OWN frame, so `foo() { function inner() { this.x; } }` does not
//     count: the inner `this` is a different `this`.
//   - An ARROW function does not, because an arrow has no `this` of its own. So
//     `foo() { return () => this.x; }` does count, and the corpus pins it.
//
// # A static block is its own frame, and that is not the same as being exempt
//
// Upstream pushes a frame for `StaticBlock` and pops it, without ever judging it. The frame is what
// matters: a static block has its own `this`, so a `this` inside one must not satisfy the enclosing
// member. Reproduced here for the same reason.
//
// # `super` counts as using `this`
//
// Upstream's `Super: markThisUsed` is easy to read past. `foo() { super.bar(); }` is CLEAN, because
// a `super` reference is only meaningful against an instance. Both a super call and a super property
// access count.
//
// # TypeScript options with no corpus coverage
//
// `ignoreOverrideMethods` and `ignoreClassesWithImplements` are typescript-eslint concepts that
// upstream's schema accepts and its JavaScript corpus never exercises: no case in the 45 sets
// either. They are implemented against the schema and the reference implementation rather than
// against fixtures, and each carries its own test written here rather than imported, because a
// TypeScript tree is exactly where they matter and upstream had no way to write one.
var ClassMethodsUseThis = rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		resolved := DefaultClassMethodsUseThisOptions()
		if given, isClassMethodsUseThisOptions := rule.OptionsAs[ClassMethodsUseThisOptions](options); isClassMethodsUseThisOptions {
			resolved = given
		}
		enforceForClassFields := true
		if resolved.EnforceForClassFields != nil {
			enforceForClassFields = *resolved.EnforceForClassFields
		}
		exceptMethods := map[string]bool{}
		for _, name := range resolved.ExceptMethods {
			exceptMethods[name] = true
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checker := &classMethodsUseThisChecker{
					ctx:                   ctx,
					exceptMethods:         exceptMethods,
					enforceForClassFields: enforceForClassFields,
					ignoreOverrideMethods: resolved.IgnoreOverrideMethods,
					ignoreWithImplements:  resolved.IgnoreClassesWithImplements,
				}
				checker.walk(node)
			},
		}
	},
}

// classMethodsUseThisChecker carries the walk state that upstream keeps on its stack.
type classMethodsUseThisChecker struct {
	ctx                   rule.Context
	exceptMethods         map[string]bool
	enforceForClassFields bool
	ignoreOverrideMethods bool
	ignoreWithImplements  string

	// stack mirrors upstream's array of booleans: one frame per function-like construct that has
	// its own `this`. An arrow function deliberately pushes NO frame; see the rule doc.
	stack []bool
}

// walk recurses the file, opening a frame for each construct that has its own `this`.
func (checker *classMethodsUseThisChecker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	if classMethodsUseThisMarksThisUsed(node) {
		checker.markThisUsed()
	}

	// A PROPERTY DECLARATION opens a frame of its own, unconditionally, and this is the arm that is
	// easiest to leave out because it is not a function.
	//
	// Upstream spells it `"PropertyDefinition > *.key:exit": pushContext` and `"PropertyDefinition:exit":
	// popContext`, so the frame opens after the key and closes with the member. Its purpose is
	// containment rather than judgment: a class field's initializer is an implicit function with its
	// own `this`, so a `this` written there must not satisfy whatever encloses the class.
	//
	// The corpus case that needs it is `class A { foo () { return class { foo = this }; } }`, where
	// upstream reports the OUTER method. Without this frame the inner field's `this` bubbles up and
	// marks the outer method as satisfied, and that one finding is lost.
	//
	// The `> *.key:exit` half of upstream's selector is equally load-bearing, and getting it wrong
	// costs a finding in the other direction. The frame opens AFTER the key, so a `this` written in
	// a COMPUTED key belongs to whatever encloses the class rather than to the field. That is
	// `class A { foo() { return class { [this.foo] = 1 }; } }`, upstream's valid[21], which is clean
	// precisely because the key's `this` still reaches the outer method. A frame opened at the
	// property instead of after its key swallows it and the outer method reports.
	//
	// Rather than model "after the key", the key is walked OUTSIDE the frame and the rest inside,
	// which is the same thing expressed as a traversal.
	opensFrame := classMethodsUseThisOpensFrame(node) ||
		(checker.enforceForClassFields && classMethodsUseThisIsClassFieldArrow(node))

	if node.Kind == ast.KindPropertyDeclaration {
		checker.walkPropertyDeclaration(node)
		return
	}
	if opensFrame {
		checker.stack = append(checker.stack, false)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		checker.walk(child)
		return false
	})

	if !opensFrame {
		return
	}

	usedThis := checker.stack[len(checker.stack)-1]
	checker.stack = checker.stack[:len(checker.stack)-1]

	if checker.isIncludedInstanceMember(node) && !usedThis {
		checker.ctx.ReportNode(
			classMethodsUseThisReportAnchor(node),
			classMethodsUseThisMissingThis(classMethodsUseThisNameWithKind(node)),
		)
	}
}

// markThisUsed sets the innermost frame, matching upstream's `markThisUsed`.
//
// Upstream guards with `if (stack.length)`, so a `this` at module scope is simply dropped rather
// than crashing. Same guard here, and it is reachable: a top-level `this` is legal in a script.
func (checker *classMethodsUseThisChecker) markThisUsed() {
	if len(checker.stack) == 0 {
		return
	}
	checker.stack[len(checker.stack)-1] = true
}

// classMethodsUseThisOpensFrame answers whether a node has its own `this` binding.
//
// An ARROW FUNCTION is deliberately absent, and it is the single most important line here: an arrow
// closes over the enclosing `this`, so `foo() { return () => this.x; }` must count as using `this`.
// Adding arrows to this list would give that arrow its own frame and lose the finding.
//
// A static block IS present, for the opposite reason: it has its own `this`, so a `this` inside one
// must not satisfy the member around it. It is never itself judged, because
// `isIncludedInstanceMember` declines its kind.
func classMethodsUseThisOpensFrame(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// classMethodsUseThisMarksThisUsed answers upstream's `ThisExpression` and `Super` listeners.
//
// `super` counts, which is easy to read past in the reference implementation: a super reference is
// only meaningful against an instance, so `foo() { super.bar(); }` is clean.
func classMethodsUseThisMarksThisUsed(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindThisKeyword, ast.KindSuperKeyword:
		return true
	}
	return false
}

// isIncludedInstanceMember answers upstream's `isIncludedInstanceMethod`.
//
// The order of the filters is upstream's and matters: `override` and `implements` are consulted
// before the name, and a COMPUTED member returns true before `exceptMethods` is consulted at all,
// because a computed key has no name to compare.
func (checker *classMethodsUseThisChecker) isIncludedInstanceMember(node *ast.Node) bool {
	member := classMethodsUseThisMemberOf(node)
	if member == nil {
		return false
	}
	// A member with no BODY cannot use `this`, and judging it is a false positive rather than a
	// finding. Two TypeScript shapes reach here and upstream's JavaScript corpus contains neither:
	//
	//	abstract parse(value: unknown): TOutput;    an abstract method
	//	is(a: T): this;                             an overload signature, declared beside its
	//	is(a: T): unknown { ... }                   implementation, which IS judged
	//
	// Both are declarations rather than definitions, so `this` is not merely unused but
	// unmentionable. Measured against ESLint on `BaseSchema.ts`: it reports 2 and an earlier draft
	// here reported 8, with the six extra all bodiless. The overload case is the sharper one,
	// because the same NAME then reports twice and reads as a duplicate rather than as a wrong
	// verdict.
	if classMethodsUseThisHasNoBody(member) {
		return false
	}
	if !checker.isInstanceMember(member) {
		return false
	}

	if checker.ignoreOverrideMethods && classMethodsUseThisHasOverrideModifier(member) {
		return false
	}

	if checker.ignoreWithImplements != "" && classMethodsUseThisClassImplements(member) {
		if checker.ignoreWithImplements == "all" {
			return false
		}
		// `public-fields` exempts only members that are publicly visible: not private-named, and
		// either unannotated or explicitly `public`.
		if checker.ignoreWithImplements == "public-fields" &&
			!classMethodsUseThisHasPrivateName(member) &&
			classMethodsUseThisIsPublic(member) {
			return false
		}
	}

	if classMethodsUseThisIsComputed(member) {
		return true
	}

	name, readable := classMethodsUseThisMemberName(member)
	if !readable {
		return true
	}
	return !checker.exceptMethods[name]
}

// isInstanceMember answers upstream's `isInstanceMethod`.
//
// A constructor is exempt by kind, and a static member is exempt by modifier. A property
// declaration counts only when `enforceForClassFields` is on.
func (checker *classMethodsUseThisChecker) isInstanceMember(member *ast.Node) bool {
	if ast.HasStaticModifier(member) {
		return false
	}
	switch member.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	case ast.KindPropertyDeclaration:
		return checker.enforceForClassFields
	}
	return false
}

// classMethodsUseThisMemberOf finds the class member a function-like node belongs to.
//
// A method, accessor or constructor IS the member. A function or arrow assigned to a class field is
// the field's initializer, so the member is its parent -- and only when it is the initializer
// rather than, say, a default value nested somewhere inside it.
func classMethodsUseThisMemberOf(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// `KindMethodDeclaration` covers an object literal's shorthand method as well as a class
		// method, and this rule is about CLASSES: `({ a(){} })` is upstream's valid[7] and must
		// stay clean. The parent kind is what separates them.
		if node.Parent == nil || !ast.IsClassLike(node.Parent) {
			return nil
		}
		return node
	case ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		// Never judged. Upstream exempts the constructor by `kind !== "constructor"` and never
		// calls the check for a static block at all.
		return nil
	case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction:
		if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration {
			if declaration := node.Parent.AsPropertyDeclaration(); declaration != nil &&
				declaration.Initializer == node {
				return node.Parent
			}
		}
	}
	return nil
}

// classMethodsUseThisReportAnchor picks the node the finding points at.
//
// Upstream reports on the FUNCTION and then narrows the location to the function's head with
// `getFunctionHeadLoc`, so the span covers the signature rather than the whole body. For a method
// the closest equivalent here is the member's NAME, which is what a reader is shown.
func classMethodsUseThisReportAnchor(node *ast.Node) *ast.Node {
	if member := classMethodsUseThisMemberOf(node); member != nil {
		if name := member.Name(); name != nil {
			return name
		}
		return member
	}
	return node
}

// classMethodsUseThisMemberName reads a member's name, and says whether it could be read.
//
// The kind guard before the text read is load-bearing rather than defensive: `Node.Text()` panics
// on several kinds rather than returning empty, and the walk recovers per FILE rather than per rule,
// so one such member would cost every rule in this package its verdict on that file.
//
// A private name is returned WITH its hash, matching upstream's `hashIfNeeded + name`, because that
// is the spelling `exceptMethods` entries use.
func classMethodsUseThisMemberName(member *ast.Node) (string, bool) {
	name := member.Name()
	if name == nil {
		return "", false
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
		return name.Text(), true
	case ast.KindPrivateIdentifier:
		// Upstream builds `hashIfNeeded + name` because ESTree's `key.name` for a private member
		// omits the hash. Ours does NOT: `Text()` on a private identifier returns `#foo`, measured
		// directly by printing it from a probe rule rather than inferred.
		//
		// The measurement is recorded because reasoning from mutation results got it backwards once.
		// Removing the `HasPrefix` guard SURVIVES, which reads as dead code and invites deleting the
		// hash handling entirely; adding the hash unconditionally is CAUGHT. Both are consistent with
		// "Text() already carries it", and neither distinguishes that from the opposite claim -- only
		// asking the parser does. The guard is therefore correct and its condition is genuinely
		// always true here, which is why it is kept as a cheap assertion of a fact upstream's own
		// AST does not share.
		text := name.Text()
		if !strings.HasPrefix(text, "#") {
			text = "#" + text
		}
		return text, true
	}
	return "", false
}

// classMethodsUseThisIsComputed answers upstream's `node.computed`.
func classMethodsUseThisIsComputed(member *ast.Node) bool {
	name := member.Name()
	return name != nil && name.Kind == ast.KindComputedPropertyName
}

// classMethodsUseThisHasPrivateName answers `key.type === "PrivateIdentifier"`.
func classMethodsUseThisHasPrivateName(member *ast.Node) bool {
	name := member.Name()
	return name != nil && name.Kind == ast.KindPrivateIdentifier
}

// classMethodsUseThisHasOverrideModifier answers TypeScript's `override` modifier.
func classMethodsUseThisHasOverrideModifier(member *ast.Node) bool {
	return classMethodsUseThisHasModifier(member, ast.KindOverrideKeyword)
}

// classMethodsUseThisIsPublic answers upstream's `!node.accessibility || node.accessibility === "public"`.
func classMethodsUseThisIsPublic(member *ast.Node) bool {
	if classMethodsUseThisHasModifier(member, ast.KindPrivateKeyword) ||
		classMethodsUseThisHasModifier(member, ast.KindProtectedKeyword) {
		return false
	}
	return true
}

// classMethodsUseThisHasModifier answers whether a member carries a given modifier keyword.
func classMethodsUseThisHasModifier(member *ast.Node, kind ast.Kind) bool {
	modifiers := member.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}

// classMethodsUseThisClassImplements answers whether the enclosing class implements an interface.
func classMethodsUseThisClassImplements(member *ast.Node) bool {
	class := member.Parent
	if class == nil {
		return false
	}
	data := class.ClassLikeData()
	if data == nil || data.HeritageClauses == nil {
		return false
	}
	for _, clause := range data.HeritageClauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage == nil {
			continue
		}
		if heritage.Token == ast.KindImplementsKeyword {
			return true
		}
	}
	return false
}

// classMethodsUseThisNameWithKind reproduces upstream's `getFunctionNameWithKind` for the shapes
// this rule can reach.
//
// The rendered forms the corpus asserts, all of which are pinned by fixtures:
//
//	class method 'foo'            an ordinary named method
//	class method                  a computed key, which has no name to print
//	class private method #foo     a private name, printed bare rather than quoted
//	class getter 'quux'           an accessor
//	class setter                  an accessor with a computed key
//	class generator method 'x'    a generator
//
// Note the quoting difference: a private name is printed as `#foo` with no quotes while an ordinary
// name is `'foo'`. That is upstream's, and it is the kind of detail an id-only fixture cannot see,
// which is why the message text is asserted whole.
func classMethodsUseThisNameWithKind(node *ast.Node) string {
	member := classMethodsUseThisMemberOf(node)
	if member == nil {
		return "method"
	}

	tokens := []string{}
	if classMethodsUseThisHasPrivateName(member) {
		tokens = append(tokens, "private")
	}
	if classMethodsUseThisHasModifier(member, ast.KindAsyncKeyword) {
		tokens = append(tokens, "async")
	}
	if classMethodsUseThisIsGenerator(node) {
		tokens = append(tokens, "generator")
	}

	switch member.Kind {
	case ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	default:
		tokens = append(tokens, "method")
	}

	if classMethodsUseThisIsComputed(member) {
		return strings.Join(tokens, " ")
	}
	name, readable := classMethodsUseThisMemberName(member)
	if !readable {
		return strings.Join(tokens, " ")
	}
	if classMethodsUseThisHasPrivateName(member) {
		return strings.Join(tokens, " ") + " " + name
	}
	return strings.Join(tokens, " ") + " '" + name + "'"
}

// classMethodsUseThisIsGenerator answers `node.generator`.
func classMethodsUseThisIsGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		if declaration := node.AsMethodDeclaration(); declaration != nil {
			return declaration.AsteriskToken != nil
		}
	case ast.KindFunctionDeclaration:
		if declaration := node.AsFunctionDeclaration(); declaration != nil {
			return declaration.AsteriskToken != nil
		}
	case ast.KindFunctionExpression:
		if expression := node.AsFunctionExpression(); expression != nil {
			return expression.AsteriskToken != nil
		}
	}
	return false
}

// DefaultClassMethodsUseThisOptions is the unconfigured answer.
//
// `enforceForClassFields` defaults to TRUE, which is why the field is a pointer: an absent option
// must not decode to the zero value.
func DefaultClassMethodsUseThisOptions() ClassMethodsUseThisOptions {
	enforceForClassFields := true
	return ClassMethodsUseThisOptions{EnforceForClassFields: &enforceForClassFields}
}

// DecodeClassMethodsUseThisOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because `enforceForClassFields` defaults to true:
// the generic helper would decode an absent option to false and silently drop every class-field
// finding. A bare `"error"` configuration also arrives as empty input, which the generic decoder
// errors on.
func DecodeClassMethodsUseThisOptions(raw []byte) (any, error) {
	options := DefaultClassMethodsUseThisOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var decoded ClassMethodsUseThisOptions
	if err := rule.UnmarshalOptions(raw, &decoded); err != nil {
		return options, err
	}
	if decoded.EnforceForClassFields == nil {
		decoded.EnforceForClassFields = options.EnforceForClassFields
	}
	return decoded, nil
}

// classMethodsUseThisIsClassFieldArrow answers upstream's
// `"PropertyDefinition > ArrowFunctionExpression.value"` listener.
//
// An arrow normally opens NO frame, because it closes over the enclosing `this` -- that is what
// makes `foo() { return () => this.x; }` clean. But an arrow that IS a class field's value is the
// member's whole body, so upstream registers enter/exit for exactly that position, gated on
// `enforceForClassFields`. Without this, `class A { foo = () => {} }` is never judged and two of
// upstream's reporting cases go silent.
//
// The `.value` half of upstream's selector is load-bearing and reproduced: an arrow nested anywhere
// else inside the field, such as a default argument, is not the field's value and still opens no
// frame.
func classMethodsUseThisIsClassFieldArrow(node *ast.Node) bool {
	if node.Kind != ast.KindArrowFunction || node.Parent == nil ||
		node.Parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	declaration := node.Parent.AsPropertyDeclaration()
	return declaration != nil && declaration.Initializer == node
}

// walkPropertyDeclaration walks a class field the way upstream's two selectors do.
//
// Upstream opens the field's frame at `"PropertyDefinition > *.key:exit"` and closes it at
// `"PropertyDefinition:exit"`, so the key is visited OUTSIDE the frame and everything after it
// inside. Reproduced here as a traversal rather than as a position test: the key is walked first
// with the enclosing frame still current, then the frame opens for the remainder.
//
// Both halves are pinned by the corpus, in opposite directions. `foo = this` inside a nested class
// must NOT satisfy the enclosing method (invalid[19]), and `[this.foo] = 1` inside one MUST
// (valid[21]).
func (checker *classMethodsUseThisChecker) walkPropertyDeclaration(node *ast.Node) {
	declaration := node.AsPropertyDeclaration()
	if declaration == nil {
		node.ForEachChild(func(child *ast.Node) bool {
			checker.walk(child)
			return false
		})
		return
	}

	name := node.Name()
	if name != nil {
		checker.walk(name)
	}

	checker.stack = append(checker.stack, false)
	node.ForEachChild(func(child *ast.Node) bool {
		if child != name {
			checker.walk(child)
		}
		return false
	})
	usedThis := checker.stack[len(checker.stack)-1]
	checker.stack = checker.stack[:len(checker.stack)-1]

	// The field's own frame is opened for CONTAINMENT rather than for judgment: it keeps a `this`
	// written in the initializer from satisfying whatever encloses the class. A field whose value is
	// a function is judged by that function's own frame, inside the walk above, so there is nothing
	// to report here and `usedThis` is deliberately discarded.
	//
	// Upstream is the same shape: its `PropertyDefinition:exit` is a bare `popContext` with no
	// report, while every finding comes from `exitFunction`.
	_ = usedThis
	_ = declaration
}

// classMethodsUseThisHasNoBody answers whether a member is a declaration rather than a definition.
//
// An abstract method, an overload signature, and a method signature in an interface or type literal
// all have a nil body. `ast.IsClassLike` already keeps interface members out of this rule, so what
// this actually excludes is the first two, and both are TypeScript-only shapes that upstream's
// corpus has no way to express.
func classMethodsUseThisHasNoBody(member *ast.Node) bool {
	switch member.Kind {
	case ast.KindMethodDeclaration:
		declaration := member.AsMethodDeclaration()
		return declaration == nil || declaration.Body == nil
	case ast.KindGetAccessor:
		accessor := member.AsGetAccessorDeclaration()
		return accessor == nil || accessor.Body == nil
	case ast.KindSetAccessor:
		accessor := member.AsSetAccessorDeclaration()
		return accessor == nil || accessor.Body == nil
	}
	return false
}
