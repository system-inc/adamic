# Combined host proof

24/25 adapted fixtures agree with Node on both backends. All 25 were rerun after merging afdb4c94 (merge d6149508). The conditional identity refusal is fixed; fixture 25 now refuses NonNullable<T> seen as T at 1097:26. The WebAssembly host test suite passes. Pristine controls remain 16/25 on each backend.

Final additions: Buffer fallback ccb8a69 via merge 469bb8b9; exact audited adapted fixtures/status/Node records from 21ef072e via 33e60869; Date conversion 05635aa via merge 0e7a2cc3; scalar concatenation 998fb3eb via merge b78d7ea7. stage3/fixtures/host is identical to the adapted branch at 9eba5d10. The pristine run uses the unchanged pre-adaptation fixture snapshot at 08b5b2c4. Agrees requires exact stdout, stderr and exit against Node, including exit 1 and 2 fixtures.

| Fixture | Pristine native | Pristine JS | Adapted native | Adapted JS | First adapted blocker | Owner |
|---|---|---|---|---|---|---|
| 01_readFile_utf8.a | Checker | Checker | Agrees | Agrees | None | None |
| 02_readFile_utf16le.a | Checker | Checker | Agrees | Agrees | None | None |
| 03_readFile_utf16be.a | Checker | Checker | Agrees | Agrees | None | None |
| 04_readFile_missing.a | Checker | Checker | Agrees | Agrees | None | None |
| 05_writeFile.a | Refused | Refused | Agrees | Agrees | None | None |
| 06_fileExists.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 07_directoryExists.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 08_getDirectories.a | NotYet | NotYet | Agrees | Agrees | None | None |
| 09_realpath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 10_getModifiedTime.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 11_setModifiedTime.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 12_deleteFile.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 13_createDirectory.a | Checker | Checker | Agrees | Agrees | None | None |
| 14_getCurrentDirectory.a | Disagrees | Disagrees | Agrees | Agrees | None | None |
| 15_getExecutingFilePath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 16_getEnvironmentVariable.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 17_write.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 18_exit_0.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 19_exit_1.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 20_exit_2.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 21_createHash.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 22_createHash_fallback.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 23_newLine.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 24_useCaseSensitiveFileNames.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 25_readDirectory.a | Checker | Checker | Refused | Refused | NonNullable<T> seen as T at 1097:26 | Compiler |

Fixture 25 passes checking on both backends. Exact next diagnostic:
```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:530:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```
Verified one-line reproductions for the new adapted blockers, on both backends:

```typescript
type Path = string & { readonly __pathBrand: any }; function ensureTrailingDirectorySeparator(path: Path): Path; function ensureTrailingDirectorySeparator(path: string): string; function ensureTrailingDirectorySeparator(path: string) { return path; } console.log(ensureTrailingDirectorySeparator('x'));
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/08_overload_path.a:1:53: Adamic 0.1 refuses overload 1 of ensureTrailingDirectorySeparator result Path cannot be served by implementation result string; make the implementation result covariant with every overload result

```typescript
function isArray(value: any): value is readonly unknown[] { return Array.isArray(value); } console.log(String(isArray([])));
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/25_array_predicate.a:1:61: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)

The now-resolved 01 array-of-never and surviving 14 callback-cycle, and narrowed-number value! reproductions and the isolated area/runtime scratch check are preserved in REPORT-before-final-additions.md. The graph-region prerequisite was absent from the fetched runtime area; the scratch merge remains excluded.

Previous validation (superseded where noted by the recount below):

- Final original and adapted commands: node_fs_file_host_check.py --all --compiler /tmp/fs-combined-adamic, with separate logs and JSON reports. Both exit 1 because named blockers remain, not because of a miscomparison; all 50 source cases observed on both backends. Pristine runner uses an unchanged temporary snapshot. Adapted count uses the full audited set, not the earlier five-fixture adaptation-47 approximation.
- TestNodeBufferRefusals PASS, 3.177s. node_buffer_fallback native/JS oracle PASS, 0.482s. fallback_identity mutant CAUGHT only by Node stdout in both backends; source restored.
- Date tests: zones, named refusal, runtime mutants and WASI refusal PASS. Private runtime mutants do not modify the working tree. Final scalar concatenation and primitive spelling mutants PASS, caught only by Node stdout on both backends.
- Final ADAMIC_ORACLE_WASI=1 host leg plus scanner mutants and Date WASI refusal: exit 1, 21.327s. Sole failure remains the routed empty symlink target mismatch, Node ENOENT versus WASI EINVAL. Runtime archive compiles, nine host refusal probes pass; honest target skips are not counted as successes.
- Full go test ./... was not run. Earlier full affected-package failures and counts regeneration limitations are preserved in the historical report. Counts merge retains each side's fixture rows once; full regeneration is not claimed. No compiler or routed library blockers were implemented here. No macOS execution.

## Recount with typed stat, WASI symlink and contextual empty arrays

Inputs: exact fixture tip 17c5385a via commit 027f8e5d; library/merge-p2b a5d5dc9 via merge 35026fd0; compiler array-literal-never-element b3578751 via merge a564f4bd. No rebase, force push or main/area push. The empty-array merge keeps existing Buffer/array-predicate dispatch and the compiler's contextual empty-literal path.

All 25 adapted and all 25 pristine controls rerun after the final compiler merge, on native and JavaScript: 21/25 and 15/25 agree with fresh Node stdout/stderr/exit. Both commands exit 1 solely for named blocked fixtures; a refusal is not a success. Reports are node_fs_file_combined_adapted_status.json, node_fs_file_combined_original_status.json and the composed node_fs_file_combined_all_status.json. Fixture 01 now agrees on both backends.

Fixture 08 includes Stats | Dirent | undefined at line 199, replacing any. The earlier overload refusal at line 160 still blocks it on both backends: ensureTrailingDirectorySeparator's Path overload result cannot be served by the implementation's string result. Its verified reproduction above remains applicable. Fixture 25 is compiler-owned: core.ts isArray returns a readonly unknown[] predicate via Array.isArray; compiler's host-blockers worker is handling this builtin predicate.

Final validation commands, output recorded directly to logs:

