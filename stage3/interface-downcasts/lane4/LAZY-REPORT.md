Built: accepted heap/slot probes and primitive contract planning; corrected the finite-key overclaim.
Commits: __String 4ba489e5; finite-key component 0bae12f8; this probe/correction checkpoint.
Checks: all checked-view oracles pass 80.374s; scoped vet passes; IR 32.511s and JavaScript 2.307s pass; five native graph failures reproduce at 0bae12f8.
Mutants: seven probe defects caught by semantic release output or exit-70 pins; detailed list below.
Uncovered: nine mixed primitive candidate pairs / twenty-seven reads, source boxed selection, full tsc compilation and exact runtime reachability.

This current summary and the correction below supersede historical tables,
including the finite-key group's incorrect completion claim. Earlier __String
source evidence remains valid. The following chronological checkpoints preserve
the observed work and prior estimates rather than rewriting their history.

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

## Finite string-key unions

Correction: this checkpoint incorrectly counted the two tsc pairs as complete.
Real CompilerOptions has a string index signature; the correction below
supersedes its completion claim and remaining totals. The fixtures are finite-key
component coverage only.

Built: Diagnostic.skippedOn and ReusableDiagnostic.skippedOn, two candidate pairs / four reads.
Commits: builds on 4ba489e5; this source fixture group.
Checks: TestCheckedViewPrimitiveKeyPairs passes 6.307s, with Node controls, native release, JavaScript and sanitized native positives.
Mutants: bypassing either helper literal-member read runs valid release code and violates the wrong-value pin in both backends.
Uncovered: seven mixed-family candidate pairs / 23 reads; production tsc reachability remains unmeasured.

Each reduced CompilerOptions preserves a finite keyof union with undefined.
Fixtures read both string members, explicit undefined and optional absence.
An unknown string, number and null each refuse at the helper read with the field,
declared union and found value/category pinned. No compiler changes were needed;
this group exercises the implemented finite-string and undefined-member path.

| Family | Remaining candidate pairs | Remaining candidate reads |
| --- | ---: | ---: |
| __String | 0 | 0 |
| Mixed primitive union | 7 | 23 |

The mixed-family working target remains October 9, 2026, 17:00 MDT (23:00 UTC).
Automatic approval review rejected the proposed broad boxed-union rewrite across
lowering, slot probing and emitters because its single component test did not
establish freedom from silent miscompiles or memory-safety failures. None of that
rewrite ran. These independent fixtures complete unaffected work; a narrower
runtime probe with dedicated tests is the next step before enabling dispatch.

## Corrected inventory and accepted probe foundation

The prior finite-key group incorrectly completed 9761:skippedOn (three reads)
and 97180:skippedOn (one read). Stock TypeScript types.ts:7571 declares a string
index signature on CompilerOptions. Its keyof admits strings and numbers, not
just named property literals. The reduced finite-key fixture omitted that
signature, making its unknown-string refusal and number refusal inappropriate
for the real pair. Those fixtures are now explicitly named component coverage;
their two pairs and four reads are restored to pending. No history is rewritten.

The corrected gap witnesses preserve the full string/number/undefined shape via
an interface extending the library's Record<string,...>. A numericKey declaration
of type keyof CompilerOptions is checker-clean. This avoids introducing runtime
dictionary storage while retaining the inherited index signature's key domain.
Source Node accepts notAnOption, 42, undefined and optional absence; malformed
boolean and null show their actual values. Adamic currently refuses the mixed
field at compilation. These gap witnesses complete no pair.

| Family | Completed candidate pairs / reads | Remaining candidate pairs / reads |
| --- | ---: | ---: |
| __String | 14 / 511 | 0 / 0 |
| Mixed primitive union | 14 / 511 | 9 / 27 |

| Remaining pair | Candidate reads |
| --- | ---: |
| 6849:<element> | 8 |
| 46428:value | 6 |
| 9761:skippedOn | 3 |
| 97934:pendingEmit | 3 |
| 37515:peerDependencies | 2 |
| 97892:signature | 2 |
| 6995:constantValue | 1 |
| 97180:skippedOn | 1 |
| 97923:forEach | 1 |

