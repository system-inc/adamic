Built: stable ascending UTF-16 default string Array.sort, with undefined before holes.
Commits: codex/library-array-sort-strings from c6d4f1a2267ff3f2efa12407347b0072920c34e2; implementation SHA in the handoff.
Commands and outputs: three fixtures agree on native sanitized/release, JavaScript and WASI; test262 sort improves 6 to 8 passes, zero disagreements.
Mutants: code-point comparison, descending, unstable equal-key runs and undefined-as-holes caught only by Node stdout; numeric admission checked separately.
Not covered: complete repository gate; remaining stage-1 checks were stopped. Host fixture 08 now reaches node:fs.symlinkSync's separate refusal.

## Observations

Node v24.19.0; test262 c8c798898646638cd0c24879f8e0374e847e7d74.
Every passing test compares with Node. Both new passes are S15.4.4.11_A2.1_T1.js
and S15.4.4.11_A2.1_T2.js. All six old passes remain passes.

| built-ins/Array/prototype/sort | Pass | Disagreements | Refused | Crashed | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Before | 6 | 0 | 37 | 0 | 11 | 54 |
| After | 8 | 0 | 35 | 0 | 11 | 54 |

The lowering accepts strings, literal unions and string | undefined. Its synthetic
comparator is ordinary IR; both backends and ownership analyses see its work.
The existing string comparison orders UTF-16 units. The runtime wrapper gathers
present strings, uses the existing stable V8 TimSort, then writes undefined before
leaving holes. Sparse arrays visit present numeric keys, and non-index properties
survive. No new libc calls: existing runtime APIs and trunc, available in wasi-libc.

Assumptions recorded in the commit: copy readonly arrays with slice() before sort(),
since TypeScript defines no sort on ReadonlyArray. Equal primitive strings have no
observable identity, so tagged objects with a string-key comparator expose stability
of the shared sorter. Explicit sort(undefined) keeps its existing behavior; this
unit admits calls with no arguments.

Fixtures cover ASCII, mixed case, non-ASCII, supplementary characters against BMP
above the surrogate range, lone surrogates printed as code units, duplicates, empty
arrays, literal unions, readonly copies, dynamically allocated and optional strings,
explicit undefined versus holes, aliases, out-of-order writes, repeated sorting,
non-index properties and distant sparse slots. The stability probe has 160 tags.

All four emission/runtime mutants compile with warnings as errors, exit 0 without
ASan/UBSan/LeakSanitizer diagnostics and fail only raw-source Node stdout comparison.
Code-point and descending mutations replace the emitted string comparator. The
unstable sorter preserves key order but reverses equal tagged runs. Undefined-as-holes
deletes explicit undefined after sorting. A separate Go overlay admits number elements
and must fail the unchanged-refusal assertion without modifying repository sources.

Unmodified host fixture 08 gets beyond default sort and refuses at line 208:
stage 0 can't lower node:fs.symlinkSync yet. Its complete execution is outside this unit.

## Commands and evidence

Shells source /workspace/adamic-tools/env.sh and export GOPROXY='https://proxy.golang.org|direct'.
Every test's output went directly to a log. Compressed logs and both test262 JSON
reports are under evidence/. The exact commands were:

```sh
bash cloud/setup.sh --wasi-sdk > /tmp/adamic-sort-strings-setup.log 2>&1
npm ci --prefix stage3/api > /tmp/adamic-sort-strings-npm.log 2>&1
ADAMIC_GATE_UNCACHED=1 go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array/prototype/sort > /tmp/adamic-sort-strings-before.json 2> /tmp/adamic-sort-strings-before.log
ADAMIC_GATE_UNCACHED=1 go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array/prototype/sort > /tmp/adamic-sort-strings-after.json 2> /tmp/adamic-sort-strings-after.log
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestArrayStringSort|^TestArrayHolesRefusals$|TestNativeAgreesWithNode/internal/oracle/testdata/(sorting.a|sorts.a|sort_releases.a|timsort.a|sort_top_level.a|library_array_copy.a|library_array_holes)' -count=1 -timeout 10m -v > /tmp/adamic-sort-strings-regression.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/adamic-sort-strings-counts-final.log 2>&1
go test -overlay /tmp/adamic-sort-strings-policy-overlay.json ./internal/oracle -run '^TestArrayStringSortRefusals$' -count=1 -v > /tmp/adamic-sort-strings-policy-mutant.log 2>&1
gofmt -l cmd internal > /tmp/adamic-sort-strings-gofmt.log
go vet ./... > /tmp/adamic-sort-strings-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/adamic-sort-strings-gate.log 2>&1
go run ./cmd/adamic c stage3/fixtures/host/08_getDirectories.a > /tmp/adamic-sort-strings-host08.c 2> /tmp/adamic-sort-strings-host08.log
```

Focused regression passed in 24.775s; counts update passed in 90.921s. Counts also
refresh the performance-core row and the directory-system row (the latter observes
its own fixture directory, now containing three new files).
The complete gate passes lower (105.556s), native (733.000s), load, fresh, IR,
the bridge and the test262 runner. It reports the inherited process_bad_code flow
failure and the two inherited raw-Node pipe-loss assertions, documented on the base
in ../array-eight/COMPILER-FIX.md. Fuzz seed 30 also refuses type never at line 178;
restoring the changed compiler files from c6d4f1a in a Go overlay reproduces the same
failure (7.476s). That generated program contains no sort calls.

Additional broad-gate failures: RegExpLongBacktrackNode exceeds its 3m deadline;
RegExpNativeTiming/quadratic_exec is killed at its 3s limit; CanonicalizeUnicodeNode
has a killed Node batch; LoadersMatchGoCohere reports house versus style on raw Node,
native and emitted JavaScript alike. Observation: none is a new string-sort fixture
failure. Inference: concurrent compilation and large Unicode checks contribute to
the timing failures. No repository-wide pass is claimed. After these core results,
the remaining stage-1 CSS, formatfiles, JSON, lint and helper checks were stopped.
Their termination is recorded in the archived gate log and gate-stopped.log.

The fuzz baseline command was:

```sh
go test -overlay /tmp/adamic-sort-strings-base-overlay.json ./internal/fuzz -run '^TestOverridesShapesAndLower$' -count=1 -v > /tmp/adamic-sort-strings-base-fuzz.log 2>&1
```

The overlay restores library_array.go, library_array_holes.go, ir.go and
emit_expressions.go from c6d4f1a without editing repository sources.

## Setup

Setup succeeded; nproc=5, cgroup cpu.max=400000 100000. Timing lines in seconds:
Node 0.027; Go 0.035; markdown 0.095; submodules 0.158; clang 0.254; WASI SDK 0.291;
build 35.898; tests deferred 36.023; cache warm 36.025; done 36.051.
Go 1.27.1, clang 20.1.8, WASI SDK 27. npm ci succeeded.
