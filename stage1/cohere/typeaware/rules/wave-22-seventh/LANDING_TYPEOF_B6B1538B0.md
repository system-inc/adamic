Built: rebased eighteen existing ports onto lint integration d3a37422c, including main b6b1538b0; no new eligible rule remains.
Commits: tested compiler/source revision da8eae8bd; published d3bbc1d8e ancestry preserved for a normal wave-22 push.
Commands: owned Go byte oracles, options, sanitizer corpora, released handles, bridge, filtered Node/runtime mutants, registry and vet pass; older gate 920.991s; setup 256s, nproc 5.
Mutants: eighteen clean rule mutants, scope export, three regression guards, Unicode question, bridge lifetime/ownership guards and typeof/one-byte Node mutants caught.
Uncovered: full required-input repository gate; dynamic native RegExp; whole-rule emitted JavaScript/shared checker binding; three parked React analysis rules.

The rebase onto origin/area/stage1-lint completed cleanly. Upstream typeof null,
constructor and lookup-presence changes were accepted without editing compiler,
runtime, registry, generator or shared harness files. Owned rule implementation
sources are unchanged. The rebuilt compiler records da8eae8bd372d1b8543dcd02a95147234c812f35
and vcs.modified=false.

All eighteen owned ports match production Go cohere byte for byte on findings,
fixes and suggestions, including supported options. Normal and ASan/UBSan runs
agree over the frozen 287 repository and 77 pinned TypeScript compiler roots.
Selected checks supplied the pinned inputs and none skipped. Positive controls
and mutants hold corpus silence. Recent default control streams match at 52,437
and 73,462 bytes; older control findings total 425, 346, 104 and 183.

| Mutation | First differing byte from Go |
| --- | ---: |
| fragment-id | 735 |
| undef-id | 358 |
| adjacent-judgment | 32497 |
| default-id | 772 |
| nested-id | 18310 |
| sort-id | 53364 |
| computed-trivia | 66117 |
| jsx-attribute-flag | 18244 |
| class-fallback-flag | 933 |
| numeric-prefix | 14733 |
| has-own-fix-span | 18706 |
| spread-parens | 313 |
| promise-condition | 16736 |
| spread-await-edit | 21926 |
| lost-write-span | 64704 |
| nullish-suggestion | 361 |
| qualifier-fix | 22460 |
| private-read | 4152 |
| scope-export-mutant | 29967 |
| global-provenance | 5919 |
| global-declaration-span | 85043 |
| timer-string | 62995 |

Every listed mutation builds and runs with exit 0 and empty stderr; the byte
oracle alone catches the disagreement. The Unicode checker mutation instead
fails its independent TestUnicodeNodeText at scalar 1, field 1 for ßName.
Both symbol-declaration-syntax and unicode-node-text refuse a released checker
program under sanitizers with exit 70 and the exact invalid-or-released-handle
panic. Required retained-handle refusals remain covered by the owned and generic
bridge suites. Bridge input/output bounds, stale retention, incorrect position,
unlinked calls, C output ownership and region ownership mutants were caught.
The bridge compares 1,600 positions in four pinned compiler files, 54,982 bytes,
under ASan/UBSan/LSan; package duration 292.431s.

The filtered Node oracle passed in 102.806s with source/native/emitted-JavaScript
comparisons for typeof, regexp, inherited static field and proven fixtures.
The one-byte mutation was caught. Additional upstream constructor, string, null
and slot-presence mutations passed their expected Node-only detection checks in
2.839s. Registry descriptor/discovery mutation checks passed in 0.629s; vet passed.

| Batch / corpus | Native seconds | Go seconds |
| --- | ---: | ---: |
| sixth / repository | 0.572 | 0.302 |
| sixth / compiler | 4.863 | 0.595 |
| seventh / repository | 6.986 | 0.329 |
| seventh / compiler | 36.947 | 0.578 |

These observations ran alongside other landing checks, not as an isolated
benchmark. Native remains slower. Logs, exact commands, source hashes, streams,
statuses, timings and the 653-ref claim inventory are preserved in
[landing-typeof-b6b1538b0](landing-typeof-b6b1538b0/summary.json).

Commands ran with output redirected to named wave-22-typeof files:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -v -timeout=30m
go build -o /workspace/wave-22-seventh-adamic ./cmd/adamic
go build -buildmode=c-archive -o /workspace/wave-22-seventh-checker.a ./bridge/tsgo/archive
python3 stage1/cohere/typeaware/rules/wave-22-sixth/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/prove_flags.py
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(typeof.*\.a|regexp\.a|inherited_static_field_read\.a|proven_.*\.a)$|^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m
go test ./internal/oracle -run '^TestTypeOf(ConstructorMutant|StringLiteralMutant|NullMutant|NullSlotPresenceMutant)$' -count=1 -v -timeout=10m
go run ./cmd/lint-registry
go test ./stage1/cohere/lint/registry -count=1 -v
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
```

The full required-input repository gate was not run. Fresh blocker probes retain
exact reproducers: compiling no-unstable-nested-components/gaps/dynamic-pattern.a
fails at options.a:12:46 because stage 0 cannot lower nonconstant RegExp; emitting
JavaScript for the owned driver refuses the native checker call at
common/unicode_node_text.a:5:24. Shared RuleContext still exposes no checker-program
binding. No integration or unsupported option coverage is claimed. Three React
HIR/SSA/capture claims remain parked under #dnv6f2c.

All 197 ranked rules are accounted for by existing ports or claims across origin:
33 distinct Markdown claim blobs and 653 origin refs. No new reservation was
made. Only codex/typeaware-wave-22 is pushed; integration merges.
