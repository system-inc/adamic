# visitBindingElement

Source: `src/compiler/transformers/declarations.ts:619`, TypeScript 6.0.3 after the
registered `ed6e2975` adaptations. Ranking example: `transformers/declarations.ts:259`.
An enclosing stopped body can be the example; the actual diagnostic identifies
the extracted function inside it.

Rank 10: **63,225 attributed bytes**,
4 deduplicated boundaries; exact reason:
`overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem` (Refused).

Overload and full implementation are retained. Reduced AST types preserve the broad Node constraint and distinct binding/omitted variants. The driver exercises initializer removal and the omitted-element identity branch. No computed property is supplied; visibility hooks are outside these inputs.

The source-Node golden is [expected.stdout](expected.stdout).
[The manifest](../status.json) records its bytes, exit, and the complete
`ed6e2975` diagnostic. Supporting AST views are reduced to the fields used
by this extract; this is not a full tsc integration test.

The first-line a-check header is for the topic compiler; the full negative
golden is pinned to ed6e2975. See [the two-pin explanation](../README.md).
