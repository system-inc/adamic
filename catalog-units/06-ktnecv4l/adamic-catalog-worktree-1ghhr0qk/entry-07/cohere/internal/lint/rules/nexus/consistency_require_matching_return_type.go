package nexus

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const consistencyRequireMatchingReturnTypeBareId = "bareReturnInValueFunction"

const consistencyRequireMatchingReturnTypeUndefinedId = "undefinedReturnInVoidFunction"

// consistencyRequireMatchingReturnTypeOrigin is where a function's return type came from, which both
// messages say.
type consistencyRequireMatchingReturnTypeOrigin string

const (
	consistencyRequireMatchingReturnTypeFromAnnotation consistencyRequireMatchingReturnTypeOrigin = "Annotation"
	consistencyRequireMatchingReturnTypeFromSignature  consistencyRequireMatchingReturnTypeOrigin = "Signature"
	consistencyRequireMatchingReturnTypeInferred       consistencyRequireMatchingReturnTypeOrigin = "Inferred"
)

// The rule's two messages and the options they pick from, whose wording lives in
// `policy/messages/consistency-require-matching-return-type.json`.
var (
	consistencyRequireMatchingReturnTypeBareText          = policy.MessageOf("nexus/consistency-require-matching-return-type", consistencyRequireMatchingReturnTypeBareId)
	consistencyRequireMatchingReturnTypeUndefinedText     = policy.MessageOf("nexus/consistency-require-matching-return-type", consistencyRequireMatchingReturnTypeUndefinedId)
	consistencyRequireMatchingReturnTypeConstructorReason = consistencyRequireMatchingReturnTypeUndefinedText.Option("voidReason", "constructor")
	consistencyRequireMatchingReturnTypeSetterReason      = consistencyRequireMatchingReturnTypeUndefinedText.Option("voidReason", "setter")
)

// consistencyRequireMatchingReturnTypeSources says where a value verdict's type came from, in the
// bare-return message.
var consistencyRequireMatchingReturnTypeSources = map[consistencyRequireMatchingReturnTypeOrigin]policy.MessageOption{
	consistencyRequireMatchingReturnTypeFromAnnotation: consistencyRequireMatchingReturnTypeBareText.Option("source", "annotation"),
	consistencyRequireMatchingReturnTypeFromSignature:  consistencyRequireMatchingReturnTypeBareText.Option("source", "signature"),
	consistencyRequireMatchingReturnTypeInferred:       consistencyRequireMatchingReturnTypeBareText.Option("source", "inferred"),
}

// consistencyRequireMatchingReturnTypeVoidReasons says why a `void` type makes a verdict void, by where
// the type came from, in the undefined-return message. One option per source rather than a source
// inside the reason, since phrases do not nest.
var consistencyRequireMatchingReturnTypeVoidReasons = map[consistencyRequireMatchingReturnTypeOrigin]policy.MessageOption{
	consistencyRequireMatchingReturnTypeFromAnnotation: consistencyRequireMatchingReturnTypeUndefinedText.Option("voidReason", "annotatedVoid"),
	consistencyRequireMatchingReturnTypeFromSignature:  consistencyRequireMatchingReturnTypeUndefinedText.Option("voidReason", "signatureVoid"),
	consistencyRequireMatchingReturnTypeInferred:       consistencyRequireMatchingReturnTypeUndefinedText.Option("voidReason", "inferredVoid"),
}

