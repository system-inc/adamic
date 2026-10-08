Built: merged integration ba59427 and replaced the old queue with the lazy candidate inventory.
Commits: integration ba59427; this measurement checkpoint on codex/views-mixed-unions.
Checks: supplied gzip inventory parses; exact family counts assert 14/511 and 23/538.
Mutants: no compiler check is claimed by this measurement checkpoint; read fixtures are in progress.
Uncovered: full __String integration, mixed primitive source selection and production tsc reachability.

The integration REPORT.md was read, including each hunk choice and inherited gate
limits. Its tip contains our 9a385606 and merged without conflict. Only the
integration branch is consumed for other-lane dependencies.

The queue now comes from lazy/census/read-demand-pairs.json.gz, with its exact
compressed source hash preserved in lazy-pair-progress.json. This supersedes the
old lane-4 213/1111 ledger for current scheduling. The supplied lazy REPORT calls
these static candidate reads and explicitly leaves production allocation-flow
reachability unmeasured. The table must not be presented as certified runtime
reachability. Whole-tsc checker errors do not prevent individual source fixtures.

| Family | Remaining pairs | Remaining candidate reads |
| --- | ---: | ---: |
| __String | 14 | 511 |
| Mixed primitive union | 23 | 538 |

Columns overlap: all fourteen branded pairs occur in the mixed-primitive family.
The leading Identifier.escapedText and Symbol.escapedName pairs account for
224 + 223 = 447 candidate reads in this inventory. Completion still requires
source lowering on both backends, Node controls, wrong-value pins and semantic
mutants; the previously pushed plain intersection fixtures remove no full union.

Revised working targets: __String October 8, 2026, 17:00 MDT (23:00 UTC), and
both families October 10, 2026, 17:00 MDT (23:00 UTC). The remaining intersection
receiver and callable/array consumer entries may need other-lane hooks; their
counts will remain pending until source fixtures establish completion. No date
is contingent on making the whole tsc program checker-clean. Report actual
integration blockers rather than silently removing those entries.

The in-progress implementation uses the existing registry's Undefined bit to
separate a required string field containing undefined from optional absence.
It uses the shared readiness/optional slot reader, with no second flow graph or
readiness bitmap. Source evidence and revised remaining totals follow in the
first green compiler group; no in-progress code is included in this table push.

## First source group: required branded strings

Built: complete __String scalar reads with branded-void and internal-name members.
Commits: based on integration ba59427 and measurement 8538b03c; this source-group commit.
Checks: source oracle 5.969s; expanded lower 2.228s; uncached checked-view oracle 69.389s; scoped vet passes.
Mutants: both backend undefined-admission overlays fail semantically; read-removal and first-member IR substitutions are caught in both release backends.
Uncovered: remaining twelve branded pairs, general mixed primitive selectors and whole-tsc production reachability.

Identifier.escapedText (224 reads) and Symbol.escapedName (223 reads) are now
held end to end to source Node. The source preserves __String's string brand,
branded-void arm and InternalSymbolName string alternatives. Node controls cover
string, undefined and an internal name; malformed number/object/null payloads
stop at the helper's field read, naming __String and the found category. A missing
required field also stops there, rather than becoming an accepted undefined.
Positives run native release, ASan/UBSan with leak checking, and JavaScript.
The Symbol fixture renames the reduced interface ViewSymbol to avoid a global
library Symbol declaration; its declared member contract and field are preserved.

The runtime uses shared presence/readiness machinery. Required undefined payload
admission uses the registry's Undefined bit, without granting optional absence.
JavaScript tagged-object selection is called only for object-valued reads, so a
scalar union cannot accidentally require an object discriminant. Only phantomBase
approved brands are erased; branded void is an undefined contract, not an object.
Existing literal-brand refusal pins remain in the broader checked-view oracle.

