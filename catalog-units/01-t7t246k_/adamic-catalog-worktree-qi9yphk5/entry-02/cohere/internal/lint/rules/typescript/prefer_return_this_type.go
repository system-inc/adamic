package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferReturnThisType flags a class method or property function annotated with the class's own
// name where every value it returns is the receiver, because `this` says the same thing and keeps
// saying it in a subclass.
//
//	valid:   class Foo { f(): this { return this; } }
//	valid:   class Foo { f(): Foo { return new Foo(); } }
//	valid:   class Base {} class Derived extends Base { f(): Base { return this; } }
//	invalid: class Foo { f(): Foo { return this; } }
//	invalid: class Foo { f = (): Foo => this; }
//
// The annotation `Foo` and the annotation `this` describe the same values inside `Foo`, and they
// stop agreeing the moment anybody subclasses it. `class Bar extends Foo {}` gives `bar.f()` the
// static type `Foo`, so the chain `bar.f().barOnly()` fails to compile even though the value really
// is a `Bar`. Writing `this` instead makes the return type polymorphic and the chain works. That is
// why upstream ships this as a fix rather than a suggestion: the rewrite widens what the caller can
// do and takes nothing away.
//
// # The judgment is type identity, not name matching, and both halves are load-bearing
//
// Upstream compares the returned expression's type against two REFERENCE-identical types: the
// class's own declared type, and that type's `thisType`. Name matching would be wrong in both
// directions and the corpus pins each:
//
//	class Base {} class Derived extends Base { f(): Base { return this; } }
//	                          silent: `Base` names a real supertype, and rewriting to `this` would
//	                          narrow the signature rather than widen it
//	class Foo { f(): Foo { const self = this; return self; } }
//	                          REPORTS: `self` has no annotation so it keeps the `this` type, and the
//	                          name in the return statement is not `this` at all
//
// So the rule reads the returned expression's type through the checker rather than its syntax, and
// the only syntactic shortcut it keeps is upstream's own: a bare `this` keyword is accepted without
// asking, which is a cost optimization and not a separate judgment.
//
// # Two returns disagree, and the class type wins
//
// A body that returns both `this` and a fresh instance is silent, and upstream's bookkeeping says
// why: `hasReturnClassType` is checked with priority over `hasReturnThis`, so one return of the
// class type disqualifies the whole method however many returns of `this` sit beside it. Measured:
//
//	class Foo { f(): Foo { if (1) return new Foo(); return this; } }        silent
//	class Foo { f(): Foo { if (1) return; return this; } }                  REPORTS
//
// The second one reports because a bare `return;` has no expression at all, so it is skipped rather
// than counted as returning something that is not `this`. The union arm does the same job for
// `Foo | string`: a returned value whose type is a union containing the class type also sets
// `hasReturnClassType`, which is what makes upstream's last passing case silent while its
// near-identical last failing case reports. The two differ only in whether the union member is the
// class or a primitive.
//
// # Where the returns are looked for, and where they deliberately are not
//
// `ast.ForEachReturnStatement` descends through statement containers only: blocks, `if`, loops,
// `switch`, `try`, labels. It does not enter a nested function or arrow, and neither does
// upstream's `forEachReturnStatement`, whose case list this matches statement for statement.
// Verified by reading both: `TypeScript/tsc/internal/ast/utilities.go:1157` and
// `typescript-eslint/packages/eslint-plugin/src/util/astUtils.ts:52` carry the same fifteen kinds,
// because both descend from the TypeScript compiler's own helper.
//
// That shared omission is a behavior rather than an oversight, and it is the right one: a `return
// this` inside a callback returns from the callback, not from the method. Measured on the installed
// 8.67.0 build, `class Foo { f(): Foo { [1].forEach(() => { return this; }); return this; } }`
// reports on the strength of its OUTER return, and removing that outer return makes it silent.
//
// # Parentheses, and the reason this rule reads TWO trees
//
// This is the one part of the port that is not a transcription, and getting it wrong cost this port
// two false positives before a probe caught them.
//
// typescript-eslint's rule straddles two trees. It reads the estree tree for the return TYPE, where
// there is no parenthesized-expression node at all, and it reads the TYPESCRIPT tree for the return
// VALUE, through `services.esTreeNodeToTSNodeMap`, where parentheses survive exactly as they do
// here. So a parenthesis is invisible on one side and load-bearing on the other, and our parser
// matches upstream on the second side already.
//
// What that means at each site, all measured against the installed 8.67.0 build:
//
//	RETURN VALUE, expression position: DO NOT unwrap.
//	  Upstream's `expr.kind === ts.SyntaxKind.ThisKeyword` is tested on the TypeScript node, so
//	  `return (this);` is not that keyword upstream either and falls through to the type comparison.
//	  In an ordinary method that comparison succeeds and the case reports; in a static body or a
//	  function expression it fails and the case is SILENT. An earlier revision of this rule
//	  unwrapped here, which made the keyword arm fire on shapes the comparison would have declined,
//	  and reported `static f(): Foo { return (this); }` and
//	  `f = function (): Foo { return (this); }` where upstream reports nothing. Both are now pinned
//	  as silent, beside their unparenthesized siblings which report.
//
//	CONCISE BODY of an arrow: unwrap.
//	  `func.body` on the TypeScript node is the parenthesized expression, and upstream compares the
//	  CHECKER's answer for it rather than testing a node kind. A parenthesized expression has the
//	  type of its operand, so upstream's comparison sees through the parens and ours must too. This
//	  is a checker question, so the unwrap is inert for the verdict and kept only so the code reads
//	  the same as the arm above; `(): Foo => (this)` reports either way.
//
//	RETURN TYPE, type position: unwrap, and it is required.
//	  `f(): (Foo | undefined) | null` goes silent without it, because the search here is SYNTACTIC:
//	  it walks type nodes looking for an identifier, and a parenthesized type node is not a type
//	  reference however transparent it is to the checker. A mutant disabling it fails a fixture.
//
// The rule that falls out is worth more than the three instances: unwrap before a question you put
// to the CHECKER, never before a question you put to a node KIND, because upstream's own keyword
// tests run on a tree that kept the parentheses.
//
// The unwraps that remain are loops rather than single steps, because `((this))` nests, and are
// written out rather than routed through `ast.SkipParentheses`, which dereferences its argument
// while a concise body is optional.
//
// # What upstream does not catch, reproduced rather than improved
//
// An intersection is invisible to upstream's `tryGetNameInType`, which recurses through unions and
// nothing else, so `f(): Foo & {}` is silent even though it is the same mistake. Measured on the
// installed build and reproduced here. Widening the search to intersections would be a defensible
// rule and it would not be this rule.
//
// A qualified name is likewise invisible: `f(): N.Foo` inside `namespace N` is silent because the
// type node's `typeName` is a qualified name rather than an identifier. Measured.
//
// # What the repair writes
//
// One edit replacing the matched type reference with `this`, and the reference's own text includes
// its type arguments: `Animal<T>` becomes `this`, not `this<T>`. Upstream anchors the report on the
// same node it rewrites, so the span and the edit are the same range and a reader is always shown
// the text that is about to change.
//
// A union with the class name written twice reports ONCE and rewrites ONCE, leaving the second
// occurrence. `f(): Foo | Foo` becomes `f(): this | Foo` in a single pass. Upstream's own tester
// shows `this | this` only because `cohereAndFix` re-lints until the source stops changing;
// measured against a single application of the same messages, upstream writes `this | Foo` too.
// This port is a single pass and matches it.
var PreferReturnThisType = rule.Rule{
	Name: "@typescript-eslint/prefer-return-this-type",

	// Every discrimination this rule makes is a type question. Name matching alone reports the
	// supertype case upstream is silent on and misses the aliased-local case it reports.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// checkFunction is upstream's `checkFunction`: the class must have a name, the function must
		// carry a return type, that return type must mention the class name, and the function must
		// really return only `this`.
		checkFunction := func(function *ast.Node, class *ast.Node) {
			className := classNameTextOf(class)
			if className == "" {
				return
			}

			// `Node.Type()` answers a function-like node's return type through its
			// FunctionLikeData, which covers a method, an accessor, a function expression, and an
			// arrow with one accessor rather than four.
			returnType := function.Type()
			if returnType == nil {
				return
			}

			matched := typeReferenceNamed(className, returnType)
			if matched == nil {
				return
			}

			if !functionReturnsOnlyThis(ctx, function, class) {
				return
			}

			ctx.ReportNodeWithFixes(matched, buildUseThisTypeMessage(),
				rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, matched), "this"))
		}

		// checkProperty is upstream's `checkProperty`: a class property whose initializer is a
		// function expression or an arrow. A property holding anything else has no return type to
		// read, which is why `f = ''` and `f?: string` are among upstream's passing cases.
		checkProperty := func(node *ast.Node, class *ast.Node) {
			initializer := node.Initializer()
			if initializer == nil {
				return
			}
			// CRASH PROTECTION as well as a filter, and the crash is the part a fixture cannot
			// see. `checkFunction` calls `Parameters()`, which dereferences the function-like data
			// a non-function node does not have. A type assertion is the reachable shape:
			// `f = <Foo>this;` answers `Type()` with `Foo`, so it clears the return-type gate, and
			// then takes the process down. Removing these two lines panics on that input while
			// every findings fixture stays green.
			if initializer.Kind != ast.KindFunctionExpression &&
				initializer.Kind != ast.KindArrowFunction {
				return
			}
			checkFunction(initializer, class)
		}

		// Upstream's three selectors are all rooted at `ClassBody >`, so a method or property has to
		// be a direct member of a class. Anchoring on the class rather than on the member reproduces
		// that containment without asking each member who its parent is, and it means an object
		// literal method, which is a different node kind anyway, can never arrive here.
		visitClass := func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}

			for _, member := range node.Members() {
				switch member.Kind {
				// A get accessor and a set accessor are `MethodDefinition` in estree, which is what
				// upstream's `ClassBody > MethodDefinition` selector matches, so a getter annotated
				// with the class name is in scope. Measured: `get f(): Foo { return this; }` reports
				// on the installed build.
				case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
					ast.KindConstructor:
					checkFunction(member, node)
				case ast.KindPropertyDeclaration:
					checkProperty(member, node)
				}
			}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: visitClass,
			ast.KindClassExpression:  visitClass,
		}
	},
}

