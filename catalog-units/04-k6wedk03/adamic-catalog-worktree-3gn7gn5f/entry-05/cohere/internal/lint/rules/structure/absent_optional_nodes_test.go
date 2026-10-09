package structure

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every rule in this package is run over shapes where a node it reaches for is legitimately absent.
//
// This exists because react-component-no-display-name shipped a crash that review did not catch and
// only a fixture did. `ast.SkipParentheses` dereferences its argument, so a nil check placed after
// the call rather than before it panics, and the shapes that produce a nil are ordinary: a
// `declare const` with no initializer, a `let` with no initializer, a call with no arguments, an
// object shorthand property with no explicit value.
//
// A crash is worse than a wrong finding. A wrong finding is visible and arguable; a panic takes
// down the walk for the whole file, so every other rule's verdict on that file is lost too, and on
// a large tree the file that crashed is one line in a log nobody reads.
//
// Kept as a standing guard rather than deleted after the fix, because the pattern recurs: nearly
// every rule in this family reaches for an initializer, an argument, or a type annotation that does
// not have to exist. Add new rules to the list.
//
// # What a pass here does and does not mean
//
// The shape list below is the measurement, and it is short: a handful of shapes one person thought
// of in a few minutes. A green run means every listed rule survived those shapes. It is not a claim
// that the package cannot panic, and the difference matters because the two look identical from
// here.
//
// This guard has found exactly one crash, the one it was built from. Reading "found nothing since"
// as coverage is the same error as reading a clean probe as safety: a probe that does not reach a
// line proves nothing about the line, only about the inputs. So when a new rule reaches for
// something optional in a shape not listed here, the fix is to add the shape, not to trust the
// green.
func TestNoRuleCrashesOnAbsentOptionalNodes(t *testing.T) {
	t.Parallel()

	sources := []string{
		"declare const memo: unknown;\nexport const value = 1;\n",
		"import { networkService } from './NetworkService.ts';\nlet useThingRequest;\nexport const value = useThingRequest;\n",
		"import { networkService } from './NetworkService.ts';\ndeclare const Document: unknown;\n" +
			"export function useThingRequest(o?: unknown) { return networkService.useGraphQlQuery(Document, { shorthand }, o); }\n" +
			"declare const shorthand: unknown;\n",
		"import { networkService } from './NetworkService.ts';\nexport function useThingRequest() { return networkService.useGraphQlQuery(); }\n",
		"declare const Field: { displayName: string };\nlet other;\nexport const value = other;\n",
		"export function Page() { return null; }\n",

		// A declaration with no initializer inside a component, which is what crashed
		// react-hook-require-result-naming on 42 real files. `let x;` is ordinary code and every
		// rule walking variable declarations reaches it.
		"export function Panel() {\n    let pending;\n    return pending;\n}\n",
		"export function Panel() {\n    declare const value: unknown;\n    return value;\n}\n",
		"export default function ThingPageRoute() {\n    let pending;\n    return pending;\n}\n",

		// A binding pattern with no initializer. A syntax error, and it reaches the walk while
		// someone is mid-edit, which is exactly when a crash is least welcome.
		"export function Panel() {\n    let { label };\n    return label;\n}\n",

		// A shorthand property, whose initializer is nil where a longhand one is not.
		"import { networkService } from './NetworkService.ts';\ndeclare const Document: unknown;\n" +
			"declare const identifier: string;\n" +
			"export function useThingRequest(variables: { identifier: string }) {\n" +
			"    return networkService.useGraphQlQuery(Document, { identifier });\n}\n",
	}

	rules := []rule.Rule{
		NetworkNoStringLiteralQuery,
		NetworkRequireHookRequestSuffix,
		NetworkRequireHookOptionsParameter,
		NetworkRequireHookVariablesType,
		ReactElementNoAnchor,
		ReactElementNoHorizontalRule,
		ReactComponentNoForwardRef,
		NextNoPageState,
		NextRequireApiParameterName,
		NextRequirePageDefaultExport,
		ReactComponentNoDestructuring,
		ReactComponentNoDisplayName,
		ReactComponentNoSeparateNamedExport,
		ReactComponentRequirePropertiesParameter,
		ReactComponentRequirePropertiesTypeSuffix,
		ReactHookNoDestructuring,
		ReactHookRequireEffectComment,
		ReactHookRequireResultNaming,
		ReactHookRequireResultNaming,
		ImportRequireReactNamespace,
		StorageNoDirectLocalStorage,
	}

	for _, currentRule := range rules {
		for index, source := range sources {
			t.Run(fmt.Sprintf("%s/shape-%d", currentRule.Name, index), func(t *testing.T) {
				t.Parallel()
				// A panic fails the test. The assertion is only that this returns at all, so
				// findings are deliberately not checked: what each rule concludes about these
				// shapes belongs in its own pair, and asserting it here would make this guard
				// fail for reasons that are not crashes.
				// Both extensions, because a React-gated rule declines a .ts file before reaching
				// any shape below it. Running only .ts made this guard silent for every rule with
				// an IsReactFile check, which is most of this package, and it is why the guard
				// missed a nil dereference in react-hook-require-result-naming that crashed 42
				// real files. The guard had the right shape in its list and never delivered it.
				//
				// That is the shared-guard failure appearing inside the thing built to catch it: a
				// gate upstream of the fixture silences the fixture, and from a green run the two
				// are indistinguishable.
				for _, fileName := range []string{
					"/repository/source/api/ThingRequest.ts",
					"/repository/source/components/ThingRequest.tsx",
					"/repository/app/thing/page.tsx",
				} {
					rule_testing.Run(t, currentRule, fileName, source)
				}
			})
		}
	}
}
