package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeMemberAccess flags reading a property off a value typed `any`.
//
//	valid:   declare const x: { a: number }; x.a;
//	valid:   declare const x: any; x?.a;                   with allowOptionalChaining
//	valid:   declare const x: { a: number }; x[1];
//	invalid: declare const x: any; x.a;
//	invalid: declare const x: { a: number }; declare const k: any; x[k];
//	invalid: declare const x: any; x.a.b.c;                reports once, on `.a`
//
// Reading a property off an `any` is where a chain stops being checked. Nothing downstream of that
// dot is verified, and the type system will not mention it again, so a typo in the property name
// survives to runtime.
//
// # Six messages across two anchors, and the pairing carries the judgment
//
// A property access reports on the property, saying whether the object was an `any` the author
// wrote, a type the checker could not resolve, or an implicitly-`any` `this`. A computed access
// reports a second, separate thing: the KEY itself being `any`, which is a different defect at the
// same site. `x[k]` where `k` is `any` is unsafe even when `x` is perfectly typed.
//
// # One report per chain, and the innermost one
//
// `x.a.b.c` on an `any` reports ONCE, on `.a`, because the outer accesses are unsafe only as a
// consequence. Upstream gets that with a memoized recursion: each access asks its object first, and
// an object that came back unsafe short-circuits without reporting again. Measured on the installed
// 8.67.0 build, which reports exactly one finding spanning `a`.
//
// That recursion is also why the cache exists rather than being an optimization. Our walk is
// pre-order, so the OUTERMOST access is visited first and recurses inward; the inner ones are then
// visited again by the walk and must not report a second time. The cache is what makes the second
// visit a lookup.
//
// # Upstream's heritage exclusion has no shape to exclude here
//
// Upstream's selector excludes a member expression inside a class `implements` or an interface
// `extends`, because estree parses `implements Foo.Bar` as a MemberExpression and the rule would
// otherwise type it as a value. Probed here: our parser gives that a `KindTypeReference` holding a
// qualified name, which is not a property access at all and never reaches either listener. The
// exclusion is therefore not ported, and that is a shape difference rather than a dropped rule.
// Measured on the installed build to confirm the verdict rather than only the mechanism: both
// `class C implements Foo.Bar {}` and `interface I extends Foo.Bar {}` are clean there, and both
// are clean here.
//
// # `allowOptionalChaining` marks a link chained, not safe
//
// With the option on, `x?.a` is skipped. It is NOT treated as safe, and the difference shows one
// link further out: measured, `x?.a.b` with the option on still reports on `.b`, because the
// chained state is recorded and then the outer access asks the checker about its own object anyway.
// A port that recorded "safe" would go silent on that second access.
//
// # The `this` branch, and why no fixture in this package can reach it
//
// Upstream runs this corpus under `tsconfig.noImplicitThis.json` and our harness pins `strict:
// true`. Measured against the installed 8.67.0 build under BOTH settings: thirty-five of
// thirty-six cases are identical, and the object-literal `this` case reports three
// `unsafeThisMemberExpression` findings under upstream's configuration and is COMPLETELY SILENT
// under ours, because `this` in an object literal is typed as the literal rather than `any` when
// the option is on. The branch is ported and the fixtures assert the silence our harness can
// observe, with the other column recorded here.
//
// # Cost
//
// Two anchors, both extremely common, and this is the most expensive rule in this batch on the real
// tree. The cache keeps a chain to one checker question per link rather than one per link per
// visit, and the computed-key listener declines a literal and an update expression before asking
// anything, which is upstream's own optimization and is worth keeping for the same reason.
var NoUnsafeMemberAccess = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-member-access",

	// Every judgment is a type question about an object or a computed key.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoUnsafeMemberAccessOptions](options)
		if !ok {
			// A rule configured as a bare "error" is handed nil, and the zero value of the settings
			// struct happens to be upstream's default here because the only option defaults to
			// false. Stated rather than relied on silently: if a second option is ever added with a
			// true default, this line is where that breaks.
			settings = DefaultNoUnsafeMemberAccessOptions()
		}

		// stateCache is upstream's `stateCache`. See the doc comment: it is load-bearing rather
		// than an optimization, because our pre-order walk visits an inner access twice.
		stateCache := map[*ast.Node]noUnsafeMemberAccessState{}

		var checkMemberAccess func(node *ast.Node) noUnsafeMemberAccessState

		// checkMemberAccessInside descends into whatever member access an expression contains,
		// looking through a call's callee, so an access written earlier in the source is visited
		// before the one that encloses it. It exists only to fix the ORDER of findings; the state it
		// computes is discarded.
		checkMemberAccessInside := func(expression *ast.Node) {
			switch expression.Kind {
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				checkMemberAccess(expression)
			case ast.KindCallExpression:
				callee := expression.AsCallExpression().Expression
				if callee != nil {
					switch callee.Kind {
					case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
						checkMemberAccess(callee)
					}
				}
			}
		}

		checkMemberAccess = func(node *ast.Node) noUnsafeMemberAccessState {
			objectExpression, propertyNode, isComputed, hasQuestionDot :=
				noUnsafeMemberAccessParts(node)
			if objectExpression == nil || propertyNode == nil {
				return noUnsafeMemberAccessSafe
			}

			if cached, seen := stateCache[node]; seen {
				return cached
			}

			// The chained short-circuit descends BEFORE it returns, which upstream gets for free and
			// this port has to arrange.
			//
			// Upstream's visitor reaches every access on its own, so an inner one is reported by
			// that visit even when an outer one skipped it. Ours reports through this recursion, so
			// returning here without descending leaves the inner access to be reported by the walk
			// LATER, and `value.outer?.middle.inner` comes out inner-then-outer against upstream's
			// outer-then-inner. Measured against the installed 8.67.0 build, which emits every case
			// in source order.
			//
			// Descending first and then recording Chained gives the same state to the caller and the
			// same findings in the same order.
			if settings.AllowOptionalChaining && hasQuestionDot {
				checkMemberAccessInside(objectExpression)
				stateCache[node] = noUnsafeMemberAccessChained
				return noUnsafeMemberAccessChained
			}

			// An access reached THROUGH a call has to be visited too, and it is not the object of
			// this access. `x.a().b` puts a call between the two accesses, so the object of `.b` is
			// a call expression and descending only through member accesses never reaches `.a`.
			// Upstream's visitor reaches it independently, so it reports both in source order;
			// measured on the installed build, which reports `a` then `b`.
			//
			// It is descended for ORDER rather than for the verdict: the call's own state does not
			// propagate, since calling an `any` yields an `any` and the object test below decides
			// this access on its own.
			if objectExpression.Kind == ast.KindCallExpression {
				checkMemberAccessInside(objectExpression)
			}

			if objectExpression.Kind == ast.KindPropertyAccessExpression ||
				objectExpression.Kind == ast.KindElementAccessExpression {
				// Recursing FIRST is what puts the findings in source order, and it is not merely
				// tidier. Upstream's visitor reaches the outermost access first too, and its
				// recursion descends before it reports, so an inner finding is emitted before an
				// outer one. Nothing downstream re-sorts a rule's findings, so a port that reported
				// on the way in emits `value.outer?.middle.inner` as inner-then-outer while
				// upstream emits outer-then-inner. Measured against the installed 8.67.0 build,
				// whose findings are in source order on all nineteen reporting cases.
				if checkMemberAccess(objectExpression) == noUnsafeMemberAccessUnsafe {
					// The object is already unsafe and has already been reported on, so this access
					// is unsafe only as a consequence. Recording and returning without reporting is
					// what keeps `x.a.b.c` to one finding.
					stateCache[node] = noUnsafeMemberAccessUnsafe
					return noUnsafeMemberAccessUnsafe
				}
			}

			objectType := ctx.TypeChecker.GetTypeAtLocation(objectExpression)
			state := noUnsafeMemberAccessSafe
			if objectType != nil && type_checking.IsTypeAnyType(objectType) {
				state = noUnsafeMemberAccessUnsafe
			}
			stateCache[node] = state
			if state != noUnsafeMemberAccessUnsafe {
				return state
			}

			propertyText := noUnsafeMemberAccessText(ctx, propertyNode)
			rendered := "." + propertyText
			if isComputed {
				rendered = "[" + propertyText + "]"
			}

			messageId := ""
			if !type_checking.IsStrictCompilerOptionEnabled(
				ctx.Program.Options(), ctx.Program.Options().NoImplicitThis) {
				// See the doc comment: correct, and unreachable through this package's fixtures
				// because the harness pins the option above the rule.
				thisExpression := type_checking.GetThisExpression(node)
				if thisExpression != nil {
					thisType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
					if thisType != nil && type_checking.IsTypeAnyType(thisType) {
						messageId = "unsafeThisMemberExpression"
						if type_checking.IsIntrinsicErrorType(thisType) {
							messageId = "errorThisMemberExpression"
						}
					}
				}
			}
			if messageId == "" {
				messageId = "unsafeMemberExpression"
				if type_checking.IsIntrinsicErrorType(objectType) {
					// The error type carries the `any` flag, so only the intrinsic name separates
					// them. Getting this backwards renames every unresolved-type finding into one
					// claiming an `any` the reader never wrote.
					messageId = "errorMemberExpression"
				}
			}

			// Upstream reports on the PROPERTY rather than on the whole access, so `x.a.b.c`
			// underlines `a` alone.
			ctx.ReportNode(propertyNode, noUnsafeMemberAccessMessageFor(messageId, rendered))
			return state
		}

		// checkComputedKey is upstream's second selector, which asks a different question at the
		// same site: not whether the OBJECT is `any`, but whether the KEY is.
		checkComputedKey := func(node *ast.Node) {
			access := node.AsElementAccessExpression()
			key := access.ArgumentExpression
			if key == nil {
				return
			}
			if settings.AllowOptionalChaining && access.QuestionDotToken != nil {
				return
			}

			switch noUnsafeMemberAccessUnwrapParentheses(key).Kind {
			case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral,
				ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral,
				ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
				ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
				// Upstream's own optimization, and it names the reason for the second half: every
				// update expression is typed `number` regardless of its argument, because the engine
				// answers NaN rather than throwing. A literal obviously cannot be `any` either.
				//
				// Our parser splits estree's single `Literal` across several kinds and its
				// `UpdateExpression` across the prefix and postfix ones, so the list is longer here
				// while naming the same two categories. `KindPrefixUnaryExpression` also covers
				// `-1` and `!x`, which estree models as a UnaryExpression and does NOT skip.
				//
				// This whole arm is EQUIVALENT and no fixture can see it, which a surviving mutant
				// established and a probe then explained rather than left as a blind spot. Measured
				// across every kind listed here, including with an `any` operand: `anyKey++` types
				// as `number`, `-anyKey` as `number`, `!anyKey` as `boolean`, and each literal as
				// its own literal type. None can be `any`, so the checker declines them one line
				// below and the skip only saves the question. It is kept because upstream keeps it
				// and because this is the most-visited anchor in the batch, where one avoided
				// checker call per computed literal key is real.
				return
			}

			// estree has no parenthesized-expression node, so upstream's selector binds to the inner
			// expression and its `getText` renders it without the parentheses. Our parser keeps the
			// node, so both the reported SPAN and the interpolated property text would carry them.
			// Measured on the installed build: `x[(y += 1)]` reports on `y += 1` and renders
			// `[y += 1]`, without the parentheses either time.
			//
			// A loop rather than one step, since `((y))` nests, and written out rather than calling
			// ast.SkipParentheses, which dereferences its argument.
			key = noUnsafeMemberAccessUnwrapParentheses(key)
			if key == nil {
				return
			}

			keyType := ctx.TypeChecker.GetTypeAtLocation(key)
			if keyType == nil || !type_checking.IsTypeAnyType(keyType) {
				return
			}
			messageId := "unsafeComputedMemberAccess"
			if type_checking.IsIntrinsicErrorType(keyType) {
				messageId = "errorComputedMemberAccess"
			}
			ctx.ReportNode(key, noUnsafeMemberAccessMessageFor(messageId,
				"["+noUnsafeMemberAccessText(ctx, key)+"]"))
		}

		if ctx.TypeChecker == nil {
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberAccess(node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkMemberAccess(node)
				checkComputedKey(node)
			},
		}
	},
}

