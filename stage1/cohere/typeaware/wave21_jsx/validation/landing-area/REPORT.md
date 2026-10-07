# Wave 21 integration-area landing

The owned branch was rebased onto area/stage1-lint b46914832, then onto
its d3a37422c update containing current main b6b1538b0. Only owned prerequisite
checks and documentation changed. Integration's runtime, JSX parser, registry,
.a support and finding model were accepted without shared-file edits.

## Closed dependency and remaining blockers

The native parser now accepts the four JSX witnesses in normal and sanitized
builds. Owned prerequisites require a nonempty parsed tree; removing parsing
produces an empty answer and is caught. This prerequisite evidence does not
count as findings agreement for the prepared React cores.

The two previously blocked no-object-constructor JSX cases now compare complete
findings, fixes and suggestions against production Go, normal and ASAN:
671 identical bytes for each. The keyword-label control still refuses at byte 37:

```typescript
var yield = 5;

                yield: while (foo) {
                    if (bar)
                        break yield
                    new Object();
                }
```

Production Go reports useLiteral. Native parsing reports
`parser slice expected semicolon at 37`. The shared parser is outside this unit.
The full upstream literal in no_object_constructor_test.go remains the reproducer;
no finding/fix/suggestion parity is claimed for this control.

The four prepared JSX/style cores still need production source/checker adaptation.
RuleContext exposes the filename through parser.path, but no project configuration,
checker program lifetime or binding/type API. It therefore cannot currently feed
ordered declarations, opaque symbols, type flags or resolved foreign signatures
into these cores. Changing that shared context, registration generator or harness
is prohibited for this unit. Their existing raw Go provider is explicitly test
preparation, not native parsing or production bridge integration. No lint verdict
comes from that provider. The three HIR/SSA/capture React reservations remain parked.

No new reservation was taken. These source-integration gaps are retained claims,
not releases. The prior all-origin audit found no available ranking entry; this
landing unit stops at its shared dependencies rather than claiming more.

## Required correctness inputs

The pinned installer from origin/area/developer-tools was exported into isolated
/tmp scratch without modifying shared repository files. npm locks/integrities,
TypeScript 6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8, deterministic
100 MiB gitignore and this checkout's C checker archive were installed.
The exact 17-row driver runs uncached Go tests with JSON events and never counts
an absent check as a pass. The first area run encountered root disk exhaustion;
Go's regenerable build cache occupied 23 GB. Automatic approval review rejected
bulk scratch/checkout deletion because the exact destructive scope was not
explicitly authorized. No rejected deletion was performed. Supported
`go clean -cache` restored 22 GB of space; all affected checks were rerun.

The repeat has 16 PASS, one ABSENT, zero SKIP and no semantic failures. The missing
check is TestSplitTSGoAgrees in internal/native/units_tsgo_test.go, present on
origin/area/developer-tools but absent from both main and the requested area base.
Reproducer: `go test ./internal/native -run '^TestSplitTSGoAgrees$' -count=1 -v`
reports no tests to run. It is outside this unit's protected-file territory.
Consequently the required 17-check gate and full repository gate are not claimed
as green. The real TypeScript compiler expression and whole-tree comparisons,
registered lint corpus, postcss/graphql-js/parsers, JSON and CSS printers all run
with their inputs. No test was skipped, relaxed or removed to gain a pass.

## Validation record

Initial b469 results: combined seven-suite command exited 1 because it was compiled
before the obsolete JSX-refusal assertion was replaced. Its other six suites pass;
the corrected core-only rerun passes 233.221 s. The initial failure is retained.
JSX package PASS 1780.496 s, style package PASS 290.132 s. External source/emitted
Node passes all 22 default batches. These are initial-base evidence, not substitutes
for the final d3 rebuilds. Final-base results, mutants and timings are appended
below after their logged runs complete.

Setup reports Go ready 0 s; clang, Node and submodules 1 s; build cache warm 152 s;
total 152 s, nproc 5 and CPU quota 4. All test output went directly to log files.

## Final d3 results

All seven TestWave21 suites PASS together **1047.698 s**, including all ten
source ports and three prepared HIR cores. JSX package PASS **1187.016 s**:
216 Go-valid controls in three modes, 66 batches, both 77/287-root frozen
corpora, normal/ASAN byte comparisons, three rule mutants and real lifetime
contracts. Style package PASS **166.135 s**: 82 controls in three profiles,
complete corpora, original/emitted Node, sanitizers, rule mutant and lifetimes.

Checker PASS **0.329 s**; uncached one-byte and lastIndexOf Node checks PASS
**7.476 s**. The first Typeof filter used the wrong capitalization and selected
none of those tests. The corrected TypeOf filter plus all typeof fixtures PASS
**3.142 s**, including main's four mutants, 31 fresh native and 22 fresh Node
observations. Vet exits 0 with empty output. No full repository gate is claimed.

All seventeen owned rule mutants compile, exit 0 and write empty stderr. Only
independent complete finding/fix/suggestion bytes catch them. The full ledger
also records listener, adjacency, released registry, strict-question and
prerequisite mutants, with their distinct checks. Temporary-path lengths can
shift a first differing byte across runs; these offsets describe the final run.

