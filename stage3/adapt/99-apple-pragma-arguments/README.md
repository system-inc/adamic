# pragma argument maps state their `types` key

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 2 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C13 (G43, 1 site)

`getNamedPragmaArguments` builds `{}` and writes keys from `pragma.args`. addPragmaForMatch checks the selected kind, and `commentPragmas` (types.ts:10273-10316) gives the positional single-line and multi-line pragmas no `types` key; `reference` has one only in its XML kind. So `types?: never` on the private return and `argMap` annotations states absence for every current positional constructor. It rests on the immutable pragma table and the kind check, and Correctness's constructor audit (the scout's evidence/constructors.md and write-audit.json) is its witness, never a string index signature taken as proof.

- `src/compiler/parser.ts`, function `getNamedPragmaArguments`: `{ [index: string]: string | undefined; }` becomes `{ [index: string]: string | undefined; types?: never; }` (2 times).

## C14b (G44, 1 site)

`extractPragmas` builds a fresh argument dictionary. For `reference` and `types` the fixed metadata has `captureSpan: true`, and parser.ts:10714-10720 writes the value, pos and end object; no other XML pragma writes a `types` key. So `types` is declared with that shape, and the index value takes `undefined`, since the scout measured the form without it (C14) at TS2411.

- `src/compiler/parser.ts`, function `extractPragmas`: `const argument: { [index: string]: string | { value: string; pos: number; end: number; }; } = {};` becomes `const argument: { [index: string]: string | { value: string; pos: number; end: number; } | undefined; types?: { value: string; pos: number; end: number; }; } = {};` (1 time).

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.
