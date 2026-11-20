# Type parameter append and token source-map ranges (#dt55mtb)

This type-only adaptation pins both function declarations through the stock
TypeScript 6.0.3 AST. It validates both owners before writing, preserves the
source line endings, rejects partially adapted input, and writes nothing on
replay.

- `checker.ts:19948:37` in `getInferTypeParameters`. The
  `SymbolFlags.TypeParameter` guard selects the type parameter symbols in the
  conditional node's locals. The declared-type dispatcher resolves these to
  their TypeParameter. Keep that dispatcher, use the compiler's existing
  `as TypeParameter` spelling, and call `append<TypeParameter>` so the
  accumulator's writable element type stays narrow.
- `factory/nodeFactory.ts:7490:71` and `:93` in the sole
  `mergeTokenSourceMapRanges` call. Both arguments are
  `EmitNode.tokenSourceMapRanges`, whose defined type is
  `(SourceMapRange | undefined)[]`. The helper copies source entries into
  the destination and returns it, allocating an empty destination if
  necessary. Both parameters and the result use that same array type.
  The type-only import is local to the pinned signature. Adaptation 30
  checks the two indexed expressions and declines both edits.
