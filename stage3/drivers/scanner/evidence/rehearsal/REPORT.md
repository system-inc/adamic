First scanner stop: corePublic.ts:9:5, index signature, split 0 and 1.
Rehearsal only, pending, not a milestone; no native scanner binary or byte comparison.
Main 54cbc125 plus area-stack 64c59784 and placeholder-nonnull 71131229 build; six requested merges conflict and are skipped.
Four isolated controls match Node and catch their one-byte mutants; both Node scanner token-end mutants are caught.
Exact scanner checkpoints behind earlier blockers remain pending; the whole scratch branch never lands.

## Stack and commands

The configured fetch was initially narrow; an explicit all-heads fetch found the
requested area-stack tip and placeholder topic. library/area-on-next-2 was not
advertised, so area/library was tried. Topic search found all five requested
compiler topics. All SHAs and exact conflict paths are in stack.json and the
per-topic *-conflicts.txt files. No conflict was hand-resolved.

The existing runner stops on a conflicting merge. measurement-preflight.py
therefore merges the requested refs in order in a never-pushed scratch,
aborts each conflicting merge and continues. It passes only accepted pinned
SHAs, in the same order, to the unchanged scratch-run.sh:

```sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-rehearsal-retry 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8 64c59784b61239a6429a4aa02478b8bb70fe17c3 711312292569ca30e0ac075f600da9bdba926def > /tmp/scanner-rehearsal-retry.log 2>&1
python3 -B /tmp/scanner-rehearsal-controls.py > /tmp/scanner-rehearsal-controls.log 2>&1
```

| Requested ref | SHA | Result |
| --- | --- | --- |
| origin/compiler/area-stack | 64c59784b61239a6429a4aa02478b8bb70fe17c3 | merged |
| origin/codex/records-maplike-next | 8c013f1cf3ec69fe80e60501740c2c45e1d1114c | skipped, 3 conflicting paths |
| origin/area/library | 7778b3606a22ef53a181f7c1508428a6491e781b | skipped, 61 conflicting paths |
| origin/codex/placeholder-nonnull | 711312292569ca30e0ac075f600da9bdba926def | merged |
| origin/codex/entries-provenance | 05a6006690f4e69b270b17f00ccff8afef59fee9 | skipped, 4 conflicting paths |
| origin/codex/assignment-proofs | 4d17f1d4d8206111171c4a8ff54fc3da8beb7a47 | skipped, 2 conflicting paths |
| origin/codex/scanner-cast-checks | 77ed4464215592f8f95354c8f7400bac31de2f53 | skipped, 60 conflicting paths |
| origin/codex/stricter-options-next | 8f32e51e8fc41b8f1177453213ca5453ce764486 | skipped, 12 conflicting paths |

Compiler scratch SHA: `ceb93620339d8d63e2df0a0420d4b2a92998740e`. Binary SHA-256:
`5cd341acf6de4e1834e7dca23f1a893ef84f8106bd800a2b447245923fe959cd`. Both scratch branches remain unpushed.

## Pending evidence block

summary.json explicitly records rehearsal, pending, milestone=false and
whole_rehearsal_branch_lands=false. It links the actual runner result rather
than promoting an isolated admission control or a stub-dependent observation.

Both Node slice dumps match the full-tree reference byte for byte: 509,014
skip-trivia tokens, 860,418 retained-trivia tokens, 466 errors, 81 files.
Both comparison controls pass and both token-end mutants fail diff with exit 1.
There is no native scanner runtime, scanner output mutant or scanner timing.

Setup timings and nproc are retained verbatim in runner-summary.json and
raw-logs.json.gz. Setup sources /workspace/adamic-tools/env.sh; nproc=5,
cgroup quota=4. Measured wall seconds:

| Phase | Seconds |
| --- | --- |
| setup | 183.122 |
| compiler-build | 5.873 |
| apply | 28.507 |
| reference | 10.116 |
| slice | 3.644 |
| scanner-0 | 6.748 |
| scanner-1 | 6.997 |

Runner total: 273.621s.

## #whkxbc7 marks

Owners are the user's supplied assignments. Gone is scoped to the stated
control; pending names the scanner proof it awaits. No unvisited site is
retired just because it did not stop this bounded traversal.

