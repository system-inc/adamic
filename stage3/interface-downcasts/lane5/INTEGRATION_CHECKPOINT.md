Built: merged integration 1dd7b382 and preserved complete callable checks through nullable-field dispatch.
Commits: stored markers b6e53fb4; original witness contracts d1153e5d; this merge consumes only codex/views-integration.
Commands/results: uncached callable/nullish oracle 58.416s; full IR 19.626s; focused lower/native/JavaScript pass; vet passes; restored broad oracle 138.136s.
Mutants: 14 compiler erasures caught this session: ten source checks, two original-pair arity erasures, two nullable-dispatch erasures; production restored.
Limits: 4 static candidate pairs/34 reads certified in fixtures; 304 pairs/1469 reads remain; highest Unknown predicate contract remains blocked, before adapter rank 53.

Three conflicted hunks were resolved individually:

- Plan: preserve both the lane-5 checkpoints and the incoming dictionary/array
  hook records; append the named nullable-callable hooks.
- interface_cast.go: retain callableViewContract's optional-callable recognition;
  use the owner's expanded nullable/object-primitive admission and remove the
  obsolete unconditional nullish refusal. Accessor and write-slot guards remain.
- object.go: attach complete callable metadata before the owner's nullable
  dispatch and possible Union/Narrow wrapping; retain both paths.

The initial merge gate exposed a real overlap: the newer nullable dispatch
bypassed callable shape checks. A supported nullable callback also bypassed the
old same-name proof, leaving its wrong result unchecked. The owned read adapter
now completes signatures on boxed Union reads, and the minimal hooks in native
and JavaScript view_nullish.go call owned view_callables_nullish.go helpers after
presence/readiness/kind validation. Permitted null/undefined remain unchanged.
Every present callable must match its independently recorded producer signature.
The prior owner compile-refusal test now pins exit 70 at node.value for the
incompatible result in both native modes and JavaScript.

The separate nullable-wrong-arity source returns a string and never reads its
missing numeric argument. Native and JavaScript hook-omission mutants therefore
finish valid executable code with exit 0 and forbidden output; the test catches
the lost named exit-70 refusal without relying on an ABI fault. Logs and the
restoring runner are committed. It adds no candidate-pair certification.

Original witness verification covers all 1503 static source spans. FileWatcher
uses its complete original declaration; Program's canonical callback witness
preserves its member alias and source handoff with a reduced receiver/helper.
These are member-contract fixture certifications, not complete original Program
or whole-tsc execution certificates. Exact runtime reachability remains unknown.
Matching the four certified contracts into the Unknown inventory accounts for
4 pairs/34 candidate reads there too: 2814 pairs/11029 reads remain in that upper
bound. The inventories overlap and must not be added.

Blocking boundary before rank 53:

The highest Unknown-fallback candidate, Debug.assert (rank 1, 439 reads), has a
predicate-valued callback contract with unknown and optional parameter slots.
The pinned reached-read fixture currently stops at its predicate declaration:
adamic/no-type-predicate, no body proving this parameter. Its original source
runs on Node and prints function. This is a per-fixture declaration/proof
boundary, independent of the whole-tsc checker diagnostics. Removing that proof
guard would trust a predicate claim on a viewed callable without certifying its
producer. It needs predicate-contract/proof support and the corresponding
boxed/optional signature handling. b131a36d is not an ancestor of integration
1dd7b382; no other lane branch was merged directly. Integration's own report also
retains the predicate-admission boundary. This checkpoint does not claim that
merging the predicates worker alone resolves every callable proof/ABI obligation.

The highest static candidate is an intrinsic string method, not an ordinary
closure field. Its detached read retains the language's unbound-method refusal.
The original direct-call census span is verified, but that does not certify an
intrinsic producer signature or justify counting an invented callback object as
138 completed original reads. Intrinsic production/dispatch remains unclaimed.

The October 12 working date remains conditional, with the predicate and intrinsic
boundaries outstanding. No completion date is represented as a verified result.

Gate limits and inherited failures:

The first broader oracle failed three optional-read diagnostic-pin suites:
TestOptionalCheckedReads, TestOptionalClassReads, TestOptionalNullAndLiteralReads.
All twelve failing cases reproduce on a detached unmodified 1dd7b382 checkout;
actual reads still exit 70 with the field, expected type and found value. Their
pins expect the older field reader wording instead of the incoming nullable
reader wording. No production check or unrelated optional test was relaxed.
The restored broad filter excludes those three suites and the two nominal-write
baseline cases previously reproduced on ba59427. Its exact command is:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View|TestPrepareViewCallableRead|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload' -skip '^TestOptionalCheckedWrites$/^nominal-(subclass-)?slot$|^TestOptional(CheckedReads|ClassReads|NullAndLiteralReads)$' -count=1 -timeout 15m
```

Full repository gate is not claimed. The carried integration patch artifact has
pre-existing whitespace issues; it was preserved. Own source diff/format checks
pass. New source allocation counts include the nullable arity witness and refresh
the optional-absence row for the incoming nullable reader's ownership operations.
