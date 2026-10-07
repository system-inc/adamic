Rebased fifteen implemented wave-11 rule ports onto current main e8ba3d5d; no new rules claimed.
Previous pushed tip 569ca634 is replaced by the rebased code tip 322e84ac and a separate evidence commit.
All five final-base byte-oracle suites, bridge checks, expanded compiler oracle and vet passed.
Every existing per-rule mutant was caught again; raw-question and released-handle mutants passed their checks.
Uncovered: three blocked React ports, full repository gate, rule JavaScript comparison and option-matrix limits.

## Landing scope

The only branch this worker has pushed is codex/typeaware-wave-11. It was not
on main. Its 21 commits rebased cleanly onto e011f8f6 and all 16 type-aware
package tests passed across three disjoint filters, with full bridge checks,
a filtered compiler oracle, vet and formatting clean. The package runs took
757.292s, 887.015s and 334.816s respectively. Those results are retained.

Main advanced during validation to e8ba3d5d, landing compiler devirtualization.
The branch was rebased again and all five worker batches rerun independently
in parallel, together with bridge tests and an expanded compiler oracle.
Both range-diffs preserve all 21 patches. No production source was edited for
the landing work; the protected compiler files are unchanged relative to main.
The final-base inherited baseline suites were not rerun; their full previous-base
results are retained separately and are not presented as final-base validation.

The final remote check confirmed main still at e8ba3d5d and the wave-11 remote
still at 569ca634. Publishing the explicitly requested rebase uses an exact
force-with-lease for that previous remote SHA, refusing concurrent updates.
No PR was opened and no merge to main was performed.

## Final-base native agreement

Each batch reran its independent unchanged production Go oracle, controls,
the frozen 287 repository roots and 77 compiler roots, complete serialized
findings/fixes/suggestions, normal and sanitized execution, existing mutants
and released-handle checks. Reports for each implementation retain their
production-default and other coverage limits. Native sources remain .a.

| Batch | Controls compared | Findings | Repairs | Identical control bytes | Suite seconds |
| --- | ---: | ---: | ---: | ---: | ---: |
| first | 25 | 53 | 31 | 22500 | 216.887 |
| next | 28 | 67 | 0 | 29186 | 383.800 |
| third | 211 | 191 | 0 | 120564 | 212.017 |
| fourth | 98 | 71 | 0 | 31373 | 198.124 |
| fifth | 374 | 259 | 216 | 97678 | 246.249 |

All normal and sanitizer comparisons passed. These production-default rules
emit no suggestions; the empty suggestion fields were compared. The fifth
batch still excludes its two explicitly logged strict-parser-rejected octal
escape inputs. Artifact byte lengths differ from historical control runs
because generated input paths differ; Go and native bytes within each run
are compared without normalization.

## Mutants caught again

Each diagnostic/raw-question mutant compiled, exited zero and emitted empty
stderr; only the independent Go byte comparator caught the altered answer.
The released-registry mutants are separate lifetime checks: exit zero fails
the required real stale-handle panic 70. The full bridge suite also reproves
C-output and region ownership mutants through sanitizer failures.

