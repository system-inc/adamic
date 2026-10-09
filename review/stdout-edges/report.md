Built Node-compatible stdout edge handling, panic encoding, spread diagnostics and child permission checks.
Commits: claims 7f0d3cd; initial implementation 872c2e9; sanitizer follow-up on codex/stdout-edges, from origin/main 5d4c801.
Checks after review: native 71.081 s; all output tests uncached 34.860 s; both passed; formatting clean. Initial full oracle passed in 95.240 s.
Mutants: original edge mutants all caught; reinstating the SIGSEGV flush handler also fails the sanitizer stack-report test.
Not covered: other operating systems, Node's SIGUSR1 inspector, SIGKILL flushing, signals inside an in-flight flush, concurrency.


Review correction: adamic_start leaves SIGSEGV, SIGBUS, SIGILL, SIGFPE, SIGTRAP, SIGSYS and SIGABRT at their existing dispositions. It also leaves platform-specific fault signals STKFLT and EMT alone. The flush handler remains on TERM, INT, HUP, QUIT, USR1, USR2, ALRM, VTALRM, PROF, XCPU, IO, PWR and public realtime signals. This preserves sanitizer handlers and their stack evidence.

TestStartupPreservesSanitizerSegvReport compiles a C harness with native.Build's ASan/UBSan flags, calls adamic_start, then writes through a null pointer. Only the crash function's UBSan checks are disabled: otherwise UBSan exits on the null store before the hardware fault reaches ASan's handler. ASan remains enabled, and the test requires its SEGV report and a crash stack frame. Before the correction it failed with exit -1 and empty stderr; after the correction it passed with exit 1 and a symbolized ASan SEGV stack. The sanitizer-segv mutant in mutants.py puts SIGSEGV back on the flush-handler list and fails the same assertion. Workspace logs: logs/sanitizer-before.log, logs/sanitizer-after.log and logs/sanitizer-mutant.log.

Logs are workspace artifacts only, under /workspace/adamic/review/stdout-edges/logs. All 30 previously tracked log files are removed from the branch's final diff and remain on disk. The already-pushed 78c6561 commit still contains them: CLAUDE.md prohibits rewriting history and force-pushing, so this removal is a forward commit. report.md, mutants.py and benchmark.py stay tracked. Log references below resolve in this workspace; the logs are not part of a fresh checkout of the final tree.

Node 24.19.0 on Linux is the oracle for these observations. The Linux-specific descriptor, resource-limit, signal and PTY tests are build-tagged accordingly. Optional signal names in the runtime are guarded, but their behavior on macOS or other systems was not measured. Node's handling of terminals, pipes and signals depends on the platform; this report does not claim the Linux observations for Windows or macOS.

The first commit claimed the requested territory and was pushed before implementation. The implementation changes adamic.c, lower/input.go, input_test.go, new tests and fixtures, and four required counts rows. input.c needs no production edit: its create mode already is 0666. adamic_panic changes only the message write to the existing surrogate-converting path. There are no new runtime threads or locks, and the protected compiler files are unchanged.

The output comparison tests run source on Node before comparing native and the JavaScript backend. Finite fixtures also join the ordinary oracle, including release builds, ASan/UBSan, leak checks for successful programs and recorded counts. Python's standard-library POSIX helpers set limits and signal dispositions only in children, and open a PTY with pty.openpty. Python runs in isolated mode so inherited optimization settings cannot disable its assertions.

| Item | Observation before the fix | Result after the fix |
|---|---|---|
| 1, SIGXFSZ | At a 102400-byte file limit, writeTextFile writes that many bytes then native dies of signal 25, with no result or stderr. Node returns `cannot write <dir>/big.txt: failed`, prints `after`, and exits 0. For stdout itself, native dies of 25 while Node exits 70. | Ignore SIGXFSZ as Node does. Both file-write runs exit 0 with identical result and stderr; both stdout runs exit 70 with identical 102400 bytes and stderr. |
| 2, EAGAIN | Slow non-blocking stdout: Node delivers 1660112 bytes and exits 0; native delivers 65520 bytes and exits 70 in this run. | poll for POLLOUT and retry through EINTR and partial writes. All three deliver 1660112 identical bytes and exit 0. |
| 3, closed descriptors | Closed stdout gives native exit 70 versus Node 0. | Open /dev/null on closed 0, 1 or 2 before any program file can occupy those numbers. Individually closed stdin, stdout, stderr, and all three closed give matching bytes and exit 0. |
| 4, inherited ignored signals | With SIGINT inherited ignored, Node terminates by INT with its 39-byte line; native remains running until the test deadline. | Reset inherited INT, HUP and TERM. All three produce the same line and terminate by the sent signal on Node, native and the backend. The comment now describes the reset. |
| 5, fatal signals | SIGUSR2 kills both, but native loses its buffered 39-byte line. | Flush for external stop signals, then restore the default action and re-raise. The revised probes compare 11 non-realtime signals (USR1 has Node inspector behavior), and all public realtime signals 34-64 here, preserve the line and retain the signal status. A soft CPU limit also gives SIGXCPU with the same line. SIGPIPE and SIGXFSZ remain ignored. |
| 6, panic | stdout substitutes U+FFFD, but panic writes ED A0 80 on stderr. | Both streams write EF BF BD and match Node byte for byte, exit 70. |
| 7, tuple spread | The checker accepts writeTextFile(...pair); lowering blames the checker for the argument count. | NotYet names `a SpreadElement` at the exact spread location, as readTextFile(...one) does. Node itself accepts both calls: write gives Ok and bytes `text`; read of the missing path gives Error. |
| 8, umask | Under umask 022, the 0644 create-mode mutant passes the old writing-fixture harness. | A child shell sets umask 0 before exec, for Node and native alike. The same mutant now fails the file snapshot: Node's 0666 versus native's 0644. The Go process's umask is untouched. |

