# Slot 04 wave 21 report

Built classTokensIn, classTokensOf and utilityEvaluation.resolveBareArgument, one helper per `.a` file. All 59 prior helpers were complete and pushed before claim 3f7ed758. This wave brings the retained total to 62. Claim was pushed before code; the final handoff names the implementation/evidence commit.

Claim and initial validation used main 39638d9e278d38bb5aeae887f46d55a70e47aaad and lint area d65a8f931c98655936ae04c6899f38f14862b73e. Main and the lint area advanced during the final checks; landing rebase evidence is recorded separately below. The requested area rebase and incoming shared finding-model/allocator changes are retained. No main or area ref is pushed. Wildcard fetch checked every claim on all twenty origin helper branches, pinned in evidence/claims.json. The three selected helpers were unclaimed and tied the highest available fan-out, four consumers each; the previously delivered comment bundle remains owned.

## Behavior and rules helped

Tokenization uses bytes, preserves exact text and source offsets, groups separator/non-separator runs, and retains nil versus allocated-empty success state. Source mismatch, negative start, end outside source and reversed ranges refuse classTokensIn. Non-ASCII bytes are not decoded into a Unicode whitespace test. Byte-to-UTF-16 conversion belongs to a future finding position builder and is not performed here.

Bare argument resolution preserves the four-type allowlist, ratio Fraction selection, empty/inference refusal, two positive-integer ratio components, space-padded slash printing, canonical quarter-number acceptance and integer-percentage guard. Callback dependencies remain explicit. The generic parser result is projected to canonical CSS text for this unit's comparisons.

Each token helper removes prerequisites for enforce-consistent-class-order, no-deprecated-classes, no-duplicate-classes and no-unnecessary-whitespace. The bare helper removes prerequisites for enforce-consistent-class-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. All names below have the better-tailwindcss/ prefix.

| Consumer | New dependencies removed | Other frozen dependencies remain |
|---|---:|---:|
| enforce-consistent-class-order | 3 | 261 |
| enforce-shorthand-classes | 1 | 200 |
| no-conflicting-classes | 1 | 220 |
| no-deprecated-classes | 2 | 20 |
| no-duplicate-classes | 2 | 21 |
| no-unknown-classes | 1 | 203 |
| no-unnecessary-whitespace | 2 | 20 |

Twelve rule/helper edges removed across seven distinct rules. The table subtracts only this wave from the frozen shared inventory, not prior waves or other workers' delivered work. No rule becomes completely helper-ready from these three alone. Readiness is conditional on byte source access, Go isSpace, inference/segmentation/numeric/trim/parser dependencies. No rule implementation, listener, finding, fix or suggestion was added.

## Observations

Two private Go exports expose actual unexported methods in tailwind and tailwind/collapse. Go supplies external callback observations; actual claimed methods provide the expected output. The Adamic implementations run on source Node, emitted JavaScript and ASan/UBSan native. Every run must exit zero with empty stderr, including mutants. Raw invalid UTF-8 token bytes are printed as decimal byte sequences so output comparison never replaces malformed byte strings.

Capture retained 195 unique asserted fixtures from all seven consumer rules. The selected actual Go tailwind suite passed in 1.166s. It made 259 token helper entry calls, yielding 211 distinct inputs: 110 classTokensIn and 101 classTokensOf. No resolveBareArgument call was reached in that live suite; direct actual-Go controls supply its coverage. Consumer source texts are additional byte-token controls, not native findings parity. Live token inputs retain source text, decoded value and exact Go ranges.

The deterministic corpus has 1,538 controls: 768 raw-byte controls covering all 256 byte values at negative/zero/positive offsets, 42 range/text controls, and 728 bare argument/value/fraction combinations. Text controls include Unicode, Unicode spaces, all six ASCII whitespace bytes and escaped text. Bare controls include unsupported types, empty values/fractions, ratios with decimals/leading zeros/extra parts, canonical and noncanonical numbers, percentages, precision-boundary integers and nonnumeric values.

Parity matches 7,232 control output lines, 5,879 consumer-source token lines and 1,586 live-call lines, 14,697 physical output lines per mode. Token outputs compare every token's bytes, separator flag and byte range; bare outputs compare success, ratio and canonical replacement text.

## Commands and outputs

Every test wrote directly to its evidence log. Environment Go 1.27.1, clang 20.1.8, Node 24.19.0, source /workspace/adamic-tools/env.sh after setup. Timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (94s)
setup: done in 94s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. Setup warms package/test binaries with no tests selected and is not a full correctness gate.

