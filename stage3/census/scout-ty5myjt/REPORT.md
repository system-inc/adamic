# Scout ty5myjt — checkpoint, measurements pending

No shape is recommended yet. Candidate diagnostics, full single-file census deltas,
and JavaScript byte comparisons have not been measured. This checkpoint is not a
completed scout result and must not be used to authorize an adaptation.

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
