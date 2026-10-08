Built the merged slices and a row-keyed production census over all 173 dated ledger rows, with exact source and exclusion checks.
Commits: enum merge c2acd9dc, census merge 67e27a37, slice d merge 5d80bb93, integration repairs 8d96e943, final slice a merge 9610256b; pushed only codex/stricter-indexed-all.
Commands and outputs: seven witness suites, filtered oracle, scoped packages, vet and updated allocation counts pass; production schedules 168 contracts, retains 5 catch errors and 0 ordinary errors; all 99 indexed reads remain behind sys.ts namespace preflight.
Mutants: eight census/exclusion mutants caught; native witness guard erasures and the ownership ASan mutant are recorded in the attached complete test logs.
Not covered: 0 original-program checks emitted, 99 indexed read-local outcomes not reached; 5 catch-variable contracts pending; the unfiltered lower package has two namespace-preflight failures.

## Current result after enum, census and final slice merges

This section supersedes every historical result below. The enum initialization refusal at core.ts:19:52 is lifted. The final tree includes stricter-options-checks 51eab0da (and c132be26), records b150f83c (and 28d30cd3), slices a cf8048da, b 3ccc3268, c 00fa8dfc and d 1a83ba1a, and enum-init-reach 2152fc3b. All merge batches were pushed to this branch. Conflicts were inspected individually; newer census, namespace and finite readonly-record support were preserved alongside incoming optional-field proofs.

| Original ledger group | Scheduled | Emitted as a check | Refused at checker | Read-local not reached |
|---|---:|---:|---:|---:|
| Indexed reads | 99 | 0 | 0 | 99 |
| Optional writes | 67 | 0 | 0 | 67 |
| JSON stringify | 2 | 0 | 0 | 2 |
| Catch variables | 0 | 0 | 5 | 5 |
| Total | 168 | 0 | 5 | 173 |

[Every row by ID](evidence/census-row-outcomes.csv), [full row details](evidence/census-row-outcomes.json), [summary](evidence/census-summary.json) and [raw production output](evidence/census-production-result.json) distinguish census scheduling, emitted guards, pending checker errors and the common program blocker. The 99 indexed rows have zero measured read-local refusals and zero measured read-local checker errors because lowering has not reached them. Scheduling alone does not establish a native compilation date.

The production loader uses each .ts file's owning project configuration and converts supported stricter diagnostics into scheduled contracts. It does not apply literal tsc strict options. All 79 dated source hashes match. The census retains all 173 IDs and their exact locations/codes/options, including indexed-owned D069's joint attribution.

