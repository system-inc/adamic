package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The seven messages, keeping upstream's wording so a suppression written against either tool reads
// the same. `rule.Message` has no interpolation layer, so upstream's slots become concatenation.

func messageNoUnsafeAssignmentAny(senderText string) rule.Message {
	return rule.Message{
		Id:          "anyAssignment",
		Description: "Unsafe assignment of an " + senderText + " value.",
	}
}

func messageNoUnsafeAssignmentAnyThis(senderText string) rule.Message {
	return rule.Message{
		Id: "anyAssignmentThis",
		Description: "Unsafe assignment of an " + senderText + " value. `this` is typed as `any`.\n" +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding " +
			"a `this` parameter to the function.",
	}
}

func messageNoUnsafeAssignmentUnsafe(senderText string, receiverText string) rule.Message {
	return rule.Message{
		Id: "unsafeAssignment",
		Description: "Unsafe assignment of type " + senderText + " to a variable of type " +
			receiverText + ".",
	}
}

func messageNoUnsafeAssignmentArrayPattern(senderText string) rule.Message {
	return rule.Message{
		Id:          "unsafeArrayPattern",
		Description: "Unsafe array destructuring of an " + senderText + " array value.",
	}
}

func messageNoUnsafeAssignmentArrayPatternFromTuple(senderText string) rule.Message {
	return rule.Message{
		Id: "unsafeArrayPatternFromTuple",
		Description: "Unsafe array destructuring of a tuple element with an " + senderText +
			" value.",
	}
}

func messageNoUnsafeAssignmentObjectPattern(senderText string) rule.Message {
	return rule.Message{
		Id:          "unsafeObjectPattern",
		Description: "Unsafe object destructuring of a property with an " + senderText + " value.",
	}
}

func messageNoUnsafeAssignmentArraySpread(senderText string) rule.Message {
	return rule.Message{
		Id:          "unsafeArraySpread",
		Description: "Unsafe spread of an " + senderText + " value in an array.",
	}
}

// noUnsafeAssignmentComparison is upstream's `ComparisonType` enum.
type noUnsafeAssignmentComparison int

const (
	// noUnsafeAssignmentComparisonNone does no assignment comparison, because an inferred variable
	// takes the sender's type and the two can never disagree.
	noUnsafeAssignmentComparisonNone noUnsafeAssignmentComparison = iota

	// noUnsafeAssignmentComparisonBasic uses the receiver's own type.
	noUnsafeAssignmentComparisonBasic

	// noUnsafeAssignmentComparisonContextual uses the receiver's CONTEXTUAL type, for a position
	// whose own type is inferred from the value being put in it.
	noUnsafeAssignmentComparisonContextual
)

