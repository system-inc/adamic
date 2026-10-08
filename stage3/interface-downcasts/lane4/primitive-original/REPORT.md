Certified the remaining four array pairs / nine reads and three original dictionary scalar/nullish components / fourteen reads.
Commits: dca0b478 for arrays; this dictionary checkpoint on codex/views-mixed-unions-2.
Checks: final original + component + array-safety oracle PASS 107.925s; lower/ir/JS PASS 1.502s / 0.083s / 2.233s; proof tests and vet PASS.
Mutants: arrays eight member, eight first-member and one metadata kills; dictionaries eight member, eight first-member and eight null-erasure backend kills.
Uncovered: original rich dictionary reference alternatives, other lanes' families and exact whole-tsc reachability; no full-gate claim.

| Original pair | Candidate reads | Status |
| --- | ---: | --- |
| EvaluatorResult<string \| number \| undefined>.value | 61 | Certified |
| NodeLinks.isExhaustive | 4 | Certified, including rejection of number 1 where only 0 is allowed |
| EvaluatorResult<string \| number \| undefined> \| undefined, value | 3 | Certified |
| EmitNode \| undefined, constantValue | 1 | Certified |
| EmitNode.constantValue | 1 | Certified original destructuring |

Fixtures import the complete original declarations from pinned TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8. Preparation verifies all 78 declaration
hashes, original source read locations and complete receiver field sets. Required
undefined and optional missing values remain distinct. Each valid member is held
to Node in both backends; positive native runs use sanitizers and leak checks.
Wrong boolean, null, wrong numeric literal and wrong string values have named
refusal pins. Helpers receive viewed values through the shared flow.

Run: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)$'
-count=1 -v. Output is logs/oracle.log. The integrated nullable primitive selector supplies the runtime checks.


The additional destructuring hook is in internal/lower/view_primitive_reads.go,
with minimal named calls from collections.go, interface_cast.go and readiness.go.
It requires a complete declared primitive contract. The helper can lower before
the cast, so bindings intern their descriptor before checking admissibility.
The ordinary binding control exposed removal of a required primitive-to-box
conversion; primitiveBindingConversion retains it even without a viewed value.
The three ordinary members now match Node in both backends and sanitizers.
Object unions and missing/unsupported descriptors retain refusal. Unknown-flow
guards are unchanged. No runtime ABI or emitter change is needed.

The complete original EmitNode graph exposed a separate callable producer-scan
panic on an unrelated binding pattern. viewCallableProducerDeclaration in
view_callables.go restricts name extraction to the four existing producer kinds.
The fixture catches the former panic; callable signature tests pass unchanged.
Both failures before their fixes are retained as diagnostic logs, not green gates.

Current command: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^Test(CheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)|PrimitiveOrdinaryBinding)$'
-count=1 -v. Exit 0, 100.134s, native cache misses 22, Node misses 80, no hits;
logs/bindings-oracle.log. Lower command: go test ./internal/lower -run
'^Test(Primitive(ViewDestructuringAdmission|BindingConversionRequiresCompleteDeclaration)|ViewFallbackScopesRetainWiderHelperGuard|ViewAggregateNullishConstantsAreNotAllocations|ViewUnionTargetAdmission|LazyViewDemandUsesSharedFlow|LazyViewArrayDemand|ViewCallable.*|CheckedViewCallable.*|DestructuredMethodsCannotLoadOwnSlots)$'
-count=1 -v. Exit 0, 0.872s, logs/bindings-lower.log. go vet
./internal/lower ./internal/oracle exits 0. No full repository gate is claimed;
the nine full lower-package baseline failures recorded at 6ed3ced0 are unchanged
and were not rerun for this focused checkpoint.

Each wrong-member read bypass executes valid release code in C and JS and
violates the named exit-70 pin. Each first-member substitution replaces the
undefined result with the first present primitive (string or numeric zero),
executes valid release code and disagrees with the original Node control.
The five groups contribute ten kills of each mutation class. This family has
no object member whose transitive check could be independently dropped.

Original mixed-primitive column now has 39 / 621 certified and 24 / 69 remaining.
One remaining intersection pair / one read belongs to lane 7. Nonbrand remainder
is 23 / 68, using candidate counts. __String is delivered early; mixed primitives
retain October 13, 2026, 17:00 MDT as the estimate. Whole-tsc exact reachability
still depends on a checker-clean program. Continue most-read primitive groups;
there is no overnight time stop.


