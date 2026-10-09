# React helper package

Ten selected `ecmascript/react` symbols from the frozen triage at `d7ab0bc4`.
Each helper has its own Adamic file. Files were extracted, not branch-merged.

| Go symbol | File | Source branch and commit |
| --- | --- | --- |
| IsEs6ComponentClass | react_es6_component_class.a | codex/lint-helpers-01, 98f7e5f |
| IsLikelyComponentName | react_likely_component_name.a | codex/lint-helpers-01, 98f7e5f |
| IsNamespacedMember | namespaced_member.a | codex/lint-helpers-02, ff5090b7 |
| IsEs5ComponentCall | es5_component_call.a | codex/lint-helpers-02, ff5090b7 |
| isCreateClassName | create_class_name.a | codex/lint-helpers-02, ff5090b7 |
| isComponentBaseName | component_base_name.a | codex/lint-helpers-03, 3e85502d |
| isComponentBase | react_component_base.a | codex/lint-helpers-05, cd24dd58 |
| isIdentifierNamed | react_identifier_named.a | codex/lint-helpers-05, cd24dd58 |
| IsHookName | react_is_hook_name.a | codex/lint-helpers-05, cd24dd58 |
| IsHookCall | react_is_hook_call.a | codex/lint-helpers-05, cd24dd58 |

The triage's overlap choices are retained: the component-base-name predicate
comes from 03 rather than 01; the identifier predicate comes from 05 rather than
01/02/03; the ES6 class predicate comes from 01 rather than 03; the ES5 call
predicate comes from 02 rather than 04. The Unicode data accompanying the likely
component-name predicate comes from 01. Hook-name data comes from 05.

`nodes.a` describes the shared, immutable Go AST projection. `projection.a`
adapts stage 1 parser identities for rule consumers. The ES5 and component-base
callbacks compose the selected leaf predicates. The namespaced-member helper
uses the selected identifier helper instead of repeating its implementation.
`main.a` composes all ten helpers for independent oracle comparison.

`react_test.go` captures actual asserted cases through Go cohere's testing hook
from every consumer in the frozen readiness ledger. It verifies that all 31
consumers supplied cases, parses those sources with the pinned Go parser, and
queries all ten helpers over every applicable projected node and name. The
namespaced-member predicate is tested with hook-name, constant-true and
constant-false callbacks; these cover both possible predicate answers, including
short-circuit rejection of non-React receivers. Unicode witnesses cover every
uppercase scalar and its adjacent scalar, including supplementary characters.

The oracle overlay only adds thin exports for the four private Go predicates.
It leaves all upstream helper bodies unchanged. Go cohere remains pinned to
7945d102a6c18dd36adf9114a758ce646e8b2359. Source Node, emitted JavaScript and
ASan/UBSan native compare byte-for-byte. Every helper has a separately compiled
semantic mutant tested against the same corpus on Node and sanitized native.

Run `go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=30m` with the
setup environment. See [REPORT.md](REPORT.md) for the completed run evidence.

The two rules whose frozen helper blockers are removed by this package alone
are `react/state-in-constructor` (ported alongside the package) and
`react/require-optimization`. The other 19 React package symbols and the
`rules/react` helper package remain outside this unit. In particular,
`react/no-children-prop` still needs `IsCreateElementCall`; it is not claimed
unblocked. No selected helper requires an unlanded package or an inexpressible
Adamic language feature.
