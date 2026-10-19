# Evidence for docs/tsc-generated-c.md

Raw data from the run the page reports, kept off main (never submitted).

- `rows.json`, `files.csv`: measure.py's output, one row per file and proxy. `summary.json`, `report.txt`: report.py's.
- `meter.json`: `adamic-meter --adapt --json` over the adapted src/compiler. `pristine.txt`: tsc's 77 files at the pin.
- `trace/`: the -O2 `-ftime-trace` of stage1/cohere/tsprinter (granularity 50 µs) and its per-pass, per-function summary.
- `probes/checker-closure-diagnostics.txt`: `adamic c checker.ts`'s stderr, the 320 diagnostics of the whole closure.
- `probes/probe-*.log`: stock tsc 6.0.3 (`stage3/api`) over the adapted tree with `types: []`, no emit, and `all` = strict, strictBindCallApply, useUnknownInCatchVariables, noUncheckedIndexedAccess, exactOptionalPropertyTypes; `noNUIA` = the same without noUncheckedIndexedAccess; `base` = tsc's own base options. None reports builder.ts:1246.
- `probes/set-iterator-repro.ts.txt` (rename to .ts): fails `adamic types` at 12:52 with TS2345; passes stock tsc with `--strict --noUncheckedIndexedAccess --exactOptionalPropertyTypes --target es2024 --lib es2024 --moduleDetection force`, and fails the same way when `adamic-iterators.d.ts.txt` (the prelude's iterator lines, renamed to .d.ts) is added.
- `logs/`: apply, meter and measure logs.