- `python3 internal/oracle/node_fs_file_host_check.py --all --compiler /tmp/fs-combined-adamic --logs /tmp/fs-recount-array-adapted --report internal/oracle/node_fs_file_combined_adapted_status.json`: exit 1, 25 observed, 21 agree on both backends. Unchanged pristine snapshot via /tmp/fs-combined-original-check.py: exit 1, 25 observed, 15 agree.
- `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIEmptySymlinkAgreesWithNode|TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals|TestWASIFileAgreesWithNode|TestNodeFSFileCombinedNullClassifier|TestDateStringWASIRefusal)$' -count=1 -timeout 15m -v`: PASS, 25.853s. Whole runtime archive builds; empty symlink now matches Node on native and WASI. Existing honest target skips remain excluded from success counts. This supersedes the earlier WASI mismatch.
- Removed-symlink-adapter mutant executes cleanly and changes ENOENT to EINVAL; caught only by Node stdout. Null classifier mutants also pass their Node-only catcher checks.
- `go test ./internal/lower -run 'TestNestedEmptyArrayElementKinds|TestEmptyLiteralGenericReturnUsesSamePath' -count=1 -timeout 10m`: PASS, 0.242s.
- `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/array_literal_empty_.*.a$' -count=1 -timeout 10m -v`: four fixtures PASS on both backends with sanitizer/leak checks, 1.095s.
- Private Go overlay changes the empty-literal element kind to Boolean: native builds, ordinary output comparisons pass, LeakSanitizer catches 74 bytes leaked in one allocation. Exit 1, 0.362s. This is a lifetime-check mutant, not a Node-output mutant; working source is unchanged.
- `python3 stage3/fixtures/host/check.py --mutants-only --logs /tmp/fs-recount-fixture-mutants`: PASS, all 25 semantic mutants caught by Node, including 08 returning files instead of directories, plus the recorded-diagnostic mutant.
- `git diff --check`: PASS. Existing toolchain reused; nproc 5. GOPROXY set before Go commands. Full go test ./..., full language WASI, complete counts regeneration and macOS execution were not run. Prior full-gate limits remain recorded above.

## Recount with a85a9cb1 and fixtures 7218520a

Compiler a85a9cb1 merged with merge commit 86454464; exact fixture tip 7218520a taken in 5e40b960. Conflicts retained enum/void type representations alongside the opaque-array unknown representation and kept both sets of counts rows. No runtime code changed.

Final full reruns: 19/25 adapted and 15/25 pristine agree with Node on native and JavaScript. All 25 sources in each set were observed on both backends; commands exit 1 for named blockers. This supersedes the earlier 21/25 adapted count.

Fixture 11 remains green on both backends, with exact fresh Node stdout/stderr/exit agreement. Its source has no diff against the prior proof. Fixtures 06, 07 and 10 remain green; 08 retains the Path/string overload Refused diagnostic and 24 remains NotYet for the regex replacement callback. No utimesSync regression occurs in this combined proof.

Two regressions are introduced by compiler a85a9cb1 (merge 86454464), not by the fixture update: 05 and 13 sources are unchanged from 5736b24, but their shared errorCode unknown parameter is now treated as an opaque Array.isArray-only input. Both backends stop at:
```
adamic: /workspace/adamic/stage3/fixtures/host/05_writeFile.a:35:19: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet
adamic: /workspace/adamic/stage3/fixtures/host/13_createDirectory.a:72:19: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet
```
Verified compiler reproducer on both backends:
```typescript
export function errorCode(error: unknown): string | undefined { return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string" ? error.code : undefined; } console.log("unused");
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/05_unknown_error_observation.a:1:79: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet

Fixture 25 remains Refused at 530:5, adamic/no-type-predicate. The delivered source still declares isArray(value: any): value is readonly unknown[], while a85a9cb1's proof supports an unknown parameter and explicitly rejects any. Taking typed stat at line 1238 does not change that earlier signature. Its one-line reproduction above remains verified; owner stays compiler. No additional source adaptation was invented.

Validation after merging:

- All 25 delivered fixtures: node_fs_file_host_check.py --all, log fs_predicate_final_adapted.log, 19 agree on each backend, exit 1. All 25 unchanged pristine controls: fs_predicate_final_original.log, 15 agree, exit 1. Node records freshly verified for every row.
- Predicate lowering regressions: `go test ./internal/lower -run 'TestArrayPredicate|TestUnknownArrayPredicate|TestPredicateBodiesAreProven|TestUnprovenPredicate' -count=1 -timeout 10m`: PASS, 1.334s.
- host_array_unknown_predicate source/native/JS oracle: PASS, 0.726s, including sanitizer/leak checks.
- Private Go overlay brand-test mutant tests undefined instead of the actual argument. Builds successfully; Node reports exit 0 while native and JS report exit 70 after incorrect false results. Both Node comparisons catch it; no compiler or sanitizer error. Working source unchanged.
- `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIEmptySymlinkAgreesWithNode|TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals|TestWASIFileAgreesWithNode)$' -count=1 -timeout 15m -v`: PASS, 21.546s, including empty-symlink controls and the removed-adapter Node mutant. Existing explicit target skips are not counted as passes.
- Full go test ./..., full language WASI and counts regeneration were not run. Earlier gate limitations remain documented. This remains a proof branch; no main/area push, no rebase and no force push.

## Final recount including ed6e30e2

Merge 904b2258 brings compiler non-null-narrowed-number ed6e30e2 and its newer typed-array/compiler ancestry into the proof. Conflicts preserve both host operations and typed-array dispatch, both Date and typed-array representations, Buffer contextual views and typed-array set arguments, fresh analyses and all unique fixture/count rows. One identical logicalAssignment helper was duplicated by the two compiler branches; the existing helper is retained and the incoming duplicate removed.

All 25 adapted fixtures at 7218520a and 25 pristine controls were rerun on both backends after this merge: still 19/25 adapted, 15/25 pristine. Fixture 11 is green on both backends, as are 06, 07 and 10. The errorCode regressions from a85a9cb1 remain in 05 and 13. Fixture 25 still has an any parameter and remains refused at the predicate proof.

The prior native-only narrowed-number gate stop is resolved. non_null_narrowed_scalar.a builds and executes on native and JS, both print 0 and compare byte-for-byte with source Node. The old double != NULL compiler diagnostic is historical, not a current blocker.

Fixture 14's first blocker now precedes the capture cycle:
```
adamic: /workspace/adamic/stage3/fixtures/host/14_getCurrentDirectory.a:17:24: Adamic 0.1 refuses a non-null assertion whose operand is exactly undefined; declare the variable optional and assign undefined
```
This is ed6e30e2's explicit exact-nullish assertion ruling applied to the source's callback = undefined! write. No adaptation was added here.

Final checks:

