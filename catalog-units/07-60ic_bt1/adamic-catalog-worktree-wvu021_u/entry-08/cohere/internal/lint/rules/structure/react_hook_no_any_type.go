package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactHookNoAnyTypeMessage is built per finding, because it names the hook and the rules that lost
// their sight, and neither is fixed. See `reactHookNoAnyTypeBlindedRules` for why the list is derived.
var reactHookNoAnyTypeMessageId = "reactHookNoAnyType"

// ReactHookNoAnyType flags a hook-shaped call whose result the checker cannot type.
//
//	valid:   import React from 'react'; React.useState(0)          resolves to a real tuple
//	valid:   import React from 'react'; React.useEffect(fn, [])    resolves to `void`
//	valid:   somethingElse.useCache(key)                            not a React hook shape
//	invalid: declare function useState<T>(initial: T): any; useState(0)
//	invalid: useState(0)                                            with no react types resolvable
//
// # What this checks, exactly
//
// For every call expression whose callee is either a bare identifier matching React's own hook
// pattern `/^use[A-Z0-9]/`, or a property access `React.<thatPattern>` — the two spellings a hook
// can have in this codebase — the rule asks the checker for the type of THE CALL, meaning the value
// the call produces, and reports when that type carries `TypeFlagsAny`.
//
// It checks the call's result and deliberately not the callee's type, and the difference is the
// whole rule rather than a detail. Measured on the exact input this rule was written for:
//
//	declare function useState<T>(initial: T): any;
//	const [count, setCount] = useState(0);
//
//	callee type    <T>(initial: T) => any     flags = Object      NOT any
//	result type    any                        flags = Any         any
//
// A rule keyed on the callee would be silent on that file, which is the file Kirk brought. The
// callee is a perfectly well typed generic function; what is destroyed is what it hands back, and
// what it hands back is the only thing any downstream rule reads.
//
// It is not an import check, and that is a constraint rather than an omission. A text search for
// hook-using files lacking a react import returns one file on Kirk's tree and that entire file is
// commented out, so the `React.useState` it matched is inside a comment block. A checker cannot see
// a comment. More importantly the import shape is a proxy pointing the wrong way in both
// directions: `declare global`, a re-export, a monorepo path mapping and an ambient declaration all
// produce real types with no local import statement and are legitimate, while a plain missing
// import is a `tsc` error that never reaches production. The cases that DO reach production are the
// ones that satisfy `tsc` — a stale `@types` package, a path mapping that half-resolves, a shim
// somebody wrote to unblock a build — and every one of those has a resolvable-looking import.
//
// `structure/import-require-react-namespace` is the neighbouring rule and the two do not overlap.
// That one polices HOW a name is imported, and the input above has no import statement at all for
// it to see. This one polices WHETHER the types behind the name are real. Neither substitutes.
//
// # Why only `any`, and what else was probed
//
// `any` is reported and nothing else is, established by probe rather than assumption. Every
// degraded arrangement collapses onto exactly `TypeFlagsAny` at the call result:
//
//	ambient shim returning any     result any        Any        reported
//	no declaration at all          result any        Any        reported
//	import from an unresolved      result any        Any        reported
//	  module specifier
//
// The three near neighbours are each a legitimate answer rather than an absence, and reporting them
// would fire on healthy code:
//
//	useEffect healthy              result void       Void       silent, and `void` is what a
//	                                                            correctly typed effect returns
//	declare function useX(): unknown    result unknown   Unknown  silent; `unknown` is a deliberate
//	                                                            annotation somebody wrote
//	declare function useX(): never      result never     Never    silent; `never` is deliberate too
//
// `unknown` is the one worth arguing about, and it is excluded because it is the opposite failure
// from `any`: `any` is the checker giving up, while `unknown` is an author refusing to guess and
// forcing every reader to narrow. Nothing in the tree writes a hook returning `unknown`, so
// including it would buy no findings and cost the rule its precision claim. An error type was
// probed for separately and does not arise here: a call to something the checker cannot resolve at
// all comes back as `any`, not as an error type, so the `any` test already covers it.
//
// # Why the hook shape is load-bearing rather than a filter
//
// `any` density measured across three production trees is 1.8%, 2.9% and 2.4%, so a rule reporting
// every `any`-typed call would be a tree-wide noise generator rather than a tripwire. Probed
// directly: `declare function used(x: number): any; used(1)` also produces an `any` result, and it
// is not a hook and not this rule's business. The shape test is what makes the rule mean "a React
// analysis went blind here" rather than "there is an `any` here".
//
// The pattern is React's own, `/^use[A-Z0-9]/`, and it is written here rather than taken from
// `react.IsHookName`, which rejects a digit and would go silent on `use2Things`. That shelf helper
// has thirteen callers whose semantics were settled against other authorities, so it is left alone
// and this rule carries its own predicate — the same conclusion `react/unsupported-syntax` and
// `react/void-use-memo` reached independently. Kirk's tree contains zero `use<digit>` names against
// a control of 4,055 `use<uppercase>` names, so matching React here buys nothing today and costs
// nothing, and it means a `use2Things` file cannot slip through unguarded later.
//
// # Scope: every TypeScript file, not only `.tsx`
//
// Kirk settled that `.jsx`/`.tsx` may be assumed to mean React, but restricting to those extensions
// would miss the files this rule most exists for. Measured on his tree: 74 `.ts` files call
// `React.use<Something>(` directly, and reading them shows genuine hooks — `useMessagesFeedRequest.ts`
// declares four `React.useState` setters and renders no JSX at all, because the JSX lives in the
// `.tsx` component that imports it. A custom-hook file is exactly where a setter is created and
// exactly where a type failure would be invisible. So the rule runs wherever the walk offers it a
// file and lets the hook shape do the discriminating, which is the same reason the shape test
// exists.
//
// # The message names the cost, and derives it
//
// A message reading "`useState` resolves to any" is a type-cleanliness nit and gets suppressed. The
// message here names which correctness rules are disabled by that suppression, so silencing the
// tripwire is an explicit choice rather than a tidy-up. The list is read off the live rule catalog
// at report time rather than written into a string, because a hardcoded list is wrong the week a
// new checker-based rule ships and it is confidently wrong: it keeps naming the old set while
// reading as current. See `reactHookNoAnyTypeBlindedRules`.
var ReactHookNoAnyType = rule.Rule{
	Name:             "structure/react-hook-no-any-type",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The typed-rule guard. A rule that reads a nil checker here would decline every file
		// vacuously, which is precisely the silence this rule exists to make loud.
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			// No second nil-checker guard here, and the absence is measured rather than an
			// oversight. Two guards were written, one at Run and one at the top of this listener,
			// and each SURVIVED its own mutation while neutralizing both together was caught: the
			// listener is only reached through a Run that already declined a nil checker, so the
			// inner test could not answer any input the outer one had not. Mutual subsumption, so
			// the reachable one stays and the unreachable one is gone rather than tested.
			//
			// The outer one is the one worth keeping because it is the cheaper path: declining at
			// Run registers no listener at all, where an inner guard pays a call per node.
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil {
					return
				}
				if !reactHookNoAnyTypeIsHookCall(call.Expression) {
					return
				}

				// The CALL's type, not the callee's. See the rule comment: an ambient shim
				// declaring `useState<T>(initial: T): any` gives a perfectly typed callee.
				resultType := ctx.TypeChecker.GetTypeAtLocation(node)

				// Crash protection rather than a behavioural filter, and it is a SURVIVING mutant
				// recorded as such rather than argued into equivalence. `shimchecker.Type_flags`
				// casts through an unsafe pointer with no nil test of its own, so a nil here is a
				// segfault rather than a wrong answer, and no `ExpectFindings` fixture can see a
				// panic — the sweep reports SURVIVED identically whether this line matters or not.
				//
				// Whether a nil is reachable at all is NOT established. Five shapes were probed
				// looking for one — a member call on an `any`, an optional call, a comma-sequence
				// callee, an unterminated call that error recovery synthesizes, and a spread
				// argument — and every one came back with a real type. That is evidence the branch
				// is hard to reach, and it is not proof that it cannot be, so the guard stays.
				// Deleting it would trade a measured cost of one nil comparison per hook call for
				// an unmeasured risk of taking a lint run down.
				if resultType == nil {
					return
				}
				if shimchecker.Type_flags(resultType)&shimchecker.TypeFlagsAny == 0 {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id:          reactHookNoAnyTypeMessageId,
					Description: reactHookNoAnyTypeDescription(reactHookNoAnyTypeCalleeName(call.Expression)),
				})
			},
		}
	},
}

