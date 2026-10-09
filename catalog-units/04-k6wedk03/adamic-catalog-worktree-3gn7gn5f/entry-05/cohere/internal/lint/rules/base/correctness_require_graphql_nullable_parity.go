package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// graphQlNullableParityDecorators are the decorators whose `nullable` flag this rule holds to the
// TypeScript type beside it.
var graphQlNullableParityDecorators = map[string]struct{}{
	"GraphQlArgument":      {},
	"GraphQlField":         {},
	"GraphQlQuery":         {},
	"GraphQlMutation":      {},
	"GraphQlFieldResolver": {},
}

// graphQlNullableParityRelationDecorators are the ORM relation decorators that take a property out
// of this rule's scope.
//
// A property carrying one of these has two truths that legitimately disagree: the GraphQL schema
// declares the business contract while the TypeScript type reflects the load state, so a relation is
// undefined until hydrated and its schema field is not nullable. `relation-must-be-optional` owns
// that pair. Enforcing parity there would demand the schema lie about the contract.
var graphQlNullableParityRelationDecorators = map[string]struct{}{
	"OrmManyToOne": {},
	"OrmOneToMany": {},
	"OrmOneToOne":  {},
}

// graphQlNullableParityInputClassDecorators mark a class whose fields a client sends rather than
// receives, which is where a nullable field must admit null itself.
var graphQlNullableParityInputClassDecorators = map[string]struct{}{
	"GraphQlInputType": {},
}