| Batch | Mutant and observation |
| --- | --- |
| first | regexp mutant: exit 0, empty stderr, independent Go bytes catch byte 464 |
| first | array mutant: exit 0, empty stderr, independent Go bytes catch byte 13831 |
| first | branches mutant: exit 0, empty stderr, independent Go bytes catch byte 16556 |
| first | signature mutant: exit 0, empty stderr, independent Go bytes catch byte 20567 |
| first | index mutant: exit 0, empty stderr, independent Go bytes catch byte 13828 |
| first | syntax mutant: exit 0, empty stderr, independent Go bytes catch byte 1323 |
| first | released-registry mutant exits 0, caught by required panic 70 |
| next | collection mutant: exit 0, empty stderr, independent Go bytes catch byte 66 |
| next | outcome mutant: exit 0, empty stderr, independent Go bytes catch byte 24473 |
| next | pure mutant: exit 0, empty stderr, independent Go bytes catch byte 14665 |
| next | lineage mutant: exit 0, empty stderr, independent Go bytes catch byte 13293 |
| next | awaited mutant: exit 0, empty stderr, independent Go bytes catch byte 25373 |
| next | released-registry mutant exits 0, caught by required panic 70 |
| third | eval mutant: exit 0, empty stderr, independent Go bytes catch byte 63 |
| third | extend mutant: exit 0, empty stderr, independent Go bytes catch byte 19690 |
| third | assign mutant: exit 0, empty stderr, independent Go bytes catch byte 27959 |
| third | global mutant: exit 0, empty stderr, independent Go bytes catch byte 1753 |
| third | anchor mutant: exit 0, empty stderr, independent Go bytes catch byte 35349 |
| third | released-registry mutant exits 0, caught by required panic 70 |
| fourth | func mutant: exit 0, empty stderr, independent Go bytes catch byte 65 |
| fourth | nonconstructor mutant: exit 0, empty stderr, independent Go bytes catch byte 6529 |
| fourth | wrappers mutant: exit 0, empty stderr, independent Go bytes catch byte 7684 |
| fourth | global mutant: exit 0, empty stderr, independent Go bytes catch byte 62 |
| fourth | released-registry mutant exits 0, caught by required panic 70 |
| fifth | throw mutant: exit 0, empty stderr, independent Go bytes catch byte 732 |
| fifth | backreference mutant: exit 0, empty stderr, independent Go bytes catch byte 126 |
| fifth | arrow-fix mutant: exit 0, empty stderr, independent Go bytes catch byte 10735 |
| fifth | provenance mutant: exit 0, empty stderr, independent Go bytes catch byte 122 |
| fifth | regex-path mutant: exit 0, empty stderr, independent Go bytes catch byte 4593 |
| fifth | self-resolution mutant: exit 0, empty stderr, independent Go bytes catch byte 10451 |
| fifth | symbol provenance question mutant: exit 0, empty stderr, independent Go bytes catch byte 122 |
| fifth | released-registry mutant exits 0, caught by required panic 70 |

## Native time against Go

These are instrumented whole-process observations during parallel validation,
including startup, program load, rule work, output and teardown. They are not
quiet medians or a devirtualization speed comparison. Native remains slower
on all measured populations; historical quiet measurements remain in the
individual batch reports.

| Batch | Repository Go / native | Compiler Go / native |
| --- | --- | --- |
| first | 342.865501ms / 713.141057ms | 612.919248ms / 2.802933158s |
| next | 309.555998ms / 851.829996ms | 817.266485ms / 38.987817552s |
| third | 471.346953ms / 665.308942ms | 815.621739ms / 4.463841578s |
| fourth | 358.27826ms / 875.830327ms | 1.178421758s / 4.140600692s |
| fifth | 326.773022ms / 526.908047ms | 390.974302ms / 2.44087648s |

## Commands and additional checks

Toolchain environment: source /workspace/adamic-tools/env.sh. Setup was reused
from the earlier successful cloud/setup.sh run: ready 0s, cache warm 86s,
total 86s; nproc 5. Every test wrote output to a log, without test pipelines.

On the final base, run each of these five filters separately with
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript and its corresponding
ADAMIC_WAVE_11[_NEXT|_THIRD|_FOURTH|_FIFTH]_ARTIFACTS directory:

```
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11AgreementAndMutants$'
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11NextAgreementAndMutants$'
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11ThirdAgreementAndMutants$'
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11FourthAgreementAndMutants$'
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11FifthAgreementAndMutants$'
go test -v -count=1 -timeout 10m ./bridge/tsgo/checker ./bridge/tsgo
go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures|devirtualize|call_targets_(closure|element|region|reuse|sort))\.a$'
go vet ./...
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
git diff --check
```

The final bridge tests passed: checker 0.481s, bridge 166.082s. The expanded
Node/native/emitted-JavaScript compiler oracle passed in 3.241s, exercising
13 fixtures including the newly landed call-target/devirtualization cases,
and the one-byte mutant. This is compiler-oracle coverage, not an emitted-JavaScript
execution of the lint rules. Vet, gofmt and whitespace logs are empty.

Full logs, exact diagnostic streams, current 364 corpus-root hashes, manifests,
the old/new range-diffs and structured results are in validation-wave-11-landing.
Build products remain in /workspace/wave-11-landing-current.

## Existing unfinished claims

react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components remain claimed and unported. The native React
HIR/SSA/capture/memoization/control-dominance substrate is absent. On the final
rebased compiler the shared parser again exits 70 on the minimal self-closing
JSX probe, expecting GreaterThanToken at the slash. Adding JSX alone would
not resolve the hook-only effect/render dependency. WAVE_11_SIXTH_REPORT.md
retains the prerequisite map and the instruction to stop at other blockers.
No shared parser, generator or harness was edited, and no additional rules
were reserved. These three have no native rule agreement, mutants, sanitizer
rule runs or lint timings. The full repository gate, rule JavaScript comparison
and the earlier per-rule option-matrix gaps remain uncovered.
