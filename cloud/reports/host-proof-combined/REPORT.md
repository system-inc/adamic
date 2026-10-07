# Combined host proof

21/25 adapted fixtures and 15/25 pristine fixtures execute and agree with fresh Node observations on both backends. Every fixture was rerun on each backend after the final merges. This is a proof, not a landing; no main/area push or force push.

Final additions: Buffer fallback ccb8a69 via merge 469bb8b9; exact audited adapted fixtures/status/Node records from 21ef072e via 33e60869; Date conversion 05635aa via merge 0e7a2cc3; scalar concatenation 998fb3eb via merge b78d7ea7. stage3/fixtures/host is identical to the adapted branch at 17c5385a. The pristine run uses the unchanged pre-adaptation fixture snapshot at 08b5b2c4. Agrees requires exact stdout, stderr and exit against Node, including exit 1 and 2 fixtures.

| Fixture | Pristine native | Pristine JS | Adapted native | Adapted JS | First adapted blocker | Owner |
|---|---|---|---|---|---|---|
| 01_readFile_utf8.a | Checker | Checker | Agrees | Agrees | None | None |
| 02_readFile_utf16le.a | Checker | Checker | Agrees | Agrees | None | None |
| 03_readFile_utf16be.a | Checker | Checker | Agrees | Agrees | None | None |
| 04_readFile_missing.a | Checker | Checker | Agrees | Agrees | None | None |
| 05_writeFile.a | Refused | Refused | Agrees | Agrees | None | None |
| 06_fileExists.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 07_directoryExists.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 08_getDirectories.a | NotYet | NotYet | Refused | Refused | ensureTrailingDirectorySeparator overload Path/string result at 160:1 | Compiler / adaptation ruling |
| 09_realpath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 10_getModifiedTime.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 11_setModifiedTime.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 12_deleteFile.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 13_createDirectory.a | Checker | Checker | Agrees | Agrees | None | None |
| 14_getCurrentDirectory.a | Refused | Refused | Refused | Refused | memoize callback capture cycle at 12:28 | Compiler / runtime |
| 15_getExecutingFilePath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 16_getEnvironmentVariable.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 17_write.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 18_exit_0.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 19_exit_1.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 20_exit_2.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 21_createHash.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 22_createHash_fallback.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 23_newLine.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | NotYet | NotYet | regex replacement callback at 56:12 | Library |
| 25_readDirectory.a | Checker | Checker | Refused | Refused | isArray predicate return not proven at 530:5, adamic/no-type-predicate | Compiler |

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
