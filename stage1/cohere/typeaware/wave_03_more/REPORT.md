Built: no-throw-literal, no-useless-backreference and prefer-arrow-callback in native Adamic .a files, with one isolated compiler fact question.
Commits: claim eb012d3e; implementation and comparison evidence f8c455b34bdd6efb7831127baba3dd94887020c4; report commit follows.
Commands/results: TestWave03MoreAgreementAndMutants PASS 128.965s, 566 controls / 318 findings and all 364 frozen corpus roots byte-identical; sanitizer and released-handle checks pass.
Mutants: throw predicate, backreference direction and callback repair spacing caught by byte comparison; four compiler fact inversions and malformed-request guard caught by exact-access controls; retained released handle caught by panic-70 expectation.
Not covered: shared registration, emitted JavaScript execution of external-checker rules, a full repository gate, exhaustive regex-engine validation, and one independently rejected octal-escape fixture.

## Reservation and implementation

The six earlier claims were implemented, tested and pushed through 8fb10ebb before fetching all 348 origin refs. The first three unported and unclaimed names in the complete checker-dependent ranking were these three zero-volume core rules. All 33 origin claim files were inspected, including exact bare TypeScript suffixes and reservations for incomplete ports. The selection audit is in [continuation-2-selection.json](../validation-wave-03/continuation-2-selection.json). The new claim was pushed before any implementation.

Each rule has its own .a file. There are thirteen new native modules, including the driver and classes split into their own files. Existing type-aware parser, symbol, declaration, binding, framing and diagnostic modules are reused. No new native .ts file was created. The existing shared harness, registration generator, compiler implementation files and submodule pins were not edited.

`no-throw-literal` reproduces Go's syntactic optimistic expression judgment, operator-specific recursion and global-only `undefined` message. In particular, it retains Go's treatment of parenthesized throws rather than silently broadening the judgment.

`prefer-arrow-callback` tracks function ownership, including arrows inheriting their enclosing frame; checks self-reference by symbol identity; follows logical, conditional, parenthesized and chained-bind callback shapes; and preserves every ordered repair and every repair decline. The default and both nondefault Boolean options are compared against unchanged Go production settings. The callback walker owns its state, and the repair builder returns edits rather than mutating a caller's diagnostic.

`no-useless-backreference` scans group and reference paths in native code, with numeric/named references, duplicate named groups, disjunctions, lookarounds, Unicode flag checks and nested v-mode classes. Native reference tracking follows global constructors, global-object members, assignments, aliases, defaults and object destructuring, with symbol identity and write exclusions. Native constant evaluation follows same-file bindings and their writes, templates and string concatenations. All five finding ids and descriptions remain distinct.

The sole new bridge question, `source-access-context`, returns four raw compiler facts: declaration name, write access, type position and intrinsic JSX spelling. Go and Adamic definitions are in files named for the question, and an owned dispatcher handles it. The prior owned dispatcher changes its default to forward to that dispatcher. No lint verdict, regex classification, reference-tracker verdict, constant value or repair is supplied by Go.

## Independent comparison

The byte oracle loads its own TypeScript program and invokes the unmodified production Go rules from pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Its adapter imports no bridge code. It compares whole sorted finding records, rule names, ids, byte ranges, full messages, ordered fix ranges/text and all suggestion records. These production rules return no suggestions; the zero counts are compared rather than ignored.

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Positive and negative controls | 566 | 318 | 122499 |
| Frozen repository corpus | 287 | 0 | 18485 |
| TypeScript 6.0.3 src/compiler | 77 | 0 | 5857 |

The control generator extracts the complete four upstream fixture tables and adds ownership, shadowing, alias, constant-write, duplicate trace and Unicode cases. Of 567 candidates, Go independently rejects one source containing the legacy octal string escape `RegExp('\1(a)')` in module context. The exact rejected source and status are retained in [controls.json](validation/controls.json); it is not credited as agreement.

The 318 ordinary findings comprise 19 throws (17 object, 2 undef), 81 callbacks with 174 ordered fix edits, and 218 backreferences (43 nested, 95 forward, 28 disjunctive, 22 intoNegativeLookaround, 30 backward). Default controls and both complete corpora also match under ASan, UBSan and leak checks with empty native stderr. Both callback option runs match independently.

All 76 captured command streams are gzip-compressed with decompressed SHA-256 hashes. The artifact audit verifies every hash, fifteen whole-output equalities across ordinary, sanitized, option and timed runs, and empty stderr for successful sanitized and rule-mutant executions. Frozen corpus root hashes, configs, control sources and counts are retained alongside the streams. [Full run log](validation/more-final.log).

## Mutants and handle lifetime

