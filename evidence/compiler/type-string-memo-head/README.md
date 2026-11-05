# Entry census peak memory at 05c109931001dfa58a9ed6a49bca98d4c7bc5822 (#wy2dyf0)

Adamic sha: `05c109931001dfa58a9ed6a49bca98d4c7bc5822` (head). Submodules: cohere 06af2e3, cohere/TypeScript d92d9bfe.

Step measured: the entry census on the adapted tsc.ts, first-error mode, as `stage3/meter/twice-daily.sh` builds and runs it
(lines 56 to 65 build; line 109 runs `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 entry-census <adapted>/src/tsc/tsc.ts tsc/latent.jsonl`).
Upstream TypeScript v6.0.3 cloned from the tag, adapted with `stage3/apply.sh` to a fresh path. Nothing else ran during the measured step.

## Box
- CPU: Intel Xeon Processor @ 2.10GHz, 4 cores (nproc 4)
- Memory: 16086 MB total, no swap
- Linux 6.18.44, Go 1.27.1, Node v22.23.4, GNU time 1.9, Python 3.13

## Numbers
- Max RSS (`/usr/bin/time -v`): **4631240 kB** (about 4.42 GiB)
- Max RSS (1 Hz `ps` sampler, last sample, taken just before exit): 4631632 kB
- Wall: **2:23:33** (10:45:53Z to 13:09:27Z)
- Exit status: **0**
- `sha256sum tsc/latent.jsonl`: `481adbe6305ebc3e28e0182e660b05b1772341dc5c852a69e44484a214951fed` (17329350 bytes)

RSS climbed across the run and was highest at the end (4.63 GB in the last samples); the sampler has 8477 rows.

## Files
- `latent.jsonl.gz`: tsc/latent.jsonl, gzipped
- `entry-census.time`: `/usr/bin/time -v` output
- `rss.csv`: epoch, rss_kb, vsz_kb at 1 Hz
- `logs/`: build, overlay, entry-build, upstream clone, npm ci, apply and run logs (`latent.log` is the run)

Not covered: the other meter steps (census, compiler-directory census, latent census) were not run.
