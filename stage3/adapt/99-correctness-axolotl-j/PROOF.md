# 99j fourteen-site proof record

Branch base: origin/main `95f07db0bb7ede61e91b2dfe2e0582f51c1e28b0`. This single commit contains the fourteen-site adapter and its records. The old branch is unchanged. Exact measured adapter bytes are pinned by SHA-256 in proof-results.json; lane execution records HEAD at the base because the final commit was made after verification.

Fourteen sites ship: four narrowings and ten checks, implemented as twelve statement edits in six files. Three sites are left: the builder null-path cast and both commandLineParser string-array casts. The four identity text.split casts are also untouched.

| Proof | Result |
| --- | --- |
| Fresh full apply | All 68 adapters pass; 136 before/after refs; source equals exactly the twelve inventoried statement edits |
| Fresh replay | Full cached replay byte-identical, excluding Git metadata; direct second adapter run has zero edits and writes |
| Refusal contract | 20 changed/partial fixtures rejected before writes; all three left statements unchanged |
| Fresh whole-program checker | Current-main compiler: all 79 sorted roots in one invocation, zero diagnostics on main and zero with fourteen sites |
| Fresh Go lane fixtures | Pass: go test -count=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/ with GOMAXPROCS=1 |
| Fresh shard check | Pass: python3 -B stage3/lane/generate_shards.py --check |
| Fourteen-site oracle | PASS: 106,366 passing, one sanctioned failure, zero pending |
| Main comparison | Exact complete pass/fail identity multisets match the retained main oracle; no new failures |
| Public declarations | Built typescript.d.ts and local API baseline declaration bytes match main exactly |
| Null path evidence | Two-line Node run against the built main tree returns "..", the value written for outDir in build info |

The first fourteen-site oracle attempt encountered eighteen 40-second timeouts under resource pressure (106,293 passing, nineteen failing including the sanctioned failure). Its full report is preserved in proof-results.json. Clean scratch file-cache pages were released with POSIX_FADV_DONTNEED. Verified replay copies and failed-run source copies were then deleted to reduce cached filesystem metadata; all raw logs, reports, measured main/fourteen/fifteen-site sources and the built baseline were retained. The complete unmodified lane was rerun. No timeout, worker count, test selection or expected count was changed. The retry passed and its complete pass/fail identity multiset equals main.

The unmodified oracle uses all runners, eight workers, no filter, and the existing expected counts. The one common failure is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`. The javascriptProjectEmit test that failed under the fifteen-site prototype now passes. Test identities are captured from Mocha's upstream parallel host replay by a read-only observer. No test, baseline, timeout or expected count changes were made.

The retained main oracle was measured on `5d521b9076fc432852013274416c1fdbab72581b`. Between that commit and the new base, no source adaptations, oracle scripts, lane checker, lane expected counts, source pin or parser lock changed. Fresh full apply also checks the unchanged main composition as its source reference. The new Go fixture tests run from current main.

## Derived census, not a fresh run

As explicitly requested, the fourteen-site delta is derived from the existing sequential main and fifteen-site first-error censuses. Their compiler pin is `5d521b9076fc432852013274416c1fdbab72581b`; no new-main compiler census is claimed. Both original runs exit zero over 79 roots with LATENT_FULL=0 and LATENT_ASSERT_NO_OUTPUT=1. Restore the builder cast and main getBuildInfo observations; keep the other fourteen removals. Fresh source checks prove that toReusableCompilerOptionValue is identical to main, and that the only difference from the measured fifteen-site composition is reverting that one assertion. All other five changed source files are byte-identical to that composition. The original main and fifteen-site getBuildInfo lowering stops already match.

Derived totals: Refused 4294 to 4280 (-14); NotYet 560 to 560; SkippedDependency, error and panic remain zero. Exactly fourteen sites are removed, zero sites are new, and three inventoried casts remain. [CENSUS.md](CENSUS.md) gives all seventeen per-unit results. [census-delta.json](census-delta.json) records raw-input hashes, provenance, source checks and full observations. This derivation is not represented as a newly executed fourteen-site census or successful native compilation.

Neither original census was killed for memory. Main observed kernel VmHWM was 8,132,056 KiB; after an environment reconnect its timeout exit zero was recovered from Linux proc stat, so this is an observed peak rather than wait4 peak. The fifteen-site run's maximum recorded RSS was 8,339,156 KiB (sampled VmHWM), with wait4 reporting 8,338,716 KiB. OOM-kill counters remained zero.

The README records both array-accessor evidence and the null-domain note for stage3/upstream/LEDGER.md. The ledger is unchanged. Exact fresh commands, exits, elapsed times, lane reports, declaration hashes and original census memory records are in [proof-results.json](proof-results.json). All raw logs remain under /workspace/scratch, with the 99j or 99j2 prefix.
