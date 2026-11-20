# Token source-map ranges (#dt55mtb)

The single caller of `mergeTokenSourceMapRanges` passes two
`EmitNode.tokenSourceMapRanges` arrays. The helper copies the source entries
into the destination and returns that array. Its parameters and return now
use `(SourceMapRange | undefined)[]`, the declared element type of both arrays.
The existing named import from the compiler namespace supplies a type-only
`SourceMapRange` specifier next to `SourceMapSource`.

The stock TypeScript 6.0.3 AST identifies and pins the helper. The adapter checks
the import and signature together and refuses partial input. Adaptation 30
records two declined indexed accesses and checks their exact expressions.
Neither those expressions nor their occurrence counts change. The affected
refusals are at `factory/nodeFactory.ts:7490:45`, `:71`, and `:93` in the
baseline. This is type-only.