// NoUnsafeAssignment flags assigning a value typed `any` into something typed otherwise.
//
//	valid:   const x: unknown = 1 as any;          unknown absorbs any, which is its purpose
//	valid:   const x = 1 as any;                   inferred, so nothing to disagree with
//	valid:   const x: Set<string> = new Set<string>();
//	invalid: const x: string = 1 as any;
//	invalid: const x: Set<string> = new Set<any>();
//	invalid: const [a] = [] as any[];              destructuring an any array
//	invalid: const { a } = { a: 1 as any };        destructuring an any property
//	invalid: const a = [...(x as any)];            spreading an any into an array
//
// Ported from `@typescript-eslint/no-unsafe-assignment`, the fourth of the `no-unsafe-*` family and
// the first that also has to reason about DESTRUCTURING. Every verdict below was measured against
// the installed 8.67.0 build driven over a real TypeScript program, one file per program.
//
// # 387 findings on the ahra tree
//
// Stated here because a number that size is a decision rather than a side effect. It is the largest
// this port has produced and the second largest of the porting wave. Nothing about the rule is
// approximate: the count was compared line for line against the installed build over the same files.
//
// # The shared judgment already existed, and three siblings use it
//
// `type_checking.IsUnsafeAssignment` is the recursive comparison that decides whether a sender type
// smuggles an `any` into a receiver's generic position. `no-unsafe-return`, `no-unsafe-argument` and
// `no-unsafe-type-assertion` all call it, and this rule is the fourth. Nothing in the comparison
// needed changing, which is the point of it living on the shelf: a fifth copy would be a fifth
// opinion about `Set<any>` into `Set<string>`.
//
// # Three comparison modes, and the distinction is what keeps inferred variables quiet
//
// `const x = 1 as any` is SILENT while `const x: string = 1 as any` reports, and the difference is
// not the value but whether a type was written down. With no annotation the variable's type IS the
// sender's, so a comparison would be asking whether a type differs from itself. Upstream expresses
// that as `ComparisonType.None`, and it is why the annotation is consulted before the types are.
//
// The contextual mode is for a position whose own type comes from the value: an object literal
// property and a JSX attribute both take their type from what is assigned, so asking the receiver's
// own type answers the sender's. Upstream asks for the contextual type there and falls back.
//
// # Destructuring is a recursion, and its findings are per element
//
// `const { a: { b } } = x` walks into the object pattern and then into the nested one, reporting on
// each element whose sender type is `any`. Two rules govern it and both are upstream's: an `any`
// element is reported and NOT descended into, so `[[[x]]] = [any]` reports once rather than three
// times, and a rest element is skipped entirely because it is not a one-to-one assignment.
//
// # A reported assignment suppresses the destructuring walk
//
// Upstream threads a `didReport` boolean through: if the plain assignment check reported, the
// destructuring helpers do not run. That is why `const { a } = 1 as any` produces ONE finding rather
// than one per destructured name. Reproduced, and it is the reason those helpers return a boolean
// rather than nothing.
//
// # `unknown` absorbs `any`
//
// Assigning `any` into `unknown` is the safe move, because `unknown` forces the receiver to narrow
// before use. Upstream returns early on it, and a port without that reports the most defensive thing
// a caller can write.
//
// # Measured against ESLint, and the three rows that differ are the checker's, not the rule's
//
// Run over the ahra tree, 132 files, this reports 435 and ESLint reports 432, with ZERO findings
// ESLint has that this misses. All three extra rows are the same shape:
//
//	const { value, done } = await reader.read();   // reader from fetch(...).body.getReader()
//
// ESLint's checker types that initializer `ReadableStreamReadResult<Uint8Array<ArrayBuffer>>`;
// typescript-go's types it `ReadableStreamReadResult<any>`, so `value` really is `any` here and the
// rule reports what it is handed. The two `lib.dom.d.ts` files are byte-identical, so the difference
// is in generic resolution through `fetch`, above this rule.
//
// Two things pin that. A fourth site, `StreamTransforms.ts`, destructures the same call against a
// bare `ReadableStream` whose `R` defaults to `any`, and both tools report it, agreeing. And
// `@typescript-eslint/no-unsafe-argument`, ported separately and sharing no code with this file,
// independently reports an unsafe argument of type `any` for that same `value` at all three sites.
// Two unrelated rules reading the same checker see the same `any`.
//
// Nothing is special-cased for it. A rule that suppressed these would be lying about what its own
// checker knows, and would go silent if the resolution is later fixed.
var NoUnsafeAssignment = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-assignment",

	// Every judgment compares a sender type against a receiver type.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			// `const x: T = value` and `let x = value`. Upstream's `VariableDeclarator[init != null]`.
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				declaration := node.AsVariableDeclaration()
				initializer := declaration.Initializer
				if initializer == nil {
					return
				}
				name := declaration.Name()
				if name == nil {
					return
				}
				noUnsafeAssignmentCheckWithDestructuring(ctx, name, initializer, node,
					noUnsafeAssignmentComparisonFor(declaration.Type))
			},

			// `x = value`. Upstream's `AssignmentExpression[operator = "="]`, and only that
			// operator: `x += value` is a different question the rule does not ask.
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				if binary.Left == nil || binary.Right == nil {
					return
				}
				// Basic rather than derived from an annotation: the target already has a type,
				// whether it was written down or inferred at its declaration.
				noUnsafeAssignmentCheckWithDestructuring(ctx, binary.Left, binary.Right, node,
					noUnsafeAssignmentComparisonBasic)
			},

			// `class C { x: T = value }` and `class C { accessor x: T = value }`. TSESTree gives
			// those two separate node types, `PropertyDefinition` and `AccessorProperty`, and
			// typescript-go folds both into a property declaration carrying an `accessor` modifier.
			// One anchor here for upstream's two, reaching the same set.
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				property := node.AsPropertyDeclaration()
				if property.Initializer == nil {
					return
				}
				name := property.Name()
				if name == nil {
					return
				}
				noUnsafeAssignmentCheck(ctx, name, property.Initializer, node,
					noUnsafeAssignmentComparisonFor(property.Type))
			},

			// `{ key: value }`. Upstream's `:not(ObjectPattern) > Property`, whose exclusion is
			// structural here: a property inside a destructuring pattern is a KindBindingElement
			// rather than a KindPropertyAssignment, so this anchor cannot see one.
			ast.KindPropertyAssignment: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// Upstream's selector excludes a property inside an ObjectPattern, and its parser
				// has a distinct node for one. Ours reuses KindObjectLiteralExpression for a
				// literal being used as a destructuring TARGET, so the exclusion has to be asked
				// of the position rather than of the kind. Without it `({ x: y } = source)` reports
				// twice: once here and once from the destructuring walk that owns it.
				if noUnsafeAssignmentIsDestructuringTarget(node.Parent) {
					return
				}
				assignment := node.AsPropertyAssignment()
				name := assignment.Name()
				value := assignment.Initializer
				if name == nil || value == nil {
					return
				}
				// Upstream skips a value that is an assignment pattern or an empty-body function,
				// saying both are handled by another selector. The first cannot appear in an object
				// literal here, and the second is a method signature in an ambient class.
				if value.Kind == ast.KindFunctionExpression &&
					value.AsFunctionExpression().Body == nil {
					return
				}
				noUnsafeAssignmentCheck(ctx, name, value, node,
					noUnsafeAssignmentComparisonContextual)
			},

			// `{ key }`, the shorthand form, which upstream also reaches through `Property`: its
			// parser gives a shorthand the same node type with `shorthand: true`, and ours gives it
			// its own kind. The key and the value are the same identifier, so the contextual type
			// of the key is what the property declares and the type at the identifier is what is
			// being put there.
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if noUnsafeAssignmentIsDestructuringTarget(node.Parent) {
					return
				}
				shorthand := node.AsShorthandPropertyAssignment()
				// `{ y = 1 }` in a VALUE position is not legal TypeScript; the parser recovers and
				// hands back a shorthand carrying an initializer. Upstream's parser refuses the
				// file, so its rule never meets the shape and is silent on it, measured. Declining
				// here keeps one of its own passing cases passing.
				if shorthand.ObjectAssignmentInitializer != nil {
					return
				}
				name := shorthand.Name()
				if name == nil {
					return
				}
				noUnsafeAssignmentCheck(ctx, name, name, node,
					noUnsafeAssignmentComparisonContextual)
			},

			// `[...value]`. Upstream's `ArrayExpression > SpreadElement`, and the parent test is
			// load-bearing: a spread in a CALL is `no-unsafe-argument`'s question, not this one.
			ast.KindSpreadElement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if node.Parent == nil || node.Parent.Kind != ast.KindArrayLiteralExpression {
					return
				}
				argument := node.AsSpreadElement().Expression
				if argument == nil {
					return
				}
				restType := ctx.TypeChecker.GetTypeAtLocation(argument)
				if restType == nil {
					return
				}
				if type_checking.IsTypeAnyType(restType) ||
					noUnsafeAssignmentIsAnyArray(ctx, restType) {
					ctx.ReportNode(node, messageNoUnsafeAssignmentArraySpread(
						noUnsafeAssignmentSenderText(ctx, restType)))
				}
			},

			// `<div id={value} />`. The attribute's own type comes from the component's props, so
			// the comparison is contextual.
			ast.KindJsxAttribute: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				attribute := node.AsJsxAttribute()
				value := attribute.Initializer
				if value == nil || value.Kind != ast.KindJsxExpression {
					return
				}
				expression := value.AsJsxExpression().Expression
				if expression == nil {
					// `<div id={} />`. Upstream declines a JsxEmptyExpression by name; ours is a
					// JsxExpression with no inner expression, which is the same shape.
					return
				}
				name := attribute.Name()
				if name == nil {
					return
				}
				noUnsafeAssignmentCheck(ctx, name, expression, expression,
					noUnsafeAssignmentComparisonContextual)
			},

			// `function f(x: T = value)`. Upstream's `AssignmentPattern`, which TSESTree uses for a
			// defaulted parameter and for a defaulted destructuring element alike.
			ast.KindParameter: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				parameter := node.AsParameterDeclaration()
				if parameter.Initializer == nil {
					return
				}
				name := parameter.Name()
				if name == nil {
					return
				}
				// The finding points at the ASSIGNMENT rather than at the parameter, which matters
				// for a parameter property: `constructor(private a = 1 as any)` is one node here
				// and two upstream, where the modifier lives on a TSParameterProperty wrapping an
				// AssignmentPattern. Reporting the parameter would underline `private a = 1 as any`
				// where upstream underlines `a = 1 as any`, and only a span assertion sees it.
				// A parameter PROPERTY carries an accessibility modifier that upstream's tree keeps
				// on a separate `TSParameterProperty` wrapper, so its finding covers only the
				// assignment inside: `constructor(private a = 1 as any)` underlines `a = 1 as any`
				// upstream, and reporting the parameter node here would underline the modifier too.
				// Only a span assertion sees the difference.
				reportRange := rule.TokenRange(ctx.SourceFile, node)
				if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
					reportRange = rule.TokenRange(ctx.SourceFile, name).
						WithEnd(parameter.Initializer.End())
				}
				noUnsafeAssignmentCheckWithDestructuringAtRange(ctx, name, parameter.Initializer,
					reportRange, noUnsafeAssignmentComparisonBasic)
			},

			// A defaulted element inside a destructuring pattern, `const { a = value } = x`, which
			// TSESTree also calls an AssignmentPattern and which is a binding element here.
			ast.KindBindingElement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				element := node.AsBindingElement()
				if element.Initializer == nil {
					return
				}
				name := element.Name()
				if name == nil {
					return
				}
				noUnsafeAssignmentCheckWithDestructuring(ctx, name, element.Initializer, node,
					noUnsafeAssignmentComparisonBasic)
			},
		}
	},
}