// ConsistencyRequireMatchingReturnType makes every `return` spell what the function's return type
// says it returns.
//
//	invalid: function find(): string | undefined { if (!ready) return; return name; }
//	valid:   function find(): string | undefined { if (!ready) return undefined; return name; }
//	invalid: async function load(): Promise<string | undefined> { if (!ready) return; return name; }
//	valid:   function save(): void { if (!ready) return; write(); }
//	invalid: function save(): void { if (!ready) return undefined; write(); }
//	valid:   useEffect(function() { if (!ready) return; start(); return function() { stop(); }; });
//
// # The principle
//
// A bare `return;` is allowed only where the function's return type says `void`. Everywhere else
// every `return` carries a value, so `undefined` is written `return undefined;`. And the converse,
// so there is one shape everywhere: in a function whose type says `void`, the early exit is written
// `return;`, never `return undefined;`. `void` means "no value"; `undefined` is a value, and the
// spelling at the return tells the reader which of the two this function deals in without making
// them scroll to the signature.
//
// # Why this replaces `consistent-return`
//
// Upstream's `consistent-return` cannot see types. It fires on an exhaustive `switch` with no
// `default`, whose fall-off the checker proves unreachable, and it cannot tell a `void` callback
// (a React effect that exits early and otherwise returns its cleanup) from a function that returns
// `undefined` as a value. This rule asks the checker instead, and it judges only `return`
// statements: it never reports a function's fall-off end, which is `noImplicitReturns`' job.
//
// # Which return type, in order
//
//  1. The declared annotation, when the function has one (`function f(): string | undefined`).
//  2. The contextual signature, for a function expression, an arrow function, or an object-literal
//     method with no annotation: the return type of the one call signature its context expects
//     (the `useEffect` callback's `void | Destructor`, an `onClick`'s `void`). A union context
//     contributes every member's single signature, and a member with no call signature (the
//     `undefined` of an optional callback property) is skipped, as the checker's own
//     `getContextualSignature` skips it. A member with more than one call signature makes the
//     context ambiguous and it is not used.
//  3. The inferred type, from the function's own signature. A function with only bare returns
//     infers `void`; one that mixes `return;` with `return value;` infers `Value | undefined`,
//     which is exactly the inconsistency `consistent-return` existed to catch.
//
// A contextual type that is `any`, `unknown`, `never`, or an uninstantiated type parameter says
// nothing about the callback's returns (an event listener typed `=> any`), so it is passed over and
// the inferred type decides. A DECLARED `any` or `unknown` is the author's statement and is not
// passed over: the rule cannot know which spelling is meant and reports nothing.
//
// A contextual or inferred type of exactly `undefined` is unknowable. Both are usually the function's
// own returns read back (`return void log();` infers `undefined`, and a generic context like
// `forEachChild`'s `T | undefined` or `Array.from`'s mapper `U` is instantiated from the callback it
// types), so the type records the spelling rather than an intent, and neither reading is safe: see
// the arm in the classifier for the ahra cases each one got wrong. A declared `(): undefined` is the
// author's and stays a value.
//
// An overloaded function declaration with no annotation on its implementation is judged by its
// overload signatures, and only when they all agree on void or value; overloads that disagree are
// treated as unknowable rather than guessed at.
//
// # Async functions and generators
//
// An `async` function is judged by the awaited type: `Promise<void>` is void, `Promise<string |
// undefined>` is a value. A non-async function declared to return a promise is not unwrapped,
// because a bare `return;` there really does return `undefined` rather than a promise.
//
// A generator's `return` sets the iterator's final `value`, so the type that governs it is the
// generator's TReturn, the second type argument of `Generator`, `AsyncGenerator`, `Iterator`,
// `IterableIterator` and friends. An unannotated generator with only bare returns infers TReturn
// `void`, so it is silent; a declared `Generator<number, string | undefined>` with a bare return
// reports. A return type that is not a reference with a TReturn argument is unknowable.
//
// # The function kinds with no return type to read
//
// A constructor and a set accessor have no value to return: `return;` is their early exit and is
// always allowed, and `return undefined;` in them is reported exactly as in a `void` function. A get
// accessor is an ordinary function whose return type is the property's type.
//
// `never` is unknowable rather than a value: a function typed `never` cannot reach a `return` without
// a type error, and there is no spelling to suggest. An arrow function with an expression body has
// no `return` statement and is out of scope.
//
// # What counts as `return undefined;`
//
// The identifier `undefined`, optionally parenthesized, resolving to the global (a name that resolves
// to a declaration in source is somebody's local and is not the global). `return void 0;` is a
// different spelling and is left alone.
//
// # The fixer, and why it is safe
//
// Both directions change no behaviour: `return;` and `return undefined;` return the same value at
// runtime. The fix is offered only when the statement is exactly the canonical text, `return;` or
// `return undefined;` with nothing but whitespace between, so no comment is eaten and no automatic
// semicolon insertion changes meaning (inserting ` undefined` after a `return` with no semicolon
// could join the next line into the expression). In the value direction the fix is also withheld
// when the return type does not admit `undefined` (`(): number` with a bare return, which only
// compiles without `noImplicitReturns`): writing `return undefined;` there would surface a type
// error, which is the right outcome but not an unattended one.
var ConsistencyRequireMatchingReturnType = rule.Rule{
	Name: "nexus/consistency-require-matching-return-type",

	// The verdict is the function's return type, declared, contextual or inferred, and only the
	// checker can answer any of the three.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		// One verdict per function, however many returns it holds.
		verdicts := map[*ast.Node]consistencyRequireMatchingReturnTypeVerdict{}

		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				function := consistencyRequireMatchingReturnTypeContainer(node)
				if function == nil {
					return
				}
				verdict, seen := verdicts[function]
				if !seen {
					verdict = consistencyRequireMatchingReturnTypeJudge(ctx, function)
					verdicts[function] = verdict
				}

				expression := node.AsReturnStatement().Expression
				switch verdict.kind {
				case consistencyRequireMatchingReturnTypeKindValue:
					if expression != nil {
						return
					}
					message := rule.Message{
						Id: consistencyRequireMatchingReturnTypeBareId,
						Description: consistencyRequireMatchingReturnTypeBareText.Render(
							map[string]string{"functionKind": verdict.functionKind, "rendered": verdict.rendered},
							verdict.source,
						),
					}
					if verdict.admitsUndefined && consistencyRequireMatchingReturnTypeText(ctx, node) == "return;" {
						ctx.ReportNodeWithFixes(node, message, rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), "return undefined;"))
						return
					}
					ctx.ReportNode(node, message)

				case consistencyRequireMatchingReturnTypeKindVoid:
					if expression == nil || !consistencyRequireMatchingReturnTypeIsGlobalUndefined(ctx, expression) {
						return
					}
					message := rule.Message{
						Id:          consistencyRequireMatchingReturnTypeUndefinedId,
						Description: consistencyRequireMatchingReturnTypeUndefinedText.Render(verdict.voidValues(), verdict.voidReason),
					}
					if consistencyRequireMatchingReturnTypeCanonicalUndefined.MatchString(consistencyRequireMatchingReturnTypeText(ctx, node)) {
						ctx.ReportNodeWithFixes(node, message, rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), "return;"))
						return
					}
					ctx.ReportNode(node, message)
				}
			},
		}
	},
}

