# Checked-view blocking families and remaining lane dependencies

Historical transitive census (superseded for build priority by the read-demand inventory below). Measured before the next family implementation on the pinned 2,936 sites. These
are overlapping transitive target-contract dependencies, not first blockers,
observed field reads, execution frequencies or successful lowering counts. A site
is counted once per family. Traversal stops at array/tuple and callable contracts,
which belong to lane 2, rather than walking their intrinsic library APIs.

| Blocking family | Tagged | Untagged | Sites blocked |
| --- | ---: | ---: | ---: |
| array or tuple contracts | 1758 | 1151 | 2909 |
| nullish members | 1758 | 1145 | 2903 |
| callable contracts | 1758 | 1132 | 2890 |
| optional properties | 1758 | 1131 | 2889 |
| dictionary contracts | 1758 | 1117 | 2875 |
| empty or unsupported field contract | 1758 | 1112 | 2870 |
| mixed primitive union | 1758 | 1110 | 2868 |
| object plus primitive union | 1758 | 1110 | 2868 |
| any field contract | 1758 | 1109 | 2867 |
| intersection field contract | 1758 | 1107 | 2865 |
| never field contract | 1758 | 1104 | 2862 |
| untagged object union | 1758 | 1104 | 2862 |
| union cast admission | 261 | 25 | 286 |
| nominal class fields | 0 | 18 | 18 |
| generic field contract | 0 | 6 | 6 |

Lane 1 now builds nullish members and optional properties together, followed by
mixed and untagged unions and full union admission. Lane 2 owns native array
reads/mutations and callable contracts. The array adapter/storage prerequisite is
already pushed; it does not remove the array blocker.

## Remaining families by site

Each row in `stage3/interface-downcasts/blocking-families-sites.json` now has a
second census column, `remaining_families`, beside the original `blockers`, plus
`remaining_owner`, `lane_projection` and `other_remaining_families`. Nothing has
been removed merely because a helper or descriptor exists. Complete source
lowerings remain **0 / 1,758 tagged and 0 / 1,178 untagged**.

Exclusive counts include every missing family. “Other unresolved” means finishing
both lanes alone still leaves a recorded blocker; its complete set remains in the
per-site table. The projected counts below answer which of these two lanes a site
needs, while retaining other blockers separately. They are not unlocked counts.

| Remaining ownership | Exclusive sites | Two-lane projection |
| --- | ---: | ---: |
| Only lane 1 | 2 | 2 |
| Only lane 2 | 32 | 32 |
| Both lanes | 23 | 2901 |
| Other unresolved | 2879 | separately recorded |
| No missing families | 0 | not a lowering claim |

Lane 1 families: nullish members, optional properties, mixed primitive union,
object plus primitive union, untagged object union and union cast admission.
Lane 2 families: array or tuple contracts and callable contracts. Dictionaries,
any/unknown/generic/intersection/never/empty contracts and nominal class fields
are other unresolved families; no worker completion is assumed for them.
The summary JSON groups all sites by their exact remaining set. Rerun this column
after each admitted family, removing only support validated on both backends.

The new native `adamic_array.element_kind` records storage, not logical element
conformance: 0 unknown, 1 number, 2 boolean, 7 packed number/undefined, 10 heap
pointers. Ordinary fresh array literals without spreads populate it; all other
producers currently remain unknown. Do not infer or overwrite it from a cast.
Mutations/copies/iteration/callbacks require explicit integration; unknown is no
proof. References must still be classified from each selected heap header, then
checked against the full logical element contract. No initialization bitmap was
added. Native struct size remains 56 bytes on this clang/Linux target; the byte
occupies padding at offset 33. Literal construction adds one byte store.

`internArrayViewContract` adapts lane 2 `viewArrayContract(node,target,buildElement)`
into `viewArrayContractHook`, reserving an id before recursion. The callback error
must propagate and child ids come from the common registry. Arrays of recursive
interfaces terminate; unknown element contracts remain unsupported. Shared view
admission is deliberately still disabled for arrays.

