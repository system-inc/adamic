# Paired latent census: reachable and phantom overload results

Report and evidence only. Every scratch merge completed without conflict. Only this census directory is committed; scratch merge commits and compiler changes are not pushed.

| Input | Pinned commit |
| --- | --- |
| Main tip captured at preparation | 0c7052759671347775615af5b4a8028dd8688af8 |
| PathPathComponents interface adaptation | a3f67b011b1a810c282da390f80a3180df1be34b |
| Step 2 | 3a7ea290fca9e3f3325ffa8206a861b22f441583 |
| Step 2b (head only) | 694834e95c3dd79bc4d350f4560d5a16b175f0ad |
| Step 3 (head only) | 78933e52b0abd0da2fce91e73b4b71c1b36dc4e4 |
| Base scratch composition | be9e9cc6f77121cdb6970a18915c7c61cdb65b78 |
| Head scratch composition | 65af0b47a56f444ae54cfc2a8761e63b20f73f8e |

The base is main plus the path adaptation and step 2. The head is that exact base plus steps 2b and 3. The four merge parents and tree IDs are retained in [composition.json](composition.json). Both sides read the single adapted tree produced by the base composition’s `apply.sh`, including a3f67b01. Both head commits and the final composed head diff touch no adaptation or meter input; see [head-adaptation-input-diff.txt](head-adaptation-input-diff.txt).

## 1. Regressions

| Regression category | Raw | Distinct |
| --- | --- | --- |
| Previously finding-free units now stopping | 0 | 0 |
| New panics | 0 | 0 |
| New refusal-scan sites | 0 | 0 |


There are no regressions to name.

All panic ledgers and every regression identity are retained in [delta.json](delta.json).

Panics before/after: 0/0 raw; 0/0 distinct.

## 2. Cleared, moved and new findings

| Measure | Raw | Distinct outermost accesses |
| --- | --- | --- |
| Cleared: no finding remains at identity | 2 | 2 |
| Moved: old reason gone, another remains | 0 | 0 |
| New findings | 1 | 1 |
| New sites | 1 | 1 |
| Changed unit first stops | 16 | 16 |
| Regressions | 0 | 0 |

Raw identities retain kind, source identity, reason and text. Distinct identities use the unmodified base composition’s `stage3/census/pair` tool: source path, source SHA-256, outermost contiguous property/element access start and exact reason. Newly exposed reasons in already-blocked units are not automatically regressions. [pair-report.md](pair-report.md) and [delta.json](delta.json) retain both complete views.

| Finding kind | Base raw | Head raw | Base distinct | Head distinct |
| --- | --- | --- | --- | --- |
| NotYet | 528 | 529 | 521 | 522 |
| Refused | 4423 | 4421 | 4228 | 4226 |


Cleared raw identities:

| Location | Kind | Exact reason |
| --- | --- | --- |
| src/compiler/path.ts:492:1 | Refused | overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] |
| src/compiler/utilities.ts:9990:1 | Refused | overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] |


New raw findings:

| Location | Kind | Exact reason |
| --- | --- | --- |
| src/compiler/utilities.ts:10005:89 | NotYet | a boolean \| undefined variable a function value captures |


## 3. getPathComponents: all 11 units

“Newly admitted” follows the requested criterion: the unit stopped at base and, on head, either has no first stop or first-stops on something other than an overload-result refusal. It does not claim that an already-blocked unit compiles. Locations are one-based lines and UTF-16 columns.

| Unit | Unit location | Base first stop | Head first / next stop | Newly admitted |
| --- | --- | --- | --- | --- |
| getPathComponents | src/compiler/path.ts:492:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| getPathComponents | src/compiler/path.ts:494:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| getPathComponents | src/compiler/path.ts:495:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| getNormalizedPathComponents | src/compiler/path.ts:622:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| comparePathsWorker | src/compiler/path.ts:884:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| containsPath | src/compiler/path.ts:963:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| containsPath | src/compiler/path.ts:965:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| containsPath | src/compiler/path.ts:967:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| getPathComponentsRelativeTo | src/compiler/path.ts:1010:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |
| guessDirectorySymlink | src/compiler/utilities.ts:9559:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | Refused: an unproven relation from "" to T: the source is not assignable to the target at src/compiler/path.ts:513:45 | yes |
| normalizePathForCJSResolution | src/compiler/moduleNameResolver.ts:1966:1 | Refused: overload 1 of getPathComponents result PathPathComponents&lt;Path&gt; cannot be served by implementation result string[] at src/compiler/path.ts:492:1 | No lowering stop | yes |

**11/11 newly admitted.** Exact next-stop reason words are in the table above.

Every newly admitted unit: getPathComponents (`src/compiler/path.ts:492:1`); getPathComponents (`src/compiler/path.ts:494:1`); getPathComponents (`src/compiler/path.ts:495:1`); getNormalizedPathComponents (`src/compiler/path.ts:622:1`); comparePathsWorker (`src/compiler/path.ts:884:1`); containsPath (`src/compiler/path.ts:963:1`); containsPath (`src/compiler/path.ts:965:1`); containsPath (`src/compiler/path.ts:967:1`); getPathComponentsRelativeTo (`src/compiler/path.ts:1010:1`); guessDirectorySymlink (`src/compiler/utilities.ts:9559:1`); normalizePathForCJSResolution (`src/compiler/moduleNameResolver.ts:1966:1`)

The adapted PathPathComponents interface and its source hash are retained in [path-components-interface.json](path-components-interface.json).

## 4. getSupportedExtensions: all 5 units

