# 99m: fix flatMap's append specialization

The confirmed site is adapted `src/compiler/core.ts:409:37`: `result`, the
first argument of `append(result, v)` in `flatMap`. The `addRange(result, v)`
call is at line 406. The sole edit is:

```ts
result = append<U>(result, v);
```

Stock TypeScript 6.0.3 infers `U` at both calls. At append it selects
`(to: U[] | undefined, value: U): U[]`; the array arm is `readonly U[]`,
and the else arm is `U`. It does not report the suspected `U & any[]` arm.
Adamic nevertheless reports the writable `{}[] | undefined` view at the
append argument. Making the existing element type explicit removes that
view in the ordinary compiler. This changes no declaration or runtime code
and introduces no cast. `site-final-before.json` and
`site-explicit-append.json` retain the stock-checker measurements.

The unique top-level flatMap owner is found by its parsed name and pinned
by complete exact text and owner-scoped SHA-256 before and after. Changed,
missing, duplicate, nested, and partly adapted owners are refused before
writing. The source is reread before writing; unrelated text and LF/CRLF
are preserved. The complete core.ts JavaScript must match, with comments
both retained and removed, before writing. A second run makes zero edits
and writes and preserves bytes and mtime. `javascriptChanging` is false.

`sort.json` records the real Python directory sort: 99m is immediately
after 99l and before 99-wasm-crypto-import. No existing adapter changes.

The final 79-root census and exact next stops are in [CENSUS.md](CENSUS.md).
No second edit was selected: the next shared stop is the inferred helper
return reaching deduplicate's declared return at core.ts:726:12. It needs
its own adaptation analysis. The core.ts root instead reaches an unchecked
cast in debug.ts:167:31. Only core.ts is adapted here.

See [PROOF.md](PROOF.md) for proof commands and results. With NODE_PATH set
to stage3/api/node_modules, the retained helpers can be run as follows:

```sh
ADAMIC_OPTIONS=1 node stage3/adapt/99-correctness-axolotl-m/inspect.cjs BEFORE_TREE
ADAMIC_OPTIONS=1 node stage3/adapt/99-correctness-axolotl-m/inspect.cjs BEFORE_TREE explicit
node stage3/adapt/99-correctness-axolotl-m/contract.cjs BEFORE_TREE
node stage3/adapt/99-correctness-axolotl-m/compare.cjs BEFORE_TREE AFTER_TREE REPLAY_TREE CENSUS_TREE
```

The stage3 lane Go check remains failed on this worker: two existing API-generator
wrappers hit fixed 55-second deadlines. The unchanged underlying API test module
passes when run directly. This limitation and every attempt are retained in PROOF.md
and the logs; no deadline, test selection, or checker setting was weakened.