// GraphQlNullableParity holds a GraphQL decorator's `nullable` flag to the type it decorates.
//
//	valid:   @GraphQlField({ nullable: true })  name?: string;
//	valid:   @GraphQlField({ nullable: false }) name: string;
//	invalid: @GraphQlField({ nullable: true })  name: string;
//	invalid: @GraphQlField({ nullable: false }) name?: string;
//
// The decorator generates the schema and the type generates the code, and nothing makes them agree.
// When they drift, a field the schema promises is non-null starts returning null to clients that
// were told it could not, or a field clients could have relied on is announced as optional forever.
// Both are silent: each side compiles and each side is internally consistent.
//
// # What counts as nullable is wider than null and undefined
//
// `decorators.IsNullableType` also admits `any`, `unknown` and `void`, and that width is the
// judgment rather than a convenience. A decorator saying `nullable: false` beside a value typed
// `any` is a claim nothing checked, so the rule treats it as unmet rather than satisfied. Shared
// with the other three parity rules, which ask the same question about a different flag name.
//
// # An absent flag means false, and a non-boolean flag means skip
//
// Three outcomes rather than two, and the difference decides real cases. A decorator with no
// `nullable` key is checked AS `nullable: false`, so `@GraphQlField() name?: string` reports: the
// schema defaults to non-null and the type is optional. A decorator whose `nullable` is not a
// boolean literal is skipped entirely, which is what makes `nullable: 'items'` and
// `nullable: 'itemsAndList'` out of scope: those declare list-element nullability, a different
// question this rule cannot answer from the element type alone.
//
// # Where the type comes from differs by decorator, and the method arm is easy to miss
//
// An argument decorator reads its parameter's type. A field decorator on a property reads the
// property's. A field decorator on a METHOD reads the method's call-signature return type, because
// the underlying field decorator is valid on a getter or a method as well as a property, and a port
// handling only the property shape goes silent on every computed field. The three operation
// decorators always read a return type.
//
// A GETTER is reached and then declines, in both implementations, because a getter has no call
// signature to read a return type from. That is measured rather than assumed and it is recorded at
// the arm.
//
// A method return is unwrapped through `Promise<T>` once before the check. Once rather than
// recursively: this rule asks whether the resolved value can be absent, and `Promise<Promise<T>>` is
// not a shape anybody writes. That is the original's `unwrapPromise` rather than the recursive
// `unwrapToElementType` its sibling rule uses, and the two are deliberately different.
//
// # A nullable input must admit null itself, not just undefined
//
//	invalid: @GraphQlInputType() class I { @GraphQlField({ nullable: true }) name?: string; }
//	valid:   @GraphQlInputType() class I { @GraphQlField({ nullable: true }) name?: string | null; }
//
// On an input, `nullable: true` means a client may send an explicit null, and the value arrives as
// null. A type of `T | undefined` passes the parity check above, since undefined counts as nullable,
// and then tells every reader the null cannot happen: a `!== undefined` guard lets the client's null
// into whatever is downstream. So on the two input positions, an argument's parameter and a field of
// a `@GraphQlInputType` class, the type must admit null specifically.
//
// Outputs stay out, ruled by system_cohere (#twm9k22): a resolver returning undefined serializes as
// null, so `?: T` on an object type's field says nothing false about the wire, and requiring `| null`
// there would be style rather than a bug.
//
// # The report points at the name, not the decorator
//
// A property or method reports on its key and a parameter reports on itself, which is what the
// original does. The finding is about the declaration's nullability, so the name is where a reader
// looks to fix it.
//
// # No repair
//
// The original ships none. Both ways to satisfy the rule change behaviour: editing the flag changes
// the schema clients see, editing the type changes the code. Which is right is not recoverable from
// the source.
var GraphQlNullableParity = rule.Rule{
	Name: "base/correctness-require-graphql-nullable-parity",

	// One side of every comparison is a TypeScript type.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// checkNullability is the shared comparison: report when the declared flag and the type
		// disagree, in whichever direction, and on an input, when a nullable type admits undefined
		// but not the null a client can send.
		checkNullability := func(reportNode *ast.Node, declaredNullable bool, actualType *checker.Type,
			isInput bool) {
			if actualType == nil {
				return
			}
			typeIsNullable := decorators.IsNullableType(actualType)
			if declaredNullable == typeIsNullable {
				if isInput && declaredNullable && !typeIncludesNull(actualType) {
					ctx.ReportNode(reportNode, buildDecoratorNullableButInputExcludesNullMessage(
						ctx.TypeChecker.TypeToString(actualType)))
				}
				return
			}

			typeText := ctx.TypeChecker.TypeToString(actualType)
			if declaredNullable {
				ctx.ReportNode(reportNode, buildDecoratorNullableButTypeNotMessage(typeText))
				return
			}
			ctx.ReportNode(reportNode, buildTypeNullableButDecoratorNotMessage(typeText))
		}

		// returnTypeOf reads a method's call-signature return type, with one Promise unwrap.
		returnTypeOf := func(methodLike *ast.Node) *checker.Type {
			methodType := ctx.TypeChecker.GetTypeAtLocation(methodLike)
			if methodType == nil {
				return nil
			}
			signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, methodType,
				checker.SignatureKindCall)
			if len(signatures) == 0 {
				return nil
			}
			returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signatures[0])
			return unwrapPromiseOnce(ctx, returnType)
		}

		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				decoratorName := decorators.CallName(node)
				if _, isTarget := graphQlNullableParityDecorators[decoratorName]; !isTarget {
					return
				}

				// An absent `nullable` is read as false, which is the schema default and is what
				// makes an optional property under a bare decorator report. A present but
				// non-boolean value is skipped: `nullable: 'items'` is about list elements.
				declaredNullable, skip, _ := decorators.BooleanOption(node, "nullable")
				if skip {
					return
				}

				owner := node.Parent
				if owner == nil {
					return
				}

				if decoratorName == "GraphQlArgument" {
					// Load-bearing, and more directly so than the same guard in this rule's
					// sibling: the argument arm has nothing below it to decline a wrong owner, so
					// removing this reports a bogus finding on a decorator placed on a property, a
					// method, a class or a getter, reading that node's own type as if it were a
					// parameter's. All four measured, and all four silent upstream, where
					// `resolveDecoratedParameterNode` answers null for a non-parameter owner.
					if owner.Kind != ast.KindParameter {
						return
					}
					// Read at the parameter NODE, which is where the original reads it and which
					// is also where the report points. Reading at the parameter's name instead is
					// EQUIVALENT, measured across a plain, an optional and a defaulted parameter:
					// the checker answers the same type at both locations for all three, so a
					// mutant swapping them survives every fixture and no input can separate them.
					checkNullability(owner, declaredNullable,
						ctx.TypeChecker.GetTypeAtLocation(owner), true)
					return
				}

				if decoratorName == "GraphQlField" {
					// A relation's schema contract and its load state legitimately disagree, and a
					// sibling rule owns that pair.
					if decorators.HasDecoratorInSet(owner, graphQlNullableParityRelationDecorators) {
						return
					}

					switch owner.Kind {
					case ast.KindPropertyDeclaration:
						// A property's class decides whether a client sends it. A field on a
						// method or getter is computed for output, so only this arm asks.
						checkNullability(owner.Name(), declaredNullable,
							ctx.TypeChecker.GetTypeAtLocation(owner.Name()),
							decorators.HasDecoratorInSet(owner.Parent,
								graphQlNullableParityInputClassDecorators))
					case ast.KindMethodDeclaration, ast.KindGetAccessor:
						// The field decorator is valid on a method or a getter as well as a
						// property, so a port handling only the property shape goes silent on
						// every computed field.
						//
						// The get accessor arm is reached and then declines, which MATCHES the
						// original and is worth stating because it looks like a defect. A getter
						// has no call signature: the type at the node is already the property
						// type, so `returnTypeOf` finds no signature and answers nil, and
						// `checkNullability` returns on nil. estree reaches the same verdict by
						// the same route, since `getCallSignatures()` on its getter is empty and
						// the original's `if(!signature) return` fires.
						//
						// Measured against the real rule on a seeded file holding a disagreeing
						// getter beside an equivalent method: the method reports, the getter does
						// not. The arm is kept naming the kind rather than dropped, so the next
						// reader sees the shape was considered rather than missed.
						checkNullability(owner.Name(), declaredNullable, returnTypeOf(owner), false)
					}
					return
				}

				// The three operation decorators, which always sit on a method.
				if owner.Kind != ast.KindMethodDeclaration {
					return
				}
				checkNullability(owner.Name(), declaredNullable, returnTypeOf(owner), false)
			},
		}
	},
}

