# Append the declared infer type parameter (#dt55mtb)

Inside `getInferTypeParameters`, the TypeParameter flag guard selects symbols
from the conditional node's locals. The append calls
`getDeclaredTypeOfTypeParameter(symbol)` directly and infers TypeParameter.

The stock TypeScript 6.0.3 AST locates the exact function before and after
the edit. The adapter preserves all surrounding bytes and makes no write on
replay. The emitted callee changes. See the evidence branch for the dispatch
proof and whole-oracle object identity count.
