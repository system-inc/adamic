# Affected gate: saved proof status

Status saved at 2026-10-07T15:49:34.774613+00:00. Fixed main is `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.
The full observed reference is green: **44 packages, 31 with tests, 13 without**.
All five shapes have completed their skipped-package runs: every skipped package
passed with identical test names, actions and genuine output to that reference.
Four selected gates finished; the oracle selected gate is still running. This is
a bounded, conservative proof, not a claim that all selected gates are green.
Only `bridge/tsgo/checker` and `internal/fresh` are test packages eligible to skip;
uncertain observations keep 30 packages selected. No result cache was added.

## Shape results

| Shape | Probe SHA | Selected / skipped | Skipped events | Selected gate |
|---|---|---|---|---|
| stage3: `stage3/affected-proof.a` | `575b32d2dd985cc1285723065c507bb9070ccf03` | 32 / 12 | **Identical**, 12 passed | Finished: 32 pass, 0 fail |
| docs: `docs/0.1.md` | `449254ddac9376d3382861c90d7f1d1e919c764a` | 30 / 14 | **Identical**, 14 passed | Finished: 30 pass, 0 fail |
| slice: `stage1/cohere/json/formatter.ts` | `4f8e0a3657ad34b2c96b625f4c87f554e96dd72e` | 30 / 14 | **Identical**, 14 passed | Finished: 28 pass, 2 fail |
| runtime: `internal/native/runtime/string.c` | `f36b912a865e35f7cbd4f74db32dade7f821fe26` | 33 / 11 | **Identical**, 11 passed | Finished: 32 pass, 1 fail |
| oracle: `internal/oracle/testdata/numbers.a` | `18a6c54abae99bb5b69a8bd5a2baf0f7b74a4d52` | 31 / 13 | **Identical**, 13 passed | Running: 20/31 completed, 1 fail so far |

Stage3 and docs selected gates passed. Slice retained scanner ENOSPC and a
60-minute markdownblocks timeout; scanner passed a separate uncached retry after
freeing space. Runtime retained a 60-minute markdownblocks timeout. Oracle has
correctly failed `TestCountsAreRecorded`: its changed fixture measured
`37 | 37 | 1 | 39 | 9 | 0` versus recorded `36 | 36 | 0 | 36 | 9 | 0`.
Original failures and retry events remain saved; no ledger was edited to hide them.
Markdownblocks passed alone in the reference and plain control. The selected
branch runner uses four workers; its contention failures are retained as failures.

| Loop | Before: whole observed gate | After: selected only | Instrument |
|---|---|---|---|
| stage3 | 92m21.615s | 72m47.271s | `proof.py --root /tmp/affected-main --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60/branches --binary /tmp/adamic-affected`; per-package literal argv in `stage3/state.json` |
| docs | 92m21.615s | 80m15.435s | `proof.py --root /tmp/affected-main --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60/branches --binary /tmp/adamic-affected`; per-package literal argv in `docs/state.json` |
| slice | 92m21.615s | 86m47.321s | `proof.py --root /tmp/affected-main --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60/branches --binary /tmp/adamic-affected`; per-package literal argv in `slice/state.json` |
| runtime | 92m21.615s | 91m27.978s | `proof.py --root /tmp/affected-main --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60/branches --binary /tmp/adamic-affected`; per-package literal argv in `runtime/state.json` |
| oracle | 92m21.615s | Still running; no final timing | `proof.py --root /tmp/affected-main --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60/branches --binary /tmp/adamic-affected`; per-package literal argv in `oracle/state.json` |

Build flags for every timing: the fixed/probe SHA above, recorder `1bd43a6`
(proof-driver reporting edits were uncommitted), nproc **5**, cpu.max
**400000 100000**, **Go 1.27.1 linux/amd64**, **clang 20.1.8**, **Node v24.19.0**;
`go test -c`, direct `-test.v=test2json -test.timeout=60m`,
**ADAMIC_GATE_UNCACHED=1**, existing Go compilation cache, no test-result reuse.
Each exceeds five minutes and ran once. These are not pure selection speedups:
the reference includes observation and serial markdownblocks, whereas selected
gates use four unobserved workers. The slice overlaps evidence compression and
scanner retry; runtime overlaps moving old evidence off the small `/tmp` disk.
The failure rows are diagnostic timings. Exact versions, command arrays and load
boundaries are saved per phase/package. Each package is built with
`go test -c -o BINARY PACKAGE`, then executes `BINARY -test.v=test2json -test.timeout=60m`
from its package directory; separate stdin `go tool test2json -t -p PACKAGE` saves JSON.

| Loop | Load before | Load after |
|---|---|---|
| Reference | 0.36 0.21 0.15 | 1.53 1.98 1.95 |
| docs selected | 1.78 1.95 2.46 | 1.03 1.38 2.11 |
| runtime selected | 2.17 2.12 2.44 | 1.95 2.16 2.75 |
| slice selected | 2.48 1.77 2.17 | 1.63 2.05 2.45 |
| stage3 selected | 1.08 1.54 1.73 | 1.51 2.02 2.53 |

## Markdownblocks observer measurement

| Loop | Before: observed | After: plain | Instrument |
|---|---|---|---|
| Fixed main, isolated markdownblocks | 2333.291s | 2319.646s | `032.timing.json` observer argv versus identical direct test binary flags without observer |

Observed minus plain is **13.645s (0.59%)**. The test
itself took **38m39.646s**. This requested single pair excludes Go compilation and
later closure decoding; the difference includes system noise and is not a precise
causal overhead estimate. Both passed with identical names/actions/verdicts.
Variable benchmark rates, temporary paths and preflight attribution differ in
raw output; `calibration-differences.json` preserves those differences. This
package is always selected. Skipped-package output comparisons remain strict.

Build flags: same fixed SHA/toolchain/quota above, uncached, one package alone,
60-minute deadline. Load observed `1.24 1.41 2.50` to `1.61 2.07 1.98`; plain `1.49 1.96 1.95` to `1.56 1.93 1.86`.

```sh
/tmp/affected-proof/main-60.json.logs/run-3685698845/notification-observer -o /tmp/affected-proof/main-60.json.logs/run-3685698845/032.trace -- /tmp/affected-proof/main-60.json.logs/run-3685698845/032.test -test.v=test2json -test.timeout=60m
go test -c -o /tmp/affected-proof/proof-60/markdown-plain/github.com_system-inc_adamic_stage1_cohere_markdownblocks.test github.com/system-inc/adamic/stage1/cohere/markdownblocks
/tmp/affected-proof/proof-60/markdown-plain/github.com_system-inc_adamic_stage1_cohere_markdownblocks.test -test.v=test2json -test.timeout=60m
```

## Mutants: completed evidence and remaining work

Controlled real-process CLI tests caught all three requested mutants: removing
observed inputs wrongly skipped changed relative fixture bytes; removing listings
wrongly skipped a new fixture; a separately compiled toolchain-comparison mutant
wrongly skipped packages after changing the recorded Node version. Fresh uncached
fixture tests and the independent actual Node version caught those omissions.
Compiled event-hash and attempt-directory-rollover mutants are also exercised.

In the actual five-shape matrix, stage3/docs/slice/runtime have completed all
three selections: the Node-comparison mutant is caught in every completed shape;
observed/listing omissions are not exposed by those four shapes. Oracle matrix
execution waits for its selected run to finish. Its already-completed fresh test
passed but changed output from **1040 writes / 985 acyclic** to **1041 / 986**,
the independent witness intended to catch omission of the relative fixture.
The listing mutant then needs its supplemental new-fixture probe with the primary
fixture restored, preventing that file hash from masking the directory omission.
That final real-repository mutant evidence is **not yet claimed completed**.

## Record, resume and artifact locations

```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/adamic-affected ./cmd/adamic-affected
ADAMIC_GATE_UNCACHED=1 /tmp/adamic-affected record -main e011f8f60899586d6373a5ccb07335ad82cfbf3c -jobs 4 -isolate github.com/system-inc/adamic/stage1/cohere/markdownblocks -out /tmp/affected-proof/main-60.json > /tmp/affected-proof/proof-60/reference.log 2>&1
ADAMIC_GATE_UNCACHED=1 /tmp/adamic-affected select -record /tmp/affected-proof/main-60.json > /tmp/selected-packages.txt 2> /tmp/select.stderr
python3 -B /workspace/adamic/cmd/adamic-affected/finish_proof.py --root /tmp/affected-main --main e011f8f60899586d6373a5ccb07335ad82cfbf3c --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60 > /tmp/affected-proof/proof-60-resumed-workflow.log 2>&1
```

The complete main record is `/tmp/affected-proof/main-60.json`; original full JSON
is `main-60.json.reference.jsonl`. Per-package closures, original events and timing
logs are in `main-60.json.logs/run-3685698845/`. These large integration artifacts
are deliberately not committed. This report and the resumable runner are committed.
`proof-60/workflow.json`, `branches/summary.json` and each shape's `state.json`
persist the current proof. Five detached worktrees retain probe commits; execution
switches the same checkout so paths/environment cannot cause vacuous all-selection.
No scratch branch or main branch is pushed.

Records and reference JSON are fsynced as packages finish. Reinvoke the same
destination to validate identity and run only pending packages. An interrupted
pending attempt gets fresh `resume-*` logs; original logs are preserved. Failed
packages require explicit `-retry-failed` after diagnosis; failure history remains.
Incomplete records select all. Completed destinations cannot become fresh runs.
The user-authorized 60-minute deadline and isolation profile are recorded in
format-5 identity. Older 30-minute attempts cannot resume into it. CLAUDE.md and
`cmd/adamic-gate` were not edited. Direct test execution avoids command-mode
test2json's inherited-SIGINT change; all actual oracle signal cases passed.

## Inputs

For each current package the command asks `go list -deps -test -json` for the actual
Go dependency graph, including its internal and external tests. It hashes Go files,
CGo files, embedded files, test files, C/C++ sources, headers, assembly and object
files, plus every dependency directory's complete `testdata` tree and module file.
It also records root `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `.gitignore` and
`.gitmodules`, including their absence. Selection recomputes the static closure, so
adding a source file or an embedded file cannot hide behind the old filenames.
Generated test-main files returned as absolute paths stay absolute.

