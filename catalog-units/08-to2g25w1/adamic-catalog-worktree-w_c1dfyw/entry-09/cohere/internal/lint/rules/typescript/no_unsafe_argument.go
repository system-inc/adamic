package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeArgument flags passing a value typed `any` where the parameter is typed something else.
//
//	valid:   declare function foo(arg: any): void; foo(1 as any);
//	valid:   declare function foo(arg: unknown): void; foo(1 as any);
//	valid:   declare function foo(...arg: any[]): void; foo(1, 2, 3, 4 as any);
//	valid:   declare function acceptsMap(arg: Map<string, string>): void; acceptsMap(new Map());
//	invalid: declare function foo(arg: number): void; foo(1 as any);
//	invalid: declare function foo(arg: number): void; declare const x: any[]; foo(...x);
//	invalid: declare function foo(a: number, b: string): void; declare const x: [any, any]; foo(...x);
//
// An `any` crossing into a typed parameter is where a type system stops being checked without
// saying so. The declaration keeps promising `number` and the call site quietly hands it anything,
// so every use inside the function is trusting a guarantee nobody made.
//
// # Four messages, and the pairing is the judgment
//
// An ordinary argument reports `unsafeArgument`. A spread reports one of three depending on what is
// being spread: `unsafeSpread` for a bare `any`, `unsafeArraySpread` for an `any[]`, and
// `unsafeTupleSpread` once per unsafe member when the spread is a tuple, since a tuple maps onto
// several parameters and only some of them may be unsafe. Anything else iterable is deliberately
// not handled, which upstream marks with a TODO and this reproduces rather than improves on.
//
// # The parameter walk is the port, not the `any` test
//
// Deciding whether an argument is unsafe is one call to the shared `IsUnsafeAssignment`. Deciding
// WHICH parameter an argument lands on is the work, and upstream keeps it in a `FunctionSignature`
// helper that this file reproduces as `unsafeArgumentSignature` below. Arguments consume parameters
// in order, a rest parameter absorbs everything after it, and a rest parameter that is itself a
// tuple hands out its members one at a time before repeating its last. A spread of a tuple whose
// target carries a variable element also marks every remaining argument consumed.
//
// # Two whole-call declines, both upstream's and both load-bearing
//
// A call with no arguments returns immediately, and a call whose CALLEE is `any` returns too. The
// second matters: `no-unsafe-call` already reports that call, and without this test every argument
// of it would be reported here as well. Six of upstream's passing cases are calls the checker
// cannot resolve, which fall out of the same test because an unresolved callee types as `any` or
// the resolved signature is nil.
//
// # A tagged template consumes its first parameter before any argument is seen
//
// `foo` + a template literal passes the strings array as the first argument implicitly, so the
// signature has to be advanced once before the interpolations are matched. Without that, every
// interpolation is compared against the wrong parameter and the messages name types nobody wrote.
//
// # Cost
//
// Two anchors, both common. The first thing each does is test the argument count, so a call with no
// arguments costs a slice length, and the checker is only asked once the call has arguments to
// judge.
var NoUnsafeArgument = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-argument",

	// Every judgment is a comparison between an argument's type and a parameter's type.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Expression == nil || call.Arguments == nil {
					return
				}
				checkUnsafeArguments(ctx, call.Arguments.Nodes, call.Expression, node, false)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()
				if newExpression.Expression == nil || newExpression.Arguments == nil {
					return
				}
				checkUnsafeArguments(ctx, newExpression.Arguments.Nodes, newExpression.Expression,
					node, false)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tagged := node.AsTaggedTemplateExpression()
				if tagged.Tag == nil || tagged.Template == nil {
					return
				}
				// Upstream passes `node.quasi.expressions`, which is the interpolated expressions
				// and not the string pieces. A template with no interpolation is a
				// NoSubstitutionTemplateLiteral here and has no spans at all, which is the same
				// empty list upstream gets and the same early return.
				if tagged.Template.Kind != ast.KindTemplateExpression {
					return
				}
				spans := tagged.Template.AsTemplateExpression().TemplateSpans
				if spans == nil {
					return
				}
				expressions := make([]*ast.Node, 0, len(spans.Nodes))
				for _, span := range spans.Nodes {
					expression := span.AsTemplateSpan().Expression
					if expression != nil {
						expressions = append(expressions, expression)
					}
				}
				checkUnsafeArguments(ctx, expressions, tagged.Tag, node, true)
			},
		}
	},
}

