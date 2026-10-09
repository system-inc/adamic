Both assigned rows are defended in the completed 28-row bounded matrix.
D1 uniquely catches TestInputAgreesWithNode; D2 uniquely catches TestImportCycleLoadTimeReads.
Production and tests are restored; no deletion or rewrite is recommended.

Code under test: Adamic lowering, native C and JavaScript emission, plus the native input runtime. The actual mutated functions are adamic_write_text_file in internal/native/runtime/input.c and lowering.staticDeclaration in internal/lower/class_static.go. Neither is the oracle or test harness. Oracle: fixture source executed by Node through oracle/node.mjs, compared against native and generated JavaScript stdout, stderr and exit status. Input additionally compares every written file's bytes and permissions and checks permission-denial behavior. Node and the harness were unchanged.

Scope and baseline

origin.txt records the fresh origin/main commit. CLAUDE.md, the audit REPORT.md and PLAN.md, all target and subsumer test files, and the relevant production code were read before mutation. Audit report/plan/rows are copied with audit- prefixes. tests.txt records the current full test list. New current rows were considered, rather than retaining the audit's old 15-row slice.

The uncached full baseline reached the binary's 90-second timeout with no earlier test failure. A six-row baseline passed, and coverage runs for each target and TestNumericEnumNeverPinned passed. A broader baseline excluding the main fixture/review/WASI aggregates also timed out; count checks remained a large aggregate. Both broad mutant matrices timed out too and are retained as cooked, incomplete evidence. Their partial passed/failed observations establish no uniqueness.

The final matrix consists of the 28 exact top-level names in narrow-matrix-rows.json, including current class, module, namespace, input and file-order rows. Its clean baseline completed successfully before these matrix runs. Each variant used a separate ADAMIC_BUILD_CACHE_DIR and ADAMIC_GATE_UNCACHED=1. Exact commands, completed passes and failures are in attempts.json. No row skipped in the completed matrix. Its results are bounded: corpus/count/WASI aggregates and other rows outside this list remain unknown. No full-package or repository-wide uniqueness is claimed. Replay of the standalone diffs can settle that later.

Coverage and leads

Each target and its subsumer ran alone with -coverpkg covering internal/lower, internal/native and internal/javascript. Profiles, commands and logs are preserved. exclusive-blocks.json lists 248 cycle-row blocks and the input-row blocks absent from the subsumer. cycle-functions.txt and input-functions.txt name reached Go functions. These profiles measure Go, not C runtime coverage.

The cycle row reaches staticDeclaration's imported class-base readiness branch. Its classes/c.a entry loads a cycle in which Middle extends Base before Base's public binding is initialized. The enum subsumer has no imported class-base initialization. The corresponding safe classes/a.a entry also runs, so the target checks both failure and successful ordering.

The input row writes again.txt first with a long value, then with 'short'. The enum subsumer has no file writes. write-callers.txt records the source search: write_files.a and write_stdout_order.a/write_stderr_order.a use the runtime operation. input_spread_local.a declares a local writeTextFile function and does not use that runtime entry. The completed matrix includes the input and stream-order parent rows. Other caller reachability, including transitive corpus imports, was not exhaustively proven; this is another reason to retain the bounded qualification.

Mutants and outcomes

D1, input.c:395: remove O_TRUNC from the open flags. This is a change of an existing option. Strict C11 syntax validation used clang -std=c11 -Wall -Wextra -Werror -pedantic with the runtime include directory; the matrix also built native products. Only TestInputAgreesWithNode fails. input_test.go:159 reports again.txt expected 'short', but native left 'shortg first text, longer than the second'. input_test.go:169 also reports stdout differs. All 27 other matrix rows passed, including the named enum subsumer and TestFileWritesLandInNodesOrder. This proves the overwrite/truncation subcase's independent value, not uniqueness of every input subcase.

D2, class_static.go:248: change checked = !l.provenModuleReads[expression] to checked = l.provenModuleReads[expression]. This flips the existing readiness condition. go vet ./internal/lower/ passed; the matrix rebuilt its native products with a separate cache. Only TestImportCycleLoadTimeReads fails, at import_cycles_test.go:117 with 'stderr differs' in classes/c.a. All 27 other rows passed, including the enum subsumer and current namespace/readiness controls. It guards imported class-base initialization independently within this bounded matrix.

Both standalone diffs apply independently to the starting origin/main, as checked in apply-validation.txt. Sources were restored between variants; no switch, test edit, harness edit or oracle edit remains. rows.json includes all passed-row names. replay.py reproduces the bounded mutants. No survivors occurred in these completed matrix runs. Neither target is a cost row or a Node/native twin pair.

Time, complications and limitations

The initial disk check showed 5.3 GB free in /tmp and 15 GB in /workspace. /tmp's total capacity is 8.8 GB, so the requested 15 GB free is impossible. Only prior-unit /tmp/regexp-defense was removed; tools and the repository were untouched. Warm env.sh worked, setup was skipped, npm ci in stage3/api succeeded before baseline, and nproc is 5.

Full and broad baseline timeouts, plus two cooked broad mutant runs, consumed about six minutes of binary time. Their outer command times include compilation. The three coverage commands took about 13 seconds in total. Completed bounded matrix commands took 18.425 and 17.977 wall seconds, including Go/build preparation; native rebuild time was not separately measured. Overall work took approximately 16 minutes. Each cooked binary stopped at its 90-second timeout; none was treated as a semantic kill. The restored bounded baseline is retained separately.

The supplied audit evidence strings are truncated, so the full commands and context were taken from the fetched report. The audit subsumption rested on a single coarse omitted-main mutant; it did not establish coverage of these class and I/O differences. Current test count is recorded in tests.txt, not inferred from the audit. One relative filename lookup missed a moved production file and was corrected using rg; one attempted evidence write used the read-only sandbox and was repeated with authorized write access. Go coverage cannot supply per-line C runtime measurements. Broad matrices were too large even after initial narrowing, so exact final scope and unknown results are explicit. No tests were deleted, rewritten or weakened, and no main push or PR was made.

Both names match their assertions: the cycle row compares actual load-time reads, and the input row compares Node execution and filesystem effects. Neither promises an unasserted performance threshold. The bounded finding is a reason to keep both rows while central replay resolves excluded rows.
