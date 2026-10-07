Built transitive strong-field reads, borrowed field declarations, and retained ownership at stores, returns and captures.
Implementation c7bc30b; current-main integration 50dedcd (origin/main e8ba3d5); branch codex/borrow-chains.
Integrated native PASS 122.617s, complete uncached oracle PASS 108.626s, counts PASS 24.858s; batch8 removes 15,334,163 retain/release pairs.
Eight mutants caught: five compiler lifetime mutants and one capture-C mutant by ASan; positive-borrow and recorded-count mutants by their assertions.
Not covered: weak-target ownership, borrowed call-result provenance, parameter conventions, for-of borrowing or devirtualization; full-gate status below.

# Borrow chains

## What changed

`borrow.go` extends the existing lent-read emission to strong property chains rooted
in a stable local or parameter. `chainRoot` follows property loads and `Defined`,
not weak reads, optional receivers or call results. `chainUnchanged` checks the
whole function and every reachable named target using `CallTargets`; bounded
callbacks use `ClosureTargets`, and Unknown refuses. Recursion is checked once per
target, including every statement in its body. A matching `SetProperty` refuses,
as do unknown operations and object spreads. A field name matches every holder,
which is conservative because native IR does not carry related-holder types.

A root must belong to this function, be neither global nor captured, and have no
assignment anywhere in its body. Emission also refuses roots consumed, moved or
reused by the existing reuse plan. Stable owned locals hold their own counts;
existing borrowed parameters rely on their callers. Parameter conventions are not
changed: ownership transfer still goes through the existing helpers.

`element_borrow.go` also plans property-initialized locals, excluding captured or
reassigned declarations. They get `Local.Borrowed`, with their root marked lending
before reuse. Their declaration snapshots the field without holding or releasing
it. A store, owned return or capture cell still retains what it keeps. Passing a
safe chain to a borrowed parameter removes its argument temporary count. The
`consumes` and `pureKind` lists are unchanged. No edits to emit.go, lower.go,
native.go, function signatures, for-of emission, returns or devirtualization.

The proof is intentionally whole-function rather than live-range precise. A write
before the read, an unrelated holder with the same field name, an object spread,
or an unknown runtime operation can prevent an otherwise safe borrow. These are
missed optimizations, not permission to borrow through uncertainty.

## Oracle and counts

All nine new programs are `.a`. They run from source on Node, native with ASan and
UBSan plus leak checking, and the JavaScript backend on Node. Strings are built at
runtime. Positive cases cover recursive children and multiple ancestor levels,
and a chained argument to a read-only function. The listener fixture uses a class
through a method-signature interface. This base gives that structural call Unknown
targets, so its argument conservatively remains counted.

The original property-style listener signature was refused by lowering as erasing
the prototype origin. The supported method-signature form passed. No lowering or
interface-target inference was added here.

| Fixture | Retains before | After | Releases before | After |
|---|---:|---:|---:|---:|
| borrow_chain_walk.a | 22 | 13 | 31 | 22 |
| borrow_chain_argument.a | 1 | 0 | 5 | 4 |
| borrow_chain_listener.a | 3 | 3 | 8 | 8 |
| borrow_chain_write.a | 2 | 1 | 6 | 5 |
| borrow_chain_reassigned.a | 2 | 1 | 6 | 5 |
| borrow_chain_override.a | 6 | 4 | 14 | 12 |
| borrow_chain_unknown.a | 3 | 3 | 8 | 8 |
| borrow_chain_capture.a | 4 | 4 | 7 | 7 |
| borrow_chain_store.a | 5 | 5 | 12 | 12 |

The refused parent/argument/capture retains stay. The decreases in the write,
reassignment and override fixtures are other stable reads, including the saved
parent's kind; they do not remove the lifetime-protecting retain.

[fixture-counts.csv](evidence/fixture-counts.csv) contains before/after retains and
releases for every one of the 320 pre-integration fixtures, including the new programs measured
with the base compiler. 31 rows change: 513 fewer retains and 508 fewer releases,
no rise on any row, and no allocation/free/peak/region change. The one unequal
pair delta is `literal_optional_shapes.a` (40/68 to 33/66): five absent optional
text reads in `show` lose retain(NULL) calls, whose NULL temporaries `??` already
passed on without a matching release. Two ordinary pairs also disappear there.

## Mutants actually run

Each compiler mutant starts from the finished source and restores it in a finally
block. Sanitizer failures, not compiler errors, are required for lifetime mutants.
[tools/mutants.py](tools/mutants.py) preserves a complete log for each invocation.

