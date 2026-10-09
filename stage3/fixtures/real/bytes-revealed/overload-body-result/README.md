# transformFunctionBody

Source: `src/compiler/transformers/es2018.ts:1220`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/es2018.ts:154`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 13: **42,328 attributed bytes**,
4 deduplicated boundaries; exact reason:
`overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody` (Refused).

Both overloads and full implementation are retained. Drivers exercise a block with no leading declarations and an expression body converted to a block with one lexical declaration. copyPrologue returns zero for the supplied return statement, and no object-rest parameter assignments exist. Range metadata is ignored.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
