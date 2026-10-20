Added 35 measured a-check headers to the records change, preserving all 109 program bodies byte for byte.
Commit: branch HEAD at handoff, one commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 611cbc943a28e88567bcb5fb9a43031518a69d3d.
Commands/results: the final real-compiler a-check accepts all 109 files: 64 checked, 3 NotYet, 9 refused, 33 type errors; affected readers, reader guard and vet pass.
Mutants: removed header and wrong refusal reason both fail unchanged Gate.aCheck; the old oracle line pin fails its exact location assertion; the stock CLI input control is caught.
Not covered: whole packages, full gate, Loom/ahra, new compiler feature validation or a new census; this pass checks headers and their readers, not the unchanged compiler implementation.

The exact requested base overrides the generic current-main instruction. Fetched origin, checked out compiler/loom-r-records-ac at dbdf977f5b, and applied only 611cbc94 with cherry-pick --no-commit. There were no conflicts. The result includes that change plus headers, shifted expectation metadata and this evidence. No cohere code was copied. Compiler and runtime sources are identical to 611cbc94, as scope-proof.json proves.

Read CLAUDE.md, docs/front-door.md, README.md and stage3/a-check-headers/README.md whole. The unchanged docs/0.1.md and docs/memory.md were already read during the preceding unit. The README's pinned fast-gate object initially lacked its file locally; fetching devtools/fast-gate made it available. gate.py is byte-identical to fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py. predicate-source.json records its hash. Only Gate.aCheck is invoked, never the whole gate.

Scan scope and results:

- Every added or modified .a in git diff dbdf977f5b was selected without directory exclusions, including archived review fixtures, input-project main.a files, and node_modules declarations ending in .d.a. There are 109, with no omitted paths or unclassified failures.
- STATUS.md is the complete per-file status table, including every NotYet reason and checker code. before.json, after.json and final.json retain full real compiler diagnostics, exits and output sizes.
- Each scan uses this branch's own compiler through the exact Gate.aCheck command, adamic c <path>. This checks and lowers to C; it does not claim every program was linked or executed natively. Every compiler invocation is limited to 45 seconds, its compiler build to 120 seconds, and each batch to 600 seconds.
- Before changes, the predicate rejected exactly 35 files for missing headers. Both post-header and final scans accept all 109, exit 0. Gate internally combines clean and NotYet as checked; this report separates them as requested.

| Location | Checked | NotYet | Refused | Type error | Total |
|---|---:|---:|---:|---:|---:|
| internal/oracle | 27 | 1 | 8 | 3 | 39 |
| review evidence | 36 | 2 | 1 | 30 | 69 |
| stage3/records-maplike-main | 1 | 0 | 0 | 0 | 1 |
| Total | 64 | 3 | 9 | 33 | 109 |

HEADERS.md lists every first line added, and headers.json includes every file's old/new header and body SHA-256. Exactly nine refusal headers use the measured full reason and fix text; 26 new type-error headers use the measured TS code. Seven existing checker headers are already correct and unchanged. All 67 clean/NotYet files take no header. No program body was edited, reformatted, normalized, renamed or regenerated. body-and-position-proof.json verifies all 109 bodies against source 611cbc94 and proves every file without a header edit is wholly identical.

Position expectations updated:

1. internal/oracle/records_refusal_test.go, recordRefusalPin: prototype_read 2:54 to 3:54; prototype_in 2:53 to 3:53; prototype_set 2:34 to 3:34; compare_properties_left 9:9 to 10:9; compare_properties_right 9:20 to 10:20; environment_boundary 3:12 to 4:12; named_invalidated 9:20 to 10:20; narrowed_number 3:58 to 4:58. Reason and fix pins are unchanged. The eight tests and three wrapper readers all pass.
2. Archived review/compiler/loom-records-3/stage3-mutant/records/status.json: 16 diagnostic locations in 11 rows advance by one line. They are 04_own_values, 08_option_boolean, 09_option_undefined, 10_option_wrong_type, 11_option_string, 12_nested_entries, 14_group_by, 15_inherited_read, 17_compare_missing_object, 18_compare_missing_scalar and 19_compare_empty_objects. archived-position-updates.json lists every exact before/after location. All eleven full diagnostics match fresh compiler output, not only their line numbers. The real stage3 fixture runner passes all eleven Node and stage0 comparisons using this archived root. Other status fields and every Node observation in this table are unchanged. The archive's intentional 07_optional_view Compiles mutant remains untouched and was excluded from this positive reader selection.
3. Archived input-fixtures/observations.json: five stock Node compiler stdout positions advance from main.ts(1,23) to main.ts(2,23): 01_paths-missing_constructor, 02_paths-missing_toString, 03_paths-missing_hasOwnProperty, 04_paths-missing___proto__, 08_paths-own___proto__. The first four were already stale under pre-existing headers in 611cbc94. Conservative choice: align these line pins too, using actual stock TypeScript 6.0.3 output, rather than retain stale expectations. A structural assertion allows only line-number differences. Codes, columns, messages, stdout apart from positions, stderr, exits and behavior classifications are unchanged. The unmodified probe-inputs.py then passes all 30 exact observations and its real-input mutant. archived-input-position-updates.json records the five changes.

