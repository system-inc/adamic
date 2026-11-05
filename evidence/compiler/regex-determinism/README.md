# Task #61ya8ya: does `adamic c` lower 040_runtime.a / 081_runtime.a nondeterministically?

Result: not reproduced. 11,120 loaded runs on two builds gave one identical outcome, and the race build reported no data race.

## Box and toolchain (measured)
- 4 cores, Linux 6.18 (Firecracker VM), Go 1.27.1, clang 18.1.3 (not 20.1.8; `adamic c` only lowers and never calls clang, so it does not matter here).
- Builds: `go build ./cmd/adamic` and `go build -race ./cmd/adamic`, each at f735047e and at main's tip 1f7df6a2, with `git submodule update --init --recursive --depth 1`.
- No stress-ng on the box. Load was 4 busy loops (`while :; do :; done`) beside 8 parallel workers on 4 cores.

## Reproduction (measured)
`stress.sh` runs, 8 at a time under load, a shuffled job list: 040_runtime.a x2000, 081_runtime.a x2000, and every `plain/*_runtime.a` (78 files) x20. `one.sh` records the exit code and sha256 of stderr (with the checkout path stripped) and of stdout.

| build | runs | 040 | 081 | other 76 files |
|---|---|---|---|---|
| f735047e | 5,560 | 2,020 x exit 1, one stderr sha, empty stdout | 2,020 x exit 1, one stderr sha, empty stdout | 20 runs each, one outcome each |
| 1f7df6a2 | 5,560 | same | same | same |

- Zero runs exited 0. Every one of the 78 files stops with `RegExp with a nonconstant pattern` at the same position, and every file's 20 runs had one stderr sha. The two builds' per-file summaries are identical (`summary-wA.txt`, `summary-wB.txt`).
- Also tried: other cwd with an absolute path, `ADAMIC_PROGRAM_REGION=1`, `ADAMIC_OWNERSHIP_QUERY=1`, and `adamic build`. All refuse the same way. `adamic c` takes one file only (two files is a usage error).

## Race detector (measured)
`adamic -race` at main's tip, each file 50 times alone and 50 times 8 at a time (100 runs each way): 0 `DATA RACE`, and every stderr line is the NotYet refusal (`race-alone.log`, `race-par.log`).

## Goroutines on the path (read, not measured)
- `grep` for `go func`, `errgroup`, `WaitGroup` and `go name(` in cmd/adamic, internal/load and internal/lower (non-test) finds none.
- load.go:134 and :160 build the program with `SingleThreaded: core.TSTrue`, so typescript-go runs one checker and no checker pool goroutines. lower.go:45 says the same. `GetTypeCheckerForFile` is called from declarations.go:46 and load.go:193 on the calling goroutine.
- So no goroutine touches a checker on `adamic c`'s path. (I did not read cohere's typescript-go beyond what SingleThreaded implies.)

## What decides line 100 (measured, scratch build only)
A scratch build printed, at regexp.go:100, the type flags `GetTypeAtLocation(pattern)` returns and the checker pointer. 400 loaded runs of 040 and 081: all 400 read `flags=32`, which is `TypeFlagsString`. `pattern` is `string` (`adamic types` agrees: `variable pattern: string`), so the Undefined test is false and the code falls to the declaration path, where the initializer `argumentsList[0] ?? ''` is not a constant, giving the refusal. The checker pointer differs per run (a new process each time); it is one checker per process.

## Is line 100's `""`-for-undefined rule sound? (measured + reasoned)
- `TypeFlagsUndefined` is set only on the exact `undefined` type. `string | undefined` has the Union flag, not Undefined, so a pattern whose type merely includes undefined is not taken as `""`; it goes on to the declaration path.
- For exactly `undefined`, `""` matches Node: `new RegExp(undefined)` is `/(?:)/`, the same as `new RegExp("")`.
- In practice Adamic's `RegExp` signature already rejects `undefined` and `string | undefined` arguments with TS2769 (programs a.a and b.a in my scratch run), so the branch is only reachable through narrowing to undefined. I did not find a case where it is unsound when the checker is right. I did not write a test for the narrowed-to-undefined case.

## What could have made the original sweep exit 0 (inferred, not tested)
I cannot see the sweep, so these are guesses:
1. `$?` taken after a pipeline or after a command substitution (`adamic c f | sha256sum`, `$(...)`), which reports the last stage rather than `adamic`.
2. Results keyed by basename. Here 9 basenames repeat across stage1 (not these two), so it would only matter for other files.
3. A different file set: this checkout has 306 `.a` files and 783 `.a`/`.ts` under stage1, not 761, so the sweep may have used another tree, a different sha, or a sibling box state.
4. A stale or different binary (for example, one built before the refusal existed).
5. Output read from a previous run's file because a worker's temp file path was shared.
None of these is established.

## Fix
None made. There is no defect to fix, and I did not add a determinism test: it would pass today and has no cause to turn red at. If the sweep recurs, ask for its script, the binary's sha, and one failing run's stderr.
