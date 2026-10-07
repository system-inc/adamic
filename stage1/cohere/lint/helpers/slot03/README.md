# Slot 03 lint helpers

Three separate Adamic `.a` files provide pure prerequisites. Import `isComponentBaseName` from `component_base_name.a`, `isTailwindSpace` from `tailwind_space.a`, and `listenerKinds` from `listener_kinds.a`.

- `isComponentBaseName(name)` matches exactly `Component` or `PureComponent`. No alias resolution, case folding or Unicode normalization.
- `isTailwindSpace(codePoint)` takes a decoded Go rune as a number and matches only space, tab, line feed, carriage return, vertical tab and form feed. Decode strings by code point before calling; do not pass UTF-8 bytes or UTF-16 halves. Unicode White_Space and JavaScript trim are broader than this helper.
- `listenerKinds()` returns a fresh mutable array containing `JsxAttribute`, `CallExpression`, `VariableDeclaration`, in that order. These are the existing stage-1 parser kind names. A caller must register every listed kind; the helper does not collect class literals itself.

`readiness.json` removes only this slot's three dependencies from the frozen ledger. It records 38 dependency removals across 26 distinct rules. No rule loses its final recorded helper blocker from these three alone, and no rule is marked implemented. Other slots' work is not counted here.

The fixture generator captures actual runtime source arguments, including dynamically assembled fixtures, through Go overlays. For Tailwind fixtures needing unavailable external installations, capture occurs before the skip. The external Go rule-package gate remains failed and is reported explicitly. The helper oracle itself calls real pinned Go functions, not copies. Node runs the unchanged `.a` source; native is built under ASan/UBSan and Linux leak checking. Four compiling source mutants change semantic output: the React name, whitespace, listener values and fresh-array behavior.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot03/testdata/regenerate.py > /tmp/slot03-regenerate.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > /tmp/slot03-tests.log 2>&1
```

The generator may finish successfully after recording the eight known unavailable-Tailwind/live-corpus failures; that records fixture inputs and does not turn those upstream gates into passes. An unknown failure or a missing consumer aborts regeneration. `testdata/coverage.json` and the test enforce all 26 ledger consumers and the pinned cohere commit. Repeated captures reproduce identical compressed source bytes. Tests also run a missing-consumer mutant.

## React name helper consumers

- `react/no-arrow-function-lifecycle`
- `react/no-did-mount-set-state`
- `react/no-did-update-set-state`
- `react/no-direct-mutation-state`
- `react/no-string-refs`
- `react/no-this-in-sfc`
- `react/no-typos`
- `react/no-unused-class-component-methods`
- `react/no-will-update-set-state`
- `react/prefer-stateless-function`
- `react/require-optimization`
- `react/state-in-constructor`
- `structure/consistency-require-matching-file-name`
- `structure/react-component-no-display-name`

## Both Tailwind helpers' consumers

Each of `isSpace` and `ListenerKinds` removes its own dependency for all twelve rules below:

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-concatenated-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`
