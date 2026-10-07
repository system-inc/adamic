# Helpers from lint wave 1 slot 12

`collapse_normalize_value_function_argument.a` exports `normalizeValueFunctionArgument(argument: string): string`. It uses the shared table's four JS RegExp literal translations, preserving Go replacement order and the final namespace suffix guard in `collapse/utility_nodes.go`. RE2 `\s` is ASCII space, tab, newline, carriage return and form feed; vertical tab, Unicode spaces and BOM remain. The nested-key match is lazy and cannot cross a newline through its `.*?` capture. The helper accepts decoded strings, with no AST or filesystem dependency.

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

## Segment scanner

`collapse_segment.a` exports `segment(input: string, separator: number): string[]`. Supported separators are integer ASCII bytes 0 through 127, which cover every observed consumer/engine call. Go's byte API can also split inside a multibyte UTF-8 character with separators 128 through 255, yielding invalid byte fragments. Adamic's decoded strings cannot preserve those fragments, so those separators refuse explicitly rather than silently substituting a UTF-16 interpretation. Negative and fractional inputs also refuse.

The scanner preserves the original precedence: a top-level separator wins before escape, quote or bracket handling. Backslashes skip the next byte, quotes ignore nesting, only a matching top closer pops, and the final part is always returned, even empty. A UTF-16 scan agrees for supported ASCII separators because remaining bytes/code units of a skipped non-ASCII character cannot be ASCII syntax.

Claim 8aeb70fd was pushed before implementation. Regenerate by setting `ADAMIC_CAPTURE_HELPER=segment` when running testdata/capture.py. The same six consumers contribute 138 captured fixture inputs, with 248 distinct real scanner input/separator pairs observed across the consumer and utility tests. The final corpus has 30,857 rows: actual pairs, each consumer source under space/comma as additional robustness inputs, and 30,333 controls. All 128 ASCII separators are exercised, including quote, bracket, backslash and NUL separators, with Unicode, escapes, empty parts, mismatched/unclosed nesting, bounded exhaustive and deterministic generated strings.

Actual Go, source Node, ASan/UBSan native and emitted JavaScript match all 855,507 observation bytes. Three semantic mutants compile and are caught by comparison on every backend: dropping final empties (row 1), popping unmatched closers (row 4259), and disabling escapes (row 913). Four unsupported separator cases refuse identically on all backends with empty stdout, exit 70 and the exact panic message. Disabling the separator guard also compiles, finishes, and is caught only by the refusal comparison; that is a supported-domain check, not a claim that Go rejects its wider byte API.

The complete owned helper package now passes in 32.477s; vet is clean. Both helpers remove two listed dependencies from each of the six named rules. None loses its final blocker from these two helpers alone. No rule port, production adapter or shared registry has been changed.

## Counter support blocker

The third claim, `nextBuildCount`, was pushed in b3041334 after fetching all 420 origin refs and inspecting all 17 unique claim blobs. It ties the highest remaining consumer count at six. The proposed helper must preserve a process-wide, mutex-protected Go `int` counter, not a per-file counter or a floating-point approximation. This machine's Go `int` is 64 bits.

No production counter helper is delivered. `gaps/atomic-counter.a` is an exact signed-64-bit support probe using a shared BigInt64 cell and atomic increment. Node agrees byte for byte with the actual Go helper for starts 0, 9007199254740991 and 9223372036854775806, three increments each. The actual Go function is called through an oracle-only adapter that seeds and restores its real process-scoped state; a separate arithmetic imitation is not the expected result. Results include 9007199254740993 and wrapping 9223372036854775807 to -9223372036854775808.

The observed compiler failure is exactly `stage 0 can't lower a function returning bigint yet` at gaps/atomic-counter.a:4:10. Loading succeeds; lowering refuses before either native or JavaScript backend emission. This proves the selected exact primitive path is unavailable on the requested base. It is not a claim that all possible alternate integer representations have been disproved. Replacing the public integer representation or adding compiler/runtime primitives would require an explicit integration decision outside this helper's ownership. Shared compiler and harness files remain untouched.

`gaps/number-counter.a` is a compiling approximation mutant, not an alternative helper. It finishes on all three backends; the actual Go comparison alone catches the lost increment at observation 5. The gap test will fail when the exact primitive path becomes available, requiring the helper to be implemented rather than retaining a stale blocked claim. These nine serial observations do not prove concurrent worker sharing; neither native counter execution nor emitted-JavaScript counter parity is claimed. No new rule or final helper blocker is unblocked by this probe.

The counter claim remains explicitly blocked. Under the later parking continuation, three additional CFG helpers are now delivered: normalizeBigIntLiteral, isBreakableStatement and newBlock, each in a separate .a file. Their API, four-consumer handoff, actual-Go comparisons, mutants and limits are recorded in [CONTROL-FLOW.md](CONTROL-FLOW.md). The complete owned package passes in 52.416s. Other unclaimed helpers remain; the exact counter primitive is still undelivered and the unit stops at that documented boundary without changing shared compiler or harness files. No helper-inventory exhaustion is claimed.

Additional supplied-node CFG helpers are now delivered: [constant truthiness](TRUTHINESS.md) and [nested label collection](LABELS.md). Both cover all four consumer fixture AST families on actual Go, source Node, emitted JavaScript and sanitized native. Each removes one more dependency for each consumer, zero final blockers.

The effective-now regex instruction is applied to the earlier normalizer: its hand-written equivalents are replaced by literal translations from codex/lint-regex. Actual Go/source Node/emitted JS/sanitized native still match 9,364 cases, with all three compiling semantic mutants caught. See [REGEX-MIGRATION.md](REGEX-MIGRATION.md) for provenance and final scoped verification.
