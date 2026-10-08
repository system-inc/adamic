# Checked views V1

This slice serves roadmap step 11, task #a03mesg. Its base and comparison are
`d72728e570fe09d91cf55b37b564dbad0fefc24d`. It rebuilds the frame from the lane
history. It does not merge the pending views integration branch.

V1 admits checked structural object, inherited interface and scalar field reads.
Tagged casts still check their tag at the cast, with `cast failed:` on a mismatch.
An untagged structural downcast adds checked reads without inventing a tag test.
Writable-slot and nominal-identity proofs remain required.

`view()` interns one-based contracts and records actual operands. `markViewRead`
keeps the declared checker type, contract id and source location until readiness
finalization. Recursive contracts refer to already interned ids. Zero and Unknown
supply no erasure proof. Unsupported members are descriptors, not certificates;
an unread member does not reject its containing object view. Demanded unsupported
callable reads refuse at the read through joined helpers, generics, callbacks and
stored fields. Unknown origins retain the refusal. Reads proved disjoint from all
viewed allocations keep ordinary dispatch.

The may-flow query is the lanes' shared allocation graph. V1 assigns allocation
identities and queries projections; it never selects graph ownership or changes
cycle admission. Base strong-backreference, capture and optional-write refusals
remain. Shared presence, physical-storage and readiness checks remain the runtime
boundary. No second readiness bitmap is introduced.

Both backends route field reads and calls through the frame boundary. This keeps
the existing counted and ordinary code-pointer signatures, optional/rest/count
layouts, receiver order, ownership and canonical closure cells. Union, array,
callable, dictionary and intersection builders return NotYet; their lazy
unsupported descriptors do not authorize execution. Erasure returns no proof.
Optional and mixed-union reads remain outside V1. An optional member that is never
read may remain in a lazy object descriptor. Existing index-signature frontend
refusals still precede view lowering.

The approved representation judgment reserves null=12, undefined=13, Record=14
and moves typed arrays to 15/16/17. V1 does not emit the new null/undefined/Record
representations. The C table and Go table are held together by
`TestRepresentationTagsAgree`. Numeric consumers of typed-array representations
use the Go symbols; their runtime element-flavor numbers are separate. The C
record cases 13 and 14 are string lengths, not representation consumers.

## Sources

These commits supply the selected hunks, not complete cherry-picks:

| Commit | Parts taken |
| --- | --- |
| df7bc5da3155fe7811f758d67aed0c7e97cdf0d8 | Nested object read checks, object fixtures and their controls |
| 94cf71560d582553d46c745945dc0045ce8e8b43 | Inherited and recursive interface contracts and fixtures |
| de8b2ec763fb781bd35e42d21e30238d0b18af07 | Lazy descriptors, read obligations and actual view origins |
| f7100da75d253a0ff33b00dc58e4d399b640f0ab | Wider-helper family fallback and helper/generic/callback/store controls |
| 59db5d63a4caebcfeca67322712691d2f1ae85c2 | Shared may-flow query and contract metadata only |
| e7f420293e94746095c2f1e2d14124484e1afb63 | Field projections in the shared query |
| 198b1271e7c9b48b20fe150c223222fe6ece9bbd | Pointer/map allocation-id copying |
| 6d61c52d9ff5f0ae8bcbabf5bd1c1fa7bb1ff892 | Allocation classification metadata only |
| f40970920539189c9180e5b3e4561c831aa922d8 | Shared call-target readers in the query |

The shared query is extracted from 7302410e's final form so the later call-target
repairs are retained. No graph-region selection, shape-check erasure or later
family admission is taken. 609ed395's union admission, f8ed279d's optional reads,
46a28295's writes, runtime 2724dabd, fresh_refused and graph_regions fixtures are
left out. The new readiness controls use class definite-assignment fields in .a;
no new .ts assertion witnesses are introduced.

## Applied merge judgments

The judgments come from c923d1cc's two merge-judgment documents. On changed files:

- IR h3/h4: distinct semantic and typed-array tags, with the typed-array reference
  predicate retained and Record's reference predicate reserved.
- cast.go: classified structural/interface plans retain base class proofs,
  sameKeeping and the original fallback refusal. Cast proof probing does not
  leave a dummy operand in the view-origin ledger.
- cast_proof.go: existing qualified-name and marker proofs stay; V4 marker
  admission is not added.
