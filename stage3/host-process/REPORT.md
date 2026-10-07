Catchable host errors are missing: the runtime Error shape has name/message but no code, and cwd/measure failures have no pending-exception CFG integration; synchronous host coverage is implemented.
Commits: ad56ba382e1e3124807758766c6aa852bf6e9c49, correction cb907016760b2badf006fac6fe98e1a2adb42f4e; branch codex/host-process extends dd4b67e7bdedd27f31d2e3900e7caf1fd5575374.
Validation before declaration correction: Node v24.19.0 on both backends, sanitizers/leaks and 20 clean mutants; final declaration integration awaits the shared hook.
Mutants before declaration correction: all 20 compile/run cleanly and fail Node output, stderr or exit comparison; rerun against @types/node 25.3.3 remains pending.
Not covered: shared @types/node hook integration, timers, nextTick callback scheduling, Date-valued System.now, full System, catchable errors and native tsc proof.

## Scope and evidence

This is a partial host-3 delivery, not a completed stage-3 tsc proof. Census source is origin/codex/tsc-census at 429c1177f0130f785c19cf590d1860513b2ddbfc, TypeScript 6.0.3. The requested dependency branch takes precedence over the general instruction to start from main. Existing exitCode, exit, env reads, isTTY and exit draining were extended, not replaced.

Per the later correction, no private declarations or loader hook remain. The fs_file worker owns the shared @types/node 25.3.3 hook; its commit SHA has not yet been supplied. Native code is node_process.c/.h; lowering and backend extensions are library_node_process.go. Eight .a fixtures and two external Python drivers cover input arguments, real Unicode directories, pipes, files, TTYs and inherited nonblocking pipes. No prohibited emitter/lowering/oracle files were edited.

The initial implementation used narrow private declarations for fixture development. Those files and loader edits were removed after the correction. Existing dependency-branch declarations were restored unchanged. Static node:* imports will use the shared hook. Unsupported os/performance operations have named NotYet paths, and existing process refusal paths include the member name. Exhaustive refusal of the broader pinned declarations, imported process identity, namespace/default import resolution and aliased result members must be checked/adapted once the hook supplies declaration provenance. They are not proven yet. No replacement loader or require implementation is supplied.

A refresh of origin/codex/tsc-census still produced 429c1177f0130f785c19cf590d1860513b2ddbfc, whose tree does not include stage3/api. The new host fixtures cannot compile against the restored baseline declarations: the shared hook is an explicit outstanding dependency. The shipped fixture registry intentionally retains these checks, so the full oracle gate will fail until that dependency is merged; nothing is skipped to conceal this.

## Contract

| Member | Native behavior and proof boundary |
| --- | --- |
| process.cwd | UTF-8 path, cached on first success like observed Node; Unicode cwd and removal after first call tested. Missing cwd panics; catching is refused. |
| process.argv | Stable mutable array; UTF-8 and invalid-byte replacement; slice(2) matches passed arguments including empty strings and --prof. First two entries are native executable path rather than Node and script paths. |
| process.execArgv | Stable empty native array: no Node VM flags exist. A command-line --prof remains an ordinary argv member. |
| process.env | Existing indexed reads retained. Census contains no env iteration; enumeration is not implemented. |
| platform, os.platform, os.EOL | linux and LF on gate machine; darwin and LF in macOS code. |
| pid | Number, positive integer; machine-specific, not byte-equal across processes and not monotonic. |
| stdout.columns, isTTY | Undefined on pipe/file; true and ioctl column count on TTY. Fixed 93-column PTY compared against Node. |
| stdout.write | Ignored string writes only, preserving raw encoding/output order and exit draining. Observing its backpressure boolean is refused. Buffer/encoding/callback overloads are not built. |
| stdout._handle.setBlocking | Missing handle on regular files; present on pipe/TTY. Flush and fcntl O_NONBLOCK setter; inherited 4 KiB nonblocking pipe tested. Other character devices and handle identity are not proven. |
| exit | Existing branch implementation and mutant suite retained. No profiling callback integration. |
| memoryUsage.heapUsed | Finite positive numeric bytes; no monotonic guarantee. ASan allocator allocated bytes under sanitizers; Linux mallinfo2 uordblks+hblkhd normally; macOS malloc zone size_in_use. Includes runtime/allocation footprint, not V8 live heap, so cannot match Node magnitude. |
| nextTick | typeof and double-negation feature observations only; callback scheduling refused. |
| performance.now | Nonnegative finite milliseconds from native startup using CLOCK_MONOTONIC; nondecreasing, machine-specific, not byte-equal. |
| performance.timeOrigin | Stable finite positive epoch milliseconds from CLOCK_REALTIME at native startup; machine-specific, not monotonic across processes. |
| performance.mark | Returned name/type/startTime/duration shape; duplicate marks resolve newest first. startTime obeys now's range/order; duration zero. |
| performance.measure | String start/end mark overload, omitted endpoints, negative duration, newest duplicates; name/type/startTime/duration shape. Resolves end before start; missing-mark panic text matches Node. Catching refused. |
| clearMarks | Named deletion or all deletion; cleanup at process exit. |
| clearMeasures | No-op for this non-enumerating surface; returned measure entries work, but no measure registry exists. Tested that clearing measures preserves same-named marks. |

