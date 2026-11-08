# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before Boundary | 10 | 10 |
| before NotYet | 13 | 13 |
| before Refused | 3 | 3 |
| after Boundary | 10 | 10 |
| after NotYet | 13 | 13 |
| after Refused | 3 | 3 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| a method read as a value (compare would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a value of type NonNullable<T> | 2 | 2 | 2 | 2 | 0 |
| an ambient value declaration | 1 | 1 | 1 | 1 | 0 |
| an array of NonNullable<T> | 1 | 1 | 1 | 1 | 0 |
| an array of T | 2 | 2 | 2 | 2 | 0 |
| failed statement; state rolled back; continued at next statement or declaration | 10 | 10 | 10 | 10 | 0 |
| inherited library member compare read as an own field | 1 | 1 | 1 | 1 | 0 |
| reading comparer | 1 | 1 | 1 | 1 | 0 |
| reading deduplicated | 1 | 1 | 1 | 1 | 0 |
| reading next | 1 | 1 | 1 | 1 | 0 |
| the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | 5 | 5 | 5 | 5 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