// consistencyRequireMatchingReturnTypeCanonicalUndefined is the only `return undefined` text the
// fixer rewrites: whitespace between the two words and a closing semicolon, nothing else.
var consistencyRequireMatchingReturnTypeCanonicalUndefined = regexp.MustCompile(`^return\s+undefined;$`)

type consistencyRequireMatchingReturnTypeKind int

const (
	consistencyRequireMatchingReturnTypeKindUnknown consistencyRequireMatchingReturnTypeKind = iota
	consistencyRequireMatchingReturnTypeKindVoid
	consistencyRequireMatchingReturnTypeKindValue
)

// consistencyRequireMatchingReturnTypeVerdict is what one function's returns are judged against.
type consistencyRequireMatchingReturnTypeVerdict struct {
	kind consistencyRequireMatchingReturnTypeKind

	// functionKind names the function in the message: "function", "async function", "method",
	// "callback", "getter", "generator", "constructor", "setter".
	functionKind string

	// rendered and source describe a value verdict's type and where it came from.
	rendered string
	source   policy.MessageOption

	// voidReason says why a void verdict is void, as a clause, and written is the type it names when
	// the reason is a `void` type rather than a constructor or a setter.
	voidReason policy.MessageOption
	written    string

	// admitsUndefined is whether `undefined` is a member of a value verdict's type, which is what
	// makes the bare-return fix type-safe as well as behaviour-safe.
	admitsUndefined bool
}

