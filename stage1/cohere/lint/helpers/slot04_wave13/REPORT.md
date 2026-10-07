# Slot 04 wave 13

Built `checkGroupQuantifier`, `errNothingToRepeat`, and `checkGroupConstruct`, one `.a` module each. Prior 35 retained helpers were complete, tested and pushed through d715558, based on current origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. Main remained unchanged at the final fetch, so no rebase was needed. The final scan explicitly fetched and inspected all 20 origin codex/lint-helpers* branches and every claims document; no competing reservation for the final three was found. The delivered comment bundle remains owned under HELPERS.md.

## Claims and ownership

Original claim f2d65497 was pushed before code, reserving checkGroupQuantifier, errNothingToRepeat and wordBoundary. The final refresh revealed slot 01's wordBoundary claim 74dfcf68 at 04:48:20 UTC preceded ours at 04:49:03 UTC by forty-three seconds. That reservation was yielded. Replacement claim a86ca357 reserved checkGroupConstruct and was pushed before its code. No duplicate wordBoundary implementation is delivered.

## Observations

The actual private helpers in pinned Go cohere are invoked using temporary overlays. No persistent shared harness, registration generator, compiler, or rule files were edited. The adapters supply actual Go quantifier byte widths. The port does not approximate that separately owned parser. The group-kind and Unicode checks preserve Annex B's non-Unicode lookahead exception. Error output preserves the quantifier prefix alone and Go's wrapped ErrUnsupportedSyntax rendering. The new group construct check accepts the exact permitted prefixes, and renders one UTF-8 rune after `(?` for unsupported groups. Raw invalid UTF-8 bytes retain Go's one-byte invalid-rune precision. APIs encode nil/success as an empty byte array and every failure as nonempty ErrUnsupportedSyntax bytes; error wrappers must retain that category.

Final `python3 testdata/capture.py` recorded 389 distinct asserted source fixtures from all four consumers, and 97 distinct input records from 12,443 instrumented helper-entry calls across the rule suites. The Go suites passed: next 0.173s, typescript 5.454s, core 6.362s. Captured records combine helper entry points and are not attributed separately to individual rules. Consumer fixture text is also supplied as raw helper input; that comparison is not a complete rule execution. Explicit controls exercise quantifiers, malformed bounds, valid/unsupported groups, Unicode and invalid/truncated/overlong UTF-8, surrogate encodings and out-of-range encodings. Every input is expanded across three legal group kinds, sixteen Go option flag combinations and both negated values; flags unused by these helpers are inert.

`go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave13 -run '^Test(AssertionsGoNodeNativeJavaScript|ConsumerCoverage)$'` passed in 52.380s. Actual Go, source Node, sanitized native and emitted JavaScript matched 13,536 control observations, 112,032 consumer-source observations and 27,936 captured-input observations: 153,504 total. Success requires exit zero and empty stderr. Four consumer-omission mutants, one per listed rule, failed the coverage check. `go vet ./stage1/cohere/lint/helpers/slot04_wave13` and `git diff --check` passed with empty logs.

Final `go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave13 -run '^TestCompilingMutants$'` passed in 29.449s. Every mutant compiled and ran successfully in source Node, sanitized native and emitted JavaScript, then differed from actual Go:

- errNothingToRepeat omitted the quantifier prefix.
- errNothingToRepeat changed assertion wording to atom wording.
- checkGroupQuantifier reversed the Unicode lookahead exception.
- checkGroupQuantifier removed the non-quantifier success guard.
- checkGroupConstruct rejected the permitted colon opener.
- checkGroupConstruct truncated a four-byte rune to one byte.

Logs are in evidence/. Earlier comparisons of the subsequently yielded boundary implementation are not counted as final evidence or delivered as code.

Inherited setup passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5. Builds source /workspace/adamic-tools/env.sh. The toolchain was not reinstalled for this continuation.

## Readiness inference

Each helper removes one prerequisite from each of:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

This is twelve prerequisite removals across four distinct rules. The frozen ledger lists 54, 45, 46 and 47 remaining dependencies respectively before this batch, so no rule reaches zero blockers from these three alone. Other workers' unintegrated implementations are not treated as landed dependencies.

Not covered: complete Adamic rule diagnostics/fixes/suggestions, regexp rewriting/compilation/matching, a port of quantifierWidth, arbitrary incorrect dependency callbacks, all possible long raw byte strings, or Go error object allocation/stack identity. No new rule or listener was added. The full repository gate was not rerun; earlier helper packages were already green against this unchanged main base.