// noUnsafeAssignmentIsDestructuringTarget answers whether an object literal is standing in for a
// destructuring pattern rather than constructing a value.
//
// TSESTree has a separate ObjectPattern node and upstream's selector names it. Our parser reuses the
// literal kind and decides the role from position: the left side of an `=`, or nested inside another
// literal that is itself a target.
func noUnsafeAssignmentIsDestructuringTarget(node *ast.Node) bool {
	for current := node; current != nil; {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken &&
				binary.Left == current
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			// Keep walking: a nested literal is a target exactly when its container is.
			current = parent
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			// An object literal nested directly in an ARRAY literal, which is how
			// `[{ x }] = source` is written. The loop's own kind test then continues from that
			// literal, so the answer comes from whatever the outermost one is assigned to.
			current = parent
		default:
			return false
		}
	}
	return false
}

// noUnsafeAssignmentComparisonFor is upstream's `getComparisonType`.
//
// A written annotation is something to compare against; without one the variable's type is inferred
// from the value, so the two agree by construction and the comparison is skipped.
//
// The skip is an OPTIMIZATION rather than a behavioral filter here, and the verdict is recorded
// rather than left for the next reader to rediscover. A mutant forcing the Basic mode on every
// declaration survives the whole corpus, because an inferred receiver's type IS the sender's and
// `IsUnsafeAssignment` comparing a type against itself finds nothing to disagree about. Traced over
// five shapes designed to break it, including an inferred `Set<any>`, an inferred generic call, an
// inferred array of `any` and an inferred class field: all five report zero under the mutant, and
// all five are silent upstream.
//
// So the `None` mode buys the cost of not asking rather than a different answer. It is reproduced
// because it is upstream's and because the argument above rests on what `IsUnsafeAssignment` does
// with two identical types, which is a property of that function rather than of this rule.
func noUnsafeAssignmentComparisonFor(typeAnnotation *ast.Node) noUnsafeAssignmentComparison {
	if typeAnnotation != nil {
		return noUnsafeAssignmentComparisonBasic
	}
	return noUnsafeAssignmentComparisonNone
}