- All adapted and pristine fixtures rerun in fs_predicate_nonnull_adapted.log and fs_predicate_nonnull_original.log; exit 1 solely for named blocked fixtures, no output mismatches.
- `go test ./internal/lower ./internal/fresh -run 'TestNonNull|TestExactlyNullish|TestNodeFSFile|TestArrayPredicate|TestUnknownArrayPredicate' -count=1 -timeout 10m`: PASS, lower 7.442s, fresh 0.024s.
- `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIEmptySymlinkAgreesWithNode|TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals|TestWASIFileAgreesWithNode|TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullStopsAtAssertion)$' -count=1 -timeout 15m -v`: PASS, 51.857s. Whole runtime archive, empty-symlink Node mutant, all selected WASI host tests and exact-nullish/possible assertion checks pass. Explicit target skips remain excluded from success counts.

- Full `go test ./internal/flow -count=1 -timeout 10m`: FAIL, 109.401s. The program sweep includes incoming non_null_deinitialize_* fixtures now explicitly refused by ed6e30e2's exact-nullish assertion ruling; it cannot lower them for SSA/range/trace checks. Full diagnostics, including all other flow findings, are in fs_predicate_nonnull_flow.log. No flow success or full repository gate success is claimed.
- `git diff --check`: PASS. No new POSIX runtime calls were added; merged runtime archive passes the WASI host leg. These compiler branches remain proof inputs, not area/main landings.

## Regex callback recount and fixture tip 0555bc27

Regex branch a8ddb3a7 merged in 04529736. Exact fixture tip 0555bc27 taken in 9ecec994. The merge brings newer async/concurrency ancestry along with regex callbacks; no extra host-blockers fix was merged. Required reconciliations retain the argument-count closure ABI, compiler optional/default/rest support, regex callback rest specialization, typed-array dispatch, atomic packed field caches and initialized-field bits. Readiness indexes use actual returned slots rather than removed cache fields. Unsupported typed-array publication uses the incoming sharing path's existing non-shareable panic instead of silently publishing mutable storage; cross-worker typed-array publication is not claimed supported.

All 25 newest adapted and all 25 pristine sources rerun on native and JavaScript after the final rest-parameter reconciliation: 20/25 adapted, 16/25 pristine agree with fresh Node stdout/stderr/exit. Both commands exit 1 solely for named blocked sources, not mismatches. Fixture 24 is now green on both backends; 11, 06, 07 and 10 remain green. 05 and 13 retain the a85a9cb1 regression. The requested later compiler fix for them was held.

Fixture 14 on both backends: Refused at 17:24, exactly undefined non-null assertion. The delivered 0555bc27 does not contain the described optional memoize signature: it preserves callback: () => T and callback = undefined!. Its ADAPTED.md says adaptation 48 was declined and source statements were unchanged. The fixture is taken exactly, without inventing a replacement adaptation. This first refusal precedes the capture-cycle check; the recount does not claim that cycle has disappeared.
```
adamic: /workspace/adamic/stage3/fixtures/host/14_getCurrentDirectory.a:17:24: Adamic 0.1 refuses a non-null assertion whose operand is exactly undefined; declare the variable optional and assign undefined
```

Final validation:

- fs_regex_memoize_adapted.log and fs_regex_memoize_original.log: all 25 observed per set on both backends, counts 20/25 and 16/25.
- Regex callback family: all selected source/native/JS oracles PASS, 0.428s, with sanitizer/leak checks and cache fingerprints reflecting final generated artifacts. The earlier callback rest refusal was corrected and independently rerun successfully; the full family then passed.
- Three field-readiness/accessor/narrowed-number oracles PASS, 25.602s. This covers the packed-cache/readiness reconciliation.
- Regex guard and source mutants plus `ADAMIC_ORACLE_WASI=1` full selected host leg: PASS, 63.344s. Native and WASI empty symlink checks/mutant pass; explicit target skips remain excluded. No POSIX call was added by the merge fixes. Earlier compiling-runtime share.c switch failure was resolved and is retained in the intermediate logs, not counted as a behavioral mutant.
- Selected compiler tests `go test ./internal/lower ./internal/fresh -run 'TestCensus|TestArrayPredicate|TestUnknownArrayPredicate' -count=1 -timeout 10m`: lower PASS 0.719s; fresh no matching tests. No full-package fresh success is claimed for that filtered command.
- Full flow comparison requested by the user: proof checkpoint 57fd72ae FAIL 109.401s; pure main 48c05d09 PASS 117.159s; pure library/merge-p2b a5d5dc9 PASS 150.472s. Same full flow command and matching cohere pin. All four top-level failure names and first diagnostic lines are in FLOW-COMPARISON.md; every failing subtest first line is in flow-failure-first-lines.json. All 42 failing source files are absent on both baselines. The comparison applies to the cited pre-regex proof, not a new full flow run after regex integration.
- Full go test ./..., full language WASI and counts regeneration were not run. All prior broader gate limits remain explicit; no main/area push or force push.

## Phantom-brand recount

Merged d90994da with the advanced census overload checks preserved. All 25 original and adapted fixtures rerun on both backends: 20/25 adapted and 16/25 pristine agree with fresh Node. 11 remains green; 24 is green. Adapted fixture 08 now refuses at 41:20:

```
adamic: /workspace/adamic/stage3/fixtures/host/08_getDirectories.a:41:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom
```

Its exact declaration is `export type Path = string & { __pathBrand: undefined; };`. Verified standalone one-line probe on both backends: `probes/08_required_undefined_brand.a` (1:20, same diagnostic). No adaptation invented. 14 remains Refused at 17:24 on both backends, before capture-cycle checking: the exact 0555bc27 tip still contains `callback = undefined!`, and ADAPTED.md says adaptation 48 was declined. 05/13 stay NotYet; no later host-blockers fix merged.

Validation: targeted lower phantom/census tests PASS 1.836s; targeted fresh phantom/fs tests PASS 0.010s; WASI host tests plus phantom review tests PASS 19.562s. Diagnostic assertions were reconciled with the newer census wording and the earlier non-void brand refusal; unsafe results, parameters, readonly removal and function-value reads still refuse. Added freshness handling for PhantomMember and a regression test. Two executed mutants: dropping the census phantom overload proof fails TestPhantomOverloadResultCastsAreErased; forgetting PhantomMember freshness fails TestPhantomMemberEvaluatesOperandWithoutEscaping. Sources restored. Full go test ./... not run; counts conflict retains both fixture families without claiming a full Linux regeneration. The prior full flow comparison is in FLOW-COMPARISON.md: main48c05d09 and librarya5d5dc9 pass, proof57fd72ae has four failing top-level tests.

Incoming six phantom runtime fixtures agree with Node on both backends and native release/sanitizer legs: focused TestNativeAgreesWithNode PASS 1.637s (see fs_phantom_oracle.log).

