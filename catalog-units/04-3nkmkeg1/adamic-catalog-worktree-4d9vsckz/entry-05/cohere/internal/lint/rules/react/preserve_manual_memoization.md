# Manual memoization parity

## Measured outcome, 2026-09-07

The measured ahra population has **zero false positives and one defensible
finding**. This is the task's correctness target, not a claim that every possible
program is covered or that the warning count should be forced to zero.

An archived binary from cohere `6f04852` reports exactly one occurrence on both
the frozen source snapshot and the live ahra checkout. The live comparison uses
`s l no-cache --linter eslint`, which reports zero occurrences of this rule.
Resolved ESLint configuration on the remaining file enables the rule at severity
2. The binary SHA256 is
`932daf0210614b0f996b12b8caef3796968a448d53ed08ec6e0b19f020f5d5e2`.

All comparisons retain diagnostic multiplicity rather than deduplicating
declaration anchors. The final bounded repairs measure separately:

| Commit | Repair | Before | After | Removed | Added |
| --- | --- | ---: | ---: | ---: | ---: |
| `0740825` | Primitive property constraints | 9 | 7 | 2 | 0 |
| `44d6bf9` | Imperative handle effects | 7 | 6 | 1 | 0 |
| `04501c3` | Computed read kinds | 6 | 1 | 5 | 0 |
| `6f04852` | Mixed-capture conditional closure | 1 | 1 | 0 | 0 |

The last repair covers an additional reference-clean reproduction discovered
during reduction, even though it does not change the live population. The
implementations and corpus attribution are documented beside the HIR passes in
PRIMITIVE_PROPERTY_CONSTRAINTS.md, IMPERATIVE_HANDLE_EFFECTS.md,
COMPUTED_READ_KINDS.md and MIXED_CLOSURE_CALLS.md.

## Known-correct divergence: mount-only field seed

The remaining source is
`libraries/structure/source/components/forms/fields/height/FieldInputHeight.tsx`.
Cohere anchors its one ValueUnmemoized finding at the `field` declaration on line
51; the responsible useMemo begins on line 62 and reads `field.value` with an
empty dependency array. The declaration anchor is an existing reporting-span
limitation, not evidence that useField itself is a memoization error.

The reference's original-source verdict is Suppression: two
`react-hooks/exhaustive-deps` disable comments cause React Compiler to skip the
component before preservation validation. That bailout alone did not adjudicate
the cohere finding. The following controlled interventions do:

| Source | React Compiler 1.0.0 | Cohere |
| --- | --- | --- |
| Original | Suppression bailout | 1 finding |
| Remove only the suppression comments | 1 preservation CompileError | 1 finding |
| Also replace the memo's empty deps with `[field.value]` | CompileSuccess | 0 findings |

The preservation error names the missing inferred dependency `field.value.value`.
No executable source changes in the first intervention. The live and frozen
files are byte-identical, SHA256
`b7b895b3835cc7cdd58920fdf50a4f747813dcb68357fd5d3317b1280c2f9b19`.
The reduced missing/correct pair is permanent in
TestPreserveManualMemoizationReportsMissingSeedDependency. It prevents a parity
sweep from deleting a diagnostic merely because ESLint's compiler bailed out.

The source's stated intent is initialization only. No application source was
changed to lower the count. If that code is revised, preserve its initialization
semantics explicitly rather than blindly following a count or widening the
dependency array without considering the intent.

## Evidence boundaries and regression protection

Compiler 1.0.0 ran with validatePreserveExistingMemoizationGuarantees enabled.
Independent named and namespace bad-dependency controls each produced one
preservation error; the correctly dependent control compiled successfully.
WisdomHome and InputMultipleSelect originally had earlier Todo bailouts. Private
copies without the unsupported try wrappers reached CompileSuccess while cohere
initially retained the same diagnostics. Permanent, reference-checked
reproductions isolate the repaired mechanisms without those bailouts.

Each repair has tests demonstrated red against its previous implementation and
nearby controls that retain genuine diagnostics. The full lint suite and build
pass. The preservation corpus retains 28/28 positive goldens and 69/69 clean
fixtures, with zero unsupported clean fixtures. No preservation-validator
exemptions, rule disabling, application suppressions or fixer runs were used to
achieve the result.
