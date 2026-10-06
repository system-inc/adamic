# Object prototype integration review

Base: codex/library-object at cdf632b47fefa315aefdeaeda468c13abed4d551.
Linux: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.
`bash cloud/setup.sh`: go, clang, node and submodules ready, each 0s.

## 1. Private mangled slots

`adamic_object_has` skips shape names beginning with #. The JavaScript
backend uses `adamicHasOwn`, evaluates both arguments once, and filters the
same private names before `Object.hasOwn`.
The supplied inherited-private fixture prints `false false true` on Node,
sanitized native, release native (-O2), and the JavaScript backend.

Uncached command: `go test -count=1 -timeout 10m ./internal/oracle -run
TestNativeAgreesWithNode/internal/oracle/testdata/library_object_private_mangled`.
Passed (9.321s). Mutant scanning # slots again: Node `false false true`,
native `true true true`, caught only by stdout comparison (9.213s).
No shared lowering or emission hooks changed for this item.

## 2. Collection iterators

Only `prototype.go` changes iterator behavior. MapIterator and SetIterator
get their intrinsic tags. Own-property reads evaluate the receiver and key
once, in order, and return false for all keys: next is inherited and internal
iterator slots are not observable. Result objects keep own done and value.
A structural view that hides a collection iterator is refused with a reason.
No Map/Set lowering or runtime files changed.

Both supplied fixtures passed uncached Node/native/release/backend comparison
(0.439s). Mutant returning the native own-slot answer again: stdout differs
on both native and JavaScript backend. Node prints `false false`, `true true`,
`false false false`; mutant prints `true true`, `true true`, `true false false`.

## 3. Null-prototype named groups

Known RegExp named-group dictionaries now refuse inherited Object calls with
an explicit NotYet reason: null prototype and Node TypeError. Both supplied
probes are registered as refusals and independently require that reason.
When the guard is mutated away, the refusal check executes the accepted
artifact against Node. Both probes fail only by behavior comparison:
Node exits 70 through the oracle wrapper's TypeError handling; native exits 0
printing `true` and `[object Object]`, or just `[object Object]`.
Uncached targeted tests passed (0.123s); mutant caught in 1.290s.
This does not implement catchable native TypeError or arbitrary prototype
chains. No RegExp runtime or lowering file changed.

## 4. Failure and boundary fixtures

`library_object_try_assign.a` is the supplied frozen-target probe, registered
as NotYet and independently held to the Object.assign failure reason. No
change to exceptions.go remains in the branch. The missing libraryFailure
assign guard mutant prints nothing and panics (exit 70); Node prints `caught`
and `1` (exit 0). The behavior comparison catches it.

`library_object_prefix_names.a` has own names a and ab, queried dynamically
alongside abc and ac. The exact-length check is unchanged. A prefix mutant
allows shorter shape names to match longer queries: both missing names become
own/enumerable. Only stdout comparison catches it. Uncached item tests passed
(0.497s), prefix mutant caught (10.202s), assign mutant caught (0.319s).

Boundary coverage also holds two representation refusals. A public # property
cannot be distinguished from reserved private shape names by the runtime's
filter, so own-property calls on such shapes refuse. Without that guard Node
prints `true true` and native `false false` (0.307s). A structurally hidden
iterator refuses rather than receiving the plain-object tag. The original
empty-object-view mutant was masked by the generic primitive-view refusal;
the narrowed Pick<MapIterator<[string, number]>, 'next'> view isolates the guard.
Its mutant prints `[object Object]` against Node's `[object Map Iterator]`
(0.336s). These are explicit refusals, not approximated results.

## Mutants and reproducibility

All final mutants below compiled and ran. Each was caught by comparison with
Node, not clang, a sanitizer, or a check that lowering alone failed. Files were
restored after each run. Raw receipts are in [prototype-evidence](prototype-evidence/).
`run-prototype-mutant.py` accepts file, old text, new text, oracle test pattern,
and label. Source the toolchain, set ADAMIC_GATE_UNCACHED=1, and redirect its
output to a log. It saves and restores the file in a finally block.

| Mutant | Fixture/check that caught it | Difference |
|---|---|---|
| Native scans # slots again | private_mangled | stdout false false true vs true true true |
| JS scans # slots again | private_mangled | backend stdout false false true vs true true true |
| Iterator next own again | iterator_own | next becomes true on native and backend |
| Map iterator given Object tag | iterator_tag | Map Iterator vs Object on native and backend |
| Groups treated as ordinary objects | TestObjectGroupsHaveNoPrototype, both probes | Node TypeError vs successful native output |
| Missing Object.assign libraryFailure guard | TestObjectAssignFailureStaysRefused | caught/1 and exit 0 vs panic and exit 70 |
| Prefix matching replaces exact length | prefix_names | abc and ac incorrectly own/enumerable |
| Public # name accepted | TestObjectPrototypeRepresentationsStayDistinct/public_hash | true true vs false false |
| Hidden iterator accepted | TestObjectPrototypeRepresentationsStayDistinct/iterator_view | Map Iterator vs Object |

No V8 algorithm was ported. Runtime change is confined to object.c; prototype
lowering changes are confined to prototype.go. The sole backend dispatch hook
is ir.HasOwn in javascript.go, calling the small adamicHasOwn helper. Fixtures
register in library_object_prototype_test.go; shared oracle_test.go and
cmd/adamic-test262 are unchanged. map.c and map_set.c are unchanged.

## Linux validation and measurements

All test output went to log files. The complete repository gate was not run;
the following changed-package and complete Object-oracle checks were run instead:

| Command | Result |
|---|---|
| `gofmt -l cmd internal` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `go test -count=1 -timeout 30m ./internal/lower ./internal/javascript ./internal/native` | lower passed 10.676s; JS has no package tests; native passed 83.454s |
| `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestObject\|TestNativeAgreesWithNode/internal/oracle/testdata/(library_object_\|object_prototype\|has_own)'` | passed 6.789s |
| `go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts` | passed 19.510s; one new prefix fixture row plus three previously measured rows put in generated order; no existing counts changed |
| `git diff --check` | exit 0, no output |

The oracle runs source on Node, sanitized native, release native (-O2), and
the JavaScript backend, and checks leaks on successfully finishing programs.
Four added passing fixtures agree in all those checks; five added refusal
fixtures stay explicit NotYet. Refusal-regression tests execute a mutant
artifact against Node when lowering wrongly accepts it.

The unmodified test262 runner measured built-ins/Object with `-adapt -json`,
checkout `/workspace/scratch/test262` at
5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81. Before was an isolated git archive
of cdf632b using the same toolchain and pinned cohere dependency; after was
the current branch. Runner command:

```sh
go run ./cmd/adamic-test262 -test262 /workspace/scratch/test262 \
  -root <before-archive-or-current-repository> -adapt -json built-ins/Object
```

| built-ins/Object | pass | disagreement | refused | crashed | skipped |
|---|---:|---:|---:|---:|---:|
| Before | 58 | 0 | 2277 | 0 | 1076 |
| After | 58 | 0 | 2277 | 0 | 1076 |

The test262 totals are unchanged. The added fixtures reproduce the integration
holes independently; no newly passing test disagrees with Node.
Not covered: general prototype chains, catchable native TypeError for groups,
catchable Object.assign failures, or iterator spreading (another worker owns
that path). The branch continues to refuse unsupported reads with a reason.
