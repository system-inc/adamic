# Entry census peak memory at 006b368745

Measured step: the entry census on tsc.ts, first-error mode, as stage3/meter/twice-daily.sh builds and runs it
(`LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 entry-census <adapted>/src/tsc/tsc.ts tsc/latent.jsonl`).
`run.sh` is the exact script used (the build, overlay, npm ci and apply steps copied from twice-daily.sh lines 56 to 65 and 109; no census/latent-census of src/compiler was run, since only this step was asked for).

- Sha: 006b36874 (`006b368745`, main), cohere submodules initialized (cohere 06af2e3, TypeScript d92d9bf)
- Box: Intel Xeon @ 2.10GHz, 4 cores, 16086 MB RAM (15 GiB), no swap, Linux 6.18.44, Go 1.27.1, Node v22.23.4, GNU time 1.9. Nothing else ran during the measured step.
- Started 2026-11-05T10:45:08Z, ended 13:25:23Z.

## Numbers

| | |
|---|---|
| Max RSS (GNU time) | **7531240 kB (about 7.18 GiB)** |
| Max RSS (1 Hz ps sampler, entry-census pid) | 7531120 kB, at the last sample before exit |
| Wall | **2:40:14** (user 10766.65 s, sys 106.98 s, 113% CPU) |
| Exit | **0** (time reports Exit status 0) |
| sha256 tsc/latent.jsonl (ungzipped, 23407807 bytes) | `61e36ec281de95f6ac7298acf7416379684c148c86e4743505b214d6bb6297ec` |

RSS grew steadily across the run (1.5 GB early, 5.75 GB at 2h21m, 6.06 GB shortly after, 7.53 GB at exit), so the peak is at the end.

## Files
latent.jsonl.gz, latent.log.gz, entry-census.time, rss.csv (epoch,rss_kb,vsz_kb,pid; the `time` wrapper pid is also listed, the census is the larger one), phase.txt, exit.txt, and the build/apply/npm logs.