Observed inputs are collected separately for each test binary and all of its children:

The embedded Linux seccomp notification observer is compiled outside the repository
for each record invocation. Its inherited syscall filter pauses pathname operations
in the test process tree, records them, and lets the kernel execute the original
syscall with `SECCOMP_USER_NOTIF_FLAG_CONTINUE`. It covers open/openat, exec, stat,
access, readlink and directory enumeration. Cwd and directory descriptors resolve
through `/proc/<tid>/cwd` and `/proc/<tid>/fd`; pathname memory uses
`process_vm_readv`, with `/proc/<tid>/mem` as a fallback. It does not attach ptrace,
so LeakSanitizer remains enabled. Trace entries use the parser's strace-like format.
Withdrawn notifications have not received CONTINUE and have not executed a file
operation; restarts produce new notifications. Other notification errors abort.

Directory fingerprints include names and entry types; file fingerprints include
mode and bytes, and symbolic links include their target and the target's fingerprint.
Missing paths and parents are explicit inputs. Undecodable paths, unsupported
operations and repository writes prevent skipping. Non-UTF-8 input names refuse
recording rather than allowing JSON to replace bytes in pathname keys. Trace logs
omit expanded environments; the environment identity stores a digest.

This is an experimental equivalent observer, not a complete integration proof.
Node on this box invokes `io_uring` syscalls. They are intercepted but undecoded,
so the package is uncertain and always selected. `openat2`, multi-path operations,
foreign syscall architectures and unreadable child memory also force selection.
A supported observer must additionally prove namespaces, inherited descriptors,
other asynchronous input interfaces and arbitrary external inputs before deployment.