// functionReturnsOnlyThis is upstream's `isFunctionReturningThis`.
//
// Three gates in upstream's order: an explicit `this` parameter disqualifies outright, a function
// with no body cannot be judged, and then the body is read either as a single expression or as a
// set of return statements.
func functionReturnsOnlyThis(ctx rule.Context, function *ast.Node, class *ast.Node) bool {
	// `f(this: Foo): Foo` pins the receiver type explicitly, so the author has already said what
	// they mean and the rule declines. Upstream tests only the FIRST parameter, because that is the
	// only position the grammar allows a `this` parameter in.
	parameters := function.Parameters()
	if len(parameters) > 0 {
		firstName := parameters[0].Name()
		if firstName != nil && ast.IsIdentifier(firstName) && firstName.Text() == "this" {
			return false
		}
	}

	body := function.Body()
	if body == nil {
		return false
	}

	// The class's own type, and the `this` type that belongs to it. `GetTypeAtLocation` on a class
	// declaration answers the instance type, which is the InterfaceType carrying `thisType`.
	classType := ctx.TypeChecker.GetTypeAtLocation(class)
	if classType == nil {
		return false
	}
	classThisType := thisTypeOf(classType)

	// An arrow's concise body is a single expression rather than a block, and upstream compares its
	// type against `thisType` directly with no return-statement walk.
	//
	// The parentheses unwrap belongs here rather than at the report site: `(): Foo => (this)` gives
	// us a parenthesized expression where upstream's parser gives the `this` keyword.
	if body.Kind != ast.KindBlock {
		bodyType := ctx.TypeChecker.GetTypeAtLocation(skipParenthesesSafely(body))
		return bodyType != nil && classThisType != nil && classThisType == bodyType
	}

	hasReturnThis := false
	hasReturnClassType := false

	ast.ForEachReturnStatement(body, func(statement *ast.Node) bool {
		expression := statement.AsReturnStatement().Expression
		if expression == nil {
			// A bare `return;` returns undefined, which is neither `this` nor the class type.
			// Upstream skips it rather than counting it against the method, which is what makes
			// `f(): Foo | undefined { if (1) return; return this; }` report.
			return false
		}

		// Deliberately NOT unwrapped: see the parentheses section in the doc comment. Upstream
		// tests this keyword on the TypeScript node, which keeps parentheses, so `(this)` has to
		// fall through to the comparison below.
		//
		// Upstream calls this a "fast check" and it is NOT one. A literal `this` keyword is
		// accepted without asking the checker, and that is a separate judgment rather than an
		// optimization, because the checker does not answer the same on four measured shapes:
		//
		//	f = function (): Foo { return this; }   `this` in a plain function expression is `any`
		//	static f(): Foo { return this; }        `this` in a static body is the constructor
		//	const Foo = class Bar { f(): Bar ... }  a class expression's declared type differs
		//
		// In each, `GetTypeAtLocation(this)` is not reference-identical to the class's `thisType`,
		// so removing this arm takes all of them silent while upstream reports every one. A mutant
		// neutralizing the arm fails four fixtures, which is how this was established after the
		// doc comment had already claimed the opposite.
		if expression.Kind == ast.KindThisKeyword {
			hasReturnThis = true
			return false
		}

		expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
		if expressionType == nil {
			return false
		}

		// Returning the class type itself disqualifies the method, and upstream stops the walk here
		// by returning true. `new Foo()` is the shape.
		if classType == expressionType {
			hasReturnClassType = true
			return true
		}

		if classThisType != nil && classThisType == expressionType {
			hasReturnThis = true
			return false
		}

		// A union containing the class type disqualifies too. This is what separates upstream's last
		// passing case from its last failing one: `BaseUnion | string` where the returned value is
		// `BaseUnion | string` is silent, and where it is `number | string` it reports.
		if unionContainsType(expressionType, classType) {
			hasReturnClassType = true
			return true
		}

		return false
	})

	return !hasReturnClassType && hasReturnThis
}

