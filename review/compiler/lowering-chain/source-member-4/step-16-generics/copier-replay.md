Cherry-picked the census copier repair d2651d95 as 9a02e97c, without conflicts.
The selected generic source replay now reproduces NotYet at core.ts:220:62, a value of type T.
Selected-unit hidden intersection: 379 bytes before, 224 after, 155 bytes revealed to the census, 0 newly hidden.
Cache rollback proof passes; removed-cache-entry and skip-all-private mutants are caught; five hidden-arithmetic mutants are caught.
The contains body remains hidden; no compiler implementation or native acceptance gain, and no whole-corpus revealed-byte total is claimed.

The earlier landing report's generic replay blocker is now cleared. Its original
record remains historical. The exact prior selector (core.ts:220:1, reason
'a generic function as a value') now exits 1 for a signature mismatch, rather
than a copier panic. Its new diagnostic is:

```text
/tmp/generics-scout-adapted/src/compiler/core.ts:220:62: stage 0 can't lower a value of type T yet
```

The corrected selector positively reproduces it, exit 0:

```sh
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay \
  -project /tmp/generics-scout-adapted/src/compiler \
  -where /tmp/generics-scout-adapted/src/compiler/core.ts:220:62 \
  -kind NotYet -reason 'a value of type T' \
  > /tmp/generics-copier-reproduced.log 2>&1
```

This is an independently attempted, unspecialized contains declaration in a
checker-rejected corpus. The stop is its value: T parameter. It does not prove
that a concrete call of contains would fail there. The executable generic
reductions retain their previous outcomes; no generic body-relations code is
added. compiler/generic-body-relations remains the separate implementation.

## Selected-unit byte measurement

[The machine record](copier-replay.json) identifies the source SHA-256, stock AST,
compiler and helper pins, exact diagnostics, and byte ranges. Both replays use
the same persistent adaptation. The selected declaration and body spans are
identical in their records and agree with stock TypeScript 6.0.3.

| Selected UTF-8 range | Before | After |
| --- | --- | --- |
| Declaration [7486, 7865) | 379 bytes | 379 bytes |
| Failed boundary / residual hidden range | [7486, 7865), 379 bytes | [7641, 7865), 224 bytes |
| Independently examined declaration prefix | none | [7486, 7641), 155 bytes |

The existing hidden.py union/subtraction operations from ec0b16c0 and a separate
byte mask agree on 155 bytes removed from the selected hidden intersection, and
zero added. That prefix includes declaration trivia and signature before the
body. It is census examination credit under the existing definition, not proof
that those bytes compile. The remaining 224-byte body is hidden behind the new
signature failure. Other declarations, enclosing failures and the whole compiler
corpus were not remeasured. The imported d72728e5 hidden totals are that commit's
historical evidence, not this branch's totals.

## Checks and logs

make_overlay.py generated /tmp/generics-copier-overlay.
cache_copy_audit.py ran its focused overlay tests, exit 0. Its baseline proves
actual failed-initializer rollback, subsequent lowering, recomputed packed facts
matching a fresh program, and independent original state. Removing the one
allowlist entry fails with the argumentFacts panic; skipping all private fields
fails with 'unreviewed private field accepted'. No compile failure is credited.

The unchanged hidden arithmetic tests from ec0b16c0 pass. Independently running
ignore-dependency, ignore-skipped, double-count, forget-checker-child and
forget-independent makes each test run exit 1 at an assertion. The stock AST
manifest was generated with that commit's units.cjs, using TypeScript 6.0.3.

Logs:

- /tmp/generics-copier-cherry-pick.log
- /tmp/generics-copier-proof.log and /tmp/generics-copier-proof/*.log.txt
- /tmp/generics-copier-replay.log (original selector's expected mismatch)
- /tmp/generics-copier-reproduced.log (new selector reproduced, exit 0)
- /tmp/generics-copier-hidden-tools/stock.log, baseline.log, and mutant logs

The repair changes measurement-only census code and imports its hidden evidence.
No new .a fixture or ownership-count row is introduced. Production lowering and
runtime code are unchanged; the prior fixture/backend/sanitizer and counts sweep
remains applicable. No main or area branch is pushed.