// checkUnsafeArguments is upstream's single body, shared by all three anchors.
func checkUnsafeArguments(
	ctx rule.Context,
	arguments []*ast.Node,
	callee *ast.Node,
	node *ast.Node,
	isTaggedTemplate bool,
) {
	if ctx.TypeChecker == nil || len(arguments) == 0 {
		return
	}

	// An `any` callee is `no-unsafe-call`'s finding, not this rule's, and upstream declines here so
	// the same code does not collect a finding from both rules.
	//
	// Against this checker the test is currently SUBSUMED, and that was measured rather than
	// assumed after a mutant neutralizing it survived every fixture. The first hypothesis, that the
	// fixtures simply had no `any`-callee case with arguments, was wrong: four such shapes were
	// tried and the two versions agreed on all four. The mechanical reason is that an `any` callee
	// resolves to a signature with ZERO parameters and no rest, probed across a bare identifier, a
	// parameter typed `any`, a property access on an `any`, and an `any`-typed property. So
	// `nextParameterType` answers nil for every argument and the loop skips them all.
	//
	// Kept because it is upstream's, because the redundancy rests on what this checker happens to
	// return for an unresolvable call rather than on anything guaranteed, and because it states the
	// intent that the two rules do not double-report. The measurement is here instead of a
	// deletion.
	calleeType := ctx.TypeChecker.GetTypeAtLocation(callee)
	if calleeType == nil || type_checking.IsTypeAnyType(calleeType) {
		return
	}

	signature := newUnsafeArgumentSignature(ctx, node)
	if signature == nil {
		// Upstream asserts the signature is non-nil, on the argument that a node from the AST map
		// always resolves. That holds for a call the checker understands and does not hold for one
		// it cannot resolve at all, which our parser hands us anyway because it recovers from
		// broken source. Declining is the same verdict upstream reaches by a different route: six
		// of its passing cases are unresolvable calls.
		return
	}

	if isTaggedTemplate {
		// The strings array is passed implicitly and occupies the first parameter.
		signature.nextParameterType()
	}

	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			checkUnsafeSpreadArgument(ctx, argument, signature)
			continue
		}

		parameterType := signature.nextParameterType()
		if parameterType == nil {
			continue
		}
		argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
		if argumentType == nil {
			continue
		}
		_, _, unsafe := type_checking.IsUnsafeAssignment(argumentType, parameterType,
			ctx.TypeChecker, argument)
		if !unsafe {
			continue
		}
		ctx.ReportNode(argument, rule.Message{
			Id: "unsafeArgument",
			Description: "Unsafe argument of type " + describeUnsafeArgumentType(ctx, argumentType) +
				" assigned to a parameter of type " + describeUnsafeArgumentType(ctx, parameterType) + ".",
		})
	}
}

