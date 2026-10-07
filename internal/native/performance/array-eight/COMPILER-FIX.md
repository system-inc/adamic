Built: merged the compiler fix and reconciled closure ABI, labels, enum, optional storage, and host adapter conflicts.
Commits: merge 17b8d76c58b3a0c91d4d5b386c412ef0bdab6e88; parents 712469a73c6a891df2b41ed7d0a163d530a09e51 and 0127854e7aa4aaa7c217630663440cf2afe268ea.
Commands and outputs: focused eight 8/8; full Array 145 pass, 0 fail, 2327 refused, 0 crashed, 610 skipped; reduced native, JS and WASI exit 0.
Mutants: boxed search changed to pointer identity, caught only by Node stdout; inherited one-byte and WASI byte/exit mutants also pass their catcher tests.
Not covered: full repository gate; three unrelated process assertions still fail identically before and after the merge.

## Array observations

Test262: c8c798898646638cd0c24879f8e0374e847e7d74. Node: v24.19.0.
Every successful program was compared with Node. All 126 old passes remain passes; 19 programs newly pass.
The eight indexOf/15.4.4.14-5-{10,11,31,32} and lastIndexOf/15.4.4.15-5-{10,11,31,32}
all independently pass in the focused run and the full run.

| Scope | Run | Pass | Fail / disagreements | Refused | Crashed | Skipped | Total |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Eight | Before | 0 | 0 | 0 | 8 | 0 | 8 |
| Eight | After | 8 | 0 | 0 | 0 | 0 | 8 |
| built-ins/Array | Before | 126 | 0 | 2338 | 8 | 610 | 3082 |
| built-ins/Array | After | 145 | 0 | 2327 | 0 | 610 | 3082 |

The first full post-merge run also had 145 passes and no disagreements or crashes.
A repeat under concurrent baseline compiler builds had 144 passes and one compiler timeout on
sort/stability-2048-elements.js. Inference: resource contention caused that timeout. The final
uncached repeat, after competing builds completed, restores 145 passes and zero crashes.
All three full reports are preserved, including the timeout; the table uses the completed final run.

All commands below used `source /workspace/adamic-tools/env.sh` and
`export GOPROXY='https://proxy.golang.org|direct'`. Output went directly to log files.

```sh
ADAMIC_GATE_UNCACHED=1 go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array > /tmp/adamic-array-eight-fix-all-complete.json 2> /tmp/adamic-array-eight-fix-all-complete.log
ADAMIC_GATE_UNCACHED=1 go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array/prototype/indexOf/15.4.4.14-5-10.js built-ins/Array/prototype/indexOf/15.4.4.14-5-11.js built-ins/Array/prototype/indexOf/15.4.4.14-5-31.js built-ins/Array/prototype/indexOf/15.4.4.14-5-32.js built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-10.js built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-11.js built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-31.js built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-32.js > /tmp/adamic-array-eight-fix-eight.json 2> /tmp/adamic-array-eight-fix-eight.log
```

New passes:

- `built-ins/Array/isArray/15.4.3.2-0-4.js`
- `built-ins/Array/isArray/15.4.3.2-0-7.js`
- `built-ins/Array/isArray/15.4.3.2-1-1.js`
- `built-ins/Array/isArray/15.4.3.2-1-3.js`
- `built-ins/Array/isArray/15.4.3.2-1-5.js`
- `built-ins/Array/prototype/filter/15.4.4.20-10-2.js`
- `built-ins/Array/prototype/indexOf/15.4.4.14-5-10.js`
- `built-ins/Array/prototype/indexOf/15.4.4.14-5-11.js`
- `built-ins/Array/prototype/indexOf/15.4.4.14-5-31.js`
- `built-ins/Array/prototype/indexOf/15.4.4.14-5-32.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-10.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-11.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-14.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-31.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-32.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-8.js`
- `built-ins/Array/prototype/lastIndexOf/15.4.4.15-5-9.js`
- `built-ins/Array/prototype/map/15.4.4.19-6-1.js`
- `built-ins/Array/prototype/map/15.4.4.19-8-c-ii-4.js`

## Reduced program and mutant

[evidence/compiler-fix/minimal.a](evidence/compiler-fix/minimal.a) contains exactly:

```typescript
const object = {};
const array = [0, object];
```