## Audited brands and additive unknown recount

Took the exact stage3/fixtures/host tree at daa638ed, including status.json and Node records; merged b788e96e with a merge commit. Phantom d90994da and regex a8ddb3a7 remain ancestors. a85a9cb1 remains in history, with its restrictive unknown path replaced by b788e96e; history was not rewritten. No forthcoming adaptation 48 source was invented. All 25 rerun on both backends: 22/25 adapted, 16/25 pristine. 05 and 13 recover and agree byte-for-byte with fresh Node; 11 and 24 remain green.

08 is Refused at 219:9 on both backends: `Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]`. Verified one-line source `const files: string[] = ['z', 'a']; files.sort(); console.log(files.join(','));` produces the same refusal at 1:37 on both backends. This is a compiler/language ruling, not an fs runtime gap.

25 is Refused at 579:22 on both backends: `Adamic 0.1 refuses a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold; take it as never, or constrain U to something readonly, which can't write (adamic/invariant-mutable)`. Verified one-line source `export const emptyArray: never[] = [] as never[]; export function fallback<U extends {}>(result: U[] | undefined): readonly U[] { return result ?? emptyArray; }` produces the same refusal at 1:148 on both backends. Owner: compiler.

14 remains Refused at 17:24 on both backends (exact undefined non-null assertion); awaiting the promised re-extraction. Targeted lower unknown/predicate/phantom/qualified-type tests PASS 5.366s. WASI host tests and imported unknown mutants PASS 37.984s. Mutants in_always_true and skip_inner_typeof are caught by Node stdout; the first checks native too. Test output in logs/fs_recount_daa_*.log. No new production implementation beyond merge reconciliation, no full repository gate, no macOS execution, no full counts regeneration.

Four requested flow tests: all FAIL, package time 111.071s. The old unknown-array-only refusal is absent from all lowering diagnostics. First diagnostics:

- TestEveryFunctionIsInSingleAssignment: `flow_test.go:95: Lower: /workspace/adamic/internal/oracle/testdata/non_null_deinitialize_alias.a:5:15: Adamic 0.1 refuses a non-null assertion whose operand is exactly null; declare the variable optional and assign undefined`
- TestEveryPathNodeTakesIsInTheGraph: `trace_test.go:36: Lower: /workspace/adamic/internal/oracle/testdata/non_null_literal_statement.a:2:1: Adamic 0.1 refuses a non-null assertion whose operand is exactly undefined; declare the variable optional and assign undefined`
- TestEveryMutationIsInItsRange: `ranges_test.go:28: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_union_logical.a:2:2: stage 0 can't lower a logical assignment to this target yet`
- TestLivenessHoldsOnEveryPath: `trace_test.go:390: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_union_logical.a:2:2: stage 0 can't lower a logical assignment to this target yet`

Focused Node unknown/errorCode oracle PASS 1.233s, including native sanitizer/release and JS for four fixtures. Full flow logs retained; no claim that flow is green.

## Adaptation 48 recount

Took the exact stage3/fixtures/host tree at 0d11046e. Its memoize parameter is now `callback: (() => T) | undefined` and clearing is `callback = undefined`. No compiler changes. Re-ran all 25 original and adapted fixtures on both backends: 22/25 adapted, 16/25 pristine agree byte-for-byte with fresh Node. 05/13/11/24 remain green. 08 still refuses default sort (219:9); 25 still refuses never-to-generic-U (579:22).

14 now reaches the capture-cycle check: Refused on both backends at 12:28, owner compiler/runtime graph ownership. Exact diagnostic:

```
adamic: /workspace/adamic/stage3/fixtures/host/14_getCurrentDirectory.a:12:28: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function callback() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
```

Verified one-line reproduction in probes/14_adapted_capture_cycle.a on both backends (1:28, same diagnostic). verify-14.cjs passed: fresh Node output equals the historical golden; omitting callback clearing changes stdout to true/false and is caught by exact comparison. This proves the source adaptation keeps memoization; it does not claim native execution of 14. Test logs fs_recount_0d_*.log; no full gate, flow rerun or WASI rerun for this source-only recount. Prior four flow tests remain failed as recorded in the previous recount. Only the proof branch is pushed; no main/area push, force-push or rebase.

## Whole-statement deinitialization recount

Merged compiler 553cbfe9 with a merge commit; b788e96e retained. Merge resolution preserves the existing explicit regex/process refused-fixture exclusions plus incoming refusedNonNullFixture, which omits intentionally rejected non-null forms from flow tracing while keeping standalone deinitializations eligible. No further compiler implementation. Audited fixture tree remains exactly 0d11046e.

Command: `go test ./internal/flow -run '^(TestEveryFunctionIsInSingleAssignment|TestEveryPathNodeTakesIsInTheGraph|TestEveryMutationIsInItsRange|TestLivenessHoldsOnEveryPath)$' -count=1 -timeout 15m -v`. All four FAIL; package time 116.968s. The only remaining lowering diagnostic is `stage 0 can't lower a logical assignment to this target yet`. No exactly-nullish or array-only-unknown lowering refusal remains. First diagnostics:

- TestEveryFunctionIsInSingleAssignment: `flow_test.go:95: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_array.a:39:1: stage 0 can't lower a logical assignment to this target yet`
- TestEveryPathNodeTakesIsInTheGraph: `trace_test.go:36: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_union_logical.a:2:2: stage 0 can't lower a logical assignment to this target yet`
- TestEveryMutationIsInItsRange: `ranges_test.go:28: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_union_logical.a:2:2: stage 0 can't lower a logical assignment to this target yet`
- TestLivenessHoldsOnEveryPath: `trace_test.go:390: Lower: /workspace/adamic/internal/oracle/testdata/non_null_write_union_logical.a:2:2: stage 0 can't lower a logical assignment to this target yet`

Distinct failing fixtures: non_null_write_array.a, non_null_write_field.a, non_null_write_local.a, non_null_write_logical.a, non_null_write_typed_array.a, non_null_write_union_logical.a. Owner: compiler. Verified one-line reproduction on both backends (1:71): `function update(value: number | string | boolean | undefined): void { value! ||= 7; console.log(`${value}`); } update(0);`.

Targeted lower tests PASS 1.428s; incoming deinitialization readiness and impossible-assertion oracle tests PASS 2.013s. Executed mutant dropping the standalone exemption: TestStandaloneDeinitializationUsesReadiness fails with the exact-nullish refusal; restored the source. All 25 adapted host fixtures rerun on both backends: 22/25 agree with Node; 05/13 remain green, 14 still capture-cycle Refused, 08 and 25 unchanged. Pristine count 16/25 is the immediately preceding recount, not rerun in this step. No full gate, WASI rerun, macOS run or full Linux counts regeneration. No runtime implementation changes.

