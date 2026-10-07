# Affected gate input records

The requested skipped-package claim passed for all five shapes at fixed main
`e011f8f60899586d6373a5ccb07335ad82cfbf3c`. Every package the selector would skip
was run uncached and had identical test names, actions and genuine output to its
reference. The selected gates had failures, reported below; the runner exits
unsuccessfully for those failures. This is a conservative, bounded proof: only
`bridge/tsgo/checker` and `internal/fresh` are test packages eligible to skip.
Uncertain observation keeps 30 packages selected. No test result cache was added.

Assumption: the fixed main and five saved probe commits remain the proof inputs.
Existing selected failures remain gate failures; they do not invalidate an independently
verified skipped-package comparison. The user requested their pass/fail report.

## Commands and durable evidence

Build outside the repository, then record at a clean fixed-main checkout with
initialized submodules. Recording itself runs the whole uncached observed gate.

```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/adamic-affected ./cmd/adamic-affected
/tmp/adamic-affected record -main e011f8f60899586d6373a5ccb07335ad82cfbf3c -jobs 4 -isolate github.com/system-inc/adamic/stage1/cohere/markdownblocks -out /tmp/affected-proof/main-60.json > /tmp/affected-proof/proof-60/reference.log 2>&1
/tmp/adamic-affected select -record /tmp/affected-proof/main-60.json > /tmp/selected-packages.txt 2> /tmp/select.stderr
python3 -B /workspace/adamic/cmd/adamic-affected/finish_proof.py --root /tmp/affected-main --main e011f8f60899586d6373a5ccb07335ad82cfbf3c --record /tmp/affected-proof/main-60.json --output /tmp/affected-proof/proof-60 > /tmp/affected-proof/proof-60-resumed-workflow.log 2>&1
```

`FILE.partial` and `FILE.reference.jsonl` are fsynced as packages finish.
Resume the same destination to validate commit, repository inventory, environment,
package plan, static and observed inputs, and saved event hashes before running
only pending packages. Interrupted pending logs get a fresh `resume-*` directory;
old logs remain untouched. Failed packages require explicit `-retry-failed` after
diagnosis, preserving logs and `Retries` history. Incomplete records select all.
A completed destination cannot be reused as a fresh record. Keep these artifacts
with integration output; none is committed. Check command exit status and record
commit before publishing a reference. No main branch or probe branch was pushed.

The user authorized 60 minutes because integration already uses it; CLAUDE.md's
30-minute line was not edited. Format 5 records the exact deadline and direct
runner profile in toolchain identity. Changing that identity selects everything.
Test binaries execute with `ADAMIC_GATE_UNCACHED=1 -test.v=test2json -test.timeout=60m`.
The recorder and proof runner set that variable for tests internally. Preserve
the selector invoking environment used when recording; adding an environment
variable only on select deliberately causes conservative all-selection.
A separate stdin-mode `go tool test2json -t -p PACKAGE` records output. Command-mode
test2json changed inherited SIGINT handling and failed an earlier real oracle
interrupt test; the direct runner passed all signal cases. Go compilation caching
remains enabled; test processes and Adamic fixture results are uncached.

The reference is green: 44 packages, 31 with tests, 13 without. Its record is
`/tmp/affected-proof/main-60.json`; original full events are in the adjacent
`.reference.jsonl`, and individual events are in `.logs/run-3685698845/`.
`proof-60/branches/summary.json`, each shape's `state.json`, and
`proof-60/verified-proof.json` retain selections, original JSON, event hashes,
identity checks, timings, instruments, failures, and mutant results. The restart
added CODEX_APPLY_PATCH_PRESERVE_LINE_ENDINGS; omitting only that editor flag
reproduced the original environment SHA256 exactly. Tool versions/configuration
also matched, and the baseline selection revalidated all saved closure identities.
No toolchain or environment comparison was disabled. The final
independent verification rehashed all recorded and branch event streams and
recomputed every skipped-package comparison.

## Five shapes and timing results

The scratch probes change one path each. Five detached review worktrees retain
the commits; execution switches the same `/tmp/affected-main` checkout so absolute
paths and environment identity do not cause a vacuous relocation selection.
Each shape runs exactly the skipped packages, then exactly the selected packages.
All run once because these gates exceed five minutes. The reference ran once.
No pure selection speedup is inferred: the reference includes observation and
queues markdownblocks last and alone; selected gates use four unobserved workers.
Only two inexpensive test packages are eligible to skip. Stage3 skips no test package.

