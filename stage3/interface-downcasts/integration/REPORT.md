Integrated lazy admission first from the lane-1 base ab4d6f902.
Owner tip e6aec805 includes lane-2 merge 592f1f71 and shared-flow merge c01ae313.
Focused package tests and vet pass; filtered Node comparison excludes one reproduced baseline failure.
Nested-loop rollback mutant is caught by string_views_policy.a: Node exits 0, mutant native exits 70.
Full repository gate and production tsc census were not rerun; baseline failures remain listed below.

Merge order follows the user's owner-first ruling. This first merge is clean,
with no conflicting hunks and no bulk or whole-file resolutions.

The integration repair in internal/lower/readiness.go applies the existing
markProgramViewArrayUse policy inside recursive struct traversal, so nested
for-of statements receive the same policy as top-level statements. Ordinary
nested loops previously retained preliminary view metadata despite no admitted
array view. The filtered Node fixtures library_array_copy_within.a,
string_views_methods.a and string_views_policy.a prove the repair. The rollback
Go overlay removes exactly that traversal repair; string_views_policy.a then
executes valid C and exits 70 with uncertified storage, against Node exit 0.
The owner's demanded callable-read refusal and tests are retained unchanged.

Validation commands (each writes full output to its named log):

```
go test ./internal/lower ./internal/ir ./internal/native -run 'TestView|TestLazyView|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View|Test.*Phantom|TestNativeAgreesWithNode/internal/oracle/testdata/.*(view|union|brand|optional|array)' -skip 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_holes_callbacks.a' -count=1 -timeout 30m
go vet ./...
```

The unskipped filtered oracle was run first: library_array_holes_callbacks.a
refuses at line 12:37 with adamic/no-type-predicate for the inline predicate
(value) => value !== undefined. The same refusal is reproduced in detached
c01ae313. This baseline fixture is explicitly excluded from the green rerun.
A broader package filter also reproduces TestPhantomArrayCastsAreErased,
TestPhantomArrayRequiredCastsAreErased and TestPhantomArrayProofs/cycle failures
on c01ae313. The parent also fails the tuple_representation refusal wording pin.
These failures were not suppressed by compiler or test changes.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Node ready 0.072s, Go 0.098s,
clang 0.504s, Markdown 0.948s, submodules 662.275s, Go build 1177.123s,
setup done 1177.419s; nproc=5 (CPU quota 4). Stage3 API pinned Node declarations
were installed with npm ci --prefix stage3/api. Early gate attempts before
submodule readiness failed on missing cohere/TypeScript/tsc/go.mod; subsequent
concurrent cold gate attempts were interrupted and not counted as evidence.

Existing carried evidence has whitespace errors; historical logs and raw
fixture evidence are preserved. No cohere code was copied.

## Lane 4 reconciliation

Merged tip 9a385606. Eleven conflict hunks across eight files were reviewed
individually. No whole-file selection or merge strategy was used.

| File | Hunk choice and evidence |
| --- | --- |
| docs/checked-views-blockers.md | Keep both distinct dated demand inventories and their limits. |
| docs/checked-views-plan.md | Keep lazy/callable handoffs and lane-4 scope rulings; latest primitive-brand ruling supersedes earlier runtime-brand requirement. |
| internal/lower/expression.go | Keep owner resolved-result path; retain shorthand symbol resolution and lane-4 overloaded-value refusal. Phantom overload and owner result-stop fixtures prove both boundaries. |
| internal/lower/functions.go | Keep owner implementation resolution and generic overload proof; incorporate approved phantom-result exception in shared census proof. |
| internal/lower/modules.go | Keep owner bodyless-header skip and registration-time overload proofs, including unused-result refusals. |
| internal/lower/interface_cast.go | Both read hunks retain owner optional/nullish/accessor policies and named lazy refusals while using receiver-aware primitive-brand recognition. Literal hunk retains both phantom-base recursion and open numeric enums. |
| internal/lower/view_contracts.go | Retain undefined, nominal, recursive and lazy metadata; classify only approved phantom bases as scalar and retain their finite literals. |
| internal/lower/view_objects.go | Keep owner MaybeNumber/MaybeBoolean support and optional write registration, plus lane-4 brand-aware data predicate. |

The shared overload proof keeps generic alpha-renaming, strict parameters,
nullable result checks and predicate result handling. Only the existing
phantomOverloadResult proof admits a phantom result difference; nongeneric
invalid results use lane-4 refusal diagnostics. Incoming argument fitting is
applied before the owner's resolved-result conversion, so undefined checks are
not bypassed by premature fitting. One parameter test now pins the owner's
more specific diagnostic, keeping the same unsafe mutable-parameter source.
The acceptance-all-results Go overlay mutant is held to the branded-literal
refusal pin. This does not introduce a second flow graph or callable convention.

Brand-string-good/wrong and brand-string-literal-good/wrong match positive Node
controls and retain both backend exit-70 checks on malformed payloads. Their
valid-release read-removal mutants run as part of the checked-view oracle.
General intersections remain deferred; only phantomBase-approved primitive
intersections bypass that unsupported-family classification. Standalone mixed
selectors remain components; complete source union admission is not claimed.

