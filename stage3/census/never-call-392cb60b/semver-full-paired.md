# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 4 | 3 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 6 | 6 |
| Regressions | 0 | 0 |
| before Boundary | 9 | 8 |
| before NotYet | 7 | 7 |
| before Refused | 1 | 1 |
| after Boundary | 7 | 7 |
| after NotYet | 5 | 5 |
| after Refused | 1 | 1 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| a case whose type differs from the switch's | 1 | 1 | 1 | 1 | 0 |
| a constructor object escaping before static fields are initialized | 1 | 1 | 1 | 1 | 0 |
| a destructured name that isn't plain | 1 | 1 | 1 | 1 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target Comparator[][] | 1 | 0 | 1 | 0 | 1 |
| checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } | 1 | 0 | 1 | 0 | 1 |
| failed statement; state rolled back; continued at next statement or declaration | 9 | 7 | 8 | 7 | 1 |
| reading major | 1 | 1 | 1 | 1 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never
- `/workspace/never-call-adapted/src/compiler/debug.ts:252:33`: NotYet: a value of type never

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
