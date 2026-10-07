# 33: Required indexed reads in the emit partition

Pinned TypeScript 6.0.3 source receives per-occurrence required-read assertions at the original evaluation point. Node erases each assertion, while Adamic treats it as a loud runtime check. No undefined or null initializer is asserted. An honest unresolved invariant is recorded as a decline.

This adapter follows adaptation 30's parsed-expression plus occurrence ledger. Stock npm TypeScript 6.0.3 parses current text, validates every occurrence count and existing assertion/default shape, and plans every file before any write. It fails on drift. Insertions preserve CRLF. There are no source regex rewrites, skips, optional chains, hoists, or general defaults. Only numeric bitwise U-zero sites may use `?? 0`.

[sites.json](sites.json) records each reviewed occurrence, its class, action, and one-line invariant or decline reason. Coordinates are documentation, never edit offsets. Pure stores and intentionally optional reads are declined. Existing upstream assertions need no adaptation. The ledger grows in three pushed waves:

- [Wave A](proof/a/README.md): emitter.ts and sourcemap.ts.
- Wave B: transformers/ts.ts, es2015.ts and declarations.ts.
- Wave C: transformer.ts, visitorPublic.ts and remaining transformers recursively.

Run `node adapt.cjs <tree>` with `NODE_PATH` pointing at stock typescript@6.0.3. `node adapt.cjs --check <tree>` validates contracts without writing. `node verify.cjs <before> <after>` compares emitted JavaScript, verifies site contracts and checks idempotence. `node mutant.cjs <before> <after> <new-mutant-tree>` proves the JavaScript and contract checks can fail. `bash census.sh <tree> <output>` runs the unchanged Adamic loader through a temporary Go overlay and retains full diagnostic chains for TS2345, TS18048, TS2532, TS2322 and TS2538.

Proof trees contain 00, 10, 30 and 33, leaving 20 out. No TypeScript source is committed. The stock oracle builds upstream and runs the default suites against its checked-in baselines. Per-wave proof directories retain census tables, every remaining finding with its reason, oracle reports and baseline diff, idempotence, JavaScript verification and mutant results.

Preparation: `bash cloud/setup.sh` took 113 seconds (Go 0, clang 0, Node 0, submodules 0, buildcache 113). `nproc` returned 5; oracle uses its default 4 workers. Dependencies: type imports a3ef0dc and indexed reads 07f637c, merged into origin/main e011f8f. All repository additions are confined to this directory.
