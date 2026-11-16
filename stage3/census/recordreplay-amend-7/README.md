# Recordreplay amendment 7 proof

Base amendment: `163bdeafe750d70b2d6f2540eba71ef1f83dffca`. Parent: `66fc7356a943372caa1b494d8d05959f91e0fb7b`.

The adapter restores exactly ff0e5859's declaration bytes and erased import. The AST lookup, composed adapter edits, owner text and owner-pin file remain unchanged. No pins were regenerated because the restore is outside the pinned owner and node: import block.

| Proof | Exit |
|---|---|
| Actual TestProduct_Stage3AdaptedTree, hereby local + tests | 0 |
| Same test, file physically removed before hereby local | 1 |
| Fresh stage3/apply.sh | 0 |
| Recordreplay check: real Adamic loader + independent stock checks | 0 |
| Recordreplay AST/replay contract | 0 |
| Full sys.ts census: CLI-only control | 0 |
| Full sys.ts census: restored amendment | 0 |
| Census comparison: zero newly stopped units | 0 |
| Source/transpile scope | 0 |
| Require-result contract | 0 |
| Watch-directory contract | 0 |
| Declaration-drift, owner-drift, out-of-scope mutants | 1 each, expected |

The actual command was `go test -run '^TestProduct_Stage3AdaptedTree$' -v -timeout 45m ./stage3/lane/adaptedtree`. It builds through hereby local and hereby tests. Its cold-cache negative control uses the same test with the saved Go overlay; that overlay changes only the build recipe to delete src/recordreplay.d.ts after adaptation and before hereby local. The negative run fails in hereby local with TS2339 (and TS2307 for the now-missing erased import target). Production Go code is unchanged. The two runs use separate cold product caches.

The product's locked bootstrap TypeScript compiler is 5.9.3, with @types/node 25.3.3. The separate API stock checker is unmodified TypeScript 6.0.3 with @types/node 25.3.3; it loads the real imported declaration, injects no augmentation, and reports zero diagnostics with the file versus TS2339 without it. The actual Adamic loader also exits 0 on all 79 roots. This distinguishes the two environments that the earlier proof conflated.

Both full sys.ts streams completed in CommonJSBundle with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1. Coverage is 135 → 136 units: every original unit remains, and the restored erased import adds one successful unit. Findings are 155 → 194; all 44 new finding records belong to already-stopped units. **Zero units newly stop; the recordreplay as-any read is gone.** The full classifications are in census-comparison.json.

All 748 fresh source files match the composed input used for the checks and census. Relative to the line amendment's adapted output, only the erased import and the declaration are restored. The CLI-only control and reused compiler binaries have identical compiler sources to 66fc7356; the complete CLI tree matches the prior re-cut's tree. Binary hashes, source pins and compiler versions are recorded in provenance.json.

Raw logs and census streams are gzip-compressed. The physical-removal overlay is mutant-adaptedtree.go.gz; decompress it before using mutant-overlay.json. The file is proof instrumentation and is not part of the production Go package. Paths in the scripts name this run's scratch workspace. The nested stock-check.json and adamic-types.log.gz come from the recordreplay check itself.

No Loom submission or PR was made.