A signal handler calls only the POSIX async-signal-safe write/poll path, sigaction and raise. Its default sigaction is prepared at startup rather than initialized by a possible memset inside the handler. The handler retains the existing whole-line snapshot scheme. No claim is made about a signal interrupting an in-flight flush: that existing window still clears the buffer before the write completes. SIGKILL cannot be caught, so pipe buffering still loses pending bytes then.

The supplied fsize_out.a on this Node version emits `after\n` on stderr and exits 70; it does not emit an additional error message. This is what the tests compare, rather than assuming the integration reading's message.

SIGUSR1 is a separate observed exception: Node stays running and starts its inspector, printing a debugger URL. The runtime registers a flushing fatal handler for its default terminating action, as requested, but does not implement Node's inspector. SIGUSR1 is consequently excluded from the terminating-signal Node comparison. The direct observation is logged (`logs/node-usr1.log`).

The timing test calls the actual native runtime line writer from a small C driver, compiled through native.Build under ASan/UBSan, ten times 150 ms apart. Node's driver uses console.log and its own timer. This avoids assuming that equal computational loops take equal time on the two runtimes; Adamic's library has no timer. Under a PTY, every line arrives separately. The terminal mutant holds all ten until about 1.503 seconds. The same test with a regular file observes Node's lines as they are written; before the fix all native lines arrive together at about 1.514 seconds.

I chose to match Node for regular files as well as terminals. fstat selects line flushing for a regular file; pipes keep the 64 KiB buffer. A controlled benchmark used two release binaries differing only in that regular-file selection, stdout a fresh regular file, nine interleaved rounds with alternating order. Every run, Node included, wrote the same 68 bytes in six lines from bench/word_count.ts.

| Runtime | Best | Median |
|---|---:|---:|
| Native, regular file buffered | 140.261 ms | 149.821 ms |
| Native, regular file line flushed | 143.030 ms | 148.814 ms |
| Node 24.19.0 | 259.628 ms | 267.475 ms |

The best is 2.0% slower; the median is 0.7% faster. The difference is within the observed run-to-run noise. This measures this six-line benchmark, not a workload printing millions of lines. Live file observability is the reason for the decision. Raw rounds (`logs/benchmark.log`) and [benchmark driver](benchmark.py) are saved.

Every requested mutant was run alone with source restored afterward. Each failed at runtime or at the diagnostic assertion, not at C compilation. The permission mutant was also run under the old harness and shown to pass, then caught under the fixed harness. [The repeatable runner](mutants.py) writes one log per mutant and restores files even if a run fails.

