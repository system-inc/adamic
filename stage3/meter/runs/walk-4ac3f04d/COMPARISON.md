# False-direction walk: 76ea8b17 → 4ac3f04d

First-error counts are floors: each unit contributes only its first stop. The latent stream measures lowering on a checker-rejected program; it does not establish native output. The main and area labels repeat the same pin. Tables use main once, without adding area duplicates.

## Headline

| SHA | Whole program | Own file | Lowering | tsc entry diagnostics | NotYet findings | Refused findings | Panic findings | Error findings |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 76ea8b17 | 2/79 | 58/79 | 2/79 | 144 | 1122 | 3083 | 1 | 0 |
| 4ac3f04d | 2/79 | 58/79 | 2/79 | 144 | 1122 | 3083 | 1 | 0 |

Whole program is the number of per-file roots whose entire resolved program passes the checker. Own file excludes files with a located checker diagnostic. Lowering is ordinary successful IR lowering. These gate statistics come from census.jsonl.gz; latent.jsonl.gz alone cannot reconstruct them. NotYet/Refused/panic/error columns use the meter definition: unique (kind, where, reason, text) findings.

## Changed lowering reasons

Exact reason strings are grouped after normalizing scratch prefixes to src/. Counts use distinct file:line:column sites, reusing the earlier count-predicates.py set logic. Only NotYet and Refused are lowering reasons in this table; other errors and panics are separate. A decrease may reflect an earlier stop hiding a prior reason, rather than successful lowering.

| Reason | Before | After | Difference |
| --- | --- | --- | --- |
| None; all distinct-site counts match | — | — | 0 |

## Files with a changed first stop

The first finding in each file record is its recorded first stop: the file refusal scan runs before independent top-level-unit attempts. This table does not claim to list every unit whose stop changed.

No files changed first stop.

## Panics

### src/compiler/moduleSpecifiers.ts:335:1

Recorded in: 4ac3f04d/area/latent.jsonl.gz, 4ac3f04d/main/latent.jsonl.gz, 76ea8b17/area/latent.jsonl.gz, 76ea8b17/main/latent.jsonl.gz

```text
runtime error: invalid memory address or nil pointer dereference
```

## Latent driver timing

The system /usr/bin/time was absent and system installation was denied by Unix permissions. The first timing attempt exited 127 before launching the latent binary. GNU time was instead extracted from the official Debian time_1.9-0.2_amd64.deb package and invoked at /workspace/adamic-tools/gnu-time/extracted/usr/bin/time with -v. This is a tool-path deviation from the requested command; the measured compiler binaries and flags are unchanged. Failed prerequisite attempts are preserved under timing/prerequisite-failure/. Package and tool provenance follows.

```text
Package: time
Version: 1.9-0.2
Architecture: amd64
time (GNU Time) UNKNOWN
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Written by David Keppel, David MacKenzie, and Assaf Gordon.
10257c6fe888e6578550b1f20705d282dd630b76f8df0c006aec3031fefe5e23  /workspace/adamic-tools/gnu-time/time_1.9-0.2_amd64.deb
efc0d1112e36ec76c14bf7508e287935cf1631d9d0d91fb532802d07aef51cea  /workspace/adamic-tools/gnu-time/extracted/usr/bin/time
```

Each invocation uses the exact retained main-label latent-census binary built by twice-daily.sh. LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 is set. The checker.ts scope supplies that file as the sole root; a whole-program scope would supply the adapted src/compiler directory, but is omitted below. Imports are loaded and checked. This is end-to-end latent driver time, including walking input paths, loading/checking, the no-output assertions, declaration collection, refusal scans, independent unit lowering, JSON encoding and fsync; it is not isolated production Lower time. All timing invocations run sequentially. Each reported value is the median of three samples; CPU is user+system per sample and RSS is the per-invocation maximum, then median across samples.

