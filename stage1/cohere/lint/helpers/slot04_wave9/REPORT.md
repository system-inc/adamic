Built: loadDesignSystemThrough, ingestUtilityBlock and normalizeValueFunctionNodes, one .a file each.
Commits: claim bc590dc pushed before code; branch based on current main e8ba3d5, with previous landing tip 55cd2f8.
Checks: new helper package PASS 12.116s; six uncached Node oracle probes PASS 1.674s; vet and formatting clean.
Mutants: six compiling semantic mutants caught on source Node, sanitized native and emitted JavaScript; loader-body drift and all six consumer omissions caught.
Not covered: complete engine/rule integration, full repository gate, raw invalid UTF-8, nil/alias/cyclic inputs outside the documented adapters.

## What changed

The loader preserves all early error branches, a nonnil empty-message error,
stylesheet stat order and duplicates, and table construction after stats. The
utility collector preserves independent static/functional membership, appends
functional definitions, replaces only static bodies and rejects malformed names.
Value normalization descends ordinary functions and reparses only --value and
--modifier argument lists. Earlier missing engine dependencies are explicit
callbacks, consistent with existing helper seams; they do not prevent testing
this orchestration. Their production implementations remain integration work.

All 18 remote helper branches were fetched before claiming. These three tie the
highest unclaimed concrete dependency count at six. Original comment helpers
are already owned by the base worker in HELPERS.md. The branch satisfied the
landing cap before reservation; main remains e8ba3d5 after the final refresh.

Final refresh found a later utility-ingestion reservation on slot 01:
3475c2a at 03:21:39 UTC. Our bc590dc claim is earlier, 03:21:16 UTC.
Under the established earliest-claim rule, slot 04 retains this helper. The
other two reservations have no competitors. No branch besides our own is pushed.

## Rules whose prerequisites this removes

Each of the three helpers removes one dependency from each rule below, 18
entries total. None removes the last listed blocker; these are dependency
removals, not completed or integrated rule implementations.

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

## Observations and commands

With /workspace/adamic-tools/env.sh sourced, all test output went directly to
logs. Toolchain setup passed earlier in this session: Go 0s, clang 1s, Node 1s,
submodules 2s, build cache warm 165s, total 165s; nproc 5. Setup was not repeated.

```
python3 stage1/cohere/lint/helpers/slot04_wave9/testdata/capture.py
go test ./stage1/cohere/lint/helpers/slot04_wave9 -count=1 -v -timeout=15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m
go vet ./stage1/cohere/lint/helpers/slot04_wave9/...
gofmt -l stage1/cohere/lint/helpers/slot04_wave9
git diff --check
```

The capture uses Tailwind 4.3.3 and temporary Go overlays. It captured 149 unique
asserted source fixtures across all six consumers, and 186 distinct helper
inputs from 209 calls: 137 loader states, 48 utility inputs and one empty
normalization call. Loader capture preserves observed load failures and loaded
stylesheet order; the helper oracle controls external dependency effects.
Captured utility calls are isolated inputs; ordered duplicate registration is
held separately by controls. The rule suites do not exercise meaningful value
normalization here. Eleven explicit normalization sources therefore hold nested
calls, both targets, empty args, commas in quotes, ordinary nested functions,
namespace rewrites, escaped stars, Unicode and the no-revisit rule.

The final four-way comparisons match Go over 51 control output lines and 234
captured output lines. Every successful process exits zero with no stderr;
native builds use ASan/UBSan and default Linux leak checking. The filtered
oracle has zero probe cache hits and six misses. Vet and formatting logs are
empty. Evidence is in evidence/.

Two superseded runner attempts failed explicitly: string-plus-number was not
lowerable, then an untyped empty-array fallback became never[]. Supported
interpolation and an explicitly typed fallback resolved these. Their logs are
retained as failures and are not counted as successful evidence. No shared
compiler or harness changes were made.

## Every mutant and what caught it

Each semantic mutant compiles and exits successfully on all three Adamic
execution paths before its output is rejected against independent Go.

| Helper | Mutation | Witness |
| --- | --- | --- |
| Loader | Stat a different path | Ordered stylesheet trace differs |
| Loader | Ignore an empty-message load error | Explicit nonnil empty error differs from successful table build |
| Utility | Replace membership with functional-only | Static plus functional foo loses its static bit |
| Utility | Drop functional definition append | Definition sequence loses foo |
| Normalizer | Return instead of descending other functions | Nested target within calc remains unnormalized |
| Normalizer | Discard parsed replacement edges | --value(--color) keeps its original argument |

The loader-body drift mutant changes the copied Go stat call; exact body
comparison catches it. Six separate consumer omissions remove every captured
fixture of each consumer in turn; readiness-based coverage catches each.

## Limits and inference

Observed: helper output and bounded side effects match Go on the captured and
control domains. Inference: the three frozen dependency entries per consumer
can be removed once these helpers are wired with their real external adapters.
No claim is made that a callback test completes LoadDesignSystem, NewTable,
ParseValue, compiler-program identity, recording filesystem integration, error
sentinel identities, concurrency or any whole-rule findings/fixes/suggestions.
Normalization nonempty inputs are controls, not observed consumer calls. The
full repository gate was not run. Previous helpers retain their landing evidence
because their code, compiler base and oracle inputs are unchanged.