| Mutation | Observation and independent catch |
| --- | --- |
| Invert `!this.couldBeError(argument)` | Native builds, exit 0, empty stderr; byte oracle differs at byte 55. |
| Invert the forward-reference direction condition | Native builds, exit 0, empty stderr; byte oracle differs at byte 37342. |
| Change arrow repair text from ` =>` to ` => ` | Native builds, exit 0, empty stderr; byte oracle differs at byte 8399. This changes fixes while preserving finding counts. |
| Invert declaration-name fact | Go mutant builds; exact compiler-access fields disagree. |
| Invert write-access fact | Go mutant builds; exact compiler-access fields disagree. |
| Invert type-position fact | Go mutant builds; exact compiler-access fields disagree. |
| Invert intrinsic-JSX fact | Go mutant builds; exact compiler-access fields disagree. |
| Disable malformed-question guard | Go mutant builds; malformed-request refusal assertion fails. |
| Retain a released registry handle | Native mutant builds and exits 0; required panic 70 catches it. |

The exact-access control checks 33 nonzero-width compiler nodes with both positive and negative cases for every field. Zero-width EOF nodes are excluded because the bridge refuses invalid ranges. A query after release produces precisely `adamic: panic: invalid or released checker handle` and exit 70. The retained-handle mutant proves this check fails when release is neutralized. Logs for all five fact/request mutants are in validation.

## Native time against Go

Single warm-process executions include loading and rendering the same complete output. The corpus timings were collected without the later parallel regression jobs running. These are observations from one round, not a throughput claim.

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository, 287 roots | 0.529153234 | 0.194678015 | 2.72 |
| Compiler, 77 roots | 4.810373187 | 0.517981163 | 9.29 |

Repository bridge load/query/run: 95.844193 / 11.389091 / 403.812850 ms, 311 queries. Go load/rule/run: 95.575051 / 5.420635 / 77.076192 ms. Compiler bridge load/query/run: 367.498875 / 355.507410 / 4274.624351 ms, 538 queries. Go load/rule/run: 351.031212 / 94.721447 / 134.567105 ms. Go rule time and bridge run time measure different boundaries and are not an isolated rule-to-rule speed comparison. Native is slower on both populations.

## Commands and limits

Commands began with `source /workspace/adamic-tools/env.sh`; all test output was redirected to files. The original successful setup remains applicable: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, warm cache 174s, total 174s; nproc 5, cgroup four CPUs. Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- `ADAMIC_WAVE03_MORE_ARTIFACTS=/workspace/wave-03/more-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03MoreAgreementAndMutants$' -count=1 -v`: PASS 128.965s.
- `go test ./bridge/tsgo/checker ./bridge/tsgo/code_path_graph -count=1`: checker PASS 0.678s; graph package has no direct tests. The graph remains exercised by the previous owned rule comparison.
- Five `go test -overlay ... ./bridge/tsgo/checker -run '^TestWave03SourceAccessContext$' -count=1 -v` compiling mutants fail their intended assertions; normal control PASS 0.094s.
- `go vet ./bridge/tsgo/checker ./bridge/tsgo/code_path_graph ./stage1/cohere/typeaware`: exit 0, empty log. Touched Go files are gofmt-clean; git diff --check is clean.
- `ADAMIC_WAVE03_QUICK=1 ADAMIC_WAVE03_NEXT_QUICK=1 go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants)$' -count=1 -v`: PASS 61.619s, prior-six regression including 96 Nexus controls, 12 ordering findings, one imported finding and 46 + 4 original findings.
- `TMPDIR=/tmp/adamic-gate go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(functions|maps_and_text|lone_surrogates|sorting|closures|string_index|method_closures|generic_functions)[.]a$' -count=1 -v`: eight source-Node / emitted-JavaScript / native / released / sanitizer and leak fixture checks PASS 34.299s. The separate one-byte oracle control passed in 8.270s; its first fixture filter selected no fixture subtests and was corrected explicitly with the full hierarchy.
- The production source gate, using the same scratch .a loader/reparse adapters as continuation 1, checks all thirteen native modules: 276 rules, 100% Adamic-ready, no findings, exit 0. Adapter sources and config are preserved. Production lint decisions are unchanged; these adapters are not used by the independent finding oracle.

An initial control config used an empty `files` list and was rejected before linting. Source-gate catches drove helper-class splits, walker ownership and property-alias removal. Superseded failure logs are retained and are not counted as passing checks.

The fetched `origin/codex/lint-harness-dot-a` is f4d98cab. Its report documents `.a` support and four-way certification for `stage1/cohere/lint`, with `rules/<slug>/rule.a` factories and that framework's Finding/Suggestion APIs. These ports use the separate existing typeaware Rules API and external tsgo C library. That branch does not add JavaScript external-checker execution to this typeaware driver. Running `adamic js wave_03_more/main.a` here exits 1 with `Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>` at the shared source-module query. The prelude and CLI also explicitly document native-only checker execution. No checker shim or shared compiler/harness edit was made to conceal this gap. Generic registration and emitted-JavaScript comparison for these checker rules remain integration work; the standalone .a native driver is tested and complete.

No full repository gate was run. The prior shared `TestInspectRequestRefusals` unknown-question mutant still assumes its mutation anchor is in shared facts.go; this unit leaves that harness untouched. The new owned malformed-access and released-handle checks are independently proven above. Regex validation intentionally follows production Go's partial validation rather than claiming a complete ECMAScript parser; numeric-overflow and arbitrary malformed-regex fuzzing were not covered.
