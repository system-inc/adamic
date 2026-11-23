# Carry the pragma argument count beside each tuple

Adamic ruled family 142 on Nov 22 2026 (#bbssdcy, "yes as sketched", hers wherever JavaScript moves) and accepts the head after its review and census; on Nov 23 at 09:06Z she answered the first read's runtime-export question with (a): "A runtime mutation of an internal table by outside code is outside the contract, the same answer as 41." The refusal
is at `parser.ts:10769:25` in the `getNamedPragmaArguments` unit on the
reported cut. The loop keeps its absent-args guard and uses a count stored on
the definition. The indexed read and the tuple types that drive
`PragmaArgumentType` retain their existing forms.

Ruling row 334 governs tuple operations and preserves readonly tuple views.
Row 397 governs changes to the public TypeScript API. The table, its
specification interfaces and the parser helper are internal.

## Producers

Stock TypeScript 6.0.3's reference search identifies one producer of pragma
definitions with arguments, the `commentPragmas` table. Each tuple is an
object-literal property with its full arity visible at construction. These
are the positions on upstream pin 050880ce59e30b356b686bd3144efe24f875ebc8.

| Definition | Argument literal | Arity |
| --- | --- | ---: |
| reference | `src/compiler/types.ts:10242` | 6 |
| amd-dependency | `src/compiler/types.ts:10253` | 2 |
| amd-module | `src/compiler/types.ts:10257` | 1 |
| jsx | `src/compiler/types.ts:10267` | 1 |
| jsxfrag | `src/compiler/types.ts:10271` | 1 |
| jsximportsource | `src/compiler/types.ts:10275` | 1 |
| jsxruntime | `src/compiler/types.ts:10279` | 1 |

`ts-check` and `ts-nocheck` omit args. The two parser lookups read the original
entries. The classifier checks membership in the table. The named-argument
helper's sole caller supplies a table entry. The uses neither construct
another definition nor copy or mutate a definition's args. The six-field
`reference` entry uses the separate XML parser path.

The adapter locates the declarations and loop with the stock syntax tree
and checks their snapshot hashes. It obtains each literal count from the
array syntax node and types each corresponding field as the concrete
tuple's `["length"]`. A wrong count at any producer is a type error. The
parser's readonly optional count is present wherever its existing guard
finds args, as established by the table's concrete types and sole-caller
audit.

Run `node stage3/adapt/99-seats-build22-142-pragma-arg-counts/test.cjs <adapted-tree>`
from the repository root. The test reads the adapter's before snapshot,
checks the clean apply and byte-identical no-op, drives the emitted helper
and table beside their originals, and compares the stock and adapted
bundle's parser pragmas. The helper cases catch both a shortened loop
bound and an extra counted AMD dependency argument. An optional second
argument names a scratch adapter to exercise.
