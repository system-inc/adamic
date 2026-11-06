# 99y measurement, before editing

Base and compiler: origin/main 322f80565edaa978d928ca559803c578ed67c105.
Upstream: 050880ce59e30b356b686bd3144efe24f875ebc8, stock TypeScript 6.0.3.
The complete baseline `stage3/apply.sh --build-product /workspace/y-before`
exited 0 before these measurements or any 99y source edit.

`measure.cjs` uses checker.getTypeAtLocation and checker.getResolvedSignature.
It also compares type object identity with combine's type parameter, rather
than assuming every displayed T is the same type parameter.

| Baseline call | Narrowed array | Narrowed value | Resolved append signature | Type argument |
| --- | --- | --- | --- | --- |
| core.ts:966 append(xs, ys) | xs: T[] | ys: T | (to: T[], value: T or undefined): T[] | combine's own T |
| core.ts:967 append(ys, xs) | ys: T[] | xs: T | (to: T[], value: T or undefined): T[] | combine's own T |

Before either isArray test the tested argument is T or T[]. Inside the
concatenate arm both xs and ys are T[]. In each append arm the array is T[]
and the scalar is exactly combine's own T. Stock TypeScript already infers T,
not {}. The reported {} widening is not stock TypeScript's inferred type.

The smallest useful type-only edit is to spell the existing type argument:
`append<T>(xs, ys)` and `append<T>(ys, xs)`. This adds six characters total.
The in-memory candidate has the same narrowed types and resolved signatures,
and each array parameter's element type is identical to combine's own T.
No cast or runtime check is needed. This makes the missing type information
explicit for Adamic without claiming to change stock inference.

Raw results: measure-before.json and measure-explicit.json. Commands:

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-y/measure.cjs /workspace/y-before
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-y/measure.cjs /workspace/y-before explicit
```

The explicit experiment substitutes only the implementation owner in a stock
CompilerHost in memory. It never writes the adapted source.
