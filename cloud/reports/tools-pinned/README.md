# Tools pinned at serve

Task 8xfsf4x. Based on origin/devtools/fast-gate 95eb1c8f. No deployment, candidate queue work, push-main edit, or main push.

Each watcher dispatch snapshots tools-good and a Unix-nanosecond serve timestamp under the promotion lock. The running/<pid> token retains that tools SHA. Box launchers use detached, SHA-specific worktrees which promotion never switches. Pool jobs retain --tools, add tools_sha and tools_served_at, and pass both to the runner in their env. fast.json retains those fields; a runner/launcher tools mismatch refuses the job.

Promotions append `<sha>\t<promoted-at Unix nanoseconds>\n` to tools-good-history.tsv beside tools-good, then atomically replace tools-good. History append and serve snapshot share tools-history.lock. Startup records the existing good tools; an external tools-good edit is recorded when observed, without inventing an earlier promotion time. Existing pre-upgrade records with no serve timestamp have no new historical proof.

## Integration's push-main check

Integration owns push-main; it is deliberately outside this change. For every record it accepts, require a tools-good-history.tsv row with SHA equal to record.tools_sha and promoted_at <= record.tools_served_at. Both timestamps are Unix nanoseconds. Do not require the SHA to equal today's tools-good. Missing serve evidence or a tools SHA promoted only after serve fails closed. Apply this to each additional gate record too, preserving its metadata.

The local helper expresses that rule:

```sh
python3 cloud/tools-history.py accept "$ADAMIC_FAST_GATE_WATCH_STATE" fast.json
```

A newly disqualified rule needs an input-hash-based rerun of the units it touches, not whole-candidate re-serve; integration owns that judgment. External Loom must forward job.env to the runner and retain runner fast.json. The job-writer and runner contracts are tested here; no live Loom deployment was exercised.

## Promotion-sensitive paths audited

* dispatch -> stopOlderRuns formerly stopped an older run of the same candidate SHA when a new run used different box tools, marked race-lost, and discarded its verdict. Removed completely. Two served runs on different tools now both retain their chance to finish.
* stage-canary scheduling -> yieldRaceForCanary formerly stopped a candidate's box race to make a promotion canary slot, marked race-lost, and let its pool route answer alone. It now waits for a slot.
* promoteStaged -> placeGoodTree formerly switched a shared tools-good-tree while launchers could still read it. Replaced by detached worktrees under tools-pinned/<sha>; promotion cannot change an in-flight launcher's source.
* A box-side HEAD push previously let dispatch's canaryToken become a candidate's tools token when not staging. Candidates now always capture tools-good once; a deployment canary can promote the new tools for future jobs only. Mac-only pushes likewise cannot change a served candidate's SHA.
* dispatch's pool path formerly read tools-good more than once for cancellation comparison, token, and --tools. It now uses one captured SHA and timestamp. A promotion cannot split the job and running token across two tools versions.
* tools-good changes and stopPromotedCanary can stop a redundant `canary/main :staged` only. The branch and token guards exclude candidates; this behavior remains.
* Reaping compares testedHead to the current canaryToken only in canary branches. A candidate's real green/red is not compared to current tools-good or toolsHead and is not invalidated by promotion. toolsHead changes adjust staging/barriers for future dispatch only.
* pool-job.sh has no promotion reader/cancellation path. It writes the explicit --tools into the job; its stale-verdict/cancel cleanup occurs on actual job submission, not on promotion. There is no new promotion-triggered submission.

Other stop paths were inspected and retained for independent causes: stopGate is the backend cancellation primitive; reapStopped terminates already-stopped waiters; endRace cancels a losing route after the other route answers; stopSkipped enforces explicit skip directives; stopOverCeiling handles runtime limits; stopSuperseded handles newer codex/devtools/area candidate tips; stopStaleRed handles a newer reserved-family candidate; preemptForStar handles star resource priority. Pool/box voids, missing verdicts, ceilings and star preemption can still requeue for their existing reasons. None is triggered by tools promotion.

## Proof

All shell commands were bounded with timeout 900; test output went directly to log files.

* Watcher: `timeout 900 python3 -m unittest -v cloud/fast-gate/watch_test.py`, 79/79 passed ([watch.txt](watch.txt)). Box and pool promotion cases hold an old job, promote, serve a second job, then verify both finish green without cancellation/requeue and that records name old/new tools respectively. A same-SHA, different-tools regression also proves the old served run is not stopped.
* Runner: 174 isolated tests passed; with the first four history tests this run was 178/178 ([runner.txt](runner.txt)). Five PhaseInputs inventory tests were excluded from that isolated run after the full 183-test attempt encountered absent cohere/TypeScript/tsc/go.mod. The full attempt had one failure and three errors, all in inventory checks depending on that missing submodule ([inventory-limit.txt](inventory-limit.txt)); it is not a complete suite pass.
* History/pool writer: 5/5 passed ([history.txt](history.txt)), including append preservation, old-record acceptance, post-serve/unknown-tools rejection, malformed-history refusal, watcher pool metadata and direct pool serve timestamps.
* Mutant: scratch watcher reinserts a promotion loop which calls stopGate and marks served candidates preempted. `WATCH_TEST_SCRIPT=/tmp/tools-pinned-requeue-mutant.sh timeout 900 python3 -m unittest cloud.fast-gate.watch_test.WatchTests.test_promotion_keeps_served_candidates_and_their_tools` fails with the running candidate SHA unexpectedly present in stops ([promotion-mutant.txt](promotion-mutant.txt)). Production source was not replaced or committed with this mutant.
* History mutant: dropping promoted-at <= served-at admits a SHA promoted after serve and fails the old-served-record test ([history-mutant.txt](history-mutant.txt)).
* The roadmap refresh test exposed the known timing race: step-globs publishes before the same background refresh's --waves call, so it could snapshot calls too soon. Its test now waits for the --waves invocation from that refresh before checking no repeated refresh; the assertion is retained. The failed pre-synchronization run is saved in [watch-before-sync.txt](watch-before-sync.txt).
* bash -n on the three shell entrypoints and git diff --check pass.

Assumption: tools-good identifies the serve-authorized tools even after a Mac-only tools commit; the exact SHA, not merely a box-side content fingerprint, is retained. History timestamps establish no retrospective authority for records served before this feature was installed.