CGo source lists do not establish a complete closure for arbitrary `#include` paths
used while building a cached binary. Repository CGo dependencies therefore force
selection until their build input tracing is proven as well.

The shared submodule identity includes cohere's HEAD, its binary working-tree diff
against HEAD, recursive submodule commit status, and every non-Git working-tree
file and directory below cohere. This includes staged, unstaged, untracked and nested
working-tree bytes; it intentionally selects more packages than the minimum when
cohere changes.

Toolchain identity contains `go version`, full `clang --version`, `node --version`,
stable Go configuration and the invoking environment's digest. The audit found
literal and dynamic `os.Getenv` calls and `os.Environ` forwarded to children, so
hashing only a list of literal names would omit inputs. Stable Go configuration
excludes `GOGCCFLAGS`, whose generated temporary prefix changes on each `go env`
invocation. No environment values are written into the JSON record.

This first implementation keeps absolute external Go input paths and exact environment
identity. Changing checkout location, `PWD`, cache paths or workspace configuration
can select everything. That conservative behavior is observed in the design, not
claimed as a worktree speedup. External corpus/library directories named by environment
variables require an additional closure audit before production use. The parser records
repository paths, not arbitrary external data or network responses.

## Tracer blocker, observed on this box

A standalone leak-free C program, compiled with `clang -fsanitize=address,undefined`,
exited 0 normally and 1 under the requested `strace -f` filter. Its output was:

```text
LeakSanitizer has encountered a fatal error.
HINT: LeakSanitizer does not work under ptrace (strace, gdb, etc)
```

The actual repository test `TestFreedValuesAreCaughtWithSlabs` likewise passed
normally and failed under tracing, for both clean control binaries. These checks
were left enabled. The recorder refuses a failed traced package rather than turning
that failed observation into a green record.

An unprivileged seccomp listener is available on this kernel. The replacement
observer passes both the leak-free sanitizer control and the actual native
`TestFreedValuesAreCaughtWithSlabs` test, with leak checking enabled. The latter
log is `/tmp/affected-notify-native.log`. This removes the specific ptrace conflict,
not arbitrary closure completeness beyond the bounded shapes above.

## Validation and retained attempts

Final command tests passed (63.492s), `go vet` passed, and all five Python
comparison tests passed. Validation overlapped the unfinished oracle selected run;
its eventual wall time must disclose that extra load.

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout 5m ./cmd/adamic-affected > /tmp/affected-proof/final-command-tests.log 2>&1
go vet ./cmd/adamic-affected > /tmp/affected-proof/final-vet.log 2>&1
python3 -B -m unittest discover -s cmd/adamic-affected -p test_proof.py > /tmp/affected-proof/final-python-tests.log 2>&1
```

The exhaustive path audit is `/tmp/affected-path-audit-all.log`: 4,488 matches,
1,702 outside cohere. It includes sibling stage1 files, native benchmark fixtures,
oracle fixture globs and scanner cohere tables. Dynamic environment reads and
environments forwarded to children justify the full environment digest.
Optional external corpora and unsupported asynchronous interfaces remain limits.

Older failed attempts remain uncertified: interrupted ordinary full gate;
format-3 signal-harness failure; format-4 absent width dependencies and repeated
30-minute markdown deadlines; and the initial calibration controller rejecting
variable output from an always-selected package. That controller resumed the
saved green record and plain run rather than repeating them. Missing dependencies
were pinned to emoji-regex@10.6.0, get-east-asian-width@1.6.0 and narrow-emojis@0.0.3.
Completed traces and older failed-attempt binaries were losslessly compressed and
decompressed SHA256-checked. Old logs were moved to `/workspace/affected-evidence/`
with verified hashes and original-path symlinks. Retention manifests remain in
`/tmp/affected-proof/`. Original JSON streams remain unchanged. Only rebuildable
successful-reference test binaries were removed.

Setup reported tools/submodules ready 0s each, build warm 280s, done 280s;
nproc 5 and cpu.max 400000 100000. Before/after load was not captured for setup,
so that script-reported duration is not a paired benchmark.

Observed-path processing was measured on the actual saved fresh-package trace,
interleaving before/after three times on this box. The trace was loaded before the
timed loop; file bytes were fingerprinted during every trial. All three comparisons
produced identical closures and uncertainty lists. This is not a gate speedup.

Build flags: worker `2e18756`, input main
`e011f8f60899586d6373a5ccb07335ad82cfbf3c`, nproc 5, cpu.max `400000 100000`,
Go 1.27.1 linux/amd64, clang 20.1.8, Node v24.19.0, uncached test results,
existing Go compilation cache. Best before load: `0.44 1.76 3.29` to
`0.41 1.73 3.27`; best after load: `0.45 1.72 3.26` to `0.45 1.72 3.26`.
All six trial metadata rows are in `/tmp/affected-proof/observation-replay.log`.

| Loop | Before | After | Instrument |
|---|---|---|---|
| Saved fresh trace, best of 3 | 3.253772s, hash every open | 0.102878s, hash the path union | `ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout 30m ./cmd/adamic-affected -run '^TestObservationReplayTimings$' -args -affected-replay-trace /tmp/affected-proof/green-main.json.logs/run-1538579916/014.trace -affected-replay-root /tmp/affected-main > /tmp/affected-proof/observation-replay.log 2>&1` |

The remaining obligations are oracle selected-run completion and its observed
and new-listing mutant witnesses. All five skipped-package comparisons are done.
No expensive-gate speedup or general external-input completeness is claimed.
