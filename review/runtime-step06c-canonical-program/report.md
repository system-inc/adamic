# Program canonical closure boundary

Named stop: constructor nested-frame finalization is missing on base `325aded7`.
This branch pins the boundary; it does not ship the unvalidated runtime change.

The two `.a` witnesses reference a nested declaration as a value from a constructor,
compare its identity, retain it past its frame, and create a second instance. The
ordinary witness captures a string and its owner. The counted twin uses an optional
argument and `arguments.length`. A saved function reference makes the closure itself
a selected Program member, rather than merely its enclosing object.

## Observation and implication

With the existing refusal temporarily removed during diagnosis, both nested `read`
functions were selected (`ProgramRegion=true`, `NestedParent=2`). Their parent,
`Reader_new_initialize`, had `FrameIdentity=0` and no frame environment. Generated C
used two separate `adamic_closure_new` calls, or the counted twin, and no canonical
constructor. The diagnostic log records these observations.

`internal/native/emit_expressions.go:287` only selects the canonical cache when the
parent frame identity is positive. `finishNestedEnvironment` in
`internal/lower/nested_functions.go:249` establishes that frame; ordinary function
finalization calls it in `internal/lower/functions.go:321`. The constructor path
needs the equivalent compiler finalization before adopt-before-cache can be tested
through this language shape. Inferring broken identity from the fresh allocations
is a diagnosis, not an observed native execution result.

The existing named refusal remains unchanged:
`Program canonical closure before canonical-cache adoption`.
No runtime C/header, emitter, or production lowering changes are committed.
The proposed adopt-before-cache implementation and preliminary tests were reverted.

## Evidence

Node v24.19.0 executes both witnesses successfully. Ordinary stdout is
`true, true, node1, true, false, node1, node2` (one value per line). Counted stdout is
`true, true, node1:0:true:true, node1:1:true:false, true, false,
node1:0:true:true, node2:0:true:true`.

`TestProgramCanonicalConstructorBoundary` requires the exact named refusal for
both witnesses. Its mutant removes that guard: both cases fail because lowering
returns no error. The guard was restored and the test rerun successfully.

No native identity comparison, ASan/UBSan/leak result, balanced counts result, or
cache-before-adoption ASan mutant is claimed. The requested cache-order mutant
could not reach a canonical allocation. No accepted oracle/counts row was added;
the full gate was not run at this named stop.

## Commands and machine

Linux x86_64, `nproc` 5, cgroup CPU quota 4 cores; Go 1.27.1, clang 20.1.8.
Setup cumulative timing lines: Node 0.074s, Go 0.076s, Markdown 0.284s, clang
0.538s, submodules 194.910s, Go build 629.193s, test binaries deferred 629.471s,
warm cache 629.474s, done 629.614s. Setup ran in the background and was checked.

All final test/build/vet commands carry an external 90-second kill:

```sh
source /workspace/adamic-tools/env.sh
timeout --signal=KILL 90s go build ./cmd/adamic
timeout --signal=KILL 90s go vet ./internal/...
timeout --signal=KILL 90s go test ./internal/lower -run '^TestProgramCanonicalConstructorBoundary$' -count=1 -timeout 85s -v
timeout --signal=KILL 90s node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/program_region/canonical.a
timeout --signal=KILL 90s node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/program_region/canonical_counted.a
gofmt -l internal/lower/program_canonical_constructor_test.go
git diff --check
```

Final build and vet exited 0; formatting and diff checks were clean. The restored
refusal test passed. Each command's output was written to a log file. Included logs record the diagnostic,
Node controls, final refusal check, and guard-removal mutant (expected exit 1).
