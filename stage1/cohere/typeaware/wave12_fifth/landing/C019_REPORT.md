Built: rebased all owned work onto fetched main c01907a70 and replaced exhaustive-deps effect-name matching with the shared JS RegExp literal; no new claims.
Commits: previous published tip 5308e63bed47097629877b233b32a1b2a42d3c3b; tested code commit 0f278dd99091e7fa3b9277292dba673daa090647 on codex/typeaware-wave-12.
Checks: twelve rules match Go on 1,126 controls and frozen 77 compiler/287 repository roots, normal and ASan/UBSan/LSan; partial JSX rendering, regex boundaries, checker, Node and vet pass.
Mutants: twelve rule plus one regex, five raw-question, four retention, JSX refusal, Node byte, twelve listener metadata, three rendering, two rendering guards and three JSX metadata mutations caught.
Uncovered: three active JSX source ports, numeric supplied-node runtime dispatch, custom rule options, inherited 26-rule/full repository gate; no new claims while source ports are incomplete.

## Landing and shared dependencies

Only codex/typeaware-wave-12 belongs to this unit and is pushed. All owned commits rebased cleanly from f8013f0b onto origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. The fetched main advance changes Stage 3 material and internal/oracle/stage3_hook_test.go; the shared JSX parser and numeric callback adapter are still absent on this base. Main changes were retained. No protected compiler, shared parser, shared harness, finding model, registration generator or submodule pin was edited. Publication uses an exact lease against the previous own-branch published tip, as required by the explicit rebase instruction. No main or area branch is pushed.

The supplied harness SHA is recorded and was inspected: ab70f38d47de1d4974082b38f84a56af2368b7af on origin/lint-rules/harness. It is not an ancestor of fetched origin/main. Fetched origin/area/stage1-lint is also c01907a70. Its separate harness context still exposes string kinds and index-based lookup; reportNode/reportRange and unchanged wire compatibility do not provide the required numeric supplied-node source adapter. No shared files were copied or changed to work around this dependency.

Three HIR-dependent claims remain PARKED with their previously named source HIR/SSA/capture/memo blockers. The three AST claims react/jsx-fragments, react/jsx-no-constructed-context-values and react/jsx-no-undef remain ACTIVE and INCOMPLETE. The production-positive JSX control still receives exact native parser panic 70: expected GreaterThanToken, got SlashToken. Shared parser and numeric source event integration block source verdict/span discovery. Existing native reporting functions and numeric rule.json declarations are preserved and revalidated. No additional claim was taken.

## Regex change

Shared table commit 071fb012848ce0408428c61aba0857cca472236f, row react/exhaustive_deps.go:451, translates Go Effect($|[^a-z]) to the literal /Effect((?![\s\S])|[^a-z])/gu. effect_hook_name.a now uses that literal; exhaustive_deps.a calls it instead of a fixed name list. The source table row and its option-family limitation are recorded in wave12_fourth/regex_translation.json. Every default supported hook is classified by the literal. AdditionalHooks and other nondefault options are not implemented or claimed; the table's required constructor for that option is new RegExp(pattern, 'u').

Independent Go regexp, native, ASan/UBSan/LSan, source Node and emitted JavaScript agree on twenty effect-name boundary cases, including LF, CR, Unicode letters, astral text and U+2028: forty output bytes with empty runtime stderr. Full-source exhaustive-deps findings remain unchanged against production Go. Replacing Effect with EffectX in the literal compiles, exits 0 with empty stderr and is caught only by full Go byte comparison at byte 8,339. Prefer-regex-literals' syntax/rewrite parser is unchanged; the upstream syntax walker is not a Go regexp compile site.

## Commands and observations

All toolchain commands source /workspace/adamic-tools/env.sh. bash cloud/setup.sh: Go 0s, clang 0s, Node 0s, submodules 0s, cache 51s, total 51s. nproc 5, CPU quota four cores, memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Output was written directly to log files.

The root gate runs go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'. Next, third and fourth run go test -v -count=1 -timeout 30m on their own packages with TestAgreement. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus and each batch's COMPILER_MANIFEST, REPOSITORY_MANIFEST and ARTIFACTS variables select the unchanged frozen roots and /workspace/wave-12/c019 artifacts. Root PASS 124.361s, next PASS 150.429s, third PASS 101.204s, fourth PASS 181.302s. These gate durations overlap other builds and are not speed measurements.

Control findings are 89/138/56/596, totaling 879 across 187/142/94/703 controls. Canonical control bytes are 63,155/86,191/25,463/320,187. Every finding, fix and suggestion field matches. Root compiler/repository findings remain 5/1 with 8,014/18,903 bytes; the other batches have zero findings with 5,010/18,485 bytes. Normal and sanitized output agree. Frozen roots are the same manifests used before the rebase, not an expanded self-generated corpus.