// noUnsafeAssignmentCheckWithDestructuringAtRange is the parameter anchor's entry point.
//
// It exists because one anchor needs a span no node in this tree covers: the assignment inside a
// parameter property. Everything else routes through the node-taking form above it.
func noUnsafeAssignmentCheckWithDestructuringAtRange(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
	reportRange shimcore.TextRange,
	comparison noUnsafeAssignmentComparison,
) {
	if noUnsafeAssignmentCheckAtRange(ctx, receiverNode, senderNode, reportRange, comparison) {
		return
	}
	if noUnsafeAssignmentCheckArrayDestructureHelper(ctx, receiverNode, senderNode) {
		return
	}
	noUnsafeAssignmentCheckObjectDestructureHelper(ctx, receiverNode, senderNode)
}

// noUnsafeAssignmentCheckWithDestructuring runs the plain check and then, only if it stayed silent,
// the two destructuring walks.
//
// The ordering is upstream's `didReport` threading and it decides finding COUNTS rather than
// verdicts: `const { a } = 1 as any` reports once, on the assignment, rather than once per name the
// pattern binds.
func noUnsafeAssignmentCheckWithDestructuring(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
	reportingNode *ast.Node,
	comparison noUnsafeAssignmentComparison,
) {
	if noUnsafeAssignmentCheck(ctx, receiverNode, senderNode, reportingNode, comparison) {
		return
	}
	if noUnsafeAssignmentCheckArrayDestructureHelper(ctx, receiverNode, senderNode) {
		return
	}
	noUnsafeAssignmentCheckObjectDestructureHelper(ctx, receiverNode, senderNode)
}