## Generic empty-array recount

Merged compiler 4e022aae with a merge commit; conflicts preserve erased-unknown element refusal alongside the incoming empty-never-literal proof, and both fixture count families. Source tree remains exactly audited 0d11046e. All 25 original and adapted fixtures rerun on both backends: 22/25 adapted, 16/25 pristine agree byte-for-byte with fresh Node. 05/13/11/24 remain green; 08 and 14 unchanged.

25 advances past never-as-U but remains Refused on both backends at 628:59:

```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:628:59: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

Verified one-line probe in probes/25_overload_predicate.a on both backends, same diagnostic at 1:59. Owner: compiler.

Targeted lower empty-array and unknown-array tests PASS 0.234s. Incoming focused Node oracle and both semantic runtime mutants PASS 2.177s: empty-array-falsy and shared-fresh-number-array mutants finish with exit 0 and no sanitizer errors, caught solely by Node stdout. The incoming scalar fallback test assumed scalar OR was unsupported; this combined proof already supports it via taste-not-soundness. Reconciled its assertion to require accepted scalar truthiness and forbid the array-specific Coalesce representation. Both backends agree with Node (`7 missing 2 present`) on the scalar probe. Widening the array-only condition to all scalar OR fails that assertion; source restored. Initial direct JS probe failed to resolve the runtime module, then passed with oracle/node.mjs; the tool failure is retained in the logs. No production semantic workaround added.

No full repository gate, full counts regeneration, flow rerun or WASI rerun in this compiler-only step; no runtime calls added. Previous four flow failures remain the recorded logical-assignment blockers. Only proof branch pushed; no main/area push, force or rebase.

## Proven-predicate recount

Merged compiler a58ba402 with a merge commit; exact audited fixture tree remains 0d11046e. Required prior host/compiler merges retained. 25 passes the bodyless overloaded some predicate declaration and advances to Refused on both backends at 691:5:

```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:691:5: Adamic 0.1 refuses debugger; remove it
```

Verified one-line source `debugger;` refuses at 1:1 on both backends. It is in Debug.fail, a source adaptation/language ruling item; no debugger workaround or fixture edit invented. 08 and 14 retain their default-sort and capture-cycle refusals.

Merge reconciliation preserves packed atomic slot caches by deriving field indices from actual slot pointers, retains initialized-bit updates after every store, keeps tagged field-view metadata across spread/reuse, preserves truthiness, Array.isArray flow and census predicate marker proofs, and retains host/async registrations. No new libc/POSIX call added. Incoming checked-view/runtime code is tested by native/JS oracle checks.

Targeted lower predicate/condition/callback/empty-array/unknown-array/phantom tests PASS 13.956s, including TestPredicateOverloadRuntime. Checked cast runtime mutants and tree/object-spread oracles PASS 1.277s. Executed mutants: skip tag caught by exit, wrong tag caught by exit, twice operand caught solely by Node stdout with clean exit/sanitizers. Required-field primitive and operand-once checks PASS. Broad field-view test command FAIL 17.380s in three top-level tests, all before IR construction:

- TestNarrowedFieldUsesSharedReadiness: readiness-identifier.a:4:65, readiness-identifier-uninitialized.a:3:85 (exact undefined); readiness-number.a:4:55, readiness-number-uninitialized.a:3:71 (exact null).
- TestViewFieldInheritedStaticReadiness: non_null_static_initialized.a:1:38 (exact undefined).
- TestDefaultTaggedSourceViews: default-staged.a:4:70, default-boxed-write.a:3:79, default-read-before-set.a:3:57 (exact undefined).

Each collision says `Adamic 0.1 refuses a non-null assertion whose operand is exactly undefined` (or null), followed by `declare the variable optional and assign undefined`. These incoming initializer fixtures are incompatible with the earlier exact-nullish ruling retained in this proof; no compiler policy changed to make them pass. Other tested field-view cases pass. Full repository gate, flow rerun, full Linux counts regeneration and macOS not run in this step. Only proof branch pushed, no main/area push, force or rebase.

Final recount: all 25 adapted and all 25 pristine fixtures freshly rerun on both backends: 22/25 adapted and 16/25 pristine agree byte-for-byte with Node. Final WASI host leg PASS 18.748s, compiling the complete updated runtime archive. Logs fs_predicate_final_*.log.

## Debugger and non-null logical checkpoint

Merged 7d2cbc89 via 87f19194 and 63a202a3 via cc090e0a, both merge commits. Kept the newer truthiness/concatenation/operator and label semantics in the documentation and JavaScript dispatch; debugger now emits no native operation and is preserved by JS. Exact fixture tree remains 0d11046e; no fixture workaround. All 25 adapted and original fixtures rerun on both backends: 22/25 adapted and 16/25 pristine agree byte-for-byte with fresh Node. 05/13 remain green; 08 and 14 unchanged.

25 advances past debugger to Refused at 693:10 on both backends:

```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:693:10: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
```

Verified one-line source `const value = Error as any;` on both backends gives the same cast refusal at 1:15. Owner: adaptation (the explicit any in Debug.fail).

WASI host tests plus debugger artifact and non-null logical readiness/mutant checks PASS 91.108s. Formerly blocked non-null write/logical fixtures agree with Node on sanitized/release/native/JS oracle legs, PASS 49.944s. Always-evaluate RHS mutant finishes cleanly but prints 3 0 against Node 0 0, caught only by stdout. The WASI removed-symlink-adapter mutant also remains caught. No new runtime .c files or calls; complete WASI archive exercised.

Field-view reproducers from 2f192d71 reverified on both backends with this checkpoint compiler:

- TestNarrowedFieldUsesSharedReadiness: `interface Identifier { escapedText: string; } const node: Identifier = { escapedText: undefined! };` (1:87, exact-undefined refusal). Null variant `interface NumberNode { value: number; } const node: NumberNode = { value: null! };` (1:75, exact-null refusal).
- TestViewFieldInheritedStaticReadiness: `class State { static value: number = undefined!; }` (1:38, exact-undefined refusal).
- TestDefaultTaggedSourceViews: `const raw: { value: string | number } = { value: undefined! };` (1:50, exact-undefined refusal).

Each diagnostic is `Adamic 0.1 refuses a non-null assertion whose operand is exactly undefined` (or null), followed by `declare the variable optional and assign undefined`. These remain initializer forms; the statement-only exemption does not cover them. Complete fixture/line list is in the preceding predicate recount section. No full repository gate, macOS execution or full Linux counts regeneration in this checkpoint. Only proof branch pushed, no main/area push, force or rebase.

Final four-flow command: `go test ./internal/flow -run '^(TestEveryFunctionIsInSingleAssignment|TestEveryPathNodeTakesIsInTheGraph|TestEveryMutationIsInItsRange|TestLivenessHoldsOnEveryPath|TestDebuggerHasNoFlowInstruction)$' -count=1 -timeout 15m -v`. PASS 152.741s. All four formerly failing tests PASS; no Lower diagnostic. No refusal exclusion was added in this checkpoint.

Six debugger mutants executed and restored: native trap caught by artifact test; native trap caught by Node exit; dropping JS statement caught by preservation count; dropping lowered IR caught by preservation count; restoring syntax refusal caught by compiler test; adding a flow instruction caught by TestDebuggerHasNoFlowInstruction. All fail the intended checks, no build/clang failure. Raw logs preserved.

## Default string sort recount

Merged ce02406 with merge commits only, preserving both sides of conflict resolutions. All 25 adapted and all 25 pristine fixtures were rerun on both backends. Adapted results: native 22/25, JavaScript 23/25, joint 22/25. Original controls remain 16/25 each. Fixture 08 now compiles and agrees on JavaScript but fails natively with exit 70:

```text
adamic: panic: compiler bug: a field the checker proved is there is missing
```

The native directory predicate reads Dirent's `type` on a Stats object (which carries `_fsFileMode`). Temporary diagnostic instrumentation confirmed the field and was restored. Verified one-line native reproducer (Node prints true and exits 0):

```typescript
import { statSync } from 'node:fs'; import type { Dirent, Stats } from 'node:fs'; function isDirectory(value: Dirent | Stats): boolean { return value.isDirectory(); } console.log(String(isDirectory(statSync('.'))));
```

Owner: library, Stats/Dirent method dispatch. Fixtures 14 and 25 retain their capture-cycle and Error-as-any refusals respectively. Since the normal measurement runner stops at a runtime mismatch, this recount used a copy that records observed/expected stdout, stderr and exit as Disagrees and continues, still returning failure. It does not change the compiler or production harness. Complete JSON records preserve the mismatch.

Validation: filtered lower tests PASS (15.272s); string-sort oracle, four semantic mutants, string-sort refusals and selected WASI host tests PASS (46.855s). Mutants for code-point order, descending order, instability and undefined-as-holes were caught by Node stdout comparisons with clean execution. Focused ArrayHoles/debugger flow tests are recorded in fs_sort_flow.log. The four full flow tests were green at the previous checkpoint and were not rerun here. No full gate, counts regeneration or macOS execution is claimed. No main or area push, no force push; this remains a proof.

## Error-value and storage exemption recount

Inputs: Error branch 6469f37 via merge 6c978f67, compiler storage exemption 3a6c8231 via merge a5bbb1a0, exact fixtures/records 9eba5d10 via 8db8a198. All 25 adapted and all 25 pristine fixtures rerun on both backends. Adapted native 22/25, JavaScript 23/25, joint 22/25; pristine remains 16/25 each. Fixture 25 now gets past the removed Error cast and captureStackTrace no-op. Next diagnostic on both backends:

```text
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:1088:12: Adamic 0.1 refuses a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

