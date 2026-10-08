First native stop: debug.ts:8:5, mutable namespace export, split 0 and 1.
Main d72728e5 and records 8c013f1c merge cleanly; library bb9bb3f9 conflicts in 56 paths and is skipped.
Both full-tree Node diffs pass; fifteen distinct Node-success/compiler-failure witnesses are retained.
Historical controls 1 (MapLike) and 11 (computed field) now build and match Node; their one-byte mutants are caught.
No native scanner execution; exact #whkxbc7 list unavailable, so historical marks and suggested owners are provisional.

## Run and refs

```sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-main-next-combined origin/main origin/codex/records-maplike-next origin/area/library > /tmp/scanner-main-next-combined.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-main-next-records origin/main origin/codex/records-maplike-next > /tmp/scanner-main-next-records.log 2>&1
```

Fetched main: `d72728e570fe09d91cf55b37b564dbad0fefc24d`. Records:
`8c013f1cf3ec69fe80e60501740c2c45e1d1114c`. Library:
`bb9bb3f9f3da1eeb4ca487a499f29f5d9c0bb921`. Never-pushed compiler scratch:
`0e38f5dc59972303f1380c5de5d9250a2f9b88ac`. Compiler binary SHA-256:
`e689caf0b422d273b3663f2ee94201b6f0cd1481bc16753c1c482f298cfd086a`.

No records rebuild or library/area-on-next-2 was advertised at fetch time.
The first command exits 2: main and records merge cleanly, library conflicts,
and its merge is aborted. The second command omits library, builds the compiler
and exits 1 with scanner-blocked. Exact 56 paths are in conflicting-paths.txt
and conflict-summary.json. No conflicts were hand-resolved. The delivery branch
merged origin/main cleanly. Compiler, adaptation and runner sources are unchanged
by this unit; only scanner evidence and measurement helpers are added.

## Comparison first and timings

Both complete slice Node streams match the full-tree reference byte for byte:
81 files, 509,014 skip-trivia tokens plus 860,418 retained-trivia tokens,
466 error rows. Raw SHA-256 `6b9d9e0fdb5eea142ee2021d436fd82ae77526487b4ee350861b4f180874644d`,
108,021,879 bytes. Absolute pass-header paths change
between output directories. Both native builds stop at the same unchanged
mutable namespace export. No native token diff, native-output scanner mutant,
scanner binary size or native scanner runtime timing exists.

| Phase | Exit | Wall seconds |
| --- | --- | --- |
| setup | 0 | 196.746 |
| compiler-build | 0 | 5.759 |
| apply | 0 | 31.246 |
| reference | 0 | 10.32 |
| slice | 0 | 4.197 |
| scanner-0 | 1 | 7.457 |
| full-tree-check-0 | 0 | 0.215 |
| scanner-1 | 1 | 6.843 |
| full-tree-check-1 | 0 | 0.215 |

Records-only command total: 310.477s. nproc 5; cgroup cpu.max 400000 100000. Setup timing lines:

```text
setup: node ready (0.025s)
setup: go ready (0.033s)
setup: submodules ready (0.091s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: markdown dependencies ready (0.101s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.245s)
setup: go build ready (196.465s)
setup: test binaries deferred (use --warm-tests) (196.662s)
setup: build cache warm (196.664s)
setup: build-flags commit=0e38f5dc59972303f1380c5de5d9250a2f9b88ac nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.51 0.18 0.06 1/222 41865 load-after=7.87 4.24 1.70 1/226 43371
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (196.695s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.Ohy7SM
```

## Ordered stops and suggested owners

Every listed witness runs on stock TypeScript 6.0.3/Node with exit 0 and fails
on this compiler with exit 1 and the matching diagnostic. Owners are technical
routing suggestions inferred from the diagnostic, not verified task assignments.
Only stop 1 is unchanged scanner evidence. Every later observation depends on
private throwing placeholders, with shifted positions in those copies.

| Order | File:line:column | Diagnostic | Suggested owner |
| --- | --- | --- | --- |
| 1 | debug.ts:8:5 | stage 0 can't lower a mutable namespace export; use a module or export functions around private state yet | area/compiler: namespace state lowering |
| 2 | debug.ts:14:14 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | area/compiler: cast proof; stage3 adaptation: remove an unchecked cast |
| 3 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | area/compiler: cast proof; stage3 adaptation: diagnostic factory typing |
| 4 | utilities.ts:65:22 | stage 0 can't lower typed array element type Uint16Array yet | area/compiler and area/library: typed-array representation |
| 5 | scanner.ts:3510:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | area/compiler: cast proof; stage3 adaptation: UTF16 worker typing |
| 6 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | area/compiler: aliased array constructors |
| 7 | utilities.ts:13:17 | stage 0 can't lower a function returning U &#124; undefined yet | area/compiler: generic nullable returns |
| 8 | scanner.ts:696:32 | stage 0 can't lower a BinaryExpression with a string and a number yet | area/compiler: mixed string/number addition |
| 9 | scanner.ts:687:15 | stage 0 can't lower destructuring a string yet | area/compiler: string destructuring |
| 10 | scanner.ts:1221:26 | stage 0 can't lower a BinaryExpression with a string and a number yet | area/compiler: mixed string/number addition |
| 11 | scanner.ts:1676:62 | stage 0 can't lower a PostfixUnaryExpression yet | area/compiler: postfix values |
| 12 | debug.ts:15:74 | stage 0 can't lower a generic function as a value yet | area/compiler: generic function values |
| 13 | scanner.ts:1826:22 | stage 0 can't lower a call through ?. (an optional call) yet | area/compiler: optional calls |
| 14 | core.ts:81:29 | stage 0 can't lower for...of over a union of differently held members yet | area/compiler: union iteration representations |
| 15 | scanner.ts:2296:130 | stage 0 can't lower a generic function as a value yet | area/compiler: generic function values |

