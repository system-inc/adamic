# Phantom brand members become void

@system_adamic, October 7, #bw3xg7c: the Path overloads are accepted because their result differs only by a phantom brand. This unit handles the brand annotations; overload compiler work remains with phantom-brands. A brand member read as any would leak any. TypeScript's existing `__escapedIdentifier: void` is the precedent.

This revises adaptation 40's existing 36 any-to-undefined sites to any-to-void. It preserves every key and required/optional flag. There are 31 distinct names in eight compiler files, with 27 public API lines. No runtime member use was found or retained. The one class declaration is `declare`, so it emits no field.

Each name was searched with `rg -n -F <name> <tree>/src/compiler`. Every occurrence was then checked by the stock 6.0.3 parser: it must be a property signature or an uninitialized declare field, never a read, write or test. Exact searches, owners and positions are in [member-audit.json](evidence/brands-void/member-audit.json). The original pinned locations below match class-rules.json; later adaptations can shift line numbers.

| Member | Pinned compiler locations | Reads / writes / tests | Decision |
|---|---|---|---|
| `__incrementalBuildInfoFileIdBrand` | builder.ts:1074 | 0 / 0 / 0 | any to void |
| `__incrementalBuildInfoFileIdListIdBrand` | builder.ts:1076 | 0 / 0 / 0 | any to void |
| `_symbolLinksBrand` | checker.ts:1452, types.ts:6059 | 0 / 0 / 0 | any to void |
| ` __sortedArrayBrand` | corePublic.ts:18, corePublic.ts:22 | 0 / 0 / 0 | any to void |
| `__compilerOptionsKey` | moduleNameResolver.ts:975 | 0 / 0 / 0 | any to void |
| `__modeAwareCacheKey` | moduleNameResolver.ts:1113 | 0 / 0 / 0 | any to void |
| `__pathComponensBrand` | path.ts:457 | 0 / 0 / 0 | any to void |
| `_usingBrand` | transformers/esnext.ts:774 | 0 / 0 / 0 | any to void |
| `__pathBrand` | types.ts:23 | 0 / 0 / 0 | any to void |
| `_jsdocContainerBrand` | types.ts:958 | 0 / 0 / 0 | any to void |
| `_localsContainerBrand` | types.ts:968 | 0 / 0 / 0 | any to void |
| `_flowContainerBrand` | types.ts:974 | 0 / 0 / 0 | any to void |
| `_declarationBrand` | types.ts:1757 | 0 / 0 / 0 | any to void |
| `_autoAccessorBrand` | types.ts:1944 | 0 / 0 / 0 | any to void |
| `_objectLiteralBrand` | types.ts:1981 | 0 / 0 / 0 | any to void |
| `_functionLikeDeclarationBrand` | types.ts:2064 | 0 / 0 / 0 | any to void |
| `_typeNodeBrand` | types.ts:2185 | 0 / 0 / 0 | any to void |
| `_expressionBrand` | types.ts:2397 | 0 / 0 / 0 | any to void |
| `_unaryExpressionBrand` | types.ts:2412 | 0 / 0 / 0 | any to void |
| `_updateExpressionBrand` | types.ts:2418 | 0 / 0 / 0 | any to void |
| `_leftHandSideExpressionBrand` | types.ts:2449 | 0 / 0 / 0 | any to void |
| `_memberExpressionBrand` | types.ts:2453 | 0 / 0 / 0 | any to void |
| `_primaryExpressionBrand` | types.ts:2457 | 0 / 0 / 0 | any to void |
| `_literalExpressionBrand` | types.ts:2785 | 0 / 0 / 0 | any to void |
| `_optionalChainBrand` | types.ts:2982, types.ts:3010, types.ts:3034, types.ts:3168 | 0 / 0 / 0 | any to void |
| `_propertyAccessExpressionLikeQualifiedNameBrand` | types.ts:2997 | 0 / 0 / 0 | any to void |
| `_statementBrand` | types.ts:3322 | 0 / 0 / 0 | any to void |
| `_classElementBrand` | types.ts:3568 | 0 / 0 / 0 | any to void |
| `_typeElementBrand` | types.ts:3573 | 0 / 0 / 0 | any to void |
| `_jsDocTypeBrand` | types.ts:3903 | 0 / 0 / 0 | any to void |
| `_prototypePropertyAssignmentBrand` | utilities.ts:4289 | 0 / 0 / 0 | any to void |

The sorted-array key is literally `" __sortedArrayBrand"`, including its leading space. Both pinned owners have required any markers and no runtime uses; they therefore meet the requested exception and become void. The adapter leaves an already required undefined sorted marker alone, independently checked by the proof. `_optionalChainBrand` has four owners, and `_symbolLinksBrand` has the interface and its implementation; each owner receives its own line edit.

Public sanctions already existed on origin/area/stage3 for these 27 lines as undefined. Only their replacement types and corresponding normalized declaration entries are revised to void; declaration identities and counts remain unchanged. The rule-file proof hash is refreshed. Same-named service markers are left alone because API projection matches the owner as well as the key.

## Reproduce and verification

Source `/workspace/adamic-tools/env.sh` for every command, with Node 24.19.0 first on PATH and `NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules` for stock parser proofs. Before setup, set `GOPROXY='https://proxy.golang.org|direct'`.

Setup passed: Go ready 0.084s, Node ready 0.088s, clang ready 0.521s, markdown dependencies ready 0.935s, submodules ready 22.018s, Go build ready 59.300s, cache warm 59.395s, done 59.427s. `nproc` is 5; CPU quota is 4. Go 1.27.1, clang 20.1.8, Node 24.19.0. Full setup output is `/tmp/any-brands-setup.log`.

