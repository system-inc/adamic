# LineMapSource, baobab variant (e)

Implementation: `e56c5d7809b44fe297c26c22302f4cd13965c271` on `typescript/line-map-source`.

Base: `efe887763`; pinned TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Ruling: Adamic, baobab, Nov 11 11:28Z, accepted by Compiler.
Measurement precedent: the full fourth section of `ddbfff50`, and
getPositionOfLineAndCharacter's existing internal overload.

Owner is `71-writable-views`, which already guards SourceFileLike.lineMap.
The full-source comparison finds exactly the measured two-file edit. The whole
documented SourceFileLike and both helper bodies remain unchanged. No compiler
backend or rule changes are included.

| Proof | Result |
| --- | --- |
| Final keyed apply | Child exit 0 |
| Production adamic types | All 79 compiler roots, child exit 0 |
| Services full program | 0 diagnostics, 0 TS2386 |
| Ten emitted JavaScript files | Byte-identical |
| Built runtime export keys | 2,256, identical |
| Parser and SourceMapSource-shaped witnesses | Byte-identical at six positions, including 0, a line start, and end |
| stripInternal declaration emit | 278 files, 0 emit diagnostics |
| Upstream declaration bundler | Child exit 0 |
| Normalized public API delta | 0 |
| Actual stage3/lane api_check | 0 errors and 0 unsanctioned changes |
| sanctioned-api.json | Unchanged |
| Guard and whole adaptation idempotence | Second application is a no-op, all source bytes unchanged |
| Drift and half-state controls | All 11 rejected before writes |
| Baseline core.ts first stop | tracing.ts:206:67 |
| Edited core.ts first stop | tracing.ts:232:34, a cast the runtime cannot check |
| Internal-overload restoring mutant | tracing.ts:206:67 |

The original production census driver is used unchanged and reports both its
per-file record and its one-root whole-program repetition. Their results agree.
The mutant retains LineMapSource, the new implementation parameter and the new
getLineStarts parameter; only the first internal overload is removed, leaving
the public SourceFileLike overload first. Its source mutation and hashes are retained.

The bundler implementation is unchanged. Its TypeScript package import selects
the pinned stage3 API parser by absolute path; its other imports use upstream's
lockfile dependencies. The original script hash is retained in bundler-status.json.
The runtime witness bundles each actual completed source tree with esbuild.

All subprocess logs are files and statuses report the child's own exit. The
abandoned after-apply build was terminated and reaped after documentation was
finalized; only final-apply is used as proof. The initial setup control began
before main was fetched and failed on its old cohere pin; setup was repeated
successfully at main. The initial bundler dependency-resolution failure is under
controls and is excluded from final counts. No API sanction changed.

Large raw artifacts are gzip-compressed after comparison. artifact-index.json
records their original hashes and sizes; SHA256SUMS covers the retained files.
No raw proof artifact is added under stage3/adapt in the implementation commit.
