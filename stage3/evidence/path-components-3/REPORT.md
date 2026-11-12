# Generic PathPathComponents interface proof

Implementation: `a3f67b011b1a810c282da390f80a3180df1be34b` on `typescript/path-components-interface-3`.
Main tip and proof base: `0c7052759671347775615af5b4a8028dd8688af8`.
Upstream: TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
This is the generic correction to the earlier PathPathComponents amendment.
The earlier evidence remains in ../path-components; this directory is a fresh run.
No production compiler source differs from main, and no lane files change.

The full applies ran at fb1c8d3b72de2686c486cf97685079589308dafd. Main then
advanced only in bridge/tsgo/units_test.go; tip-equivalence.json proves that this
is an excluded external Go test and all compiler-production, adaptation and
emission inputs are byte-identical to the final base. Compiler builds, types and
censuses were refreshed after that rebase. Apply and emission results therefore
validate the exact same inputs as final main, without repeating identical work.

## Contract

Final rule 103 target:
`export interface PathPathComponents<T extends Path = Path> extends Array<T> { __pathComponensBrand: undefined; }`

The planner accepts exactly the upstream any intersection, undefined intersection,
previous non-generic interface and generic interface. Every other owner spelling
is refused. The final interface replays with zero edits. The default fills all
13 existing bare uses, with no consumer edits. declarationProjection parses its
replacement from rule.after, checks rule.owner, derives its upstream base check
from rule.upstream, and retains the emitted scope's modifiers and indentation.
It contains no hardcoded owner or Path array spelling. Eight projection cases
(four states in both module and namespace output) match stock declaration emit.
Planner cases also cover LF/CRLF, unchanged JavaScript and five rejected drifts.

## Proofs

- Both full applies pass from independently reset pinned checkouts, with the
  existing locked dependencies already installed. Only the requested declaration
  differs among 747 source files (source-scope.json).
- Fresh emission builds remove only built/, preserving adapted diagnostic source.
  All ten JavaScript outputs and both public declarations are byte-identical;
  comparison.json records every size and SHA-256 hash. Public d.ts files are
  compared in both lib and built/local. Internal declaration changes are allowed.
- Adamic types exits 0 on all 79 roots before and after. Stock TypeScript under
  Adamic's options and matching prelude, pinned Node and regexp library views
  reports zero diagnostics on both trees. Commands, roots and logs are retained.
- Complete adaptation 40 replay on the recorded generic post-40 tree makes zero
  edits across 747 files. This avoids conflating 40 with downstream owner changes.
- Restoring the **non-generic interface** makes path-components-check.cjs exit 1.
  The generic tree passes before mutation and after restoration; the fully
  composed generic tree also passes.
- PathPathComponents remains internal: absent from the upstream public API golden
  and sanctioned-api.json. Inspection hashes and the full competing-pin scan are
  retained in public-api.json and pin-scan.txt. No other adapter pins it whole.

## All identifier uses

13 uses before and after: one declaration, one overload result, one import,
ten Readonly<PathPathComponents> references, zero casts. Every type reference
has zero explicit type arguments. All consumer rows are unchanged; only the
owner declaration changes from alias to generic interface.

| File | Line:column | Use |
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

Both runs use LATENT_FULL=1, LATENT_ONLY_FILE=<tree>/src/compiler/path.ts and
LATENT_ASSERT_NO_OUTPUT=1, with the same guarded main-tip binary. Raw JSONL,
unit times, environment and child exit records, and .go.txt overlay sources are
retained. Counts are unique (kind, where, reason, text) sites; only tree prefixes
are normalized. This is a path.ts-only measurement.

| Kind | Before | After | Delta |
|---|---:|---:|---:|
| NotYet | 20 | 20 | +0 |
| Refused | 6 | 6 | +0 |
| Boundary | 32 | 32 | +0 |
| error | 0 | 0 | +0 |
| panic | 0 | 0 | +0 |
| All unique sites | 58 | 58 | +0 |

Removed sites: 11. Added sites: 11.
All eleven replacements change only diagnostic rendering from PathPathComponents to PathPathComponents<Path>. Counts and locations are unchanged; census-display-changes.json verifies the exact replacement at each site. The existing getPathComponents overload-result refusal remains.

- removed: Boundary at `src/compiler/path.ts:1011:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:1012:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:492:1` — function signature or prologue failed; continued at next declaration.

- removed: Boundary at `src/compiler/path.ts:494:1` — function signature or prologue failed; continued at next declaration.

- removed: Boundary at `src/compiler/path.ts:495:1` — function signature or prologue failed; continued at next declaration.

- removed: Boundary at `src/compiler/path.ts:623:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:908:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:909:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:977:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Boundary at `src/compiler/path.ts:978:5` — failed statement; state rolled back; continued at next statement or declaration.

- removed: Refused at `src/compiler/path.ts:492:1` — overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[].

- added: Boundary at `src/compiler/path.ts:1011:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:1012:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:492:1` — function signature or prologue failed; continued at next declaration.

- added: Boundary at `src/compiler/path.ts:494:1` — function signature or prologue failed; continued at next declaration.

- added: Boundary at `src/compiler/path.ts:495:1` — function signature or prologue failed; continued at next declaration.

- added: Boundary at `src/compiler/path.ts:623:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:908:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:909:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:977:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Boundary at `src/compiler/path.ts:978:5` — failed statement; state rolled back; continued at next statement or declaration.

- added: Refused at `src/compiler/path.ts:492:1` — overload 1 of getPathComponents result PathPathComponents<Path> cannot be served by implementation result string[].

Full site text and unit/deferred coverage are in census-delta.json and sites-*.json.
Both census headers report checker_rejected=false and empty diagnostics; the
legacy measurement label saying "checker-rejected program" is not authoritative.

## Process integrity

Every subprocess wrapper reports its child's own exit code. run.py enforces a
1740-second limit and records command, elapsed time, exit code and stopping reason.
types-run.py and census-run.py exec the actual binary. All required children exit
0 except the expected restoring mutant, which exits 1. No measured process runs
for 30 minutes. The first equivalence inspector assumed Go's TestGoFiles field;
this test is external and appears in XTestGoFiles. The corrected inspector checks
both categories and passes; the initial failed inspector record is retained. All Go evidence filenames end .go.txt. No evidence is in adapt/.

Scripts retain the exact /workspace paths used for reproducibility. Source
/workspace/adamic-tools/env.sh for Go/Node and set NODE_PATH to
/workspace/adamic/stage3/api/node_modules for adaptation helpers. comparison.json,
census-delta.json and use inventories are reproduced by compare.py,
census-delta.py and uses.cjs. SHA256.json hashes every retained artifact except
itself. Binaries and the replay tree are omitted; their hashes and source
provenance are retained.
