Built a graph-result refusal, exclusive join design, parent binder and deterministic pool controls.
Branch runtime/step37-join-handoff, base 9bc8201f; query merge a111d464, scout merge 29d45001.
Full fresh/lower, filtered uncached oracle, counts, gofmt and vet passed; logs are attached.
Worker-survivor, reverse-merge, skipped binder write and three isolated guard mutants were caught.
Stopped: missing whole-graph return inventory; cyclic handoff and Program adoption are not built.

## Findings

The cyclic SourceElements reduction built before the guard, then native printed
`adamic: panic: a graph region can't cross into parallel work yet`, exit 70.
Source Node exited 0 and printed all 128 files after the parent changed declaration
values. The initial native build used the CLI's shipped -O2/ThinLTO policy.
After the guard it refuses at cyclic.a:51:16 before either backend emits code,
with the default-off ownership query both disabled and enabled. A counted wrapper
around the returned graph also refuses. A worker storing a child in a global
is independently refused by the existing task effect proof.

The actual ownership extractor returns Unknown for graph demand because its
region inventory is incomplete. Complete supplied snapshots prove an exclusive
cyclic region, refuse a worker root reaching a child or disconnected member,
refuse a Weak observer, and return Unknown for incomplete inventory. These
normalized proofs are not a worker-return extraction implementation.

The supported acyclic reduction returns mutable nodes and binds them on the
parent. The pool reduction creates separate intern Maps and caches per file,
returns name/use projections, then remaps on the parent in file/insertion order.
Canonical spelling order is beta, alpha, __proto__, constructor, βeta, 𐐀name,
gamma. Both controls and source mutants agree with their own Node executions
and the JavaScript backend. The controls match pristine Node at 1, 4 and 16
threads, with ASan/UBSan and internal/leakcheck; Linux controls pass TSan three
fresh runs at each size. The main oracle also passes its ordinary release,
ASan, sanitized slabs and TSan variants at one thread and the default pool.
Ordinary test release uses -O2 without ThinLTO; sanitized lanes use -O1 with
ASan/UBSan, separate TSan builds instrument the runtime too.

Only two count rows were added by this change. Existing rows did not move.
The scout merge separately brings its three already recorded rows.

| New fixture | Allocations | Frees | Retains | Releases | Peak | Statement regions | Graph regions | Merges |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| step37_bind.a | 1799 | 1799 | 1288 | 2317 | 523 | 0 | 0 | 0 |
| step37_intern_merge.a | 2498 | 2498 | 2831 | 3286 | 861 | 0 | 0 | 0 |

## Mutants run

| Mutant | Observed catcher |
| --- | --- |
| Worker keeps a declaration in a global | Compile-time task effect refusal with query off and on; source Node still matches the control |
| Reverse file pool order | Compiles, matches its own Node, exits 0 and stays leak-clean at 1/4/16; pristine Node stdout differs |
| Omit parent's binder update | Same compiling/clean conditions; pristine Node stdout differs |
| Remove post-classification result guard | Compiling Go overlay; refusal test fails with `graph handoff reached emission: <nil>` |
| Ignore second surviving owner | Compiling Go overlay; complete snapshot's kept-child and disconnected-member checks fail |
| Ignore inventory completeness | Compiling Go overlay; incomplete snapshot returns Proven and fails its check |

No production file is edited by the overlay runner. No build-error kill counts.
The generic frontier rule is conservative; no separate generic source mutant
was run. The future runtime handoff/count-ending protocol was not implemented
or mutated.

## Commands and observed output

All test output was written to logs. Every test/build shell sourced
`/workspace/adamic-tools/env.sh`. Logs here are gzip-compressed exact outputs.

```sh
bash cloud/setup.sh > /tmp/step37-final-setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/fresh ./internal/lower -count=1 -timeout 30m > /tmp/step37-analysis-packages.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/step37-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestStep37Join|^TestScout37ShapesAndMutants$|^TestOwnershipQueryNodeShapes$|^TestGraphParallelMapRefusesRegions$|^TestNativeAgreesWithNode/internal/oracle/testdata/concurrency/accepted/graph_regions_per_task' -count=1 -v -timeout 30m > /tmp/step37-final-focused.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/fresh ./internal/lower ./internal/oracle -run '^TestStep37Join|^TestOwnershipQueryMoveAgreement$|^TestNativeAgreesWithNode/internal/oracle/testdata/concurrency/accepted/step37_' -count=1 -v -timeout 30m > /tmp/step37-final-gate.log 2>&1
python3 internal/oracle/testdata/step37/prove.py > /tmp/step37-final-mutants.log 2>&1
gofmt -l cmd internal > /tmp/step37-final-gofmt.log 2>&1
go vet ./... > /tmp/step37-final-vet.log 2>&1
git diff --check
```

Observed exits were 0. Full analysis packages: fresh 34.735s, lower 126.394s.
Counts regeneration: oracle 119.950s. Expanded uncached oracle: 198.217s.
Final filtered gate: fresh 0.008s, lower 80.339s, oracle 4.712s. Move agreement
covered 79 fixtures, one admitted IR transfer and four independent query
refusals. Final mutant runner reports three compiling mutants caught. Formatting,
vet and whitespace checks print nothing.

Final setup: Go ready 0.022s, Node ready 0.022s, submodules ready 0.066s,
markdown ready 0.073s,
clang ready 0.178s, go build ready 39.783s, cache warm 39.942s, done 39.971s.
`nproc=5`, `cpu.max=400000 100000`, memory 17.6 GB. Tools are Go 1.27.1,
Node v24.19.0 and clang 20.1.8. A prior setup overlapped dependency merging and
failed on stale compiler package definitions (`could not import os`, undefined
QueryOwnershipTransfers/Proven and missing ParallelMap.Site); current-setup.log.gz
records that failure. The final setup above ran after those merges and passed.
The checkout's main-only fetch refspec initially left area/runtime at a stale
4271aab5; explicit remote-tracking ref updates corrected the base before changes.

## Limits

Named stop reason: **complete dynamic-region and worker-return root inventory
is missing**, so there is no Proven extracted return certificate to authorize
cyclic graph handoff. Existing plain graph counts and shared-publication backstop
remain unchanged. Step 06 at ff91354d is not merged; only its parent-only
zero-copy adoption seam is designed in [step37-join.md](../step37-join.md).
The ownership-query flag stays off by default. Protected lower.go changes come
only from the requested query starting-point merge, not this guard implementation.

The full repository gate, full oracle package, macOS, WASI/Workers, whole 77-file
parser, constructor/method task admission, persistent worker caches, general
return extraction, runtime transfer cleanup and parallel checking were not run
or built by this unit. The positive AST fixture is acyclic and uses existing
atomic shared publication. No parallel parsing throughput or speedup is claimed.
