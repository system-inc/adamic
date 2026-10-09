# Census and Error runtime facts

The census fix is commit 2bf78f6a872bd75d55437e5a8f54e2196a469183, pushed separately.
Configured per-file checks retain project roots, lib and options, but report file
errors only for requested files. Configuration and global errors remain visible.

| Composite fixture | Before | After |
| --- | ---: | ---: |
| main alone | 1 TS6307 | 0 |
| values alone | 0 | 0 |
| whole project | 0 | 0 |

Stock TypeScript 6.0.3 independently gives 1/0/0 when deliberately single-rooted.
The requested-root mutant restores TS6307 and fails TestCompositeProjectCensus.

The Error prelude adds the pinned @types/node ErrorConstructor signatures.
The loader also honors an explicit project types:["node"] request without a
node: import, using the already pinned embedded Node declarations.

On the adapted TypeScript 6.0.3 tree, before fixing either issue, per-file loads of
debug.ts and watch.ts each returned 137 diagnostics across the import graph.
Of these, debug.ts itself had three (TS6307 and two missing captureStackTrace
members); watch.ts itself had one TS6307. Both also reported a global TS2688
for the explicit Node type library. Afterward each requested file reports zero.
The debug witness reproduces the two captureStackTrace sites at lines 200-201.

Both backends implement captureStackTrace by defining a writable, non-enumerable
own stack string. Native stores it outside the fixed object shape, retains and
releases owned values, and makes reads, writes and own-property checks consult
that storage. JavaScript defines the equivalent property. Recapture, immutable
aliases, detached capture, and evaluation of constructorOpt are exercised.
stackTraceLimit is ordinary shared numeric IR state, initialized to 10, with
read-back tests for writes from the entry point and a function.

Assumptions: frame text and constructorOpt frame filtering have no specified
portable content; the implementation uses an empty string. Fixtures print only
presence, ownership and typeof, never stack text, length or frames. Uncaptured
Error stacks remain refused. Native descriptor exceptions are not implemented,
so programs that freeze objects are refused. Targets must be provably plain
objects or Errors, with no existing shaped stack field; unknown layouts,
existing shaped stack fields, stack destructuring, and compound limit updates receive
named refusals. Captured Error read proofs do not cross function boundaries.
These are conservative lowering boundaries, not declaration lies or no-ops.

Mutants are applied to lowered IR without modifying the working tree:
- Remove the capture body: clean exit, own-property fixture prints false instead
  of Node's true.
- Remove limit assignments: clean exit, read-back differs from Node.
Both mutants run natively with ASan/UBSan, on the JavaScript backend and on WASI;
only Node stdout comparison catches them, with no sanitizer finding.

Four fixtures are registered on both backends and the WebAssembly leg. The
allocation-count table was regenerated. The only existing row that changes is
node_fs_directory_system.a, which enumerates the fixture directory: four added
files increase its directory-entry allocation counts.

Commands (setup environment sourced, Node v24.19.0):
```text
python3 internal/load/testdata/project-census/run_mutant.py
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -timeout 20m
ADAMIC_ORACLE_WASI=1 go test ./internal/load ./internal/lower ./internal/oracle -count=1 -timeout 15m
go test ./internal/lower -count=1 -timeout 10m
go test ./internal/load -count=1 -timeout 10m
go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
ADAMIC_ORACLE_WASI=1 go test ./internal/load ./internal/lower ./internal/oracle -run 'TestErrorCaptureMutants|TestErrorCaptureBoundaries|TestErrorRuntimePrelude|Test(Native|WASI)AgreesWithNode/internal/oracle/testdata/library_error_' -count=1 -timeout 10m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go run ./stage3/ledger/checker-259/probe-stage0.go /workspace/load-error-adapted/src/compiler/debug.ts /workspace/load-error-adapted/src/compiler/watch.ts
```

Logs live in /tmp/load-census-* and /tmp/load-error-*. The census full oracle uses
an isolated checkout at its pushed commit, with the shared TypeScript checkout.
The environment restart interrupted earlier full gates; final gate logs are
load-census-final-oracle.log, load-error-final-all.log, and
load-error-load-final.log. No new C libc dependency is introduced beyond wasi-libc.

Census gate: whole oracle with WASI PASS, 716.500s in the isolated checkout.
Final load gate: PASS, 20.696s. Native package gate: PASS, 194.698s.
Final conservative-boundary lowering gate: PASS, 107.768s.
Focused final fixtures and mutants pass on native, JavaScript and WASI.
Final Error whole oracle with WASI: PASS, 556.643s, zero Node disagreements.
Green recorded at 04:37 UTC October 8, before 22:45 MDT October 7.
