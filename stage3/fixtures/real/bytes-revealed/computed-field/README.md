# visitEachChildOfQualifiedName

Source: `src/compiler/visitorPublic.ts:624`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `visitorPublic.ts:623`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 4: **113,387 attributed bytes**,
4 deduplicated boundaries; exact reason:
`a computed field name` (NotYet).

Full table entry and function body. Explicit parameter types replace contextual types. The supplied visitor uppercases identifier text; the factory retains left/right results. QualifiedName uses its TypeScript 6.0.3 kind, 167.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
