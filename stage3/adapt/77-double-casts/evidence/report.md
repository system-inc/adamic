# Observed adaptation 77 results

Partial implementation: 18 sites reviewed, four bridges replaced by explicit single downcasts, fourteen left unresolved. README and sites.json list each outcome and contract limitation. No public declaration changed. No native checked-cast proof is claimed.

Base `031a1259bc7973934792dc6cb1bd4074fc2204b9`. The measured tree was the uncommitted adaptation on that base; the execution records therefore identify the base commit. provenance.json records exact adapter/proof/site hashes.

## Commands and results

The README records the exact setup, lane and proof commands. Main lane `/tmp/adapt77-before` and final adapted measurement `/tmp/adapt77-accepted` both have lane status pass. The main lane exits 0. The recovered run applied successfully but hit a hook timeout in its eight-worker oracle. The final run retains that exact apply output, runs the full oracle at eight workers, and passes the unchanged lane/check.py. Both have 106366 passing, 1 failing, 0 pending tests. The sole failing test is the sanctioned public API acknowledgment: `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`. This is identical to main, not an assertion that the raw TypeScript suite has zero failures.

Proof exits 0: 711 source files compared; only the four documented type expressions differ. All 10 real built JavaScript files and 2 API artifact files have identical SHA256/bytes. Lane verdict, counts, failure identity, API fields and baseline.diff bytes match main. Stock compiler checking reports zero diagnostics on both actual source trees. Adapter idempotence and LF reconstruction pass. No whole Go gate or native tsc run was performed.

The first adapted lane was interrupted by an environment restart that killed its coordinator. Its raw tests log finished but no lane report was produced; it is not used as proof. The recovered eight-worker lane then failed on a 40-second before-all timeout in declarationEmitPrivatePromiseLikeInterface.ts (106360 passing, 2 failing); its verdict is archived. The exact case passes six checks in seven seconds alone. A four-worker full oracle control then matched all 106366/1/0 counts and baseline differences, but the unchanged lane correctly rejected workers=4 (expected 8). That control is not an accepted lane verdict. The accepted final full oracle uses eight workers on the same successful apply tree and passes the unchanged lane checker. Additional runs address observed interruption/timeout, not repeat confirmation.

## Mutants

All 23 mutants are caught. Source guard mutants deliberately prove review drift detection, including unresolved sites; they do not prove native type soundness. Each substitutes a string assertion into that site's real source expression and exits 1 before writes. Duplicate expressions are checked by their group multiplicity.

| Site | Scout location | Exit | Catch |
|---|---|---|---|
| C01 | src/compiler/checker.ts:10950:32 | 1 | reviewed AST site multiplicity guard |
| C02 | src/compiler/checker.ts:49445:40 | 1 | reviewed AST site multiplicity guard |
| C03 | src/compiler/core.ts:736:36 | 1 | reviewed AST site multiplicity guard |
| C04 | src/compiler/core.ts:759:12 | 1 | reviewed AST site multiplicity guard |
| C05 | src/compiler/core.ts:764:12 | 1 | reviewed AST site multiplicity guard |
| C06 | src/compiler/core.ts:810:89 | 1 | reviewed AST site multiplicity guard |
| C07 | src/compiler/debug.ts:849:36 | 1 | reviewed AST site multiplicity guard |
| C08 | src/compiler/debug.ts:850:8 | 1 | reviewed AST site multiplicity guard |
| C09 | src/compiler/program.ts:3284:116 | 1 | reviewed AST site multiplicity guard |
| C10 | src/compiler/transformer.ts:338:108 | 1 | reviewed AST site multiplicity guard |
| C11 | src/compiler/tsbuildPublic.ts:992:23 | 1 | reviewed AST site multiplicity guard |
| C12 | src/compiler/tsbuildPublic.ts:993:22 | 1 | reviewed AST site multiplicity guard |
| C13 | src/compiler/tsbuildPublic.ts:1355:12 | 1 | reviewed AST site multiplicity guard |
| C14 | src/compiler/watch.ts:862:41 | 1 | reviewed AST site multiplicity guard |
| C15 | src/compiler/watchPublic.ts:150:38 | 1 | reviewed AST site multiplicity guard |
| C16 | src/compiler/watchPublic.ts:151:24 | 1 | reviewed AST site multiplicity guard |
| C17 | src/compiler/watchPublic.ts:555:22 | 1 | reviewed AST site multiplicity guard |
| C18 | src/compiler/watchUtilities.ts:839:13 | 1 | reviewed AST site multiplicity guard |

Two real artifact mutants append a byte to built JavaScript and typescript.d.ts; SHA256 equality rejects both. One lane-count mutant decrements the measured passing count; verdict comparison rejects it. One actual transformer source mutant replaces the SourceFile intersection with string; stock checking produces TS2339/TS2352. One real built debugger mutant changes m1 to m0; exact successful composite/merged mapper output rejects it. proof.json records diagnostics, artifact hashes and all outcomes.

The debugger control covers both nested mapper casts with prototypes installed, and separately observes TypeError for a child without the debug prototype. This establishes the Node behavior and its limit, not unconditional native admission of the class view.

## Setup and limits

GOPROXY was set to `https://proxy.golang.org|direct` before setup. Setup exit 0; nproc 5. Readiness lines: Go 0.033s, Node 0.039s, submodules 0.103s, markdown 0.107s, clang 0.208s, Go build 71.535s, test binaries deferred 71.732s, cache warm 71.734s, total 71.778s. Actual env source was /workspace/adamic-tools/env.sh. Complete setup.log and setup.json are retained.

Fourteen unresolved bridges remain: C01, C03-C06, C09, C11-C18. Their producer, phantom brand, comparator, builder-generic and enum contracts are unfinished. Four single downcasts still require checked-cast work #b5w3ycg. No .a fixture was added; a-check headers and counts.md need no update. No compiler files or public declarations were edited, and no additional API sanction was introduced.
