# transformPrivateIdentifierInInExpression

Source: `src/compiler/transformers/classFields.ts:671`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/classFields.ts:354`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 6: **97,864 attributed bytes**,
3 deduplicated boundaries; exact reason:
`a value of type PrivateIdentifierInExpression` (NotYet).

Full body. Both resolved and unresolved private names are exercised. The helper records its brand/receiver arguments as emitted text; the fallback visitor retains the node. The original intersection includes the phantom token: InKeyword field, separately from BinaryExpression.operatorToken.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