The workspace restarted after 26 oracle selected packages were checkpointed.
Only the five pending packages resumed, with markdownblocks last and alone.
Original unfinished events were retained in their original directory; new events
use a resume directory. Oracle wall time is accumulated checkpointed execution
plus the resumed session, not an uninterrupted gate wall time; downtime and
uncheckpointed lost execution are excluded. It is a diagnostic lower bound.
Other four shapes retain their original complete gate measurements.

| Loop | Before: whole observed reference | After: selected only | Selected / skipped | Skipped events | Selected result | Instrument |
|---|---|---|---|---|---|---|
| stage3: `stage3/affected-proof.a` | 92m21.615s | 72m47.271s | 32 / 12 | PASS, identical | PASS | `run_set(selected, ...)`; exact package argv in `stage3/state.json` |
| docs: `docs/0.1.md` | 92m21.615s | 80m15.435s | 30 / 14 | PASS, identical | PASS | `run_set(selected, ...)`; exact package argv in `docs/state.json` |
| slice: `stage1/cohere/json/formatter.ts` | 92m21.615s | 86m47.321s | 30 / 14 | PASS, identical | FAIL: stage1/cohere/markdownblocks, stage1/typescript/scanner | `run_set(selected, ...)`; exact package argv in `slice/state.json` |
| runtime: `internal/native/runtime/string.c` | 92m21.615s | 91m27.978s | 33 / 11 | PASS, identical | FAIL: stage1/cohere/markdownblocks | `run_set(selected, ...)`; exact package argv in `runtime/state.json` |
| oracle: `internal/oracle/testdata/numbers.a` | 92m21.615s | 94m24.208s | 31 / 13 | PASS, identical | FAIL: internal/oracle | `run_set(selected, ...)`; exact package argv in `oracle/state.json` |

The exact measured per-package command is `go test -c -o BINARY PACKAGE`, then
`BINARY -test.v=test2json -test.timeout=60m` from its package directory, with output
converted by `go tool test2json -t -p PACKAGE` from `/tmp/affected-main`.
Every `state.json` result stores literal `build`, `test`, `converter`, cwd and
JSON destination arguments; each phase stores its exact package list and load
boundaries. The orchestrator command above is the instrument for all five loops.

Build-flags line for every table row: fixed main above; probe SHA below;
recorder `1bd43a6` (proof-driver reporting edits were uncommitted during execution);
nproc **5**, cgroup cpu.max **400000 100000** (four CPU quota cores);
**Go 1.27.1 linux/amd64**, **clang 20.1.8**, **Node v24.19.0**; `go test -c`,
`-test.v=test2json -test.timeout=60m`, **ADAMIC_GATE_UNCACHED=1**; existing Go
build cache, no test-result reuse. Full compiler lines are in each phase's metadata.

| Loop | Input SHA | Load before | Load after |
|---|---|---|---|
| Whole reference | `e011f8f60899586d6373a5ccb07335ad82cfbf3c` | 0.36 0.21 0.15 | 1.53 1.98 1.95 |
| stage3 selected | `575b32d2dd985cc1285723065c507bb9070ccf03` | 1.08 1.54 1.73 | 1.51 2.02 2.53 |
| docs selected | `449254ddac9376d3382861c90d7f1d1e919c764a` | 1.78 1.95 2.46 | 1.03 1.38 2.11 |
| slice selected | `4f8e0a3657ad34b2c96b625f4c87f554e96dd72e` | 2.48 1.77 2.17 | 1.63 2.05 2.45 |
| runtime selected | `f36b912a865e35f7cbd4f74db32dade7f821fe26` | 2.17 2.12 2.44 | 1.95 2.16 2.75 |
| oracle interrupted segment | `18a6c54abae99bb5b69a8bd5a2baf0f7b74a4d52` | 2.05 2.12 2.68 | Not captured at interruption |
| oracle resumed segment | `18a6c54abae99bb5b69a8bd5a2baf0f7b74a4d52` | 0.96 0.36 0.14 | 1.01 1.38 1.58 |

Let **A** be these 30 always-selected uncertain packages (paths relative to the module):

```text
bridge/tsgo
bridge/tsgo/archive
cmd/adamic-meter
cmd/adamic-stage1-progress
cmd/adamic-test262
internal/flow
internal/fuzz
internal/load
internal/lower
internal/native
internal/oracle
internal/regexp
internal/unicodeproperties
stage1/cohere/css
stage1/cohere/cssnumbers
stage1/cohere/cssstrings
stage1/cohere/formatfiles
stage1/cohere/gitignore
stage1/cohere/graphql
stage1/cohere/json
stage1/cohere/lint
stage1/cohere/markdownblocks
stage1/cohere/markdowninline
stage1/cohere/mediaquery
stage1/cohere/selector
stage1/cohere/suppression
stage1/cohere/typeaware
stage1/cohere/values
stage1/typescript/parser
stage1/typescript/scanner
```

