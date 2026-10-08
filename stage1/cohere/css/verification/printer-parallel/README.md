# CSS printer parallel verification

Base: origin/main `45487a809f89885a3fc651cd590e7dabf31362dc`.
Cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`;
TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`.

## Checks and shared state

`askedCases` uses an overlay in Go cohere's PostCSS package to collect test
strings, repository CSS/SCSS/LESS files, deterministic generated inputs and
truncations, in both CSS and SCSS modes. It writes a corpus and raw parser
answers once. The printer test copies the composed TypeScript port once and
loads/lowers it with stage 0 once.

For each of the unchanged `default` and `narrow` modes, a separate Go overlay
formats the entire corpus and encodes successful strings or refusals. The test
compares those exact bytes against native ASan/UBSan output, the source on Node
through `oracle/node.mjs` and `oracle/adamic.mjs`, and the JavaScript backend.
It separately checks leaks, then checks that each semicolon, indentation and
remaining-width mutant disagrees with Go on native and Node. Optional npm and
fork Prettier comparisons remain unchanged.

The two modes and their six mutant subtests now call `t.Parallel`; Go's testing
scheduler therefore honors `-test.parallel`, with no separate process pool.
The corpus and source copy remain shared read-only files. C emission mutates
IR borrowing flags, so emission and the unmutated sanitized native build and
JavaScript emission finish once before the parallel children. Darwin's
unsanitized leak binary is likewise built once before the children. Executables
and emitted JavaScript are shared read-only; runtime globals, including the
oracle's UTF-8 cache and panic state, live in separate child processes. Parent
`TempDir` cleanup waits for all children. Each mode's Go request, overlay and
answers, and each mutant's source copy, IR and binary are private. The shared
native runtime-library cache already uses a mutex and atomic publication.
No check was removed and no existing case was left serial.

## Output proof

Scratch directory: `/tmp/css-printer-proof`. Temporary capture patches and the
measurement script used here are retained beside this report; capture code was
removed from the committed test before the race run.

Every run captured one corpus file and 20 output files: for each of two modes,
Go plus three agreement backends, and three mutants times two backends. All 21
files have **24,014 cases** in every run. There are **48,028 mode/case pairs**;
the test still has two mode subtests and six mutant subtests.

These commands each exited 0 and produced empty diff files:

```sh
diff -r /tmp/css-printer-proof/before-2 /tmp/css-printer-proof/before-8
diff -r /tmp/css-printer-proof/before-2 /tmp/css-printer-proof/after-2
diff -r /tmp/css-printer-proof/before-8 /tmp/css-printer-proof/after-8
```

`outputs.json` records the common case counts, byte lengths and SHA-256 hashes
of every captured file. Native, source Node and JavaScript backend agreed with
Go in both modes; separate LeakSanitizer checks passed. All six mode/mutant
subtests caught disagreement on both native and Node in all four runs (12
catches per run); the logs retain the differing line and byte.

`ADAMIC_CSS_FIXTURES` and `ADAMIC_CSS_PRINTER_LIBRARY` were unset throughout.
The optional external fixture corpus and npm/fork library comparisons therefore
were not exercised; their configuration and skip behavior did not change.
The repository corpus included 165 CSS and 90 SCSS files.

## Same-machine measurements

Linux amd64, Intel Xeon Platinum 8573C; **5 visible/allowed CPUs**, cgroup
`cpu.max = 400000 100000` (**4 CPU equivalents**). Go 1.27.1, Node 24.19.0,
clang 20.1.8. Runs were sequential on the same machine. Go package and native
runtime caches were warmed before timing; test binary compilation was excluded.
Each timed run generated fresh corpus/source/temp files and executed all checks.

| `-test.parallel` | Before wall (s) | After wall (s) | Before/after |
| --- | ---: | ---: | ---: |
| 2 | 630.039 | 323.666 | 1.95x |
| 8 | 508.602 | 376.640 | 1.35x |

`measure.py` measured `time.perf_counter()` immediately around `subprocess.run`
of each precompiled test binary from the CSS package directory, with:

```sh
-test.run='^TestCSSPrinterAgreesWithGo$' -test.parallel=2 -test.timeout=0 -test.v
-test.run='^TestCSSPrinterAgreesWithGo$' -test.parallel=8 -test.timeout=0 -test.v
```

The temporary capture hooks were enabled in all four measurements. `timings.json`
also records child user/system CPU time and exit status. These are single-run
measurements: the two serial baselines vary, and the new limit-8 run is slower
than limit 2 on this four-CPU-quota machine. Logs show parallel pause/continue
scheduling and all six mutant children released at limit 8.

## Final package validation

Passed: `go test -race -count=1 -timeout=0 -parallel=2 -v ./stage1/cohere/css`.
The full package finished with `PASS` in 736.807 seconds and no race reports;
see `race.log`. Formatting and `git diff --check` are clean.

The capture patches use zero context; apply with `git apply --unidiff-zero`
to the base revision or final revision respectively, only in a scratch checkout.
