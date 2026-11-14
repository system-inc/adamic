# Require result annotation: acceptance checkpoint

Task #baobab item 9. Base 8e3a784aa5a601f57342c07b68f145fe56252c21.
Upstream v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

- Full main apply: exit 0.
- Stock TypeScript 6.0.3 with pinned Node types: 79 roots, main 0 diagnostics,
  candidate 0 diagnostics. The earlier missing-Node-types probe is invalid and
  retained as raw/stock*.gz; only stock-valid.json is acceptance evidence.
- Full sys.ts census: LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1,
  LATENT_ONLY_FILE=<adapted sys.ts>, CommonJSBundle; all three streams completed
  with relation trailers and zero checker diagnostics.
- Findings: 185 -> 184; 0 new, 0 changed, 1 gone (1610:22, refusal_scan).
  delta.json retains the full removed finding and its phase.
- Restoring original arrow: all 185 main findings return, exactly.
- Annotated return module:any: no new finding. Existing computed-path require
  host model treats that call as never; no model or compiler rule is changed.
- JavaScript: 604 files, 1208 comparisons, byte-identical with comments kept/removed.
- adamic types: 79 roots, exit 0.
- Whole-owner amendment contract: pass (contract.log).
- Separate upstream npm ci + node_modules/.bin/hereby dts: both pass.
  typescript.d.ts and tsserverlibrary.d.ts are byte-identical; both empty diff
  files are retained, with per-side hashes in declarations.json.
- Whole unfiltered lane: PASS, 106,366 passing, one sanctioned API failure, zero pending.
  Candidate full apply, install and build exit 0; oracle/tests exit 1 only for
  the sanctioned API baseline row. See lane/report.json and complete logs.
- Final source scope: 747 files; only sys.ts changes, with the exact measured
  annotation bytes. Actual adapter replay writes nothing and preserves bytes/mtime.
- Existing crypto source and JavaScript scope checks pass (747-file inventory).
- Implementation is one commit on the exact base:
  ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8.

Source snapshots and original streams/logs are gzip-compressed in raw/.
All source work is in isolated scratch worktrees. No Loom submission is made.