Use `brands-void-proof.cjs prepare <before-tree> <proof>` on an isolated fresh tree to restore the brand annotations before building. The before proof uses a complete composed pipeline with only the 36 compiler brand annotations and their 27 owner-specific API-reference lines restored to any. All other adaptations remain. Build that tree with `npm ci` and `npm run build` before taking the snapshot. `brands-void-proof.cjs before <before-tree> <proof>` records all compiler sources, declarations and JavaScript hashes in scratch `state-before.json`, then `brands-void-proof.cjs after <fresh-void-tree> <proof>` compares the fresh pipeline against an independent owner projection. State is kept in scratch because it contains upstream sources. Committed hash lists and per-member evidence are under `evidence/brands-void`.

The brand-only proof passed: all 10 JavaScript files are identical, all 715 declaration files are compared, and only eight change through the reviewed owner projection. All eight touched source files also produce byte-identical standalone JavaScript. The nine negative controls listed in evidence/brands-void/proof.json fail their intended checks; the brand plans are idempotent and the required-undefined sorted control is unchanged.

The premature first oracle build was stopped and discarded after it exposed stale declaration outputs; the brand artifact proof used the separate fresh lane build. The earlier brand-only gate runs were superseded and stopped when the internal isArray site was added. Final combined commands and results are recorded below. Full logs always go to files, never through a pipe.

## Mutants and limits

The real upstream build mutant changes the checker class's `_symbolLinksBrand` to string while its types.ts interface is void. `npm run build` exits 1, with TS2416 at the class member and TS2322 at an existing SymbolLinks assignment. It observes the incompatibility through tsc's own code, rather than an invented assignment. Exact diagnostics and the result are in `evidence/brands-void/build-mutant*`; full output is `/tmp/any-brands-string-mutant.log`. Both mutated source files were restored byte for byte. This proves the pair's assignability check; it does not claim that every unique primitive marker has an assignment observing its value type.

The proof also runs negative controls for a marker read, write, string-key membership test, materialized class field, changed JavaScript bytes, missing JavaScript file, unrelated public API line, and an unreviewed owner type. Each must fail its intended check. Brand plan second passes must change zero bytes.

Native Adamic compilation, the full Go gate, source-map and tsbuildinfo identity, and soundness of remaining upstream assertions are not covered. The overload compiler implementation remains outside this territory. Historical undefined evidence remains intact and is explicitly superseded by this report.


## Internal isArray input

The follow-up changes only `export function isArray(value: any)` to `export function isArray(value: unknown)` at original core.ts:1750. The existing class is Generic array predicate input, ID 58. The predicate remains `value is readonly unknown[]` and its body remains `return Array.isArray(value)`. This follows compiler host-blockers a85a9cb1. Because it is @internal, there is no public line or new sanction.

`isarray-proof.cjs before` takes the brand-only composed tree and resolves every internal isArray call through the stock checker, rather than counting all same-named methods. `after` checks the final composed tree. **64 compiler caller sites are unchanged; zero diagnostics before and after.** Every compiler source is compared, with only the owning annotation permitted to differ. All ten emitted JavaScript files and the public typescript.d.ts are byte-identical. Stock standalone emission of core.ts also matches exactly. The unknown-input program has zero compiler diagnostics across the complete src/compiler project, not only at a filtered caller subset. See [isarray/proof.json](evidence/isarray/proof.json), the caller list and JavaScript hashes beside it.

An input-type mutant changes unknown to string using the stock compiler host. Existing compiler callers then fail with TS2345, with all 56 new TS2345 diagnostics retained in the proof JSON. A predicate-body mutant changes the real core.ts body to `return false`; running the actual adapter rejects it as `unreviewed uses: isArray`. No mutant is retained in the production tree.

This site removes one more any token: adaptation 40 has now removed 69 of the original 210, leaving 141. The earlier 36-brand count is unchanged by revising undefined to void.

## Final combined gates

With Node 24.19.0 first on PATH, these commands were run with both output streams redirected to logs:

```sh
bash stage3/apply.sh /tmp/any-brands-isarray-final > /tmp/any-brands-isarray-apply.log 2>&1
bash stage3/oracle/run.sh /tmp/any-brands-isarray-final /tmp/any-brands-isarray-oracle > /tmp/any-brands-isarray-oracle.log 2>&1
bash stage3/lane/run.sh /tmp/any-brands-isarray-lane > /tmp/any-brands-isarray-lane.log 2>&1
```

Apply exits 0. The standalone full oracle installs and builds successfully, then reports **106,366 passing, one failing, zero pending**, with only `api/typescript.d.ts` differing. Its exit 1 is the expected public API acknowledgement failure. Total oracle wall time is 403.217s. The independent lane checker accepts this completed result: all 222 composed declarations are sanctioned, including adaptation 40's 28 reference declarations (27 brand revisions and the existing ErrorCallback line). No sanction is added for isArray. Reports are retained under `evidence/gates`.

An additional lane mutant changes only Path's sanctioned replacement from VoidKeyword to StringKeyword in a scratch sanction file. The real lane checker exits 1 and reports both an unsanctioned Path reference and a changed sanctioned Path declaration. The original sanction file is untouched; the normal checker is rerun and passes. See `evidence/brands-void/api-sanction-mutant.json`.

The independent full `lane/run.sh` completes with exit 0 and **PASS stage3 landing lane**, no errors: 106,366 passing, one sanctioned API failure, zero pending. Its fresh install and build exit 0; its oracle takes 327.995s. The lane records 769.28s including the temporary pause while the standalone oracle and proof finished. Its API result is exactly 28 adaptation reference declarations and 222 composed/sanctioned declarations, with only the 27 reviewed brand replacements revised by this unit. See `evidence/gates/lane.json` and `lane-oracle.json`.
