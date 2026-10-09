# The merge queue

Ruled by Kirk and @system_adamic, Oct 9 19:35Z (#qth06pv): integration's merge queue replaces hand-cut trains. It
stacks candidates on main's tip, gates every stack at once, lands the longest green prefix, kicks a red member out
with its record, and rebuilds everything behind it. This file is its design; `merge-queue.py` beside it is the code.

## Members

A member is one line of integration's candidates file (`~/.adamic-integration/candidates`, the same file the gate
lane reads): landing branch, full sha, task id, infra names. `cut.sh` still makes the cut and adds the line. The
queue adds what the file lacks by reading the cut commit itself:

- `Gate-tier` from the cut's trailer (40 miscompile and P0, 30 the rest, 20 low).
- `kind`: `fast` for a change gated by the fast gate, `ruled` for one landed through `push-main --ruled-gate`
  (no product code, its own tests plus the census). A ruled member never rides a stack (see Ruled members).

Order: tier, highest first, then the order lines entered the file (a cut appends). A recut replaces its old line in
place, so it keeps its turn.

## Stacks

A stack is `main + A + B + ...`, built the way `speculate.sh` builds a prefix: a chain of merge commits on main's
tip made with `git merge-tree` and `git commit-tree`, so no checkout is touched. Stack k is stack k-1 plus member k.
Each stack's top merge carries the trailers the watcher already reads (`Gate-runs: deferred`, `Gate-tier` of its
highest member) plus `Queue-members: <sha8> <sha8> ...`, and is pushed as `cloud/land-queue-<k>-<sha8>`, so the
fast-gate watcher and Loom's pool gate it like any cut.

- Depth is capped at 3 (`QUEUE_DEPTH`). Each stack is a whole fast gate of every member's units until a unit's
  verdict can be kept by input hash (#pgnnb67, behind selection's 091e06f1), and the pool is the scarce thing.
- A member that conflicts with the prefix before it is not stacked. It goes back to its owner with the conflicting
  file names (`git merge-tree --name-only`), and the members behind it stack without it.
- A member whose tree adds nothing to the prefix (already an ancestor) leaves the file with a note on its task.

## States

Per member: `queued`, `stacked`, `kicked`, `landed`. Per stack: `gating`, `green`, `red`, `void`.

The state file is `~/.adamic-merge-queue/state.json`, written whole under the gate lane's lock (`fcntl` on
`~/.adamic-gate-lane/lock`, the same lock `cut.sh` takes, so a cut never races a rebuild):

```
{ "main": "<sha the stacks are built on>",
  "stacks": [ { "k": 1, "sha": "...", "branch": "cloud/land-queue-1-...", "members": ["<sha>"],
                "state": "gating", "record": null, "voids": 0, "submitted": "<utc>" } ],
  "kicked": [ { "sha": "...", "task": "...", "record": "gate-logs/...", "at": "<utc>" } ] }
```

## Verdicts

Each run reads every stack's newest finished record with the gate lane's own `records()` (fast, fast-phases or
full-main; a gate merge's records count as the stack's).

- **Void** (no verdict: a ceiling, a cancel, a pool death): requeued by recutting the same tree with a fresh empty
  commit, never read as a pass or a red. Three voids in a row on one stack go to Loom with the three record refs, and
  the stack waits.
- **Red**: if stack k is red and stack k-1 is green (or k is 1), member k is the cause. It is kicked: its line leaves
  the file, its record ref and first failure are posted on its task and sent to its owner. Every stack from k on is
  withdrawn (`~/.loom/bin/withdraw.sh <full sha>`) and rebuilt without it. A red whose every failing name is in the
  member's infra names (push-main's `--infra-red` class check) is not a kick; it lands as push-main would land it.
- **Green**: marks the prefix. A green stack whose prefix holds a red or gating stack waits for it.

## Landing

The longest green prefix lands through `push-main.sh --fast-gate <its record> <stack sha>`, the only judge, which
already refuses anything not complete, zerorun-clean and green, honors main's pause rule, and handles a main that
moved under the stack by records and tests only. Its landing commit names every member in `Branches`. Each landed
member's line leaves the file and its task gets the main sha. A push-main refusal is posted once per record and the
queue holds; exit 3 waits.

## Retest

After a landing or a kick, the remaining members are restacked on the new main and resubmitted; the old stack shas
are withdrawn. A main that moved by anything but the queue's own landing (a test-only lane landing, a ruled landing)
restacks too, but only when the next push-main would refuse on it: a moved test in a stacked member's package, or a
non-test commit. Test-only moves outside the members' packages are left to push-main's moved-main path, since
restacking on every one would starve the queue (main takes about 12 test-only landings an hour).

## Ruled members

A ruled member (gocacheprog, floor1, a setup.sh guard) has no product code and its gate is its own tests plus the
compiler dependency census of the merged tree. It is landed alone, ahead of the stacks, with
`push-main --ruled-gate "<ruling>"`, which runs `census-check.sh` on the landing tree since 471e035b. The stacks
then restack on the new main as for any main move.

## What Oct 9's hand work taught it

1. **Stale counts when stacked** (train 0ff31e41 red on `TestCountsAreRecorded`): rows right for each member alone go
   wrong together. A `counts.md` conflict is a stacking conflict, so the member goes back to its owner with the
   rows; a counts red in stack k with k-1 green kicks member k. The queue never regenerates counts or picks a side.
2. **Cuts going stale as main moves** (library 75473437 against `compiler-dependencies.json` and `fuzz_test.go`;
   the setup guard over the census fix): every stack is rebuilt on the new main after a landing, and a conflict
   goes to the owner with file names, never into a guess.
3. **Decided voids nobody requeued** (2c99ff09 and ab8989c3 sat at the pool ceiling until requeued by hand): every
   void is requeued by the queue itself, and the third in a row goes to Loom.
4. **The census a ruled gate skipped** (floor1 landed without it and reddened main fc7252a6): ruled members land
   only through `push-main --ruled-gate`, which runs the census on the merged tree.
5. **A known red still burning the pool** (after-chain c54a6d56 carried ba611349's lint link red): a member whose
   owner has a fix coming is withdrawn and its line held, not regated.

## Steps

`#dyre8as` stacks, `#g7hc1gj` verdicts, `#48d24b4` land, `#p3qhepw` retest, `#epjzksq` the launchd loop beside the
gate lane and the proof: three real candidates land their green prefix with no hand step, a planted red member is
kicked while the rest land, and candidates' submit-to-landing time falls. Until the loop lands, the gate lane keeps
landing single candidates from the same file; the queue replaces it when the proof passes.