Pure primitive follow-up: StringLiteralType | NumberLiteralType.value is now
certified, six candidate reads. The helper retains the exact original union
receiver type; each fixture feeds it a viewed value of one complete original
member. No narrowed receiver declaration is substituted. Both full member field
sets and the stock original union read witness are checked. This does not certify
a direct cast to the untagged object union, which is lane 4c's obligation.

New preparePrimitivePropertyRead hook plus minimal object.go/interface_cast.go
admission and readiness conversion retention reuse the existing runtime selector
with null/undefined alternatives disabled. Ordinary pure-union field reads match
Node and sanitizer/leak checks too. One now-supported NotYet field-read row is
removed; writes, class fields, arrays and closure storage refusals remain tested.
Wrong boolean, null, undefined and missing fields have exact named exit-70 pins.
The new bypass and untested-first-member mutants execute valid release code in
both backends and are caught: four more backend kills, 24 in the six-group suite.

Uncached six-group + ordinary binding/property oracle PASS 132.186s, exit 0,
native misses 25 and Node misses 96, no hits; logs/pure-oracle.log. Filtered lower
refusal, primitive, callable and lazy-flow tests PASS 1.698s. Vet passes. The full
repository gate is not claimed. Counts now: 40 / 627 certified, 23 / 63 remaining
in the original mixed-primitive column; one remaining pair/read is lane 7.
Nonbrand remaining is 22 / 62. Larger object-plus-primitive entries remain lane
4b and are not credited here. Continuing with three-read nullable primitive
fields; October 13, 17:00 MDT remains the estimate.


Nullable follow-up adds six pairs / thirteen reads:

| Pair | Candidate reads |
| --- | ---: |
| Diagnostic.skippedOn | 3 |
| IncrementalBundleEmitBuildInfo.pendingEmit | 3 |
| Resolved.originalPath | 3 |
| PackageJsonInfoContents.peerDependencies | 2 |
| ReusableDiagnosticRelatedInformation.file | 1 |
| ReusableDiagnostic.skippedOn | 1 |

Original interfaces and aliases are imported whole from the hash-pinned 78-file
emission. Private Resolved is printed from its complete original interface with
all five fields and its PackageId dependency; its generated declaration and
source are hash-pinned separately. The stock read witness and all receiver field
sets are verified. No proxy receiver or shortened private interface is used.

The true-only member of originalPath caught a JavaScript diagnostic mismatch.
The minimal nullishMemberSelection emitter hook now uses native's named union
message. Both backends share exact pins for wrong runtime kinds and wrong finite
boolean members. Missing required pendingEmit, peerDependencies and file fields
have separate named initialization pins. Missing optional fields return undefined.

The precise skip-member mutants retain the original field read and physical
producer storage validation, remove declared masks/literal/member obligations,
and print the original wrong value in valid release C and JS. Only the member
refusal can catch these mutants; there is no constant-output proxy in this rerun.
The bundle first-member substitution uses its actual numeric member None (0).

Run: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^Test(CheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)|PrimitiveOrdinary(Binding|Property)|CheckedViewNullish.*)$'
-count=1 -v -timeout 10m. PASS 299.401s, native misses 78, Node misses 372,
zero hits; logs/nullable-all-oracle.log. Exact-member rerun uses
'^TestCheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)$/[^/]+/(wrong|wrong-number|undefined)$'
and passes in 61.710s, native misses 11, Node misses 80, zero hits. Combined
with the pure-union numeric control, this proves 30 exact member-check and 24
first-member backend kills. JavaScript package passes in 1.563s; vet passes.
No full lower package or repository gate is claimed.

The original mixed-primitive column now has 46 / 640 certified and 17 / 50
remaining, including lane 7's branded intersection 1 / 1. Nonbrand remaining is
16 / 49. These are candidate counts; object-plus-primitive rows remain owned by
lane 4b and dictionary rows overlap lane 6. Continue through the remaining
private receivers; October 13, 17:00 MDT remains the estimate.

Private receiver batch certifies FlowGraphNode.circular (one read) and the
FilePresentOnHost | FilePresenceUnknownOnHost.version helper (one read).
Complete private interfaces and dependencies are printed from pristine original
source, hash-pinned, and field sets checked. Both backends match Node for every
member, including circularity versus boolean and string versus false. Named
wrong-kind, wrong-literal, missing and null refusals are pinned. Actual-value
member-check removal and first-member mutants fail in both backends.

Builder signature has two candidate reads but reaches an unsupported compound
intersection receiver at child. Seven Node controls and named compiler refusals
are pinned; no certification credit, lane 7 owns that receiver.