// reactHookNoAnyTypeIsHookCall reports whether a callee is one of the two hook spellings.
//
// A bare identifier covers a destructured or ambiently declared hook, and `React.<name>` covers
// house style, which is `React.useState` by 599 occurrences to 3 across Kirk's `.tsx` files.
// `react.IsNamespacedMember` is reached for rather than hand-rolled because it additionally skips
// parentheses on the receiver, so `(React).useState(0)` is the same call and a hand-rolled receiver
// test declines it while looking correct.
//
// Any other callee shape is deliberately not a hook here. `somethingElse.useCache(key)` reads as a
// method on an unrelated object, and treating every namespaced `use*` name as a React hook would
// report on code that has nothing to do with React — the same judgment `isHookCall` in
// `react_detection.go` makes for the same reason.
func reactHookNoAnyTypeIsHookCall(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return reactHookNoAnyTypeIsHookName(callee.Text())
	case ast.KindPropertyAccessExpression:
		return react.IsNamespacedMember(callee, reactHookNoAnyTypeIsHookName)
	}
	return false
}

// reactHookNoAnyTypeIsHookName answers React's own `/^use[A-Z0-9]/`.
//
// Byte comparisons rather than `unicode.IsUpper`, because React's character class is ASCII: `useÉ`
// is not a hook to React and a Unicode test would answer that it is. The digit arm is the half
// `react.IsHookName` gets wrong in the other direction, which is why this is written here rather
// than lifted — see the rule comment.
func reactHookNoAnyTypeIsHookName(name string) bool {
	if len(name) < 4 {
		return false
	}
	if name[0:3] != "use" {
		return false
	}
	fourth := name[3]
	return (fourth >= 'A' && fourth <= 'Z') || (fourth >= '0' && fourth <= '9')
}

