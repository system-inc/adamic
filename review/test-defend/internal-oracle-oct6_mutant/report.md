One requested row is defended within the bounded reached-row matrix.
The other two were not defended after three aimed production attempts each.
All production edits were restored; no test, oracle or harness was changed.

## Base, scope and baseline

Start: origin/main e77a4ae41f473c149aee910c51b73637686a804a; audit commit 9e378707ce011528e61eaf9c01250bd735d62012. The audit scope.md, REPORT.md, results.json and menu.json were read from the explicitly fetched full remote ref. Their subsumption rests solely on M4, a general Lower entry rejection. Current go test -list inventory has 198 top-level names, including all three requested names. They remain in omitted_arguments_test.go, oracle_test.go and output_test.go.

Warm env.sh worked; setup skipped. npm ci in stage3/api ran before the baseline. The whole clean package exhausted 90 seconds (90.199 binary seconds) without any completed failing test; 125 top-level rows completed successfully. This is an oversized package, not a red code baseline. Clean per-row compiler coverage passed for original, file order, prompt, and six bounded NativeAgrees fixtures. Restored comparison controls passed for the union of all selected top-level rows and all selected relevant NativeAgrees fixtures. Exact selections, observed results and uncompleted rows are saved in runs.json, controls.json and matrix.json. New policy, argument-length, Node filesystem and math regression rows were included when relevant.

## Code, oracle and coverage leads

See code-and-oracle.md. Code is Adamic Lower, native argument emission and C runtime; truth comes from unmodified Node source observations, explicit expected strings, and sanitizers/leaks. No source fixture or comparison was edited. Per-test Go coverage used -coverpkg=./internal/lower,./internal/native. The NativeAgrees profile is explicitly bounded to six fixtures because its complete row is too large. Original versus bounded Native has 24 exclusive blocks, file order versus its audit subsumer prompt has 21, and bounded Native versus original has 1235. Exact blocks and function profiles are saved. These are leads, not uniqueness claims. Embedded runtime C is not instrumented by Go coverage, so C leads use function bodies and source-call inventories.

Original's distinct source is preserved .ts with start! in a conditional arm skipped for omitted length; Node prints 11. Attempts touch source-extension admission, scalar absence, and conditional branch selection. D4/D5/D6 each fail this row, but other current rows and NativeAgrees catch each. Newly introduced non-null policy rows are real catches beyond the original audit matrix.

File order puts both streams on one pipe and writes a file alias between logs. D1 removes its pre-write flush and exposes second/first/third rather than first/second/third, but NativeAgrees' stdout-alias fixture catches it too. D2 changes truncate to append: target passes and InputAgrees rejects the changed filesystem behavior. D3 omits the last nonempty file-write byte: both target and NativeAgrees reject it, alongside InputAgrees. Thus all three attempts give no unique catch.

NativeAgrees' navigation fixture uses atan2 with negative x, including signed zero y. D7 changes this function's pi constant to 3.0. Only NativeAgrees fails among completed substantive selected rows, at oracle_test.go:774: stdout differs. Its method_coverage_math fixture passes. D7-rows.log records all passing controls, including both other requested rows and current math/method regression rows. atan2-source-reach.txt identifies navigation and method_coverage_math as the source consumers; both are registered in the main fixture row and allocation-count table, and potentially WASI execution. Defended here means uniqueness in this bounded reached-row matrix, not a completed whole-package proof. WASI execution/emission remain skipped behind opt-in with no configured WASI_SYSROOT, so those outcomes are unknown.

## Allocation-count selector artifact

The D7 allocation-count probe selected only navigation and method_coverage_math. Both fixture subtests pass. The parent reconstructs the entire table, so selecting two entries necessarily leaves the others absent and fails counts_test.go:221 with no recorded/measured row difference. The same selector on restored clean source reproduces that failure, and both fixtures still pass. Neither log reports any changed measured row. This parent failure is not a mutation kill and is excluded from rows_failed; all raw events are retained. The full allocation-count row remains unmeasured under D7. No harness was altered to make filtering work.

## Mutants and verification

Seven mutants were planned and saved before execution, within the whole-menu choices: drop statement, change constant/option, off-by-one bound and swap arguments. D3 preserves the zero-length write case to avoid unsigned underflow; it omits exactly one byte only for nonempty writes. Runtime mutants were checked with the native sanitizer clang flags in runs.json and rebuilt as actual native products. Go mutants passed go vet for their modified package. Each has its own ADAMIC_BUILD_CACHE_DIR. Every standalone diff applies to the pinned starting commit and contains no selector switch. matrix.json preserves every raw failure including witness preparation failures, which are not counted as evidence that their guarded check fails. No repository-wide replay was run.

## Owner findings and brief friction

* OriginalProbePolicy does verify that the supplied .ts can lower and execute with Node's omitted-argument result. It does not compare the source text against an original-authority pin despite its preserve-verbatim comment. No broader name/assertion mismatch was demonstrated.
* FileWritesLandInNodesOrder directly asserts the promised combined-stream order for two alias cases. No name/assertion mismatch was found. Its shared catches do not imply that its stream protocol is unnecessary.
* The brief omitted named subsumers and failing commands for these rows. The audit supplies multiple mutual subsumers, all based on one generic entry mutant. We used actual named audit subsumers for coverage comparisons and named observed catchers for the defense result.
* The 15 GB free requirement cannot fit /tmp's entire 8.8 GB filesystem. Earlier-unit scratch had already been cleaned, leaving only about 12 MB used initially, and no substantial old unit directories remained. /workspace started with 12 GB free. Neither repository nor tools was removed. No test failed from disk exhaustion.
* The complete package is over budget; native row profiles/matrices are bounded. Other fixtures and unrelated current rows have unknown outcomes, rather than inferred passes.
* Go profiles cannot measure runtime C execution. Source reach and shared-line semantic differences supplied the C leads.
* Filtering TestCountsAreRecorded does not filter its final whole-table comparison. The clean reproduction resolves that misleading failure and cost another control run. Full-table mutation sensitivity remains unknown.
* Optional WASI rows are not exercised; no SDK was configured. Their skips are not passes, and the native defense is conditional on the observed scope.

Timing: nproc 5, warm setup skipped. Full baseline 90.199 binary seconds; coverage runs wall total 17.28 seconds; mutant group runs wall total 237.61 seconds. Controls and compile commands/timings are retained separately. Session approximately 18 minutes. No test was deleted, rewritten or weakened; no main push or PR.
