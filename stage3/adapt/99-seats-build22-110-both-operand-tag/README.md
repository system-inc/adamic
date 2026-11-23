# Compare both JSX tag kinds before the property access

Family 110 (one site), lane build22-110 on #bbssdcy. Compiler accepts under
the split (Adamic, Nov 22, 2026). This applies the cast and checked-view
rulings in `docs/rulings.md` rows 13, 15, 18 and 73 at Adamic
`0efbdb083a60086fec3a29687a162bbf21105336`.

Reviewer lane review-j measured a scoped clear in the compiler-directory
scan and selected attempt. The two enum-slot records for #snypn8c at
`parser.ts:10804:68` and `parser.ts:10805:102` came from the right-hand
`rhs as PropertyAccessExpression` casts. The two outer recursive cast
records at `parser.ts:10805:31` and `parser.ts:10805:101` also disappeared.
The selected compiler-directory unit cleared with zero findings. The
entry-root replay retained the existing missing import-declaration route
panic and its boundary. Loom owns the whole-meter disposition.

The site is the closing return of `tagNamesAreEquivalent` in
`src/compiler/parser.ts` (`parser.ts:10804:68` in the scout census).
The adapter finds that declaration and its final return through the syntax
tree, and checks the exact function text. It accepts the original and
adapted snapshots and writes only the return.

The preceding statements establish equal kinds and handle the other three
members of `JsxTagNameExpression`.

```ts
    if (lhs.kind !== rhs.kind) {
        return false;
    }

    if (lhs.kind === SyntaxKind.Identifier) {
        return lhs.escapedText === (rhs as Identifier).escapedText;
    }

    if (lhs.kind === SyntaxKind.ThisKeyword) {
        return true;
    }

    if (lhs.kind === SyntaxKind.JsxNamespacedName) {
        return lhs.namespace.escapedText === (rhs as JsxNamespacedName).namespace.escapedText &&
            lhs.name.escapedText === (rhs as JsxNamespacedName).name.escapedText;
    }
```

For tsc's own ordinary parser nodes, `kind` and `name` are stable,
side-effect-free data fields. At the final return, the left tag must be
`SyntaxKind.PropertyAccessExpression`. The first return establishes that
the right tag has the same kind. Both comparisons are true whenever
execution reaches the return. The guard narrows both operands to
`JsxTagNamePropertyAccess`, whose `expression` is itself a JSX tag name.
The existing name comparison still short-circuits the recursive call, and
the call receives the same expressions. These parser nodes give the same
result with either return.

Review-j's getter witness is outside that guarantee. Its right-hand
`kind` getter always returns `PropertyAccessExpression` and changes the
name exposed by `name` on its second read. Base returns true and the
guarded return gives false. A getter that changes the discriminant also
gives different results. Adamic's decision on externally supplied tags is
pending.

Run the adapter regression on the pristine TypeScript v6.0.3 checkout.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-110-both-operand-tag/test.cjs /path/to/TypeScript
```

The test runs the production adapter on a scratch copy of `parser.ts`,
checks ordinary JSX parser nodes including `a.b.c` against `z.b.c`, and
preserves both getter witnesses.
