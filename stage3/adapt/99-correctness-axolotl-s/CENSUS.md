# 99s initial compiler census (historical)

The queue amendment has separate current evidence in [QUEUE-BOX-PROOF.md](QUEUE-BOX-PROOF.md).

These are actual frozen-main measurements, not a successful lowering
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

The meter-only filterMutate mutant restores just `array.length = outIndex` on
the complete accepted source snapshot. Its body remains deferred, so that run
does not prove it compiler-red. meter-mutant-result.json preserves that result.
The follow-up authorized direct concrete probes instead. Both splice programs
lower; both restored controls are NotYet. direct-mutant-result.json retains
the exact diagnostics and source hashes. No finding count substitutes for this
direct proof.

Current main's internal/lower/class.go:471 and :706 call l.notYet for a field
assignment whose object has this representation. The ruling in docs/rulings.md:285 and :439 is Refused, with enforcement
owed by #53r37a3. Adamic explicitly authorized building these splices while
recording main's current NotYet classification. No production compiler change
or measurement relabeling is included.

The complete ordinary per-file census measured all 79 TypeScript compiler roots
on the accepted stack, using unmodified frozen-main stage3/census/tool. The
full-applied source and measured source are byte-identical across all 746 files;
snapshot-comparison.json records the check. All raw rows are retained. The three
non-TypeScript entries and the ordinary whole-program row are excluded from the
79-root table, exactly as in 99q.

| Roots | Kind | First stop | Reason |
| ---: | --- | --- | --- |
| 73 | Refused | `src/compiler/core.ts:736:36` | a cast the runtime can't check |
| 2 | accepted | - | - |
| 1 | NotYet | `src/compiler/commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `src/compiler/debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `src/compiler/diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `src/compiler/semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

census-results.json retains each exact root, kind, location and reason. No root
is checker-rejected. Census timing includes a deliberate pause of program.ts
while both full oracles and lane Go ran with the full CPU quota. No throughput
claim is made from those wall times. Historical partial rows from the stopped
attempt remain under the initial- prefix.

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
