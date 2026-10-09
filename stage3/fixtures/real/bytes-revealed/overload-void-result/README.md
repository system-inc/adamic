# writeTokenText

Source: `src/compiler/emitter.ts:4949`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `emitter.ts:1211`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 9: **79,336 attributed bytes**,
10 deduplicated boundaries; exact reason:
`overload 1 of writeTokenText result void cannot be served by implementation result number` (Refused).

Both overloads and the full implementation are retained. tokenToString is specialized to EqualsToken (=, 64). Driver exercises missing, positive, and negative positions and records all writer calls.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
