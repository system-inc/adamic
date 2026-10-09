TestStringIndexCacheStatesMatchNode is defended within the five-row direct-call matrix.
D1 removes the literal-sentinel condition; only the target fails, with ASan global-buffer-overflow.
Production and tests are restored; full-package uniqueness remains unknown outside this bounded matrix.

Starting origin/main: 7d113268b1903e4e47289f226e94ec2fb609e15b.
Audit base: 0942c5169d0ea736d9dfaa19881af1ad8adad162.
Audit REPORT.md, scope.md, results.json and menu.json are retained as audit-* files.

Code under test: internal/native/runtime/string_index.c and the C string search operations calling it. Oracle: executed Node indexOf and charCodeAt, comparing complete stdout, plus ASan memory checks. No oracle, harness or test was mutated.

Semantic difference: the target runs indexOf on a long mixed-Unicode literal before a length or character query. It also copies that literal onto the stack, clears its index, and exercises a heap string. The subsumer builds heap strings, queries lengths, reads characters and slices, then searches. D1 targets units_before's sentinel handling, shared production code with a distinct initial state. Removing its literal condition dereferences the one-byte adamic_literal_mark as an index. Node produced its expected answer before the native process ran.

Coverage commands:
  timeout 120 go test -json -count=1 -timeout 90s -coverpkg ./internal/native -coverprofile review/test-defend/internal-native-split_units/target.cover ./internal/native/ -run '^TestStringIndexCacheStatesMatchNode$' > target-coverage.log 2>&1
  timeout 120 go test -json -count=1 -timeout 90s -coverpkg ./internal/native -coverprofile review/test-defend/internal-native-split_units/subsumer.cover ./internal/native/ -run '^TestStringIndexMatchesNode$' > subsumer-coverage.log 2>&1
The five exclusive Go coverage blocks are in coverage-exclusive.txt. They belong to cString, source preparation, and were not mutated. Go coverage does not instrument the compiled C runtime. The defense therefore rests on the semantic input history, not exclusive Go preparation lines.

Matrix selection: rg -l 'indexOf|lastIndexOf|adamic_string_units_before' internal/native --glob '*test.go' finds string_test.go, string_index_test.go, string_build_test.go. All five top-level tests in those files are included, from the current commit. No new direct search caller files were found. units_before is called by index_of_at and last_index_of in string_search_impl.h. Compiler-generated runtime callers in other test files may also reach these operations and remain unknown.

Full clean baseline exceeded 90 seconds, with no observed assertion failures before timeout. It is an over-budget baseline, not a red assertion baseline. The five-row baseline passed in 8.290 binary seconds. Coverage target: 0.714 binary seconds; subsumer: 10.079. D1 matrix: 14.178 binary seconds, 22.502 wall seconds including Go compilation and runtime rebuild. Restored five-row baseline passed in 6.975 binary seconds. Setup reused env.sh, nproc 5. npm ci succeeded. Separate runtime build duration was not measured.

Standalone D1.diff applies to the starting commit and passed clang -std=c11 -Wall -Wextra -Werror -pedantic -fsyntax-only -Iinternal/native/runtime internal/native/runtime/string_index.c. The native test build also completed successfully. Each mutant command sets its own ADAMIC_BUILD_CACHE_DIR. Runtime archives actually use os.UserCacheDir with a content hash over source, compiler and flags; the mutant rebuilt a changed archive, as confirmed by the ASan stack.

Friction and limits:
- The user supplied no failing audit line. The fetched report provides it and shows that subsumption rests on one mutant.
- /tmp has a total capacity of 8.8 GB, so 15 GB free is impossible. Earlier-unit /tmp/oracle-defense was removed; 5.2 GB remained free. /workspace reported 15 GB free. No disk-full failure occurred.
- Full native suite is too large for the requested 90-second budget. Uniqueness is proven only within the explicitly listed direct-call matrix, not the package or repository.
- Go coverage observes build preparation, not runtime C branches. Semantic state differences supply the lead.
- Initial npm redirection used the wrong working directory and failed before running npm. It was corrected, npm succeeded before baseline.
- One report read exceeded the output limit; relevant source and report portions were reread, and whole report copies are retained.
- The environment refreshed during final evidence work and removed the old polling tool. The completed restored log was read directly.

Name/assertion finding: CacheStatesMatchNode promises semantic agreement across cache states, which its assertions check. It asserts no timing or instruction threshold and makes no performance guarantee. One unique bounded mutant suffices to keep the row; no additional mutations or test rewrites were needed.
