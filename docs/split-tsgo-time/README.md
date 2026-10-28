# TestSplitTSGoAgreesUnit00 and Unit01: where the seconds go (#dgcy49s)

Measured only. No compiler or test code changed; nothing submitted.

## The box and the tree

- Tree: origin/main at `95a7b63f12219f756b9bcd10f9aad953781add96`, cohere `06af2e399c6c`, typescript-go `d92d9bfee114`.
- Box: a cloud container, 4 processors (`nproc` 4, Intel Xeon Processor @ 2.10GHz), 15.7 GB RAM, no swap, Linux 6.18.
- Tools from `cloud/setup.sh --wasi-sdk`: clang 20.1.8 (the release tarball in `/opt/adamic-tools/llvm`, picked first by
  setup's sanitizer probe), Node v24.22.1, go 1.27.1, `WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot`.
  Setup's log is `logs/setup.log`.
- Each run's `logs/<run>.box` holds the date, sha, `nproc`, CPU, RAM, clang, Node and `/proc/loadavg` before and after.

This box is not a house box. Its quiet times are about half of main's reported 23.4 s, so the absolute seconds do not
carry over; the shares and the critical path are what to read.

## How it was measured

`tools/run.sh <run> <regex>` runs `go test ./internal/native -run <regex> -count=1 -v -timeout 30m` under
`/usr/bin/time -v`, with `ADAMIC_BUILD_LOG` set to a fresh `logs/<run>.build`. Beside it `tools/poll.py` stamps, every
20 ms, each new line of the `-v` log and the build log, the 1-minute load each second, and the size of every unit `.c`
file that appears under `$TMPDIR/adamic-units-*/.unit-*` (`logs/<run>.timeline`). `tools/analyze.py` reads the three
files into `logs/quiet-summary.txt` and `logs/loaded-summary.txt`.

The test logs only two phases (product fetch, whole checker run). The rest comes from the timeline, so it is read off
20 ms polling of when lines landed, not from timers in the test:

- split build: from the `=== RUN .../cache_N` line to the `cLink` line landing in the build log.
- split run: from the `cLink` line to the subtest's "identical bytes" log line. This includes the split binary's run and
  the compare; it can't separate them.

"Cleared" runs: the repository's removal guard (`.claude/hooks/guard-removals.mjs`) refuses a move outside the repository and
/tmp, so instead of moving `~/.cache/adamic-build/native-units` aside, the cache's other two directories (`local`, which
holds the products, and `native-runtime`) were copied to `/tmp/adamic-gate/cache-cleared*` and the run pointed at the
copy with `ADAMIC_BUILD_CACHE_DIR`. The products and the runtime archive still hit; only the unit objects were missing.
The real cache was left as it was.

Loaded runs: `stress-ng --cpu 8` (2x nproc) started first, and each run waited for the 1-minute load to reach 7.5.

## Products built ahead (run 00)

`go test ./internal/native -run '^TestProduct_(TSGoArchive|SplitTSGoSource|SplitTSGoBaseline)$' -count=1 -v`, from an
empty build cache: 2:43.58 wall, max RSS 2.3 GB (`logs/00-products.time`). TestProduct_SplitTSGoSource 4.22 s,
TestProduct_TSGoArchive 121.70 s, TestProduct_SplitTSGoBaseline 148.50 s (it waits on the other two).

## The runs

