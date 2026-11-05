# 99s site census and partial ordinary census

STOPPED. These are actual current-main measurements, not a successful lowering
claim. Main is 82b0c076395636800764ed1af7ada569681fc2ae. The ordinary binary
has no overlay. The separate full site census uses the meter's guarded overlay,
LATENT_FULL=1, LATENT_ONLY_FILE set to the complete core.ts absolute path and
LATENT_ASSERT_NO_OUTPUT=1. Both full site streams contain zero checker diagnostics.
The overlay's historical measurement label says checker-rejected even with an
empty diagnostics list; the raw label is preserved rather than rewritten.

| Edited site | Before | After |
| --- | --- | --- |
| filterMutate, core.ts:307 | Owner at 299:1: deferred_generic_instantiation, instantiated=false. No body finding at 307. | Same deferred status. No claim that truncation was examined. |
| clear, core.ts:312:5 | **NotYet**, assigning a field of a value. Boundary records rollback. | **NotYet**, an array of unknown with erased element storage (retain its declared element type before reading elements). Boundary records rollback. |
| createQueue.dequeue truncation, core.ts:1596 | createQueue at 1569:1 and dequeue at 1581:5 are deferred_generic_instantiation, instantiated=false. | Same deferred status. No claim that truncation was examined. |
| skipWhile, core.ts:2601:16 | **Refused**, a cast the runtime can't check, refusal_scan, unit 2594:1. | Exact cast finding absent. Its generic body still has deferred_generic_instantiation status. |
| Unbuilt dequeue cast, core.ts:1586:24 | **Refused**, a cast the runtime can't check. | Unchanged **Refused** finding. |

Raw records, including all other core.ts findings, are in site-before.jsonl.gz,
site-after.jsonl.gz and mutant.jsonl.gz. site-summary.json retains exact locations,
reasons and the selected unit coverage records. The unchanged takeWhile cast at
2585:16 must not be counted as skipWhile.

The required filterMutate mutant restores only `array.length = outIndex` on the
complete accepted source snapshot. Its body remains deferred, so this run does
not prove it compiler-red. mutant-result.json explicitly records NOT_PROVEN_RED.
Do not infer a successful mutation test from an unchanged total finding count.

Current main's internal/lower/class.go:471 and :706 call l.notYet for a field
assignment whose object has this representation. This conflicts with the task's
required Refused classification. No production compiler change or measurement
relabeling is included.

An ordinary per-file census was started on all 79 roots in the complete accepted
source snapshot, with the unmodified main stage3/census/tool binary. It was
terminated when the delivery contract could not be met. census.jsonl.gz and
census.log.gz preserve only completed rows. census-partial-summary.json counts
them and explicitly marks the census incomplete. They are not a 79-root result.

Reproduction, after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage3/census/latent/make_overlay.py /workspace/adamic-main /workspace/scratch/s-overlay > /workspace/scratch/s-overlay.log 2>&1
gofmt -w /workspace/scratch/s-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/s-overlay/overlay.json -o /workspace/scratch/s-bin/latent ./stage3/census/latent/tool > /workspace/scratch/s-build-latent.log 2>&1
LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/s-before/src/compiler/core.ts LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/s-bin/latent /workspace/s-before/src/compiler /workspace/scratch/s-site-before.jsonl > /workspace/scratch/s-site-before.log 2>&1
LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/s-probe/src/compiler/core.ts LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/s-bin/latent /workspace/s-probe/src/compiler /workspace/scratch/s-site-after.jsonl > /workspace/scratch/s-site-after.log 2>&1
```

The site and ordinary census binaries are separate. The no-output measurement
binary is never used to claim successful compilation.
