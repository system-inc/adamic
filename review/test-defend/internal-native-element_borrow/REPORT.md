Both rows are defended within the complete 30-row bounded matrix, each with 29 other passing rows.
Starting commit: 7d113268b1903e4e47289f226e94ec2fb609e15b. nproc: 5. No tests or production source remain changed.
Whole-package baseline timed out at 90.076 seconds; bounded baseline passed in 9.015 seconds without skips.

Code under test and oracle

TestSizeClassesShareTheirChunks: runtime/heap.c allocation and cross-class chunk recycling. Reached production chain: adamic_string_allocate, adamic_allocate, take, new_chunk, list_chunk, unlist_chunk, adamic_release, release_last, list, free_one, deallocate, give. Oracle is self: correct string contents, acquired chunks <= largest generation + 16, and churn/control RSS <= 1.5. The new absolute chunk-count bound was added after the audit at bc8b353b. It catches the audit's former M10 survivor. D1 disables only spare reuse, retaining working allocation and string bytes. Clean control/churn acquired 79/85 chunks; D1 acquired 1171/769, both with largest generation 79. String-content checks and the other 29 matrix rows pass. This is the required answer-preserving resource-cost attempt.

TestCEndsInNewline: C/cProgram and expression emission, particularly emitter.value's ArrayReverse branch. Oracle is self: final newline and, for its 6000-expression fixture, output length greater than 256 KiB. Compared with TestRegexProgramsKeepCheckedFieldReads, it reaches 16 exclusive Go coverage blocks, including emit_expressions.go:412-414. D5 returns NULL early instead of emitting the array operand and reverse call. Only this row fails, at library_array_test.go:23, because output becomes too small. Its small subcase still passes. This proves unique protection for the additional size assertion, not unique protection for the final newline byte. D2 removes main's final newline and D4 removes the translation unit's final byte; strict clang builds catch both too. D3 drops completed bodies and is also caught by source-shape and compilation rows.

Coverage and scope

Raw per-test profiles use -coverpkg=./internal/native. TestSizeClassesShareTheirChunks has no exclusive Go blocks against the other 29 bounded rows. Its semantic difference is the 15-generation allocation/release history, across multiple classes versus one class, and its absolute chunk-count bound. Go coverage does not instrument embedded C; it proves the build path, not runtime/heap.c branch coverage. No claim of C coverage is made. Source search found no other native top-level test directly naming ArrayReverse or .reverse(). Reached-function selection remains conservative, not an exhaustive transitive proof.

matrix-rows.json lists every matrix row; excluded-rows.json lists every current package row outside the bounded matrix. No outside-set kill is known. The package list was freshly obtained, and the matrix includes current runtime/cache/string rows beyond the old 14-row audit slice. A defense here is a unique observed kill within the bounded matrix, not established package-wide uniqueness.

Commands and validation

Warm /workspace/adamic-tools/env.sh used; no setup rerun. npm ci --prefix stage3/api ran before baseline. Whole baseline: timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ > baseline.log 2>&1. No assertion failure was observed before its timeout. Bounded baseline and matrices use the exact regex in scope.regex. Each matrix sets ADAMIC_BUILD_CACHE_DIR=/tmp/defend-element/cache/Dn. D1 physically changed the embedded C source during its run and was restored afterward. D2-D5 use Go overlays; their standalone diffs contain no selector or overlay dependency.

Every Dn.diff passed git apply --check against the unchanged starting tree. Each Go mutation passed go vet with its corresponding overlay. D1 passed clang using the exact release Flags(Options{}) from native.go and passed go vet; its native build and execution also succeeded, apart from the expected resource assertion. Runtime archive caching is content-keyed under os.UserCacheDir(), not ADAMIC_BUILD_CACHE_DIR, so changed embedded bytes forced a fresh runtime archive. No stale product reuse was assumed.

matrix.json includes complete pass/fail lists, no skips and no unknown outcomes within each matrix. D1 took 22.270 seconds, D2 9.318, D3 8.891, D4 8.542, D5 9.028. Coverage comparison over the other 29 rows passed in 11.935 seconds. Runtime rebuild time is included in D1; it was not separately measured.

Brief ambiguities and costs

- /tmp has only 8.8 GB total, so the requested 15 GB free floor cannot be met there. Workspace free space started at 7.2 GB. Only the identifiable previous /tmp/defend-checked-views directory (40 KB) was removed. Unidentified shared directories were left intact. Final free space was 3.0 GB in /tmp and 6.9 GB in /workspace. No command failed for disk exhaustion.
- The audit refers to an older test. The chunk-sharing test now has an absolute chunk-count bound that addresses the audit's documented RSS-ratio weakness. test-changes-since-audit.diff records this without modifying any test.
- The full package cannot finish within the brief's binary budget. All unique-kill claims are explicitly bounded; results outside the 30 listed rows remain unknown. This is a limit on the evidence, not permission to delete tests.
- There were four C-emission experiments, exceeding the requested three by one. D4 was an alternate end-byte encoding of the already caught D2 behavior. It is retained transparently. D5 was the coverage-led semantic difference that ultimately defended the row. No verdict rests on the duplicate encoding.
- Go -coverpkg profiles cannot instrument runtime C. The heap defense therefore rests on the documented input history and a direct production C mutation, not invented exclusive C lines.
- TestCEndsInNewline's successful defense is for its size assertion. The name understates that extra check; it does assert the newline it promises. TestSizeClassesShareTheirChunks now explicitly asserts chunk reuse as its name promises. Neither row is being deleted, weakened, or rewritten.
- An initial D4-generation script matched two return statements and stopped at its assertion. It made no production change; the corrected script matched the cProgram return specifically. No failed generation was counted as a mutant result.

Artifacts

rows.json: structured defense report, exact commands and failing lines.
matrix.json: complete outcome and passed-row lists.
D1.diff-D5.diff: standalone production mutants.
Dn.log and Dn-vet.log: raw matrix and validation logs; D1-clang.log: direct runtime compilation.
*.cover and *-exclusive.txt: per-row/comparison coverage evidence.
provenance.json, tests-list.log, baseline*.log, npm-ci.log: starting state and baseline evidence.

No repo-wide replay, full package completion, C coverage, or exhaustive native reachability was obtained. Production sources were restored and no test was edited.
