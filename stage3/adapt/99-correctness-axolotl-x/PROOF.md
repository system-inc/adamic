# Current continuation: baseline equality is the only replay condition

The user superseded both earlier replay conditions. A shared refusal is a
baseline outcome regardless of file or owner names. 40-explicit-any's builder.ts
brand refusal is a recorded baseline, not a check failure. The paired replay
has resumed after its eight matching outcomes and is comparing every remaining
adapter; only a difference between the trees stops it. Earlier stopping notes
below are historical.

Check 2 has now passed: adamic types received all 79 sorted compiler roots
from /workspace/bind-after in one invocation, exited 0 in 3.650 seconds and
reported zero diagnostics. The complete type listing is types-after.log.gz;
arguments, root count, timing and exit are in types-after.json. All 46 sites
remain rewritten, zero are left as written, and no cast was added.

# 99x proof: revised check 1 stops at the affected-file guard

The user replaced the original zero-refusal replay rule with a paired replay
comparison. The new records are identical through 40-explicit-any, but both
refusals name builder.ts, one of the six files covered by the revised rule's
affected-file condition. That condition fails. The run stopped there; checks
2 through 6 remain unrun. No new 99x regression is established by this result.
The adapter is implemented, but the requested verification remains incomplete.

## Results

| Check | Result |
| --- | --- |
| 1 baseline full apply, --build-product | Pass, exit 0 |
| 1 after full apply, --build-product | Pass, exit 0; 46 rewrites, zero left, six files |
| 99x owner contract | Pass: six refused malformed states; foreign edit accepted |
| 99x replay on actual after tree | Zero files and edits; bytes and nanosecond mtimes preserved across 79 roots |
| CRLF contract | Pass |
| 1 original ordered replay | Superseded by the user |
| 1 revised paired replay comparison | Eight adapters compared, identical records, empty diff |
| 1 revised refusal file/owner guard | Failed at 40-explicit-any: both refusals name builder.ts |
| 2 Adamic types, all 79 roots | Not run: stopped at check 1 |
| 3 nine emitted JavaScript files | Not run: stopped at check 1; no emitted diff hunks |
| 4 upstream oracle and shards | Not run: stopped at check 1 |
| 5 restored-site negative witness | Fixture and patch prepared; detection not run |
| 6 latent and ordinary per-file census | Not run: stopped at check 1 |

## Historical stop under the original rule

The first three adapters, 00-setup, 10-type-imports and
20-optional-declarations, exited 0 with no byte or mtime changes. The next
adapter, 30-indexed-reads, exited 1 with:

```text
declined site changed: core.ts:992:21
```

Its sites.json row names addRange, expression from[i], occurrence 2 of 2,
action decline. On the completed after tree the expression is at core.ts:1002:21
and its parent is from[i]!. The stop is in addRange, which is outside this
adaptation's 15 owners. It is a declined-site guard, not an overlapping owner
pin refusal. No other adaptation directory was edited.

core.ts is byte-identical between the successful main baseline and the after
tree. Its SHA-256 is
3dde48b17a7ea7c9475f28b0dfa1b8c3a407f6b4966c4f6f6c4e6769e4920f3e.
That hash is measurement evidence only, never an adaptation precondition.
replay-stop-detail.json records the complete row, actual AST parent and owner.
ordered-replay/results.json records exits and changed-file lists; per-adapter
logs and the driver log are retained alongside it. The failure made zero writes.

## Commands and provenance

