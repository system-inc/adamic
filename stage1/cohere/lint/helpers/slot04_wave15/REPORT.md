# Slot 04 wave 15

Built IsPositiveInteger, roundTripsAsJavaScriptNumber and isValidSpacingMultiplier, one `.a` module each. Claim 0b4db840 was pushed before code. All 41 prior retained helpers were already complete, based on current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, green across fifteen helper packages and pushed through 58492c83. Final wildcard fetch and all claim trees on twenty origin codex/lint-helpers* branches found no competing reservation for these three. Main remains c01907a, so no rebase was needed for this batch.

## Observations

IsPositiveInteger uses one module-level JS RegExp literal for nondigits, then preserves empty-input, leading-zero and long-value handling. Lengths reaching the length checks are ASCII, so UTF-16 length and Go byte length agree there. The other two helpers reuse the verified numeric modules from slot04_wave14. Go strconv.ParseFloat is an explicit external callback dependency; JavaScript Number acceptance is not substituted for Go parser acceptance. No Go regex in a rule was ported, no hand-rolled matcher was introduced, and no findings or position conversion are involved. The inspected codex/lint-regex table has no row for the byte-loop Go integer helper.

Actual private Go functions are invoked and instrumented through temporary overlays. No persistent shared harness, registration, rule, compiler or inherited test files were edited. Capture recorded 112 distinct asserted inputs from all four consumers and 29 actual helper entry calls. Six distinct records include one IsPositiveInteger input and five isValidSpacingMultiplier inputs; no roundTripsAsJavaScriptNumber call was recorded. Precision-boundary controls exercise that helper directly. The actual Go tailwind suite passed in 0.408s.

`go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave15` passed in 14.004s. Actual Go, source Node, sanitized native and emitted JavaScript matched 638 control lines, 112 consumer-source lines and six captured-input lines. Each line contains all three predicate results: 756 lines and 2,268 boolean comparisons per mode. Controls include 512 deterministic integer strings of lengths one through thirty-two, seventy values around 2^53, empty/zero/leading-zero values, overflow, noncanonical decimal and hex spellings, signed zero, NaN/infinities, Unicode digits, line terminators, NUL, BOM and 2048-character values. Consumer sources are also supplied as helper input; this is not complete native rule execution. All successful runs require exit zero and empty stderr.

Nine semantic mutants compiled and ran successfully in all three Adamic modes, then differed from Go:

- Integer regex rejected the digit nine.
- Integer empty-input guard was removed.
- Integer long-value cutoff changed from fifteen to one hundred.
- Integer leading-zero guard checked one instead of zero.
- Round-trip parse-success guard was inverted.
- Round-trip 1e21 threshold guard was inverted.
- Round-trip exact-format comparison was omitted.
- Spacing divisor changed from 0.25 to one.
- Spacing result was negated.

Four consumer-omission mutants, one per rule below, failed the coverage check. `go vet ./stage1/cohere/lint/helpers/slot04_wave15` and `git diff --check` passed.

`ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/regexp_unicode.a$'` passed in 0.857s: three native misses, two Node misses, zero cache hits. Logs are in evidence/.

The named shared harness ab70f38d4 on origin/lint-rules/harness was fetched and inspected. It is not an ancestor of current main; its common base with main is ef3d907. These helpers do not use context.report or the finding model. No shared or automatic leak-check changes were reverted. Inherited setup passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5. Build shells source /workspace/adamic-tools/env.sh.

## Readiness inference

Each of the three helpers removes one prerequisite from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve prerequisite removals across four distinct rules. No rule reaches zero remaining blockers from these three alone. Unintegrated work from other workers is not counted as landed dependencies.

Not covered: complete rule diagnostics/fixes/suggestions, native registration into the finding model, a port of strconv.ParseFloat, arbitrary incorrect dependency callbacks, exhaustive integer strings or float bit patterns, and live reachability of the long-value helper in those existing consumer fixtures. The full repository gate was not rerun; the fifteen earlier helper packages were already green on this unchanged main base, and the new package plus a relevant uncached oracle were run.