| Mutant | Catching test and evidence |
|---|---|
| 1: omit SIGXFSZ ignore | TestOutputEdges/fsize and fsize_out: signal 25 versus Node exits 0 and 70. log (`logs/mutants/1-sigxfsz.log`) |
| 2: omit EAGAIN retry | TestOutputEdges/nonblock: 65520 bytes, exit 70 versus 1660112 bytes, exit 0. log (`logs/mutants/2-eagain.log`) |
| 3: omit standard-descriptor repair | TestOutputEdges/closed: closed stdout exits 70 versus 0. log (`logs/mutants/3-closed.log`) |
| 4: preserve inherited SIG_IGN | TestOutputEdges/ignored: INT does not stop native; deadline 124 versus Node signal 2. log (`logs/mutants/4-inherited.log`) |
| 5: install only TERM, INT and HUP handlers | TestOutputEdges/signals: SIGUSR2 preserves 0 native bytes versus Node's 39. log (`logs/mutants/5-fatal.log`) |
| 6: panic message written raw | TestOutputEdges/panic: ED A0 80 versus EF BF BD on stderr. log (`logs/mutants/6-surrogate.log`) |
| 7: restore the argument-count error before spread handling | TestInputTupleSpreadsAreNotYet: internal checker-blaming error instead of NotYet. log (`logs/mutants/7-spread.log`) |
| 8: create with 0644 | TestInputAgreesWithNode/internal/oracle/testdata/write_files.a: file permissions differ. old harness passes (`logs/mode-old-harness.log`), fixed harness fails (`logs/mutants/8-mode0644.log`) |
| Force terminal output to buffer mode | TestOutputLinesArriveWhileRunning/terminal: all arrivals at 1.503 s. log (`logs/mutants/terminal.log`) |
| Restore regular-file buffer mode | TestOutputLinesArriveWhileRunning/file: all arrivals at exit instead of 150 ms apart. log (`logs/mutants/file.log`) |
| Omit realtime signal handlers | TestRealtimeSignalsLeaveWhatWasPrinted/34: correct signal status but 0 native bytes versus Node's 39. log (`logs/mutants/realtime.log`) |

The initial Node-inclusive failing runs are saved in edges-before (`logs/edges-before.log`), corrected signal and file probe (`logs/edges-before-corrected.log`), lower-before (`logs/lower-before.log`), and realtime-before (`logs/realtime-before.log`). The initial probe included SIGUSR1 as terminating; Node disproved that assumption, so the corrected probe excludes it and the inspector observation is recorded separately. Final checks below use the corrected probes.

Toolchain: bash cloud/setup.sh completed successfully. Its timing lines were Go ready 0 s, clang ready 0 s, Node ready 0 s, submodules ready 0 s, build cache warm 78 s, total 78 s. nproc was 5, with cgroup CPU quota 400000/100000. Versions were Go 1.27.1, clang 20.1.8, Node 24.19.0. Each shell sourced /workspace/adamic-tools/env.sh. Setup log (`logs/setup.log`).

Commands and outputs from the initial full verification:

```text
gofmt -l cmd internal > /tmp/stdout-gofmt.log
  empty output
go vet ./... > /tmp/stdout-vet.log 2>&1
  exit 0, empty output
go test -count=1 -timeout 30m ./internal/native/... ./internal/lower/... > /tmp/stdout-packages.log 2>&1
  ok github.com/system-inc/adamic/internal/native 76.376s
  ok github.com/system-inc/adamic/internal/lower 9.585s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -v > /tmp/stdout-oracle.log 2>&1
  PASS
  native hits=0 misses=1046
  node hits=0 misses=581
  probe hits=0 misses=24
  ok github.com/system-inc/adamic/internal/oracle 95.240s
```

Full oracle log (`logs/oracle.log`), package log (`logs/packages.log`), counts update log (`logs/counts.log`). Counts were regenerated with `go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts`; only the four new fixture rows were added, and the initial full uncached oracle rechecked them. Test output always went to files, never through a pipe.

I did not run `go test ./...`; I ran the requested native/lower packages and the whole oracle, plus `go vet ./...`. Thread safety belongs to codex/concurrency and is unchanged here. The output signal probes deliver signals from outside. The new null-write harness separately proves preservation of ASan crash diagnostics after startup; it does not prove recovery from faults. The timed C driver proves the runtime flushing decision; the ordinary fixtures separately prove compiler output dispatch.

Review follow-up verification (logs stay in the workspace):

```text
go test -count=1 -timeout 30m ./internal/native/... > /tmp/stdout-revision-native.log 2>&1
  ok github.com/system-inc/adamic/internal/native 71.081s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestOutput|TestStartupPreservesSanitizerSegvReport|TestRealtimeSignalsLeaveWhatWasPrinted|TestASignalLeavesWhatWasPrinted|TestClosedStdoutEndsAsOnNode|TestOneFileHoldsNodesOrder|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead' -v > /tmp/stdout-revision-output.log 2>&1
  PASS
  native hits=0 misses=0
  node hits=0 misses=0
  probe hits=0 misses=11
  ok github.com/system-inc/adamic/internal/oracle 34.860s
python3 review/stdout-edges/mutants.py sanitizer-segv > /tmp/stdout-sanitizer-mutant.log 2>&1
  sanitizer-segv caught by TestStartupPreservesSanitizerSegvReport$ exit 1
```

Copies are at logs/revision-native.log and logs/revision-output.log under the workspace directory given above. The new sanitizer test passed in the final output run too. The full oracle and lower tests were not rerun for this review correction; the requested native package and all output tests were rerun. No log files remain tracked.