The unchanged whole compiler graph, entered through src/compiler/_namespaces/ts.ts, refuses at one of these two source locations (the preflight's traversal order varies):

* src/compiler/sys.ts:1492:142, `_fs.realpathSync.native`.
* src/compiler/sys.ts:1502:51, `memoize(() => process.cwd())`.

Both messages are: `stage 0 can't lower a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration yet`.

All 16 entry attempts (the actual compiler namespace plus each of the 15 indexed-row source files) retain the complete checked graph and stop at that preflight. No native artifact or indexed IR guard is emitted. No lowering refusal is bypassed.

Observation: both constructs resolve through pinned Node declarations; fs.d.ts declares namespace realpathSync and process.d.ts declares namespace process. Inference: the namespace initialization preflight treats these ambient declaration namespaces as pending executable initialization. Route to internal/lower/namespaces.go, namespaceInitialization / namespaceDeclaration / namespaceRuntime and its reachable-call graph. Origin branches namespaces-tsc 75a1d221, enum-tag-narrowing-2 64952252 and compiler/stage3-front-3 7a1fd0a9 were inspected; none contains a lift for this ambient namespace case. Their broader tips were therefore not imported as an unrelated workaround.

## Exact diagnostic exclusions

The authorized manifest remains exactly the dated 72 IDs: 67 primary optional-write rows and D108, D128, D132, D152, D153. The loader already schedules all 67 optional-write contracts, so those contracts and all their lowering maps remain active. The isolated probe removes only the five pending catch diagnostic messages. Normal production loading still reports those five checker errors. The probe overlay is confined to scratch files; source TypeScript, tsconfigs and the normal loader are not altered for this run.

The policy checks every ID, absolute file, UTF-16 line/column, code and single-option attribution, rejects joint D069 as an exclusion, and requires exactly five actual removals given the census. It preserves ordinary and indexed diagnostics byte for byte. Source identity, the exact manifest and overlay hashes are attached with the production evidence.

## Validation and mutants

The final validation logs are in logs/census-final-*.txt. Full slice suites cover both native backends in release and sanitized builds, Node present/absent behavior, pinned stderr, explain counts and runtime guard-erasure mutants. All four slices, records, base stricter options and integration routing are run together after the final merge. Earlier failures and repair logs remain available separately; they are not counted as mutant catches.

Eight probe mutants: omit one authorized exclusion; shift one exclusion location; disguise joint D069 as optional-only; drop one indexed audit; fabricate a built entry; substitute an excluded row ID; remove a census row; fabricate an emitted state. The first three must fail before lowering, and the other five must fail classification. Full output identifies each caught assertion. The ownership mutant builds successfully and ASan reports heap-use-after-free after removing standalone retains; a previous unused-variable compilation failure was rejected as inadequate evidence and the mutant was repaired.

D212 and D220: generators.ts, transformGenerators.hasImmediateContainingLabeledBlock and transformGenerators.tryEnterOrLeaveBlock. Their redundant receiver `!` assertions now lower, with one indexed-presence guard each, so the old redundant-assertion refusal is no longer a routing blocker. Their original whole-program sites still have not been reached.

The full lower gate's two node:fs.readFile subcases fail because the namespace preflight arrives before the expected unsupported-member diagnostic. They are reported, not silently made green. A separate lower run explicitly skips those named subcases to check the remaining package. No full repository gate is claimed. The filtered independent oracle, native package, loader/IR/CLI packages and vet commands and outputs are retained in logs. The setup run succeeded in 106.643 seconds with nproc=5 and CPU quota=4; timing lines: Go 0.204s, Node 0.213s, submodules 0.398s, markdown dependencies 0.464s, clang 0.948s, Go build 106.465s, warm cache 106.614s.

## Final-tree verification results

All seven witness packages pass on 9610256b: slice a 160.926s, b 157.739s, c 133.602s, d 240.248s, records 98.426s, stricter options 53.787s, integration routing 1.773s. Together the four slices cover all 99 minimal indexed witnesses, including D069 and nested D119. This minimal-program proof remains distinct from the blocked original whole program.

The scoped package run passes: loader 24.067s, lower 80.894s with exactly the two known namespace subcases skipped, CLI 5.600s. Vet passes. The independent filtered oracle passes in 60.261s with zero cache hits, 114 native misses and 142 Node misses. All eight final-tree probe mutants are caught. The full native package passed in 256.135s before the final slice a merge; the final merge's native behavior is covered by the complete witnesses and filtered oracle, not a claimed repeated full native-package run.

The allocation-count gate initially failed with stale recorded retain/release totals across 86 rows; full evidence is retained. The prescribed update-counts command passes in 32.578s and the verification passes in 12.935s. It measures the merged tree rather than changing compiler behavior; both outputs are attached separately. The earlier two unfiltered lower failures remain unresolved.

## Reproduction commands

With the toolchain environment sourced and GOPROXY=https://proxy.golang.org|direct, the final compiler revision is 9610256b. Test output is redirected to files.

```sh
python3 stage3/stricter-indexed-all/build-production-probe.py --tree /tmp/stricter-indexed-all-typescript --scratch /tmp/stricter-indexed-all-final-production
/tmp/stricter-indexed-all-final-production/production-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json /tmp/stricter-indexed-all-final-production/exclusions.json /tmp/stricter-indexed-all-final-production/native
python3 stage3/stricter-indexed-all/classify-census.py --result /tmp/stricter-indexed-all-final-production/result.json --output stage3/stricter-indexed-all/evidence
python3 stage3/stricter-indexed-all/probe-controls.py --scratch /tmp/stricter-indexed-all-final-production --tree /tmp/stricter-indexed-all-typescript
ADAMIC_GATE_UNCACHED=1 go test ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-records ./stage3/stricter-options ./stage3/stricter-indexed-all -count=1 -timeout 10m -v
go test ./internal/load ./internal/lower ./cmd/adamic -skip 'TestNodeLibraryNamesUnimplementedMembers/node:fs.readFile' -count=1 -timeout 10m
go vet ./internal/load ./internal/lower ./internal/native ./cmd/adamic ./stage3/stricter-indexed-all
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnumInitialization|TestNamespace|TestParameterPropert|TestOptional.*Guard|TestOptionalImplementsRepresentation|TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers|records|catch_values|field_access_paths)' -count=1 -timeout 10m -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m
```

The original unfiltered lower run is retained in logs/census-full-lower-failures.txt. The final-tree package run names its two skipped subcases explicitly above. The standalone census probe and classifier exited zero because they recorded all attempts. Every lowering attempt refused; that evidence-collection success is not a successful native build.

## Historical evidence before enum and census integration

The following is retained for provenance only; its old counts and blockers are superseded above.

Built the ruling-mode whole-program probe with exactly 72 authorized row exclusions, merged all current slices/records, and classified every indexed row.
Commits: scoped probe 0820cc0a, newest base 4fe06650, final owner-fix merge 783509f4; all pushes go only to codex/stricter-indexed-all.
Commands and outputs: seven merged witness packages, compiler package checks, filtered oracle and vet pass; 79 source hashes match; production checker succeeds; all 16 lowering entries refuse at core.ts:19:52.
Mutants: merged runtime guard erasures and null/undefined conflation caught; six probe/report mutants caught; exclusion policy preserves ordinary and indexed diagnostics unchanged.
Not covered: original whole-program indexed checks are not reached (0 emitted, 99 program-refused, 0 checker errors); excluded optional/catch checks, nested D119 records and D069's absent optional-field storage context remain unproven.

## Current production-mode whole-program result

This is the ruling's mode: the unchanged .ts files are checked with their owning project's options. The production loader audits the stricter flags and retains all 99 indexed rows for lower.checkedIndexedRead / indexedPresenceGuard. The probe does not enable literal tsc noUncheckedIndexedAccess on the owning project.

| Indexed row status on the original program | Rows |
|---|---:|
| Compiles as an emitted check | 0 |
| Still refuses in its enclosing original dependency graph | 99 |
| Still checker-errors | 0 |

All 99 rows are enumerated individually, with the measured refusal and its scope, in [production-indexed-status.csv](evidence/production-indexed-status.csv) and [full JSON](evidence/production-indexed-status.json). Their read-local outcome is explicitly `not reached`. This is 99 rows behind one common program refusal, not 99 independently unsupported receiver representations. No native artifact or indexed IR guard was emitted. The program count cannot establish a completion date for native TypeScript.

Production loading succeeds with zero remaining checker diagnostics, all 79 dated source hashes equal, and all 99 exact indexed audit identities retained, including jointly attributed D069. Lowering from the real compiler namespace entry, src/compiler/_namespaces/ts.ts, then stops at:

`src/compiler/core.ts:19:52: stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet`

The exact source construct is the module-level emptyMap initialization, `new Map<never, never>()`. Route this to internal/lower/enum_initialization.go, (*lowering).enumInitialization. The probe additionally selects each of the 15 original indexed-row files as the single lowering entry of that same complete checked program. All 16 entry attempts report the same NotYet refusal before any indexed-site lowering. The 79 checker roots remain present; selecting a lowering entry is not removing a source or dependency. Production Lower requires exactly one entry, so it is never called with the old harness's 79-entry invocation.

[Raw production result](evidence/production-result.json) includes every option audit, every attempted entry and the exact error. [Summary](evidence/production-summary.json) keeps checker success, emitted checks, native artifacts and unmeasured read-local outcomes separate. No enum check or other lowering refusal was bypassed. Clang is reached only if loading and lowering succeed.

## Exactly 72 authorized diagnostic exclusions

The exclusion list is generated by row ID from the dated 173-row ledger: the 67 primary exactOptionalPropertyTypes rows and five useUnknownInCatchVariables rows. [Every excluded ID, original location/code, option and observed diagnostic](evidence/production-excluded-rows.csv) is retained. [Loader manifest](evidence/production-exclusions.json) records the same IDs and exact absolute file, UTF-16 line/column and diagnostic code.

The build tool makes a scratch Go overlay of internal/load/load.go. Its ordinary Load/LoadOverlay calls retain their original behavior. Only the separately generated production probe calls the overlay's LoadIndexedLedgerProbe. Immediately before the existing CheckError gate, its policy:

1. Requires exactly 72 unique identities, with 67 optional and five catch exclusions.
2. Requires an exact identity match in the actual production option audit, and exactly the requested single-option attribution.
3. Removes only the complete observed audit messages for those identities, requiring exactly 72 removals. Duplicate or unmatched exclusions refuse.
4. Leaves all indexed and unrelated diagnostics intact. D069 is indexed-owned and has two options, so it cannot be excluded as an optional row.

The original loader file is unchanged on disk; the Go overlay changes no TypeScript source, project configuration, declaration input or indexed check. [Overlay provenance](evidence/production-overlay-provenance.json) records loader hashes and the 72 IDs; [source identity](evidence/production-source-identity.json) records 79 matches and zero mismatches. This scopes the experiment to indexed work; it supplies no optional-write/catch runtime proof.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-all-followup-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/stricter-indexed-all/build-production-probe.py --tree /tmp/stricter-indexed-all-typescript --scratch /tmp/stricter-indexed-all-production-complete > /tmp/stricter-indexed-all-production-complete-build.log 2>&1
/tmp/stricter-indexed-all-production-complete/production-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json /tmp/stricter-indexed-all-production-complete/exclusions.json /tmp/stricter-indexed-all-production-complete/native > /tmp/stricter-indexed-all-production-complete/result.json 2> /tmp/stricter-indexed-all-production-complete/run.log
python3 stage3/stricter-indexed-all/classify-production.py --result /tmp/stricter-indexed-all-production-complete/result.json --tree /tmp/stricter-indexed-all-typescript --output stage3/stricter-indexed-all/evidence > /tmp/stricter-indexed-all-production-complete-classify.log 2>&1
```

The build/classifier commands and probe reporting process exit zero. This reports compiler refusals successfully; it does not mean a native compilation succeeded. Source reproduction uses the exact 3b255125 adaptation described in the historical section below, not newly merged adaptation scripts. Setup passed in 69.544s: Node/Go 0.022s, submodules 0.053s, markdown 0.067s, clang 0.136s, Go build 69.360s, cache warm 69.509s; nproc=5, CPU quota four. [Setup log](evidence/production-setup.log).

## Latest integrated tips and validation

Merged into codex/stricter-indexed-all and pushed after each merge: stricter-options-checks c132be26, records 28d30cd3, c 00fa8dfc, a through d1390ca8 and d through 71897d7e; b remains 3ccc3268. The implementation merge checkpoints include 2b8938f7, 4c6c41a4, 2de83a52, 090c9f55, 4fe06650, b409085e, 8bf7cb55, 26ed3c2c, 47d45bd0, 9d398b66 and 783509f4. The current base brings its own upstream main ancestry, explicitly requested with c132be26; no main or area branch was pushed.

Conflict resolutions were hunk-specific: a's implicit binding receiver/element parameters plus d's null-sentinel condition; both record and typed-array expression boundary checks; c's equivalent sparse initialization order; d's null observation fields plus chained guard counts; and d's explicit record-hole and independently observed outer/inner variants. The dependency gate caught the overly broad record cast precheck masking the typed-array refusal. The final d owner fix scopes that precheck to records; its scoped condition and record-specific message supersede the interim generic diagnostic fix.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-records ./stage3/stricter-options ./stage3/stricter-indexed-all -count=1 -timeout 10m -v > /tmp/stricter-indexed-all-final-merged-suites.log 2>&1
go test ./internal/load ./internal/lower ./internal/ir ./internal/javascript ./internal/native -count=1 -timeout 10m > /tmp/stricter-indexed-all-current-compiler-packages.log 2>&1
go test ./internal/native -count=1 -timeout 10m > /tmp/stricter-indexed-all-final-native-package.log 2>&1
go test ./internal/lower -count=1 -timeout 10m > /tmp/stricter-indexed-all-complete-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers|enums_open_never_index|cast_enum[^/]*)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-all-current-oracle.log 2>&1
go vet ./stage3/stricter-indexed-all ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-records ./stage3/stricter-options ./internal/lower > /tmp/stricter-indexed-all-final-vet.log 2>&1
```

Final merged gate exits zero: stricter-indexed-a 113.630s, stricter-indexed-b 110.618s, stricter-indexed-c 101.888s, stricter-indexed-d 129.065s, stricter-records 41.880s, stricter-options 51.016s, stricter-indexed-all 0.738s. Final lowering package passes in 39.679s; load in 22.597s, IR in 47.856s, native in 113.498s; JavaScript has no package tests. The uncached filtered oracle passes in 19.218s (native misses=29, Node misses=24). Both final witness/lowering vet and overlay probe vet exit zero. An environment restart interrupted earlier aggregate runs; their completed package observations are retained, and every interrupted package was rerun. All seven witness packages were then rerun together successfully on the final merged tree. [Final gate](evidence/production-final-suites.log), [compiler packages](evidence/production-compiler-interrupted.log), [native](evidence/production-native.log), [final lowering](evidence/production-final-lower.log), [oracle](evidence/production-oracle.log), [vet](evidence/production-final-vet.log), [overlay vet](evidence/production-final-overlay-vet.log).

The merged minimal witnesses prove 98 ledger rows and pin one refusal: D119's nested record of records is outside the readonly scalar/scalar-array representation. D037 is now proven with its real overload/array-read attribution; a's ordinary optional-return control still errors. D069's array check is proven with its optional receiving field present; the original absent-field storage context remains a separately named dependency. These are minimal read/contract witnesses, not 98 native original compiler sites.

Runtime erased-check mutants are rerun in the full merged suites, including implicit bindings, sparse holes, typed-array reads, record-key and inner-array guards, and null/undefined sentinel conflation. Exact exit/stdout/stderr and explain assertions catch them; compilation failure is never a mutant kill. The full [merged suite log](evidence/production-final-suites.log) retains each site and mutant observation.

Six scoped probe/report mutants pass: delete one exclusion; shift a row identity; try to exclude joint indexed D069; drop an indexed audit row; fabricate a built entry; relabel an exclusion ID. [Control results](evidence/production-controls.log). The overlay unit probe additionally removes the actual 72 optional/catch messages while preserving an ordinary and an indexed diagnostic unchanged ([log](evidence/production-preservation.log)). The template is in probe/diagnostics_test.go.txt and is compiled only via the recorded scratch controls-overlay.json, with ADAMIC_INDEXED_LEDGER_PROBE_RESULT pointing at the raw result. These are probe-boundary controls, separate from runtime check mutants. The classifier refuses any future successful/later-stage attempt until original source-to-guard attribution is implemented, so a built-stage claim cannot masquerade as this refusal count.

D212/D220's exact receiver-assertion reductions still name the same functions and message in the routing section below. Their current rerun is included in the final suite log; the whole program did not reach those assertions. Remaining whole-program indexed work starts with the enum-initialization dependency, not the excluded optional/catch diagnostics. No full repository/lint gate was run.

## Historical pre-exclusion measurements

The following pre-exclusion and literal-strict observations are retained as history. They are superseded for the ruling's current count by the production result above.

## Whole-program result

The original whole program did **not** reach lowering or clang. This is the count that constrains the date, rather than extrapolating from minimal witnesses.

| Profile | Compiler roots | Checker diagnostics | Indexed checker errors | Original indexed native outcomes |
|---|---:|---:|---:|---|
| Literal project strict options enabled | 79 | 171 | 99 | all 99 unmeasured |
| Production project options plus Adamic option conversion | 79 | 72 | 0 | all 99 unmeasured |

The literal project profile enables noUncheckedIndexedAccess, exactOptionalPropertyTypes, useUnknownInCatchVariables and strictBindCallApply in an in-memory overlay of the owning compiler tsconfig. Its 171 diagnostics are exactly the ledger's 99 indexed, 67 optional and five catch rows. No source expression was rewritten. Every indexed ID and its measured diagnostic appears in [indexed-status.csv](evidence/indexed-status.csv), with full metadata in [indexed-status.json](evidence/indexed-status.json).

The production profile preserves the project's own settings. Adamic audits the stricter options and defers the 99 indexed diagnostics to runtime guards. It still rejects the 67 optional and five catch diagnostics, so none of those indexed sites can be measured in native output. Both raw results retain the full diagnostics and option attribution: [literal strict](evidence/whole-strict.json) and [production](evidence/whole-production.json). Two JSON option sites are also recorded by production attribution; these are not checker diagnostics.

Consequently this run establishes **99 indexed errors under literal strict settings**, and **99 whole-program indexed outcomes blocked by 72 earlier errors under production conversion**. It establishes neither 99 native successes nor a per-site native refusal count. Errors were not suppressed to manufacture a lower-stage result. The optional/catch branches were outside this integration's authorized branch list and were not merged.

## Exact tree and reproduction

The named checker-259 directory has LEDGER.md, not a README. Read the dated ledger at a1a16427a46149435e25f847c45ad136f1f1e55c. Its adaptation input is 3b25512566206bc603b93264e8d072c55e075d64, rather than this integration's current stage3 patches. Apply that revision's stage3/apply.sh and run npm ci. All 79 implementation source hashes match the dated ledger ([identity](evidence/source-identity.json), [expected hashes](evidence/ledger-source-hashes.json)). The options inventory's eightieth root is its virtual prelude declaration; production Load injects its own prelude, so the probe passes the 79 implementation roots, not that virtual filename.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-all-setup.log 2>&1
source /workspace/adamic-tools/env.sh
git worktree add --detach /tmp/stricter-indexed-all-ledger-base 3b25512566206bc603b93264e8d072c55e075d64
bash /tmp/stricter-indexed-all-ledger-base/stage3/apply.sh /tmp/stricter-indexed-all-typescript > /tmp/stricter-indexed-all-apply.log 2>&1
npm ci --prefix /tmp/stricter-indexed-all-typescript --ignore-scripts --no-audit --no-fund > /tmp/stricter-indexed-all-npm.log 2>&1
go build -o /tmp/stricter-indexed-all-probe ./stage3/stricter-indexed-all > /tmp/stricter-indexed-all-probe-build.log 2>&1
/tmp/stricter-indexed-all-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json project-strict /tmp/stricter-indexed-all-strict-native > /tmp/stricter-indexed-all-whole-strict.json 2> /tmp/stricter-indexed-all-whole-strict.log
/tmp/stricter-indexed-all-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json production /tmp/stricter-indexed-all-production-native > /tmp/stricter-indexed-all-whole-production.json 2> /tmp/stricter-indexed-all-whole-production.log
python3 stage3/stricter-indexed-all/classify.py --tree /tmp/stricter-indexed-all-typescript --strict stage3/stricter-indexed-all/evidence/whole-strict.json --production stage3/stricter-indexed-all/evidence/whole-production.json > /tmp/stricter-indexed-all-classify.log 2>&1
```

Both probe commands exit 1, as expected from their recorded checker rejection. The probe calls native.Build only after successful loading and lowering. Initial harness construction mistakenly counted the virtual prelude among implementation roots and was corrected before these recorded runs. Setup passed: Node 0.017s, Go 0.019s, submodules 0.051s, markdown 0.056s, clang 0.132s, Go build 37.223s, warm 37.331s, done 37.354s; nproc=5, CPU quota four, Go 1.27.1, clang 20.1.8, Node 24.19.0. [Setup log](evidence/setup.log).

## Merged witness coverage

Base dfb82dabbf7b548de84a334ae5e72935bd08ce90 already implements bounded sparse-array presence. Git merges were clean; semantic conflicts were reconciled by inspecting the affected hunks: c D155/D156 now initialize one slot for the present run and leave a hole for the absent run, and d's hole variants follow the same runtime/explain/mutant assertions as its dense variants. No compiler implementation file was edited.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-options ./stage3/stricter-indexed-all -count=1 -timeout 10m -v > /tmp/stricter-indexed-all-final-suites.log 2>&1
go vet ./stage3/stricter-indexed-all ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d > /tmp/stricter-indexed-all-vet.log 2>&1
```

All pass: a 56.460s, b 59.668s, c 44.942s, d 72.540s, base options 18.317s, integration routing 0.694s; vet exits zero. [Complete merged log](evidence/merged-suites.log). Coverage is 90 minimal ledger rows proven, nine refused, zero missing manifests. D060/D061 share a program and other downstream rows can share an original read; 90 rows is not 90 original whole-program guard sites.

Each supported witness verifies Node present/undefined observations, backend JavaScript and release/sanitized native observations, exact named stderr and exit 70, and explain coverage. Each erase-panic mutant must compile and fail the pinned contract; successful compiler errors are never counted as mutant kills. [Every logged mutant outcome](evidence/mutants.log) records the supported rows and sparse variants. Existing slice a logs include its individual erasures; the merged test reruns them. The classifier additionally rejects a missing D220 diagnostic and a fabricated built-stage result ([log](evidence/classifier-mutants.log)).

| Minimal refused row | Observed reason |
|---|---|
| D037 | Type 'SourceFile \| undefined' is not assignable to type 'SourceFile' |
| D071 | destructuring anything but a tuple into [names] |
| D072 | destructuring anything but a tuple into [names] |
| D073 | destructuring anything but a tuple into [names] |
| D119 | Adamic 0.1 refuses an index signature; use a Map |
| D129 | Adamic 0.1 refuses an index signature; use a Map |
| D130 | Adamic 0.1 refuses an index signature; use a Map |
| D131 | Adamic 0.1 refuses an index signature; use a Map |
| D151 | stage 0 can't lower a value of type string \| null yet |

D037's optional return is an attribution issue, not an established indexed guard. D069 preinitializes its receiving optional field; slice a retains the separate field-write counterexample in gaps/optional-field.json. Minimal proofs preserve read shapes, not every enclosing compiler operation.

## D212 and D220 routing

Both are in src/compiler/transformers/generators.ts:

| Row | Function | Original receiver syntax |
|---|---|---|
| D212 | transformGenerators.hasImmediateContainingLabeledBlock | line 2445: blockStack![j]; downstream diagnostic line 2446 column 48 |
| D220 | transformGenerators.tryEnterOrLeaveBlock | line 2983: blockOffsets![blockIndex]; diagnostic column 57 |

With the receiver made a guaranteed array, each standalone exact-`!` reduction still refuses in lowering:

`Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`

[Routing test and observations](evidence/routing.log) establish the redundant-assertion refusal, not a claim that the checker-blocked whole program reached it. Slice b's runtime witnesses erase the redundant receiver assertion to test the indexed guard itself. Route assertion support separately from indexed presence.

## Integration tips and remaining work

Merged and pushed after each merge: b 3ccc3268, c f5a2212c, d c019ca20 then d22099f0 and 9191d209, a b7214299 then 90342391, a17a300f and f5a63e24. Integration evidence b250367c follows f2d8c2d4; retained logs are committed as 19339ca6. The final documentation-only D151 investigation was merged as 976cf0f5; no source or witness changed, so the completed gate remains applicable. Origin was checked each batch and after the final suite. No codex/stricter-records branch was published at the final check. After validation, stricter-options-checks advanced to 7d70c572 via a large origin/main merge and enum-length sparse-array change; this integration retains its start-time newest base dfb82dab and the user-requested later a/c/d tips. The new base was observed, not merged or tested here. Newer optional-write tips were observed but not merged. Only codex/stricter-indexed-all was pushed.

Remaining: remove the 72 whole-program optional/catch checker blockers in their owning units, rerun this exact native pipeline, then classify original indexed sites from actual lowering/native results. Records, array destructuring, nullable string representation and redundant assertion support remain separately routed. No full repository gate was run; the exact six-package gate and three-package vet above were run. Minimal witness completion cannot establish the whole-program completion date.
