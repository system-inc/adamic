# Proof

Measured on the adapted tree stage3/apply.sh builds at main 932a6344af, with this adaptation run last over it (its
members are in types.ts and resolutionCache.ts, which 99-wasm-crypto-import, the one directory sorting after it, doesn't
touch). Every run was a sealed Loom proof job on a rented machine. The scripts are on
`runtime-performance/method-values` under measure/rv75m9k/: scan.cjs finds every method read as a value with the stock
checker, owners.cjs pins the members, gate.sh runs the checks below before and after, and compare.py pairs the censuses.

## Sites

The meter's run 20261116T134940Z.SUGhBh counts the family at 140 sites over 71 names. On the adapted tree, the stock
checker finds 129 method value reads under those names (J-0wk70y8). This adaptation takes 80 of them, and 49 remain:
5 on the written members it leaves as methods, and 44 on public members, which are road B's (a later adaptation, never
a declaration move). The 32 members, before and after, are owners.json.

## Checks, before and after

| Check | Result | Job |
|---|---|---|
| Stock tsc 6.0.3 over the 79 roots, strictFunctionTypes on | 0 diagnostics before, 0 after | J-0wk70y8, J-tb5tn3f |
| `adamic types` over the 79 roots | exit 0, 0 diagnostics before and after; the listing gains the 32 members as declarations | J-tb5tn3f |
| Emitted JavaScript of each touched file (transpileModule, ES2020, CommonJS) | byte-identical | J-tb5tn3f |
| Declarations of the 79 roots, emitted with stripInternal | all 79 byte-identical (maps left out: they record source columns) | J-tb5tn3f |
| Replay | writes nothing | J-0wk70y8 |
| Restoring mutant | every touched file back byte for byte | J-0wk70y8 |

## Full single-file censuses

Each run loads all 79 roots, CommonJSBundle, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE, with the
census tool built from main's make_overlay.py. Findings pair as a multiset by (kind, phase, reason) with positions
stripped. Every run exited 0 with checker_rejected false and a relation trailer.

| File | Before | After | Removed | Added | Job |
|---|---:|---:|---:|---:|---|
| checker.ts | 8,548 | 8,546 | 2 | 0 | J-g8g8afm |
| nodeFactory.ts | 3,677 | 3,677 | 0 | 0 | J-tb5tn3f |
| program.ts | 1,500 | 1,500 | 0 | 0 | J-tb5tn3f |
| types.ts | 0 | 0 | 0 | 0 | J-tb5tn3f |
| emitter.ts | 1,112 | 1,112 | 0 | 0 | J-x834psv |
| transformers/es2015.ts | 1,113 | 1,113 | 0 | 0 | J-2m5tpex |
| transformers/module/module.ts | 398 | 398 | 0 | 0 | J-2m5tpex |
| transformers/module/system.ts | 411 | 411 | 0 | 0 | J-2m5tpex |
| transformers/es2018.ts | 496 | 496 | 0 | 0 | J-2m5tpex |
| transformers/generators.ts | 523 | 523 | 0 | 0 | J-2m5tpex |
| transformers/declarations.ts | 262 | 262 | 0 | 0 | J-2m5tpex |
| resolutionCache.ts | 211 | 210 | 1 | 0 | J-e2xxmxe |
| factory/parenthesizerRules.ts | 181 | 181 | 0 | 0 | J-e2xxmxe |
| factory/nodeConverters.ts | 151 | 151 | 0 | 0 | J-e2xxmxe |
| sourcemap.ts | 132 | 132 | 0 | 0 | J-e2xxmxe |
| visitorPublic.ts | 95 | 95 | 0 | 0 | J-e2xxmxe |
| transformers/destructuring.ts | 78 | 78 | 0 | 0 | J-e2xxmxe |

**0 new anywhere.** Removed: resolutionCache.ts:556:32, the target refusal for globalCacheResolutionModuleName; and in
checker.ts at 8074, a lowering NotYet, "a presence test of a method whose presence the receiver's construction doesn't
prove", with its rollback Boundary: the member is a field now, so its presence test lowers.

**What the censuses can't show:** most of the 80 reads sit in units that stop earlier (the printer's `var` gate in
emitter.ts, earlier stops in the transformers and visitorPublic.ts), so a census of those files doesn't reach them. That
they're gone is the stock checker's answer: scan.cjs over the adapted tree reads 129 before and 49 after, with every
road-A read gone (J-0wk70y8). The meter will say which become visible as earlier stops fall.

## Not run

- `hereby dts` on the built upstream tree. The stripInternal declaration emit of the 79 roots stands in for it here, and
  every member is `@internal`; Correctness, say if you want hereby's own bytes.
- The whole lane. The emitted JavaScript is byte-identical, so the lane runs the same code; say if you want it run.