Seconds. "cCompile sum" is the sum of the `cCompile` lines' seconds (clang's own time, waits for a token left out),
across 23 units: 22 keyed units plus `main.c`, which always compiles uncached ("main.c holds a directive past its first
line"). Unit01's 23 are all `off`. Every run also had 3 `build ... hit` product lines and 1 `cArchive runtime.a hit 0.000`.

| run | test | quiet or loaded | load 1m (before; during) | total | product fetch | whole run | split build | cCompile sum over 23 units | cLink | split run |
|---|---|---|---|---|---|---|---|---|---|---|
| 01 | Unit00 | quiet, units cache empty | 0.97; 1.0-1.3 | 11.52 | 0.904 | 0.076 | 10.45 | 17.34 (22 miss, 1 off) | 0.345 | 0.08 |
| 02 | Unit01 | quiet | 0.99; 1.0 | 10.83 | 0.908 | 0.077 | 9.75 | 17.54 (23 off) | 0.407 | 0.10 |
| 03 | Unit00 | quiet, warm | 0.77; 0.8 | 4.46 | 1.010 | 0.108 | 3.22 | 1.36 (22 hit, 1 off) | 0.407 | 0.12 |
| 04 | Unit01 | quiet | 0.81; 0.8 | 11.33 | 0.981 | 0.084 | 10.17 | 18.66 (23 off) | 0.404 | 0.08 |
| 05 | Unit00 | quiet, warm | 0.84; 0.8 | 3.59 | 0.906 | 0.088 | 2.48 | 0.28 (22 hit, 1 off) | 0.439 | 0.10 |
| 06 | Unit01 | quiet | 0.86; 0.9-1.2 | 11.00 | 0.932 | 0.090 | 9.90 | 17.42 (23 off) | 0.391 | 0.08 |
| 07 | Unit00 | quiet, cleared | 0.78; 0.8-1.1 | 11.11 | 0.878 | 0.086 | 10.05 | 16.57 (22 miss, 1 off) | 0.377 | 0.08 |
| 11 | Unit00 | loaded, warm | 7.50; 7.6-7.7 | 7.25 | 2.314 | 0.164 | 4.49 | 0.18 (22 hit, 1 off) | 0.789 | 0.26 |
| 12 | Unit01 | loaded | 7.73; 8.1-8.8 | 27.01 | 2.107 | 0.200 | 24.46 | 46.76 (23 off) | 0.910 | 0.23 |
| 13 | Unit00 | loaded, warm | 8.80; 8.7-8.9 | 7.13 | 2.030 | 0.260 | 4.53 | 0.12 (22 hit, 1 off) | 0.888 | 0.29 |
| 14 | Unit01 | loaded | 8.92; 9.2-9.5 | 27.00 | 2.358 | 0.278 | 24.11 | 44.03 (23 off) | 0.814 | 0.25 |
| 15 | Unit00 | loaded, cleared | 7.94; 8.2-8.8 | 28.94 | 2.401 | 0.283 | 26.00 | 48.33 (22 miss, 1 off) | 0.911 | 0.27 |

Run 01 was the first Unit00 after the products were built, so its units cache was empty, the same as a cleared run. Runs
03, 05, 11 and 13 hit all 22 keyed units. No run repassed, and no run printed a unit cache mismatch.

The ten slowest `cCompile` lines, per run, are in the summaries. Every uncached or missing run has the same top three:
`functions_0016.c` 6.66-7.31 quiet and 16.87-16.99 loaded, `functions_0009.c` 1.90-2.31 quiet and 5.25-6.49 loaded,
then `functions_0015.c` or `functions_0008.c` near 1.0-1.5 quiet and 3.0-3.7 loaded.

### The critical path (from the timelines, `logs/critical-path.txt`)

Seconds after the cache subtest starts:

| run | first unit file written | `functions_0016.c` written | `functions_0016.c` done | cLink line |
|---|---|---|---|---|
| 01 Unit00 quiet, empty | 1.73 | 3.12 | 10.10 | 10.45 |
| 02 Unit01 quiet | 1.16 | 2.59 | 9.36 | 9.75 |
| 04 Unit01 quiet | 1.16 | 2.45 | 9.76 | 10.17 |
| 06 Unit01 quiet | 1.08 | 2.48 | 9.49 | 9.90 |
| 07 Unit00 quiet, cleared | 1.77 | 3.01 | 9.69 | 10.05 |
| 12 Unit01 loaded | 2.76 | 6.58 | 23.56 | 24.46 |
| 14 Unit01 loaded | 2.70 | 6.40 | 23.29 | 24.11 |
| 15 Unit00 loaded, cleared | 3.67 | 8.07 | 25.07 | 26.00 |

In every compiling run `functions_0016.c` was the last unit to finish, and the link followed it.

## The split source's shape

From run 02's timeline (Unit01 compiles every unit, so every unit's `.c` was seen): BuildSplitTSGo made 23 units plus the
shared `units.h`, 4,282,122 bytes of C in all.

| file | bytes | | file | bytes |
|---|---|---|---|---|
| `units.h` (in every unit) | 476,762 | | `functions_0010.c` | 168,608 |
| `state.c` | 490,252 | | `functions_0011.c` | 325,287 |
| `functions_0000.c` | 6,289 | | `functions_0012.c` | 210,219 |
| `functions_0001.c` | 6,037 | | `functions_0013.c` | 209,761 |
| `functions_0002.c` | 6,004 | | `functions_0014.c` | 190,651 |
| `functions_0003.c` | 6,095 | | `functions_0015.c` | 182,487 |
| `functions_0004.c` | 5,941 | | `functions_0016.c` | 1,026,899 |
| `functions_0005.c` | 6,308 | | `functions_0017.c` | 18,910 |
| `functions_0006.c` | 25,529 | | `functions_0018.c` | 74,015 |
| `functions_0007.c` | 121,808 | | `functions_0019.c` | 150,185 |
| `functions_0008.c` | 229,042 | | `functions_0020.c` | 714 |
| `functions_0009.c` | 340,391 | | `main.c` | 3,928 |