- cycles.go: all base cycle and captured-backreference refusals stay. Shared
  allocation queries do not admit graphs.
- object.go: base Union storage, erased-unknown array refusals, CallbackType,
  optional comparator proof and Map/callback refusals stay. No array-family
  ViewRead admission is added.
- JavaScript common dispatch: counted/plain signatures, receiver/count/rest
  layouts, canonical interior cells, shared readiness and ArrayIsArray stay.
- Native common dispatch: typed call pointers, receiver evaluation, owned holds
  and all callback refusals stay. There is no function-pointer cast adapter.

Judgments 24 through 27 leave the existing Map callback, detached intrinsic,
fixed-object enumeration and direct-overload extra-argument paths in the base;
their files need no V1 changes. Plain Error widening and missing optional-own-field
writes keep their base refusal. No graph fixture expectation is weakened.

## Evidence

The reduced lane fixtures retain their source Node observations. Invalid payloads
pin the inserted checked-view stop independently in native and JavaScript. Native
controls use ASan/UBSan and release C; successful controls also check leaks.

The .a audit checks every new source with the configured compiler, then records
its lowering outcome against the base compiler. No upstream census retirement is
inferred from contract eligibility.

The baseline full compiler run passes flow, lower and IR; JavaScript has no
package tests. The baseline full oracle has one failing leaf:
`TestNativeAgreesWithNode/internal/oracle/testdata/normalize_coverage_long.a`.
Its sanitized process exceeds the harness deadline. An uncached single-fixture
run repeats the timeout; release C and source Node exit normally. The generated
C is byte-identical on base and V1: 10,063 bytes, SHA-256
`db6c4874d179e8bdf4724d4c52bdfcbfd3427f9677678caca6a3d955e6066c8b`.
This establishes unchanged emission, not a passing sanitizer observation.
No deadline or expectation is weakened.

| Mutant | Independent pin that catches it |
| --- | --- |
| Drop nested field check | Wrong numeric payload must stop with exit 70 and the boolean-field diagnostic; release C and JavaScript continue when the marker is dropped |
| Drop readiness | Uninitialized `view.child.ready` must stop with exit 70; the mutated slot produces `false\n` with exit 0 in sanitized C, release C and JavaScript |
| Drop tagged-cast check | Wrong kind must stop at the cast with `cast failed:`; valid mutated C and JavaScript print `casting\notherother\n` and exit 0 |
| Evaluate cast operand twice | Source Node and both backends pin `once2` and `calls 2`; valid mutated C and JavaScript print `once3` and `calls 3` |
| Drop lazy demand guard, local Go overlay | Joined/wider/unknown/missing-origin lower tests fail their refusal pins; helper, generic, callback and stored-field oracle controls report that the unsupported read ran on |
| Change C Record tag from 14 to 20, local Go overlay | C compiles and runs; the independent Go/C table comparison fails on 20 versus 14 |

The direct callable helper control pins Node's `3\n` and the existing refusal
for a class method through a view that erases its prototype origin. That guard
runs before the lazy field guard; V1 does not admit the method call.

A failed reflection implementation in the new mutant harness and relative Node
cache paths were repaired before rerunning the controls. Those failed runs are
not mutant evidence. A new typed-array family test was corrected to expect its
existing dictionary classification; no production family admission changed.
The array-shaped interface probe also exposed a changed diagnostic in
`objects/05_polling_array_metadata.a`. Requiring actual object representations
before either object-view cast path restores its base unchecked-cast refusal.

The final focused suite passes: `TestCheckedViewObjects`,
`TestCheckedViewInterfaces`, `TestCheckedViewLazyUnread`,
`TestCheckedViewLazyHelperMutant`, both `TestCheckedViewFrame*` tests and the
base `TestInterfaceCastRuntimeMutants`. The command completes with PASS in
338.943 seconds; its log is `/tmp/views-v1-fixtures-complete.log`.

All 19 new .a files pass `adamic types`. Fourteen lower on V1; five keep a
Refused/NotYet result at their demanded use. Each is also compiled by the d72728e5
binary. The comparison is `/tmp/views-v1-a-comparison-final.json`.

