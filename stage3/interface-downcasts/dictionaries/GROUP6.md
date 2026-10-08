Built four optional-receiver candidate pairs with 15 new source controls and repaired required-field presence checking.
Commits: this group follows 07603363; shared hooks are adamic_object_optional_view and adamicViewField, documented in the lane plan.
Checks: targeted lower/native/JS/oracle gate passed; oracle 64.371s, lower 0.570s, native 8.777s, JS built with no matching tests.
Mutants: compiling optional-absence conflation mutant loses the pinned required-field refusal in both backends; all dictionary and targeted readiness/lazy mutants pass.
Not covered: direct optional-base downcast form, optional dictionary-index receiver, remaining lookup/union/enumeration work; 14 candidate pairs / 61 candidate reads remain.

The working date remains October 11, 2026, 23:00 UTC. Counts are candidate
inventory counts; exact runtime reachability remains unmeasured.

Certified: ParsedCommandLine | undefined watchOptions (5 reads), options
(type IDs 94919 and 11560, 5 and 2 reads), and wildcardDirectories (1 read).
Original representative fixtures use a guarded checked downcast followed by
an optional receiver read. The initial direct optional-base-to-optional-target
required-options cast was refused by existing admission; this group does not
claim to change that separate cast form.

The guarded fixture exposed a native soundness bug: optional chaining admitted
a missing required dictionary field on a present receiver. The native helper
now distinguishes absent receiver, absent optional field and missing required
field. Its undefined-payload path likewise requires field absence permission
or a supported maybe scalar contract. The JS helper now allows undefined due
to optional chaining only for an absent receiver. The changes do not add a
record representation or inspect unread dictionary elements.

Every shape has valid, wrong-number, missing-field, absent-receiver and
explicit-undefined cases. Valid/allowed absence agrees with Node in sanitized
C, release C and JS, with the leak check. A missing required container pins:

    adamic: panic: field read failed: view?.options is not initialized; expected CompilerOptions, found missing

A present required field containing undefined pins:

    adamic: panic: field read failed: view?.options is not a CompilerOptions; expected CompilerOptions, found nullish

The C/JS optional-absence mutant changes field-absence permission to
absent || optional. Both mutated programs compile, exit 0 and print undefined,
violating the exact exit-70 required-field control. Existing dictionary skip,
wrong-shape, transitive, container and producer-certificate mutants reran.

Final command:

    ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestCheckedViewDictionary|^TestCheckedViewLazy|^TestOptionalCheckedRead|^TestOptionalClassReads$|^TestOptionalNullAndLiteralReads$|^TestLiteralOptionalOracleCatchesMutant$|^TestFieldReadinessRepresentation$|^TestReadinessMutants$|^TestUninitializedIsNotNullishMutant$|^TestUniformFieldsMatchNode$|^TestOptionalWriteMissingSlotRemainsChecked$|^TestOptionalCastKeepsHiddenFieldCheck$|^TestViewDictionaryDescriptors$' -count=1 -v

Final raw log: logs/group6-final-gate.log. Initial cast/refusal and missing-field
observations are retained. Full repository baseline failures from GROUP3.md
remain; this is a targeted gate, not a whole repository green claim.
