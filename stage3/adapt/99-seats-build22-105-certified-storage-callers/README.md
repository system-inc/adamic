# Certified storage callers in builder.ts

Family 105 applies Adamic's Nov 22 21:49Z ruling under `docs/rulings.md`
rows 133, 135 and 136. Compiler accepts 105 under the widened split.
Family 106 applies the same ruling with Adamic's Nov 23 08:12:04Z and
10:18:13Z answers and row 50. Adamic accepts 106.
The representation comparison at `internal/lower/overload_certificate_array.go:541`
still gates certified overloads. This adaptation changes JavaScript.

## Family 105

At the base census site `builder.ts:1421:25`, in unit `builder.ts:1258:1`,
`toFileIdListId` calls the mapped `arrayFrom` overload.

Before

```ts
const fileIds = arrayFrom(set.keys(), toFileId).sort(compareValues);
```

After

```ts
const fileIds: IncrementalBuildInfoFileId[] = [];
for (const path of set.keys()) {
    fileIds.push(toFileId(path));
}
fileIds.sort(compareValues);
```

The loop applies `toFileId` in set iteration order. The existing comparator
orders the resulting IDs before the caller uses them.

## Family 106

At the base census site `builder.ts:2416:31`, in unit `builder.ts:2277:1`,
`toPerFileEmitDiagnostics` calls the mapped-value overload of `arrayToMap`.
This edit applies rulings rows 133, 135, 136 and 50.

> **2. 106: (a).** Keep the skip: `const path = toFilePath(value[0]); if (path !== undefined) result.set(path, value[1]);`, spelling the local `Path | undefined` so the guard types. That matches stock's JavaScript exactly. That `toFilePath`'s `!` panics natively first on corrupt input is :50's loud stop, and that's fine.

> **106: keep `if (!diagnostics) return diagnostics;`.** It's what makes the head value-identical to stock for a JSON `null` in build info, which the reader showed is real. My ruled text was a sketch, and value identity is the rule it serves. It's mine to accept after review-k's read and 105's several-key sort test.

Before

```ts
return diagnostics && arrayToMap(diagnostics, value => toFilePath(value[0]), value => value[1]);
```

After

```ts
if (!diagnostics) return diagnostics;
const result = new Map<Path, readonly ReusableDiagnostic[]>();
for (let i = 0; i < diagnostics.length; i++) {
    const value = diagnostics[i]!;
    const path: Path | undefined = toFilePath(value[0]);
    if (path !== undefined) result.set(path, value[1]);
}
return result;
```

Adamic's 10:18:13Z ruling keeps the early return because the real build-info
reader can pass through JSON null, which stock's `diagnostics &&` returns as
given. The call becomes a loop in emitted JavaScript and the return values
follow stock. The indexed walk preserves that rule for the empty object and
the array-like object from review-k's real-reader cases. A duplicate key keeps its insertion position and
receives the later diagnostics. On Node, an out-of-range file ID yields an
undefined path and the caller skips that record. Adamic's native build stops
at `toFilePath`'s postfix `!` with exit 70 under row 50.

Run `NODE_PATH=stage3/api/node_modules node stage3/adapt/99-seats-build22-105-certified-storage-callers/test.cjs <adapted-tree>`
from the checkout with a tree produced by `stage3/apply.sh`. The test reads the
adapted helpers and this adapter's output from its after snapshot, and the
unrecognized-call input from its before snapshot. Later edits to
`builder.ts`, `core.ts` or `corePublic.ts` leave the test valid.
The several-key case checks the actual 105 sort and mapper
order beside stock. The 106 table covers the six original inputs and four
truthy non-array values. The reader cases exercise ten build-info inputs.
The adapter cases check clean reapply, each leading-indent drift, and an
unrecognized 106 call with 105 still in its original form.

The sort-removal plant fails the several-key case. Removing the undefined
path guard fails the missing-path case. Replacing the indexed walk with
`for...of` fails the object cases in the table and the reader. Moving an
indentation check back into its rewrite branch fails that block's drift case.
Writing 105 before validating 106 fails the unrecognized-call case.
