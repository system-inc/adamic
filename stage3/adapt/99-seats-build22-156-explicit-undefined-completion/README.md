# Source-map decoder completion

Family 156 on the Nov 22 meter, #bbssdcy. At `src/compiler/sourcemap.ts:562:18`, the `value` property in `decodeMappings.stopIterating` used `undefined!` under a declared `never` slot. Adamic had no supported storage representation for that slot.

The completion result declares `value: undefined`, and the literal writes `value: undefined`. The declaration describes the value Node already places in the own enumerable `value` property. The `done: true` arm and property order are preserved. The adapter locates the nested helper through the TypeScript 6.0.3 syntax tree and checks its exact snapshot.

Adamic's Nov 22 21:49Z ruling applies `docs/rulings.md` lines 54, 55, 76, 291 and 292. Line 292 assigns gaps in Adamic's iterator prelude to the compiler.

At ac44317c, Adamic's ruling counts the iterator and prelude rows as 205, 206, 288, 289 and 290.

## Manual pin

Use an apply output with the pre-156 snapshot ref, or a tree containing the pre-156 source directly.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
  stage3/adapt/99-seats-build22-156-explicit-undefined-completion/test.cjs \
  /tmp/adapted-tsc /tmp/family-156-completion.json
```

The test runs the production adapter on a temporary copy of the source. It transpiles both sides with stock TypeScript 6.0.3 and drives the decoder through done and one more step. For both completions it compares the value descriptor, own-property check, ordered keys and done flag with the base and with the ruled expectation. The JSON path receives the traces. The test exits 1 on a differing row and checks that replay preserves bytes after all rows pass.