// typeReferenceNamed is upstream's `tryGetNameInType`: the type node itself when it is a reference
// to the given name, or the first such reference inside a union, recursively.
//
// Deliberately NOT extended to intersections. Upstream recurses through unions only, so
// `f(): Foo & {}` is silent on the installed build, and that gap is reproduced rather than closed.
func typeReferenceNamed(name string, typeNode *ast.Node) *ast.Node {
	if typeNode == nil {
		return nil
	}

	if typeNode.Kind == ast.KindTypeReference {
		reference := typeNode.AsTypeReferenceNode()
		typeName := reference.TypeName
		// A qualified name such as `N.Foo` is not an identifier, so upstream's `typeName.type ===
		// Identifier` test declines it. Measured silent on the installed build.
		if typeName != nil && ast.IsIdentifier(typeName) && typeName.Text() == name {
			return typeNode
		}
		return nil
	}

	if typeNode.Kind == ast.KindUnionType {
		members := typeNode.AsUnionTypeNode().Types
		if members == nil {
			return nil
		}
		for _, member := range members.Nodes {
			// A union member can itself be parenthesized: `(Foo | undefined) | null`. Our parser
			// keeps that node and upstream's does not, so unwrap before recursing. Measured
			// reporting on the installed build.
			if found := typeReferenceNamed(name, skipParenthesesInTypeSafely(member)); found != nil {
				return found
			}
		}
	}

	return nil
}