// checkUnsafeSpreadArgument handles the three spread shapes upstream distinguishes.
//
// The fourth shape, something merely iterable, is deliberately not handled. Upstream carries a TODO
// there and this reproduces the gap rather than closing it, because closing it would report inputs
// upstream is silent on.
func checkUnsafeSpreadArgument(
	ctx rule.Context,
	spread *ast.Node,
	signature *unsafeArgumentSignature,
) {
	inner := spread.AsSpreadElement().Expression
	if inner == nil {
		return
	}
	spreadType := ctx.TypeChecker.GetTypeAtLocation(inner)
	if spreadType == nil {
		return
	}

	// The finding points at the whole spread including its dots, which is upstream's `node:
	// argument` where `argument` is the SpreadElement rather than its inner expression.
	switch {
	case type_checking.IsTypeAnyType(spreadType):
		ctx.ReportNode(spread, rule.Message{
			Id:          "unsafeSpread",
			Description: "Unsafe spread of an " + describeUnsafeArgumentType(ctx, spreadType) + " type.",
		})

	case type_checking.IsTypeAnyArrayType(spreadType, ctx.TypeChecker):
		ctx.ReportNode(spread, rule.Message{
			Id: "unsafeArraySpread",
			Description: "Unsafe spread of an " +
				describeUnsafeSpreadType(ctx, spreadType) + " array type.",
		})

	case checker.IsTupleType(spreadType):
		// A tuple spreads across several parameters, so each member is judged separately and the
		// finding is repeated on the same span once per unsafe member.
		for _, memberType := range checker.Checker_getTypeArguments(ctx.TypeChecker, spreadType) {
			parameterType := signature.nextParameterType()
			if parameterType == nil {
				continue
			}
			// The sender node is nil here and upstream says why at the line: the members of a
			// spread variable are not expressions we can point at, so the `new Map()` special case
			// inside IsUnsafeAssignment cannot apply and must not be given a node that would make
			// it fire on the wrong thing.
			_, _, unsafe := type_checking.IsUnsafeAssignment(memberType, parameterType,
				ctx.TypeChecker, nil)
			if !unsafe {
				continue
			}
			ctx.ReportNode(spread, rule.Message{
				Id: "unsafeTupleSpread",
				Description: "Unsafe spread of a tuple type. The argument is " +
					describeUnsafeTupleType(ctx, memberType) +
					" and is assigned to a parameter of type " +
					describeUnsafeArgumentType(ctx, parameterType) + ".",
			})
		}

		// A tuple whose target carries a variable element ends in a rest, so everything after it
		// is absorbed and no later argument gets its own parameter.
		if target := spreadType.TargetTupleType(); target != nil &&
			checker.TupleType_combinedFlags(target)&checker.ElementFlagsVariable != 0 {
			signature.consumeRemainingArguments()
		}
	}
}

// describeUnsafeArgumentType renders a type for the message, matching upstream's `describeType`.
//
// The error type gets prose rather than a name, because the checker's own rendering of it is the
// word `any` and a message saying "of type `any`" about a type the checker could not resolve sends
// the reader hunting an `any` that is not written anywhere. The backticks around a real type name
// are upstream's message text rather than markup added here.
func describeUnsafeArgumentType(ctx rule.Context, t *checker.Type) string {
	if type_checking.IsIntrinsicErrorType(t) {
		return "error typed"
	}
	return "`" + ctx.TypeChecker.TypeToString(t) + "`"
}

// describeUnsafeSpreadType is upstream's `describeTypeForSpread`, which differs from the plain one
// in exactly one place: an array OF the error type renders as the bare word `error`, so the message
// reads "Unsafe spread of an error array type" rather than naming the array.
func describeUnsafeSpreadType(ctx rule.Context, t *checker.Type) string {
	if checker.Checker_isArrayType(ctx.TypeChecker, t) {
		typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
		if len(typeArguments) > 0 && type_checking.IsIntrinsicErrorType(typeArguments[0]) {
			return "error"
		}
	}
	return describeUnsafeArgumentType(ctx, t)
}

// describeUnsafeTupleType is upstream's `describeTypeForTuple`. The difference from the plain one is
// the words rather than the type: this slot sits after "The argument is", so a resolved type reads
// "of type `X`" while the error type reads "error typed" with no preposition.
func describeUnsafeTupleType(ctx rule.Context, t *checker.Type) string {
	if type_checking.IsIntrinsicErrorType(t) {
		return "error typed"
	}
	return "of type `" + ctx.TypeChecker.TypeToString(t) + "`"
}