| New source group | Base | V1 |
| --- | --- | --- |
| lane1/interfaces-good, interfaces-missing-inherited, interfaces-uninitialized-object, interfaces-wrong-inherited | NotYet nested Child field | Compiles with checked descendants |
| lane1/objects-good, objects-missing-nested, objects-uninitialized-nested, objects-wrong-nested | NotYet nested child field | Compiles with checked descendants |
| lane1/objects-untagged-good, objects-untagged-wrong | Refused unchecked cast | Compiles with checked structural reads |
| lazy/unread, ordinary-disjoint | Refused callable member at cast | Compiles; callable member is unread on viewed allocations |
| lazy/unread-untagged | Refused unchecked cast | Compiles with checked structural reads |
| lazy/optional-unread | NotYet optional member at cast | Compiles; optional member is never read |
| lazy/helper-mutant, generic-mutant, callback-mutant, field-mutant | Refused callable member at cast, line 8 | Refused unsupported callable at demanded read, line 9 |
| lazy/call-mutant | Refused callable member at cast, line 8 | NotYet prototype-origin method call, line 9 |

Only three existing stage3 records move to Compiles:
`assertions/11_parenthesized_kind.a`, `assertions/13_flag_downcast.a` and
`assertions/20_type_flag_mask.a`. Their source Node outputs are respectively
`110\n110\n`, `8\n0\n` and `2\n0\n`. The existing lower test
`TestDefaultTaggedInterfaceAdmission/optional_target` moves from NotYet to
admitted because it never reads its optional member. Optional member reads and
writes remain outside V1. Each moved result is named in the commit trailers.

`TestCountsAreRecorded -args -update-counts` passes in 1,053.212 seconds.
The table adds exactly 17 rows: the fourteen newly lowered .a sources and the
three moved stage3 reductions. Every old row remains byte-identical; none is
removed. `/tmp/views-v1-counts-final.jsonl` records the full refresh.

The pinned cohere submodule is built directly, without copied source. Its types
only audit checks the nineteen V1 sources and prelude through a scratch config:
`20 checked`, exit 0. `/tmp/views-v1-cohere-a-audit.log` records that observation.
This is a type audit; lint and formatting are not reported as checked.

Both final stage3 runs pass. The normal run completes in 678.262 seconds; the
locally unguarded run completes in 180.091 seconds. Each preserves all 609
baseline test results and adds exactly three native runs, one for each admitted
assertion reduction. No other stage3 record changes. Both worktrees run
`npm ci` in `stage3/api` before the fixture tests.

The local Go overlay changes only the platform skip condition to `false && ...`.
It is not committed. The comparison also includes the baseline normal run
(751.565 seconds) and locally unguarded run (588.251 seconds).

The base and V1 .a type results all match, exit 0. Their cohere type audits both
report `20 checked` and exit 0. The protected compiler files and cohere submodule
have no source changes. The changed Go files pass `gofmt -l` with empty output
and the diff passes `git diff --check`.

The final full compiler run passes flow, lower and IR; JavaScript again has no
package tests. All 3,267 baseline test results match; the sixteen added lower
results pass. This includes the family no-proof, joined lazy-demand, descendant
contract and source-slot write controls.

| Check | d72728e5 | Final V1 |
| --- | --- | --- |
| internal/flow | PASS, 2,057.910 seconds | PASS, 1,060.854 seconds |
| internal/lower | PASS, 609.052 seconds | PASS, 146.789 seconds |
| internal/ir | PASS, 9.031 seconds | PASS, 3.623 seconds |
| internal/javascript | No test files | No test files |
| Full internal/oracle | One normalization timeout, 3,361.759 seconds | Full comparison still running |
| stage3/fixtures, normal | PASS, 751.565 seconds | PASS, 678.262 seconds |
| stage3/fixtures, platform guard lifted locally | PASS, 588.251 seconds | PASS, 180.091 seconds |
| Changed .a type audit | Nineteen source files and cohere's twenty-file scope pass | Same |
| Counts | Existing table | Seventeen added V1 rows; no old row changes |

The final compiler log is `/tmp/views-v1-packages-complete.jsonl`; the baseline
log is `/tmp/views-v1-base-packages-serial.jsonl`. The full oracle comparison is
recorded after its run finishes.


## Commands and machine

All test output is redirected to logs. No full repository gate is run. The V1
unit explicitly requests the full named packages and full oracle.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-v1-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc

# Both baseline and final V1, in their respective worktrees.
go test -json -count=1 -p=1 -parallel=2 -timeout=90m \
  ./internal/flow ./internal/lower ./internal/ir ./internal/javascript > LOG 2>&1
