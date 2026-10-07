Built: require-await, symbol-description and valid-typeof in .a with numeric handed-node listeners; nine owned algorithms total.
Commits: claim 78198b914 (pushed before implementation as 89abf07e3); port 52179be0e; refinement source 52a034bbadded6fe04069f95ecdb9cc5e35a8d88; base origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30. The report commit is pushed only to codex/typeaware-wave-19.
Commands and outputs: nine-rule gate PASS 793.491s on c019; latest-compiler rebuild PASS, 50 normal/sanitized comparisons and 801774 identical canonical bytes; full bridge PASS 111.508s; current-base checker, vet and filtered Node PASS. Setup 195s, nproc 5.
Mutants: three new rule mutants plus union-mask, strict-option and ASI-initializer mutants compile and exit cleanly; only Go bytes catch them. Four raw-fact mutants, one retained-registry mutant and one numeric metadata mutant also fail their intended checks.
Uncovered: production registrations, shared .a harness integration, lint emitted-JavaScript comparison and full repository gate. React and require-atomic-updates analysis prerequisites remain skipped; no additional claims.

The claim was pushed before implementation. Main advanced twice during validation. Both rebases were conflict-free; the first retained 23 patch-equivalent commits. Main b8fb957aa fixes inherited static field reads. Every owned native runner was rebuilt with that compiler, both normally and with ASAN/UBSAN/LSAN, and compared again to unchanged production Go cohere. Only unchanged checker archives and independent Go oracle binaries were reused. No checker, bridge, parser or owned lint source changed between main c019 and b8. All nine algorithms are green again on the latest base. No main or area branch was pushed.

The new rules declare parser numeric SyntaxKind arrays in rule.json kinds and listenerKinds. Owned drivers convert legacy parser kind strings once, dispatch only relevant numeric kinds and hand each listener the fetched node. No rule compares node kinds as strings. Names are converted only for the existing C ABI. The earlier six rules retain their documented legacy driver limitation.

Symbol-description checks the ambient global through declaration-ancestry. Valid-typeof reproduces default and requireStringLiterals behavior, shadowed/global undefined and suggestions. Require-await searches its own body, stops at nested functions/classes, recognizes for-await and await-using, exempts empty bodies/generators, preserves head spans and reproduces async-removal suggestions. Adamic judges declared generic, contextual promise and heritage contracts from raw Go signatures, parameter/rest, member/index/tuple and declared/resolved-call facts. Go returns facts, never a lint verdict. ASI suggestions use exact initializer-token metadata from the owned compatibility driver.

Simple controls: 63 independently parse-valid sources, 31 findings, 15248 bytes. Strict typeof: 17253 identical bytes normally and sanitized. Require await: 28 parse-valid controls retained from 40 extracted/handwritten candidates, 19 findings, 12795 bytes; twelve extracted expected-head strings are not complete valid sources and are filtered by the independent Go parser. Three contract fixtures cover a typed promise callback, an implemented interface and self-inference through map. Four ASI controls match 2528 bytes, distinguishing modified, string-name and optional uninitialized fields from initialized fields.

Each runner matches Go over the frozen 77-root TypeScript src/compiler corpus and 287-root repository corpus on complete findings, fixes and suggestions, normally and sanitized. These three rules produce zero corpus findings, making positive controls and mutations essential. The old let.if() parser refusal remains explicitly tested and excluded from original-rule agreement. This is corpus and targeted-control coverage, not a claim to execute every upstream test-table form.

Latest-compiler byte-only mutant receipts:

| Mutation | Check that caught it | First differing byte |
| --- | --- | --- |
| Symbol changed to SymbolWrong | Go findings | 56 |
| typeof accepts invalid array string | Go findings | 2365 |
| await search ignored | Go findings | 8388 |
| Wrong TypeScript union mask | Go contract findings | 66 |
| Strict typeof option ignored | Go findings/suggestions | 3334 |
| Uninitialized fields marked initialized | Go suggestion bytes | 628 |