Stock checker inspection corrects two tuple positions / four candidate reads:
[number, string][0] is number and [1] is string. The old whole-index union
classification was a false positive. See tuple-candidate-corrections.json and
measure-tuple-candidates.cjs; this reclassification earns no certification credit.
Lane 1 should reconcile this correction into its census.

| Candidate scope | Pairs | Reads |
| --- | ---: | ---: |
| Raw original column certified | 48 | 642 |
| Raw original column remaining | 15 | 48 |
| Corrected column remaining | 13 | 44 |
| Primitive array indexes, owned | 4 | 9 |
| Rich dictionaries, shared selector component | 3 | 14 |
| Object alternatives, lane 4b | 4 | 18 |
| Intersection receivers, lane 7 | 2 | 3 |

The dictionary handoff subset CompilerOptions and BuildOptions is two pairs /
eleven reads within the three / fourteen above. Exact whole-tsc reachability
still waits for a checker-clean program. __String is delivered; October 13,
17:00 MDT remains the mixed-primitive estimate, including shared components.

Integration d718a9ff reconciliation: six conflict files resolved hunk by hunk.
Primitive conversion, dictionary admission, untagged structural preparation and
callable guards are retained together. Unknown fallback remains unchanged.
Focused lower/javascript tests PASS 0.705s / 1.407s; original literal, graph and
watch oracle PASS 143.190s (native misses 7, Node misses 50, zero hits); vet PASS.
Broader NotYet table fails two expectations for primitive-union callable
parameters now accepted by integration. Those failures are retained in
logs/integration-refusal-table.log; no full gate claim and no pair count change.

Read-only primitive array batch certifies four original pairs / nine candidate
reads: (string | number)[] positions 0, 1 and 2 (4/3/1 reads), and
DiagnosticArguments dynamic indexing (1 read). Stock-checker source witnesses and
original declared element types are pinned in array-candidates.json. Both valid
members match Node in both backends; wrong boolean and missing reads have exact
named exit-70 pins. Eight actual-value member-admission and eight first-member
backend mutants are caught. A native forged producer-metadata mutant prints
the actual 42 and is caught by the named unknown-storage refusal.

Storage, producer ABI and writes are unchanged. The indexed-read adapter retains
the receiver before index evaluation, validates physical producer metadata,
snapshots one slot, selects a complete primitive descriptor, then owns its box.
The same adapter serves typeof. Ordinary and viewed values pass through the
same helper. Alias updates, callback replacement/evaluation once, sparse holes,
bounds, finite literals and null refusals have controls. Sanitized runtime
controls release arrays before using retained string and scalar boxes.

Clean pair oracle PASS 17.738s (native misses 8, Node misses 40, no hits); safety
oracle PASS 13.940s (native misses 6, Node misses 20); metadata mutant/storage
test PASS 1.144s. Focused native/lower/ir/JS packages PASS 91.391s / 38.965s /
0.108s / 3.343s, vet PASS. Earlier safety harness failures are retained: console
typing corrected, and sparse typeof controls preserve the unrelated holey
UnionToString refusal. No full repository gate claim. Cumulative primitive
selector mutants: 50 actual-value member backend kills, 36 first-member backend
kills, plus one native storage-certificate kill.

Corrected candidate column: 52 pairs / 651 reads certified; 9 / 35 remaining.
Raw column: 11 / 39 remaining, including two scalar tuple false positives.
Owned remaining: three shared rich dictionary selector components / fourteen
reads. Dictionary null hooks are in progress and receive no credit in this batch.

Original dictionary scalar/nullish batch:

| Original receiver | Candidate reads | Certified component |
| --- | ---: | --- |
| CompilerOptions | 10 | string, number, boolean, null, undefined and missing lookup |
| OptionsBase | 3 | same scalar/nullish selection |
| BuildOptions | 1 | same scalar/nullish selection |

Stock source witnesses, original index types and whole receiver field sets are
recorded in dictionary-candidates.json. Fixtures import full hash-pinned original
modules, verify all original receiver fields, and assert the original unsupported
reference descriptor remains present. Helpers retain the exact original types.
Ordinary boxed records and viewed fixed objects use the same generic helper.

Read admission previously consulted the whole nullable descriptor, so unread
dictionary objects and multiple array alternatives poisoned a selected scalar.
Named dictionaryPrimitiveReadContract creates a distinct complete primitive
certificate. It preserves the original full contract and type ID. The get/read
dispatch uses that certificate only after PrimitiveDictionaryReadCertificate
proves all admitted kinds scalar/nullish. Runtime source tags and storage are
checked at every read, including unknown flow. Global unknown fallback and
unsupported descriptors remain intact. Reference values are excluded and stop
with named exit-70 pins until their owners implement those alternatives.