// voidValues are the values the undefined-return message takes: the function's kind, and the type its
// reason names when the reason is a `void` type.
func (verdict consistencyRequireMatchingReturnTypeVerdict) voidValues() map[string]string {
	values := map[string]string{"functionKind": verdict.functionKind}
	if verdict.written != "" {
		values["written"] = verdict.written
	}
	return values
}

// consistencyRequireMatchingReturnTypeContainer is the function a `return` belongs to, or nil.
//
// Written out rather than `ast.GetContainingFunction`, because that walks through a class static
// block to whatever function encloses the class. A `return` in a static block is a grammar error, and
// attributing it to an outer function would judge it against the wrong signature.
func consistencyRequireMatchingReturnTypeContainer(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return current
		case ast.KindClassStaticBlockDeclaration, ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// consistencyRequireMatchingReturnTypeJudge decides whether a function is void, a value, or unknowable.
func consistencyRequireMatchingReturnTypeJudge(ctx rule.Context, function *ast.Node) consistencyRequireMatchingReturnTypeVerdict {
	flags := ast.GetFunctionFlags(function)
	isAsync := flags&ast.FunctionFlagsAsync != 0
	isGenerator := flags&ast.FunctionFlagsGenerator != 0
	functionKind := consistencyRequireMatchingReturnTypeFunctionKind(function, isAsync, isGenerator)

	switch function.Kind {
	case ast.KindConstructor:
		return consistencyRequireMatchingReturnTypeVerdict{
			kind:         consistencyRequireMatchingReturnTypeKindVoid,
			functionKind: functionKind,
			voidReason:   consistencyRequireMatchingReturnTypeConstructorReason,
		}
	case ast.KindSetAccessor:
		return consistencyRequireMatchingReturnTypeVerdict{
			kind:         consistencyRequireMatchingReturnTypeKindVoid,
			functionKind: functionKind,
			voidReason:   consistencyRequireMatchingReturnTypeSetterReason,
		}
	}

	// 1. The declared annotation is the author's statement, and is final even when it is `any`.
	if annotation := function.Type(); annotation != nil {
		declared := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, annotation)
		return consistencyRequireMatchingReturnTypeClassify(ctx, declared, true, isAsync, isGenerator, functionKind, consistencyRequireMatchingReturnTypeFromAnnotation)
	}

	// 2. The contextual signature, passed over when it says nothing.
	if contextual := consistencyRequireMatchingReturnTypeContextualReturnType(ctx, function); contextual != nil {
		verdict := consistencyRequireMatchingReturnTypeClassify(ctx, contextual, false, isAsync, isGenerator, functionKind, consistencyRequireMatchingReturnTypeFromSignature)
		if verdict.kind != consistencyRequireMatchingReturnTypeKindUnknown {
			return verdict
		}
	}

	// 3. The inferred type, from the function's own signatures.
	functionType := ctx.TypeChecker.GetTypeAtLocation(function)
	if functionType == nil {
		return consistencyRequireMatchingReturnTypeVerdict{}
	}
	if function.Kind == ast.KindGetAccessor {
		// A get accessor's type is the property's type, which is its return type.
		return consistencyRequireMatchingReturnTypeClassify(ctx, functionType, false, isAsync, isGenerator, functionKind, consistencyRequireMatchingReturnTypeInferred)
	}
	signatures := type_checking.GetCallSignatures(ctx.TypeChecker, functionType)
	if len(signatures) == 0 {
		return consistencyRequireMatchingReturnTypeVerdict{}
	}
	var agreed consistencyRequireMatchingReturnTypeVerdict
	for index, signature := range signatures {
		returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
		verdict := consistencyRequireMatchingReturnTypeClassify(ctx, returnType, false, isAsync, isGenerator, functionKind, consistencyRequireMatchingReturnTypeInferred)
		if index == 0 {
			agreed = verdict
			continue
		}
		// Overloads that disagree on void against value are unknowable, not guessed.
		if verdict.kind != agreed.kind {
			return consistencyRequireMatchingReturnTypeVerdict{}
		}
		agreed.admitsUndefined = agreed.admitsUndefined && verdict.admitsUndefined
	}
	return agreed
}

