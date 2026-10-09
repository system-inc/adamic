Merged stdout edge handling into the runtime area lineage with native-only POSIX guards.
Commits: merge f1f2590070b6dfe4a92305a33b69dd369505903e; counts order 94583ffb828923abf69007f89f6170b4ebaa2d41.
Checks: native 645.339 s; TestWASI 192.191 s, 35/35; full Wasm oracle 913.388 s, 415/415; final full uncached native oracle 304.364 s, PASS.
Mutants: all 13 individual behavioral reversions were caught by the named assertions after successful compilation.
Not covered: other operating systems, Node inspector behavior, SIGKILL flushing, signals during an in-flight flush, concurrency.

The branch is codex/stdout-edges-area, based on origin/area/runtime f955656353315db75e86043d3f550e0c4dff11de. It merges origin/runtime/stdout-edges-usr1-counts 5f15c17c8750a5f52c5f086e13e8fc0bb61b8a0e. Freshly fetched origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 is already an ancestor through the area base. No main or area branch was changed or pushed.

The merge commit explains the two conflicts. In adamic.c, fcntl/poll/signal/stat includes, EAGAIN polling, standard-descriptor repair with /dev/null, signal state and handlers, startup dispositions and external/realtime stop lists, and regular-file line flushing are native-only. WASI keeps its argument setup and existing terminal-versus-buffered output selection. The two portable incoming changes remain on both targets: a zero-byte write fails rather than spinning, and panic messages use the existing lone-surrogate replacement path. The latter makes the new panic_surrogate fixture agree with Node on Wasm as well as native.

Independent preprocessing verified that native C is identical to the incoming branch after ignoring blank lines introduced by guards. WASI preprocessing contains no new pollfd, default_action, /dev/null, stop_with, SIGUSR1 or F_GETFD paths. Compared with the area base, its only executable differences are the zero-write check and panic message encoding. Scratch evidence is /tmp/stdout-edges-preprocessed, including wasi-area.diff.

The counts conflict preserves every newer area row and imports the five output-edge rows. The first full uncached oracle passed all comparisons and probes but failed the final table byte comparison because the five rows were in the wrong registration order. Its diagnostic listed no changed measurement. Regenerating with TestCountsAreRecorded changed only those five rows' position. A map comparison of all 420 fixture rows confirmed every number was identical; no existing fixture count rose. The final full uncached rerun passed, including the exact table comparison, all output probes and all ordinary fixtures.

| New fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| fsize.a | 8 | 8 | 3 | 12 | 5 | 0 |
| fsize_out.a | 60000 | 60000 | 0 | 60000 | 3 | 0 |
| closed.a | 0 | 0 | 0 | 0 | 0 | 0 |
| panic_surrogate.a | 3 | 1 | 0 | 2 | 2 | 0 |
| usr1.a | 2 | 2 | 0 | 2 | 2 | 0 |

Observed Linux edge results: at RLIMIT_FSIZE=102400, file writing returns the same failed-write result, prints after and exits 0; stdout produces the same 102400 bytes, prints after on stderr and exits 70. Slow nonblocking stdout delivers all 1660112 bytes and exits 0. Closing each standard descriptor, or all three, matches Node's output and exit. INT/HUP/TERM inherited ignored still stop the program as Node does. External stop signals and public realtime signals 34 through 64 preserve the same 39-byte line and signal status. Timed terminal and regular-file lines arrive while the process runs. Sanitizer startup retains the ASan SEGV stack report.

For SIGUSR1, Node and native both print started, about to work for a while, then finished 166656, and exit 0. Node additionally starts its inspector and writes a debugger URL to stderr; native ignores SIGUSR1 and implements no inspector. TestOutputEdges/usr1 deliberately compares stdout and exit for this observation. The ordinary and Wasm fixtures receive no signal and compare both streams. Panic writes printed U+FFFD first on stdout and adamic: panic: lone U+FFFD here on stderr, with exit 70, on all three ordinary backends and Wasm. No new WASI exception was added.

All commands source /workspace/adamic-tools/env.sh. Test output goes directly to the named log files; logs are workspace artifacts, excluded from commits by .gitignore.