Let **N** be these 12 certain packages without tests:

```text
bench
bench/regex
bridge/tsgo/cost
bridge/tsgo/oracle
bridge/tsgo/spec
cmd/adamic
cmd/adamic-fuzz
internal/ir
internal/javascript
stage1/cohere/markdownblocks/tools/generate_classes
stage1/cohere/markdownblocks/tools/generate_entities
stage1/cohere/markdownblocks/tools/generate_width
```

| Shape | Exact selected set | Exact skipped set |
|---|---|---|
| stage3 | A + `bridge/tsgo/checker`, `internal/fresh` | N |
| docs | A | N + `bridge/tsgo/checker`, `internal/fresh` |
| slice | A | N + `bridge/tsgo/checker`, `internal/fresh` |
| runtime | A + `bench/regex`, `cmd/adamic`, `cmd/adamic-fuzz` | N minus `bench/regex`, `cmd/adamic`, `cmd/adamic-fuzz` + `bridge/tsgo/checker`, `internal/fresh` |
| oracle | A + `internal/fresh` | N + `bridge/tsgo/checker` |

The slice scanner failed with `no space left on device`; after freeing space its
separate uncached retry passed. Original and retry JSON are retained, including
`slice/scanner-retry.json`; the original gate remains failed. Slice and runtime
markdownblocks reached 60 minutes under contention. The oracle fixture correctly
failed `internal/oracle/TestCountsAreRecorded`: recorded counts
`36 | 36 | 0 | 36 | 9 | 0` versus measured `37 | 37 | 1 | 39 | 9 | 0`.
Its count ledger was not changed to hide that selected failure. No timeout failure was
reclassified as success. The slice timing overlaps evidence compression and the
scanner retry; runtime overlaps moving older evidence off `/tmp`. These failure
rows are diagnostic timings, not clean performance improvements. Oracle also
overlapped final command validation before the interruption; its resumed
markdownblocks run executes alone. No previous failure was replaced with a pass.

## Observer overhead

| Loop | Before: observed | After: plain | Instrument |
|---|---|---|---|
| Isolated markdownblocks at fixed main | 2333.291s | 2319.646s | observed `032.timing.json` argv; plain direct binary and stdin test2json |

Observed minus plain is **13.645s**, **0.59%** of plain execution. The test itself took
about 38m40s. This single requested pair excludes Go compilation and later closure
decoding. It measures a difference under system noise, not a precise causal estimate.
Both passed with identical test names, actions and verdicts. Their raw outputs
differ in benchmark rates, temporary paths and shared preflight output attribution;
`calibration-differences.json` preserves those differences. Markdownblocks is always
selected; no such normalization is permitted for a skipped package.

Build flags: same fixed main, worker, toolchain, nproc and quota above; uncached;
both isolated; direct `-test.v=test2json -test.timeout=60m`. Observed command:

```sh
/tmp/affected-proof/main-60.json.logs/run-3685698845/notification-observer -o /tmp/affected-proof/main-60.json.logs/run-3685698845/032.trace -- /tmp/affected-proof/main-60.json.logs/run-3685698845/032.test -test.v=test2json -test.timeout=60m
```

The plain binary was built with
`go test -c -o /tmp/affected-proof/proof-60/markdown-plain/github.com_system-inc_adamic_stage1_cohere_markdownblocks.test github.com/system-inc/adamic/stage1/cohere/markdownblocks`,
and invoked with the same test flags from its package directory. Load averages:
observed `1.24 1.41 2.50` to `1.61 2.07 1.98`; plain `1.49 1.96 1.95` to `1.56 1.93 1.86`.

## Mutants across all five shapes

Each mutant retains uncertainty and removes only its named component. A separately
compiled selector drops toolchain comparison; its input record changes Node's
version string. The original selector requires all 44 packages for that mismatch.

| Shape | Remove observed | Remove directory listings | Drop toolchain comparison |
|---|---|---|---|
| stage3 | Not caught by this shape | Not caught by primary shape | Caught: wrongly skips 12 packages despite Node mismatch |
| docs | Not caught by this shape | Not caught by primary shape | Caught: wrongly skips 14 packages despite Node mismatch |
| slice | Not caught by this shape | Not caught by primary shape | Caught: wrongly skips 14 packages despite Node mismatch |
| runtime | Not caught by this shape | Not caught by primary shape | Caught: wrongly skips 11 packages despite Node mismatch |
| oracle | internal/fresh | internal/fresh | Caught: wrongly skips 13 packages despite Node mismatch |