// skipParenthesesSafely unwraps parenthesized expressions without dereferencing a nil.
//
// `ast.SkipParentheses` dereferences its argument, and every caller here holds something optional:
// a concise body, a return argument. Written as a loop because `((this))` nests.
func skipParenthesesSafely(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// skipParenthesesInTypeSafely is the type-position counterpart: `(Foo | undefined) | null`.
func skipParenthesesInTypeSafely(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
	}
	return node
}

// thisTypeOf reads the `thisType` off a class's instance type.
//
// Only an interface type carries one, and a class's instance type is one, so a type that fails the
// conversion simply has no `this` type and every identity test against it is false.
func thisTypeOf(classType *checker.Type) *checker.Type {
	interfaceType := classType.AsInterfaceType()
	if interfaceType == nil {
		return nil
	}
	return checker.InterfaceType_thisType(interfaceType)
}

// unionContainsType answers upstream's `isUnionType(type) && type.types.some(part => part === class)`.
//
// Reference identity per constituent, matching upstream exactly. A non-union answers false rather
// than comparing itself, because upstream guards on `isUnionType` first and the caller has already
// tested plain identity.
//
// The `IsUnion` test is CRASH PROTECTION rather than a behavioral filter, and it is not
// interchangeable with a nil check. `AsUnionType` is an unchecked type assertion on the type's data
// pointer: handed an intrinsic it panics with `checker.TypeData is *checker.IntrinsicType, not
// *checker.UnionType` rather than returning nil. That is reachable from real source and this port
// shipped it briefly: `class Foo { f = function (): Foo { return this; } }` gives `this` inside a
// plain function expression the type `any`, which is an intrinsic, and it arrives here whenever the
// `this` keyword fast path above does not consume the return first. The cohere walk recovers per
// FILE rather than per rule, so one panic here costs every rule its verdict on that file.
//
// Found by a mutation that neutralized the fast path, which is worth recording: the fast path was
// masking a crash on the exact shape upstream's own corpus writes, so no fixture could see it while
// the optimization stood.
func unionContainsType(candidate *checker.Type, wanted *checker.Type) bool {
	if candidate == nil || wanted == nil {
		return false
	}
	if !candidate.IsUnion() {
		return false
	}
	for _, part := range candidate.AsUnionType().Types() {
		if part == wanted {
			return true
		}
	}
	return false
}

