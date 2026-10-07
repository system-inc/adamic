Built: rebased eighteen existing native ports onto the current lint integration branch; rule/checker sources unchanged; no new claims.
Commits: compiler/rebased tree 01255f6a4; integration b84a9d931; main c7991b900; published ac44e3e3b preserved.
Commands: owned eighteen-rule oracles, bridge, filtered Node/JavaScript, registry and vet pass; older gate 1002.893s; setup 296s, nproc 5.
Mutants: eighteen rule mutations, scope export, three regression guards, Unicode case, released registries, seven bridge and Node one-byte detected.
Uncovered: full required-input repository gate, configured dynamic RegExp, whole-rule JavaScript/shared checker context, three parked React analysis rules.

The rebase replayed only this branch's first-parent changes onto b84a9d931,
which includes main c7991b900. It completed without conflicts. The upstream
changes introduce compiler proof handling and native record support; the compiler
was rebuilt rather than reusing the previous green binary. Its build metadata
records 01255f6a411ce8ec4af3879d0a1dc119008aef10 with vcs.modified=false.
Owned checker and rule sources have no difference from published ac44e3e3b.
No shared registration, harness, parser, lowerer or native source was edited.

The twelve older ports passed in 1002.893s, including their controls, every
rule mutant, scope-export mutation, released-handle refusals and retained-handle
mutation, and normal/ASan/UBSan runs over all 287 frozen repository and 77 pinned
compiler roots. The sixth and seventh batches passed their complete finding,
fix and suggestion streams, options, mutants and both normal/sanitized corpora.
The accepted control counts remain 176 and 272; the independent Go parser
rejects 188 and 251 extracted nonsource/malformed inputs respectively. Their
default streams are 52,437 and 73,462 identical bytes. Every corpus comparison
for these six rules remains 18,485 repository bytes and 5,857 compiler bytes;
positive controls and comparison-only mutants prove this silence can fail.

| Mutation | First differing byte | Detector |
| --- | ---: | --- |
| numeric-prefix | 14792 | Go byte oracle |
| has-own-fix-span | 18785 | Go byte oracle |
| spread-parens | 314 | Go byte oracle |
| promise-condition | 16849 | Go byte oracle |
| spread-await-edit | 22057 | Go byte oracle |
| lost-write-span | 65053 | Go byte oracle |
| nullish-suggestion | 361 | Go byte oracle |
| qualifier-fix | 22460 | Go byte oracle |
| private-read | 4152 | Go byte oracle |
| scope-export-mutant | 29967 | Go byte oracle |
| global-provenance | 5940 | Go byte oracle |
| global-declaration-span | 85544 | Go byte oracle |
| timer-string | 63292 | Go byte oracle |
| fragment-id | 735 | Go byte oracle |
| undef-id | 358 | Go byte oracle |
| adjacent-judgment | 32497 | Go byte oracle |
| default-id | 772 | Go byte oracle |
| nested-id | 18310 | Go byte oracle |
| sort-id | 53364 | Go byte oracle |
| computed-trivia | 66117 | Go byte oracle |
| jsx-attribute-flag | 18244 | Go byte oracle |
| class-fallback-flag | 933 | Go byte oracle |

Every tabled mutant compiled and ran with exit 0 and empty stderr. The Unicode
question's wrong simple-case mutation fails TestUnicodeNodeText on ßName.
Both owned question probes panic 70 with exactly invalid-or-released-checker-
handle after program release, under sanitizers. The bridge gate passed in
259.08s with 1,600 positions and 54,982 identical bytes under ASan/UBSan/LSan.
Its input/output length mutations trigger ASan; retained handles fail the stale
assertion; wrong position differs at byte 6; removed link opt-in fails refusal;
missing C free and region cleanup trigger LSan. The filtered external Node and
emitted-JavaScript oracle passed in 69.288s for regex, inherited static fields,
and all five newly landed proof fixtures; its one-byte mutant was caught.
Registry regeneration and vet passed. No selected correctness test skipped.

| Batch / corpus | Native seconds | Go seconds |
| --- | ---: | ---: |
| Sixth / repository | 1.481 | 0.376 |
| Sixth / compiler | 8.271 | 0.578 |
| Seventh / repository | 9.034 | 2.084 |
| Seventh / compiler | 39.629 | 0.812 |

These are single process observations while other landing checks ran, not an
isolated benchmark. Native remains slower. All exact commands and timings,
compiled source hashes, canonical streams and fresh logs are in
[landing-main-c7991b900](landing-main-c7991b900/summary.json).

The commands, with output redirected to named /workspace/wave-22-new-*.log files:

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
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(regexp\.a|inherited_static_field_read\.a|proven_.*\.a)$|^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m
go run ./cmd/lint-registry
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
```

The full repository gate, including the complete required-input correctness
suite for other stage 1 packages, was not run. The owned corpus checks supplied
the pinned TypeScript input and bridge corpus; no skip was accepted or weakened.
The new dynamic-pattern probe still fails at options.a:12:46 with stage 0 can't
lower RegExp with a nonconstant pattern yet. Its required new RegExp(source, 'u')
implementation is preserved. Whole-rule adamic js still refuses the native
checker call, and shared RuleContext still lacks a checker-program binding.
The three prior HIR/SSA/capture-dependent React claims remain parked under
#dnv6f2c. These limitations are not advertised as completed integration.

At setup, the workspace had less than 1 GB free. Only 94 individually identified
obsolete ELF executables and C archives from this unit's prior scratch runs were
removed, reclaiming 3,853,617,278 bytes. Sources, logs and committed evidence were
retained; the removed-file inventory is included with the fresh evidence.

The refreshed snapshot contains 632 origin refs, 33 distinct Markdown claim
blobs and all 197 ranked rules. Zero rules remain both unported and unclaimed.
The full current inventory is committed alongside this report. No new claim was
made. The branch is pushed only to codex/typeaware-wave-22, never main or area.