| Mutant | Fixture/check | Observed failure |
|---|---|---|
| Ignore matching field writes | borrow_chain_write.a | ASan heap-use-after-free |
| Same unsafe borrow into a storing override, before a sibling destroys its holder | borrow_chain_store.a | ASan heap-use-after-free |
| Ignore root reassignment | borrow_chain_reassigned.a | ASan heap-use-after-free |
| Consult only the static virtual target | borrow_chain_override.a | ASan heap-use-after-free |
| Treat Unknown callback targets as empty | borrow_chain_unknown.a | ASan heap-use-after-free |
| Store the captured chained string in its generated cell without retain | borrow_chain_capture.a | ASan heap-use-after-free |
| Disable chain declarations | TestBorrowChainDeclarations | parent, kind and children did not borrow |
| Record 14 walk retains instead of 13 | TestCountsAreRecorded | recorded 14, measured 13 |

The capture mutant changes only the retain at `adamic_cell_new` in generated C,
leaving cell layout, closure capture and cleanup intact; this isolates the lifetime
check rather than triggering a missing-cell C compilation error. See
[tools/capture-mutant.py](tools/capture-mutant.py). All eight mutants failed their
intended checks. No mutant survived or was credited for a compile-time failure.

## Batch8 release measurement

The report's scratch driver branch is not advertised remotely. The same four
inputs were merged in an isolated worktree from this unit's exact base: batch8
4189abd, batch4-typescript 63782c5, integer fast paths a183e50, and numeric Map
hashing e7ea1a4. Scratch merge head is d360f5d8ee7f6a727f9f8bb99986de736004022f
at `/workspace/borrow-chains-profile`. The after compiler differs only by this
unit's borrow.go and element_borrow.go. Runtime files are identical before/after;
the profiling worker's runtime fixes are not overlaid. Other unit branches are
not merged into the measurement checkout. The landing branch separately includes
current-main devirtualization; these timings isolate this unit before integration. This isolates chains against f64641f, rather than claiming the profile
report's later runtime wall times as this baseline.

The pinned input is all 77 TypeScript compiler files at v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Go cohere is 715ba94 and its
TypeScript-go pin is unchanged. Both drivers report 161 findings. Whole output,
including findings and fixes, matches Go and Node byte for byte before/after, and
the changed driver passes ASan/UBSan with empty stderr: 11,442,907 bytes, SHA-256
886ab0d88967bb93f678afa133e8583e77a0b533de437b5ed4ed3f8ae698da49. Different
absolute corpus paths explain the byte-size difference from the prior report.

[tools/measure.py](tools/measure.py) builds timed executables with clang 20.1.8,
C11, `-O2`, no `-g` and no sanitizers. It alternates before/after order in five
interleaved before/after/Go rounds. Compilation is excluded; startup, reading,
parsing and rule execution are included. `--count` here selects findings-only
output; timed binaries have no ADAMIC_COUNT or visit instrumentation. Separate
untimed binaries use ADAMIC_COUNT and a counter inserted into the generated
recursive visit function. The exact complete clang command lines, including every
runtime translation unit, are in [clang-commands.txt](evidence/clang-commands.txt)
and [clang-commands.json](evidence/clang-commands.json). PATH resolves clang to
`/workspace/adamic-tools/llvm/bin/clang`.

| Metric | Before | After | Go |
|---|---:|---:|---:|
| Best of five, seconds | 2.165802 | 2.113378 | 0.341332 |
| Ratio to Go | 6.345x | 6.192x | 1x |
| Retains | 110,222,456 | 94,888,293 | |
| Releases | 100,087,461 | 84,753,298 | |
| Recursive visitor calls | 887,803 | 887,803 | |
| Whole-driver retains per visitor call | 124.151930 | 106.879897 | |
| Whole-driver releases per visitor call | 112.736115 | 95.464082 | |
| Allocations / frees | 8,203,664 / 8,203,664 | 8,203,664 / 8,203,664 | |
| Peak live | 730,193 | 730,193 | |

The 15,334,163-pair reduction is directly observed (13.912 percent of retains).
The 2.421 percent best-time change is small and may be noise; it does not close the
native-versus-Go gap. Retains per visitor call divide *all driver ownership work*,
including parsing/scanning, by the observed visit count. They are not a claim that
all those retains execute inside visit.

Machine: Linux x86_64, AMD EPYC 7763, Go 1.27.1, Node 24.19.0, nproc 5,
cgroup CPU quota 400000/100000 (four CPUs). No affinity or frequency pinning.
Load averages before timing: 1.30/2.94/2.23; after: 1.20/2.79/2.19. No task
compilation or tests ran during the timing rounds. The redundant recursive
submodule fetch was stopped before timing. Raw five samples are in
[timing.json](evidence/timing.json); full parity hashes are in
[parity.json](evidence/parity.json).

## Which compiler work remains

Chains already remove counts without the other units. The finish line of a
read-only visitor counting zero is not reached. Current generated visit still
retains its returned node once through Parser.node, then retains its children
array and retains that array again for the iterator, releasing all three.
There are exactly 887,803 visits: this particular path still contributes
2,663,409 pairs, including 1,775,606 array pairs. Its elements are numbers, so
borrowed reference-element bindings alone remove zero of these visitor pairs.
Iterator borrowing or moving the already-owned array temporary is also needed.