| SHA | Root scope | Median wall (s) | Median user+sys (s) | Median max RSS (KiB) | Peak RSS across samples (KiB) |
| --- | --- | --- | --- | --- | --- |
| 76ea8b17 | checker | 16.13 | 21.44 | 735456 | 749700 |
| 4ac3f04d | checker | 16.60 | 22.51 | 718844 | 729064 |

Whole-program timing repeats were omitted under the requested 60-minute fallback. The before whole-program latent process was observed running for more than 57 minutes; three repeats of that scope alone, before adding after and checker samples, would exceed the budget. Only checker.ts is timed, three sequential invocations per pin.

## Box and run elapsed time

```text
2026-10-27T01:45:47Z
5
400000 100000
               total        used        free      shared  buff/cache   available
Mem:            17Gi       3.1Gi        12Gi       212Mi       2.3Gi        14Gi
Swap:             0B          0B          0B
               total        used        free      shared  buff/cache   available
Mem:     18882674688  3339489280 13653753856   223047680  2503553024 15543185408
Swap:              0           0           0
v24.19.0
go version go1.27.1 linux/amd64
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
Target: x86_64-unknown-linux-gnu
Thread model: posix
InstalledDir: /workspace/adamic-tools/llvm/bin
Linux 9fe78be11265 6.18.55 #1 SMP Wed Oct 14 19:45:18 UTC 2026 x86_64 GNU/Linux
```

| SHA | Run directory | Meter elapsed (s) |
| --- | --- | --- |
| 76ea8b17 | 20261027T014837Z.mugwl6 | 3935 |
| 4ac3f04d | 20261027T025412Z.HCmJYO | 3806 |

Meter elapsed includes fetching, compiler/overlay builds, adaptation, census and reporting. Preparation/cache states differed (see the apply logs), so these totals are not isolated lowering times.

## Attribution and reproducibility

4ac3f04d is the direct child of 76ea8b17; there is no other intervening commit. There are no stage3 or submodule changes between these pins. Checker gate differences cannot be explained by this lowering-only walk and must be treated separately. The comparison script checks repeated labels and input metadata below rather than pooling them.

The latent step has no per-file timing fields. Its header records checker diagnostics and spans; each LatentFile records declarations, unit statuses, findings and optional boundaries. Ordinary census seconds measure a different gated path and are not latent lowering times. Its stderr reports file and count summaries only. See the pinned main.go.txt and lower.go.txt and the unmodified twice-daily.sh build commands.

Run `python3 stage3/meter/runs/walk-4ac3f04d/compare.py` from any directory to regenerate this file. The script reads committed compressed latent and census streams plus raw time -v logs. See commands.sh for the exact setup, run, timing and publication commands, timing/commands.log for every timed invocation, and run-status.txt for external elapsed times and exit codes. Initial setup overlapped a branch switch and failed; the preserved initial log is not a meter failure. Setup was repeated successfully on the stable target checkout.

- 76ea8b17: timed binary SHA-256 55d89add610dfb13ac2dd0a3a1521879ce6e0e28f1e517d2e02923cee2e247ec matches the meter; each of the three timing streams contains only the checker.ts file record.
- 76ea8b17: pin metadata verified; main/area raw streams agree after path and census-seconds normalization.
- 4ac3f04d: timed binary SHA-256 9ed61cdc7c4d8f0a60d4805f10f0d32ca8a2f18e643b14144ee5075496e05169 matches the meter; each of the three timing streams contains only the checker.ts file record.
- 4ac3f04d: pin metadata verified; main/area raw streams agree after path and census-seconds normalization.
- Before/after latent.jsonl.gz: identical after source-path and census-seconds normalization.
- Before/after census.jsonl.gz: identical after source-path and census-seconds normalization.
- Before/after tsc/census.jsonl.gz: identical after source-path and census-seconds normalization.
- Before/after tsc/latent.jsonl.gz: identical after source-path and census-seconds normalization.
- The live origin/main advanced during the second fetch; both explicit run pins remained fixed, as verified by compiler-mode.json. No intervening commit was included in this comparison.
