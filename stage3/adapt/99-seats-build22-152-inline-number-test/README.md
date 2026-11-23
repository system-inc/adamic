# Inline the number predicates in getPendingEmitKind

Compiler accepts under the split, per Adamic's Nov 22 ruling on #bbssdcy,
family 152. Rows 141 and 142 of `docs/rulings.md` cover `typeof` narrowing
and proven predicate bodies. Row 348 governs entry into loose `unknown` or
`object` slots. The inlines test each parameter directly.

The adapted `core.ts:1780` defines the predicate as follows.

```ts
export function isNumber(x: unknown): x is number {
    return typeof x === "number";
}
```

Both arguments are parameter reads declared at `builder.ts:291` and `:292`.

```ts
    optionsOrEmitKind: CompilerOptions | BuilderFileEmit,
    oldOptionsOrEmitKind: CompilerOptions | BuilderFileEmit | undefined,
```

- At `builder.ts:294`, the old-value call becomes
  `typeof oldOptionsOrEmitKind === "number"`. The outer
  `oldOptionsOrEmitKind &&` preserves zero, `-0`, `NaN` and `undefined`
  as the local `oldEmitKind`. The function returns the current emit kind
  when that local is falsy.
- At `builder.ts:295`, the current-value call becomes
  `typeof optionsOrEmitKind === "number"`.

The adapter pins the predicate text, both parameters and both enclosing
statements. The stock checker loads `src/compiler/tsconfig.json` and follows
each call's binding through aliases to the exact pinned declaration. On
replay it validates the binding in scope at each former call before accepting
the inline. Only the two call spans are edited.

Round 2 measured the entry and compiler roots in one base/head census job
at `0efbdb08`. Each pass had 2,108 sites and 602 units with a first stop on
both sides. Family 152 cleared at `builder.ts:294:59`. The first lowering
stop in `getPendingEmitKind` (`builder.ts:290:1`) moved to
`builder.ts:294:135`, family 44: "a narrowed union member whose object tag
cannot be checked with typeof; keep differently held object kinds in
separately typed variables". The unit still stops. Each pass cleared one,
exposed one, added zero new sites and had zero newly stopping units. Loom's
unit census at main follows review.

Run `node stage3/adapt/99-seats-build22-152-inline-number-test/test.cjs TREE`
with a full applied tree. The test uses this adapter's before snapshot,
checks the exact edits and replay, and checks that both shadow forms and
four drift controls fail without writes. It also compares eleven runtime
vectors and option-read counts from the original and adapted functions in
a VM using stock TypeScript 6.0.3.