Sizes are of each unit's own `.c` as compileUnitObject wrote it, `units.h` and the runtime headers beside it not counted.

## What dominates

Measured:

- **The split build is the test.** It is 89-91% of the total in every compiling run (9.75-10.45 of 10.83-11.52 quiet;
  24.11-26.00 of 27.00-28.94 loaded). Product fetch is 0.88-1.01 s quiet and 2.03-2.40 s loaded, the whole checker run
  under 0.3 s, the split run plus compare under 0.3 s.
- **Inside the build, clang is the time, and one unit is most of it.** `functions_0016.c` (1.03 MB, 27% of the 23 units' own C)
  took 6.66-7.31 s of clang quiet and 16.87-16.99 s loaded, and was on the critical path every time: it is written 2.5-3.1
  s into the subtest quiet (6.4-8.1 s loaded) and the build ends about 0.4 s (0.9 s loaded) after it finishes. Its compile
  alone is 66-72% of the split build quiet and about 65-70% loaded.
- The cLink is 0.35-0.44 s quiet and 0.79-0.91 s loaded. The units cache, warm, saves Unit00 about 7 s quiet: 2.48-3.22 s
  of split build against about 10 s, the remaining time being 1.1-1.2 s before the runtime archive line, about 0.65 s
  more to the first `cCompile` line, and the link.
- Before any unit compiles there is 1.1-1.8 s quiet (2.7-3.7 s loaded) of work in the subtest: 1.0-1.2 s (2.3-2.7 s
  loaded) to the runtime archive's line in every run, and in Unit00 about 0.6 s (1.0 s loaded) more, both runs whether hit
  or miss.
- Load hurts the compile about 2.4x: the uncached build went from about 10 s at load 1 to about 24.3 s at load 8-9.5 on 4
  processors. A warm Unit00 went from 3.6-4.5 s to 7.1-7.3 s.
- Unit00 against an empty units cache (runs 01, 07, 15) costs the same as Unit01 (02, 04, 06, 12, 14): a miss is a
  compile plus a write into the cache.

Inferred, not measured:

- Splitting the test cannot bring a compiling run under `functions_0016.c`'s own compile plus the fixed time around it:
  about 7 + 3 s quiet and 17 + 6 s at 2x load here. Each of Unit00 and Unit01 already holds one cache mode, so a finer test
  split doesn't shorten either unless it stops compiling the whole program. What would shorten them is in the build, not
  the test: a smaller `functions_0016.c` (the splitter's units are very uneven, 714 bytes to 1.03 MB), or compiling the
  biggest units first. `functions_0016.c` was the 18th unit file to appear for 5 workers; started first, it would begin about 1.3 s
  sooner quiet and 3.7 s sooner loaded (the gap between the first unit file and its own in the critical path table).
- The 23.4 s on main at load 6 is about twice this box's quiet 11 s. If that run compiled (a fresh units cache on that
  box, or Unit01), the same shares would put `functions_0016.c` near 14-16 s of it. I have no build log from that box, so
  this is a guess from shares.
- Unit00 is cheap only when its units are in this box's cache. A Loom box starting with no units cache pays Unit01's cost
  for Unit00 too.
- What clang spends on `functions_0016.c` (front end, optimizer, sanitizer passes, debug info) was not measured: no
  `-ftime-trace` was taken, since the unit's `.c` lives in a directory BuildSplitTSGo removes, and the poller kept only
  sizes. That is the next measurement if the unit itself is to be split.

## Files

- `logs/<run>.box`, `.v` (the `-v` output), `.build` (the build log), `.timeline`, `.time` (`/usr/bin/time -v`).
- `logs/quiet-summary.txt`, `logs/loaded-summary.txt`: the per-run summaries, with the ten slowest `cCompile` lines.
- `logs/critical-path.txt`, `logs/stress-ng*.log`, `logs/setup.log`.
- `tools/`: the scripts. `run.sh` and `quiet.sh` hard-code this session's paths.