The census matches all 2,936 source spans and texts and stock TypeScript reports
zero diagnostics. TypeScript 6.0.3 source commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Per-site witnesses are blocking-families-sites.json; the executable census is
blocking-families.cjs, both under stage3/interface-downcasts. The existing rejection
audit still establishes zero complete eligible contracts. No whole-file lowering
pass or unlocked site is inferred from these dependency counts.


## Read-based demand measured by Lane 2

The transitive graph counts above are historical dependencies, not the scheduling
measure: an unread descendant needs no checked-view read contract. Following
c898009b, a full pinned compiler census seeded by all 2,936 downcast targets finds
2,825 (checker type, field) pairs and 25,848 source reads. Exact declared families,
read witnesses and provenance are in [Lane 2's per-pair table](../stage3/interface-downcasts/lane2/read-census/pairs.json).

| Lane 2 read family | Distinct pairs | Read sites | Coverage |
| --- | ---: | ---: | --- |
| array contracts | 334 | 3189 | partial |
| array element/consumer reads | 251 | 1602 | partial |
| callable contracts | 308 | 1503 | partial |
| array intrinsic contracts | 179 | 794 | partial |
| array own-property reads | 30 | 72 | missing integration |
| tuple contracts | 6 | 9 | missing |
| Lane 2 deduplicated total | 920 | 6362 | partial demand |

Rows overlap and are demand, not a claim that every read is blocked. Currently
refused array intrinsics specifically account for 35 pairs / 128 reads; join is
largest at 5 pairs / 49 reads, exceeding tuples' 9 reads. Equal static receiver
types elsewhere count conservatively. This is not allocation/alias flow or an
unlocked lowering count. Full family/lane tables, methodology and limits are in
[Lane 2's read census report](../stage3/interface-downcasts/lane2/READ_CENSUS_REPORT.md).

The c898009b compiler still eagerly builds descendant contracts and has global
array/callable use scans. Lane 1 owns the pending cast merge and lazy admission,
and will push with Lane 2's tip merged. Lane 2 must merge that SHA before claiming
unsupported members refuse only at the reached read, naming field and family.

## Read demand, October 7 lane 4 checkpoint

The old 2,936-site columns above must not set implementation priority. Lazy
contracts depend on reads, including reads inside shared helpers, generics,
callbacks and stored aliases. A lexical scan after each cast is unsound.

This checkpoint measures **conservative Unknown-fallback demand**, not exact
reaching-view counts. The pinned, unadapted TypeScript compiler has zero stock
checker diagnostics. Across executable compiler sources, the inventory records
66,960 explicit property/element/destructuring reads with receivers that may hold
objects, and 12,181 distinct checker `(type id, field)` pairs. Every receiver is
Unknown for reaching-view provenance in this checkpoint, so each retains the
cheap view-tag test. This intentionally includes unrelated object and namespace
receivers. It is an upper bound on explicit reads needing guards, not a claim
that a cast reaches all of them. It does not yet inventory implicit property
reads in object spread/rest or destructuring assignment. Exact closed-world
reaching-view counts and complete all-read coverage remain unfinished.

The enumerator contains no flow solver. A scratch-only receiver adapter now
reuses lane 3 c1f4c5a7's original source adapter and
`allocationFlowGraph.ReachingAllocations` from `shape_flow.go`/`graph_flow.go`.
The production loader and lowering output remain disabled in that overlay.
On the pinned **unadapted, native-checker-rejected** program it records 6,037
diagnostics, 412 allocation schemas and 69,621 receiver queries: 67,141 have
unknown allocation frontiers and 2,480 have known allocation sets. These are
not lane 3's adapted-program diagnostic counts.

Exact UTF-16-to-UTF-8 span joining matches all 66,148 property/element inventory
reads; the other 812 are destructuring bindings, which retain Unknown checks
without a shared-graph query. Among matched receivers, 2,457 have known
allocation sets and 40 overlap allocations reaching a cast. The graph does not
prove the other receivers non-viewed: unknown cast operands, property loads,
callbacks, generics and diagnosed producers can route known allocations through
untracked paths. All 66,960 inventory receivers therefore retain checks. A known
allocation is not a proof that an object never passes through a view. No second
flow solver or erased guard was introduced. `read-flow-summary.json` preserves
these counters and source hashes; `read-flow-result.json.gz` preserves the raw
query/cast/diagnostic evidence. Known overlap is conservative may-flow evidence,
not a claim that an actual execution has passed a view to that read.

The table classifies only the field's immediate declared type, without expanding
its descendant contract graph. Families overlap. These are **family demand**
counts; backend admission/support has not been measured for each read, so they
must not be relabeled as proven missing-family counts. Lane 2's existing partial
array/callable support does not make every row supported or unsupported.

| Field family | Owner | Distinct (type, field) pairs | Explicit read sites |
| --- | --- | ---: | ---: |
| callable contracts | lane 2 | 2,818 | 11,063 |
| nullish members | lane 1 | 2,011 | 9,084 |
| optional properties | lane 1 | 1,681 | 7,863 |
| tagged object union | lane 1 | 488 | 3,116 |
| dictionary contracts | unassigned | 409 | 2,256 |
| array or tuple contracts | lane 2 | 389 | 2,117 |
| intersection field contract | unassigned | 198 | 1,145 |
| mixed primitive union | lane 4 | 63 | 690 |
| object plus primitive union | lane 4 | 73 | 267 |
| untagged object union | lane 4 | 84 | 186 |
| any field contract | unassigned | 74 | 138 |
| generic field contract | unassigned | 91 | 118 |
| nominal class fields | unassigned | 5 | 52 |
| unknown field contract | unassigned | 3 | 12 |

Deduplicated within each lane (counts overlap across lanes):

| Lane | Distinct pairs | Read sites |
| --- | ---: | ---: |
| lane 1 | 2,370 | 11,321 |
| lane 2 | 3,207 | 13,180 |
| lane 4 | 213 | 1,111 |
| unassigned | 780 | 3,721 |

Top lane 4 declared shapes by actual explicit read count, under the same
Unknown fallback (these are not cast counts):

| Shape | Distinct pairs | Read sites |
| --- | ---: | ---: |
| `__String` | 30 | 543 |
| `string \| number \| undefined` | 4 | 66 |
| `string \| NodeArray<JSDocComment> \| undefined` | 24 | 63 |
| `"string" \| "number" \| "boolean" \| "object" \| Map<string, string \| number> \| "list" \| "listOrElement"` | 1 | 33 |
| `Identifier \| ObjectBindingPattern \| ArrayBindingPattern \| PrivateIdentifier \| StringLiteral \| NumericLiteral \| BigIntLiteral \| ElementAccessExpression \| ComputedPropertyName \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| PropertyAccessEntityNameExpression` | 2 | 26 |

For comparison, the old leading `false | string[] | undefined` and
`false | VersionPaths | undefined` shapes are no longer the leading read demand.
`__String` still requires the intersection/brand family; a scalar tag alone does
not prove its phantom brand. No family is declared completed by this table.

The full pair table includes declared type displays, families, count and a
source witness in `stage3/interface-downcasts/lane4/read-demand-pairs.json.gz`;
the site inventory in `read-demand-sites.json.gz` includes exact UTF-16 spans,
receiver type identity, field, family and Unknown reason. Type displays may be
truncated except lane 4 union shapes; checker type ids distinguish pairs. The
summary and executable enumerator are `read-demand-summary.json` and
`read-demand.cjs`. `audit-read-demand.py` independently verifies source spans,
pair/family totals and `module.valueDeclaration` in the shared helper
`getSourceFileOfModule`, utilities.ts:993. That helper receives both viewed and
ordinary Symbol values and must not be excluded for lacking a lexical cast.

### Contract admission is still eager

Observed in the current lane 4 branch and fetched lane 1 9b85eada:
`view()` invokes `viewObjectFields` across the target graph and then constructs
recursive `viewContract` children. Both propagate unsupported-child failures.
There is also a whole-program scan keyed by field name, rather than reaching
receiver identity. Supported runtime payload checks are lazy, but compile-time
contract admission is not. An unread unsupported descendant can refuse a cast. The
`unread-unsupported.a` control reads only `ready` and has a valid unread
`string | number` value. Source Node prints true; both backend compilers refuse
at its cast with `checked view field unread of type string | number`.
`unread-observations.json` pins both refusals. This is an observation of the
current eager frontier, not a desired lazy-admission test.

Required lane 1 integration: reserve unsupported/deferred child descriptors
without admitting their payloads; cast admission checks the target tag; every
read with known viewed or Unknown reaching provenance checks its own declared
field contract. Unsupported reads must refuse with field name and family.
Use lane 3's shared query and retain tag tests in helpers, instantiated generics,
callbacks and stored-value reads. Do not turn Unknown into an erased guard.
A deferred descriptor alone cannot safely enable admission while the global
field-name scan or unchecked helper reads remain. These shared files remain
lane 1 territory and have not been changed by lane 4.

The helper oracle now passes ordinary and viewed Box values to the same helper.
Source Node prints `true` then `42` for its malformed boolean member; native and
JavaScript print `true` and stop at `value.unsupported` with exit 70, expected
boolean, found number. Removing only that helper read's `Property.View` marker
runs valid release code: native prints `true`/`false` and JavaScript
`true`/`42`, both exit 0. The pinned exit/message catches both mutants.
A helper reading the unsupported `string | number` family still compile-refuses
with the exact read location/type message. This proves current scalar helper
checks and conservative unsupported-family refusal, not lazy union admission.
The separate measurement mutant turning off an Unknown helper receiver's guard
is caught by the independent inventory audit. No lazy mixed-union or lane 1 nullish/optional completion claim follows from
this inventory. A defensible nullish/optional completion date requires lane 1
to supply the lazy dispatch hooks and resolve the shared cast merge conflict;
lane 4 cannot commit a date for another owner’s implementation. Per-cast only-lane-4/plus-other counts are **unmeasured** under
the new rule, rather than inherited from the obsolete transitive census.

Reproduce after sourcing the toolchain:

```sh
NODE_PATH=stage3/fixtures/assertions/api/node_modules node stage3/interface-downcasts/lane4/read-demand.cjs /tmp/views-mixed-typescript stage3/interface-downcasts/lane4 > /tmp/read-demand.log 2>&1
# gzip the generated sites/pairs JSON with deterministic mtime=0.
python3 stage3/interface-downcasts/lane4/audit-read-demand.py /tmp/views-mixed-typescript stage3/interface-downcasts/lane4 > /tmp/read-demand-audit.log 2>&1
```

Shared-graph reproduction uses an isolated worktree at c1f4c5a7, with its exact
cohere/TypeScript submodule pins (worktrees, no copied code):

```sh
# Run this command from that lane 3 worktree.
python3 /workspace/adamic/stage3/interface-downcasts/lane4/make-read-flow-overlay.py /tmp/views-read-flow-overlay > /tmp/read-flow-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/views-read-flow-overlay/overlay.json -o /tmp/views-read-flow-meter ./stage3/shape-conformance/latent/tool > /tmp/read-flow-build.log 2>&1
# Run the remaining commands from lane 4.
python3 stage3/interface-downcasts/lane4/map-read-flow-input.py /tmp/views-mixed-typescript stage3/interface-downcasts/blocking-families-sites.json /tmp/views-read-flow-casts.json > /tmp/read-flow-map.log 2>&1
/tmp/views-read-flow-meter /tmp/views-mixed-typescript /tmp/views-read-flow-casts.json /tmp/views-read-flow-result.json > /tmp/read-flow-run.log 2>&1
python3 stage3/interface-downcasts/lane4/export-read-flow.py /tmp/views-mixed-typescript /tmp/views-read-flow-result.json stage3/interface-downcasts/lane4 > /tmp/read-flow-export.log 2>&1
python3 stage3/interface-downcasts/lane4/read-demand-mutants.py /tmp/views-mixed-typescript > /tmp/read-demand-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewLane4' -count=1 -v -timeout 10m > /tmp/helper-oracle.log 2>&1
```

The scoped helper oracle passes in 3.142s and oracle vet passes. Tests/output go
to committed logs. Production compiler files were not changed. The full gate
was not rerun; the inherited TestSharedArrayContractAdapter disagreement from
the earlier lane 4 report remains outside this measurement change.

## Integration branch checkpoint, October 7, 2026

The dated lane checkpoints above retain their original evidence and limits.
The integrator has merged lazy owner e6aec805 first on codex/views-integration;
its first published SHA is f1c919701ec9b42b78e104fe9cc1ce96db9d7d09.
Lane-4 primitive phantom scalar reads are reconciled with that lazy entry point;
standalone selector evidence does not imply production union admission.
See stage3/interface-downcasts/integration/REPORT.md for conflict choices,
validation, reproduced baseline failures and mutants.

## Continued integration checkpoint, October 8, 2026

Continuation starts at d718a9ffa0432245b69a3545c3ca080db865e0bb on
codex/views-integration. Optional boolean a2eb65ca76816f895f846f78a210bc1717a5f7c4
and untagged unions cd32db4234fb0a65a6ad199474b802c9e7491054 are ancestors.

The attempted owner merge d5b3c4a9bde6ee28fe38568ef4b720aa9c2b68e3 was
aborted with 18 unresolved paths. The previous bd05075f rollback deleted Map
storage adapters, tuple/callable entry descriptors and producer tests that the
new nominal Map code depends on. Its incoming Map forEach calls closure code
directly; integration dispatches through viewCallableBoxedInvokeTypes. Restoring
the deleted adapters also requires reconciling ownership of converted keys and
values with that dispatch. No whole-file conflict selection was used.

This is an unresolved integration boundary, not a new runtime failure or a
certification of the owner fixtures. The existing Union callback adapter and
unknown-producer refusals are retained. Do not restore the general Union calling
convention refusal to hide this boundary: entry-live-mutation.a then refuses
lowering, and the boxed callback fixtures lose their supported behavior. The
native and JavaScript TestViewCallableBoxingUnknownProducer tests and
TestCheckedViewCallableRetainsNeverWiderHelper remain required boundary checks.
See stage3/interface-downcasts/integration/CONTINUED-REPORT.md for the final
merge ledger and validation results.


## Tuple continuation dependency, October 8, 2026

The attempted tuple merge 5c54e8c6583462ae061557d5fc4618deaf4b67ee was
aborted after nine paths were individually reconciled. Nullish Map storage
fixtures restored for measurement all refuse at Map admission; optional/rest
Map tuple reads retain the unsupported Map key/value certificate contract
refusal. They depend on the deferred owner storage/provenance changes above.
A narrow supported-tuple routing retry also found a JavaScript/Node disagreement:
original emit-tuple-good and emit-helper-good exit 70 with "a union value does
not match its narrowed type" in JavaScript while the source Node controls succeed.
Existing storage and callable refusals were kept. Exact logs and the attempted
patch are in stage3/interface-downcasts/integration/logs/continued; the complete
ledger and command outcomes are in CONTINUED-REPORT.md. The tuple item remains
out of the branch; latest green compiler merge is 03aea1cbede84da46e84c78d599f287bc8585c1b.


## Priority integration resolution, October 8, 2026

The earlier owner and tuple attempt blockers above are historical and superseded. Owner 0f55d664 and tuples 0804dec5 both include ffe428ab and are integrated with dictionaries 48d166cf7. Boxed callable dispatch and owned Map snapshots coexist. All 47 Map callback refusal sites remain, including entry-live-mutation.a at 5:69; the general callable lane remains intact. No broad primitive-array selector admission is added.

The tuple lane retains its direct array-to-tuple cast frontier. Dictionary array-entry consumers retain six named [element] tuple refusals until their producer-specific ownership proof is supplied; scalar enumeration and independent certified tuples remain supported. Recursive nominal witnesses and required non-void primitive brands retain existing refusals. Non-null assertions stay outside .a fixtures: four new owner/tuple fixtures use explicit narrowing, and the previous .ts migration remains. Final scoped controls and mutants pass; the global count refresh has 39 inherited failures, no new failures. Detailed hunk decisions and exact logs are in stage3/interface-downcasts/integration/CONTINUED-REPORT.md.
