# Slot 04 wave 16

Built modifierAsValue, resolveArbitraryArgument and Theme.variableReference, one `.a` file per helper. Claim b769e910 was pushed before code. All 44 prior retained helpers were complete and pushed through 22f6ea52, already green on current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Initial and final wildcard fetches inspected all twenty origin codex/lint-helpers* branches and every claim tree. The maximum available named fan-out was four; these three tie it. Final inspection found their reservation only on this worker's branch. Main did not advance, so no rebase was necessary.

## Observations

The modifier helper preserves nil presence, maps arbitrary modifiers to arbitrary values and every other modifier kind to named, and leaves Fraction and DataType empty. The arbitrary resolver preserves wildcard, existing typehint and inference priority, returning parser nodes unchanged on success. The theme helper preserves missing entry, prefix-before-escape composition, reference option bit 2 and the nonempty fallback guard. These contain no regex matcher or finding-position logic. No rules, shared generators, harness files, compiler files or incoming allocator changes were edited.

ParseValue, InferDataType, PrefixKey and escapeCSSIdentifier are explicit separately owned callback dependencies. The private Go oracle adapts these dependencies with real Go functions; it invokes the actual claimed functions for expected results. Parsed output is projected through ValueToCss in the comparison driver. This is a bounded resolver/output observation, not a claim that those callback dependencies or every parsed-node field were ported here. Go nil slices become empty node lists with an explicit failure flag.

The actual Go Tailwind consumer suite passed in 0.453s with temporary overlays. It yielded 112 distinct asserted fixtures across all four consumers and three distinct variableReference calls. No modifierAsValue or resolveArbitraryArgument call was reached in that selected live suite. Both are exercised directly by the Go control corpus. Consumer source texts are also supplied as helper inputs, not executed as complete Adamic rules.

`go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave16` passed in 44.263s. Actual Go, source Node, ASan/UBSan native and emitted JavaScript matched 43,200 control records, 112 consumer-source records and three captured-call records. Those 43,315 composite records emitted 48,093 physical lines, including newlines inside values. The control cross-product covers empty/ordinary/Unicode/newline values, nil/present modifiers, named/arbitrary/unknown kinds, wildcard/known/unknown/empty types, matching/mismatching/no typehint, absent/present theme entries, options 0 through 3, escaped and Unicode keys, and empty/nonempty prefixes. Every successful command must exit zero with empty stderr.

Ten semantic mutants compiled and ran successfully in each Adamic mode, then disagreed with actual Go on a deterministic seed-1604 subset of 600 controls:

- Invert the nil-modifier guard.
- Test named rather than arbitrary modifier kind.
- Erase the modifier value.
- Disable wildcard acceptance.
- Invert the typehint equality guard.
- Invert inference success.
- Invert missing theme entry.
- Use option bit 1 instead of reference bit 2.
- Invert the nonempty fallback guard.
- Omit CSS identifier escaping.

An earlier full-corpus run also passed in 227.040s, including eight of these semantic mutants against all 43,200 controls. The final suite adds the two presence-guard mutants and bounds mutant repetitions to the independently observed 600-control subset.

Four consumer-omission mutants, one per rule below, failed the independent coverage check. The compressed control corpus and mutant subset reproduce with identical SHA-256 bytes via `python3 stage1/cohere/lint/helpers/slot04_wave16/testdata/generate.py`. Logs are in evidence/.

`go vet ./stage1/cohere/lint/helpers/slot04_wave16` passed with an empty log. `git diff --check` passed. `ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/regexp_unicode.a$'` passed in 0.950s, with three native misses, two Node misses and zero cache hits.

The shared finding model ab70f38d4 is not yet in main; these helpers have no finding-model dependency. Inherited setup succeeded: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc printed 5 again. Build shells source /workspace/adamic-tools/env.sh. The full repository gate was not run. The prior sixteen helper packages were already green on this unchanged main base; this new package and a filtered uncached external oracle were run.

## Readiness inference

Each helper removes one prerequisite from each of these four rules:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve prerequisite removals across four distinct rules. None reaches zero remaining blockers from these three alone. Unintegrated implementations on other worker branches are not counted as landed dependencies. All 47 retained helpers are now complete; this batch leaves no outstanding reservation.

Not covered: complete rule diagnostics/fixes/suggestions or native registration, live reachability of modifierAsValue and resolveArbitraryArgument in the selected consumer fixtures, a new implementation of any supplied callback, every parsed-node field or arbitrary callback side effects, invalid UTF-8, malformed unbracketed arguments outside the private caller contract, arbitrary theme profiles and exhaustive input strings. See README.md for the handoff contracts.