Verified one-line compiler reproducer (Node prints X, exit 0; native lowering refuses at 1:190):

```typescript
function identity<T>(value: T): T { return value; } function lower(value: string): string { return value.toLowerCase(); } function choose(flag: boolean): (value: string) => string { return flag ? identity : lower; } console.log(choose(true)('X'));
```

08 retains its native Stats/Dirent dispatch disagreement and agrees on JavaScript; 14 retains its callback capture-cycle refusal. No routed host implementation was built here.

The Error merge brings nominal runtime error infrastructure. Conflict resolutions retain both readiness/NoMove flags, current compiler truthiness, host and sparse-array exception edges, Buffer/crypto/process/phantom hooks, optional fields and deinitialization. Both null representations remain recognized. The new null kind is an immutable leaf in share.c, allowing the whole native and WASI runtime archive to compile. Conflict-marker scanning accidentally touched TypeScript's deliberately malformed baseline fixtures; those changes were restored before finishing, and the submodules remain clean. Counts preserve unique rows from both sides; no complete regeneration claimed.

Validation:
- Filtered lower Error/nullish/deinitialization tests PASS, 2.907s.
- Error probe, typeof, debug_fail and stack refusal oracles PASS, 1.503s, including native sanitizers/release and JavaScript.
- Configured WASI host refusal tests and empty-symlink oracle PASS. Removed-symlink-adapter mutant caught by Node output.
- Four storage readiness mutants PASS: dropping local/field/static/identifier checks changes observed output and exits 0, caught by the pinned runtime comparison.
- Selected oracle package command exits 1: TestNonNullStorageUnionRemainsRefused expected a refusal but lowering returns nil on this combined proof. Initial wider run also fails TestLibraryErrorCounts on the historical library_error_original_probe.a cast; exact diagnostics are in logs. These are reported failures, not green gates. The initial WASI invocation lacked WASI_SYSROOT; cloud/setup.sh --wasi-sdk supplied it and the WASI tests were rerun successfully.
- Setup timing: node 0.021s, Go 0.022s, submodules 0.051s, markdown 0.066s, clang 0.159s, WASI SDK 0.176s; nproc 5. Initial setup overlapped conflict resolution and its warm build failed on unresolved markers; final SDK setup completed.
- No full gate, four-full-flow rerun, counts regeneration or macOS execution. Proof branch only; no main/area or force push.

## Stats/Dirent runtime dispatch merge and recount

Merged 70a2a97 via f72489a6, preserving existing nominal error/null-sentinel behavior and both fixture families. All 25 adapted and pristine sources rerun on both backends. Final adapted joint count is 17/25 (native 17, JavaScript 17). The merge imports optional-widening checks and the Node option absence proof; the proof excludes const assertions. Consequently the existing shared `{ throwIfNoEntry: false } as const` options stop fixtures 06,07,08,10,13,24 before lowering, with the optional bigint refusal. The regression begins at this merge, not at the preceding Error/storage recount. Fixture 08 no longer reaches the old layout panic because it now stops earlier; the isolated union oracle proves the dispatch fix. No global optional-widening exemption was added to hide these refusals.

Verified one-line regression reproducer (Node prints true, exit 0; Adamic refuses at 1:120):

```typescript
import * as fs from 'node:fs'; const options = { throwIfNoEntry: false } as const; console.log(String(fs.statSync('.', options)?.isDirectory()));
```