// reactHookNoAnyTypeCalleeName renders the callee the way the source spells it, for the message.
//
// `Node.Text()` panics on a PropertyAccessExpression, which is why the two arms are built by hand
// rather than by asking the node for its text. Every `Node.Xxx()` accessor in this tree panics off
// its kind, so the kind is established before either call.
func reactHookNoAnyTypeCalleeName(callee *ast.Node) string {
	if callee == nil {
		return "this hook"
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil {
			return "this hook"
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "this hook"
		}
		return "React." + name.Text()
	}
	return "this hook"
}

// reactHookNoAnyTypeBlindedRules names the rules this file's degraded types actually disable.
//
// # Why this is derived and not a string
//
// The message has to name the cost or it reads as a nit and gets suppressed. Naming the cost means
// naming rules, and a list of rule names written into a format string is wrong the week a new
// checker-based rule ships — and wrong in the worst way, because it keeps rendering a confident
// sentence about a set that has moved. That exact failure, a doc comment accurate when written and
// read as general afterwards, cost this project fourteen hours on 2026-08-23 by inoculating every
// later reader against checking. So the list is read off the live catalog at report time.
//
// # Why the predicate is `ResolvesReactValueTypes` and not `NeedsTypeChecker`
//
// This is the part that was measured rather than assumed, and the obvious answer is wrong.
//
// Forty-one registered rules declare `NeedsTypeChecker`. Only two of them go blind here. The other
// thirty-nine ask `GetSymbolAtLocation`, which resolves a BINDING, and a binding survives a
// degraded type completely. Probed on the same file, twice, on the same call:
//
//	declare function useState<T>(initial: T): any;
//	const [count, setCount] = useState(0);
//
//	GetSymbolAtLocation(useState)   symbol `useState`, 1 declaration   resolves fine
//	GetTypeAtLocation(call)         `any`, TypeFlagsAny                blind
//
// So `typescript/await-thenable` and `core/no-array-constructor` are not disabled by a fake
// `useState` and naming them would be an overclaim, while `react/set-state-in-render` and
// `react/set-state-in-effect` identify a setter by reading the type's alias and lose every finding
// in the file. A message naming all forty-one would be confidently wrong in one direction and a
// hardcoded pair confidently wrong in the other. `rule.ResolvesReactValueTypes` is the declaration
// each rule makes about its own predicate, which is the only place that knowledge honestly lives.
//
// # The one honest limitation, stated because it is invisible otherwise
//
// `rule.Registered()` returns what the LINKED BINARY registered. In `cmd/cohere` every rule package
// is imported, so this sees all of them. In this package's own test binary only `structure` rules
// are linked, so it sees zero — which would make a fixture assert a sentence the real tool never
// prints. That is not a fallback to paper over; it is a real difference between two binaries, so the
// sentence below says plainly that the set was empty rather than inventing one, and
// `TestReactHookAnyTypeNamesTheBlindedRulesFromTheCatalog` registers a stand-in rule to prove the
// populated branch renders, rather than proving the degraded branch twice.
func reactHookNoAnyTypeBlindedRules() []string {
	var blinded []string
	for _, registration := range rule.Registered() {
		if registration.Rule.ResolvesReactValueTypes {
			blinded = append(blinded, registration.Rule.Name)
		}
	}
	return blinded
}