| Unit | Unit location | Base first stop | Head first / next stop | Newly admitted | Head invariant-mutable |
| --- | --- | --- | --- | --- | --- |
| getSupportedExtensions | src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:10005:158 | yes | no |
| getSupportedExtensions | src/compiler/utilities.ts:9992:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:10005:158 | yes | no |
| getSupportedExtensions | src/compiler/utilities.ts:9994:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:10005:158 | yes | no |
| isSupportedSourceFileName | src/compiler/utilities.ts:10146:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions at src/compiler/utilities.ts:10012:105 | yes | no |
| getFileNamesFromConfigSpecs | src/compiler/commandLineParser.ts:3950:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions at src/compiler/utilities.ts:10012:105 | yes | no |

**5/5 newly admitted** by the same criterion. Exact next-stop reason words are in the head column.

Every newly admitted unit: getSupportedExtensions (`src/compiler/utilities.ts:9990:1`); getSupportedExtensions (`src/compiler/utilities.ts:9992:1`); getSupportedExtensions (`src/compiler/utilities.ts:9994:1`); isSupportedSourceFileName (`src/compiler/utilities.ts:10146:1`); getFileNamesFromConfigSpecs (`src/compiler/commandLineParser.ts:3950:1`)

`adamic/invariant-mutable` appears in **0/5 head first/next-stop diagnostics**. The table preserves each finding’s exact `reason`; the rule code is detected in its full diagnostic `text`, including its fix, and is not added to the reason words.

## 5. Overload-result first-stop counts

| Measure | Base composition | Head composition |
| --- | --- | --- |
| Overload-result first stops | 61 | 45 |
| Resolved source files | 81 | 81 |
| Checker diagnostics | 0 | 0 |
| Units | 4564 | 4564 |
| Attempted units | 4060 | 4060 |
| Attempted units with no first stop | 3177 | 3187 |


## Method and validation

The workflow reuses the tools and first-error latent mode from `stage3/census/overload-proof-3a7ea290` (commit 353218cbd01286116bd97fc1475b51bc54cdce75). No earlier measurement stream is reused. Both overlays are built from their own composed source trees, using their own `make_overlay.py` and `entry_overlay.py`, and exact recursive submodule pins. Both read the same base-produced adapted bytes, pinned to TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. Source hashes, overlay maps and executable hashes are in the evidence.

`LATENT_FULL=0` and `LATENT_ASSERT_NO_OUTPUT=1`. Ordinary Load and Lower remain disabled; no backend or usable IR is produced. Per-unit eligibility, findings and timing records are retained. A unit without a finding is not proof of whole-program compilation.

Each side runs eight disjoint attempted-unit shards over the complete project, with two shards per side concurrent and `GOMEMLIMIT=3GiB`. Selection is after unchanged checker-body eligibility and generic deferral, before an independent attempt. Dependencies and registration are not filtered. Shards 0 and 1 retain complete refusal scans and must agree exactly; shards 2–7 omit only repeated scans. Deferred/skipped unit records deduplicate, and generic call witnesses union under the original Go ordering and five-witness cap. Each attempted unit must occur exactly once. Process-scoped relation/cache reports remain separate, never summed; any overflow or global finding stops composition.

Both binary variants on both sides pass the unsharded-versus-sharded composition fixture, including a real overloaded-value first stop and cross-shard generic witnesses. Both existing latent audits and all 20 pairing tests pass. The bounded fixture validates the splitting mechanism; no full unsharded census is claimed. A stock-parser inventory independently checks every top-level unit and its source order. Final validation checks shard membership, exact archived-stream recomposition, all per-unit timings, source hashes and both requested cohort sizes.

Every timing wrapper reports its child’s own exit code and exits with it; signal exits use shell encoding. Every shell wrapper also reports and propagates its child’s status. Each child is bounded at 1,740 seconds with at most ten seconds to kill, below 30 minutes. No measured child timed out. All Go evidence is named `.go.txt`; raw JSONL and logs are gzip compressed without changing their bytes. `commands.md`, the retained scripts, provenance, merge logs, per-side raw streams and `validation.json` make the measurement reviewable.

Each census child’s own measurements (overlapping runs):

| Side | Shard | Wall s | User s | System s | Peak RSS MiB | Child exit |
| --- | --- | --- | --- | --- | --- | --- |
| base | 0 | 1301.989 | 1273.771 | 20.44 | 2758.02 | 0 |
| base | 1 | 1280.485 | 1230.362 | 25.692 | 2804.6 | 0 |
| base | 2 | 135.806 | 133.851 | 3.245 | 1322.75 | 0 |
| base | 3 | 134.319 | 132.596 | 4.147 | 1167.95 | 0 |
| base | 4 | 139.54 | 135.081 | 3.286 | 1012.97 | 0 |
| base | 5 | 143.531 | 140.302 | 2.734 | 994.9 | 0 |
| base | 6 | 136.805 | 132.305 | 3.28 | 990.72 | 0 |
| base | 7 | 140.252 | 135.879 | 3.447 | 985.51 | 0 |
| head | 0 | 1275.611 | 1240.815 | 21.369 | 2750.41 | 0 |
| head | 1 | 1257.098 | 1214.473 | 20.738 | 2783.05 | 0 |
| head | 2 | 140.974 | 138.538 | 3.228 | 1197.52 | 0 |
| head | 3 | 138.032 | 137.013 | 3.228 | 1154.38 | 0 |
| head | 4 | 135.601 | 131.869 | 3.329 | 1003.96 | 0 |
| head | 5 | 142.566 | 137.32 | 2.994 | 1015.76 | 0 |
| head | 6 | 132.765 | 127.405 | 3.413 | 998.87 | 0 |
| head | 7 | 141.54 | 136.468 | 3.673 | 989.62 | 0 |
