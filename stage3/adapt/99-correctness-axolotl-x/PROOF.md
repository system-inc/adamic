# 99x proof: stopped at check 1

The task's stop rule applies: ordered replay refused in 30-indexed-reads.
No later adapter or check was run. The adapter is implemented, but the full
requested verification is incomplete and this branch is not ready to land.

## Results

| Check | Result |
| --- | --- |
| 1 baseline full apply, --build-product | Pass, exit 0 |
| 1 after full apply, --build-product | Pass, exit 0; 46 rewrites, zero left, six files |
| 99x owner contract | Pass: six refused malformed states; foreign edit accepted |
| 99x replay on actual after tree | Zero files and edits; bytes and nanosecond mtimes preserved across 79 roots |
| CRLF contract | Pass |
| 1 ordered replay | Failed at 30-indexed-reads, exit 1, zero writes |
| 2 Adamic types, all 79 roots | Not run: stopped at check 1 |
| 3 nine emitted JavaScript files | Not run: stopped at check 1; no emitted diff hunks |
| 4 upstream oracle and shards | Not run: stopped at check 1 |
| 5 restored-site negative witness | Fixture and patch prepared; detection not run |
| 6 latent and ordinary per-file census | Not run: stopped at check 1 |

## Exact stop

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
