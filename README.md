# tracing-fs-import stopped evidence

Task nqp4p7g. Started from fetched main d35fe5e7f. The fs-view retirement
was not an ancestor of that tip, so the authorized base is
9ad919acec5955055898f4978a100a72a3742f4e on typescript/tracing-fs-views.
Pinned upstream source: 050880ce59e30b356b686bd3144efe24f875ebc8.

## Required stop

The complete emitted JavaScript inventory has ten files. Three change:
_tsc.js, run.js, typescript.js. emitted-js.diff lists every hunk.
In _tsc.js, the hunk at old line 2169 removes the hostErrors.ts section,
including errorMessage. The hunk at old line 3461 reinserts errorCode and
propertyOf later, immediately before sys.ts. The final typescript.js hunk
also moves the remaining hostErrors section. This is helper elimination
and code relocation outside tracing's import block and startTracing,
not a pure esbuild rename. The user's explicit rule therefore stops work.
Other outside hunks have not been certified as pure renames.
No final implementation branch is pushed. The draft is retained on the
checkpoint branch. The 398-shard lane and restoring mutant are not run.
Runtime export-key equality, final zero-edit replay and direct collision
probe completion are not claimed.

## Completed observations

- Required setup and stage3/api npm ci: exit 0.
- Official before and after apply.sh: exit 0.
- All 79 roots, real current adamic types: exit 0 on both trees.
- Full src inventory: only tracing.ts differs. See source-scope.json and source.diff.
- Both serial npm builds and harness builds: exit 0.
- Both public declarations byte-identical. Hashes are in public-types.json.
- Complete tracing full censuses: exit 0, no checker diagnostics, complete trailers.
  Control's ambient require refusal at tracing.ts:380:1 is gone after the edit.
- Real per-file census for core.ts: control Refused at tracing.ts:380:1;
  after NotYet at sys.ts:1654:13, node:inspector.Session.post.
- Plain real tsc runs: both exit 0, stdout, stderr and emitted files identical.
- --generateTrace runs: both exit 0, but the strict trace comparison fails:
  551 versus 561 events. Removing only named ts, dur and pid fields does
  not reconcile event inventories. This observation does not establish
  whether the adaptation caused the difference. Raw trace files retained.

## Failed or superseded observations

The stock census loader generator first refused stale roots[:len(roots)-1]
spelling. Its current source uses roots[:len(paths)]. A scratch-only generator
and config_hook.go.txt reconcile that spelling and the rooted-path API.
The lower and loader overlays were composed in scratch; production unchanged.
The initial broad statement census was terminated with child exit -15 and
wrapper exit 241, then replaced by the repository's real per-file census.
A multi-program direct binding probe was killed with child exit 137.
It is not a passing guard proof. No successful isolated rerun is claimed.
The cgroup then reported oom=2, oom_kill=1. No lane is run on this observation.

The initial apply began before evidence-script and README-only changes were
finished. Its source result is checked, but its old input-key cache is not
claimed as a product for the final draft. Further proof is required before
any final implementation publication.