| Command | Result | Workspace log |
|---|---|---|
| bash cloud/setup.sh --wasi-sdk | Go/clang/Node/WASI/submodules ready at 0 s; cache warm and total 266 s; nproc 5; cpu.max 400000 100000 | /tmp/stdout-edges-setup.log |
| go test ./internal/native/... -count=1 -v -timeout 20m | PASS, 645.339 s | /tmp/stdout-edges-native.log |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -v -timeout 40m | Initial run: 636.512 s, only counts row ordering failed | /tmp/stdout-edges-full-oracle.log |
| go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts | PASS, 45.958 s; ordering only | /tmp/stdout-edges-counts-update.log |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -v -timeout 40m | PASS, 304.364 s; native cache hits 0/misses 1657, Node hits 0/misses 906, probe hits 0/misses 25 | /tmp/stdout-edges-full-oracle-green.log |
| PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 go test ./internal/native -run '^TestWASI$' -count=1 -v -timeout 20m | PASS, 192.191 s; 48 strict-C11 units; 35/35 fixtures; stack canary; 100000 requests, memory 393216 stable, live 0, regions 6300000 | /tmp/stdout-edges-testwasi.log |
| ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWASI' -count=1 -v -timeout 40m | PASS, 913.388 s; 415/415 agreement, 415/415 emission; oracle and runner byte/exit mutants | /tmp/stdout-edges-wasm-oracle.log |
| go test ./internal/lower -run '^TestInputTupleSpreadsAreNotYet$' -count=1 -v -timeout 10m | PASS, 0.221 s | /tmp/stdout-edges-spread-control.log |
| gofmt -l cmd internal; go vet ./... | Both exit 0, empty output | /tmp/stdout-edges-gofmt.log; /tmp/stdout-edges-vet.log |

The full native oracle includes TestOutputEdges (fsize, fsize_out, nonblock, closed, signals, ignored, usr1, panic), TestOutputLinesArriveWhileRunning, all 31 TestRealtimeSignalsLeaveWhatWasPrinted cases, TestStartupPreservesSanitizerSegvReport, and the five finite output fixtures by path under TestNativeAgreesWithNode. It is fully uncached, including source-on-Node observations. The Wasm gate includes those same five finite fixture paths, plus every other ordinary fixture. No overlay or expectation patch is used.

The [guard-aware mutant runner](mutants.py) runs in /tmp/stdout-edges-area-mutant-tree, detached at the merge commit, so the unit branch and baseline gates are never mutated. Each case changes one behavior and must fail its named assertion after successful C compilation; compiler warnings and build errors do not count. Logs and exact patches are /tmp/stdout-edges-area-mutants/<name>.log and <name>.patch. Summary output is /tmp/stdout-edges-mutants.log. The initial cold-worktree build hit the old 200-second process limit before executing a test, so it is not counted as evidence; the successful rerun uses an eight-minute Go test limit and 600-second process limit. Its initial timeout log is /tmp/stdout-edges-mutants-cold-timeout.log. After the successful suite, git hash-object matched HEAD for all three temporarily mutated files, confirming byte-for-byte restoration.

| Mutant | Catching assertion |
|---|---|
| usr1: omit SIGUSR1 ignore | TestOutputEdges/usr1: native terminates instead of finishing |
| sanitizer-segv: add SIGSEGV to flush handlers | TestStartupPreservesSanitizerSegvReport: sanitizer SEGV stack missing |
| 1-sigxfsz: omit SIGXFSZ ignore | TestOutputEdges/fsize and fsize_out: native dies of signal 25 |
| 2-eagain: omit poll retry | TestOutputEdges/nonblock: shortened output and exit 70 |
| 3-closed: omit descriptor repair | TestOutputEdges/closed: closed stdout exits 70 |
| 4-inherited: preserve inherited SIG_IGN | TestOutputEdges/ignored: native fails to stop on INT |
| 5-fatal: omit external-stop list | TestOutputEdges/signals: SIGUSR2 loses the buffered line |
| 6-surrogate: write panic message raw | TestOutputEdges/panic: ED A0 80 instead of EF BF BD |
| 7-spread: omit writeTextFile spread refusal | TestInputTupleSpreadsAreNotYet: checker-blaming count error instead of NotYet |
| 8-mode0644: create files with mode 0644 | TestInputAgreesWithNode/internal/oracle/testdata/write_files.a: permission snapshot differs under child umask 0 |
| terminal: buffer terminal lines | TestOutputLinesArriveWhileRunning/terminal: lines arrive only at exit |
| realtime: omit realtime handlers | TestRealtimeSignalsLeaveWhatWasPrinted/34: loses the buffered line |
| file: buffer regular-file lines | TestOutputLinesArriveWhileRunning/file: lines arrive only at exit |

No performance benchmark was repeated: native preprocessed implementation is identical to the incoming unit, and this merge changes conditional compilation rather than its native algorithm. The complete repository go test ./... was not run; requested native/oracle packages, the imported lower check, formatting and repository-wide vet are covered. Linux Node 24.19.0, clang 20.1.8, WASI SDK 27 and Go 1.27.1 were used. Other systems, inspector implementation, SIGKILL buffering, the existing signal-during-flush window and multithreaded callers remain outside this unit.