// The rule's message and the two ways it can go, whose wording lives in
// `policy/messages/react-hook-no-any-type.json`.
var (
	reactHookNoAnyTypeText     = policy.MessageOf("structure/react-hook-no-any-type", reactHookNoAnyTypeMessageId)
	reactHookNoAnyTypeNamed    = reactHookNoAnyTypeText.Option("blindedRules", "named")
	reactHookNoAnyTypeUnlinked = reactHookNoAnyTypeText.Option("blindedRules", "unlinked")
)

// reactHookNoAnyTypeDescription builds the finding's sentence for one hook.
//
// The condition comes first and the cost second, because the cost is what makes the finding worth
// acting on. A reader who stops after the first clause has still been told what is wrong.
func reactHookNoAnyTypeDescription(hookName string) string {
	return reactHookNoAnyTypeDescriptionFor(hookName, reactHookNoAnyTypeBlindedRules())
}

// reactHookNoAnyTypeDescriptionFor is the sentence for a given set of blinded rules, apart from the
// catalog so a test can reach the branch no test binary links its way into.
func reactHookNoAnyTypeDescriptionFor(hookName string, blinded []string) string {
	if len(blinded) == 0 {
		// Not a hedge and not a default. This branch is reached only in a binary that did not link
		// the rules in question, and saying so is more useful than naming a set this binary cannot
		// see. See the derivation comment above.
		return reactHookNoAnyTypeText.Render(map[string]string{"hookName": hookName}, reactHookNoAnyTypeUnlinked)
	}
	return reactHookNoAnyTypeText.Render(
		map[string]string{"hookName": hookName, "rules": reactHookNoAnyTypeJoin(blinded)},
		reactHookNoAnyTypeNamed,
	)
}

// reactHookNoAnyTypeJoin renders a name list as prose, so the message reads as a sentence.
//
// Written here rather than with `strings.Join` because the last separator differs, and a message
// that reads as a comma-delimited machine list is the kind a reader skims past.
func reactHookNoAnyTypeJoin(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return "`" + names[0] + "`"
	case 2:
		return "`" + names[0] + "` and `" + names[1] + "`"
	}
	out := ""
	for index, name := range names {
		switch {
		case index == len(names)-1:
			out += "and `" + name + "`"
		default:
			out += "`" + name + "`, "
		}
	}
	return out
}