// classNameTextOf reads the class's own name, which upstream reads as `originalClass.id?.name`.
//
// An anonymous class expression has none, and upstream returns early on it, so `const Foo = class {
// f(): Foo { return this; } }` is silent while `const Foo = class Bar { f(): Bar ... }` reports.
// Both measured on the installed build.
// The identifier test upstream writes has no counterpart here, and that is measured rather than
// assumed. A class's name field is typed `*IdentifierNode`, and the parser answers NIL rather than
// some other node for every malformed spelling probed: `class 123 {}`, `class 'str' {}`,
// `class [x] {}`, `class {}`, `const A = class 5 {};` and `class default {}` all yield a nil name
// under error recovery. So an `ast.IsIdentifier` guard beside the nil test would be dead, and a
// mutant deleting it survives every fixture, correctly.
//
// The callers enumerated when that verdict was taken are the two listeners in this file, on
// KindClassDeclaration and KindClassExpression. A future caller passing something that is not a
// class-like node voids the measurement, because `Name()` on other kinds can answer a computed
// property name or a string literal.
func classNameTextOf(class *ast.Node) string {
	name := class.Name()
	if name == nil {
		return ""
	}
	return name.Text()
}

func buildUseThisTypeMessage() rule.Message {
	return rule.Message{
		Id:          "useThisType",
		Description: "Use `this` type instead.",
	}
}