The command's built-in catalogue stops at order 8. Supplementary private copies
add fresh witnesses for mixed addition, object-style string length destructuring,
postfix values, generic function values, optional calls and differently held
iterables, then continue to fifteen total stops. They use the existing throwing
function stub helper and retain every body/signature change. The source helpers
are committed as measurement-*.py; raw-logs.json.gz retains every candidate,
Node/build output and witness source. Failed candidates caught unrelated TS2345
console declarations and were discarded. A mixed-element-array loop candidate
built but did not reproduce the requested diagnostic, so it was not used as a
witness or claimed as a correctness control. All private changes are in
private-placeholders.json.gz. These copies never establish a native scanner pass.

## Provisional historical marks

The exact #whkxbc7 text was not supplied, found in the working tree, or found
in fetched commit messages. A clarification was requested while the run continued.
The table below uses the fifteen-item land-area-next/stops.json list already
referenced by the user's prior stop 10/12/14 instructions. Its identity with
#whkxbc7 is not verified; historical-marks.json explicitly records that limit.
Suggested owners do not replace actual task assignments.

| Historical order | Historical checkpoint | Mark | Owner |
| --- | --- | --- | --- |
| 1 | corePublic.ts:9:5 | isolated admission control now passes and matches Node; output mutant caught; no whole-scanner correctness claim | area/compiler: records |
| 2 | debug.ts:8:5 | remains: unchanged first stop | area/compiler: namespaces |
| 3 | debug.ts:15:14 | remains: current ordered stop 2; behind placeholders | area/compiler / stage3 adaptation: cast proof and source typing |
| 4 | diagnosticInformationMap.generated.ts:13:34 | remains: current ordered stop 3; behind placeholders | area/compiler / stage3 adaptation: cast proof and source typing |
| 5 | utilities.ts:65:22 | remains: current ordered stop 4; behind placeholders | area/compiler and area/library: typed arrays |
| 6 | scanner.ts:1269:28 | old isolated probe still blocked; original scanner checkpoint not established by this traversal | area/compiler: enum proof |
| 7 | scanner.ts:3497:67 | remains: current ordered stop 5; behind placeholders | area/compiler / stage3 adaptation: UTF16 typing |
| 8 | core.ts:107:20 | remains: current ordered stop 6; behind placeholders | area/compiler: aliased constructors |
| 9 | utilities.ts:13:17 | remains: current ordered stop 7; behind placeholders | area/compiler: nullable generic returns |
| 10 | scanner.ts:608:87 | scanner site adapted in prior accepted units; the original any probe remains intentionally refused | stage3/adapt/42-scanner-any |
| 11 | scanner.ts:113:5 | isolated admission control now passes and matches Node; output mutant caught; no whole-scanner correctness claim | area/compiler: computed fields |
| 12 | scanner.ts:100:23 | scanner site adapted in prior accepted units; the original any probe remains intentionally refused | stage3/adapt/42-scanner-any |
| 13 | scanner.ts:432:59 | old isolated probe still blocked; original scanner checkpoint not established by this traversal | area/compiler and area/library: entries on named records |
| 14 | driver/main.a:13:42 | scanner site adapted in prior accepted units; the original any probe remains intentionally refused | stage3/drivers/scanner |
| 15 | generated/main.c:7923:9 | old isolated probe still blocked; original scanner checkpoint not established by this traversal | area/compiler: never/string emission |

All fifteen old witnesses were rerun unchanged from their archived source on
this compiler. Only probes 1 and 11 admit and match Node; their byte mutants
fail diff with exit 1. They print admission markers, so this establishes admission
and output-control sensitivity, not complete computed-field runtime semantics.
The remaining probes still fail, with complete diagnostics in historical-controls.json.
The original any probes 10/12/14 intentionally remain refused: adaptation 42 and
the driver's real callback type previously removed those scanner sites, rather
than teaching the compiler to accept any. No unvisited checkpoint is marked
closed merely because it did not appear in this bounded traversal.

## Checks and coverage limits

```sh
python3 -B stage3/drivers/scanner/test_scratch_runner.py > stage3/drivers/scanner/evidence/main-area-next-replay/tests.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/main-area-next-replay/conflict-summary.json > stage3/drivers/scanner/evidence/main-area-next-replay/conflict-check.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/main-area-next-replay/records-summary.json > stage3/drivers/scanner/evidence/main-area-next-replay/records-check.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/main-area-next-replay/conflict-pass-mutant.json > stage3/drivers/scanner/evidence/main-area-next-replay/conflict-mutant-check.log 2>&1
```

Five targeted tests pass; both real command summaries validate. The false-pass
mutant retains the failed library merge and 56 conflict paths, supplies otherwise
complete successful compiler/scanner claims, and is rejected with exit 1 solely
by the merge-conflict rules. Both scanner modes and the full-tree reference catch
their token-end mutants. Both admitted historical controls catch their one-byte
native-output mutants. No native scanner output mutant is claimed.
No broad packages or full gate run; no oracle fixture added, counts.md unchanged.

Space was recovered only from this worker's prior completed scratch worktrees
and derived source copies; prior raw token streams were losslessly compressed.
Logs, binaries, summaries and scratch branch refs are preserved. The recovery
manifest remains /tmp/scanner-main-next-space-recovery.json. This unit's full
outputs and scratch worktree remain under /workspace/scratch/scanner-main-next-records.
No library integration, exact #whkxbc7 mapping or whole native scanner correctness
is claimed. The scanner comparison remains blocked at the unchanged first stop.