Checker command: go test -v -count=1 -timeout 30m ./bridge/tsgo/checker, PASS 0.204s. Vet on that package and the four owned rule packages passes with empty output. External TypeScript enum auditor check_listeners.py validates twelve numeric listener declarations and catches all twelve +1 mutations. The current main runtime still lacks numeric supplied-node dispatch, so metadata does not imply that the existing twelve callbacks have completed that migration.

Uncached Node command: ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/^(closures|method_closures|generic_functions|regions)\.a$'. Four selected fixtures agree across source Node, native, release native and emitted JavaScript; the one-byte oracle mutant is caught. PASS 1.840s, zero cache hits. An initial checker package typo failed setup, and two earlier Node filters selected only the byte-mutant test; their logs are retained and are not counted as fixture coverage. The corrected named fixture run is the evidence above.

Partial JSX command: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus python3 stage1/cohere/typeaware/wave12_sixth/validate.py /workspace/wave-12/c019/sixth. PASS fifteen findings (1 fragment, 1 undefined name, 13 constructed-context), 6,093 canonical bytes, normal/sanitized native and source/emitted Node. The shorter artifact paths explain the byte-count difference from the historical report. Go supplies locations/message-family selection to prepared reporting inputs: this proves rendering, not source verdicts or native span discovery. ASCII-only quoting, default options and prepared construction inputs remain the documented limitations. No new bridge question or handle is used by this reporting slice.

## Every mutation

All twelve original source-rule mutations compile and exit 0 with empty stderr, and only their Go byte comparison catches them. First: skip one redeclaration survivor, change the global-regex flag test from g to y, and increase the write-only collection finding's end offset by one. Next: direct-write versus transitive-write, lost timeout polarity and shebang reason polarity. Third: global binding polarity for Function, nonconstructors and wrappers. Fourth: comment exception, rest-parameter symbol declarations and optional dependency-path polarity. The additional effect-name literal mutation is described above. Exact first differing bytes are preserved in each complete gate log.

Five raw-question mutations also exit 0 with empty stderr and differ only in bytes: ancestry declaration-file bit, call signature body bit, import target path, constructor symbol declaration-file bit and identifier binding declaration-file bit. Four registry-retention mutations change expected released-handle panic 70 into success and are caught by separate lifetime checks. Reversing raw JSX-presence detection is caught by valid-control success; the positive JSX parser refusal is asserted separately. These refusal/lifetime mutations are not represented as byte-only mutations.

Node's one-byte mutation is caught by the external oracle. Twelve listener numeric +1 mutations fail the external enum comparison. Three JSX reporter-ID mutations preserve successful execution and finding counts and change only Go comparison bytes 75/347/758. Bypassing ASCII-name or construction-range guards changes required panic 70 into exit 0 and fails the refusal contract. Three JSX numeric-kind +1 metadata mutations fail the external enum comparison. Reporting mutants do not qualify as complete source-rule mutants.

## Fresh native versus Go time

Three quiet interleaved whole-process runs per implementation and corpus, after builds finished. Each round compares complete stdout. ADAMIC_TSGO_TIMING=1 records native load/run phases; Go records its phases independently. Measurements include program loading and source parsing. No prepared-reporting/full-source timing ratio is claimed.

| Batch | Corpus | Native median seconds | Go median seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| first | compiler | 4.276 | 0.454 | 9.42 |
| first | repository | 0.440 | 0.114 | 3.85 |
| next | compiler | 2.795 | 0.392 | 7.13 |
| next | repository | 0.501 | 0.162 | 3.09 |
| third | compiler | 3.036 | 0.378 | 8.02 |
| third | repository | 0.431 | 0.101 | 4.27 |
| fourth | compiler | 3.113 | 0.514 | 6.06 |
| fourth | repository | 0.438 | 0.133 | 3.30 |

Native remains slower. Existing numeric metadata alone does not remove the current string-kind/lookup/dispatch overhead. The measurements establish timing and byte parity, not a causal attribution of all overhead. No inherited 26-rule gate, full repository gate, Stage 3 gate or nondefault-option parity was run.

Complete compressed gate and command streams, corrected and failed attempts, code/enum hashes, shared-ref status, prepared regex/reporting fixtures, Go reference source and all timing rounds are under evidence/c019. Twenty reproducible scratch binaries were removed from named older sixth-reporting directories, reclaiming 163,565,933 bytes; their source/log evidence was retained. cleanup.json names every removed file. Benchmark commands and all exit codes are retained. The report/evidence commit follows the tested code commit without changing executable sources.
