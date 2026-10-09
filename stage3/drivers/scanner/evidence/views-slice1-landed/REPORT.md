First native stop: corePublic.ts:9:5, Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added, both split modes and both runs.
Landed main 031a1259 builds; records 8c013f1c, newer descendant cabe36b0 and library fa270a53 conflict and are skipped.
Both runs' Node scanner streams match the full-tree reference; no native scanner comparison or timing exists.
Four isolated controls per compiler match Node and catch byte mutants; scanner Node token-end mutants are caught.
Pending scanner verification behind earlier stops; no rehearsal/scratch branch is pushed or proposed to land whole.

## Commands and refs

```sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-views-landed-main-retry origin/main > /tmp/scanner-views-landed-main-retry.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-views-landed-combined 031a1259bc7973934792dc6cb1bd4074fc2204b9 > /tmp/scanner-views-landed-combined.log 2>&1
```

Main-only runs first. Because the unchanged runner stops at merge conflicts,
measurement-preflight.py attempts the requested merges in a separate never-pushed
scratch, aborting each conflicting merge. It also tries cabe36b0, a newer records
descendant containing 8c013f1c, as the newest advertised candidate before library.
No candidate is rebased onto 031a1259. All candidates conflict, leaving only main
for the second command. No hand resolutions, compiler edits or adaptation edits.

| Attempt | SHA | Result |
| --- | --- | --- |
| origin/codex/records-maplike-next | 8c013f1cf3ec69fe80e60501740c2c45e1d1114c | skipped, 3 conflicting paths |
| origin/codex/maplike-records | cabe36b00e287e959eaf1f3d78f289d1489cfdef | skipped, 10 conflicting paths |
| origin/library/area-on-next | fa270a53fcc7b595cb58fd952222e83f27d7d18b | skipped, 16 conflicting paths |

Exact paths are retained in stack.json and *-conflicts.txt. The second run does
not establish any records/library integration. Both actual fetched bases and
compiler identities are recorded separately below and in the runner summaries.

## Comparisons, timings and pending evidence

Both split modes in both commands perform the full-tree/slice Node byte
comparison with diff exit 0. The fixed upstream corpus contains 509,014
skip-trivia tokens, 860,418 retained-trivia tokens, 466 error rows and 81 files.
All four scanner comparison controls pass; all four scanner token-end mutants
fail diff with exit 1. Each full-tree reference also catches its token-end mutant.
There is no native scanner binary, native byte diff or scanner output mutant.

| Run | Fetched main | Compiler SHA | Binary SHA-256 |
| --- | --- | --- | --- |
| main | 031a1259bc7973934792dc6cb1bd4074fc2204b9 | 031a1259bc7973934792dc6cb1bd4074fc2204b9 | e45b447f62c0cfe7160a4f5144306bd6c3281d5feb5455d53aa402a15c1383d0 |
| combined | 031a1259bc7973934792dc6cb1bd4074fc2204b9 | 031a1259bc7973934792dc6cb1bd4074fc2204b9 | 61b9a6a655cb63e2dfce36cdd3c3a4409408b2dc876f9f0474237d7db7153e38 |

| Run | Phase | Exit | Wall seconds |
| --- | --- | --- | --- |
| main | setup | 0 | 190.554 |
| main | compiler-build | 0 | 5.75 |
| main | apply | 0 | 21.017 |
| main | reference | 0 | 10.137 |
| main | slice | 0 | 3.792 |
| main | scanner-0 | 1 | 6.758 |
| main | scanner-1 | 1 | 6.856 |
| main | command total | 1 | 273.27 |
| combined | setup | 0 | 176.127 |
| combined | compiler-build | 0 | 5.207 |
| combined | apply | 0 | 19.822 |
| combined | reference | 0 | 9.666 |
| combined | slice | 0 | 3.64 |
| combined | scanner-0 | 1 | 7.207 |
| combined | scanner-1 | 1 | 6.793 |
| combined | command total | 1 | 257.22 |

Setup timing lines are retained verbatim in each runner summary and raw logs.
Both setups source /workspace/adamic-tools/env.sh; nproc=5, cgroup quota=4.
summary.json marks the native result blocked and milestone=false. Isolated
controls and all behind-placeholder observations remain pending; they await
clean feature integration and actual scanner execution without discovery stubs.

## #whkxbc7 markings

Uses the authoritative checkpoints and owners from
../main-area-next-replay/WHKXBC7.md. A control pass closes only that control's
admission diagnostic. Unvisited scanner sites are not retired by absence.

