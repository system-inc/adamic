# Deferred reach revision: exact causes and controls

All three diagnostic requests declare one binder, T, and recover zero mappings.
`mapping-diagnosis.json` retains the three caller units, call sites, signature/body
entry and exit events, and the first-stop comparison to the original 79-file
census. All three first stops match, including the finding-free initializer.
The diagnostic adds no type-printing checker queries before lowering.

## createComputedCompilerOptions: unread mapped record, not unit kind

The caller is the census variable unit utilities.ts:9042:1, at :9042:26; the
declaration is :9031:1. Its parameter and inferred return have T behind a mapped
record, keyof T and indexed accesses T[K]. internal/lower/generic.go inferTypes
recovers direct binders, optional/union witnesses, reference arguments and call
signatures. It does not walk arbitrary record properties or mapped record types.
Consequently T is absent from concreteTypes. The concrete-context test is false,
although lowerBody enters and exits successfully because the body simply returns
options and needs no direct runtime representation of T.

This is a lowering type-argument readback limitation, not a missing variable-unit
recording path. Per the user's conditional stop instruction, no inference fix or
relaxation of the flag was made. The row remains unresolved under the requested
classes, with its exact cause recorded. It is not assigned an invented class.

## updateCreateProgram: unread record field, not success-only recording

The caller is executeCommandLine.ts:968:1, call :974:5; the declaration is :981:1.
T occurs in the createProgram field of the anonymous host parameter record,
{ createProgram: CreateProgram<T> }, and not in its void return. inferTypes does
not inspect that property, so it also recovers zero mappings. The body is entered
with no fully mapped concrete context and fails at :983:26 while lowering the
callback signature: NotYet, a function returning T.

The flag is set at lowerBody entry, before body lowering, when all binders are
mapped. It is not conditional on successful lowering. The README's positive
all-binders-mapped definition is correct. Its broader negative implication that
false flags prove unexamined bodies was wrong and is corrected. This row also
remains unresolved; no mapping or recording-policy change was authorized here.

## toSearchResult: request stops in its signature

The caller is moduleNameResolver.ts:2656:1, call :2661:16, passing undefined.
That absent value supplies no binder witness; the return alias puts T under a
record value property, which the readback does not inspect. The signature fails
at moduleNameResolver.ts:3414:28: NotYet, a value of type T | undefined. The body
is never entered. The report assigns the requested sixth class, reached, stops
at its signature, and quotes this stop.

## Controls, mutants and rerun scope

The expanded audit_deferred fixture calls initializerOnly<number> only from a
top-level const initializer. It records instantiated=true and the exact variable
unit, call site and number argument in legacy, full and speculative modes.
enteredButStopped<number>, also called from an initializer, records true and its
witness before failing on a written-any body. Both controls preserve the original
boolean in comparisons with the old binary, byte for byte except instantiated_by.

initializer-witness-mutant.go.txt drops only the first initializer's witness.
The audit exits 1 at the new initializer witnesses assertion while the boolean
remains true. initializer-mutant.log and initializer-mutant.exit retain that red.
The paired reader suite passes 20 tests. Its mutant removes only the metadata
exclusion and runs the unchanged positive eligibility test; that test is red with
ValueError: unit eligibility differs. The complete failure is in pair-mutant-tests.log.

No production or observation Go code changed. No lowering state was added. The
existing whole-statecopy and 79/81-file byte/meter-identity evidence still applies.
Only the three diagnostic caller units, the changed fixtures, baseline fixture
comparisons, and the pair tests were rerun. Fixed recording moves zero rows from
unreached to reached: the hypothesized recording defects were not present.
