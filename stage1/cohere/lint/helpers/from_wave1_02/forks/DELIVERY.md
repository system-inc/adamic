Built snapshotForks and restoreForks for four syntax-rule consumers; eight more dependency entries removed, zero final blockers.
Current main c01907a70; refreshed helper gate c1e716a73; parked rule branch 702db4c9e; claims 1c6b3b003 and 28d8e360a; final pushed SHA accompanies report.
Checks PASS: 4,326 Go observations, 113,214 equal bytes on source Node, emitted JavaScript and sanitized native; setup PASS 100s, nproc 5.
Six new compiling semantic mutants caught: snapshot bit/order/buffer, restore bit/suffix/write order; alias mutants pass ordinary comparisons and fail the identity/lifetime comparisons.
Not covered: complete CFG, whole-rule findings/fixes or full repository gate; rule branch remains parked on the named shared registration blocker, no shared harness edits or regex matchers.

Both helpers have the same frozen readiness consumers:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

Each helper removes one dependency occurrence for each consumer, without making any consumer fully helper-ready. With the previously delivered markFinal and markThrown, this branch delivers four helpers, sixteen total dependency occurrences and zero final blockers. This is a conservative dependency statement, not a rule-port claim.

Original consumer suites and explicit consuming-rule probes are recorded separately. Each helper observes six calls in original suites (3, 2, 0, 1 respectively) and sixteen in probes (2, 4, 8, 2). The no-unreachable-loop original suite makes zero calls, so its added live Go-rule probe is necessary. All original tests still pass and no suite is skipped. Probe diagnostic IDs are logged but no claim is made that Adamic independently computes those full-rule findings.

Snapshot coverage: 84 ordinary calls plus 31 retained-snapshot observations, 1,505 total bytes. Snapshot controls enumerate all vectors through length four and retain two snapshots while changing frame state and the first snapshot, including empty vectors. Restore coverage: 3,959 unique-frame transitions and 252 shared-frame transitions, 111,709 total bytes. Restore controls enumerate all initial vectors through length four and snapshots through length six, plus repeated frame identities. The total for this pair is 4,326 observations and 113,214 bytes. Including the prior marker pair, all four helpers cover 7,432 observations and 211,466 bytes with twelve semantic mutants.

Commands, with output redirected directly to logs:

- python3 stage1/cohere/lint/helpers/from_wave1_02/forks/verify_snapshot.py
- python3 stage1/cohere/lint/helpers/from_wave1_02/forks/verify_restore.py

The per-helper reports are SNAPSHOT_REPORT.md and RESTORE_REPORT.md. Exact Go original-suite and probe commands, overlay generation, Node loader invocation, emitted JavaScript and sanitizer builder invocation are in those verifiers. Actual private Go implementations are unmodified; only scratch overlays add state logging and additional tests. The cohere worktree and shared lint harness are unchanged.

New mutants, all compiled and executed cleanly on three backends:

1. Snapshot inversion changes each copied boolean.
2. Snapshot reverse-order changes the ordered frame copy.
3. Snapshot buffer reuse clears and returns a process-global array. Ordinary pointwise output still matches all 864 bytes; the 641-byte retained-snapshot comparison alone detects sharing.
4. Restore inversion changes each assigned boolean.
5. Restore suffix clobbering iterates every frame and substitutes false when the snapshot lacks that position.
6. Restore reversed iteration preserves every unique-frame final value, matching 105,613 bytes, but changes shared-frame last-write behavior, failing the 6,096-byte alias comparison.

The checks hold storage lifetime and identity, not just independent boolean tables. Snapshot results are mutable fresh arrays, input frame stacks are readonly views over mutable nominal ForkFrame instances, and restoration preserves snapshot input and mutates only applicable frame flags. Go nil/empty slices and nonnil frame pointers have the explicit typed adapter described in the reports. No arbitrary proxy/getter semantics or full CFG construction is promised.

Landing first was applied before these claims: existing markers rebased cleanly on c01907a70 and all six existing mutants re-green at c1e716a73. The rule landing branch rebased cleanly, preserved main oracle PASS 30.010s (751,320 equal bytes), and named owned-witness missing-registration gate FAIL 2.550s. That branch is parked and pushed at 702db4c9e under system_adamic's explicit exception, awaiting the harness commit on #zmh9v36. No shared bridge or option-text changes were made. No main or area branch was pushed, and no incoming main change was reverted.

Selection audits read all origin claim blobs: snapshot claim after 559 refs/20 distinct contents; restore after 565 refs/20. Post-push audits found no competing earlier claim for either retained helper. Setup on this main base passed with tools/submodules 0s each, cache warm 100s, total 100s, nproc 5 and cpu.max 400000 100000. Initial runner never[] refusal is retained and corrected; it earns no mutant credit. No new Go regex exists in either helper, no handwritten matcher was added, and no node-kind string dispatch was introduced.

No additional helper is reserved. Rebase and land the parked rule branch before further claims when the harness commit is named.


## Current main refresh, 2026-10-07

Rebased all eighteen owned helper commits cleanly onto origin/main
b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Incoming inherited static field
compiler changes were retained. No shared harness file changed.

Reran final/verify.py, thrown/verify.py, verify_joint.py,
forks/verify_snapshot.py and forks/verify_restore.py, each writing directly
to its own log. All five verifiers PASS. Source Node, emitted JavaScript and
sanitized native match actual Go consumer observations: marker and joint
98,252 bytes plus fork pair 113,214 bytes, total 211,466 bytes from 7,432
observations. All twelve semantic mutants compile and execute cleanly and
are caught only by comparison: marker unreachable/final guards; thrown
unreachable/thrown guards and final flag; joint wrong destination list;
snapshot inverted bits, reversed order and reused storage; restore inverted
bits, suffix clobbering and reversed shared-frame writes. Buffer reuse still
passes the ordinary 864-byte comparison; the retained-snapshot comparison
alone fails. Reversed restore still passes unique-frame 105,613 bytes;
shared-frame identity comparison alone fails. Logs: ../evidence/main-b8fb957a-*.log.

The same four consumers remain: array-callback-return, consistent-return,
no-unreachable-loop and react-hooks/rules-of-hooks. Sixteen helper dependency
occurrences removed in total, zero whole rules newly helper-ready. Full CFG,
whole-rule findings/fixes, throughput and full repository gate remain outside
this proof. No regex exists in these helpers and no kind dispatch was added.

Rule landing branch codex/lint-wave1-02-land is pushed at 1cb4def8d with
current main integrated and named harness ab70f38d4 retained. Its renewed
TestOwnedWitnesses gate FAIL 3.168s: no-restricted-types defaults ban nothing
but the shared witness gate supplies no Types configuration. Previous shared
all-row option-isolation failure remains recorded there. The rule branch is
not newly green or landing-certified. No additional helper is claimed until
that landing gate is resolved. Main and area branches were not pushed.
