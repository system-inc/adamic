# Node host runtime on wasm32-wasi

Baseline: `c762b555e7c0b39b105e2d9208732a5a311b2e6f`, Linux, Node v24.19.0,
Go 1.27.1, clang 20.1.8, WASI SDK 27. `nproc` reports 5 (CPU quota 4).

## Behavior and audit

The three Node host translation units compile for WASI. Available libc file,
directory, allocation, UTF-8, environment, clock, output, exit, and fsync
operations remain real operations. The WASI parent-path adapter preserves
symlink-before-`..` lookup and leaf identity; the runner supplies its actual
working directory through `ADAMIC_WASI_CWD` and `chdir`.

Unsupported members produce typed build refusals when generated C identifies
them: mkdtempSync, pid, platform, argv, stdout.columns, and memoryUsage.
Their runtime entries also panic explicitly if called indirectly. Custom file
and directory permission bits and unavailable timestamp precision/range refuse
at runtime. Default modes and nonnegative whole-second timestamps work.
The oracle records these exact refusals as target skips; other build failures,
traps, and panics still fail. It compares filesystem effects as well as streams.

`realpath`, `readlink`, `utimensat`, open/read/write/close, stat/lstat/fstat,
opendir/readdir/closedir, mkdir/unlink/rmdir, fsync, isatty, getenv/environ,
clock_gettime, chdir/getcwd, allocation and byte functions were checked against
SDK headers and runnable witnesses. getpid, executable-path readlink,
ioctl(TIOCGWINSZ), and allocator observations are native-only. Darwin malloc.h
remains Darwin-only. symlink creation, chmod, and getrusage are not called by
these host files; they were not replaced by invented behavior.

Compiler hooks: `internal/native/native.go` changes one Build validation call.
`internal/native/runtime/directory.c` adds one shared-adapter include so the
existing fileStatus bridge uses the same parent-path semantics. No edits to
emit.go, lower.go, or oracle_test.go.

## Setup

`GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh --wasi-sdk`
completed. Cold setup elapsed lines (seconds): node 0.209, Go 0.265,
clang 0.837, markdown 1.559, submodules 7.543, SDK 11.503,
Go build 698.895, warm 699.085, done 699.151. A repeat completed in
57.223 seconds including installer-lock waiting. Environment:
`source /workspace/adamic-tools/env.sh`. `npm ci --prefix stage3/api`
installed three packages in 14 seconds. Logs: `/tmp/wasi-setup-direct.log`,
`/tmp/wasi-setup-final.log`, `/tmp/wasi-npm.log`.

## Verification

All commands write test output directly to logs. Compressed evidence is committed beside this report as `node-host-*.log.gz`; temporary paths below identify the original logs.

- Full WASI oracle and Linux counts: `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -v -timeout 30m -args -update-counts`, `/tmp/wasi-oracle-final.log`: all seven WASI test groups pass; counts pass (380.50s). The complete package exits 1 (1324.573s) only for the two independently reproduced baseline Node pipe-loss assertions below.
- Full native oracle: `go test ./internal/oracle -count=1 -timeout 30m`, `/tmp/wasi-native-oracle-final.log`: fails (710.632s) only TestProcessExitDrainsStderr and TestProcessLargePipeExitPreservesOutput. Raw Node emits all 204800 bytes; these tests require only 4096. Adamic preserves all bytes. Both raw-Node assertion failures reproduce on untouched c762b55 with `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestProcess(ExitDrainsStderr|LargePipeExitPreservesOutput)$' -count=1 -v -timeout 10m` (30.438s), `/tmp/wasi-pipe-baseline.log`. A complete native retry exits 1 (381.463s) with the same two assertions plus a long-regex 3-minute timeout while both gates run concurrently, `/tmp/wasi-native-oracle-retry.log`. That regex test passes in the first native run and the full WASI-enabled run; the timeout is reported rather than hidden.
- `go test ./internal/native ./internal/flow -count=1 -timeout 30m`, `/tmp/wasi-native-flow.log`: native passes (403.711s); flow fails on process_bad_code.a with five identical mid-basic-block call-end diagnostics. The exact filtered flow test on the untouched baseline reproduces all five diagnostics (1.010s), `/tmp/wasi-flow-baseline.log`. This pre-existing failure is outside this unit.
- `go vet ./...`, gofmt, and git diff --check pass, `/tmp/wasi-vet.log`, `/tmp/wasi-format.log`, `/tmp/wasi-diff-check.log`.
- Strict WASI compilation covers all 57 runtime C files. Native objects for node_fs_file.c, node_fs_directory.c, node_process.c, and directory.c are byte-identical to baseline at clang -O2, `/tmp/wasi-native-objects.log`.
- A direct Node/WASI round trip verifies UTF-8 write/read, flush/fsync, whole-second utimes, stat mtimeMs and unlink, `/tmp/wasi-whole-second.log` and `/tmp/wasi-whole-second-node.log`.
- The stock wasi-agreement/check.py compiles the runtime and passes its positive comparisons but its historical adamic.h mutant lacks the newer process_status declaration. A scratch copy adds only that declaration to the scratch mutant header; all five separate-stream comparisons, two combined-stream comparisons, and three independent mutants (five witnesses) pass. Logs: `/tmp/wasi-agreement.log`, `/tmp/wasi-agreement-compatible.log`. Repository agreement code is unchanged.

Linux counts change process_observations.a and node_fs_directory_system.a;
other directory rows already match Linux. The full WASI pass also regenerates
counts, rather than relying only on a filtered counts run.

## Mutants

- Remove the mkdtemp WASI guard: the WASI dedication oracle fails compiling the runtime with undeclared mkdtemp. `/tmp/wasi-build-mutant.log`. This is the requested build sentinel, not a semantic mutant.
- Disable target-refusal lookup through a Go overlay: all six member assertions fail because they receive no refusal. `/tmp/wasi-target-mutant.log`.
- Nine independently compiled runtime mutants replace each explicit refusal by a successful fake return: mkdtemp, pid, platform, columns, memoryUsage, argv, custom open mode, custom mkdir mode, and fractional utimes. Every exact exit-70/reason check catches its mutant. `/tmp/wasi-host-mutants.log`.
- Bypass parent-path resolution: compilation and execution succeed, but the symlink/.. directory listing disagrees with Node. The independent directory witness catches it. Same mutant log.
- Existing agreement mutants: restore throwing-sort refusal, remove stack-depth enforcement (two witnesses), and remove stdout/stderr descriptor routing (two witnesses). Their exact stream/status comparisons catch them after the historical-header compatibility adaptation.

- Existing WASI oracle mutants change dedication output bytes and return exit 23; Node comparison catches both. The standalone WASI runner control also catches output-byte and exit-23 mutants. Both registered mutant tests pass in the full run.

## Limits

No claim of complete WASI host equivalence: process identity, permission bits,
temporary directories, terminal size, executable paths, allocator observations,
and some timestamp writes are explicitly refused. Specialized native tty and
unlinked-working-directory drivers are not extended to WASI. Native sanitizers
remain covered by the existing native oracle; WASI SDK does not support those
flags. Flow and the two raw-Node pipe-loss assertions remain red on independently reproduced baseline defects. No whole-package green claim is made.
