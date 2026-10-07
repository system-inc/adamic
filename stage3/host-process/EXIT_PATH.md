Built immediate process-exit policy: abandon live values without unwinding, and judge leak applicability from the executed exit path rather than status.
Base library/merge-p2-trial c762b555; branch codex/host-process-exit-land; the earlier host-process-land branch is unchanged.
Commands: uncached process/cwd subset PASS 17.057s; Linux counts regenerated PASS 38.832s; native PASS 145.759s; whole oracle PASS 394.047s; final input/policy subset PASS 6.898s; vet PASS.
Mutants: dropping the exit-path marker and dropping the normal closure release each run cleanly and fail only counted memory checks; FIFO output-tail truncation fails only Node stdout on source and both backends; pipe no-drain and process mutants are caught.
Not covered: macOS execution or unrelated compiler-front work; the macOS counted-check algorithm is exercised directly on Linux.

## Decision

Process.exit is immediate termination, like panic: flush output, report counts, and stop without frame unwinding, finally clauses or global cleanup. A releaseGlobals callback inside a live closure would not release stack-held values and does not promise complete cleanup. docs/memory.md names the rule beside Exceptions.

The actual native process-exit implementation marks only counted builds with a termination process_exit suffix on the final counts line. Normal completion keeps its original counts format. The oracle checks this runtime evidence, never infers immediate exit from zero status or a static source call. Both platforms run the counted balance test; Linux additionally runs LeakSanitizer and macOS leaks --atExit for normal completion. Immediate exit skips those tools after its counted classification. Counts still honestly report allocations held at the stopping point. Invalid exit arguments that throw do not reach the marker and remain ordinary cleanup paths.

TestProcessExitPathOwnsLeakPolicy uses the exact zero-status never-returning arrow, compares source Node and both backends, asserts its counted build has an actual live allocation, and proves exemption. It also tests exit 7. A normal-return control contains the same exit closure but never calls it; it must balance. The marker-removal mutant leaves stdout, status and allocations unchanged and fails the counted policy check. The normal-release mutant preserves stdout and status and fails counted leak detection. Thus ordinary leaks remain errors even if exit appears in the program. Input fixtures use the same counted path decision; the explicit-input-exit test supplies a deliberately nonexistent leak-tool binary and proves it is bypassed only after runtime classification.

## Front runner and cached cwd

The front already sets stdout/stderr handles blocking and makes panic exit at once; those changes are retained. The original cached-cwd runner passes 24/24 parallel observations on Linux. However, its Python driver's forced read-ahead control still loses cached true in all 24 runs: stdout.readline buffers bytes that communicate never reads. Blocking stdout does not fix Python read-ahead or verify readTextFile('/dev/stdin') succeeded. The FIFO rendezvous and continuously draining readers from c80c8bd are therefore carried onto this tree, without importing the older compiler/library branch. The fixed stress test compares 24 concurrent runs each of plain Node, source Node, native and generated JavaScript, and catches the dropped-tail driver mutant only by stdout comparison.

The blocking runner also invalidated two older tests' assumptions that runner output loses queued bytes. They now require full output against the runner while preserving an independent stock-Node pipe-loss control. The no-drain JavaScript mutant explicitly turns blocking off before writes, then exits without draining, so the byte comparison still kills it. Normal-return tests and native flush/order mutants remain intact. Panic count reports retain their existing format.

## Verification

The pinned TypeScript submodule initially lagged the requested cohere revision, causing shim build failures. The exact trial submodules are now checked out, without changing repository submodule pins. Logs retain that initial setup failure and successful synchronization.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestProcess|^TestNodeProcess(CwdBarrier|CachedDirectory)$' -count=1 -timeout 30m -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/native -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 30m
go vet ./internal/oracle ./internal/native
```

The whole oracle passed before the final input-harness extension; all input comparisons and the extended exit-policy test passed afterward. go vet passed against the final files. Linux counts were regenerated. Logs are in logs/exit-path/. No main branch was pushed or force-pushed.

The first broad process run caught stale pipe expectations from the front's blocking change. A second caught a mutant that turned blocking off only after writing; moving that change before writes made it effective. An initial native run caught the unnecessary change to panic count formatting, which was removed. Final logs distinguish those corrected attempts from final gates.

macOS confirmation with official Node 24.19.0:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestProcessExitPathOwnsLeakPolicy$|^TestNodeProcess(CwdBarrier|CachedDirectory)$|TestNativeAgreesWithNode/internal/oracle/testdata/process_exit' -count=1 -timeout 30m -v > /tmp/process-exit-path-macos.log 2>&1
```

Expect the same counted exit-path decision and complete cached-cwd observations. No macOS success is claimed by this Linux worker.
