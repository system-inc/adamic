# tsc full latent census: main versus private IR field fix

The fix removes every `latent state copy: unexported IR field methodTables` panic, but it does not make the census panic-free. Both runs completed with exit 0 on the same 81 files and 10,705 distinct units. Main took 623.536 seconds (10m 23.536s); fix took 976.730 seconds (16m 16.730s). Runs were sequential, with `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1` and a 5,400 second timeout each.

| Metric | Main | Fix |
| --- | ---: | ---: |
| Units | 10,705 | 10,705 |
| NotYet findings | 224 | 317 |
| Refused findings | 4,049 | 4,178 |
| Panic findings | 9,818 | 7,929 |
| Error findings | 0 | 109 |
| Boundary findings | 9,818 | 8,151 |
| SkippedDependency findings | 0 | 0 |

Main's 9,818 panic findings all have the exact reason `latent state copy: unexported IR field methodTables`. Example: `/workspace/scratch/tree/src/compiler/corePublic.ts:3:1`.

The fix's 7,929 panic findings affect 7,925 units and have 106 exact reasons. Every reason is `lower: source module has no import-declaration route: <path>`, optionally prefixed with `statement panic: `. There are 7,819 findings without that prefix and 110 with it. The largest exact reason names `src/compiler/checker.ts`, with 2,390 findings; example unit `/workspace/scratch/tree/src/compiler/checker.ts:1455:1`. [TABLES.md](TABLES.md) lists every exact remaining reason, its count and an example. No state-copy panic remains. This measurement does not change or diagnose the routing implementation.

Of the same 9,818 units that panicked on main, the fix produces:

| Outcome | Units |
| --- | ---: |
| Lowering attempted with no owned finding | 1,639 |
| NotYet or Refused, with no panic | 254 |
| At least one panic finding | 7,925 |

The other units retain their statuses: 823 deferred generic instantiations and 64 checker body splits. The five outcome counts sum to 10,705. A unit with both a refusal and a panic belongs to the panic outcome. Four additional panic findings occur within already counted units, explaining the difference between 7,929 findings and 7,925 units. The ledger's `status` alone misses statement-level recovered panics, so outcomes join findings by `(file, unit.where)` and inspect their kinds.

Among the 254 units with NotYet or Refused and no panic, the leading reasons are an unchecked cast (24 units), an open numeric enum used as a literal tag (13), and a value of type any (11). [TABLES.md](TABLES.md) gives the top 20 reasons with counts and examples, plus an additional top 20 including panicking units. A unit can contribute to multiple reason groups. [summary.json](summary.json) includes all reason groups and example sites. Thirteen findings per run have no matching unit key and remain in finding totals, without being assigned to a unit outcome.

Both ledgers explicitly say `measured on a checker-rejected entry-root program`, with 144 entry diagnostics. The no-finding count means this measurement's independent lowering attempt had no owned finding; it is not a claim that tsc compiles or produces usable IR. Both binaries assert that measurement produces no usable output.

The fix's requested package test passed:

```text
ok  github.com/system-inc/adamic/stage3/census/latent/statecopy  2.798s
```

The box reports `nproc=5`, a CPU cgroup quota of `400000/100000` (4 CPUs), and `MemTotal=18440100 kB` (17.59 GiB). Its memory cgroup limit is `17179869184` bytes (16 GiB); no swap was configured. Go was `go1.27.1 linux/amd64`, Node `v24.19.0`, and clang `20.1.8`. [box.json](box.json) and [setup.log](setup.log) preserve the observations. The box.json start snapshot was captured during the main run; setup.log records the pre-build box and load.

Pins:

- Main: `141ad1d244d6a2d9c5403858086d505dfdef5fd5`
- Fix: `caa19bcd7ff9ad977e6e72fde659acade7f0c4ab`
- Upstream TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`

The adapted tree was built once from main and used by both binaries. Each pinned worktree supplied its own overlay generator and entry overlay. [COMMANDS.md](COMMANDS.md), [build-entry.sh](build-entry.sh) and [run-census.sh](run-census.sh) record commands, submodule preparation, bounds and log paths. Per-run pins files contain submodule SHAs and binary SHA-256 digests. Setup passed in 397.792 seconds. Both entry overlay builds passed. The unavailable `/usr/bin/time` caused one launch failure before any census ran; both actual measurements used Bash's built-in timer.

The original ledgers are [main.jsonl.gz](main.jsonl.gz) and [fix.jsonl.gz](fix.jsonl.gz). Decompression was checked against the original files, unit key sets match exactly, and all outcome and panic totals reconcile. [analyze.py](analyze.py) reproduces the tables. No existing source under `stage3/census` or `internal` was changed; additions are limited to this measurement directory.