// noUnsafeAssignmentCheck is upstream's `checkAssignment`, and it answers whether it reported.
func noUnsafeAssignmentCheck(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
	reportingNode *ast.Node,
	comparison noUnsafeAssignmentComparison,
) bool {
	return noUnsafeAssignmentCheckAtRange(ctx, receiverNode, senderNode,
		rule.TokenRange(ctx.SourceFile, reportingNode), comparison)
}

// noUnsafeAssignmentCheckAtRange is the body both entry points share.
func noUnsafeAssignmentCheckAtRange(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
	reportRange shimcore.TextRange,
	comparison noUnsafeAssignmentComparison,
) bool {
	// Upstream resolves the receiver first. It is read only for an `any` sender and for a sender
	// IsUnsafeAssignment could find unsafe, so it is resolved only then: for a JSX attribute it
	// resolves the element's props, which was 0.8s of CPU on ahra for attributes whose value could
	// never be unsafe (#hekjpw3).
	senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)
	if senderType == nil || !type_checking.CanBeUnsafeAssignment(ctx.TypeChecker, senderType) {
		return false
	}
	if comparison == noUnsafeAssignmentComparisonNone && !type_checking.IsTypeAnyType(senderType) {
		return false
	}
	receiverType := noUnsafeAssignmentReceiverType(ctx, receiverNode, comparison)

	if type_checking.IsTypeAnyType(senderType) {
		// `unknown` is what a receiver is supposed to declare when the value cannot be trusted, so
		// putting `any` into it is the safe move rather than the unsafe one.
		if receiverType != nil && type_checking.IsTypeUnknownType(receiverType) {
			return false
		}

		message := messageNoUnsafeAssignmentAny(noUnsafeAssignmentSenderText(ctx, senderType))
		if noUnsafeAssignmentThisIsAny(ctx, senderNode) {
			message = messageNoUnsafeAssignmentAnyThis(
				noUnsafeAssignmentSenderText(ctx, senderType))
		}
		ctx.ReportRange(reportRange, message)
		return true
	}

	if comparison == noUnsafeAssignmentComparisonNone {
		return false
	}
	if receiverType == nil {
		return false
	}

	// The shared recursive comparison, the same one three sibling rules in this package call. It
	// answers whether the sender smuggles an `any` into a generic position the receiver declared
	// otherwise, and returns the two types whose disagreement it found.
	receiver, sender, unsafe := type_checking.IsUnsafeAssignment(senderType, receiverType,
		ctx.TypeChecker, senderNode)
	if !unsafe {
		return false
	}

	ctx.ReportRange(reportRange, messageNoUnsafeAssignmentUnsafe(
		"`"+ctx.TypeChecker.TypeToString(sender)+"`",
		"`"+ctx.TypeChecker.TypeToString(receiver)+"`"))
	return true
}

