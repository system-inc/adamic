# Two module-load tables build without holes (#s59veah)

Base commit `a06ab2eb1c30360e6e816efd5f622d6c3628fe90`. TypeScript v6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`.

Adamic's rule (docs/rulings.md:318, amended on #s59veah) stops an array write
that leaves a hole. The dynamic count on typescript/sparse-array-writes
(3ad1b557, REPORT.md, "Module initialization") found six sites that write
holes at module load, so under that rule every program would stop before it
reads input. Apple adapts makeReverseMap in scanner.ts. This adapter takes
the other five: the four bracket rows in emitter.ts and the one write in
core.ts's arrayToNumericMap.

Adamic's rule for the form: a table read only by index gets an explicit dense
fill at the write site, and a table whose readers can see holes gets a Map or
Set. A dense fill that would push thousands of slots may take the Map form
instead, with the reason named.

## createBracketsMap: a Map, since a dense fill would push 8,189 slots

`const brackets: string[][] = []`, then four writes, at `ListFormat.Braces`,
`Parenthesis`, `AngleBrackets` and `SquareBrackets`. Those are `1 << 10` to
`1 << 13` in src/compiler/types.ts, so 1024, 2048, 4096 and 8192. A dense fill
would push 8,189 undefined slots to hold four rows.

Its readers, every one in src/ at the pin:

- `const brackets = createBracketsMap();` at emitter.ts module scope, the table's
  only binding;
- `getOpeningBracket`, `brackets[format & ListFormat.BracketsMask][0]`;
- `getClosingBracket`, `brackets[format & ListFormat.BracketsMask][1]`.

Both readers are called from createPrinter's emitListWorker, under
`if (format & ListFormat.BracketsMask)`. Nothing enumerates the table, reads its
length or tests a key with `in`.

The table is now `new Map<number, readonly [string, string]>()` with the four
rows set in the same order. The readers read `.get` with the same masked key
and the same slot. A key with a row gives the same string. A key with no row
gave undefined from the array and gives undefined from the Map, and both forms
then stop alike: Node throws reading the slot of undefined, and Adamic stops on
the `!` that adaptation 33 put on the read.

## arrayToNumericMap: an explicit dense fill, since its one table is read only by index

`const result: (T | U)[] = []`, then `result[makeKey(value)] = makeValue(value)`
for each value. Its only caller in src/ at the pin, checked by search over the
whole of src/, is classifier.ts:

```ts
const noRegexTable: true[] = arrayToNumericMap<SyntaxKind, true>([...13 tokens], token => token, () => true);
```

The keys run from `NumericLiteral` (9) to `TrueKeyword` (112), so the filled
table holds 113 slots. Its one reader is `handleToken` in createClassifier's
`getEncodedLexicalClassifications`, on a slash:
`!noRegexTable[lastNonTriviaToken]`, a read by index, where undefined and a
hole give the same value.

The write site now fills each slot below the key with undefined first:

```ts
const key = makeKey(value);
while (result.length < key) result.push(undefined);
result[key] = makeValue(value);
```

The key is still computed before the value. The overloads return
`(T | undefined)[]` and `(U | undefined)[]`, and noRegexTable is annotated
`(true | undefined)[]`, which is what the filled slots hold. arrayToNumericMap
is `@internal`, so the public declarations don't move.

## The adapter

`adapt.cjs` finds each owner through the stock TypeScript 6.0.3 AST and pins
its whole text, as it arrives or as this leaves it: `createBracketsMap`,
`getOpeningBracket`, `getClosingBracket`, the three `arrayToNumericMap`
declarations, and the `noRegexTable` statement. As it arrives, the two bracket
readers carry adaptation 33's `!` on both reads, and arrayToNumericMap carries
adaptation 30's `array[i]!`. Any other text, a parse error, mixed line endings,
and a partly adapted tree are refused before any write. A second run writes
nothing. `rewrite(tree, { restore: true })` is the mutant: every owner back as
it arrived.

This directory sorts after 30-indexed-reads, 33-indexed-reads-emit and
99-correctness-core-*, the adapters before it that touch these files. No
directory after it touches these owners. 99-correctness-receiver-printer pins
createPrinter whole by hash, and createPrinter calls the two readers but holds
neither them nor the table.

## What is proven

It's JavaScript-changing, value-identical. Two jobs at the commit, named with
their results on #s59veah:

- **The lane count.** An instrumented copy of the adapted tree keeps each table's
  old form beside it, built by upstream's own code, and records, at each of the
  three readers, the value read from the new form and from the old one. It runs
  the lane oracle whole. It needs every reader reached and 0 differences. Its
  controls show the recorder firing on a compile, on the classifier, and on a
  difference fed to it.
- **The proof.** It applies the adaptations into two fresh trees, which must be
  byte-identical. A second run writes nothing, and stock tsc reads 0
  diagnostics over the adapted tree. The public typescript.d.ts and
  tsserverlibrary.d.ts are byte-identical with and without this adapter. The
  restore mutant gives the tree without it byte for byte. Then a write probe
  over the three owner files checks module load. Without the adapter, the
  module-load hole writes come back: four bracket rows and three
  arrayToNumericMap keys. With it, there are none.
