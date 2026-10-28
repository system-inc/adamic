# checker.ts timing: 22026c73 → 5e382ce4

Parent: `22026c7392f55556e721b6f425dd369ceafdf14f`. Change: `5e382ce4baca2f75cc9a452f9b438c8318f375a5` (its direct child).

Five samples per revision, alternating parent then change, sequentially on one box. All builds and timed invocations exited zero. checker.ts was the sole root; no whole-directory latent run was made.

This follows the commands.sh, K1T2WP9.md and compare.py method on typescript/stage3-meter-walk-4ac3f04d. GNU time -v measures end-to-end latent driver time: imports/loading/checking, declaration collection, no-output assertions, refusal scans, independent unit lowering, JSON and fsync. It does not isolate production Lower time. LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1; first-error findings are floors, and checker-rejected latent lowering does not establish native output.

| Revision | Sample | Wall (s) | User + system (s) | Max RSS (KiB) |
| --- | --- | --- | --- | --- |
| parent | 1 | 15.30 | 21.47 | 692316 |
| change | 1 | 15.13 | 20.44 | 731576 |
| parent | 2 | 14.72 | 20.13 | 724828 |
| change | 2 | 15.75 | 21.46 | 691604 |
| parent | 3 | 15.78 | 21.47 | 728432 |
| change | 3 | 14.83 | 20.39 | 769012 |
| parent | 4 | 16.88 | 22.75 | 749408 |
| change | 4 | 16.25 | 22.16 | 730680 |
| parent | 5 | 15.99 | 21.90 | 740904 |
| change | 5 | 15.25 | 20.67 | 729148 |

| Revision | Median wall (s) | Median CPU (s) | Median RSS (KiB) | Wall min (s) | Wall max (s) |
| --- | --- | --- | --- | --- | --- |
| parent | 15.78 | 21.47 | 728432 | 14.72 | 16.88 |
| change | 15.25 | 20.67 | 730680 | 14.83 | 16.25 |

Median wall change: -3.36%. Wall ranges overlap; CPU ranges overlap (20.13–22.75 vs 20.39–22.16 s).

The wall difference is not clearly past noise in these five samples: the observed ranges overlap.

Normalized full checker.ts latent records are identical. Findings are identical. Every revision reproduces its normalized stream across all five samples. Normalization replaces only the two adapted-tree path prefixes; all JSON fields and array order are retained.

No lowering reason has a changed distinct-site set.

Both revisions were built in separate detached worktrees with the exact build_census sequence from twice-daily.sh (census, latent overlay, entry overlay, entry-census, latent-census), then stage3/apply.sh. All 746 adapted src files are byte-identical.

Each binary SHA-256 is in parent/binaries.sha256 or change/binaries.sha256; timed binary hashes were checked against those manifests. Submodule pins and clean tracked-source checks are retained. /usr/bin/time was absent, so GNU time was extracted from the official Debian time_1.9-0.2_amd64.deb package; tool and package hashes are retained.

```text
2026-10-28T01:00:13Z
5
400000 100000
17179869184
               total        used        free      shared  buff/cache   available
Mem:     18882674688  1678286848 13098938368     8937472  4505468928 17204387840
Swap:              0           0           0
go version go1.27.1 linux/amd64
v24.19.0
Linux 51f1e41e10e4 6.18.55 #1 SMP Wed Oct 14 19:45:18 UTC 2026 x86_64 GNU/Linux
```

Reproduce with commands.sh after successful setup (see its preamble). Regenerate this report with `python3 stage3/meter/runs/time-5e382ce4/compare.py`. All ten raw streams are retained under timing/; parent-1.jsonl.gz and change-1.jsonl.gz are the compared pair. Timing commands, stdout/stderr, GNU time logs, setup, build, overlay, package and apply logs are retained. No compiler code was changed.
