Friction and limits:
- The audit labels were bounded, not package-wide. The current package has 282 Test functions; all were enumerated. The current exact step-limit boundary row was added to regexp matrices.
- The full clean run timed out at90.168 seconds with no observed individual failure. A broader direct-runtime baseline passed; D1's first broad run cooked at90.060 seconds. Its failures do not prove uniqueness and are retained separately. The subsequent per-function matrices completed or are explicitly flagged.
- Go -coverprofile with -coverpkg=./internal/native,./internal/regexp measures compiler and build code, not embedded C run in child binaries. All requested row/subsumer profiles are saved, with exact exclusive Go lines. No C coverage claim is made. Runtime leads use semantic inputs and assertions instead.
- Seven mutation slots cannot supply three dedicated attempts for ten rows. They were spent on invalid-radix diagnostics, undefined equality, three step-limit contracts, raw UTF16 pattern identity and low-surrogate lastIndex. Rows with fewer than three dedicated attempts are cannot-judge, not not defended. They remain candidates to keep.
- Some rows differ in input despite sharing a checker. The random wrappers are one family. A family member is never treated as its subsumer.
- The supplied radix audit excerpt truncates commands. Complete prior report, row data, scopes, plans and limits were fetched and saved.
- Every complete matrix lists its current test names and passed rows. Decoder corpora, unrelated compiler products, WASI and other emitted-program consumers were not rerun under mutants. No package-wide unique kill is asserted beyond the bounded direct-consumer scope.
Name/assertion findings for rows not defended:
- TestRecordBenchmark measures and logs timing, validates five nonnegative non-NaN durations, and compares Node workload results. It asserts no performance threshold and accepts Infinity. Its name does not establish protection against a slowdown.
- TestRegExpNativeStepLimit checks catastrophic-backtracking interruption, exit70 and a diagnostic substring. It does not check exact instruction count. The newer boundary row guards that separate promise.
- Search, Lint, Test262 and Random compare results for their own supplied inputs, including captures/groups/lastIndex. No missing name promise was established; insufficient mutants cannot justify weakening them.
- RuntimeReleasePaths checks shared ownership, live counts and a100000-object chain with two sanitizer settings. Its assertions do address the named release behavior. No exclusive production defect was tested within the budget.
No test or oracle was edited. All source edits are restored before committing evidence.