Null is classified and boxed with the existing adamic_null sentinel rather than
undefined. Stringification is admitted only after a complete scalar/nullish
dictionary selection, including record get. JS uses its existing closure type
tag with a fallback for standalone kernel probes. No lookup, absence, enumeration,
storage layout or ownership ABI is replaced.

Every valid primitive member matches Node in both backends with sanitizers and
leak checks. Wrong function values have exact named refusal pins. Each original
receiver has actual-value member-admission, first-member and null-erasure mutants;
all execute valid release code and are caught. Generic selector controls repeat
these mutations. Existing dictionary skip-check, wrong-shape and transitive
contract mutants also pass their regression gate. Reference-arm refusal controls
earn no reference certification credit.

Final command: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedView(OriginalDictionaryPrimitiveComponents|PrimitiveDictionarySelector|DictionaryComponents|DictionaryComponentMutants|PrimitiveArraySafety)$'
-count=1 -v -timeout 10m. PASS 107.925s; native misses 31, Node misses 115,
zero hits. Proof tests PASS lower 0.687s / ir 0.016s. Final focused lower/ir/JS
PASS 1.502s / 0.083s / 2.233s; vet PASS. Earlier diagnostic and conversion seams
are retained in logs; none are counted as green.

Owned candidate work now has zero pairs / zero reads remaining: all seven /
twenty-three components are delivered ahead of October 13. Full-union column
credit remains 52 / 651, with corrected 9 / 35 remaining. That remaining scope
contains the three dictionary reference components / fourteen reads, lane 4b
object alternatives four / eighteen, and lane 7 intersections two / three.
Raw remaining is eleven / thirty-nine, including two tuple scalar false positives.
Exact whole-tsc reachability still requires a checker-clean program.

Remaining-pair reconciliation: LiteralType.value (one pair / ten candidate
reads) reuses integrated lane 4b production and complete original witnesses.
Uncached TestCheckedViewObjectPrimitiveOriginalPairs/literal- passed in
14.752s with original declarations enabled. Four cases passed Node, native
ASan/UBSan, release C and JavaScript; eight independent backend mutant kills
cover wrong member, skipped union check, wrong nested shape and transitive
omission. No new production code needed for this already integrated pair.
Corrected column now 53 / 661 certified, 8 / 25 remaining. Lane 6 owns
three dictionary reference pairs / fourteen reads; this worker has five /
eleven remaining. Counts are static candidates, not exact reachability.

CommandLineOption.defaultValueDescription checkpoint: one original pair /
four candidate reads green. Full original member interfaces pass into the
CommandLineOption helper; string, number, both boolean descriptors, undefined
and DiagnosticMessage selection retain the original merged declaration.
All seven cases and six backend mutant kills passed uncached in 17.136s:
ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewOriginalCommandDefault$' -count=1 -v -timeout 10m.
Wrong null and nested code pins stop at named reads with exit 70. Positive
native cases run ASan/UBSan. Independent stock-checker measurement verifies
complete member and DiagnosticMessage field sets; all original declarations
remain hash-verified. Mutants skip the actual selector, take an untested first
member, and omit the transitive code check; each exits normally and fails its
pin. Earlier omission drafts encountered a later narrowing guard and were
not credited; the final typeof poison isolates the actual selector.
Minimal shared hooks are listed in the plan. Scoped source checks and the
global unknown fallback are unchanged. Lower/IR selected certificate and
intersection/lazy tests passed; vet passed. Full gate is not claimed.
Column now 54 / 665 certified, 7 / 21 remaining; exclude lane 6 three /
fourteen from owned work, leaving four / seven. Counts remain candidates.

Builder signature checkpoint: IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo
.signature (one pair / two candidate reads) now certified. The cause was the
intersection visitor treating a complete primitive-union descendant as a
compound unsupported payload before its field read. The minimal named hook
defers only complete primitive descendants; unsupported, never, unknown and
object alternatives retain the guard. Fixtures retain full original Omit
intersection declarations and populate all required fields. Seven cases
pass Node, release C, JavaScript, native positive sanitizers and four semantic
backend mutant kills (member omission and first-member substitution).
Standalone uncached TestCheckedViewOriginalBuilderSignature passed 16.392s;
negative certificate/intersection/lazy package checks and vet passed.
Full repository gate not claimed. Property certificate ledger is now 15 / 93.
Combined column now 55 / 667 certified, 6 / 19 remaining. Lane 6 dictionary
reference exclusion is unchanged at three / fourteen; owned remainder three /
five, including the measured tuple-position false positive pending correction.
