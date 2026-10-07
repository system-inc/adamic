Historical observations at 2972601, before the ruling to preserve all output in both backends.
The current contract and validation are in [process.md](process.md) and [process-drain-report.md](process-drain-report.md).

The follow-up adds Linux exit-output coverage on top of 3a9d1fc; the pushed tip is in the final report.
Regular-file output before immediate exit and before normal return agrees byte for byte with Node.
Interleaved stdout/stderr sharing one file or pipe agrees in order, bytes, and chosen exit status.
The 200 KiB file case agrees; dropping exit flush loses 8,192 bytes and is caught.
The 200 KiB backpressured pipe exposes a known native mismatch: Node drops 200,704 bytes.

## Scope and observations

No compiler or runtime implementation changes. Tests extend `internal/oracle/process_test.go`
and generate their TypeScript source in test scratch directories. No new registered oracle
fixture or counts row is added. `process_output.py` supplies actual Linux descriptors, preserves
raw output as base64 in its observation envelope, and records the chosen status and pipe capacity.
Source Node, sanitized native, and the JavaScript backend are observed independently.

All cases choose exit 37. Stderr is empty when separate. The shared-descriptor cases use one
descriptor for both streams, exactly `2>&1`, not a merge of independently captured streams.
The ordinary pipe in this run had capacity 65,536 bytes. The large backpressured case explicitly
sets capacity to 4,096 and withholds its reader until process exit or one second, then drains:
Node can discard its queue and exit; native's synchronous writer can unblock and finish.
The same driver and policy are used for every implementation.

| Case | Source Node | Sanitized native | JavaScript backend |
| --- | --- | --- | --- |
| File, output then `process.exit(37)` | 19 bytes, exit 37 | Identical | Identical |
| File, `exitCode = 37`, output then normal return | 21 bytes, exit 37 | Identical | Identical |
| Shared pipe, interleaved output then immediate exit | 33 bytes, exit 37 | Identical order and bytes | Identical |
| Shared file, interleaved output then immediate exit | 33 bytes, exit 37 | Identical order and bytes | Identical |
| File, 200 lines of 1,024 bytes then immediate exit | 204,800 bytes, exit 37 | Identical | Identical |
| Backpressured pipe, same 200 KiB then immediate exit | 4,096 bytes, exit 37 | 204,800 bytes, exit 37: **known mismatch** | Same 4,096 bytes as source Node |
| Backpressured pipe, same 200 KiB then normal return with exitCode | 204,800 bytes, exit 37 | Identical | Identical |

The correct shared order is `out 0\nerr 0\nout 1\nerr 1\nout last\n`.
The large payload has 1,023 ASCII x characters plus LF on each of 200 lines. It crosses
the native 65,536-byte buffer three times and leaves an 8,192-byte tail to flush at exit.

`TestProcessLargePipeExitGap` is explicitly a gap assertion. It requires the ordinary
byte comparison to find `stdout differs` for immediate exit, source/backend Node to agree
byte for byte, and native to preserve the full payload. It does not normalize output or
claim the mismatch is a passing parity result. The normal-return control must match Node
exactly. A change to this gap requires reviewing Node first and updating the test.

## Executed mutants

Tests include a scratch copy of the real output runtime in generated C, rename its public
symbols, and route this fixture's output and explicit exit through it. The checkout and
cached runtime are unchanged. The buffer, full-buffer flushes, stderr path, and `_exit`
are the actual runtime implementation, with just the named flush removed. The valid
chosen code still goes through the actual status setter. Each mutant compiles with the
normal `-Werror` flags and ASan/UBSan and runs cleanly with exit 37 and empty separate stderr.

| Requested case | Mutant | What caught it |
| --- | --- | --- |
| 1: immediate exit to a regular file | Remove flush before `_exit` | Node has 19 bytes; mutant has 0 |
| 2: stdout and stderr share a pipe | Remove stdout flush before stderr writes | Mutant writes `err 0\nerr 1\nout 0\nout 1\nout last\n`; Node order differs |
| 2: stdout and stderr share a file | Same stderr flush mutant | Same ordering mismatch |
| 3: 200 KiB to a regular file | Remove flush before `_exit` | Node has 204,800 bytes; mutant has 196,608 |

The large pipe already differs from Node without a mutant, so it cannot honestly be used
as a passing baseline for a mutant kill. Case 3's file comparison isolates the missing exit
flush instead. Normal return uses the existing finish/atexit flush and LeakSanitizer hook;
explicit `_exit` skips that hook, so no leak-check claim is made for immediate-exit runs.

## Validation

Setup: `bash cloud/setup.sh > /tmp/process-output-setup.log 2>&1`, exit 0.
`nproc` printed 5. Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux.

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (73s)
setup: done in 73s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Source `/workspace/adamic-tools/env.sh` for each command. Test output goes to files.
The initial new cases passed in 4.991s, `/tmp/process-output-first.log`. The broader
uncached process and shared-descriptor regression passed in 6.548s,
`/tmp/process-output-final.log`; it includes the previous process mutants, numeric
argument tests, object-escape refusals, all twenty terminal/environment combinations,
and existing shared-file/shared-pipe ordering probes. Final verification:

```
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./internal/oracle \
  -run '^Test(Process|OneFileHoldsNodesOrder|FileWritesLandInNodesOrder)' \
  > /tmp/process-output-verified.log 2>&1
go vet ./... > /tmp/process-output-vet.log 2>&1
gofmt -l cmd internal
git diff --check
```

Test exit 0, `PASS`, package time 6.552s. Cache observations: native hits 0,
misses 8; Node hits 0, misses 21; probe hits 0, misses 4. Both new groups and
all requested mutant comparisons passed, including the separately labelled
known mismatch assertion. Vet exit 0 with no output; formatting and diff checks
printed nothing. The final log prints both the correct and mutated merged-stream
order, and the exact Node/native/backend byte counts for the large pipe.

The complete 14-minute gate was not repeated for this tests-only unit. No runtime fix
or complete parity for immediate exit to every Linux pipe is claimed. Pipe loss depends
on buffering, backpressure, and when the reader drains it; these exact loss counts are
observations of the stated driver, not a claim that every shell pipe loses those counts.
