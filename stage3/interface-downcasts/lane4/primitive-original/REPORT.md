Certified fourteen original primitive pairs, 91 candidate reads, including private graph and watch receivers.
Commit: this checkpoint on codex/views-mixed-unions-2, after a4df9433.
Checks: private receiver oracle PASS 47.782s; extra exact-member mutants PASS 14.800s; vet passed; prior twelve-group suite PASS 299.401s.
Mutants: cumulative 42 backend kills of actual-value member-check removal and 28 of untested first-member substitution.
Uncovered: intersection receivers, primitive arrays and shared dictionary alternatives; candidate counts only.

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