Literal path/basename searches of *_test.go and testdata are saved in path-readers.json; generic main.a basename false positives are removed. dynamic-reader-search.log records the refusal-helper directory search. Dynamic stage3 readers were checked as well, rather than relying only on literal filename hits. No other active position pin was found. Historical run logs remain historical observations; new scan evidence provides the current positions.

Mutant proof:

- On records_refused/records_prototype_read.a, remove the header: the real compiler still reports refused, while unchanged Gate.aCheck expects checked and exits 1. Caught in 0.265 seconds.
- On the same file, replace the reason with a deliberately absent reason: the compiler still reports refused; the exact predicate rejects its reason and exits 1. Caught in 0.215 seconds.
- Both header mutants invoke the actual branch compiler, not a hand-written predicate or fake diagnostic. The source is restored in finally, and its complete bytes are verified afterward. header-mutants.json and their individual JSON/log files record both catches.
- Overlay the original 611cbc94 oracle test with its old :2:54 location pin: TestRecordRefusalPrototypeRead fails with diagnostic changed because the actual refusal is at :3:54. The package test takes 0.075 seconds. The compiler's intended Refused diagnostic, not a compile error or warning, triggers the assertion. old-line-pin-mutant-proof.json records it; the mutant Go source is old-line-pins.go.txt.
- The unchanged stock input reader again catches replacing the lost own __proto__ path with an ordinary own constructor path through different real TypeScript CLI output.

Validation and commands:

COMMANDS.md and commands.jsonl give exact commands, limits, exits and command wall times. Test output goes directly to logs, never a pipeline. All Go tests use -count=1, -timeout 90s and -parallel 4, with GOMAXPROCS=4.

- Eleven oracle refusal/helper readers pass, package 0.373 seconds. Individual top-level times are 0.09 to 0.19 seconds. No test was added; only the eight helper position expectations changed.
- TestCallTargetReaders passes in 1.26 seconds, with its TestProduct_CallTargetExports owner. No changed leaf exceeds the 60-second budget. test-seconds.json contains every observed top-level time.
- The archived fixture selection passes, package 3.381 seconds, with its hook owner prepared first in 9.18 seconds. All 11 Node outputs and all 11 stage0 pins agree.
- Stock TypeScript 6.0.3 input observations pass twice, including the final unchanged expectation-checking mode. Each run is bounded at 180 seconds; its child commands have 30-second limits. Neither run changed source bodies.
- TestCountsAreRecordedTable passes. counts.md is byte-identical to 611cbc94; no counted fixture or program body changed, so counts were not regenerated.
- Vet passes for internal/oracle, internal/ir and stage3/fixtures. No compiler consumer package or runtime feature set was added.

Toolchain: export GOPROXY='https://proxy.golang.org|direct', then ADAMIC_GOCACHE_OFF=1 timeout 600s bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. GOCACHEPROG was unset for local checks. Setup succeeds, no workaround. nproc=5, four-CPU cgroup quota (400000 100000), Go 1.27.1, Node 24.19.0, Linux clang 20.1.8. Timing lines: Node 0.045s; Go 0.051s; submodules 0.092s; markdown dependencies 0.117s; clang 0.286s; shared cache off 0.288s; Go build 13.273s; test binaries deferred 13.484s; cache warm 13.486s; done 13.531s. setup.log is authoritative.

The preceding unit's feature validation and two broader Python reader deadline limitations remain in its original report; this header pass does not claim they were rerun. No whole package, full gate, ahra command, Loom submission or PR was run. The final handoff reports the new single-commit SHA and remote push verification.
