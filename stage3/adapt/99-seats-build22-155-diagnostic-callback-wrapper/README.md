# Family 155 scanner diagnostic callback

This adaptation applies Adamic's Nov 22 21:49Z ruling on #bbssdcy for family 155.
Compiler is its accepting seat under the split, after review and Loom's unit census at main.

The four conflict-marker branches in `createScanner`'s nested `scan`
pass overloaded `error` as a value to `scanConflictMarkerTrivia`.
The base first-error census records the less-than branch. The
compiler-directory replay of the first head also records the equals,
greater-than, and bar branches in `scan`, with `createScanner` as the
recorded parent and top unit.

| Branch or enclosing function | Adapted base position and disposition |
| --- | --- |
| Less-than | `scanner.ts:2195:68`, wrapped |
| Equals | `scanner.ts:2224:68`, wrapped |
| Greater-than | `scanner.ts:2246:68`, wrapped |
| Bar | `scanner.ts:2286:68`, wrapped |
| `skipTrivia` | `scanner.ts:711`, helper callback omitted |
| `scanJsxToken` | `scanner.ts:3740`, keeps the overloaded callback |

The Nov 22 ruling on #bbssdcy allows a monomorphic wrapper because the
private helper uses the callback only in a presence guard and a direct call.

```ts
error(Diagnostics.Merge_conflict_marker_encountered, pos, mergeConflictMarkerLength);
```

The callback remains optional. When present, its position and length
parameters are required, as both arguments are supplied at the only
invocation. The arrow `(diag, at, len) => error(diag, at, len)` selects
the existing overload with required position and length. The helper
neither compares nor stores the callback.

The governing entries in `docs/rulings.md` at `ab9af917` are:

- Line 132, "Bodyless overload signatures are erased"
- Line 133, "A call resolving to an overload whose declared result is narrower than the implementation's"
- Line 213, "Nested overload signatures are skipped only when the same statement list holds the implementation"

The adapter checks the full reviewed helper declaration before first and
repeat application, allowing the callback type for either state. It also
checks all six calls and the direct case guard around each wrapped call.
The only identifiers with the helper's name must be the reviewed declaration
name and the six direct callees, establishing their lexical binding.
The JSX callback remains assignable under stock TypeScript and did not
surface in the selected scan unit.

## Test

After `stage3/apply.sh <applied-tree>`, run from the Adamic checkout.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-155-diagnostic-callback-wrapper/test.cjs <applied-tree>
```

The test extracts `src/` from this adapter's before snapshot, applies the
adapter in a scratch tree, and emits and loads the real scanner with
stock TypeScript 6.0.3. It checks diagnostic code, position and length,
token, and cursor for all four markers with trivia enabled and skipped,
plus JSX and omitted callbacks. It verifies a byte-identical repeat
and rejects helper identity drift, shadowing, aliasing, and call-inventory drift on both
input states without writing the planted scanner.
