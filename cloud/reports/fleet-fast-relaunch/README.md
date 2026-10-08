# Fleet setup failure relaunch proof

Based on origin/area/developer-tools 358dbccd95a909a849d1d9db370ae7122a87fc5b.

Workers receive a self-contained publisher before checkout/setup. A failure publishes
failure.json (requested commit, shard, attempt and exact reason) and setup.log on
that run's shard branch. Success appends a normal commit containing the archive and
attempt.json, removing the marker. No source checkout is modified, no branch is
force-pushed, and old attempts cannot overwrite newer publications. Success must
match the full checkout commit; failure can be published even if checkout failed.

Every controller poll reads the marker. A current-attempt failure or reported
launcher error starts a fresh brief immediately, leaving healthy shards alone.
After a replacement starts the controller checks again immediately, then resumes
the default 30-second polling interval. Stale markers cannot consume retries.
Each replacement records shard, new attempt, reason and UTC time in notes.jsonl;
the merge brief carries those notes for inclusion in merge.tgz. If writing the
notes fails, the replacement does not start. Explicit merge and the merge brief
refuse markers, including a marker with a stale archive beside it.

Assumption: the cap means three failed attempts total, initial plus two
replacements, following the requested fail-three-times proof. It does not mean
three replacements after the first failure. This assumption is also in the commit.
An environment that never starts must be reported by the launcher (nonzero exit or
marker on its behalf). A launcher that reports success and then silently loses
its box still needs the missing-log deadline. Unreachable push credentials or
origin cannot deliver a marker. No live fleet or whole compiler gate was run.

## Measurements

Original and changed scripts were run interleaved on this same box, three times,
at the same checkout commit with scratch changes. The fake launcher uses a local
bare origin, publishes the marker through the exact publisher embedded in the
brief, and has the merger read the replacement archive's synthetic summary.
The original controller attempts merge with the marker-only branch; that fake
merge fails because the archive is absent. Its old 7,200-second missing-branch
wait is a code setting, not a measured two-hour run. No timing is invented for it.

| Loop | Before | After | Instrument |
| --- | --- | --- | --- |
| 1 | no replacement; merge failed in 0.728 s | replacement 0.314 s; green in 1.069 s | commands 1 and 2 in benchmark.json |
| 2 | no replacement; merge failed in 0.669 s | replacement 0.322 s; green in 1.127 s | commands 3 and 4 in benchmark.json |
| 3 | no replacement; merge failed in 0.720 s | replacement 0.399 s; green in 1.355 s | commands 5 and 6 in benchmark.json |

Best of three: replacement starts 0.314 s after the marker, default poll 30 s.
All three replacements start under 60 s. Each successful merge reads attempt 2,
and the healthy shard starts once. A separate marker delayed until after the first
poll was replaced in 29.737 s in the fully instrumented correctness test, default poll 30 s.
The test puts the missing-log deadline at 70 seconds, outside the promised minute.

Gate command covering the delayed-marker proof (the standalone instrument is in next-poll.json):

```
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 5m -v ./cloud -run '^TestFleetFastRelaunch$' > /workspace/fleet-cloud-proof.log 2>&1
```

Build-flags for the measurements: commit=358dbccd95a909a849d1d9db370ae7122a87fc5b
with the documented scratch edits; nproc=5; cpu.max=`400000 100000`;
Go=`go version go1.27.1 linux/amd64`;
clang=`clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`;
Node=`v24.19.0`; `ADAMIC_GATE_UNCACHED=1`, no new result cache.
The six exact commands, per-run build-flags and load averages before/after are in
[benchmark.json](benchmark.json). The delayed-marker correctness observation's
exact command, build-flags and load averages are in [next-poll.json](next-poll.json):
load-before=`0.00 0.15 0.30 1/2011 259980`, load-after=`0.00 0.14 0.29 1/2032 260292`.
It used the same checkout, toolchain and CPU quota; this is a correctness proof,
not a paired performance comparison.

Required setup: `ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh > /workspace/fleet-relaunch-setup.log 2>&1`,
then `source /workspace/adamic-tools/env.sh`. Cumulative timing lines: Go 0.194 s,
submodules 0.261 s, Node 0.649 s, clang 0.700 s, markdown dependencies 0.758 s,
stage3 dependencies 1.126 s, module dependencies 23.098 s, Go build 68.020 s,
workspace sums restored 68.530 s, done 68.732 s, nproc=5. Setup's complete
build-flags line and load averages are in [setup.log.gz](setup.log.gz): cached=yes,
load-before=`0.02 0.01 0.05`, load-after=`6.08 1.74 0.64`, same commit/toolchain/quota.

## Validation

- Fleet gate test: 12 Python integration tests pass through TestFleetFastRelaunch,
  including next-poll replacement, healthy shard preservation, replacement archive
  consumption, three-failure cap naming shard-0, reported launcher failure,
  stale markers/publications, foreign-commit markers, marker-only/stale-archive
  merge refusal, successful-archive commit identity, and unwritable run notes.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 5m ./cloud -run 'TestFleetFastRelaunch|TestLintWaveCheck' > /workspace/fleet-cloud-final.log 2>&1`: pass,
  including 43 existing lint-wave tests. The final fleet-only rerun above passes all
  12 final cases; [fleet-tests.log.gz](fleet-tests.log.gz) preserves its output.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 5m ./cmd/adamic-gate > /workspace/fleet-runner-tests.log 2>&1`: pass,
  including the existing 22-shard exactly-one-archive brief proof.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 5m ./internal/oracle -run '^TestNativeAgreesWithNode$/^dedication$/^dedication.a$' > /workspace/fleet-filtered-oracle.log 2>&1`: pass, the dedication source agrees with Node, JavaScript and native variants.
- `go vet ./cloud ./cmd/adamic-gate > /workspace/fleet-vet.log 2>&1`,
  `bash -n cloud/gate/fleet.sh`, and `git diff --check`: pass.

Eight isolated code mutants fail their designated tests. [mutants.json](mutants.json)
names each test; the corresponding compressed log records the failure:

| Mutant | What caught it |
| --- | --- |
| defer-marker | replacement test times out at 65 s with no replacement; deadline is 70 s |
| four-attempts | cap test expects shard-0 failure after three attempts, sees a fourth |
| ignore-merge-marker | stale-archive marker refusal loses the required shard-0 diagnosis |
| repeat-old-marker | delayed replacement starts extra attempts |
| ignore-marker-commit | a foreign-commit marker is accepted instead of refused by name |
| stale-publication | old attempt overwrites a newer archive |
| wrong-success-commit | an archive from another checkout is accepted |
| missing-notes-guard | a replacement starts despite unwritable run notes |
