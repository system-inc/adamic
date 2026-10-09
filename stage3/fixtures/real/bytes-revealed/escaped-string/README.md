# createBaseIdentifier

Source: `src/compiler/factory/nodeFactory.ts:1304`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `factory/nodeFactory.ts:492`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 1: **273,934 attributed bytes**,
377 deduplicated boundaries; exact reason:
`a value of type __String` (NotYet).

Full body. InternalSymbolName is restricted to the two real members used here. The base allocator supplies the four fields the body writes; its initialization and SyntaxKind identifier value are retained.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
