Rebased all 65 retained slot-04 helpers onto the current lint integration branch; no new helper claimed.
Implementation tip: 7a943b17, rebased from f6598c04 onto b46914832; main c7991b90 is an ancestor.
All 23 helper packages, shared harness, helper vet, uncached Node oracle and native runtime checks PASS.
Every retained mutant/coverage check passed; emitted-output and second-suggestion-edit mutants were caught.
Full repository gate and the 17 broader input-dependent comparisons were not selected; helper and parser limits remain.

# Landing unit

The work-in-progress cap makes landing readiness this unit. The previously completed claims are implemented, tested and pushed through wave22. No new reservation or helper source change is made while refreshing their landing evidence. Only codex/lint-helpers-04 is pushed; main and area/ branches are untouched. No pull request is opened.

## Rebase and incoming behavior

Refreshed every origin codex/lint-helpers* branch plus main and area/stage1-lint. Rebased 72 commits without conflicts onto origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53. The rebased implementation tip is 7a943b1778210ff26d595523f2c2a481f146b99e. Main c7991b900362796aefd111474e65eb5398e91953 is retained as an ancestor. The wave22 claim rebased from 0e4e3f70 to 00a6965e; its ordering before code is preserved. The old published tip used for the exact push lease is f6598c04674fc66d6486eb49ca1e5abb7b5ff69b. The final response names the evidence commit and pushed tip.

Incoming legacy-rule registration, the nil-options-adapter guard, RuleContext.has, the shared finding model and allocator checks are retained. There is no diff from the new area base in the protected compiler files, lint.ts, context.ts or the shared oracle. No options adapter needed a change: the integrated comparison passed with the guard enabled. The generated registry reports 40 descriptors; it remains untracked.

## Setup and exact commands

Every command wrote directly to its own raw log; no test process was piped. Source /workspace/adamic-tools/env.sh was used for every build shell. bash cloud/setup.sh succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. Timing lines: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 88s, done 88s. nproc printed 5. Raw setup.log records the selected tools and machine summary.

- go run ./cmd/lint-registry: PASS; 40 descriptors, registry.log.
- go test -p=2 -count=1 -v -timeout=20m ./stage1/cohere/lint/helpers/...: PASS; all 23 packages, helpers.log. This includes the four inherited foundation helpers and inherited comments package in addition to all 65 retained slot-04 helpers. Per-package times are below and in verification.json. No selected test skipped or failed.
- go test -count=1 -v -timeout=20m ./stage1/cohere/lint -run '^(TestRulesAgree|TestDotARename|TestCompleteSuggestionSerialization|TestEmittedJavaScriptMismatch)$': PASS, 337.429s, harness.log. Emitted mismatch check 43.76s; .a rename check 79.27s; complete suggestion serialization 85.83s; integrated rule comparison 128.51s. The indented child FAIL in the mismatch probe is the expected mutant rejection; the parent test passes.
- go vet ./stage1/cohere/lint/helpers/...: PASS, empty vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS, 2.937s, oracle.log. Node, emitted JavaScript, release native and sanitized native agree on 758 bytes. Native misses 3, Node misses 2, cache hits 0.
- go test -count=1 -v -timeout=10m ./internal/native -run '^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$': PASS, 0.826s, runtime.log.

| Helper package | Seconds |
| --- | ---: |
| foundation and initial slot04 | 148.517 |
| comments | 215.485 |
| slot04_wave10 | 20.732 |
| slot04_wave11 | 17.499 |
| slot04_wave12 | 17.467 |
| slot04_wave13 | 108.204 |
| slot04_wave14 | 22.596 |
| slot04_wave15 | 32.037 |
| slot04_wave16 | 56.363 |
| slot04_wave17 | 33.685 |
| slot04_wave18 | 48.029 |
| slot04_wave19 | 164.433 |
| slot04_wave2 | 17.124 |
| slot04_wave20 | 32.438 |
| slot04_wave21 | 67.530 |
| slot04_wave22 | 46.299 |
| slot04_wave3_space | 20.389 |
| slot04_wave4 | 22.091 |
| slot04_wave5 | 25.416 |
| slot04_wave6 | 33.607 |
| slot04_wave7 | 28.208 |
| slot04_wave8 | 31.834 |
| slot04_wave9 | 15.454 |

## Every retained mutant

MUTANTS.md extracts the named subtests and catches from this fresh run, with raw-log line references. There are 170 named mutant subtests, additional standalone guard/drift mutants, and 78 logged omission catches plus the initial consumer-removal checks. No retained test source, mutation or assertion was relaxed. The tests keep their original Go/Node/native/emitted-JavaScript coverage and limits. A compile failure or unintended panic is not credited as a compiling semantic mutant.

The shared harness independently rejects a clean-running emitted-JavaScript extra-output mutant through its ordinary comparison. It also rejects the changed second suggestion-edit endpoint on source Node, emitted JavaScript and sanitized native. The baseline .a-to-.ts rename preserves module bytes and comparisons.

## Limits

This is a landing rerun, not new helper or rule functionality. Conditional dependency removals, callback contracts, bounded corpora, invalid-byte/configuration limits and AST projections are unchanged; see the existing wave reports, especially slot04_wave22/REPORT.md. The integrated rule test records Go's recovered findings for malformed upstream inputs and proves the ports explicitly refuse unsupported parser recovery. Those cases are not successful recovered-findings parity.

The full repository gate, compiler throughput/profile tests and the 17 broader TypeScript/postcss/graphql/parser comparisons were not selected or claimed. Their required inputs were not bypassed; no check was skipped, relaxed or deleted to obtain these selected passes. No source rule, registration generator, shared harness, compiler or allocator implementation was edited. Final remote main and area bases stayed at c7991b90 and b46914832 during verification.