- python3 stage1/cohere/lint/helpers/slot04_wave21/testdata/capture.py: 195 consumer fixtures, 211 distinct live token inputs, actual Go tailwind suite PASS 1.166s.
- python3 stage1/cohere/lint/helpers/slot04_wave21/testdata/regenerate.py: 1,538 controls, identical SHA256 after regeneration.
- Initial go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave21: baseline parity PASS 20.04s; package FAIL 83.338s because the proposed inference-guard inversion narrowed inferred to an empty string and TypeScript rejected three later comparisons (TS2367). This is not credited as a semantic detection. Twelve other compiling semantic mutants and all seven consumer omissions passed. The rejected mutant was replaced with an inference-guard omission, retaining all strict comparator/run checks. helpers.log preserves this trial.
- Final go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave21: PASS 73.536s: baseline parity 7.99s, all thirteen compiling semantic mutants 65.50s and seven consumer omissions. The initial failed trial remains preserved.
- go vet ./stage1/cohere/lint/helpers/slot04_wave21: exit zero, empty vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS 1.040s; Node/emitted JavaScript/release native/sanitized native match 758 bytes, three native misses, two Node misses, zero cache hits.
- A final fetch detected main c7991b90 and lint area b84a9d93. An ancestry check failed as intended, requiring a landing rebase; final-fetch.log records the movement.

## Every final mutant

All thirteen final semantic mutants compiled, ran cleanly and differed from actual Go output in source Node, sanitized native and emitted JavaScript. Variants are temporary copies; production sources are unchanged.

| File | Mutation | Independent check |
|---|---|---|
| classTokensOf | omit start offset | token start positions differ |
| classTokensOf | subtract one from end | token end positions differ |
| classTokensOf | invert emitted separator flag | all-byte/text flags differ |
| classTokensOf | emit one-byte runs | grouped token bytes/count/ranges differ |
| classTokensIn | reject end equal to source length | valid exact-boundary ranges refuse |
| classTokensIn | reject start equal to end | valid empty ranges lose allocated success |
| classTokensIn | invert source/value equality | exact source and mismatch controls disagree |
| resolveBareArgument | exclude integer | real integer resolutions refuse |
| resolveBareArgument | omit failed-inference guard | uninferrable known-type values are accepted |
| resolveBareArgument | require three ratio parts | valid two-part ratios refuse |
| resolveBareArgument | clear ratio result flag | successful ratios lose Go's ratio flag |
| resolveBareArgument | invert spacing acceptance | canonical/noncanonical number verdicts disagree |
| resolveBareArgument | invert integer-percentage acceptance | integer/decimal percentage verdicts disagree |

Seven separate structural mutants omit each consumer's fixtures in turn. The readiness-derived coverage check catches all seven missing consumers; these are coverage checks, not compiling semantic mutants.

## Limits and required checks

This bounded gate covers the touched helper package and a filtered uncached external Node oracle. The full repository gate and 17 broader stage 1 TypeScript/postcss/graphql/parser correctness checks were not selected. No selected check skipped, no skip was accepted as correctness evidence, and no harness/assertion/required-input check was relaxed or removed. The invalid initial mutation was replaced because a compiler rejection cannot prove output parity can fail.

Prior retained helpers and shared harness/runtime checks were green on the claim base; the changed landing base requires fresh checks. Callback implementations, complete parser AST fields, all possible rule configurations, actual bare-helper reachability in this selected rule corpus, whole native findings/ranges/fixes/suggestions and arbitrary machine-integer offsets outside exact JavaScript integer range are not covered. The byte interface avoids undocumented Unicode reinterpretation and requires callers to supply byte arrays rather than UTF-16 strings.

## Landing rebase and fresh checks

All 69 worker commits rebased without conflict onto origin/area/stage1-lint b84a9d93, containing current main c7991b90. Rebased implementation 696ce65bda727da57f9acda77a5a5cbc0dcc3e1f. All 22 helper packages passed afresh, including this wave at 71.916s and every retained mutant/omission. Shared .a/emitted-JavaScript/suggestion harness PASS 195.273s, uncached Node oracle PASS 15.279s with zero hits, runtime checks PASS 10.810s, all-helper vet clean. No selected check skipped or was relaxed. See ../slot04_landing_b84a9d93/REPORT.md for complete logs and incoming-code preservation. The final handoff names the pushed evidence tip.