// unsafeArgumentSignature tracks which parameter the next argument lands on.
//
// This is upstream's `FunctionSignature` class, which lives in its shared util directory and is used
// by exactly one rule. It is a private type here rather than a shelf helper for that reason; if a
// second rule wants it, that is the moment to promote it.
//
// The three rest shapes are upstream's and they are not interchangeable. A rest parameter that is a
// TUPLE hands out its members positionally and then repeats its last, which is what makes
// `(...params: [number, string, any])` line up. A rest parameter that is an ARRAY hands out its
// element type forever. Anything else hands out the parameter type itself.
type unsafeArgumentSignature struct {
	parameterTypes []*checker.Type

	// restKind is one of the three below, or restKindNone when the signature has no rest parameter.
	restKind  unsafeArgumentRestKind
	restIndex int

	// restType carries the answer for the array and other kinds; restTypeArguments carries the
	// tuple's members. Only one is populated.
	restType          *checker.Type
	restTypeArguments []*checker.Type

	parameterTypeIndex   int
	hasConsumedArguments bool
}

type unsafeArgumentRestKind int

const (
	restKindNone unsafeArgumentRestKind = iota
	restKindArray
	restKindOther
	restKindTuple
)

// newUnsafeArgumentSignature resolves the call and reads its parameter list.
//
// Returns nil when the call does not resolve, which upstream treats as impossible and asserts on.
// See the caller for why declining is the right answer here rather than a panic.
func newUnsafeArgumentSignature(ctx rule.Context, node *ast.Node) *unsafeArgumentSignature {
	resolved := ctx.TypeChecker.GetResolvedSignature(node)
	if resolved == nil {
		return nil
	}

	signature := &unsafeArgumentSignature{restKind: restKindNone}
	for index, parameter := range checker.Signature_parameters(resolved) {
		parameterType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameter, node)
		if parameterType == nil {
			continue
		}

		isRest := false
		if len(parameter.Declarations) > 0 {
			// Upstream reads declaration [0] here, and this reproduces that choice rather than
			// looping. The question is "is THIS parameter a rest parameter", which is a property of
			// the one declaration a parameter symbol has; a parameter symbol does not merge.
			isRest = type_checking.IsRestParameterDeclaration(parameter.Declarations[0])
		}
		if !isRest {
			signature.parameterTypes = append(signature.parameterTypes, parameterType)
			continue
		}

		constrained := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, parameterType)
		if constrained == nil {
			constrained = parameterType
		}
		signature.restIndex = index
		switch {
		case checker.IsTupleType(constrained):
			signature.restKind = restKindTuple
			signature.restTypeArguments = checker.Checker_getTypeArguments(ctx.TypeChecker, constrained)
		default:
			// The shelf already answers exactly this question, wrapping the two-argument
			// index lookup with the checker's number type. Upstream spells it
			// `getIndexTypeOfType(t, ts.IndexKind.Number)`.
			elementType := type_checking.GetNumberIndexType(ctx.TypeChecker, constrained)
			if elementType != nil {
				signature.restKind = restKindArray
				signature.restType = elementType
			} else {
				signature.restKind = restKindOther
				signature.restType = constrained
			}
		}
		// A rest parameter is last by construction, so nothing after it is read.
		break
	}
	return signature
}

// consumeRemainingArguments marks every later argument as absorbed by the rest parameter.
func (s *unsafeArgumentSignature) consumeRemainingArguments() {
	s.hasConsumedArguments = true
}

// nextParameterType hands out the parameter the next argument lands on, or nil when the call has
// more arguments than the signature has places for them.
//
// Nil is not an error and upstream skips rather than reports on it: too many arguments is already a
// compiler error, and reporting a type mismatch against a parameter that does not exist would name
// a type nobody wrote.
func (s *unsafeArgumentSignature) nextParameterType() *checker.Type {
	index := s.parameterTypeIndex
	s.parameterTypeIndex++

	if index < len(s.parameterTypes) && !s.hasConsumedArguments {
		return s.parameterTypes[index]
	}

	switch s.restKind {
	case restKindTuple:
		if len(s.restTypeArguments) == 0 {
			return nil
		}
		last := s.restTypeArguments[len(s.restTypeArguments)-1]
		if s.hasConsumedArguments {
			return last
		}
		typeIndex := index - s.restIndex
		if typeIndex < 0 || typeIndex >= len(s.restTypeArguments) {
			return last
		}
		return s.restTypeArguments[typeIndex]

	case restKindArray, restKindOther:
		return s.restType
	}
	return nil
}
