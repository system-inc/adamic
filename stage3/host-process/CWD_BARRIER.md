Fixed driver read-ahead loss and replaced the ignored /dev/stdin read with a checked, independent FIFO barrier; macOS's underlying stdin errno remains unobserved on this Linux worker.
Parent 8f76160; changes are confined to the cwd test harness and its source/typed-IR witnesses, without runner or runtime behavior changes.
Commands: cached/removed cwd tests PASS 1.288s; 24 concurrent observations per backend and plain Node PASS 2.570s; uncached process/counts PASS 70.772s; flow PASS 91.822s; vet exit 0.
Mutants: dropping the driver's final stdout chunks runs cleanly and fails only Node stdout comparison on source, native and JavaScript; existing cwd-cache and error-code mutants still fail against Node.
Not covered: macOS execution, the exact macOS stdin errno, or the pending fs-file area merge and full host acceptance rerun.

## Cause and fix

The original Python driver mixed buffered child.stdout.readline() with child.communicate(). readline can read beyond ready into its internal buffer. communicate then uses descriptor reads and never retrieves those buffered bytes. A control child writing ready and cached true in one write loses cached true in all 24 concurrent runs on Linux. This proves an observation bug independent of Node exit draining. The cwd fixture has no explicit process.exit, and normal Node termination can drain queued writes; setting stdout blocking globally would change the oracle rather than fix this driver.

The user's macOS observations additionally establish that readTextFile('/dev/stdin') frequently returns Error. Because the fixture ignored its result, it could continue before Python removed the directory or supplied continue. EAGAIN due to inherited nonblocking state is a hypothesis, not a measured errno here. Node output timing varies by destination and platform; correctness of this rendezvous must not depend on a prompt ready write or a synchronous pipe write.

The new driver drains stdout and stderr continuously with one descriptor reader per stream. It removes the child's cwd only after ready, then opens a separate named FIFO for writing. A fresh synchronous readFileSync/fopen of that FIFO has its own descriptor flags. Opening its writer succeeds only when the child has opened the reader; writing continue and closing produces a definite EOF. Both source witnesses read the FIFO path from programArguments, and check the read kind and payload. The removed-cwd typed-IR witness reads the same supplied path. Waiting for child exit and joining both output readers captures every final byte. Timeouts kill and reap the child; temporary files are cleaned up.

This changes only the test rendezvous input, not the operations being judged: cwd caching, cache invalidation and the removed-cwd own code/message still have their independent Node comparisons. No tsc host acceptance source or language blocker is rewritten. ASan leak detection stays disabled in this driver as before; the incoming Linux-only leak detection changes elsewhere are preserved when merged. The Darwin macro added in 2316c2c remains present.

## macOS confirmation

With official Node 24.19.0, source your configured toolchain and run from the repository root:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNodeProcess(CwdBarrier|CachedDirectory|CwdErrorRuntime)$' -count=1 -timeout 30m -v > /tmp/process-cwd-macos.log 2>&1
```

Expect all three tests to pass. CwdBarrier runs 24 observations concurrently for each of plain Node, oracle source, native and generated JavaScript, compares complete stdout/stderr/exit, and kills the lost-tail mutant only by stdout comparison. CachedDirectory still kills the no-cache mutant after the FIFO handoff; CwdErrorRuntime checks the exact removed-directory error and own code/message.

To determine the precise original stdin failure rather than infer EAGAIN, reproduce it with the old driver and a probe that records the fs error caught by readTextFile:

```sh
git show 8f76160:internal/oracle/testdata/node_process_cached_cwd.py > /tmp/process-cwd-old-driver.py
python3 /tmp/process-cwd-old-driver.py node --disable-warning=ExperimentalWarning "$PWD/oracle/node.mjs" "$PWD/stage3/host-process/probes/stdin-error.mjs" > /tmp/process-cwd-stdin-probe.json 2> /tmp/process-cwd-stdin-probe.stderr
```

The JSON stderr field is base64-encoded. Decode it to obtain the exact code/message and read kind. The probe logs the original caught error then rethrows it, so it does not turn failure into success. This diagnostic is for the old harness only; the fixed stress test does not depend on it.

## Linux verification

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNodeProcess|^TestProcess|TestInputAgreesWithNode/internal/oracle/testdata/node_process|^TestCountsAreRecorded$' -count=1 -timeout 30m -v > /tmp/process-cwd-fix/oracle.log 2>&1
go test ./internal/flow -count=1 -timeout 30m > /tmp/process-cwd-fix/flow.log 2>&1
go vet ./internal/oracle > /tmp/process-cwd-fix/vet.log 2>&1
```

All pass. Counts remain unchanged. No compiler lowering or native runtime was changed, so no additional native package rerun is claimed. The uncached oracle builds sanitized binaries for the changed sources and runtime witnesses. Full repository and all 25 host acceptance fixtures are not rerun here; the latter remain pending the area SHA containing fs-file.

Logs, including the independently forced original-driver loss, are under logs/cwd-barrier/. The original stdin probe returns Ok on Linux. This worker has no macOS execution environment.