The undefined-admission overlay mutants each remove just that backend's declared
undefined-member admission. Valid release executions stop instead of matching
Node's undefined output, and the fixture catches both. The IR read-removal
mutants replace one helper read with a dynamically built valid string, producing
uncheckedunchecked at exit 0 instead of the pinned refusal. First-member
substitution produces wordword for an undefined value and is caught by the Node
control. Earlier constant-string versions hit clang's tautological-address
warning; those attempts were discarded and are not counted as mutant kills.
Object-member transitive selection is outside this lane's narrowed scope.

| Family | Completed candidate pairs / reads | Remaining candidate pairs / reads |
| --- | ---: | ---: |
| __String | 2 / 447 | 12 / 64 |
| Mixed primitive union | 2 / 447 | 21 / 91 |

Columns overlap. These are supported per-pair fixtures against the supplied
candidate inventory, not successfully compiled whole-tsc reads. No checker
error is bypassed. The first scoped oracle attempt failed only because pinned
@types/node was absent; npm ci --prefix stage3/api installed the declared version,
and the uncached rerun passes. Full repository gate is not claimed.

Toolchain setup succeeded: Go 0.046s, Node 0.049s, Markdown 0.137s, clang 0.295s,
submodules 0.873s, build 436.959s, cache 437.249s, total 437.304s; nproc 5, quota 4.

With shared hooks now explicitly authorized, working targets remain October 8,
2026, 17:00 MDT for the remaining branded family and October 10, 17:00 MDT for
both families. These dates supersede the earlier handoff-dependent estimate;
remaining fixture results, not another lane's permission, determine delivery.

## Branded family source coverage complete

Built: all fourteen branded candidate pairs, including optional fields and receiver unions/intersections.
Commits: implementation 2b767776; this per-pair fixture group carries the remaining twelve pairs.
Checks: TestCheckedViewBrandCandidatePairs passes 26.446s across 60 source cases; scoped oracle vet passes.
Mutants: each of twelve malformed-number helper reads has a one-read bypass caught independently by valid native release and JavaScript runs.
Uncovered: nine remaining mixed-family candidate pairs / 27 reads; whole-tsc compilation and runtime reachability.

The twelve remaining pair fixtures preserve their candidate receiver forms,
field names and complete __String declaration. They cover string, branded void,
wrong number, wrong null and missing own fields. The three optional fields allow
absence; required fields refuse it. Helpers with Identifier|undefined and
Symbol|undefined guard their receiver, while the Identifier|PrivateIdentifier,
MemberName and LeftHandSideExpression&Identifier helpers retain their static
receiver type. The last is a read through an intersection receiver; this does not
claim general intersection cast admission or lane 7's conjunction machinery.
Module-local reduced interfaces avoid library declaration merging.

Each positive source case matches Node in release native, sanitized native with
leak checks, and JavaScript. Each wrong case pins field, declared __String union,
found category and exit 70 in both backends. The malformed-number read-removal
mutants execute valid release code and print uncheckedunchecked at exit 0,
violating the refusal pin. These are per-pair checks, not a claim that compiling
a reduced fixture compiles its entire tsc source module.

| Family | Completed candidate pairs / reads | Remaining candidate pairs / reads |
| --- | ---: | ---: |
| __String | 14 / 511 | 0 / 0 |
| Mixed primitive union | 14 / 511 | 9 / 27 |

The branded candidate family is delivered on October 8 UTC (October 7 MDT),
earlier than October 9. The remaining mixed-family working target is October 9,
2026, 17:00 MDT (23:00 UTC), with the array consumer's overlapping object/member
contract reported separately if it needs another lane's source adapter. Shared
hooks are now in this lane's hands; no wait on permission or whole-tsc checker
cleanup is included in that estimate.

No production code changed in this fixture-only group. The prior uncached
checked-view oracle, scoped lowering and backend undefined-admission overlay
mutants cover the unchanged implementation; the new pair oracle and vet cover
this group's additions. Test outputs are preserved in logs/brand-candidate-pairs*.
