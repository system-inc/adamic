package react

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The two name predicates React conventions rest on, lifted from `internal/rules/structure/` where
// four rules shared them and nothing outside that package could see them.
//
// A research pass on `nextjs/no-async-client-component` asked for exactly `IsLikelyComponentName`
// and reported that no capitalization helper existed on any shelf. It did exist, in a rule package,
// which is the same invisibility this shelf was built to remove: a helper four rules already agreed
// on, unreachable from the fifth that needs it.

// IsHookName reports the shape React's own linting recognizes as a hook: "use" followed by an
// uppercase letter.
//
// A function named "used" or "user" is not a hook, which is why the letter after the prefix is
// tested rather than the prefix alone.
//
// Two divergences here, both from the same measurement as the component predicate below.
//
// Our origin is `ReactDetectionUtilities.ts:14`, `/^use[A-Z]/`, which is ASCII-only. This is unicode.
//
// And oxc's `is_react_hook_name` accepts a digit after the prefix (`utils/react.rs:761` reads
// `c.is_uppercase() || c.is_ascii_digit()`), so `use2Things` is a hook to oxc and is not one here or
// in our TypeScript. Recorded rather than adopted: our rules are the port target for this predicate,
// not oxc, and our source agrees with this implementation on that case.
func IsHookName(name string) bool {
	if !strings.HasPrefix(name, "use") || len(name) == 3 {
		return false
	}
	return unicode.IsUpper([]rune(name[3:])[0])
}

// IsHookCall reports a direct `useFoo(...)` or a namespaced `React.useFoo(...)`.
//
// The namespaced form is checked against the React object specifically, through IsNamespacedMember: a
// call to `somethingElse.useState` is not a React hook, and reading every namespaced `use*` name as one
// would count things the gate does not. Lifted from the structure package when
// consistency-no-property-alias moved to nexus (#kjnmdb1), so the two packages share one predicate.
func IsHookCall(call *ast.CallExpression) bool {
	if call == nil || call.Expression == nil {
		return false
	}

	switch call.Expression.Kind {
	case ast.KindIdentifier:
		return IsHookName(call.Expression.Text())

	case ast.KindPropertyAccessExpression:
		return IsNamespacedMember(call.Expression, IsHookName)
	}
	return false
}

// IsLikelyComponentName reports a name starting with a capital.
//
// JSX decides between a component and an intrinsic element by that letter, so the capital is the
// author's claim that this is a component rather than an inference anyone is making about it.
//
// **The test is `unicode.IsUpper` and every source it could mirror is ASCII-only. That divergence is
// stated here rather than fixed, because nothing has established which answer we want.**
//
// Our own rules are the origin of this predicate, and `ReactDetectionUtilities.ts:15` spells it
// `/^[A-Z][A-Za-z0-9]*$/`. oxc has two spellings in one file and the component one is the ASCII
// spelling: `utils/react.rs:785` is `is_ascii_uppercase` in `is_react_component_name`, which
// `no_this_in_sfc.rs` and `display_name.rs` route their whole component test through.
// `utils/react.rs:761` is `is_uppercase` and belongs to the hook path.
//
// So an earlier version of this comment claimed "oxlint uses `char::is_uppercase`" and generalized
// from the wrong function in the right file. Caught by `@system_cohere_format` while adopting this
// for the react ports.
//
// Reachable and measured: `É` and `Ω` answer true here and false under both sources, so
// `function Émile() {}` is a component to us and not to them. Nobody writes it, no corpus case
// exercises it, and that is precisely why it would never surface once shipped.
//
// Left as-is deliberately. Changing it is a one-character edit and the question is which behavior we
// want rather than what the code does, and that is a decision with no caller pressing it today.
func IsLikelyComponentName(name string) bool {
	runes := []rune(name)
	return len(runes) > 0 && unicode.IsUpper(runes[0])
}