Exact diagnostic:

```text
adamic: /tmp/fs_union_options.a:1:120: Adamic 0.1 refuses optional property bigint in StatSyncOptions & { bigint?: false | undefined; throwIfNoEntry: false; } absent from structural source { readonly throwIfNoEntry: false; }, which can hide fields; declare bigint on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

Owner: compiler/library Node-options absence-proof boundary. The existing exemption intentionally rejects casts and thus misses the fixture's const assertion. Fixtures 14 and 25 retain their capture-cycle and generic identity callback refusals.

Validation: ADAMIC_ORACLE_WASI=1 filtered union/host tests PASS, 29.728s. Native static first-member-dispatch mutant compiles and executes then panics on Stats; Node catches it. Special kinds match Node on native and JavaScript; WASI honestly refuses ambiguous FIFO/socket kinds, and removing that guard is caught only by Node stdout with exit 0. WASI host refusals and empty-symlink error oracle pass. No full gate or full flow rerun; counts retain unique rows without a regeneration claim. stage3 host fixtures remain exactly 9eba5d10. Proof branch only, merge commits without rebasing/force/main/area pushes.

## Fixture 14 with graph regions and memoize

Authorized runtime area 440e3cdb8 merged via 57e73215, followed by compiler memoize 53e2a82b. Fixture 14 alone was measured with the unchanged 9eba5d10 host set: native Compiles/Agrees, JavaScript Compiles/Agrees, exact Node stdout/stderr/exit. Full 25-fixture JSON/table above remains the preceding recount; no new all-fixture total is claimed from this single measurement. The const-options fix remains pending.

Graph ownership and frame environments were merged while retaining host nodes, sparse arrays, packed caches, nominal errors, current truthiness and checked-storage readiness. Runtime's ordinary async IR replaces its older AsyncProgram architecture: compile errors identified stale Async/hasAsync/lowerAsync references, which were removed in favor of Generated/AsyncEntry and ordinary async functions. Header retains typed-array, null and environment kinds and sparse-array APIs. Region object allocation preserves readiness/type metadata. Closure invocation retains the existing host argument-count ABI. Both fixture/count families remain; count regeneration is not claimed. Automatic approval review rejected broad scripted conflict resolution, so reviewed small patches and byte-for-byte equivalence checks were used instead.

Tests: ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestMemoizeRegions$|^TestMemoizeEnvironmentOmissionLeaks$|^TestWASIHostRuntimeRefusals$' -count=1 -timeout 15m -v PASS, 48.472s. Memoize a/b/host14/self match Node on JavaScript, sanitized native and release; graph counts and leak checks pass. Omitting graph environment adoption is caught by LeakSanitizer (251 bytes in three allocations), proving the ownership check can fail. WASI host refusals pass, building the whole runtime archive.

Focused package command: go test ./internal/lower ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript -run 'TestMemoize|TestArrayHolesExceptionEdges|TestDebuggerHasNoFlowInstruction' -count=1. Flow passes, freshness/IR compile with no matching tests, JavaScript has no tests. Lower exits 1 solely because incoming TestMemoizeCaptureStopsAtAssertion expects the historical blanket non-null assertion refusal and receives nil with this proof's non-null/storage support; TestMemoizeSelfCaptureUsesGraph passes. This failure is reported and not waived. Full gate, full flow suite, full host recount and macOS execution were not run. Proof branch only; no main/area/force push or rebase.

## Updated memoize input 3550dbc4

The updated compiler memoize branch 3550dbc4 was merged to supersede the earlier runtime/memoize tips. `git merge-base --is-ancestor d70bc6b1 3550dbc4` exits 0: the compiler tip already carries the requested runtime area, so no separate area/runtime merge is necessary. History was preserved, without rebase or force push. Only count metadata conflicted; unique rows from both sides remain. No runtime or compiler source conflict arose in this update.

Fixture 14 was remeasured against fresh Node execution: native Compiles and Agrees; JavaScript Compiles and Agrees. The dedicated JSON record is fixture-14-regions-status.json. All 25 were not rerun here; the const-options fix is still pending and the preceding all-fixture count remains historical.

ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestMemoizeRegions$|^TestMemoizeEnvironmentOmissionLeaks$|^TestWASIHostRuntimeRefusals$' -count=1 -timeout 15m -v PASS, 1.986s. Unchanged positive native/Node observations were served by the oracle cache (native 12 hits, Node 9 hits); the ownership-omission mutant was rerun and caught. The earlier stale TestMemoizeCaptureStopsAtAssertion expectation remains present in 3550dbc4; it was not rerun because its source and relevant compiler behavior are unchanged. No full gate, count regeneration or macOS test claimed. Only codex/host-proof-combined is pushed.

## Full const-options recount

Inputs: const-options 02a997b6 via merge 630702f2; optional-widening parentheses e52e6219 via eb5f02bb. All requested commits 3550dbc4, 6469f37 and 3a6c8231 are verified ancestors. Host fixtures and Node records are still byte-identical to 9eba5d10. Both sets of all 25 were rerun on native and JavaScript after the final declaration repair. Adapted: 24/25 each and jointly. Pristine: 16/25 each; original 14 reaches the loud non-null assertion and disagrees with Node, so it is not counted green.

06,07,08,10,13,24 compile and agree again; 08's runtime union dispatch now also agrees natively. 14's graph-region memoize agrees on both backends. The sole adapted stop is compiler-owned 25 at 1088:12:

```text
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:1088:12: Adamic 0.1 refuses a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

Verified compiler reproducer (Node X, exit 0; Adamic refusal at 1:190), retained from the earlier recount:

```typescript
function identity<T>(value: T): T { return value; } function lower(value: string): string { return value.toLowerCase(); } function choose(flag: boolean): (value: string) => string { return flag ? identity : lower; } console.log(choose(true)('X'));
```

Integration repair f43d1b2b1ea94fa143cc668ee836c77f2c808c7a: the runtime area's uninitialized-declaration path had been merged for every local, although it is for captured/preallocated storage. Plain BOM byte-swap temporaries were therefore never declared. Both emitters now use that early path only for Captured or Preallocated locals, preserving ordinary typed storage/readiness and graph preallocation. The self-cycle memoize regression test verifies there is no duplicate result declaration. Three fs input fixtures also had auto-merged byte-identical duplicate errorCode helpers; the duplicates were removed without changing Node behavior. The two parentheses count rows were regenerated on Linux: existing allocation counts unchanged, two zero graph-count columns added. Whole-table regeneration is not claimed.