// noUnsafeMemberAccessState is upstream's `State`.
//
// `Chained` is distinct from `Safe` on purpose and the difference is observable: a chained link is
// skipped, while a safe one has been checked. Measured, `x?.a.b` with `allowOptionalChaining` still
// reports on `.b`, which a port collapsing the two would miss.
type noUnsafeMemberAccessState uint8

const (
	noUnsafeMemberAccessUnsafe noUnsafeMemberAccessState = iota + 1
	noUnsafeMemberAccessSafe
	noUnsafeMemberAccessChained
)

// noUnsafeMemberAccessParts splits an access into the pieces both anchors need.
//
// estree has one MemberExpression with a `computed` flag; our parser has two node kinds, so this is
// where the two shapes meet.
func noUnsafeMemberAccessParts(node *ast.Node) (
	objectExpression *ast.Node, propertyNode *ast.Node, isComputed bool, hasQuestionDot bool,
) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		return access.Expression, access.Name(), false, access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		return access.Expression, access.ArgumentExpression, true, access.QuestionDotToken != nil
	}
	return nil, nil, false, false
}

// noUnsafeMemberAccessText reads a node's own source text.
//
// Upstream calls `sourceCode.getText(node)`. Deliberately NOT `node.Text()`, which panics on a
// property access or a call expression and would take every rule's findings for that whole file
// with it, since the walk recovers per file rather than per rule.
func noUnsafeMemberAccessText(ctx rule.Context, node *ast.Node) string {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

// noUnsafeMemberAccessMessageFor renders one of the rule's six messages.
//
// The two `this` texts are two lines joined by a newline, which is upstream's own shape: the second
// line is advice about the compiler option rather than a separate suggestion, because there is no
// edit to propose.
func noUnsafeMemberAccessMessageFor(messageId string, property string) rule.Message {
	const thisAdvice = "\nYou can try to fix this by turning on the `noImplicitThis` compiler " +
		"option, or adding a `this` parameter to the function."
	switch messageId {
	case "errorComputedMemberAccess":
		return rule.Message{Id: messageId,
			Description: "The type of computed name " + property + " cannot be resolved."}
	case "errorMemberExpression":
		return rule.Message{Id: messageId,
			Description: "Unsafe member access " + property + " on a type that cannot be resolved."}
	case "errorThisMemberExpression":
		return rule.Message{Id: messageId,
			Description: "Unsafe member access " + property +
				". The type of `this` cannot be resolved." + thisAdvice}
	case "unsafeComputedMemberAccess":
		return rule.Message{Id: messageId,
			Description: "Computed name " + property + " resolves to an `any` value."}
	case "unsafeThisMemberExpression":
		return rule.Message{Id: messageId,
			Description: "Unsafe member access " + property +
				" on an `any` value. `this` is typed as `any`." + thisAdvice}
	default:
		return rule.Message{Id: "unsafeMemberExpression",
			Description: "Unsafe member access " + property + " on an `any` value."}
	}
}

// NoUnsafeMemberAccessOptions is the rule's option surface.
type NoUnsafeMemberAccessOptions struct {
	// AllowOptionalChaining permits `?.` on an `any`, on the argument that the author already
	// signalled they know the value may not be there.
	AllowOptionalChaining bool
}

// DefaultNoUnsafeMemberAccessOptions is upstream's `allowOptionalChaining: false`.
func DefaultNoUnsafeMemberAccessOptions() NoUnsafeMemberAccessOptions {
	return NoUnsafeMemberAccessOptions{AllowOptionalChaining: false}
}

// noUnsafeMemberAccessRawOptions is the wire shape.
type noUnsafeMemberAccessRawOptions struct {
	AllowOptionalChaining *bool `json:"allowOptionalChaining"`
}

// DecodeNoUnsafeMemberAccessOptions maps the wire key onto the setting the rule reads.
//
// A pointer wire field even though the default is FALSE and the Go zero value would agree. The brief
// warns about a default-true option silently inverting under a generic decode; the same shape read
// the other way is what makes an explicit `false` distinguishable from an absent key, and the day
// this option's default changes upstream, this decoder is the one line that has to move rather than
// every call site.
func DecodeNoUnsafeMemberAccessOptions(raw []byte) (any, error) {
	settings := DefaultNoUnsafeMemberAccessOptions()
	if len(raw) == 0 {
		return settings, nil
	}
	var wire noUnsafeMemberAccessRawOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}
	if wire.AllowOptionalChaining != nil {
		settings.AllowOptionalChaining = *wire.AllowOptionalChaining
	}
	return settings, nil
}

// noUnsafeMemberAccessUnwrapParentheses peels every parenthesis off an expression.
//
// A loop rather than one step, because `((y))` nests. Written out rather than calling
// ast.SkipParentheses, which dereferences its argument and is how this project once lost 167 files
// to a nil panic; the caller here can hold nil.
func noUnsafeMemberAccessUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}