| Supplied checkpoint | Owner | Main | After skipped merges |
| --- | --- | --- | --- |
| corePublic.ts:9:5 | records-maplike #rhjc4x4 | still there | still there |
| debug.ts:8:5 | compiler area-stack group 1 | gone in isolated control; pending real scanner verification | gone in isolated control; pending real scanner verification |
| debug.ts:15:14 | compiler #b5w3ycg | still there | still there |
| diagnosticInformationMap.generated.ts:13:34 | compiler #b5w3ycg | still there | still there |
| scanner.ts:3497:67 | compiler #b5w3ycg | still there | still there |
| utilities.ts:65:22 | runtime #4gkdjsz | still there | still there |
| scanner.ts:1269:28 | compiler #cvhj5fk | still there in isolated control; scanner checkpoint not reached | still there in isolated control; scanner checkpoint not reached |
| core.ts:107:20 | compiler stricter options #k881crd | still there | still there |
| utilities.ts:13:17 | compiler area-stack | gone in isolated control; pending real scanner verification | gone in isolated control; pending real scanner verification |
| scanner.ts:113:5 | compiler area-stack | gone in isolated control; pending real scanner verification | gone in isolated control; pending real scanner verification |
| scanner.ts:432:59 | compiler #d9eemrs | gone in isolated control; pending real scanner verification | gone in isolated control; pending real scanner verification |
| scanner.ts:4097:24 | compiler #9wc5q5j | still there in isolated control; scanner checkpoint not reached | still there in isolated control; scanner checkpoint not reached |
| Error.captureStackTrace | library #jj9z3qn | still there in isolated control; scanner checkpoint not reached | still there in isolated control; scanner checkpoint not reached |

Mutable namespace export, generic return and Object.entries controls now pass
on landed main; computed-field admission still passes. Namespace and computed
field controls print admission markers and do not exercise complete field
semantics. Generic-return and entries controls execute and match Node. Each of
these four controls catches exactly one changed native-output byte, on each
compiler. Their scanner checkpoints await traversal past MapLike and subsequent
blockers; none is a whole-scanner milestone. MapLike is blocked again relative
to the earlier records-integrated marking because both records merges conflict.

Error capture and the undefined! property initializer are directly witnessed
again: Node succeeds, while native build refuses the static method and the
non-null assertion respectively. The three original any controls remain
intentionally refused and do not undo the accepted adaptation/driver typing.

## Ordered observations and coverage

| Run | Order | Location | Diagnostic | Owner | Scope |
| --- | --- | --- | --- | --- | --- |
| main | 1 | corePublic.ts:9:5 | Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | records-maplike #rhjc4x4 | first unchanged stop |
| main | 2 | debug.ts:14:14 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| main | 3 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| main | 4 | utilities.ts:65:22 | stage 0 can't lower typed array element type Uint16Array yet | runtime #4gkdjsz | behind throwing discovery placeholders |
| main | 5 | scanner.ts:3510:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| main | 6 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | compiler stricter options #k881crd | behind throwing discovery placeholders |
| main | 7 | scanner.ts:778:15 | stage 0 can't lower destructuring a string yet | compiler, no task assignment supplied (new relative to #whkxbc7) | behind throwing discovery placeholders |
| combined | 1 | corePublic.ts:9:5 | Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | records-maplike #rhjc4x4 | first unchanged stop |
| combined | 2 | debug.ts:14:14 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| combined | 3 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| combined | 4 | utilities.ts:65:22 | stage 0 can't lower typed array element type Uint16Array yet | runtime #4gkdjsz | behind throwing discovery placeholders |
| combined | 5 | scanner.ts:3510:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | compiler #b5w3ycg | behind throwing discovery placeholders |
| combined | 6 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | compiler stricter options #k881crd | behind throwing discovery placeholders |
| combined | 7 | scanner.ts:778:15 | stage 0 can't lower destructuring a string yet | compiler, no task assignment supplied (new relative to #whkxbc7) | behind throwing discovery placeholders |

Each runner summary retains its ordered stops, scope and attempted witnesses.
The built-in catalogue stops at string destructuring; a fresh supplementary
witness runs on Node and fails each compiler with the matching diagnostic.
No further traversal is claimed after that boundary, and no fifteen-stop
completeness claim is made. Exact private source/body changes are preserved
in *-private-placeholders.json.gz. All later observations are pending behind
throwing placeholders: they await real declarations and implementations for
MapLike, Debug failure handling, diagnostics, Uint16Array/UTF16 and Array
construction, then an actual native scanner comparison.

Both real runner summaries validate. A false-pass summary retaining these exact
conflicts but otherwise complete success claims is rejected solely by merge
rules. Checks and mutant logs are retained. No whole-package tests, full gate,
oracle fixtures or counts refresh were added/run.

The first main attempt passed setup but npm ci failed with exit 228, ENOSPC.
A newer-records merge attempt also hit ENOSPC before a usable conflict result;
its logs and partial scratch copy are preserved. The fresh retries above are
the reported measurements. Automatic approval review initially rejected a
combined force-cleanup/run command due to potential uncommitted work loss.
Narrow cleanup was subsequently reviewed after clean statuses and dependency
pins were checked; nine old output archives passed round-trip SHA-256 checks.
The decisive disk recovery was go clean -cache: the 20 GB regenerable cache
was cleared, restoring 19 GB free, making the main retry cold-cache setup.
Prior evidence, binaries, refs and logs remain; recovery metadata is retained.
Current large token artifacts remain in scratch with hashes in
large-output-manifest.json. Other stdout/stderr, witnesses and conflict logs
are retained in raw-logs.json.gz. Only this evidence directory is delivered.
