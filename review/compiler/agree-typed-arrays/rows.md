# Source row classification

Original typed_arrays_test.go: 17 source rows, 3 accept, 14 refuse, 0 IR-only. Final: 19 source rows, 5 accept, 14 refuse, 0 IR-only. No row deleted or skipped. The two new rows also assert an IR property invisible to the JavaScript backend.

| Test | Row | Class | What I did | Why |
|---|---|---|---|---|
| TestTypedArraysLower | Uint8Array | accept | Use lowersAndAgreesWithNode; print initial length and values; remove IR assertions | Construction, conversion, writes, fill, overlapping views, iteration and out-of-bounds reads are observable |
| TestTypedArraysLower | Int32Array | accept | Use lowersAndAgreesWithNode; print initial length and values; remove IR assertions | Construction, conversion, writes, fill, overlapping views, iteration and out-of-bounds reads are observable |
| TestTypedArraysLower | Float64Array | accept | Use lowersAndAgreesWithNode; print initial length and values; remove IR assertions | Construction, conversion, writes, fill, overlapping views, iteration and out-of-bounds reads are observable |
| TestTypedArrayGaps | 1: ArrayBuffer | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 2: SharedArrayBuffer construction | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 3: SharedArrayBuffer parameter | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 4: DataView | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 5: resizable ArrayBuffer | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 6: Int8Array | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 7: buffer view | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 8: cross-kind set | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 9: slice | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 10: try around set | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayGaps | 11: try around length conversion | refuse | Unchanged | Asserts NotYet and the existing diagnostic |
| TestTypedArrayViewsCannotChangeRepresentation | 1: assignment to length shape | refuse | Unchanged | Asserts NotYet, regardless of the test name |
| TestTypedArrayViewsCannotChangeRepresentation | 2: cast to length shape | refuse | Unchanged | Asserts NotYet, regardless of the test name |
| TestTypedArrayViewsCannotChangeRepresentation | 3: nested length shape | refuse | Unchanged | Asserts NotYet, regardless of the test name |
| TestTypedArrayFromArray | [1, 2] (audit witness promoted) | accept | Add agreement row printing length and every value; named FromArray assertion | JavaScript emission ignores FromArray, so stdout cannot certify native constructor dispatch |
| TestTypedArrayFromEmptyArray | [] (audit witness promoted) | accept | Add agreement row printing length and every value; named FromArray assertion | JavaScript emission ignores FromArray, so stdout cannot certify native constructor dispatch |

The audit taste_representation label names a multi-file slice. Its Uint8Array witnesses live in review/test-audit/internal-lower-taste_representation/witness/main.go, not taste_representation_test.go. That test file contains six refusal rows and remains unchanged.

Awkward requirement: both exact audit mutants leave JavaScript output byte-for-byte unchanged because internal/javascript/javascript.go emits TypedArrayNew without consulting FromArray. Length and value printing alone cannot kill either mutant. The added targeted IR checks name this unobservable backend metadata explicitly. Mutant failures therefore prove the dispatch assertions, not the stdout comparisons. No native behavior claim is made.
