# transformParameterWithPropertyAssignment

Source: `src/compiler/transformers/ts.ts:1430`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/ts.ts:235`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 11: **61,355 attributed bytes**,
3 deduplicated boundaries; exact reason:
`a value of type ParameterPropertyDeclaration` (NotYet).

Full body. The intersection preserves a constructor parent and identifier name. Factory helpers record this.x=x; as text; clone preserves identifiers, and range/parent/emit metadata does not affect this output.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