| Rule | Mutation | Catch byte |
| --- | --- | ---: |
| no-mixed-enums | Reverse enum comparison | 53 |
| no-collection-misuse | Reverse collection judgment | 1465 |
| no-discarded-outcome | Reverse discarded outcome judgment | 7879 |
| no-discarded-pure-result | Reverse pure result judgment | 18380 |
| no-uncleared-race-timeout | Reverse timer judgment | 91 |
| no-process-exit-after-output | Reverse process-exit judgment | 6248 |
| require-blocking-standard-streams | Reverse blocking judgment | 15624 |
| no-obj-calls | Reverse direct-global test | 468 |
| no-object-constructor | Lose preceding semicolon in suggestion | 45885 |
| no-promise-executor-return | Reverse expression exemption | 77186 |
| set-state-in-render | Reverse unconditional render gate | 480 |
| set-state-in-effect | Reverse setter gate | 1577 |
| static-components | Reverse creator gate | 60 |
| jsx-fragments | Reverse fragment-name acceptance | 1720 |
| jsx-no-undef | Reverse resolved-symbol acceptance | 49 |
| jsx-no-constructed-context-values | Reverse component gate | 2336 |
| style-prop-object | Remove string from literal set | 490 |

Released handles produce exact panic 70 in the real C bridge; retained-registry
mutants exit 0 and are caught by that expectation. HIR/JSX prepared cores remain
partial source ports. The three HIR claims remain parked, four JSX/style claims
remain retained with the missing project/checker context named above.

## Native time against Go

These are observed executions with other tests/builds active, not isolated
benchmarks. Native source suites include program loading/parsing/bridge questions.
Prepared cores receive raw Go preparation outside Adamic, so their native/Go
columns use different pipelines and establish no end-to-end native speed claim.
Compilation and sanitizer runs are excluded from these normal timing rows.

| Workload | Native seconds | Go seconds | Native input |
| --- | ---: | ---: | --- |
| Three core rules, compiler | 3.656259 | 0.443489 | Source |
| Three core rules, repository | 0.807941 | 0.400963 | Source |
| Three Nexus result rules, compiler | 3.389198 | 1.165407 | Source |
| Three Nexus result rules, repository | 0.409719 | 0.236784 | Source |
| Three stream/timer rules, compiler | 2.952251 | 0.337745 | Source |
| Three stream/timer rules, repository | 0.418314 | 0.144792 | Source |
| Three HIR cores, 91 controls | 0.017986 | 0.324493 | Prepared HIR |
| Style core, three control modes | 0.031007 | 0.112394 | Raw AST graph |
| Style core, compiler | 15.725068 | 0.363046 | Raw AST graph |
| Style core, repository | 0.973362 | 0.284355 | Raw AST graph |

HIR raw Go preparation is 0.297221 s; style control preparation is 0.037457 s.
The JSX execution sum and preparation time are recorded in jsx.log.

## Required final-base check verdicts

| Check | Verdict | Seconds |
| --- | --- | ---: |
| lint | PASS | 147.975 |
| parser-expressions | PASS | 51.120 |
| parser-whole | PASS | 34.116 |
| postcss | PASS | 17.391 |
| graphql | PASS | 11.445 |
| media-query | PASS | 8.777 |
| selector | PASS | 19.184 |
| selector-nontermination | PASS | 2.593 |
| values | PASS | 10.296 |
| json-numeric | PASS | 8.509 |
| json-corpus | PASS | 67.214 |
| json-mutants | PASS | 4.937 |
| css-printer-default | PASS | 123.552 |
| css-printer-narrow | PASS | 124.727 |
| css-printer-boundaries | PASS | 3.768 |
| gitignore-100MiB | PASS | 21.443 |
| checker-split | ABSENT | 8.305 |

The final-base check summary remains **16 PASS / 1 ABSENT / 0 SKIP**. This is
not a green 17-check gate: the absent split-checker test remains an integration
blocker, not a permitted skip. No correctness check was relaxed or removed.

## Commands

Every shell sources /workspace/adamic-tools/env.sh. Owned commands set
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json,
ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest and
ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest.
The JSX command additionally sets ADAMIC_WAVE21_JSX_ARTIFACTS=/workspace/wave21-jsx-final;
style sets ADAMIC_WAVE21_STYLE_ARTIFACTS=/tmp/wave21-style-artifacts. TMPDIR=/tmp.

```sh
go test -count=1 -v -timeout 30m ./stage1/cohere/typeaware -run '^TestWave21' > /tmp/wave21-d3-owned.log 2>&1
go test -count=1 -v -timeout 30m ./stage1/cohere/typeaware/wave21_jsx > /tmp/wave21-d3-jsx.log 2>&1
go test -count=1 -v -timeout 30m ./stage1/cohere/typeaware/wave21_jsx/style-prop-object > /tmp/wave21-d3-style.log 2>&1
source /tmp/wave21-gate-inputs.env
python3 /tmp/wave21-gate-installer/cloud/run-gate-input-tests.py /workspace/adamic /tmp/wave21-d3-required-inputs with > /tmp/wave21-d3-required-inputs.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTypeOf|^TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_' -count=1 -v -timeout 10m > /tmp/wave21-d3-typeof-node.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave21-d3-checker.log 2>&1
go vet ./... > /tmp/wave21-d3-vet-final.log 2>&1
```

Uncovered: production source/checker adaptation and registry factories for four
prepared JSX/style cores, source-to-HIR/SSA/capture/memo compilation-unit analyses,
the keyword-label syntax control, suppression/repair application, native source
throughput for prepared cores, the absent seventeenth correctness test and the
full repository gate. The supported source/core comparisons above are green;
these explicit shared dependencies prevent calling the partial cores completed
native source ports. No new rules are claimed in this landing unit.

Final external Node repeat: all 22 default batches agree byte for byte on source
and freshly emitted JavaScript. Fresh-artifact JSX source prerequisites PASS
13.496 s, normal/ASAN, including the successful parser-omission mutant. Final vet
exits 0 with empty output. Final base is area d3a37422c / main b6b1538b0; the owned
prerequisite-update commit rebased to 69e1b8fe6. No subsequent source change was
made to any rule or bridge question. Only evidence/documentation follows.