// unwrapPromiseOnce strips a single `Promise<T>` wrapper.
//
// Once rather than recursively, matching the original's `unwrapPromise` and deliberately unlike the
// recursive unwrap its sibling rule uses. This rule asks whether the resolved value can be absent,
// and one level is what an async method produces; a nested promise is not a shape anyone writes and
// unwrapping further would start answering a different question.
func unwrapPromiseOnce(ctx rule.Context, subject *checker.Type) *checker.Type {
	if subject == nil {
		return nil
	}
	symbol := checker.Type_symbol(subject)
	if symbol == nil || symbol.Name != "Promise" {
		return subject
	}
	typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, subject)
	if len(typeArguments) == 0 {
		return subject
	}
	return typeArguments[0]
}

// typeIncludesNull is the source library's `typeIncludesNull`.
//
// Null itself, or `any` and `unknown`, which admit it. Narrower than `decorators.IsNullableType`,
// whose mask also counts undefined and void, because the question here is whether a client's
// explicit null fits the type. Recursive across a union, because a union admits null when any
// member does.
func typeIncludesNull(subjectType *checker.Type) bool {
	if type_checking.IsTypeFlagSet(subjectType,
		checker.TypeFlagsNull|checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return true
	}
	if subjectType.IsUnion() {
		for _, part := range subjectType.AsUnionType().Types() {
			if typeIncludesNull(part) {
				return true
			}
		}
	}
	return false
}

// The rule's messages, whose wording lives in `policy/messages/correctness-require-graphql-nullable-parity.json`.
var (
	graphQlNullableParityDecoratorNullableButInputExcludesNullText = policy.MessageOf("base/correctness-require-graphql-nullable-parity", "decoratorNullableButInputExcludesNull")
	graphQlNullableParityDecoratorNullableButTypeNotText           = policy.MessageOf("base/correctness-require-graphql-nullable-parity", "decoratorNullableButTypeNot")
	graphQlNullableParityTypeNullableButDecoratorNotText           = policy.MessageOf("base/correctness-require-graphql-nullable-parity", "typeNullableButDecoratorNot")
)

func buildDecoratorNullableButInputExcludesNullMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          graphQlNullableParityDecoratorNullableButInputExcludesNullText.Id,
		Description: graphQlNullableParityDecoratorNullableButInputExcludesNullText.Render(map[string]string{"typeText": typeText}),
	}
}

func buildDecoratorNullableButTypeNotMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          graphQlNullableParityDecoratorNullableButTypeNotText.Id,
		Description: graphQlNullableParityDecoratorNullableButTypeNotText.Render(map[string]string{"typeText": typeText}),
	}
}

func buildTypeNullableButDecoratorNotMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          graphQlNullableParityTypeNullableButDecoratorNotText.Id,
		Description: graphQlNullableParityTypeNullableButDecoratorNotText.Render(map[string]string{"typeText": typeText}),
	}
}
