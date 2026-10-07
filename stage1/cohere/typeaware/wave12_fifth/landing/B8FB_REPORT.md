Built: rebased the owned wave 12 branch onto fetched main b8fb957aa, retaining inherited-static-field support; no rule source changes or new claims.
Commits: previous published eef97d622178f18ae109e9a9925648d54b5368ee; tested rebased tip e9f805c128319529301d42c7da25abcfc74b8c9f on codex/typeaware-wave-12.
Checks: twelve full Go byte oracles green on 1,126 controls and frozen 77 compiler/287 repository roots, normal and ASan/UBSan/LSan; rendering, metadata, checker, uncached inherited-static-field Node oracle and vet pass.
Mutants: twelve rule plus regex, five raw-question, four retention, JSX guard, Node byte, three reporting, two reporting-guard and forty-eight named-metadata mutations caught.
Uncovered: three active JSX source ports remain blocked on native JSX parsing; full registration/callback migration, Unicode quoting, nondefault options and inherited 26-rule/full repository gate not covered.

Main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 adds inherited static-field reads in internal/native/emit_objects.go and its oracle fixture/counts. The rebase completed cleanly. All shared/main changes were retained. No protected compiler file, shared parser, registry generator, harness, bridge registration or submodule pin was edited. Only codex/typeaware-wave-12 is pushed, using the exact lease on its previous published tip.

Announced harness 41eb6eab2b6de45ede0a40250765be295ee25fbd is not an ancestor of fetched main. Its changes since ab70f38d4 retire the batch-8 runner onto the registry and add the dedup ledger; no ledger rows name these fifteen owned rules. Main and fetched area/stage1-lint are both b8fb957aa. The named-kind contract remains correct: kinds are pinned ast.Kind names, and active descriptors use node: true. No numeric adapter is required. The existing JSX source probe still exits 70, empty stdout, expected GreaterThanToken, got SlashToken. Three AST claims stay ACTIVE and INCOMPLETE; the earlier three HIR/SSA/capture-dependent claims stay PARKED. No new claim was taken.

Commands source /workspace/adamic-tools/env.sh. bash cloud/setup.sh PASS: Go 0s, clang 0s, Node 0s, submodules 0s, cache 123s, total 123s. nproc 5; quota four cores; memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup cache warming overlapped initial validation, so gate durations are not benchmark times. One exec-server transport disconnect prevented an auxiliary command from starting; the same commands were retried successfully. No test failed. Output went directly to log files.

Root command: go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'. Next/third/fourth use the same options on their own packages with '^TestAgreement$'. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus, per-batch COMPILER_MANIFEST and REPOSITORY_MANIFEST select the unchanged frozen roots, and ARTIFACTS selects /workspace/wave-12/b8fb/<batch>. PASS root 189.696s, next 159.168s, third 104.061s, fourth 163.154s.

Control findings 89/138/56/596 total 879 over 187/142/94/703 controls, totaling 1,126. Full finding/fix/suggestion bytes 63,155/86,191/25,463/320,187 agree with production Go, normally and sanitized. Root compiler/repository findings remain 5/1 (8,014/18,903 bytes); other batches zero (5,010/18,485 bytes). All frozen corpus roots were actually loaded and compared. Released-handle probes require exact panic 70, and removing registry deletion changes that into success and fails the lifetime check.

PYTHONDONTWRITEBYTECODE=1 python3 stage1/cohere/typeaware/wave12_sixth/validate.py /workspace/wave-12/b8fb/sixth passes fifteen production findings and 6,093 canonical bytes across native, sanitized native, source Node and emitted JavaScript. Forty-eight named-metadata mutations are caught, using the independent pinned Go ast.Kind.String() oracle (351 names). Reporting still consumes prepared ranges/message-family selection from Go and establishes rendering only; native source verdicts/span discovery remain unimplemented. No new question or handle is introduced.

Checker: go test -v -count=1 -timeout 30m ./bridge/tsgo/checker PASS 0.174s. Uncached Node command: ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/^inherited_static_field_read\.a$'. Source Node, native, release native and emitted JavaScript agree on the newly landed fixture; the byte mutation is caught, PASS 0.544s, zero cache hits. Scoped vet passes on the checker and four owned rule packages with empty output.

Every mutation from NAMED_REPORT.md was rerun. Twelve source-rule changes skip a redeclaration survivor, change g to y, move a collection finding end by one, restrict transitive writes, reverse timeout/shebang conditions, reverse global constructor binding conditions three times, reverse comment/rest/optional-dependency conditions. The effect-name regex literal is changed to EffectX. All compile and exit 0 with empty stderr and are caught only by Go bytes. Five raw-question changes erase ancestry declaration status, signature-body presence, import target, constructor declaration status and identifier declaration status; those also fail only byte comparison. Four retention mutants fail the released-handle contract; JSX presence reversal fails valid-control success. Node's byte mutant fails its oracle. Three reporter-ID changes preserve count and success but fail Go bytes; two reporter guard bypasses change required panic 70 into success. Forty-five invalid/numeric/wrong-listener metadata changes and three node: false changes fail their metadata checks. Exact observations are in mutation-observations.json and complete logs. Reporting probes are not complete JSX source-rule mutants.

Three alternating quiet whole-process timing rounds after builds finished, each comparing full bytes. The existing benchmark_wave_12.py driver records native and Go load/run phases and total times. Every round agrees. No comparable timing is claimed for prepared JSX rendering versus Go source lint.

| Batch | Corpus | Native median seconds | Go median seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| first | compiler | 4.550 | 0.470 | 9.67 |
| first | repository | 0.559 | 0.135 | 4.14 |
| next | compiler | 2.436 | 0.390 | 6.24 |
| next | repository | 0.435 | 0.106 | 4.09 |
| third | compiler | 2.605 | 0.359 | 7.25 |
| third | repository | 0.449 | 0.127 | 3.53 |
| fourth | compiler | 4.753 | 0.447 | 10.62 |
| fourth | repository | 0.665 | 0.158 | 4.22 |

Native remains slower; this rebase makes no performance claim or causal attribution. Named metadata alone does not establish shared callback migration. The shared RegExp literal is unchanged and revalidated by the fourth oracle. Nondefault options and full repository/inherited 26-rule regressions were not run.

Complete compressed streams/logs, command exits, current source and emitter hashes, final metadata/reporting summaries, shared-ref status, prepared fixture, all mutation observations and timing rounds are under evidence/b8fb. Reproducible older ELF/archive binaries were removed from named owned scratch directories, retaining source/log evidence: prior named run 701,769,077 bytes, older completed runs 5,480,112,735 bytes; completed current-run cleanup is listed separately. This prevented disk exhaustion without deleting running artifacts or frozen inputs. The report commit follows the tested rebased tip without changing executable sources.