Validation:
- Full adapted/pristine host measurement commands return 1 solely for recorded blocked/disagreeing fixtures; all rows are recorded, including exact stdout/stderr/exit for runtime discrepancies.
- Filtered TestNodeHost and parenthesized-relation lowering tests PASS, 17.476s. The fresh const proof admits consuming uses and refuses hidden/mutable/escaping shapes.
- Final ADAMIC_ORACLE_WASI=1 selected oracle command exits 1, 24.430s. WASI runtime archive compiles. Runtime host target refusals and empty-symlink comparison PASS. NodeFSDirectoryUnion tests and static-dispatch/lossy-kind mutants PASS. Memoize a/b/host14/self, all four storage-readiness mutants and parentheses Node/sanitizer/release/JS comparisons and refreshed counts PASS.
- WASI host comparison failures below remain open. Two fixtures refuse Error-to-ErrnoException optional widening; others omit error-printing output. A verified minimal native host identity probe returns false while Node and generated JS return true for `error instanceof Error`, which explains the missing output for such guarded catch paths. This is a library/runtime host error identity contract gap; the adapted tsc errorCode path still passes.

```typescript
import { readFileSync } from 'node:fs'; try { readFileSync('/adamic-deliberately-missing-file'); } catch (error) { console.log(String(error instanceof Error)); }
```

Final failing host tests:

```text
--- FAIL: TestWASIInputAgreesWithNode (9.91s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_process_performance_core.a (0.67s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_symlink.a (0.23s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_entries.a (0.62s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_realpath.a (0.55s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_stat_options.a (0.24s)
--- FAIL: TestWASIInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_permissions.a (0.72s)
--- FAIL: TestWASIFileAgreesWithNode (9.51s)
--- FAIL: TestWASIFileAgreesWithNode/read (0.62s)
--- FAIL: TestWASIFileAgreesWithNode/write (0.52s)
--- FAIL: TestWASIFileAgreesWithNode/close (0.49s)
--- FAIL: TestWASIFileAgreesWithNode/unlink (0.62s)
--- FAIL: TestWASIFileAgreesWithNode/buffer (0.67s)
--- FAIL: TestWASIFileAgreesWithNode/read_sync (0.53s)
--- FAIL: TestWASIFileAgreesWithNode/rm (0.65s)
```

The Error-to-ErrnoException refusal occurs at node_fs_directory_symlink.a:5:42 and node_fs_directory_stat_options.a:19:46: optional errno absent from Error (adamic/no-optional-widening). Full exact output differences are preserved in the compressed WASI log. No full gate, full flow suite, full counts regeneration or macOS execution. Only codex/host-proof-combined is pushed, without rebase, force, main or area push.


## Generic value and WASI Error identity recount

Merged 21e50844 with merge commit 318553b0 and 6054117 with merge commit bb526abf. The expression conflict retains Promise/optional-intrinsic hooks and adds generic-function identity checks. The obsolete generic-value refusal test now checks an uncontextualized generic value; unrelated refusals already superseded by this proof remain removed.

All adapted fixtures: 24/25 native, 24/25 JavaScript. All pristine controls: 16/25 on each backend. Adapted fixture 25 remains Refused on both backends at 1088:12:

```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:1088:12: Adamic 0.1 refuses a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

Verified one-line reproducer (C lowering refuses at 1:190):

```typescript
function identity<T>(value: T): T { return value; } function lower(value: string): string { return value.toLowerCase(); } function choose(flag: boolean): (value: string) => string { return flag ? identity : lower; } console.log(choose(true)('X'));
```

Commands and results:
- `go build -o /tmp/fs-combined-adamic ./cmd/adamic`: PASS.
- `observe-disagreements.py --all` and pristine snapshot runner: complete all 25; exit 1 because the recorded refusals remain.
- `go test ./internal/lower -run '^TestGenericFunctionValue' -count=1 -v`: PASS, 0.425s.
- `go test ./internal/oracle -run '^TestGenericFunctionValueWrongResultMutant$' -count=1 -timeout 5m -v`: PASS, 19.094s. Wrong numeric specialization returns -1; Node stdout comparison catches it on both backends.
- `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASI(InputAgreesWithNode|HostRuntimeRefusals|FileAgreesWithNode|EmptySymlinkAgreesWithNode|HostErrorIdentity)$|^TestNodeFSDirectoryUnion|^TestOptionalWideningNodeParentheses$|^TestNonNullStorageReadinessMutants$|^TestMemoizeRegions$' -count=1 -timeout 15m -v`: PASS, 65.409s. Honest target refusals stay explicit; these are not claimed as successful WASI filesystem executions.
- Host Error tag-dropping mutant caught by Node stdout; empty-symlink, static Stats/Dirent dispatch, missing WASI kind guard and storage-readiness mutants also pass their expected failure checks.

Raw logs are compressed in logs/fs_last_*.log.gz. No full `go test ./...`, full counts regeneration or macOS gate was run. This remains a proof branch, never a main or area landing; no force push or rebase.


## Conditional generic instantiation recount

Merged afdb4c94 as d6149508. Invariance conflicts preserve the census marker-predicate rule and readonly-array contextual targets alongside contextual generic signatures.

Adapted: 24/25 on each backend; original controls: 16/25 on each backend. Fixture 25 reaches this new compiler-owned first refusal:

```
adamic: /workspace/adamic/stage3/fixtures/host/25_readDirectory.a:1097:26: Adamic 0.1 refuses a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold; take it as NonNullable<T>, or constrain T to something readonly, which can't write (adamic/invariant-mutable)
```

Verified on both backends at 1:325:

```typescript
function isArray(value: unknown): value is readonly unknown[] { return Array.isArray(value); } export function flatten<T>(array: T[][] | readonly (T | readonly T[] | undefined)[]): T[] { const result = []; for (let i = 0; i < array.length; i++) { const v = array[i]; if (v) { if (!isArray(v)) { result.push(v); } } } return result; }
```

Commands: compiler build PASS; both all-25 observation runners complete (exit 1 for remaining refusals); focused `go test ./internal/lower -run '^TestGenericFunctionValue' -count=1 -v` PASS (0.631s). WASI command is the previous complete host selection plus `^TestGenericFunctionValueWrongResultMutant$`, with `ADAMIC_ORACLE_WASI=1`, `-count=1 -timeout 15m -v`: PASS (33.712s). Host Error tag-dropping and generic wrong-result mutants are caught by Node comparison; other selected host and storage mutants pass. Honest WASI target refusals remain explicit. Raw logs: logs/fs_conditional_*.log.gz. No full repository gate or macOS run. Proof branch only, no rebase or force push.