The base is unmodified main 3e395f9e5a7038d859ec8930be72cde68d8be3ff. An isolated worktree at
/workspace/bind-main provided the baseline adaptation set, excluding this
branch's initial evidence-only directory. The branch was never based on 99w.
99w's adapt.cjs, README.md and PROOF.md were read at
06dad75d0bb482b47046370a60341c054899afed. 99k's adapter and documentation were
read from main. No Apple compiler overlay was used.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/bind-setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/bind-npm-api.log 2>&1
# Setup does not supply the upstream source checkout; clone the authorized pin.
git clone --filter=blob:none --no-checkout https://github.com/microsoft/TypeScript.git /workspace/typescript-source
git -C /workspace/typescript-source checkout 050880ce59e30b356b686bd3144efe24f875ebc8
# Run from /workspace/bind-main for the baseline:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
# Run from this branch for the after tree:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/generate.cjs /workspace/bind-before /workspace/typescript-source
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/contract.cjs /workspace/bind-before
python3 stage3/adapt/99-correctness-axolotl-x/replay.py /workspace/adamic /workspace/bind-after /workspace/scratch/bind-ordered-replay
```

Pin generation ran after baseline apply and before after apply. The command
list groups apply commands together for readability. All check outputs went
to files. Full apply logs are baseline-apply.log.gz and after-apply.log.gz.
Setup, npm, pin generation, pristine symbol scan, contract and actual-tree
self replay each have their own compressed logs. The self replay checks all
79 compiler source files' SHA-256 and nanosecond mtime before and after calling
adapt on the actual after tree. An unmodified main Adamic and ordinary census
tool were built successfully, but no type or lowering measurement ran.

The initial missing-source stop remains in check1-before.log and
prerequisites.log as historical evidence, superseded by this continuation's
successful full applies. Completed evidence commits change product keys
without changing the frozen transformation. All commits touch only this
directory, and all pushes use typescript/maybebind-adapt without force.

## CW051 before and after

The six unit identifiers below come from typescript/core-walls-sort at
ad29102e9c145c5b480629227574d2ffa235a4b1. That reference records their shared
stop as core.ts:1523:12, a call through ?. (an optional call). Those historical
observations are not new baseline measurements. The container-fit selector
reference was inspected at 654092b95416580c3e5be1f968fb8b14755b4481.

| Recorded unit | New before run | New after run |
| --- | --- | --- |
| src/compiler/checker.ts:54277:1 | Not run | Not run |
| src/compiler/watch.ts:673:1 | Not run | Not run |
| src/compiler/watchPublic.ts:123:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:812:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:1011:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:1034:1 | Not run | Not run |

The ordinary 79-root census also did not run; no changed-first-stop result
is claimed. The negative fixture restores one builder.ts call and its import,
but its JavaScript comparison did not run. No Go file was added. All processes
started for this task have exited; no background work remains.

## Revised check 1: paired replay

The requested replacement condition compares adapter outcomes between main
and main plus 99x, excluding 99x itself, and requires that no refusal name any
of 99x's six source files or its named owners. paired_replay.py implements
both conditions and stops at their first failure.

Independent copies preserve the successful full-apply trees for future work:

```sh
cp -a --reflink=auto /workspace/bind-before /workspace/bind-replay-before
cp -a --reflink=auto /workspace/bind-after /workspace/bind-replay-after
python3 stage3/adapt/99-correctness-axolotl-x/paired_replay.py /workspace/adamic /workspace/bind-replay-before /workspace/bind-replay-after /workspace/scratch/bind-paired-replay > /workspace/scratch/bind-paired-replay.log 2>&1
```

For each adapter, snapshots record the actual files written, including writes
that preserve bytes. Text edits are recorded as affected files and added/removed
line counts; complete patches are retained if any edit occurs. Neither tree
had any writes or edits during these eight adapters. A nonzero exit records
the complete refusal output, with only the tree root normalized to <tree>.
Raw per-adapter logs are retained for both sides. Failures do not stop traversal
unless records differ or the refusal names a covered file or owner.

| Adapter | Before | After | Equal |
| --- | --- | --- | --- |
| 00-setup | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 10-type-imports | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 20-optional-declarations | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 30-indexed-reads | Refuses core.ts:992:21 | Same refusal | Yes |
| 31-indexed-reads-checker | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 32-indexed-reads-program | Refuses appendSourceMap runtime drift | Same refusal | Yes |
| 33-indexed-reads-emit | Refuses sourcemap.ts:215 site drift | Same refusal | Yes |
| 40-explicit-any | Refuses missing builder.ts brands | Same refusal | Yes |

The two records are paired-replay/before.json and paired-replay/after.json.
Their exact diff, paired-replay/records.diff, is empty (zero bytes).
paired-replay/status.json records the failed affected-file condition.
The full shared refusal at 40-explicit-any includes:

```text
Error: missing reviewed brands: src/compiler/builder.ts
```

40's classes.cjs matches brands by key and absolute line. The relevant owners
are IncrementalBuildInfoFileId and IncrementalBuildInfoFileIdListId, neither
of which is one of 99x's named function owners. The brand declarations are
already typed undefined in both trees. Its reviewed lines are 1074 and 1076;
main's actual lines are 1098 and 1100, and after's are 1097 and 1099. The one-line
after shift is the removed maybeBind import. Both trees miss the original
line pins, so this is a shared baseline failure. The rule's file-name condition
still fails because it names builder.ts. No other adaptation was changed.
paired-replay-stop-detail.json retains keys, owners and actual/pinned lines.

The source comparison in source-delta.json confirms that the original complete
src trees differ in exactly the intended six files, with core.ts unchanged.
The revised run did not reach 99x or later adapters. 99x's direct no-write
replay and contracts passed in the earlier completed piece. No types, emitted
JavaScript, oracle, negative detection or census measurements were added in
this continuation. The six CW051 rows above remain unmeasured. All processes
started for this continuation have exited.
