Unknown-error source execution remains Refused at `in`; the exact adaptation 47 helpers and runtime own string code/message contract are tested.
Built process/os/performance primitives and namespace/default imports; pushed checkpoints d06ed47, 157f53e, 8499196; landing branch rebased onto origin/main f8013f0.
Commands: landing process oracle PASS 10.788s, counts PASS 12.926s, lowering PASS 11.174s, affected-package vet exit 0; load/native package checks passed after adaptation merge.
Mutants: all host, blocking/cache, env/directory, error-code, temp-directory, namespace and exit-output mutants are caught by Node comparisons; complete names and effects below and in logs.
Not covered: assigned System fixtures all NotYet at fs.mkdtempSync, unknown/in source execution, timers, nextTick scheduling, Date System.now and native tsc; no macOS run.

## Landing verification

Branch codex/host-process-land was created from the pushed process branch and rebased with `git rebase --rebase-merges origin/main`; success. Current origin/main is f8013f0baac41ddc340d76f83bddde38536a8f07 and is an ancestor. No main push and no force-push occurred. The branch includes the explicitly requested loader, fixture and adaptation-47 branch merges; their source/evidence files are part of the merge, not new edits by this unit.

Additional namespace/default-import fixture node_process_os_namespace.a agrees on both backends for os.platform, os.EOL, os.tmpdir, imported process.cwd and process.env. Its EOL mutant changes only Node stdout. Namespace receivers resolve by official declaration identity; first-class os namespace objects remain NotYet. Namespace fixture is a custom oracle test rather than a row in the generic input registry.

Commands, all redirected to logs:

- `go test ./internal/oracle -run '^TestNodeProcess|^TestProcessExitOutput$|TestInputAgreesWithNode/internal/oracle/testdata/node_process' -count=1 -v`: PASS 10.788s, land-oracle.log. This includes the Refused-blocker assertion and independent typed-IR error tests; it does not mean unknown-error source compiles.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1`: PASS 12.926s, land-counts.log. Preceding full refresh PASS 13.247s; only process_bad_code retains changed (31 to 41), and process rows moved after merge. Allocations/frees are unchanged.
- `go test ./internal/lower -count=1`: PASS 11.174s, land-lower.log.
- `go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/oracle`: exit 0, land-vet.log.
- `go test ./internal/load ./internal/lower ./internal/native -count=1` after merging adaptation 47: PASS 3.661s, 14.170s, 78.865s, adaptation47-packages.log. The final error-message/path changes and namespace change are covered by the landing oracle and lowering package respectively.
- `python3 stage3/host-process/check_acceptance.py --compiler /tmp/host-process-land-adamic --logs /tmp/host-process-land-acceptance`: exit 1 with eight expected NotYet rows; each unchanged Node fixture matches status.json and each Node source mutant is caught. Exact per-fixture stages are in acceptance/land-stages.json.

| Assigned fixture | Native | JavaScript | Node |
| --- | --- | --- | --- |
| 14_getCurrentDirectory | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees |
| 15_getExecutingFilePath | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees |
| 16_getEnvironmentVariable | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees |
| 17_write | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees |
| 18_exit_0 | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees, exit 0 |
| 19_exit_1 | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees, exit 1 |
| 20_exit_2 | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees, exit 2 |
| 23_newLine | NotYet: fs.mkdtempSync | NotYet: fs.mkdtempSync | Agrees |

Additional clean native mutants beyond the original table: performance hook object's epoch changed to zero (clock predicate/stdout); os.tmpdir drops two slashes (stdout); os namespace EOL becomes CRLF (stdout); exit ERR_OUT_OF_RANGE becomes ERR_INVALID_ARG_TYPE (stdout); removed cwd ENOENT code becomes ENOTDIR (stdout); env set/delete disabled (stdout); successful chdir cache invalidation disabled (stdout); chdir ENOENT code changed (stdout). Five inherited exit-output mutants are file_immediate (flush omitted), file_normal (exit flush omitted), merged_file/merged_pipe (write ordering buffered), large_file (tail omitted); each is caught by Node stdout bytes. Logs name every run and record sanitizer-clean comparison failure.

The error tests cover all seven mapped chdir errno and uncached cwd ENOENT. Normal errors have own string code/message. Missing performance marks follow Node's SyntaxError name/message; Node's DOMException code is inherited numeric 12, so the string-only errorCode helper returns undefined for that non-process exception. Generic reflection and broader Node APIs beyond the census are not certified by this unit. Timers are unbuilt: there is no timer/event-loop scheduler in this patch; callback nextTick remains named NotYet. This is a partial stage-3 host delivery, not the native tsc --noEmit proof.

@types/node 25.3.3 is installed locally under stage3/api/node_modules, outside tracked source. No private node_*.d.ts files or replacement require loader were added. The pinned declaration installation remains necessary when reproducing tests in a clean checkout.

Earlier reports below are historical. This landing summary supersedes their older stage, dependency and proof limitations where explicitly stated.

Unknown-error source narrowing is blocked by Refused `in` at node_process_errors.a:5:66; runtime own string code/message and exact Node errors are built.
Commits pushed: d06ed47 performance first, 157f53e process/tmpdir; adaptation cf976e8 merged as dd3c56f, exact helpers now used in both error fixtures.
Commands: post-adaptation errors PASS 5.234s; affected packages load/lower/native PASS 3.661s/14.170s/78.865s; full counts PASS 13.247s; affected-package vet PASS.
Mutants: original 19 host mutants and ancillary process/exit/env/cache mutants remain green; new ERR_OUT_OF_RANGE, removed-cwd code, tmpdir and performance-hook mutants are caught only by Node observations.
Not covered: eight unchanged System acceptance fixtures stop at fs.mkdtempSync on both backends; unknown/in source execution, timers, nextTick scheduling, Date System.now and native tsc remain incomplete.

## Adaptation 47 error contract

Merged the requested complete branch at cf976e8. The source fixtures contain the helper text copied exactly from its adapt.cjs, and a drift check compares the text. Node's real errors pass through errorCode/errorMessage, which take unknown and use typeof, !== null and in. The source lowerer remains Refused; this is also the adaptation's documented compiler blocker. No assertion, private declaration, or syntax-policy exemption replaces those helpers. The separate runtime oracle explicitly constructs typed IR and is not counted as source acceptance.

All seven mapped chdir errno are exercised: ENOENT, ENOTDIR, EACCES, ELOOP, ENAMETOOLONG against real directories, ENOMEM and EIO with the same Linux libc fault shim loaded into Node and native. Empty paths, embedded NUL and a lone surrogate are tested. Node truncates the error destination at NUL and replaces lone surrogates when converting it to the native path; the runtime now does the same. Both exit and exitCode are exercised with NaN, positive/negative infinity and a fraction. Own code and message properties are checked for every caught error.

A real parent-process removal of the child's cwd tests uncached process.cwd after chdir invalidates the bootstrap cache. This exposed Node 24.19's longer ENOENT description, now copied exactly: `process.cwd failed with error no such file or directory, the current working directory was likely removed without changing the working directory`. Both backends match its full message and own fields. A code mutant exits normally and differs only in Node stdout. Sanitizers cover all runs; regular runtime tests include leak comparison, while the removal driver's sanitizer leak detection is disabled because the deleted cwd interferes with its exit probe.

Linux is the gate. macOS uses the same POSIX host code but its fault-injection branch is not run; no macOS execution is claimed. Machine-dependent pid, memory, clock types/ranges/order and native executable-path rationale remain in the contract table below.

Landing rule is acknowledged: create codex/host-process-land, rebase onto current origin/main, re-green, push that branch only. Never push main and never force-push.

Earlier checkpoint reports below are historical; this supersedes their untested ENOMEM/EIO and removed-cwd-message statements.

Process unknown-error narrowing remains blocked by the compiler's explicit refusal of `in`; runtime own string code/message and Node messages are implemented.
Built os.tmpdir and official Node process integration; performance checkpoint d06ed47 was pushed first, with shared loader 69c71d5 merged.
Commands: every process unit test plus all four registered process fixtures PASS 8.706s; layouts/uniform-field check PASS 7.352s; tmpdir matrix PASS 6.202s.
Mutants: 19 general host mutants, setBlocking/cwd-cache, environment set/delete, directory cache/code, ERR_OUT_OF_RANGE, tmpdir, and five existing exit-output mutants all caught by Node comparisons.
Not covered: all eight unchanged System fixtures are NotYet at fs.mkdtempSync; timers, nextTick scheduling, Date-valued System.now and native tsc remain incomplete.

## Process, temp directory and error checkpoint

node_process_host.a, node_process_system.a, node_process_performance.a and node_process_performance_core.a now pass on native and JavaScript. The terminal fixture compares regular file, pipe and a 93-column TTY. Local optional _handle annotations represent sys.ts's private feature probe without modifying official declarations. Blocking and memoized-cwd drivers pass; no private node_*.d.ts copies exist.

Official WriteStream declarations advertise nonoptional columns/isTTY even on a pipe where Node returns undefined. Guarded reads preserve that absence; unguarded uses that assume the advertised number/boolean receive named NotYet rather than a runtime narrowing panic. Generic Dict indexing is enabled only for process.env. Writable/Socket.write is enabled only for the supported stdout receiver. All runtime object layouts participate in the field-offset proof.

os.tmpdir is a fresh string from the first nonempty TMPDIR, TMP, TEMP, or /tmp. It removes one final slash when length exceeds one. Nine Node comparisons cover precedence, emptiness, root/multiple slashes, relative paths, Unicode and malformed UTF-8. macOS follows the same POSIX logic but was not executed.

Native directory errors already use own name/message/code slots. Numeric exit/exitCode validation now also uses own string ERR_OUT_OF_RANGE and Node's exact RangeError message. The independent typed-IR runtime oracle invokes actual ENOENT, ENOTDIR, ELOOP, ENAMETOOLONG and EACCES under uid 65534, plus NaN, both infinities and fractional exit/exitCode. Both backends agree and own-property checks are true. It proves the runtime, not source lowering. ENOMEM and EIO mappings are implemented but not induced; missing-cwd error narrowing is not yet source-proven. The .a source test catches unknown and uses typeof, non-null and in narrowing; it currently records the explicit in-language refusal. TestNodeProcessErrorNarrowingBlocker is a blocker assertion, not an acceptance pass.

The unchanged assigned acceptance fixtures 14, 15, 16, 17, 18, 19, 20 and 23 are all Node Agrees / native NotYet / JavaScript NotYet, first stopped at node:fs.mkdtempSync in their shared scratch setup. All eight Node mutants are caught. No shared fs implementation was fabricated in this unit. Prior undefined! and optional-function-value blockers remain downstream of this dependency.

Latest requested adaptation cf976e8 will be merged before the final error contract test and landing rebase. Standing rule: only codex/host-process-land is landed after rebase/re-green; never push main or force-push.

Earlier reports below are historical; newer evidence supersedes the stages and missing integrations they describe.

Built pinned-Node performance objects and tsc hook discovery; process error narrowing with `in` remains a language blocker.
Commits: shared loader merged as 6bc30e2; performance checkpoint follows this report.
Commands: performance Node/native/JavaScript oracle PASS 2.768s; eight clean native semantic mutants caught; count run 14 allocations/14 frees.
Mutants: now sign, epoch range, mark duration, measure sign, named/all mark deletion, measure clearing, hook-object epoch; all caught by Node comparison.
Not covered at this checkpoint: unchanged assigned System acceptance fixtures, timers, scheduled nextTick, os.tmpdir and unknown-error narrowing.

## Performance checkpoint, pinned Node declarations

Merged 69c71d5. Installed official @types/node 25.3.3 with npm under stage3/api/node_modules; dependency files are not committed. There are no private node declarations. The shared prelude adapter also replaces the old process constant with NodeJS.Process while checking Node imports, avoiding conflicting global declarations.

node_process_performance.a and node_process_performance_core.a pass both backends with sanitizers and leak checking. The latter uses tsc performanceCore hook discovery with its require supplied by a static namespace import. Type-only adaptations make timeOrigin readonly and optional properties explicitly include undefined; object truthiness becomes explicit undefined checks. These are documented in the fixture, not claimed as an unchanged upstream acceptance program. Generic interface calls pad omitted builtin arguments natively, release ignored entry returns, and preserve Node method receiver behavior in JavaScript. Missing-mark SyntaxError name/message is catchable, resolving the end mark first. Unbuilt getEntries and measure options refuse by member name.

Machine-dependent clocks remain predicate comparisons: now is finite, nonnegative and nondecreasing; timeOrigin is stable epoch milliseconds. Mark startTime and measure duration are finite and tied to this clock, not byte-equal to another process; duration can be negative. clearMeasures has no observable registry in the implemented non-enumerating surface.

Latest error contract: native directory errors already have own string code/message fields. Numeric exit validation still needs ERR_OUT_OF_RANGE, and the requested typeof/non-null/`'code' in error` source test is blocked by the explicit compiler refusal of `in`. One-line reproducer: `function code(error: unknown) { if (typeof error === "object" && error !== null && "code" in error) return error.code; }`. No compiler-policy exemption was introduced. This checkpoint does not claim that test passes.

Standing landing rule acknowledged: rebase current origin/main into codex/host-process-land, re-green the oracle, push only that branch; never push main or force-push.

Earlier reports below are historical evidence and limitations; this checkpoint supersedes their shared-loader and missing-mark-panic statements.

Built env string writes/deletion, chdir cache invalidation and catchable directory errors with code/message; performance.measure failures still panic and cannot be caught.
Commits: fixtures merge 5728bc7, runtime/harness 6272a910c986ae5945067f70cef0c31573c43f54 on codex/host-process; earlier implementation ad56ba3 and declaration correction cb90701 retained.
Commands: assigned-fixture harness reports 8 Checker/Checker; filtered process oracle PASS 6.642s, lower PASS 8.344s, affected-package vet PASS; exact logs retained.
Mutants: all 8 acceptance source mutants caught on Node; 4 additional clean native runtime mutants caught by Node stdout; earlier 20-mutant proof remains historical pending shared types.
Not covered: shared @types/node 25.3.3 integration, accepted native/JS fixture execution, timers, scheduled nextTick, Date-valued System.now and full native tsc proof.

## Assigned acceptance fixtures, October 7

Merged origin/codex/stage3-fixtures-host at 1037217. The merge also brings newer main code; the two conflicts were resolved by preserving both process/user-method dispatch hooks and both sets of count rows. Acceptance source, README, check.py and status.json were not changed.

| Fixture | Native stage | JavaScript stage | Recorded Node |
| --- | --- | --- | --- |
| 14_getCurrentDirectory | Checker | Checker | Agrees |
| 15_getExecutingFilePath | Checker | Checker | Agrees |
| 16_getEnvironmentVariable | Checker | Checker | Agrees |
| 17_write | Checker | Checker | Agrees |
| 18_exit_0 | Checker | Checker | Agrees, exit 0 |
| 19_exit_1 | Checker | Checker | Agrees, exit 1 |
| 20_exit_2 | Checker | Checker | Agrees, exit 2 |
| 23_newLine | Checker | Checker | Agrees |

No assigned acceptance fixture is green yet. All eight stop at missing node: module declarations, before lowering. Their unchanged sources agree byte-for-byte with status.json stdout/stderr/exit. Each of the eight mutations copied from check.py changes the recorded Node observation. Exact native and JavaScript compiler diagnostics and stages.json are in acceptance/.

The shared pinned Node hook remains an external dependency. The public fs_file tip inspected was 080789f, which still uses private node_fs_file.d.ts declarations, not the corrected @types/node 25.3.3 hook. It was not merged, and no replacement hook or private declaration was built. A request for the corrected hook SHA is pending. These fixtures also depend on shared scratch-scaffolding members mkdtempSync, mkdirSync, rmSync, tmpdir, and path helpers. Integration and trusted declaration identity for static node: imports must be adapted when the hook arrives. Broader declaration members must then be held to named NotYet refusals.

Confirmed language blockers, reported immediately:

- Fixture 14 contains callback = undefined!, which Adamic 0.1 refuses. Reproducer: `let callback = (): string => ""; callback = undefined!;` prints Refused for the non-null assertion.
- Fixtures 18, 19 and 20 use an optional parameter in an object method/function value. Reproducer: `const nodeSystem = { exit(code?: number): void { process.exit(code); } }; nodeSystem.exit(0);` prints NotYet for a function value with an optional parameter.

These reproducer outcomes are separate from the fixtures' current Checker stages. No language-wide assertion exemption or optional-parameter calling convention change was introduced.

## Independent implementation progress

String env assignment preserves the assigned value, converts native environment text to NUL-terminated UTF-8, and deletion returns true. Only recognized external process.env deletion bypasses the fixed-object-shape refusal; ordinary delete remains Refused. Non-string assignment/key forms remain unsupported. Source lowering is pending declaration integration.

Successful process.chdir releases and invalidates the native cwd cache, while separately held memoized cwd strings stay alive. cwd and chdir now use the existing pending-exception word and native cleanup paths; functions/closures propagate their throws. Directory Error objects expose name, message and code. ENOENT, ENOTDIR, EACCES, ELOOP, ENAMETOOLONG, ENOMEM and EIO are mapped; unknown errno deliberately panics. Only successful directory changes and ENOENT chdir messages were held to Node in the new runtime proof. Other mapped messages and macOS execution remain unverified. errno/syscall/path fields are not built. Missing-mark performance errors remain panic-based and catches remain NotYet.

Two explicit-IR runtime/backend tests bypass declaration loading and compare native plus generated JavaScript to the corresponding unchanged Node operations. They do not prove source checker/lowering acceptance. Environment tests cover empty, zero-string, Unicode and deletion. Directory tests cover cwd invalidation after two chdir calls, independent memoized strings, and caught ENOENT code/message. The Node code observation in the directory source is its ENOENT message prefix; the native typed-IR test reads the code field directly. ASan/UBSan and leak checks pass. Four mutants disable setenv, disable unsetenv, omit cwd invalidation, and change ENOENT code; each compiles/runs cleanly and fails only the Node stdout comparison. New source-fixture count registration remains pending the shared declarations.

Commands run with output redirected to files:

- `go build -o /tmp/host-process-adamic ./cmd/adamic`: PASS.
- `python3 stage3/host-process/check_acceptance.py --compiler /tmp/host-process-adamic --logs stage3/host-process/acceptance`: exit 1 intentionally reports unmet acceptance, all 8 Checker on both backends; Node and 8 mutant comparisons pass.
- `go test -count=1 -timeout 10m ./internal/oracle -run '^TestProcess|^TestNodeProcess(Environment|Directory)Runtime$' -v`: PASS, 6.642s.
- `go test -count=1 -timeout 10m ./internal/lower`: PASS, 8.344s.
- `go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle`: PASS, empty log.
- `git diff --check`: PASS.

No full gate pass is claimed: Node fixture declaration checks remain blocked. Deadlines received are Oct 7 06:00 MDT for the shared loader and Oct 8 01:00 MDT for acceptance green. Runtime progress was pushed immediately; no acceptance-green SHA can be supplied yet.

## Historical report at 3555c7f

The following records the earlier census-only implementation and tests. Its cwd panic/refusal statements are superseded by the catchable-directory work above; its environment read-only boundary is superseded by string mutation support. Its private declarations remain removed.

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
