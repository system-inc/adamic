Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 14 C, family 12), on the lane witness this head carries; Adamic accepts the head after its review and census.

The scout reported the invariant mutable refusal at `factory/utilities.ts:219:23` in `createJsxFactoryExpressionFromEntityName`. The qualified name arm creates an identifier from `idText(jsxFactory.right)` before assigning the original escaped text. This adaptation removes the cast to `Mutable<Identifier>` and that assignment.

The checker parses `jsxFactory` and `jsxFragmentFactory` options and the `@jsx` and `@jsxFrag` pragmas with `parseIsolatedEntityName`. The factory fallback synthesizes a qualified name with the fixed right identifier `createElement`. A missing factory at the emit helper bypasses the assignment. Parser identifiers and the synthesized fallback use `createIdentifier`, which escapes the text at construction. Unescaping and escaping again preserves each right identifier reaching the assignment.

The lane Node witness compared each incoming escaped value with the new identifier's value using `===`. Every reach matched. Emitted JavaScript and diagnostics matched for all 26 JSX programs across the stock and edited compiler copies. The corpus includes two and three leading underscores, a Unicode escape, keywords, and the synthesized default and custom namespace fallback.

Rulings [96](../../../docs/rulings.md#L96) and [265](../../../docs/rulings.md#L265) apply to the refusal of writable views of readonly fields and unproven writes through `Mutable<T>` respectively.

Run `test.cjs` with the locked API packages on the tree produced by `stage3/apply.sh`. It reads the adapter's own before and after snapshot refs and compares the name chain, printer output, and factory call sequence.