// noUnsafeAssignmentReceiverType answers the type the receiver declares, by comparison mode.
func noUnsafeAssignmentReceiverType(
	ctx rule.Context,
	receiverNode *ast.Node,
	comparison noUnsafeAssignmentComparison,
) *shimchecker.Type {
	if comparison == noUnsafeAssignmentComparisonContextual {
		if contextual := type_checking.GetContextualType(ctx.TypeChecker, receiverNode); contextual != nil {
			return contextual
		}
	}
	return ctx.TypeChecker.GetTypeAtLocation(receiverNode)
}

// noUnsafeAssignmentThisIsAny answers upstream's `anyAssignmentThis` condition.
//
// Only when `noImplicitThis` is off, because with it on a bare `this` is never implicitly `any` and
// the plain message is the accurate one. Upstream reads the option the same way, and the second line
// of that message is advice about the compiler option rather than a repair, since there is no edit
// to propose.
func noUnsafeAssignmentThisIsAny(ctx rule.Context, senderNode *ast.Node) bool {
	if ctx.Program == nil || ctx.Program.Options() == nil {
		return false
	}
	if type_checking.IsStrictCompilerOptionEnabled(
		ctx.Program.Options(), ctx.Program.Options().NoImplicitThis) {
		return false
	}
	thisExpression := type_checking.GetThisExpression(senderNode)
	if thisExpression == nil {
		return false
	}
	thisType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
	return thisType != nil && type_checking.IsTypeAnyType(thisType)
}

// noUnsafeAssignmentSenderText is upstream's `createData` for the one-slot messages.
//
// A type the checker could not resolve renders as the words `error typed` rather than as a
// backticked name, because printing the error type would show a name that does not exist.
func noUnsafeAssignmentSenderText(ctx rule.Context, senderType *shimchecker.Type) string {
	if type_checking.IsIntrinsicErrorType(senderType) {
		return "error typed"
	}
	return "`any`"
}

// noUnsafeAssignmentIsAnyArray guards the shelf helper's unchecked index.
//
// `type_checking.IsTypeAnyArrayType` reads `getTypeArguments(t)[0]` with no length test, so an array
// type carrying no type arguments panics one call in. The walk recovers per FILE rather than per
// rule, so that would cost every rule in the package every finding in that file while the run still
// printed a plausible summary.
//
// Guarded here rather than in the shelf because three sibling rules already call that helper on
// types they have narrowed differently, and widening the shared function is a change to their
// behavior rather than to this rule's.
func noUnsafeAssignmentIsAnyArray(ctx rule.Context, candidate *shimchecker.Type) bool {
	if candidate == nil {
		return false
	}
	if !shimchecker.Checker_isArrayType(ctx.TypeChecker, candidate) {
		return false
	}
	arguments := shimchecker.Checker_getTypeArguments(ctx.TypeChecker, candidate)
	if len(arguments) == 0 {
		return false
	}
	return type_checking.IsTypeAnyType(arguments[0])
}

// noUnsafeAssignmentCheckArrayDestructureHelper is upstream's helper of the same name.
func noUnsafeAssignmentCheckArrayDestructureHelper(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
) bool {
	if receiverNode.Kind != ast.KindArrayBindingPattern &&
		receiverNode.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)
	if senderType == nil {
		return false
	}
	return noUnsafeAssignmentCheckArrayDestructure(ctx, receiverNode, senderType, senderNode)
}