The observed mutant wrongly skips `internal/fresh` for the changed relative-path
oracle fixture. Its fresh run passes, but reports **1041 writes / 986 acyclic** versus
**1040 / 985** in main; the strict event comparison catches it. The other shapes
do not expose this omission, and that absence is reported.

The primary oracle file change still has a file hash, which masks the directory
mutant. Its supplemental probe restores that file to main and adds
`internal/oracle/testdata/affected_proof.a`. The correct selector runs fresh; the
listing mutant skips it. A new uncached fresh run has different program/write
counts and catches that omission. The fixture is removed and the probe commit
restored afterward. Original controls, mutant records and exact selected sets
are saved under each shape directory.

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

## Validation, failed attempts and supporting measurements

The exhaustive test-path audit is `/tmp/affected-path-audit-all.log`: 4,488 matches,
1,702 outside cohere. It found native benchmark fixtures, lowering oracle fixtures,
fresh directory globs, sibling stage1 slices and scanner cohere tables.
Literal/dynamic environment reads and environments forwarded to children explain
the full environment digest. Optional external corpora require further audit.

Controlled CLI tests run real observed process trees and fresh uncached fixtures.
Observed omission changes bytes and fails a test; listing omission adds a fixture
and fails enumeration; a compiled toolchain mutant wrongly skips packages and
the independent actual Node version catches it. Event-hash omission accepts a
corrupted reference and is caught by an independent hash. Attempt-rollover omission
overwrites unrecorded logs and is caught by their original digest. Python resume
rollover and markdown-isolation mutants also each fail the independent regression
check; logs are in resume-runner-mutants/. Six Python runner tests pass. Deadline
identity, file modes, symlink targets, missing files, testdata, isolation and
uncertainty are checked. A real native static-only supplemental mutant omitted
`bench/regex/cases.json`; changed invalid JSON failed `TestRegExpLintPatternsNode`.
The path was restored. No fixture result cache was added.

Final validation logs and exact commands:

The unchanged Go command suite passed in 63.492s and `go vet` passed. All six
final Python tests passed; both additional resume-runner mutants failed their
independent regression check as intended.

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout 5m ./cmd/adamic-affected > /tmp/affected-proof/final-command-tests.log 2>&1
go vet ./cmd/adamic-affected > /tmp/affected-proof/final-vet.log 2>&1
python3 -B -m unittest discover -s cmd/adamic-affected -p test_proof.py > /tmp/affected-proof/final-python-tests.log 2>&1
```

Older attempts remain separate and uncertified: the interrupted ordinary full gate;
format-3 signal-harness failure; format-4 missing width dependencies and repeated
30-minute markdown deadlines; and the initial calibration controller's overly
strict comparison of varying selected-package output. The latter resumed the saved
green reference and saved plain run without rerunning either. Pinned missing
dependencies were installed with `npm install --prefix /tmp/adamic-markdown-width
--ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0
narrow-emojis@0.0.3`. Logs and records are retained. Large finished traces and old
failed-attempt binaries were losslessly compressed and decompressed SHA256-checked.
`compressed-traces.json`, `compressed-old-binaries.json` and
`relocated-failed-evidence.json` document retention. Successful reference test
binaries were deleted as rebuildable outputs; their JSON and closures remain.
Older failed logs moved to `/workspace/affected-evidence/green-main.json.logs`
with verified hashes and an original-path symlink.

Setup reported Go, clang, Node and submodules ready in 0s each; build warm 280s;
done 280s. Its `go build ./...` / `go test -count=1 -run '^$' ./...` warmed Go
compilation. It ran earlier on the same main/toolchain/quota; before/after load was
not recorded, so this is a script-reported setup duration, not a paired benchmark.

| Loop | Before | After | Instrument |
|---|---|---|---|
| Setup | Not measured | 280s reported | `bash cloud/setup.sh > /tmp/affected-setup.log 2>&1` |

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

The expensive gate benefit remains unproven: uncertain packages always run and
the five shapes do not establish arbitrary external-input or asynchronous-input
closure completeness. The specific five skipped-package claims and three requested
mutant sensitivities are established; selected timeout/resource failures remain
failures. `cmd/adamic-gate` and integration configuration were not edited.
