Unknown-error source narrowing remains Refused at `in`; assigned System source fixtures remain NotYet at `node:fs.mkdtempSync`.
Merged library area 2faf682 into process landing 94df56b, retaining both fixture-registration intents and the area QualifiedName fix; node:perf_hooks is included.
Commands and outputs: full oracle PASS 329.280s; final process/counts PASS 60.957s; flow PASS 94.507s; audited host fixtures PASS 17.516s; affected packages and vet passed; Linux nproc 5.
Mutants: all 19 host/performance mutants and dedicated environment, directory, blocking, caching, error-code, tmpdir, namespace and exit-output mutants are rerun against Node.
Not covered: native execution of the eight assigned System fixtures, unknown/in source narrowing, timers, nextTick scheduling, Date System.now, native tsc, or macOS.

The merge has parents 94df56b and 2faf682. This follows the latest explicit merge instruction: no rebase, no force push, and no push to main. Only the process landing branch is pushed for integration into area/library.

The area versions of stage3/fixtures/NOTICE and stage3/fixtures/host/status.json are retained exactly. Both sides of the oracle fixture-list conflict are kept: the existing process exit/observations/shadow fixtures and area's typeof_null fixture. Counts are regenerated rather than selected from either parent.

The first flow run failed because its top-level compile-fixture glob found four deliberately runtime-only witnesses. They now live under internal/oracle/testdata/node_process_runtime with unchanged source text. Their dedicated tests still construct typed IR, run Node as the independent oracle, compare native and JavaScript backends, and execute mutants. The exact adaptation-47 helper refusal is still explicitly asserted. No flow assertions or compiler refusals were relaxed.

node:perf_hooks includes performance.now, timeOrigin, mark, measure, clearMarks and clearMeasures, including the performanceCore tryGetPerformance shape. Source fixtures node_process_performance.a and node_process_performance_core.a compare both backends with Node; machine-dependent time values are checked for types, ranges and monotonicity rather than identical bytes.

All 25 audited host fixtures match Node and their recorded compiler stages. For this unit, 14_getCurrentDirectory, 15_getExecutingFilePath, 16_getEnvironmentVariable, 17_write, 18_exit_0, 19_exit_1, 20_exit_2 and 23_newLine each remain NotYet at node:fs.mkdtempSync. Passing the recorded-stage check does not mean native source execution is green.

Verification logs are under logs/library-merge/. The full oracle passed before witness relocation; the complete process subset and counts are rerun after relocation. The original failing flow log is retained alongside the successful rerun.

| Check | Command | Result |
| --- | --- | --- |
| Counts regeneration | `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts` | PASS 34.127s |
| Full oracle | `go test ./internal/oracle -count=1 -timeout 30m` | PASS 329.280s |
| Final process and counts | `go test ./internal/oracle -run '^TestNodeProcess\|^TestProcessExitOutput$\|TestInputAgreesWithNode/internal/oracle/testdata/node_process\|^TestCountsAreRecorded$' -count=1 -timeout 30m -v` | PASS 60.957s |
| Host acceptance records | `go test ./stage3/fixtures -run '^TestFixtures/host$' -count=1 -timeout 30m -parallel 4 -v` | PASS all 25, 17.516s |
| Affected packages | `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/regexp -count=1 -timeout 30m` | load PASS 4.586s; lower PASS 40.061s; native PASS 169.919s; regexp PASS 7.359s; JavaScript has no package tests; initial flow failure retained |
| Final flow | `go test ./internal/flow -count=1 -timeout 30m` | PASS 94.507s |
| Final vet | `go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/regexp ./internal/oracle ./stage3/fixtures` | exit 0 |

The 19 host mutants are cwd, platform, EOL, nextTick feature, pid range, argv, execArgv, columns, handle feature, write, memory range, monotonic now, time origin range, performance hooks, mark, measure, clear all marks, clear named marks, and clear measures. Dedicated tests additionally kill environment set/delete, cwd-cache invalidation and directory error code, numeric exit error code, removed-cwd error code, terminal blocking/cache, temporary-directory precedence, os namespace, and stdout-drain/exit mutants. The logs record each executed comparison and expected difference; no sanitizer failure substitutes for a Node mismatch.
