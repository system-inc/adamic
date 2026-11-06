# 99y: make combine's append type argument explicit

The only source edits are append<T>(xs, ys) and append<T>(ys, xs).
MEASURE.md records the stock checker evidence collected before editing.
Stock TypeScript already infers combine's T. The annotation makes that
information explicit for Adamic. There is no cast or runtime change.

Run with stock typescript 6.0.3 on NODE_PATH: node adapt.cjs TREE.
All four combine declarations are pinned by exact text and owner-scoped
SHA-256 before and after. The pinned source has three overload signatures
plus the implementation, including the array-only overload above the excerpt.
Missing, duplicate, nested, changed, or partially adapted owners fail before
any write. Replay preserves source bytes and mtime. Complete core.ts JavaScript
must match with comments both kept and removed.

Python's real sort puts 99y after all other correctness adapters and before
99-wasm-crypto-import, matching 99w's convention. Running last overall is
incompatible with the required directory name and unchanged apply.py.
