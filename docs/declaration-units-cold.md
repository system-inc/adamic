# Declaration units, cold split builds: base against head

Question: does giving every declaration its own translation unit (head) make a cold split native build slower than
sixteen-function units (base), and by how much?

Answer, measured: yes. With the unit cache cold and the runtime archive warm, at four jobs on four cores, the lint port's
build takes 51.4 s on head against 28.8 s on base (1.79x wall, 2.35x CPU). The smaller programs pay 1.19x to 1.54x
wall. Adding one empty function misses every unit on both trees, so on this edit head's stable identities buy no
incremental reuse; the cause is `units.h` (below).

- base: `926e19f4967c0f13a2b9a58560dc1eb61601d9c3` (origin/main)
- head: `c33dafb71d07a1c5f6cf192689229feb2a156dd6` (`runtime/wb-batch-stable-identities-v3`, one commit on base)

## The box

- Linux 6.18.44, a cloud container, nothing else running (`ps` showed only the session's own agent at about 1% CPU).
- `nproc`: 4. CPU: `Intel(R) Xeon(R) Processor @ 2.30GHz`. 15 GiB memory, no swap.
- clang 20.1.8 (`87f0227cb60147a26a1eeb4fb06e3b505e9c7261`) from the official `LLVM-20.1.8-Linux-X64.tar.xz`,
  extracted without `bin/flang*`, `bin/mlir-*`, `bin/lldb*`, `bin/llvm-bolt*`, `bin/clangd` and `lib/lib*.a`, first on
  PATH.
- Node v24.22.1 next on PATH. Go 1.27.1 (go.mod's toolchain). GNU time 1.9 (Ubuntu `time` 1.9-0.2build1).
- `/proc/loadavg` before and after every build is in `raw/load-*.txt`. The 1-minute average sat between 1.4 and 3.5
  throughout: the builds ran back to back, so each reading is mostly the build before it, not outside work.

## Results

Medians of three interleaved runs (base, head, base, head, base, head). "Units" counts the program's `cCompile` lines,
`main.c` included (it compiles uncached, `off`, on both trees: "main.c holds a directive past its first line"); every
other unit line was a `miss` on every timed run. No runtime `cCompile` line appeared in a timed run: the archive was a
`cArchive hit` each time. "Sum unit s" is the sum of the `cCompile` seconds; "largest" is the largest one; "link" is the
`cLink` seconds. Peak RSS is GNU time's maximum resident set, which is the largest single process in the build, not the
sum.

### stage1/cohere/lint/main.ts (the lint port)

| | base | head | head / base |
|---|---|---|---|
| wall s | 28.79 | 51.42 | 1.79x (+22.63 s) |
| CPU s (user+sys) | 63.00 | 148.15 | 2.35x |
| peak RSS MB | 535 | 588 | 1.10x |
| units | 95 | 1487 | 15.7x |
| sum of unit s | 10.61 | 62.84 | 5.92x |
| largest unit s | 1.23 (`functions_0062.c`) | 0.70 (`Statements_statement`) | |
| link s | 8.47 | 14.19 | 1.68x (+5.72 s) |
| units.h bytes | 1,138,304 | 1,937,272 | 1.70x |

Per run, wall: base 28.55, 28.79, 28.97; head 48.74, 51.42, 51.50.

### stage1/cohere/tsprinter/main.ts

| | base | head | head / base |
|---|---|---|---|
| wall s | 11.50 | 16.00 | 1.39x (+4.50 s) |
| CPU s | 26.45 | 44.35 | 1.68x |
| peak RSS MB | 408 | 462 | 1.13x |
| units | 31 | 461 | 14.9x |
| sum of unit s | 4.63 | 14.96 | 3.23x |
| largest unit s | 1.05 | 0.56 | |
| link s | 3.70 | 5.08 | 1.37x |
| units.h bytes | 294,596 | 514,196 | 1.75x |

### stage1/cohere/json/main.ts

| | base | head | head / base |
|---|---|---|---|
| wall s | 2.88 | 3.42 | 1.19x (+0.54 s) |
| CPU s | 4.86 | 7.26 | 1.49x |
| peak RSS MB | 330 | 339 | 1.03x |
| units | 7 | 81 | 11.6x |
| sum of unit s | 0.53 | 2.01 | 3.79x |
| largest unit s | 0.19 | 0.14 | |
| link s | 0.62 | 0.66 | 1.06x |
| units.h bytes | 40,764 | 70,881 | 1.74x |

### stage1/cohere/yaml/main.ts

| | base | head | head / base |
|---|---|---|---|
| wall s | 9.19 | 14.12 | 1.54x (+4.93 s) |
| CPU s | 18.51 | 37.56 | 2.03x |
| peak RSS MB | 380 | 436 | 1.15x |
| units | 33 | 483 | 14.6x |
| sum of unit s | 2.99 | 14.55 | 4.87x |
| largest unit s | 0.39 | 0.13 | |
| link s | 2.30 | 3.46 | 1.50x |
| units.h bytes | 358,447 | 607,433 | 1.69x |

All four entries are `main.ts` under their directories, as named. `units.h` was copied out of the build's
`adamic-units-*` directory while the warm-up build ran (the directory is removed after the link); every copy ends in
`#endif`.

### Incremental, lint port

After a cold timed build, one line was added to `stage1/cohere/lint/context.ts` (identical in both trees), after
`space` at line 66 and before `class RuleContext`, which `lint.ts` imports after the parser and scanner:

```
export function benchmarkEmptyFunction(): void {}
```

`adamic c` showed the function emitted on both trees (two mentions in the C). Then the same build ran with the unit
cache warm, and the edit was reverted (`git checkout`, `git status` clean after).

| | base | head |
|---|---|---|
| program units after the edit | 95 | 1488 (the new function's own unit included) |
| miss | 94 | 1487 |
| hit | 0 | 0 |
| off (`main.c`) | 1 | 1 |
| wall s | 29.04 | 49.21 |
| CPU s | 62.77 | 139.85 |
| link s | 8.34 | 14.63 |

Control, to prove a hit was possible at all: the unedited build twice more with the cache warm. Base hit 94 of 94
(wall 26.56, 27.51 s; link 8.65, 8.47 s); head hit 1486 of 1486 (wall 33.12, 33.49 s; link 13.78, 14.52 s). So the
cache works on both trees, and the edit is what missed every unit.

Why every unit missed on head, read from the code and the logs: each of head's 1487 units kept its name across the edit
(names are by declaration and contents, so the existing units' C is the same), and every one of them has a different key
(the one unit with the same key is `main.c`, keyed empty). `newUnitKeying` in `internal/native/unitkey.go` keys every
unit by the digest of each file in `common` (line 391: "The units' own directory holds only common, keyed by name and
contents"), and `common` holds `units.h`, the one header that declares every unit's functions. The new function adds a
prototype to `units.h`, which moves every unit's key. I did not diff the edited `units.h` itself; that it gains the
prototype is inferred from it declaring every function (its last lines are the program's prototypes and `main`).

### Where head's extra time goes (inference)

On the lint port, the unit compiles at four jobs take roughly sum / 4: 2.7 s on base, 15.7 s on head, about 13 s more.
The link is 5.7 s more. That accounts for about 19 of the 22.6 s; the remaining 3 to 4 s are outside clang's unit and
link lines (emission of 15x the files, key computation and the like), not measured separately here. Even fully cached,
head's lint build is 6 s slower than base's (33.1 against 26.6 s), most of it the link.

## What was not covered

- One box, one jobs setting (4). No other `ADAMIC_NATIVE_JOBS` and no macOS.
- One kind of incremental edit (a new function). An edit to a function's body, which leaves `units.h` alone, was not
  run, and it is the case head's stable names would serve.
- The binaries were built but not run; this measures the build only.

## Exact commands

Setup:

```
git fetch origin main runtime/wb-batch-stable-identities-v3
git worktree add --detach /tmp/bench/base 926e19f4
git worktree add --detach /tmp/bench/head c33dafb7
(cd /tmp/bench/<tree> && git submodule update --init --recursive --depth 1 && go build -o /tmp/bench/<tree>/adamic ./cmd/adamic)
mkdir -p /tmp/adamic-gate && chmod 1777 /tmp/adamic-gate
```

Environment for every build (`scripts/env.sh`):

```
PATH=/tmp/dl/llvm/bin:/tmp/dl/node/bin:/usr/local/go/bin:/usr/bin:/bin:/usr/sbin:/sbin
TMPDIR=/tmp/adamic-gate
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4
ADAMIC_BUILD_CACHE_DIR=/tmp/bench/cache-<tree>
ADAMIC_BUILD_LOG=<fresh file per run>
(ADAMIC_BUILD_OUTPUT_DIR and ADAMIC_GATE_UNCACHED unset)
```

Warm-up, once per program and tree (`scripts/warm.sh <tree> <program>`):

```
./adamic build stage1/cohere/<program>/main.ts -o /tmp/bench/out/bin-<tree>-<program>
```

Each timed run (`scripts/run.sh <tree> <program> <n>`):

```
rm -rf "${cache:?}/native-units"
/usr/bin/time -v -o time-<program>-<tree>-<n>.txt ./adamic build stage1/cohere/<program>/main.ts -o /tmp/bench/out/bin-<tree>-<program>
```

The loop: `for p in lint tsprinter json yaml; do for n in 1 2 3; do for t in base head; do run.sh $t $p $n; done; done; done`.

Incremental and control: `scripts/incremental.sh <tree>` then `scripts/control.sh <tree>`, base first. Tables:
`python3 -I scripts/summarize.py`.

## Files

- `declaration-units-cold/raw/build-<program>-<tree>-<n>.log`: the build logs, plus `-inc-cold`, `-inc-warm` and
  `-control-<n>` for lint.
- `declaration-units-cold/raw/time-*.txt`: GNU time output; `load-*.txt`: load average before and after;
  `stdout-*.txt`: the build's own output; `warm-*`: the warm-up builds; `runs.txt`: the loop's lines;
  `incremental-<tree>.diff`: the edit.
- `declaration-units-cold/scripts/`: the scripts above, as run.