Raw Node, sanitized native, emitted JavaScript run through `oracle/node.mjs`, and wasm32-wasi
run through `oracle/wasi.mjs` all exit 0 with no stdout or stderr. The final compiler was built with
`go build -o /tmp/adamic-array-eight-fixed-probes/adamic ./cmd/adamic`. Exact build and execution
argument vectors and exits are in `probe-final.log.gz`. Native build used `--sanitize`; WASI build
used `build --target wasm32-wasi`. WASI successfully compiles every runtime translation unit;
the merge resolution adds no libc calls.

Initial probe invocations used plain Node for emitted JS, which could not resolve the `adamic`
package, and the low-level canary WASI runner, which requires a stack export absent from the public
CLI artifact. Corrected invocations use the repository's oracle loaders; initial errors remain archived.

For the incoming `host_empty_object_slots.a` fixture, emitted C was mutated by replacing all 15
`adamic_equal_unions` search arguments with `adamic_equal_identity`. Control and mutant both compile
and exit 0 with no sanitizer or leak diagnostics. The control agrees with raw-source Node; the mutant
has different stdout. The comparison alone catches it. Full compiler flags and observations are in
`mutant.log.gz`; both generated C files are archived. No source mutation remains in the branch.

## Merge and regression observations

The requested merge has two parents. Nineteen conflicted files were resolved by preserving existing
Array holes, open numeric/closed string enum rules, implicit returns and host adapters, while taking
incoming boxed primitive-admitting slots, labels, truthiness and receiver closures. Assumption:
independent existing features on the two branches must both survive, rather than replacing either
branch's implementation wholesale. This assumption is recorded in the merge message.

The incoming argument-count closure ABI required adapting five performance methods, stdout
setBlocking, and sparse-array map. Interface calls preserve built-in padding and release of discarded
performance entries. Read-only fs option literals preserve raw initializer slots rather than entering
new optional-field storage. Existing refusal tests were reconciled with supported logical assignment
and the retained enum diagnostics. No prohibited file was manually edited.

Passing final regression commands (logs under `evidence/compiler-fix/`):

```sh
go test ./internal/lower -count=1 -timeout 30m
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(host_|library_array_holes)|TestCountsAreRecorded|TestNodeFSFile' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(taste_|enum|switch|implicit_return|node_process)|TestTheOracleCatchesOneByte' -count=1 -timeout 10m
go test ./internal/oracle -run 'TestInputAgreesWithNode/internal/oracle/testdata/node_process_performance_core.a|TestArrayHolesMilestone|TestNativeAgreesWithNode/internal/oracle/testdata/taste_optional_join.a' -count=1 -timeout 10m
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIAgreesWithNode/internal/oracle/testdata/host_empty_object_slots.a|TestWASIOracleCatchesMutants|TestWASIRunnerCatchesMutants' -count=1 -timeout 10m
```

The lower rerun passed in 117.849s; final allocation/host/holes oracle in 254.948s;
conflict oracle in 22.105s; adapters oracle in 3.247s; WASI oracle in 11.528s.
The initial broad command `go test ./internal/load ./internal/lower ./internal/flow ./internal/native ./internal/oracle -count=1 -timeout 30m`
had load pass 7.191s and native pass 302.315s, then exposed the merge issues above.
Its stale full oracle leg was stopped after later source resolutions and replaced with the focused
final checks. Its complete emitted output, including initial failures, is archived.

The final combined flow/lower rerun has only one flow failure, process_bad_code.a, and lower passes.
The flow failure also reproduces at pre-merge 712469a:

```sh
go -C /tmp/adamic-array-eight-baseline test ./internal/flow -run 'TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/process_bad_code.a' -count=1 -timeout 5m
go -C /tmp/adamic-array-eight-baseline test ./internal/oracle -run 'TestProcessExitDrainsStderr|TestProcessLargePipeExitPreservesOutput' -count=1 -timeout 5m
```

Both before and after, invalid process.exit produces the same flow trace error, and raw Node writes
all 204800 bytes where the two tests assert documented pipe loss. These three inherited assertions
are left unchanged; no passing full repository gate is claimed. Their before/after logs are archived.

## Setup

`bash cloud/setup.sh --wasi-sdk` succeeded; `npm ci --prefix stage3/api` added 3 packages in 809ms.
`nproc` is 5, cgroup cpu.max is 400000 100000. Setup timing lines, in seconds:
Go 0.018; Node 0.022; submodules 0.052; markdown dependencies 0.062; clang 0.168;
WASI SDK 0.181; Go build 29.548; tests deferred 29.640; cache warm 29.641; done 29.665.
The first focused test run stopped compiling the old stdout setBlocking adapter against the new ABI;
fixing that merge integration allowed all subsequent requested programs to run. The initial merge
and build diagnostics are described above; setup itself did not fail.