// consistencyRequireMatchingReturnTypeContextualReturnType is the return type the function's context
// expects, or nil when it has none or more than one call signature could apply.
//
// The checker's own `getContextualSignature` is not reachable through the shim, so this rebuilds it
// from the contextual TYPE: the apparent type's single call signature, or, for a union, each member's
// single signature with members that have none skipped, their return types unioned.
func consistencyRequireMatchingReturnTypeContextualReturnType(ctx rule.Context, function *ast.Node) *checker.Type {
	var contextual *checker.Type
	switch {
	case function.Kind == ast.KindFunctionExpression, function.Kind == ast.KindArrowFunction:
		contextual = checker.Checker_getContextualType(ctx.TypeChecker, function, checker.ContextFlagsSignature)
	case ast.IsObjectLiteralMethod(function):
		// Not `getContextualType`, which dispatches on the node's PARENT and has no arm for a method
		// whose parent is the object literal itself. Measured: through that call the method's
		// contextual `() => void` went unread, and only a fixture whose inferred type disagrees with
		// its context could see it. The checker's own path for a method is this one.
		contextual = ctx.TypeChecker.GetContextualTypeForObjectLiteralElement(function, checker.ContextFlagsSignature)
	default:
		return nil
	}
	if contextual == nil {
		return nil
	}
	var returnTypes []*checker.Type
	for part := range type_checking.UnionTypePartsSeq(contextual) {
		if part == nil {
			continue
		}
		apparent := checker.Checker_getApparentType(ctx.TypeChecker, part)
		if apparent == nil {
			continue
		}
		signatures := type_checking.GetCallSignatures(ctx.TypeChecker, apparent)
		switch len(signatures) {
		case 0:
			continue
		case 1:
			if returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signatures[0]); returnType != nil {
				returnTypes = append(returnTypes, returnType)
			}
		default:
			return nil
		}
	}
	switch len(returnTypes) {
	case 0:
		return nil
	case 1:
		return returnTypes[0]
	}
	return ctx.TypeChecker.GetUnionType(returnTypes)
}

// consistencyRequireMatchingReturnTypeClassify reads one return type as void, value or unknowable,
// after unwrapping a generator's TReturn and an async function's awaited type.
func consistencyRequireMatchingReturnTypeClassify(
	ctx rule.Context,
	returnType *checker.Type,
	declared bool,
	isAsync bool,
	isGenerator bool,
	functionKind string,
	source consistencyRequireMatchingReturnTypeOrigin,
) consistencyRequireMatchingReturnTypeVerdict {
	unknown := consistencyRequireMatchingReturnTypeVerdict{functionKind: functionKind}
	if returnType == nil {
		return unknown
	}
	written := returnType

	if isGenerator {
		returnType = consistencyRequireMatchingReturnTypeGeneratorReturn(ctx, returnType)
		if returnType == nil {
			return unknown
		}
	} else if isAsync {
		returnType = checker.Checker_getAwaitedType(ctx.TypeChecker, returnType)
		if returnType == nil {
			return unknown
		}
	}

	flags := checker.Type_flags(returnType)
	if flags&(checker.TypeFlagsAnyOrUnknown|checker.TypeFlagsNever) != 0 {
		return unknown
	}

	// The union's constituents OR'd, not the type's own flags: `string | void` is a union whose own
	// flags carry no Void bit, and the shelf's `IsTypeFlagSet` would answer false for it.
	var combined checker.TypeFlags
	for part := range type_checking.UnionTypePartsSeq(returnType) {
		if part != nil {
			combined |= checker.Type_flags(part)
		}
	}
	if combined&checker.TypeFlagsVoid != 0 {
		return consistencyRequireMatchingReturnTypeVerdict{
			kind:         consistencyRequireMatchingReturnTypeKindVoid,
			functionKind: functionKind,
			voidReason:   consistencyRequireMatchingReturnTypeVoidReasons[source],
			written:      ctx.TypeChecker.TypeToString(written),
		}
	}
	// Nothing but `undefined`, and not because the author wrote it, is unknowable. An inferred type is
	// the function's own returns read back, so `return void log(); ... return;` infers `undefined`
	// purely from its spelling, and a generic context instantiated from the callback it types
	// (`forEachChild(node, (child) => { if (skip) return; visit(child); })`, whose `T | undefined`
	// resolves to `undefined`) is the same circle one call away. Such a type cannot say which kind of
	// function this is, and both readings are wrong somewhere on ahra:
	//
	//   - read as a VALUE, it asked for `return undefined;` in 27 exits of one CLI dispatcher whose
	//     `return void console.log(...)` made its inferred type `Promise<undefined>`, and in a
	//     `forEachChild` visitor: 28 of the first 36 findings, all noise;
	//   - read as VOID, it asked for `return;` in `Array.from({ length }, function() { return
	//     undefined; })` and in stub objects standing in for `() => X | undefined` methods, where
	//     `undefined` IS the value, and the fix would retype each as `void` and break its consumer.
	//
	// So it reports nothing. A DECLARED `(): undefined` is the author saying `undefined` is the value,
	// and stays a value.
	if !declared && flags == checker.TypeFlagsUndefined {
		return unknown
	}
	// A bare type parameter, conditional or indexed access could be instantiated as `void`.
	if flags&checker.TypeFlagsInstantiable != 0 {
		return unknown
	}
	return consistencyRequireMatchingReturnTypeVerdict{
		kind:            consistencyRequireMatchingReturnTypeKindValue,
		functionKind:    functionKind,
		rendered:        ctx.TypeChecker.TypeToString(written),
		source:          consistencyRequireMatchingReturnTypeSources[source],
		admitsUndefined: combined&checker.TypeFlagsUndefined != 0,
	}
}