| Checkpoint | Owner | Mark | Evidence / awaits |
| --- | --- | --- | --- |
| corePublic.ts:9:5 | records-maplike #rhjc4x4 | still there | Actual first stop in both modes; records skipped after 3 conflicts. |
| debug.ts:8:5 | compiler area-stack group 1 | gone in admission control; pending scanner verification | 02 passes, Node/native ok; awaits clean records integration and actual namespace state execution. |
| debug.ts:15:14 | compiler #b5w3ycg | still there | Ordered stop 2 at debug.ts:14:14; casts topic skipped. |
| diagnosticInformationMap.generated.ts:13:34 | compiler #b5w3ycg | still there | Ordered stop 3; casts topic skipped. |
| scanner.ts:3497:67 | compiler #b5w3ycg | still there | Ordered stop 5 at scanner.ts:3510:67; casts topic skipped. |
| utilities.ts:65:22 | runtime #4gkdjsz | still there | Ordered stop 4, Uint16Array; awaits runtime support. |
| scanner.ts:1269:28 | compiler #cvhj5fk | still there in isolated control | 06 refused; actual scanner checkpoint not reached. Assignment-proofs skipped. |
| core.ts:107:20 | compiler stricter options #k881crd | still there | Ordered stop 6; stricter-options topic skipped. |
| utilities.ts:13:17 | compiler area-stack | gone in control; pending scanner verification | 09 executes generic return and matches Node x. Awaits real scanner traversal past earlier stops. |
| scanner.ts:113:5 | compiler area-stack | gone in admission control; pending scanner verification | 11 admits and prints ok; no field read is exercised. Awaits actual field behavior and full scanner. |
| scanner.ts:432:59 | compiler #d9eemrs | gone in control; pending scanner verification | 13 executes entries count and matches Node 1 despite skipped entries topic. Actual scanner record provenance awaits clean records integration and traversal. |
| scanner.ts:4097:24 | compiler #9wc5q5j | still there in isolated control | 17 Node prints 0; build refuses undefined! assertion. Awaits placeholder property initialization support; merged topic does not close this control. |
| Error.captureStackTrace | library #jj9z3qn | still there in isolated control | 16 Node prints ok; build cannot lower ErrorConstructor.captureStackTrace. Awaits clean library integration (61 conflicts). |

Compared with the preceding marking, namespace export, generic returns and
Object.entries controls newly pass; computed-field admission remains passing.
MapLike is blocked again on this stack because its requested merge conflicts.
The namespace/computed-field probes print admission markers; they do not test
complete mutable namespace or computed property runtime behavior.

## Ordered discovery and checks

runner-summary.json retains seven ordered observations. Stops 1-6 have fresh
Node-success/compiler-failure witnesses. Stop 7, scanner.ts:778:15 string
destructuring, stops the built-in catalogue. supplemental-witness.json supplies
a fresh matching Node-success/compiler-failure witness for it, without another
scanner walk. No claim of fifteen discovered stops is made.

All observations after stop 1 are pending behind throwing discovery placeholders.
They await real MapLike declarations, Debug failure handling, diagnostic factory,
Uint16Array/UTF16 helpers and Array constructor implementations replacing the
removed pieces, then an unmodified native scanner comparison. Their absence
or success against contracts alone does not close scanner behavior.

The 17 retained controls run freshly on Node and the rehearsal compiler.
Controls 02, 09, 11 and 13 build/run with native exit 0, diff 0 and exactly
one changed output byte caught by diff 1. Remaining controls refuse at build
with Node exit 0; original any probes remain intentionally refused and do not
undo adaptation 42. Controls 16 and 17 directly cover Error capture and the
property placeholder. controls.json contains sources and complete outputs.

The real runner summary validates. A false-pass summary with otherwise complete
successful compiler/scanner claims but these exact skipped conflicts is rejected
solely by the merge rules. No whole packages, full gate, oracle fixtures or
counts refresh were run. Compiler, adaptations and runner sources are unchanged.

The first attempt failed at dependency checkout with exit 128, no space left
on device, before setup/build. Its summary and raw diagnostics are retained.
The retry is the measurement above. Space recovery losslessly compressed prior
scanner token outputs and removed regenerable source/dependency/build copies
and completed compiler worktrees belonging to this worker. Prior logs, binary
identities and scratch refs remain. Recovery manifests are retained; the failed
first attempt's partial compiler worktree was removed after retaining its logs.
Large token artifacts remain in the current scratch; their hashes and byte
counts are in large-output-manifest.json. All other measurement logs, witness
sources are retained in raw-logs.json.gz; exact private source changes are
retained in private-placeholders.json.gz.