[tools/dependencies.py](tools/dependencies.py) observes 14,879,759 Parser.node
calls across the complete driver, each with one owned node return. That is the
current count scheduled at this accessor, not a measured promise that every call
can borrow. The borrowed-accessor-return unit must prove the holder survives;
chainRoot then needs that return's owner/path provenance to follow fields of the
borrowed call result safely. A boolean borrowed-return flag alone cannot justify
an entire-function borrow across a later write to the accessor's holding field.
Devirtualization may bound structural listeners; their current Unknown calls stay
counted. This unit did not implement those flags, return rules or target inference.

## Commands and gate status

Every test writes directly to a log. Shells source
`/workspace/adamic-tools/env.sh`. `bash cloud/setup.sh` succeeded: Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 107s, total
107s, five processors. The profiling artifact helper initially rejected successful
Go dependency-download stderr; the cached rerun passed (before 37.217s, after
28.124s). A wrong scratch copy briefly caused missing-helper compile errors and
was corrected before generating either measured artifact.

```sh
go test ./internal/native -count=1 -timeout 30m > /tmp/borrow-chains-native-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(borrow|element|weak|region|exceptions)' -count=1 -timeout 30m > /tmp/borrow-chains-oracle-final.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/borrow-chains-counts-final-check.log 2>&1
gofmt -l cmd internal > /tmp/borrow-chains-gofmt.log
go vet ./... > /tmp/borrow-chains-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/borrow-chains-full-gate.log 2>&1
```

Native PASS 81.840s; filtered uncached oracle PASS 4.275s; counts PASS 4.532s.
Formatting, vet and git diff --check have empty output. The complete uncached gate
was stopped to integrate current main after native PASS 257.306s and complete
uncached oracle PASS 225.932s. Remaining stage1 packages were not completed. Baseline and changed
counts were each regenerated with `-args -update-counts`; the baseline was saved
before restoring the changed compiler. A counts check overlapped an intermediate
listener edit and failed loading that fixture; the stable final set was then
regenerated and checked successfully. Earlier fixture API mistakes and one
synthetic test lacking local metadata were corrected before the final gates.

### Landing on current main

The implementation commit is `c7bc30b`; merge `50dedcd` brings only this unit
branch up to origin/main `e8ba3d5d81de4d3773c723914fccd4c76248b965`.
No main or area branch is modified or pushed. The generated counts conflict was
resolved by rerunning the complete counts updater on the combined compiler.

```sh
go test ./internal/native -count=1 -timeout 30m > /tmp/borrow-chains-main-native.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(borrow|element|weak|region|exceptions)' -count=1 -timeout 30m > /tmp/borrow-chains-main-oracle.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/borrow-chains-main-counts-check.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/borrow-chains-main-full-oracle.log 2>&1
go vet ./... > /tmp/borrow-chains-main-vet.log 2>&1
```

Native PASS 122.617s, filtered uncached oracle PASS 8.784s, counts PASS 24.858s;
vet has empty output. All six source mutants were rerun and killed on the
integrated branch, as was the capture mutant with its CLI rebuilt from this tree.
The recorded-count mutant was also rerun against the final compiler in the
isolated checkout and caught recorded 14 versus measured 13. The optional full repository gate was not restarted.
A separate current-main checkout with all nine new fixtures supplies the
post-integration baseline, without changing this branch's source during testing.
Complete uncached oracle PASS 108.626s. The 321-row comparison changes 32 rows,
removing 514 retains and 509 releases with zero rises and identical allocation,
free, peak-live and region counts; see
[current-main-fixture-counts.csv](evidence/current-main-fixture-counts.csv).
The additional changed row is the newly landed devirtualize fixture. Its current-main
baseline is retained in [current-main-counts-before.md](evidence/current-main-counts-before.md).
Run `python3 tools/count-mutant.py /path/to/checkout` with the final compiler
and final counts table in that checkout to reproduce the counts control.

Mutants run via `python3 tools/mutants.py` and `python3 tools/capture-mutant.py`
from this report's directory; from repository root use their full paths.
The tools locate the repository from their own paths. Capture reproduction first
needs the changed CLI at `/tmp/borrow-chains-after`. Profile reproduction uses the
prior report's batch8_artifacts.go.fixture as runtime_profile_test.go in the scratch
lint package, with ADAMIC_TYPESCRIPT_SOURCE and ADAMIC_LINT_PROFILE_DIR pointing
to the pinned corpus and before/after artifact directories. Run measure.py,
parity.py and dependencies.py with the paths shown in their argument lists.

Compressed complete logs are preserved under evidence. Other architectures,
precise related-holder aliasing, weak-target provenance, and borrowed accessor
integration are not verified by this unit.