# Baseline uses its default test parallelism; final V1 bounds it at 2.
go test -json -count=1 -timeout=90m ./internal/oracle > BASE_LOG 2>&1
go test -json -count=1 -p=1 -parallel=2 -timeout=90m ./internal/oracle > V1_LOG 2>&1

# Both baseline and V1; npm ci runs in stage3/api first.
npm --prefix stage3/api ci > NPM_LOG 2>&1
go test -json -count=1 -p=1 -parallel=1 -timeout=90m ./stage3/fixtures > LOG 2>&1
go test -json -count=1 -p=1 -parallel=1 -timeout=90m \
  -overlay LOCAL_PLATFORM_OVERLAY ./stage3/fixtures > LOG 2>&1

# V1 focused controls, including the four required valid-code mutants.
go test -v -count=1 -p=1 -parallel=2 -timeout=45m ./internal/oracle \
  -run '^(TestCheckedView(Objects|Interfaces|Lazy|Frame)|TestInterfaceCastRuntimeMutants)' > /tmp/views-v1-fixtures-complete.log 2>&1

go test -json -count=1 -p=1 -parallel=2 -timeout=90m ./internal/oracle \
  -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/views-v1-counts-final.jsonl 2>&1

# Local overlay mutants; each command exits nonzero on its independent pin.
go test -v -count=1 -overlay /tmp/views-v1-mutant-demand.json ./internal/lower \
  -run '^TestLazyViewDemandUsesSharedFlow$' > DEMAND_LOG 2>&1
go test -v -count=1 -overlay /tmp/views-v1-mutant-demand.json ./internal/oracle \
  -run '^TestCheckedViewLazyHelperMutant$' > HELPER_LOG 2>&1
go test -v -count=1 -overlay /tmp/views-v1-mutant-tags.json ./internal/native \
  -run '^TestRepresentationTagsAgree$' > /tmp/views-v1-mutant-tags.log 2>&1
go test -v -count=1 ./internal/native -run '^TestRepresentationTagsAgree$' > /tmp/views-v1-tags.log 2>&1

# Uncached inherited normalization timeout, observed on the base.
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -p=1 -parallel=1 -timeout=30m ./internal/oracle \
  -run '^TestNativeAgreesWithNode/internal/oracle/testdata/normalize_coverage_long[.]a$' > /tmp/views-v1-base-normalization-uncached.jsonl 2>&1

# The pinned submodule binary; source is referenced, not copied.
go -C cohere build -o /tmp/views-v1-cohere ./command/cohere > /tmp/views-v1-cohere-build.log 2>&1
/tmp/views-v1-cohere --types --no-fix --no-cache --single-threaded \
  --directory /tmp/views-v1-build --tsconfig /tmp/views-v1-a-audit-tsconfig.json > /tmp/views-v1-cohere-a-audit.log 2>&1
# The same command is repeated with --directory /workspace/adamic for the base.
```

The fixture audit builds separate base and V1 `cmd/adamic` binaries, runs
`types` and `c` for every new .a file, and stores their exit codes and diagnostics
in `/tmp/views-v1-a-comparison-final.json`. The normalization emission comparison
uses those binaries too. Changed Go files are checked with `gofmt -l`; the patch
is checked with `git diff --check`. Both produce empty logs.

Setup runs on f0c6e6fc before the base instruction changes; every reported
comparison uses d72728e5. Its measured ready lines, in seconds: node 0.071,
Go 0.086, markdown 0.199 (validated cache step 0.026), submodules 0.205,
clang 0.583, Go build 149.709, test binaries deferred 149.989,
build cache warm 149.995, done 150.126. `nproc` prints 5;
`cpu.max` is `400000 100000`. The toolchain is Go 1.27.1, Node 24.19.0 and
clang 20.1.8. Setup succeeds and prints `/workspace/adamic-tools/env.sh`.

Initial unbounded baseline runs exceed process/test deadlines under concurrent
load. The bounded full baseline compiler and stage3 reruns replace those runs.
The baseline oracle's normalization failure remains explicitly recorded. Early
V1 reruns exposed the diagnostic and harness issues described above; no failing
run is substituted for the final controls. A first count refresh adds 13 V1 rows;
the final refresh adds the remaining four and preserves every old row.
