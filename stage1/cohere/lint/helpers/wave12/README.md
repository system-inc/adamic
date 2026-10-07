# Helpers from lint wave 1 slot 12

`collapse_normalize_value_function_argument.a` exports `normalizeValueFunctionArgument(argument: string): string`. It preserves the four Go RE2 replacements and final namespace suffix guard in `collapse/utility_nodes.go`. RE2 `\s` is ASCII space, tab, newline, carriage return and form feed; vertical tab, Unicode spaces and BOM remain. The nested-key match is lazy and cannot cross a newline through its `.*?` capture. The helper accepts decoded strings, with no AST or filesystem dependency.

Claim 61519ec4 was pushed before implementation. The branch is based on `origin/codex/lint-helpers`; it does not merge or alter the shared harness branch. Go adapters are added through build overlays, never by writing to cohere. `testdata/capture.py` runs the six consumer rule fixture files and the engine utility tests. A test-only overlay redirects each rule fixture's Tailwind package search to the pinned scratch installation; it does not change rule behavior or expected findings. Missing repository-specific source trees still cause the original tests to skip, and those skips are retained in evidence.

The captured 138 unique rule inputs cover all six consumers, and the engine/consumer runs observed 24 distinct real helper arguments. The comparison corpus combines those arguments, the complete captured sources as additional string inputs, and 9,202 controls: all byte-range characters in five contexts, bounded exhaustive short strings, Unicode and whitespace boundaries, and deterministic generated strings. Whole source strings are robustness inputs, not a claim that production passes source files as arguments. Expected helper results come from the actual unmodified Go function via an oracle-only export.

Go output is observed as comma-separated UTF-16 code units, including every unit and its order. Valid decoded text therefore matches exactly, independent of JSON escaping conventions. The same `.a` files run unchanged on source Node, Adamic's emitted JavaScript, and native under ASan/UBSan including leak checks. Three semantic mutants compile and finish on all three backends; only the external comparison catches them.

With the setup environment sourced:

```
npm install --prefix /workspace/scratch/wave12-tailwind --ignore-scripts --no-audit --no-fund tailwindcss@4.3.3 > /tmp/wave12-helper-npm.log 2>&1
python3 stage1/cohere/lint/helpers/wave12/testdata/capture.py > /tmp/wave12-helper-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=15m > /tmp/wave12-helper-tests.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-helper-vet.log 2>&1
```

Set `ADAMIC_TAILWIND_ROOT` to another scratch directory containing `node_modules/tailwindcss` when regenerating. The pinned Go cohere revision is `715ba94f3608a6500086b1076ce5cb7e51b836db`.

The helper removes one listed dependency from each of:

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

Zero rules lose their last helper blocker from this helper alone. This is a shared-helper handoff, not six completed rule ports or end-to-end Adamic findings/fix parity. Shared readiness and entry points remain untouched. Raw invalid UTF-8 is outside the decoded-string API. There is no exhaustive proof over arbitrary-length strings, and no complete repository gate claim.
