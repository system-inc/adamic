# transformClassLike

Source: `src/compiler/transformers/esDecorators.ts:670`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/esDecorators.ts:295`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 7: **93,640 attributed bytes**,
2 deduplicated boundaries; exact reason:
`a function returning ImmediatelyInvokedArrowFunction` (NotYet).

Trimmed to the undecorated, named class with no members, heritage, or lexical additions. Retains the class expression, return statement, original-node association, lexical merge, and final IIFE construction from the original body. Helpers represent the same AST shape and print generated expression text; the generated class expression is not executed.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
