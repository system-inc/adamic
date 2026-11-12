# PathPathComponents interface proof

Implementation: `f6b046353ebd03b619ee3feb6166b0b9d51b385d` on `typescript/path-components-interface`.
Base: `d4693fc61d2d01fb4e1e993667ae5028da109cc3`, verified as remote main before commit.
Upstream: TypeScript 6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
Compiler production sources are unchanged from main; binary hashes and submodule
pins are in provenance.json. No new adaptation directory or lane-file change.

## Results

- Both full applies pass (apply-before-status.json, apply-after-status.json).
- Only src/compiler/path.ts differs among 747 source files, and only by the
  requested one-line declaration (source-scope.json).
- All ten JavaScript build outputs are byte-identical. Both public declaration
  files, typescript.d.ts and tsserverlibrary.d.ts, are byte-identical in both
  lib and built/local (comparison.json has sizes and SHA-256 hashes).
- Adamic types exits 0 on exactly 79 roots before and after. Stock TypeScript
  6.0.3, under Adamic's options and matching prelude/Node/regexp library views,
  reports zero diagnostics before and after. See roots-*.json, types-*.log,
  stock-*.json and their child exit-status records.
- Complete adaptation 40 replay on its recorded post-stage tree changes zero
  of 747 files (replay40.json). This isolates 40 from later owner adaptations.
- Restoring the undefined intersection makes 40's path-components-check.cjs
  exit 1; restoring the interface makes it exit 0. See restoring-mutant-final
  and check-restored logs/statuses. The fully composed tree passes the check.
- The planner accepts upstream-any, previous-undefined, and final-interface
  states under LF and CRLF, replays with zero edits, emits identical JavaScript,
  and rejects five owner drifts. Its independent declaration projection matches
  stock module and namespace declaration emission (planner.json, projection.log).

PathPathComponents is internal: absent from the upstream public API golden and
sanctioned-api.json, and its source declaration is marked @internal.
public-api.json records the inspection hashes. No API-sanction row is needed.
pin-scan.txt contains the complete naming/line scan: outside 40, matches are
historical census reasons in 41's evidence, not whole-declaration pins.

## Uses

There are 13 identifier uses both before and after: one declaration, one overload
return reference, one import, ten Readonly<PathPathComponents> references, and
**zero casts**. Only the declaration's AST kind changes (alias to interface).
The complete AST inventory is in uses-before.json and uses-after.json.

| File | Line:column | Use after amendment |
|---|---|---|
| src/compiler/path.ts | 457:18 | InterfaceDeclaration |
| src/compiler/path.ts | 492:48 | TypeReference |
| src/compiler/resolutionCache.ts | 57:10 | ImportSpecifier |
| src/compiler/resolutionCache.ts | 269:68 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 310:66 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 330:52 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 330:103 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 353:34 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 358:48 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 402:33 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 406:34 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 444:33 | Readonly<PathPathComponents> |
| src/compiler/resolutionCache.ts | 472:34 | Readonly<PathPathComponents> |

## Single-file full latent census

Both runs set LATENT_FULL=1, LATENT_ONLY_FILE to the absolute path.ts in their
respective trees, and LATENT_ASSERT_NO_OUTPUT=1. They use the same guarded
main-tip binary, generated with make_overlay.py's normal transformation and
physical evidence names ending .go.txt. Raw JSONL, per-unit times, environment
records, overlay sources, and command/child-status records are retained.

| Site kind | Before | After | Delta |
|---|---:|---:|---:|
| NotYet | 20 | 20 | 0 |
| Refused | 6 | 6 | 0 |
| Boundary | 32 | 32 | 0 |
| All unique sites | 58 | 58 | 0 |

The NotYet/Refused total is **26 -> 26 (delta 0)**. No site changes: added and
removed lists are empty in census-delta.json. Both runs record 102 events over
81 units, with no checker diagnostics, skipped checker bodies, errors or panics.
70 functions and seven statements are attempted. Four generic units are deferred;
one is instantiated. The three never-reached forEachAncestorDirectory units
remain listed in census-delta.json. This is a path.ts-only measurement.

The existing getPathComponents overload-result refusal remains at path.ts:492:1
on both sides: implementation result string[] cannot serve PathPathComponents.
The interface amendment provides phantomExtension's required shape for the
Optimizer follow-up; it does not claim to repair the overload implementation.
The raw tool's legacy measurement label says "checker-rejected program" even
though its authoritative checker_rejected flag is false and diagnostics empty.

## Process integrity and reproduction

run.py caps each child at 1740 seconds, retains its own exit code (including
signals), and returns that code. types-run.py and census-run.py exec the actual
binary so they do not hide its status. All measured runs finish within 30 minutes.
The two full applies take about 1240 and 1197 seconds respectively.

An environment reconnect interrupted some supervisors after their child work
completed but before status files were saved. Both builds and baseline stock
checking were rerun to obtain complete child exit records; compare-final then
revalidated all bytes and diagnostics. The initial planner invocation omitted
NODE_PATH and failed before executing checks; planner-final is the successful
correctly configured run. Both initial and final records are retained.

Scripts retain the exact /workspace paths used. Set NODE_PATH to
/workspace/adamic/stage3/api/node_modules for adaptation helpers; source
/workspace/adamic-tools/env.sh for the pinned Go and Node tools. Every status JSON
records its command, child exit code, elapsed time and stopping reason. The
full applies use --adapt-checkout on independent pristine pinned checkouts;
replay40 uses the after-40 snapshot recorded by full apply. comparison.json,
census-delta.json and the inventories are regenerated by compare.py,
census-delta.py and uses.cjs. All evidence is outside stage3/adapt.