Each mutant compiles, exits 0 and has empty native stderr. Four Go overlays corrupt signature returns, generic type parameters, property types and heritage member types; direct checker snapshots reject each after compilation. A retained-registry mutant exits 0 with empty stderr while all four released-handle checks require panic 70. Normal and sanitized releases pass. Metadata changed from CallExpression 214 to NewExpression 215 fails the independent pinned-enum assertion. The earlier ten byte-only, five raw-fact and three retained-registry proofs were rerun on c019. The full bridge also reran seven ABI/memory/refusal mutants; filtered Node includes its one-byte check and main's inherited-static-field fixture.

Shared files remain untouched by this continuation. Production refuses wave19-type-signatures, wave19-generic-call, wave19-type-members and wave19-heritage-members with panic 70 and an unsupported-question message, verified by TestWave19ThirdProductionPending. The four-line registration.patch remains unapplied. These rules also require declaration-ancestry from the earlier three-line wave_19_next_registration.patch. Integration needs seven unique question dispatch lines total plus shared rule registration. Shared profile compilation and lint JavaScript comparison remain with codex/lint-harness-dot-a. No batch-8 Diagnostic SHA was supplied. No further non-React, non-analysis-dependent candidate remains in the audited selection snapshot; require-atomic-updates needs control_flow_graph and binding capture/escape analysis on #dnv6f2c.

Three alternating complete-process timing rounds, after all builds and gates finished:

| Runner | Corpus | Go seconds | Native seconds | Native / Go |
| --- | --- | --- | --- | --- |
| symbol-description + valid-typeof | compiler | 0.460613 | 2.949418 | 6.40x |
| symbol-description + valid-typeof | repository | 0.197996 | 0.425313 | 2.15x |
| require-await | compiler | 0.489978 | 3.029564 | 6.18x |
| require-await | repository | 0.390453 | 0.752269 | 1.93x |

The first runner measures both simple rules together. Native remains slower. These are complete-process observations including loading and parsing, not isolated dispatch/bridge measurements. Earlier same-unit setup reported Go, clang, Node and submodules ready in 0s each, cache warm 195s, total 195s; nproc 5, quota 4, memory 17.6 GB. Setup did not change across these rebases.

Commands, with test output redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/wave19-f801-scratch
# Persistent artifact variables use the paths in the preceding landing report and the new tests.
go test ./stage1/cohere/typeaware -run '^TestWave19(NumericListenerDeclarations|AgreementAndMutants|TimeoutPendingRegistration|ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles|ThirdSimpleRules|AwaitRule|ThirdNumericMetadata|ThirdReleasedHandles|ContractByteMutant)$' -v -count=1 -timeout=30m
ADAMIC_TSGO_CORPUS=/workspace/wave19-typescript go test ./bridge/tsgo -count=1 -timeout=15m -v
python3 stage1/cohere/typeaware/wave_19_third/rebuild_landing.py $TMPDIR /workspace/wave19-typescript
python3 stage1/cohere/typeaware/wave_19_third/prove_native_mutants.py $TMPDIR
python3 stage1/cohere/typeaware/wave_19_third/prove_contract_facts.py $TMPDIR/third-raw-mutants
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py $TMPDIR/fact-mutants
go test ./stage1/cohere/typeaware -run '^TestWave19(ThirdStrictTypeof|AwaitInitializerASI|ThirdProductionPending|ThirdNumericMetadata)$' -count=1 -v
go test ./bridge/tsgo/checker -count=1 -v
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|call_targets_.*|devirtualize|inherited_static_field_read)\.a$' -count=1 -timeout=15m -v
```

The rebuild and native-mutation scripts require persistent artifacts from the full owned tests. Local evidence at /workspace/wave19-f801-scratch/final-evidence/streams.tar.gz contains 881 output, manifest, config, snapshot, mutation and timing files; sha256.json contains their digests. Paths record run provenance. The archive is kept locally because automatic approval review rejected exporting its large logs/config payload without explicit authorization. Code, claims and this written report are the pushed payload. The full repository gate and baseline 26-rule corpus suite were not rerun.
