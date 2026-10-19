# Emit's two heads on #nda35wb, clocked

clang 20.1.8 wall seconds and peak resident memory for the C that four `adamic c` builds emit for four stage1
programs. Each number is the median of three runs. Every run is in `rows.json`, and `clock.py` is the script
that produced them.

| tree | sha | what it is |
|---|---|---|
| A | `2b22b31ccc8b` | main, the base of emit/main-parts |
| B | `71fcee578a` | emit/main-parts: main's top-level code cut into parts of about 4,000 lines |
| C | `141fec7702d4` | main, the base of emit/constant-tables |
| D | `375a979e17` | emit/constant-tables: an array literal of 64+ constant numbers becomes a static table and a loop |

Compare B with A and D with C. A and C emit byte-identical C for all four programs (checked with `cmp`):
C is one commit after A and changes only the JSON runtime, which shows up as a small difference in the
included headers. That makes A against C a measure of the box's own noise.

## How it was measured

- Each tree is a `git worktree add --detach` checkout with `git submodule update --init --recursive --depth 1`
  (cohere `0183ccf8c57a` in all four) and `go build -o <tree>/adamic ./cmd/adamic`. All four built.
- The C comes from `<tree>/adamic c main.ts`, run from the program's own directory, with stdout as the C.
  All 16 emits exited 0. The entry points are the `main.ts` files named in the task.
- clang runs with measure.py's `FLAGS` (from `origin/compiler/tsc-generated-c-table-2`), plus
  `-O0|-O1|-O2 -I <tree>/internal/native/runtime -c unit.c -o unit.o`, under `/usr/bin/time -v`.
  `clang_once`, `functions()` and `time_trace_top` are copied from measure.py unchanged, except that
  clang_once takes the tree's runtime directory as an argument.
- One compile runs at a time. Within each program and level, the repeats are interleaved across trees
  (A B C D, A B C D, A B C D), so a drift in the box falls on every tree alike. Each figure is the
  median of three. All 144 timed compiles exited 0.
- "Largest function" is measure.py's `functions()`: a top-level definition from its opening line to the
  column-0 brace.

## tsprinter (`stage1/cohere/tsprinter/main.ts`)

| tree | C bytes | C lines | O0 s | O0 MB | O1 s | O1 MB | O2 s | O2 MB |
|---|---|---|---|---|---|---|---|---|
| A 2b22b31ccc8b | 4,429,107 | 59,096 | 1.04 | 177.4 | 21.72 | 307.0 | 22.62 | 308.2 |
| B 71fcee578a | 4,429,327 | 59,112 | 0.84 | 151.3 | 14.35 | 262.9 | 14.95 | 264.2 |
| C 141fec77 | 4,429,107 | 59,096 | 1.00 | 177.7 | 22.90 | 306.7 | 22.47 | 308.3 |
| D 375a979e17 | 3,062,408 | 40,764 | 0.49 | 107.1 | 5.57 | 156.3 | 5.80 | 157.1 |

## json (`stage1/cohere/json/main.ts`)

| tree | C bytes | C lines | O0 s | O0 MB | O1 s | O1 MB | O2 s | O2 MB |
|---|---|---|---|---|---|---|---|---|
| A 2b22b31ccc8b | 1,879,342 | 25,248 | 0.61 | 147.9 | 13.67 | 345.7 | 13.86 | 345.4 |
| B 71fcee578a | 1,879,617 | 25,268 | 0.44 | 121.5 | 9.72 | 233.0 | 10.54 | 233.1 |
| C 141fec77 | 1,879,342 | 25,248 | 0.61 | 147.9 | 12.61 | 345.7 | 14.26 | 345.3 |
| D 375a979e17 | 505,131 | 6,510 | 0.08 | 66.6 | 0.63 | 81.4 | 0.70 | 80.7 |

## yaml (`stage1/cohere/yaml/main.ts`)

| tree | C bytes | C lines | O0 s | O0 MB | O1 s | O1 MB | O2 s | O2 MB |
|---|---|---|---|---|---|---|---|---|
| A 2b22b31ccc8b | 3,371,301 | 47,565 | 0.64 | 138.2 | 11.44 | 261.4 | 12.25 | 261.6 |
| B 71fcee578a | 3,371,466 | 47,577 | 0.63 | 134.1 | 11.20 | 242.8 | 12.24 | 242.8 |
| C 141fec77 | 3,371,301 | 47,565 | 0.64 | 138.0 | 11.65 | 261.2 | 12.53 | 261.5 |
| D 375a979e17 | 2,498,468 | 35,854 | 0.38 | 97.9 | 3.89 | 133.8 | 4.23 | 132.6 |

## graphql printer (`stage1/cohere/graphql/printer/main.ts`)

| tree | C bytes | C lines | O0 s | O0 MB | O1 s | O1 MB | O2 s | O2 MB |
|---|---|---|---|---|---|---|---|---|
| A 2b22b31ccc8b | 1,744,917 | 25,134 | 0.41 | 120.7 | 8.32 | 237.2 | 9.04 | 237.7 |
| B 71fcee578a | 1,745,082 | 25,146 | 0.39 | 116.3 | 8.06 | 219.4 | 8.86 | 220.8 |
| C 141fec77 | 1,744,917 | 25,134 | 0.42 | 120.6 | 8.43 | 237.2 | 8.85 | 237.8 |
| D 375a979e17 | 885,931 | 13,423 | 0.14 | 71.9 | 1.20 | 90.1 | 1.34 | 91.0 |

