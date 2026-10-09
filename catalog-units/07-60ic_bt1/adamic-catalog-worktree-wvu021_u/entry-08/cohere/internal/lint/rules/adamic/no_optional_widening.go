package adamic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/checking/flow"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// noOptionalWideningText is the rule's message, whose wording lives in
// `policy/messages/no-optional-widening.json`.
var noOptionalWideningText = policy.MessageOf("adamic/no-optional-widening", "optionalWidening")

/*
 * NoOptionalWidening reports a value seen as a type with an optional property its own type does not
 * declare (#drbrp8c).
 *
 *     invalid: const wide: { x: number; y?: number } = narrow;    narrow: { x: number }
 *     valid:   const wide: { x: number; y?: number } = { x: 1 };  a literal has exactly its keys
 *     valid:   const wide: { x: number; y?: number } = point;     const point = { x: 1 }
 *
 * # The hole
 *
 * Width subtyping lets `{ x: 1, y: 'surprise' }` be seen as `{ x: number }`, which drops `y` from the
 * type and not from the value. tsc then lets `{ x: number }` be seen as `{ x: number; y?: number }`,
 * because an optional property may be absent, and `y` is typed `number | undefined` while holding a
 * string (probe h05 on #drbrp8c: `(wide.y ?? 0).toFixed is not a function` on Node). The first step is
 * everywhere and harmless alone; the second is where the type starts lying, so this rule reports it.
 *
 * # What is exempt
 *
 * A fresh literal carries exactly its keys. So does a `const` with no annotation whose initializer is an
 * object literal, since its type is the literal's and a const is never reassigned; that exemption covers
 * the value itself, not what it holds. A class instance target is nominal-class's.
 *
 * # Noise, measured before it counts
 *
 * Expected to be the noisiest of the set: an options bag declared narrowly and passed to a wide options
 * type (`fetch(url, init)`, `init: { method: string }`) reports once per such call. That is the hole,
 * faithfully. Per #drbrp8c's ruling, the hundred measures it before it counts toward readiness.
 *
 * # No fix
 *
 * The repair is a type or a construction only the author can choose.
 */
var NoOptionalWidening = rule.Rule{
	Name:             "adamic/no-optional-widening",
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return nil
		}
		typeChecker := ctx.TypeChecker
		walker := flow.WalkerFor(ctx)
		// The judge, what it reads and what it reports are made once per file and set at each site, not a
		// closure per site (#m6tyg79).
		var exactAtTop bool
		var missing *ast.Symbol
		judge := func(pair flow.Pair) (bool, bool) {
			if isClassInstance(pair.Target) {
				return false, false
			}
			if pair.Target.Flags()&checker.TypeFlagsObject == 0 || pair.Source.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsAny) != 0 {
				return false, true
			}
			// A tuple's optional element is a position its fixed length rules out: tsc refuses `[number]`
			// as `[]`, so a value typed `[]` holds no element 0 (probe t6 on #drbrp8c). A function's optional
			// members stay judged, since Object.assign can hand a function any of them (probe t5).
			if flow.IsArrayLike(typeChecker, pair.Target) {
				return false, true
			}
			if exactAtTop && len(pair.Path) == 0 {
				return false, true
			}
			apparent := checker.Checker_getApparentType(typeChecker, pair.Source)
			if apparent == nil || apparent.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
				return false, true
			}
			for _, property := range checker.Checker_getPropertiesOfType(typeChecker, pair.Target) {
				if property.Flags&ast.SymbolFlagsOptional == 0 {
					continue
				}
				if checker.Checker_getPropertyOfType(typeChecker, apparent, property.Name) == nil {
					missing = property
					return true, false
				}
			}
			return false, true
		}
		return walker.Listeners(func(site flow.Site) {
			if site.Spread || site.Method {
				// A spread into a literal and a literal's method are invariant-mutable's; see flow.Site.
				return
			}
			if site.Fresh {
				return
			}
			exactAtTop, missing = isConstLiteralAlias(typeChecker, site.Node), nil
			found, wrong := walker.Walk(site, judge)
			if !wrong {
				return
			}
			ctx.ReportNode(site.Node, rule.Message{
				Id: "optionalWidening",
				Description: noOptionalWideningText.Render(map[string]string{
					"slot":         slotText(ctx.SourceFile, site.Node, found.Path),
					"source":       type_checking.StableTypeText(typeChecker, found.Source),
					"target":       type_checking.StableTypeText(typeChecker, found.Target),
					"property":     type_checking.StablePropertyName(missing.Name),
					"propertyType": type_checking.StableTypeText(typeChecker, checker.Checker_getTypeOfSymbol(typeChecker, missing)),
				}),
			})
		})
	},
}

// isConstLiteralAlias is an identifier bound to a `const` with no annotation whose initializer is an
// object literal: its type is exactly the literal's keys.
func isConstLiteralAlias(typeChecker *checker.Checker, node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier {
		return false
	}
	symbol := typeChecker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	return variable.Type == nil && variable.Initializer != nil &&
		ast.SkipParentheses(variable.Initializer).Kind == ast.KindObjectLiteralExpression
}
