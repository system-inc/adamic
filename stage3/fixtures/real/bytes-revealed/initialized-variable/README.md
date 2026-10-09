# transformInitializedVariable

Source: `src/compiler/transformers/module/module.ts:1926`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/module/module.ts:174`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 3: **120,003 attributed bytes**,
5 deduplicated boundaries; exact reason:
`a value of type InitializedVariableDeclaration` (NotYet).

Full body. This driver selects the plain identifier branch. Visitors preserve the input, range metadata is ignored, and the reduced expression factory records the emitted assignment as text. Binding-pattern flattening is outside this driver.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