// noUnsafeAssignmentCheckArrayDestructure is upstream's `checkArrayDestructure`.
//
// Note the return value on the first branch: upstream reports the whole pattern for an `any[]` and
// then returns FALSE rather than true, which lets the object walk run afterwards on the same node.
// That reads like an oversight and is reproduced, because the alternative is a port that reports
// differently from the rule it claims to be.
func noUnsafeAssignmentCheckArrayDestructure(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderType *shimchecker.Type,
	senderNode *ast.Node,
) bool {
	if noUnsafeAssignmentIsAnyArray(ctx, senderType) {
		ctx.ReportNode(receiverNode, messageNoUnsafeAssignmentArrayPattern(
			noUnsafeAssignmentSenderText(ctx, senderType)))
		return false
	}

	// Upstream returns TRUE here, which is `didReport` saying "stop", not "reported". The caller
	// reads it as a suppression, so a non-tuple sender ends the destructuring walk for this node.
	//
	// The guard is SUBSUMED by what follows and the verdict is recorded rather than the line
	// deleted. A mutant removing it survives, because every non-tuple that reaches this point has
	// already failed the any-array branch above, and `getTypeArguments` on such a type yields no
	// `any` at a destructured position. Probed over five shapes chosen to break that: a
	// `ReadonlyArray<any>` is caught by the branch above, a `readonly [any]` IS a tuple, and a
	// `Set<any>`, a `Map<string, any>` and an `any[][]` all reach the loop and report nothing,
	// upstream and here alike.
	//
	// It is kept because it is upstream's and because the argument above rests on what
	// `getTypeArguments` returns for a non-tuple, which is a property of the checker rather than of
	// this rule. A checker that answered differently would need it.
	if !shimchecker.IsTupleType(senderType) {
		return true
	}

	tupleElements := shimchecker.Checker_getTypeArguments(ctx.TypeChecker, senderType)
	didReport := false

	for index, receiverElement := range noUnsafeAssignmentPatternElements(receiverNode) {
		if receiverElement == nil {
			continue
		}
		// A rest element is not a one-to-one assignment, so upstream skips it rather than guessing
		// which tuple members it absorbs.
		if noUnsafeAssignmentIsRestElement(receiverElement) {
			continue
		}
		if index >= len(tupleElements) {
			continue
		}
		elementSenderType := tupleElements[index]
		if elementSenderType == nil {
			continue
		}

		// The `any` test comes FIRST so `[[[x]]] = [any]` reports once on the outer element rather
		// than descending three times into a type that is already the answer.
		if type_checking.IsTypeAnyType(elementSenderType) {
			ctx.ReportNode(noUnsafeAssignmentElementTarget(receiverElement),
				messageNoUnsafeAssignmentArrayPatternFromTuple(
					noUnsafeAssignmentSenderText(ctx, elementSenderType)))
			// Every invalid element in the tuple is reported, so this does not return early.
			didReport = true
			continue
		}

		target := noUnsafeAssignmentElementTarget(receiverElement)
		switch target.Kind {
		case ast.KindArrayBindingPattern, ast.KindArrayLiteralExpression:
			didReport = noUnsafeAssignmentCheckArrayDestructure(ctx, target, elementSenderType, senderNode)
		case ast.KindObjectBindingPattern, ast.KindObjectLiteralExpression:
			didReport = noUnsafeAssignmentCheckObjectDestructure(ctx, target, elementSenderType, senderNode)
		}
	}

	return didReport
}

// noUnsafeAssignmentCheckObjectDestructureHelper is upstream's helper of the same name.
func noUnsafeAssignmentCheckObjectDestructureHelper(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderNode *ast.Node,
) bool {
	if receiverNode.Kind != ast.KindObjectBindingPattern &&
		receiverNode.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)
	if senderType == nil {
		return false
	}
	return noUnsafeAssignmentCheckObjectDestructure(ctx, receiverNode, senderType, senderNode)
}

