# Checked-view blocking families and remaining lane dependencies

Measured before the next family implementation on the pinned 2,936 sites. These
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