The forEach candidate carries array/object/intersection consumer obligations in
addition to its mixed primitive dependency. This lane does not claim that a
primitive selector completes those other families. No production tsc allocation
reachability is measured, and no second whole-program flow solver is introduced.

Accepted source-independent foundation:

* adamic_view_union_heap classifies real borrowed heap values and unwraps scalar
  boxes. Unknown kinds stay unknown; null storage is handled separately.
* adamic_object_view_union_snapshot calls the shared readiness/static-owner
  resolver, retains its actual owner's storage evidence, and distinguishes null,
  undefined, optional absence and present uninitialized slots. It neither invokes
  getters nor enables a cast or read dispatch. Runtime probes cover plain/boxed/
  packed scalar storage, mismatched reference tags, unknown storage and inheritance.
* PrimitiveViewMembers admits only complete primitive alternatives and preserves
  finite literals and undefined. Unsupported/object/missing/recursive graphs do
  not acquire a primitive certificate. Its unit test passes in the IR package.

Seven semantic native release mutations were run independently with production
restored in finally: number-tag, boolean-payload, skip-readiness, ignore-owner,
collapse-null, ignore-reference-storage and trust-unknown. The first two are
caught by TestCheckedViewPrimitiveHeap; the remaining five by
TestCheckedViewPrimitiveProbe. The readiness/unknown/null/reference mutations
run valid code to exit 0 instead of the pinned exit 70; the owner mutation refuses
a valid inherited numeric read. No clang warning or sanitizer-only failure is
counted. Initially the number-tag mutant escaped because printing the original
payload hid the selected member; the harness now observes the selected member's
kind against source Node, and the rerun catches it. That escaped attempt is not
reported as a kill.

Full native package validation fails five graph tests. Exactly those five fail
at the previously pushed 0bae12f8 in an isolated detached worktree as well:
GraphRegionsMillion, GraphContainerBoundary, GraphClosureEnvironment,
GraphLazyRegions and GraphRegionsRuntime. Their logs are preserved; this is an
observed baseline comparison, not a claim that the full repository gate passes.
The complete checked-view oracle and scoped vet results are recorded after the
final run. IR and JavaScript packages pass. No unverified source adapter is enabled.

Automatic approval review rejected both a broad boxed-field rewrite and the
subsequent proposed array adapter, citing memory-safety and silent-miscompile risk
and stating that cross-layer work exceeded the accepted probe-only scope. Neither
rejected edit ran. The concrete adapter is preserved for review only, with a
successful git apply --check and its source/runtime validation gate in
ADAPTER-REVIEW.md. Admission and shared dispatcher hooks remain unapplied.

The __String family delivered October 8 UTC (October 7 MDT), before October 9.
For the eight direct primitive pairs / twenty-six reads, October 9, 2026 at
17:00 MDT (23:00 UTC) remains a working target only if adapter approval arrives
October 8. The overlapping forEach candidate depends on integrated object/
intersection consumer adapters; there is no honest unconditional date for all
nine pairs while those dependencies and automatic approval remain unresolved.
Whole-tsc checker cleanup is not a per-pair fixture dependency.

Final accepted-state commands:

* go test ./internal/oracle -run '^TestCheckedView' -count=1 -timeout 15m:
  PASS 80.374s, including the corrected open-key and primitive-array gap tests.
* go test ./internal/ir ./internal/native ./internal/javascript -count=1 -timeout 15m:
  IR PASS 32.511s, JavaScript PASS 2.307s; native FAIL 462.555s on the five
  graph tests listed above. The same five fail at 0bae12f8 in 11.859s.
* go vet ./internal/ir ./internal/native ./internal/javascript ./internal/oracle:
  PASS with no output. git diff --check and review patch applicability pass.

A fresh integration-tip read failed because GitHub HTTPS authentication is no
longer configured: fatal: could not read Username for 'https://github.com':
No such device or address. The environment reports no configured secret or
outbound identity, and Git has no credential helper. This is separate from the
automatic code-review rejection. The last merged integration tip remains
ba59427. Push is retried after this commit; if authentication remains absent,
the commit is exported as a format-patch for the user/integrator.
