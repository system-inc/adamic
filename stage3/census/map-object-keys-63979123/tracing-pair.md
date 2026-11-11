# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 8 | 8 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 1 | 1 |
| Regressions | 0 | 0 |
| before Boundary | 26 | 26 |
| before NotYet | 23 | 23 |
| before Refused | 9 | 9 |
| after Boundary | 22 | 22 |
| after NotYet | 19 | 19 |
| after Refused | 9 | 9 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 1 | 1 | 1 | 1 | 0 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 3 | 0 | 3 | 0 | 3 |
| a call returning any | 1 | 1 | 1 | 1 | 0 |
| a cast the runtime can't check | 2 | 2 | 2 | 2 | 0 |
| a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 1 | 1 | 1 | 1 | 0 |
| a spread after the first field | 1 | 1 | 1 | 1 | 0 |
| a spread into record storage from a fixed object | 1 | 1 | 1 | 1 | 0 |
| a value of type any | 1 | 1 | 1 | 1 | 0 |
| an ambient value declaration | 2 | 2 | 2 | 2 | 0 |
| assigning a field of a value | 3 | 3 | 3 | 3 | 0 |
| failed statement; state rolled back; continued at next statement or declaration | 26 | 22 | 26 | 22 | 4 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| reading conditionalType | 1 | 1 | 1 | 1 | 0 |
| reading indexedAccessType | 1 | 1 | 1 | 1 | 0 |
| reading objectFlags | 4 | 4 | 4 | 4 | 0 |
| reading recursionIdentityMap | 1 | 0 | 1 | 0 | 1 |
| reading substitutionType | 1 | 1 | 1 | 1 | 0 |
| reading tracingEnabled | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field without its source-slot type certificate | 3 | 3 | 3 | 3 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

- `/workspace/scratch/ff5eddz-adapted/src/compiler/tracing.ts:232:19`: NotYet: a value of type any

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

- `/workspace/scratch/ff5eddz-adapted/src/compiler/tracing.ts:232:19`: NotYet: a value of type any

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
