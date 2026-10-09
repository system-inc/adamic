Converted all 26 acceptance rows for task #41bkfdw wave 3 onto the shared Node agreement helper.
Branch compiler/agree-library; commit recorded in the delivery report.
Territory tests: PASS, 7.422s; TestCallTargetReaders: PASS; lane output recorded separately.
Mutants: parseInt radix fixed at 10 and existsSync lowered as stat both fail converted tests and pass original tests.
Not covered: native backend execution; 108 refusal and five checker metadata rows retained, no new oracle fixtures.

Census: rows.md records 139 source rows across all nine library test files and fs_method_guards_test.go. No source row was removed or skipped. The receiver ownership probe names the metadata behavior cannot observe; the four regex producer probes are negative checker-only tests.

The duplicated filesystem Node runner and the behavior-observable parseInt callback IR assertion were removed. Absolute paths isolate filesystem state under parallel tests. Silent rows print computed results; the missing unlink row prints before its expected exit 70. Buffer borrowing uses a private descriptor and prints read count, buffer content and written file content. mkdtemp prints existence and cleans up its random directory. Method-value subtests now use t.Parallel.

Commands: test-command.txt contains the exact territory selector. All go test invocations use -count=1 -timeout 90s, with external timeout 100 for the final runs. The first broad-selector attempt failed because pinned Node declarations were absent; npm ci --prefix stage3/api fixed that. The first final-selector run exposed numeric console.log, fixed using toString. Final tests passed, and test-seconds.json records every leaf. The longest touched leaf is under 2s.

Setup: exported GOPROXY='https://proxy.golang.org|direct', ran timeout 240 bash cloud/setup.sh. It timed out during dependency warming after Node 0.030s, Go 0.032s, clang 0.231s, markdown 1.388s, submodules 17.546s. Sourced /workspace/adamic-tools/env.sh and installed pinned declarations. Retry timeout 120 bash cloud/setup.sh passed: Node 0.036s, Go 0.051s, markdown 0.137s, submodules 0.153s, clang 0.339s, Go build 16.184s, total 16.315s. nproc=5; cgroup quota is four CPUs.

Mutants were temporary compiler changes only, restored before commit. Each .diff is one line. parse-int-radix-converted.log records output 10,10 versus Node 10,NaN; parse-int-radix-old.log passes the original acceptance-only table. fs-exists-converted.log records the output/exit disagreement for TestNodeFSFileNamespaceImport; fs-exists-old.log passes its original lower-only assertion. No production changes are delivered.

No fixture added, so counts.md refresh is unnecessary. No full packages or full gate run. Conservative classification: checker ownership and negative regex producer proof are IR-only rather than acceptance, since neither executes a lowered program.