## Largest function per tree

| program | A 2b22b31ccc8b | B 71fcee578a | C 141fec77 | D 375a979e17 |
|---|---|---|---|---|
| tsprinter | `main` 19,887 | `adamic_main_2` 11,410 | `main` 19,887 | `adamic_function_183_Statements_statement` 1,693 |
| json | `main` 19,857 | `adamic_main_3` 11,410 | `main` 19,857 | `adamic_function_38_Printer_print` 848 |
| yaml | `main` 12,505 | `adamic_main_1` 11,410 | `main` 12,505 | `adamic_function_207_UnistContext_document` 843 |
| graphql printer | `main` 12,534 | `adamic_main_1` 11,410 | `main` 12,534 | `adamic_function_142_Printer_generic` 988 |

On B, every program's largest part is the same 11,410 lines: one top-level statement, the `emoji` table in
`widthTables.ts`, built as 1,472 nested arrays with about 8,500 `adamic_array_push` calls. The cut can only
fall between statements, so this one part stays that size. In tsprinter, B's parts are 1,879, 3,050,
11,410 and 3,520 lines; the per-program part lists are in `rows.json` under `main_functions`. On D, `main`
shrinks to 541 (tsprinter), 240 (json), 224 (yaml) and 253 (graphql) lines.

## Time trace, tsprinter at -O2

One run per tree with `-ftime-trace -ftime-trace-granularity=50`, reading the slowest `OptFunction`:

| tree | run wall s | slowest function | seconds |
|---|---|---|---|
| A 2b22b31ccc8b | 22.99 | `main` | 14.558 |
| B 71fcee578a | 14.76 | `adamic_main_2` | 6.417 |
| C 141fec77 | 22.15 | `main` | 13.992 |
| D 375a979e17 | 6.09 | `adamic_function_179_Statements_classDeclaration` | 0.291 |

On D, `main` does not appear in the trace's top three functions at all. Each trace's full outline and
passes are in `rows.json` under `time_trace`.

## The box

- One Linux VM for the whole run (a Firecracker guest in a cloud container), kernel `6.18.44-fc-v139`.
- `nproc` 4, lscpu model name `Intel(R) Xeon(R) Processor` (the guest does not expose the exact part),
  16,479,424 kB memory.
- `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`,
  from the official `LLVM-20.1.8-Linux-X64.tar.xz` release. The tarball's sha256 `1ead36b3...ae6` matches the
  digest in the release's attestation `.jsonl`. I compared digests only; I did not verify the sigstore
  signature. The system's own clang is Ubuntu 18.1.3 and was not used.
- go1.27.1, as `go.mod` asks.
- `/usr/bin/time` was missing, so I installed GNU time 1.9 from Ubuntu for the run.
- Load average, 1, 5 and 15 minutes, before and after each program:

| program | before | after |
|---|---|---|
| tsprinter | 0.87 2.64 1.73 | 1.00 1.31 1.41 |
| json | 1.00 1.31 1.41 | 1.00 1.14 1.32 |
| yaml | 1.00 1.14 1.32 | 1.02 1.08 1.25 |
| graphql printer | 1.02 1.08 1.25 | 1.00 1.04 1.20 |

A 1-minute load of 1.0 is clang itself on one of the four cores. Every clang run's before-and-after load
is in `rows.json`.

## What was odd

- **The box is not a dedicated machine.** It is a cloud VM, and the Claude Code agent driving the run
  was resident on it (mostly idle while clang ran). I can't vouch for noisy neighbours on the host. The
  noise shows: A and C compile the same C, and their medians differ by up to 8% (json -O1 13.67 against
  12.61, tsprinter -O1 21.72 against 22.90).
- **The start was not fully quiet.** The 5-minute load was 2.64 when tsprinter began, left over from
  building the four trees and a trial emit just before. The 1-minute load was already 0.87, and the
  tsprinter -O0 runs that came first agree within 0.05 s, so I don't think it moved those numbers.
- **Repeats that disagree** (more than 10% from their median). All of them are slow outliers, and the
  medians drop each one:
  - graphql printer B -O0: 0.71 against 0.37 and 0.39.
  - graphql printer C -O0: 0.49 against 0.41 and 0.42.
  - yaml C -O0: 0.72 against 0.62 and 0.64.
  - yaml C -O2: 13.82 against 12.22 and 12.53.
  - yaml D -O2: 5.15 against 4.23 and 4.22.
  - json B -O1: 10.79 against 9.72 and 9.66.
  - json C -O0: 0.68 against 0.59 and 0.61.
  - json D -O0: 0.10 against 0.08 and 0.08. GNU time reports wall time to 0.01 s, so the sub-second
    -O0 figures carry about one hundredth of a second of resolution error.
- **Within noise:**
  - B against A on yaml (-O2 12.24 against 12.25) and graphql printer (8.86 against 9.04).
  - B's peak memory on those two still falls by about 19 and 17 MB.

  The likely reason is the 11,410-line part above: on these programs main was 12,505 and 12,534 lines,
  so cutting it barely changes its largest piece.
- Nothing failed: all four trees built, all 16 emits and all 148 clang runs (144 timed and 4 traces)
  exited 0.