Machine-dependent observations are compared through type/range/stability/order predicates, not raw numeric equality. Timestamps and measurement durations are finite numbers; durations can be negative and are not monotonic. TTY width depends on the terminal. Memory can rise or fall.

## System members

The adapter fixture exercises args, newLine, write, getExecutingFilePath, getCurrentDirectory, getEnvironmentVariable, getMemoryUsage, clearScreen, debugMode and cpuProfilingEnabled. clearScreen writes exactly the three census escape sequences. Native getExecutingFilePath returns the executable path so installed compiler libraries can live beside the binary; it cannot return Node's tsc entry-script path because there is no entry script at runtime. Linux resolves /proc/self/exe; macOS uses _NSGetExecutablePath and may require canonicalization for symlinks.

Direct process observations prove primitives for writeOutputIsTTY/getWidthOfTerminal; the original boolean-or-undefined function-value signatures are not proven. Original optional-parameter function wrappers and Date-valued now exceed this adapter. exit works directly via the dependency branch. setTimeout/clearTimeout remain absent: there is no event-loop/timer scheduling implementation in this patch. debugMode tests the inspector environment and execArgv regex; no native debugger/recordreplay or profiling engine is provided. No claim is made that unmodified getNodeSystem now compiles.

Catch refusal tests cover direct cwd, transitive measure, stdout backpressure observation and nextTick scheduling. This avoids silently pretending panic is catchable. Native errors with .code and exact host-operation messages still need error payload plus exception-flow integration; the existing Error name/message shape alone cannot meet that contract.

## Mutant proof

| Mutation | Node comparison that catches it |
| --- | --- |
| cwd returns platform | stdout path differs |
| platform wrong | stdout differs |
| EOL CRLF | stdout differs |
| nextTick feature false | stdout differs |
| pid zero | positive-integer predicate differs |
| argv drops last argument | stdout differs |
| execArgv injects --prof | profiling predicate differs |
| terminal columns 80 | stdout differs from 93-column Node TTY |
| handle presence inverted | TTY stdout differs |
| write adds newline | stdout bytes differ |
| heapUsed negative | positive finite predicate differs |
| now sign reversed | nonnegative/monotonic predicate differs |
| timeOrigin negative | epoch-range predicate differs |
| mark duration one | stdout differs |
| measure sign reversed | stdout differs |
| clearMarks all disabled | exit codes differ on missing mark |
| clearMarks named disabled | stderr missing-mark name differs |
| clearMeasures clears marks | exit codes differ |
| setBlocking fcntl skipped | Node writes 200000 bytes/exit 0; mutant 4096 bytes/exit 70 |
| cwd caching skipped | Node retains path/exit 0; mutant ENOENT/exit 70 after directory removal |

The clearMeasures test was strengthened after the initial mutant survived: it now creates a measure and mark with the same name before clearing the measure. All final mutants were caught. Existing exit-output mutants initially failed to link after the shared raw-write export was added; the harness rename list now includes that export, and the suite passes.

## Validation and environment

Run on Linux, nproc=5. Setup: Go 1.27.1 (0s), clang 20.1.8 (0s), Node v24.19.0 (0s), submodules (0s), build-cache warm (71s), total 71s; cgroup cpu.max 400000 100000, 17.6 GB. Environment file: /workspace/adamic-tools/env.sh. Logs are copied alongside this report.

Commands (each output redirected to its log, never piped):

- bash cloud/setup.sh: success; setup.log.
- go test -count=1 -timeout 10m ./internal/oracle -run '^TestProcessExitOutput$|^TestNodeProcess|^TestInputAgreesWithNode$/internal/oracle/testdata/node_process' -v: PASS, 5.512s; oracle.log.
- go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts: PASS, 10.100s; counts.log. Only three new rows changed; allocations equal frees.
- go vet ./...: success, empty vet.log.
- Post-correction go test -count=1 -timeout 10m ./internal/load ./internal/lower ./internal/native: PASS (load 1.501s, lower 14.638s, native 68.062s); post-correction-packages.log.
- Post-correction go test -count=1 -timeout 5m ./internal/oracle -run '^TestNodeProcessTerminal$' -v: expected dependency failure, TS2339 for columns, _handle and write absent from baseline declarations; missing-shared-hook.log.
- git diff --check and gofmt -l on new Go files and modified mutant harness: success, no output.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...: interrupted after the user declaration correction; all touched packages passed before interruption (native 218.772s, oracle 190.342s, lower 26.462s, load 1.967s), but this is not a final-source full-gate pass; full-gate.log. The prior gate failed only at the exit-mutant duplicate raw-write symbol; failure log retained.

The oracle, vet and count results above precede removal of private declarations. Post-correction package checks and exact missing-hook output are preserved separately in post-correction-packages.log and missing-shared-hook.log. ASan, UBSan and leak comparisons are included in oracle tests. Linux is the gate of record. macOS branches exist but were not run: platform is darwin, heap measurement uses malloc zones, executable-path resolution differs, and the F_SETPIPE_SZ backpressure test is skipped. TTY/file/pipe checks are intended to run there but are unverified.
