# The resolution stack's length, read as a property

TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Task #kr1v0tn, shape C, for Adamic's accept. It changes the
emitted JavaScript of `src/compiler/checker.ts` at one line. pushTypeResolution is local to createTypeChecker, so
typescript.d.ts doesn't move and this directory carries no `api-sanctions.json`.

## Why

When pushTypeResolution finds a cycle, it reads how deep the resolution stack is through an object pattern:

```ts
const { length } = resolutionTargets;
```

resolutionTargets is a `TypeSystemEntity[]`. Adamic holds an array as an array, never as a plain object, so destructuring
it through an object pattern stops there ("an array going where a plain object type is held"), and the loop's read of
`length` on the next line stops behind it ("reading length"). It's the only site of this shape on the full-mode census
(J-vt1zf5t).

## The edit

```ts
const length = resolutionTargets.length;
```

The same value, read once, at the same point, before the loop. adapt.cjs refuses before any write unless
createTypeChecker holds exactly one function pushTypeResolution, and its body exactly one statement reading
resolutionTargets' length, written as upstream's pattern or as this read. A second run writes nothing.

## What was measured

Sealed, over two adapted trees built by apply, without and with this directory (J-bnf7z2k at 5f71057a, main 0efbdb08):

| Measurement | Result |
| --- | --- |
| The trees' `src/` | checker.ts alone differs, at line 11501 alone |
| Stock tsc over all 713 files of src, as (file, code, message) | 2,719 before and after, 0 new, 0 gone |
| Emitted checker.js, 47,514 lines | one line differs, the statement itself |
| Second run | writes nothing |
| The statement rewritten another way | refused, nothing written |

**The value, counted in Node** (J-panmd8z). Stock typescript.js bundles the statement as
`const { length: length2 } = resolutionTargets;`. Instrumented there to read the length both ways and compare, it ran
over 155 programs, 154 of tsc's own compiler test cases whose names speak of circularity, recursion or cycles and tsc's
own src/compiler: **the statement ran 126 times, and the property read differed from the pattern 0 times.**

**The census of checker.ts, full mode**, before (the tree without this directory, which is the restoring mutant) and
after: 1,738 sites to 1,736. The two that go are this site, 11501:32 "an array going where a plain object type is held",
and 11502:57 "reading length" behind it. Nothing is created. Three more sites (30329:16, 39339:9, 39437:9) read as gone
and new in a plain comparison only because their reasons quote the tree's own scratch path, `before` against `after`;
each is the same fact at the same place.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or `CENSUS_TYPESCRIPT`, as apply runs every
adaptation.