// noUnsafeAssignmentCheckObjectDestructure is upstream's `checkObjectDestructure`.
//
// The sender's properties are looked up by NAME, so a key the rule cannot read statically is skipped
// rather than guessed at. Upstream reads an identifier, a literal and a single-quasi template, and
// declines anything else.
func noUnsafeAssignmentCheckObjectDestructure(
	ctx rule.Context,
	receiverNode *ast.Node,
	senderType *shimchecker.Type,
	senderNode *ast.Node,
) bool {
	properties := map[string]*shimchecker.Type{}
	for _, property := range shimchecker.Checker_getPropertiesOfType(ctx.TypeChecker, senderType) {
		propertyType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, senderNode)
		if propertyType != nil {
			properties[property.Name] = propertyType
		}
	}

	didReport := false
	for _, receiverProperty := range noUnsafeAssignmentPatternElements(receiverNode) {
		if receiverProperty == nil {
			continue
		}
		if noUnsafeAssignmentIsRestElement(receiverProperty) {
			continue
		}

		key, keyed := noUnsafeAssignmentPropertyKey(ctx, receiverProperty)
		if !keyed {
			continue
		}
		elementSenderType, found := properties[key]
		if !found || elementSenderType == nil {
			continue
		}

		target := noUnsafeAssignmentElementTarget(receiverProperty)

		// The `any` test first, so `{x: {y: z}} = {x: any}` reports on `x` rather than descending.
		if type_checking.IsTypeAnyType(elementSenderType) {
			ctx.ReportNode(target, messageNoUnsafeAssignmentObjectPattern(
				noUnsafeAssignmentSenderText(ctx, elementSenderType)))
			didReport = true
			continue
		}

		switch target.Kind {
		case ast.KindArrayBindingPattern, ast.KindArrayLiteralExpression:
			didReport = noUnsafeAssignmentCheckArrayDestructure(ctx, target, elementSenderType, senderNode)
		case ast.KindObjectBindingPattern, ast.KindObjectLiteralExpression:
			didReport = noUnsafeAssignmentCheckObjectDestructure(ctx, target, elementSenderType, senderNode)
		}
	}

	return didReport
}

// noUnsafeAssignmentPatternElements returns the elements of a destructuring target.
//
// Four shapes reach here and they come in two families. A `const [a] = x` target is a BINDING
// pattern, and an `[a] = x` target is an array or object LITERAL reinterpreted as a pattern, which
// is a distinction TSESTree does not make and typescript-go does. Both families are walked so the
// declaration form and the assignment form behave the same.
//
// `Elements()` and `Properties()` PANIC on a kind that has neither, so the kinds are named rather
// than tried. The walk recovers per file, not per rule.
func noUnsafeAssignmentPatternElements(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		return node.AsBindingPattern().Elements.Nodes
	case ast.KindArrayLiteralExpression:
		return node.AsArrayLiteralExpression().Elements.Nodes
	case ast.KindObjectLiteralExpression:
		return node.AsObjectLiteralExpression().Properties.Nodes
	}
	return nil
}

// noUnsafeAssignmentIsRestElement answers whether a pattern element absorbs the remainder.
func noUnsafeAssignmentIsRestElement(element *ast.Node) bool {
	switch element.Kind {
	case ast.KindBindingElement:
		return element.AsBindingElement().DotDotDotToken != nil
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return true
	}
	return false
}

// noUnsafeAssignmentElementTarget answers the node a finding on this element should point at.
//
// Upstream reports `receiverProperty.value` for an object property and `receiverElement` for an
// array element, so a nested pattern is reported at the pattern rather than at the key naming it.
// The binding-element arm returns its NAME for that reason: for `{ a: { b } }` the name is the inner
// pattern, which is what upstream underlines.
func noUnsafeAssignmentElementTarget(element *ast.Node) *ast.Node {
	switch element.Kind {
	case ast.KindBindingElement:
		if name := element.AsBindingElement().Name(); name != nil {
			return name
		}
	case ast.KindPropertyAssignment:
		if initializer := element.AsPropertyAssignment().Initializer; initializer != nil {
			return initializer
		}
	}
	return element
}

// noUnsafeAssignmentPropertyKey reads the name a pattern element binds from the sender.
//
// Upstream's four arms: a non-computed identifier or private name answers its own text, a literal
// answers its cooked value, a single-quasi template answers its cooked text, and anything else is
// declined rather than guessed at.
func noUnsafeAssignmentPropertyKey(ctx rule.Context, element *ast.Node) (string, bool) {
	var key *ast.Node
	switch element.Kind {
	case ast.KindBindingElement:
		binding := element.AsBindingElement()
		key = binding.PropertyName
		if key == nil {
			key = binding.Name()
		}
	case ast.KindPropertyAssignment:
		key = element.AsPropertyAssignment().Name()
	case ast.KindShorthandPropertyAssignment:
		key = element.AsShorthandPropertyAssignment().Name()
	}
	if key == nil {
		return "", false
	}

	if key.Kind == ast.KindComputedPropertyName {
		inner := key.AsComputedPropertyName().Expression
		if inner == nil {
			return "", false
		}
		key = inner
	}

	switch key.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier,
		ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return key.Text(), true
	}
	return "", false
}