A complete IR-package run exposed the inherited emitter-name allowlist mismatch
and the two overload readers. Both overload helpers now use Program.CallTargets;
the two permanent emitter allowlist keys name evaluateWithoutViewArrays, matching
lane 2's rename without broadening the allowlist. The complete IR rerun passes.
Focused lower/native, full IR, the filtered oracle with the same baseline skip,
and vet are the lane-4 gates. Owner overload result-stop/append controls are
included beyond the view filter. The all-results covariance mutant is caught
with a valid build: TestPhantomOverloadBrandLiteralConstraintRefused sees nil
instead of the required branded-literal refusal.

## Lane 5 reconciliation

Tip 9c4d0904 includes newer native-array lane 2 bf544d5c through 1e47e792.
Three conflict hunks across two files were individually reconciled:

- docs/checked-views-plan.md, two hunks: retain lazy/lane-4 checkpoints and both
  callable shape/read-producer handoffs as dated evidence.
- internal/native/emit_expressions.go, ArraySlice hunk: use lane 2's sparse-aware
  adamic_view_array_slice when array views exist, retaining graphArray and
  GraphTypes from the owner. Automatically merged ArrayPush keeps graph-held
  reference ownership plus sparse-aware push dispatch.

The newer array runtime consumes element_kind, certifying physical allocation
storage without a second view storage byte. Checked sparse pop/copy/join and
write-kind refusals are retained. Fixtures native-array-sparse,
native-array-sparse-pop-hole, native-array-copy-bad, native-array-write-bad,
native-array-push-bad and native-array-join-literal pin these paths, alongside
native-array-string-mutation and large-field/evaluation controls.

Fixed callable signature/read/producer adapters remain components where shared
source wiring is deferred. No unsupported callable refusal, Unknown frontier,
receiver convention or void signature was enabled by this merge. Three isolated
Go-overlay mutants were rerun and caught semantically: native-forge-producer by
TestViewCallableProducerCertificateNative/wrong-arity, javascript-forge-producer
by TestViewCallableProducerCertificateNode/wrong-arity, and eager descendants by
TestPrepareViewCallableRead/interface_Result. Production files were not mutated.

Focused lower/native/JavaScript, additional prepare-read adapter, full IR,
filtered checked-view/Node oracle (same recorded baseline exclusion), and vet
are the gates. The full repository gate remains unclaimed.

## Lane 4b reconciliation

Tip d11f3ae2 adds the assigned object/primitive plan and .a probes, with no
compiler admission change. The one plan hunk keeps the existing callable
read/producer handoff and integration checkpoint plus lane 4b's distinct scope,
ranked shapes and conditional estimate. Probes jsdoc-good/wrong,
option-type-good/wrong, node-indicator-good/wrong and diagnostic-good/wrong
are frontier evidence, not newly accepted source contracts. No mutant claim
is made for this documentation/probe-only merge. The same focused package,
full IR, filtered checked-view/Node oracle and vet gates are rerun.

## Lane 2 search follow-up

Tip 25ed1d3f (implementation 12482f1f) merges after lane 4b at the lead's request.
The only remaining conflict is one plan hunk: retain all dated owner/lane
checkpoints and append centralized integration coordination. The ArraySlice hunk
recorded by lane 2 against f1c91970 was already resolved in 8165756d: sparse-aware
slice dispatch plus graphArray/GraphTypes ownership. No emitter conflict recurs.
Search read metadata merges with owner readiness, including the nested ForOf fix.
Reference-source write refusals remain in both backends; search bad/literal, lazy,
sparse, object identity and evaluation fixtures pin behavior. All nine new search
mutants are rerun before publication; prior eleven array mutants remain carried
evidence unless explicitly rerun. Gates use the same recorded baseline exclusion.

## Lane 4c untagged reconciliation

Tip 6b3fc480 adds selector components without production source admission.
The sole docs/checked-views-plan.md hunk retains all integration/lane handoffs
and the incoming selector handoff. Owner lazy admission replaces the old eager
cast refusal with the named unsupported untagged-object-union field read. The
source-frontier assertion is updated to that observed refusal for binding-name,
option-element and structural view-read fixtures, preserving their rejection.
The initial old-message failure is logged. All 84 pairs / 186 reads remain pending;
component nested mutants prove the adapter seam, not compiler propagation.
Six selector semantic mutants are rerun and restored before normal gates.

## Lazy owner priority follow-up

Dictionary merge was aborted untouched when fetch revealed owner 5fd6445e.
This priority merge is clean: owner had consumed integration 4d863661, and no
previous conflict resolution is replaced. Native Error now physically owns only
name/message and stamps their actual string kinds; optional code is absent.
Optional-error and optional-error-fields positive controls match Node; number
and null code payload mutants retain exit-70 refusals in both backends.
The adapted census is carried evidence (1,758 tagged / 1,178 untagged descriptors,
zero production entries compiled, 259 checker diagnostic rows); it is not rerun
in this integration session. Its remaining runtime reachability is unmeasured.
A producer-certificate rollback mutant is run before normal gates.
The producer rollback is caught by optional-error-fields: Node exits 0, valid
native exits 70 at source.message with unsupported representation. The runner's
initial textual check expected uncertified storage, but this oracle prints byte
arrays; semantic exit evidence is verified from the captured log. Production
exceptions.c was restored in finally.