// consistencyRequireMatchingReturnTypeGeneratorReturn is a generator return type's TReturn, the second
// type argument of `Generator<T, TReturn, TNext>` and every iterator interface shaped like it, or nil.
//
// The reference guard is a crash guard as well as a narrowing: `Checker_getTypeArguments` panics on a
// type that is not a reference.
func consistencyRequireMatchingReturnTypeGeneratorReturn(ctx rule.Context, returnType *checker.Type) *checker.Type {
	if checker.Type_flags(returnType)&checker.TypeFlagsObject == 0 ||
		checker.Type_objectFlags(returnType)&checker.ObjectFlagsReference == 0 {
		return nil
	}
	arguments := checker.Checker_getTypeArguments(ctx.TypeChecker, returnType)
	if len(arguments) < 2 {
		return nil
	}
	return arguments[1]
}

// consistencyRequireMatchingReturnTypeIsGlobalUndefined is whether an expression is the global
// `undefined`, parentheses skipped.
//
// The global has no declarations, so a symbol with one is a local that happens to be named
// `undefined`. Note the complement: asking "does it resolve to a global" through a declarations check
// answers false for the real `undefined`, which is the trap the porting standard records.
func consistencyRequireMatchingReturnTypeIsGlobalUndefined(ctx rule.Context, expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	if expression == nil || expression.Kind != ast.KindIdentifier || expression.Text() != "undefined" {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(expression)
	return symbol == nil || len(symbol.Declarations) == 0
}

// consistencyRequireMatchingReturnTypeText is a node's own text, leading trivia excluded.
func consistencyRequireMatchingReturnTypeText(ctx rule.Context, node *ast.Node) string {
	textRange := rule.TokenRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}

// consistencyRequireMatchingReturnTypeFunctionKind names the function for the message.
func consistencyRequireMatchingReturnTypeFunctionKind(function *ast.Node, isAsync bool, isGenerator bool) string {
	var noun string
	switch function.Kind {
	case ast.KindConstructor:
		return "constructor"
	case ast.KindSetAccessor:
		return "setter"
	case ast.KindGetAccessor:
		return "getter"
	case ast.KindMethodDeclaration:
		noun = "method"
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		noun = "function expression"
		if function.Parent != nil && (function.Parent.Kind == ast.KindCallExpression || function.Parent.Kind == ast.KindNewExpression) {
			noun = "callback"
		}
	default:
		noun = "function"
	}
	if isGenerator {
		noun = "generator " + noun
	}
	if isAsync {
		noun = "async " + noun
	}
	return noun
}
