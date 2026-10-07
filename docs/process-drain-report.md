JavaScript process.exit now drains stdout and stderr before terminating with the chosen status.
Both Adamic backends preserve all 204,800 bytes on the backpressured Linux pipe.
Raw Node's 4,096-byte result is pinned as documented behavior, not the expected Adamic answer.
Ordinary exit cases still match Node byte for byte; the full uncached oracle passed.
The named no-deliberate-output-loss exception is documented beside exit 70 in process.md.

Branch: codex/process-exit-and-tty, new work on top of 2972601. No native implementation
changes. The final response records the pushed tip. The older output report is labelled
historical, so its former backend-loss observation cannot be mistaken for the current contract.

## Implementation

Generated JavaScript calls the private processExit helper imported from oracle/adamic.mjs.
It validates the chosen code before stopping execution, preserving catchable invalid-code
errors. It queues a completion callback behind existing writes on each stream, waits for
both callbacks, and then calls Node's exit with the captured status. Empty writes add no
output bytes. Raw Node source still calls Node's own process.exit: the helper does not
monkeypatch it, so the reference remains independent.

An internal stop signal aborts synchronous generated execution while the host event loop
drains the streams. Generated catches rethrow it; generated finally blocks and instrumentation
leave hooks skip it. The runtime recognizes this signal without reporting a panic. Valid
exit never returns to source code, and invalid exit still runs its source catch/finally.
This preserves the control flow held by the existing process fixtures, including numeric
status conversion, omitted/undefined arguments, and the uncaught-throw exit-70 convention.

The named exception and its rationale are in [process.md](process.md#named-oracle-exception-no-deliberate-output-loss):
Adamic never loses output on purpose, even where raw Node discards queued pipe writes.

## Observations

The Linux driver withholds a reader on a 4,096-byte pipe, then drains it after process exit
or one second. Every implementation uses the same policy. The payload is 200 lines of
1,023 ASCII x characters and LF, exactly 204,800 bytes. All cases choose exit 37.

| Case | Raw source Node | Native under ASan/UBSan | Adamic JavaScript |
| --- | --- | --- | --- |
| Backpressured stdout, explicit exit | 4,096 stdout bytes | 204,800 stdout bytes | 204,800 stdout bytes |
| Backpressured stderr, explicit exit | 4,096 stderr bytes | 204,800 stderr bytes | 204,800 stderr bytes |
| Backpressured stdout, normal return with exitCode | 204,800 stdout bytes | Identical | Identical |
| Regular file, output before explicit exit | 19 stdout bytes | Identical | Identical |
| Regular file, exitCode and normal return | 21 stdout bytes | Identical | Identical |
| Interleaved stdout/stderr sharing a pipe or file | 33 bytes in program order | Identical | Identical |
| Regular file, 200 KiB before explicit exit | 204,800 stdout bytes | Identical | Identical |

The pipe test asserts exactly the first 4,096 bytes from raw Node, including exit 37 and
empty other stream, so a Node change requires review. It independently compares both
Adamic backends with the complete expected payload. Node's lost tail is not accepted as
an Adamic result. Ordinary cases continue to compare stdout, stderr and status to raw
Node without normalization. Pipe loss depends on the driver and backpressure; these counts
are observations of this driver, not a claim about every shell pipe.

## Mutants

The requested JavaScript mutant replaces the generated draining call with raw process.exit.
It runs cleanly with exit 37 and emits exactly raw Node's 4,096 bytes; the full-output check
catches the missing 200,704 bytes. It is not killed by a syntax error, panic, or sanitizer.

Control-flow guards also have executable mutants: allowing the stop signal into a source
catch prints `must not catch` and resumes the caller; allowing source finally prints
`must not finally`; allowing instrumentation cleanup prints `must not leave`. Each exits
37 with empty stderr and is caught only by stdout differing from Node's `before\n`.
The previous native exit-flush and stderr-flush mutants remain caught by the file and
shared-descriptor comparisons. Native buffering and flushing are unchanged.

## Validation

Setup command: bash cloud/setup.sh, output /tmp/process-drain-setup.log, exit 0.
nproc: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux. Timings:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (17s)
setup: done in 17s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Source /workspace/adamic-tools/env.sh before each command. Logs are files, never pipes.

```
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  > /tmp/process-drain-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./internal/oracle \
  -run '^TestProcess' > /tmp/process-drain-final-process.log 2>&1
go test -count=1 -timeout 10m ./internal/javascript \
  > /tmp/process-drain-javascript.log 2>&1
go vet ./... > /tmp/process-drain-final-vet.log 2>&1
gofmt -l cmd internal
git diff --check
```

The complete oracle exited 0, PASS, 85.833s. Cache results: native hits 0, misses
1,064; Node hits 0, misses 608; probe hits 0, misses 24. All registered process
fixtures, counts, panic and stream-order probes, the new draining checks, and mutants
passed. After that run, only an additional instrumentation-guard probe/mutant was added
to the harness; the final focused process run holds that addition separately.
It exited 0, PASS, 6.229s, with zero cache hits, and caught all four new JavaScript
mutants (no drain, catch, finally, and instrumentation leave) only by output bytes.

The JavaScript package exited 0 and reports no test files; its behavior is held by the
oracle. Vet exited 0 with no output. Formatting and diff checks are clean. No registered
fixture or counts row was added; new sources are generated in scratch directories.
The complete repository gate was not repeated; the touched backend and entire uncached
oracle package were checked. Immediate native exit still skips LeakSanitizer's exit hook.
