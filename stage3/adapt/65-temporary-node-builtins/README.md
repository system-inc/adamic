Temporary: comes out when Adamic accepts require with a literal specifier and Node's builtin types

This replaces the previous 65 process-only adaptation. All edits erase: runtime
require calls, their lazy evaluation, exceptions and process tests stay intact.
The local process view names only nextTick and browser, as adaptation 30 already
did; tracing adds pid:number. fs names the six synchronous operations tsc uses.
The perf_hooks assertion names only its optional Performance member. Local
ambient require declarations return explicit any, admitted by current main,
without inventing a runtime loader. This clears checker dependencies, not native
builtin lowering. No tracing or performance fallback is added.

The adapter checks TypeScript 6.0.3, exact source sites, and byte-identical
emitted JavaScript (with and without comments) before writing any file.
The prior process-only evidence below is historical, not validation of this
replacement. New whole-tree apply, oracle and lane results are recorded separately.


## Validated on current main

Against main efe9f404, normal apply succeeds and measures 2 files, 7 lines
added, 2 removed. Core already has the local process view; on older parser
slices this adapter also removes the previous 65 NodeJS.Process dependency.
The previous process-cast-only proof and implementation are archived in evidence.

Both complete stage3 lanes PASS with exactly 106,366 passing, one sanctioned
API failure and zero pending. The only diff is api/typescript.d.ts, identical
to main, with the same 222 composed declaration sanctions. All ten freshly
built JS artifacts compare byte for byte. No API lines are newly sanctioned.
The first run missed writeFileSync in the local fs view; that was fixed and
fully rerun. Worker exits without counts were rerun sequentially with
NODE_OPTIONS=--max-old-space-size=1536; the fixed eight-worker gate is unchanged.

check.cjs verifies idempotence and proves the emitted-JS equality guard fails
for a ! to !! change to tsc's real browser test. The 81-file full-tree and
re-cut slice dumps compare byte for byte: 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The node-end and JSDoc mutants are caught. All 32 local lane tests pass.

The temporary removes checker dependencies only. Native builtin calls are
not replaced by a fallback, and native acceptance remains red at the Error
captureStackTrace checker fact on current main.

Reproduce with NODE_PATH pointing at the pinned stock TypeScript 6.0.3 API:
`node adapt.cjs TREE`, then `node check.cjs TREE`. Run the unmodified full lane
with a new output directory, redirecting output to a log. Compare against a
separate current-main lane. Full commands and reports live in
[the parser evidence](../../drivers/parser/evidence/front31/result.json), including
both complete lane reports, apply tables, oracle logs, runtime JS hashes and mutants.
