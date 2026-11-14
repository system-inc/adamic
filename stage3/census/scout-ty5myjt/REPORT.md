# Scout ty5myjt — stock results; full census pending

Recommended direction: a **named readonly view** for the diagnostics helper's
`file` constraint and for each resolution helper's `resolvedFileName` constraint.
These are provisional directions, not measured full-census recommendations.
Both named-view scratch shapes have **0 stock tsc 6.0.3 diagnostics across 79 roots,
0 new**, and byte-identical JavaScript. Full-census deltas are pending; no claim
that the six findings are gone is made at this checkpoint.

| Candidate | Stock diagnostics | New vs baseline | JavaScript diff |
| --- | ---: | ---: | --- |
| Baseline | 0 | — | — |
| Diagnostic literal readonly intersection (control) | 0 | 0 | byte-identical |
| Diagnostic named readonly view | 0 | 0 | byte-identical |
| Diagnostic exact DiagnosticWithLocation (negative control) | 1 | 1 | byte-identical |
| Diagnostic non-generic overloads for both caller types | 0 | 0 | byte-identical |
| Resolution named readonly view | 0 | 0 | byte-identical |
| Resolution exact ResolvedModuleFull result (negative control) | 2 | 2 | byte-identical |
| Resolution fixed readonly callback result, no R binder | 0 | 0 | byte-identical |

Stock options and every diagnostic are in `*-stock.json`. The compiler roots are
exactly the same 79 source files; Node ambient roots come from the worktree's
locked @types/node plus the existing census node-require declaration. Stock
libraries are unchanged. An initial ambient-free control had 64 diagnostics;
the complete-input baseline has zero. All candidate source patches are included.
`javascript.json` records both source files' emitted sizes and SHA256s: **14/14
comparisons identical**, stock 6.0.3 transpileModule, ES2024/CommonJS, comments kept.
The exact diagnostic specialization adds TS2322 at the semantic caller. The exact
resolution result specialization adds TS2322/TS2345 at type-reference callers.
The resolution controls retain T (cache-wrapper polymorphism); the exact-result
control removes R, and is not a claim of fully monomorphizing the cache wrappers.

Full helper signatures and bodies: [BODIES.md](BODIES.md).
Every src caller and T/R, harness included: [CALLERS.md](CALLERS.md),
[call-inventory.json](call-inventory.json), [harness-call.json](harness-call.json).

Base: `7534c1d7575db02bb7b6512119e0685987b7e333`, detached fresh worktree
`/workspace/scratch/scout-ty5myjt/adamic` before creating this report branch.
Shared checkout is unchanged apart from fetched Git objects/worktree registration.
Recursive shallow submodule initialization succeeded, including cohere at
`5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript at
`d92d9bfee114c80be2c375d72edae966176e3a4f`; no missing credentials.

Stock TypeScript was cloned using `--depth 1 --branch v6.0.3` from Microsoft and
confirmed at `050880ce59e30b356b686bd3144efe24f875ebc8`.
`npm ci` in this worktree's stage3/api succeeded (3 packages).
`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/scout-ty5myjt/TypeScript` is supplied
to apply.sh; adapted output and all build products are in the same scratch area.

The initial census overlay build encountered `go: creating work dir: stat
/tmp/adamic-gate: no such file or directory`; after setting GOTMPDIR, cgo also
reported `cgo: open /tmp/adamic-gate/cgo-gcc-input-1712609632: no such file or directory`.
The retry sets GOTMPDIR and TMPDIR to scratch and CGO_ENABLED=0.
No adaptation or compiler source has been edited.

## Early source evidence (measurements still pending)

`getDiagnosticsHelper` has no direct `file` assignment or field access. It calls
`getDiagnostics`, `flatMap`, and `sortAndDeduplicateDiagnostics`. The latter passes
diagnostics to `sortAndDeduplicate` with `compareDiagnostics` and
`diagnosticsEqualityComparer`; this is a read/sort pipeline, not a `file` writer.
Its three callers are syntactic (`DiagnosticWithLocation`), semantic (`Diagnostic`),
and declaration (`DiagnosticWithLocation`) diagnostics. A single exact
`DiagnosticWithLocation` specialization cannot serve the semantic caller.

The resolution helpers read the callback result's `resolvedFileName`. They do not
write it. Watch and stop do write `resolution.files` and mutate sets/maps and
watcher counts; this is a different object/field from the resolved result.
The watch helper is exposed through `ResolutionCache`, and the harness calls it
at `src/harness/incrementalUtils.ts:339` with `ExpectedResolution` and a callback
returning `{ resolvedFileName: string | undefined }`. Module-only specialization
would lose that caller and the type-reference callers.

The literal `Readonly<Pick<D, "file">> & D` is **not** a readonly view in stock
6.0.3: the mutable member from D permits `v.file = undefined`. The accompanying
`readonly-control.cjs` compares it with an Omit-and-replace view and records the
single expected TS2540 on the latter write in `readonly-control.json`.

## Owner pins

Direct name searches are recorded in `owner-grep.txt` and `resolver-grep.txt`.
They are not sufficient to find enclosing owner pins: diagnostics helper edits
change the whole `createProgram` owner hashed by
`95-checker-axolotl-h/adapt.cjs:76–83` using `sites.json`'s owner SHA256 values.
`99-correctness-axolotl-x/ownership.json` explicitly records that same enclosing
ownership (and createProgram.getEmitHost). No pins were changed for this scout.
`98-axolotl-d/adapt.cjs` pins `ResolutionWithResolvedFileName.packageId` and
`ResolutionWithFailedLookupLocations.files` shapes; it does not pin the full
resolution helper bodies. No direct helper-name matches were found in adapt.cjs.
