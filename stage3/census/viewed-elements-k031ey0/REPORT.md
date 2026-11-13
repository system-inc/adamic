# Viewed-elements entry census — k031ey0

## Answer

Does head show `every=true` with a non-empty trip anywhere that main shows `every=false`? **No such observation was produced. The underlying yes-or-no question is undetermined by this census, not a pass.** Both sides are missing all `viewed-elements:` lines. There is no main `every=false` witness and no head trip witness to compare.

Both separately built entry-census binaries exited **0**. Each loaded **81** source files through the adapted `src/tsc/tsc.ts`, with **zero checker diagnostics**, and attempted the selected 101-byte file's single `KindExportDeclaration` unit, with no findings. Full mode was used because only that mode honors `LATENT_ONLY_FILE`. The common adapted tree was built once with main's `stage3/apply.sh`; `git diff --exit-code c07d3b5204 a16bedb211 -- stage3` succeeded.

## Counts

Every line beginning `viewed-elements:` in each complete stderr was collected and validated by `count.py`. Totals: **main 0; head 0**. `counts.json` records all classes and the complete (empty) matching line lists.

| every-on-entry | every | trip non-empty | main | head |
|---|---|---|---:|---:|
| false | false | false | 0 | 0 |
| false | false | true | 0 | 0 |
| false | true | false | 0 | 0 |
| false | true | true | 0 | 0 |
| true | false | false | 0 | 0 |
| true | false | true | 0 | 0 |
| true | true | false | 0 | 0 |
| true | true | true | 0 | 0 |

## Verbatim trip lines and counterparts

There are **no head lines with a non-empty trip** to quote, and **no main viewed-elements lines** to place beside them. No line was discarded or truncated. Both stderr files are retained in full (119 bytes each); each contains only this same ordinary census completion line:

```text
/workspace/k031ey0/adapted/src/compiler/_namespaces/ts.performance.ts 0 measured on a checker-clean entry-root program
```

## Why the observations are missing

This is a census/probe wiring limitation, not a census process failure or checker rejection:

- In both pinned trees, `stage3/census/latent/units.go.txt:129` constructs `fresh()` as `&lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}`. It never initializes `elementKey`. `base := fresh()` is used for registration.
- The full-mode path at `stage3/census/latent/units.go.txt:154` calls `base.collectViewedElements(modules)` after module ordering. With no registration finding and a completed attempted unit, that path completes, but the collector does no type walk.
- Main's `internal/lower/view_arrays.go:102` and head's `internal/lower/view_arrays.go:122` read `key := l.elementKey`. The logger is installed only when the environment flag is set **and `key != nil`**. Immediately afterward, `if key == nil || key.every { return }` returns for this census's nil key.
- The production lowering path normally supplies an element key; the measurement overlay replaces production `Lower` with the no-IR guard. It does not supply a key for the full-mode census lowering.

Consequently, although the full program was loaded and the census reaches the collector call, this run does **not** exercise the whole-program viewed-element walk or head's nesting rule. Zero logged classes cannot establish that head leaves `every` false, sets it true, or matches main. No source or overlay was altered to repair this measurement, and no additional experiment was substituted.

## Reproduction and artifacts

See `PREPARATION.md` for exact setup commands and pins, and `measure.sh` for the exact sequential overlay builds and census invocations. Builds follow the meter's latent-overlay then entry-overlay recipe, with `-p=1` to serialize package compilation. Each census is bounded by `timeout --preserve-status 900`; neither timed out. The wrapper printed and saved each child's own exit code. Preparation, builds and runs were sequential and completed within about 15 minutes of environment startup.

Both full stderr files, stdout files, exit codes, binary SHA-256 files, JSONL census records, preparation logs and deterministic counts are included. `SHA256SUMS` hashes the evidence files. No log exceeded 20 MB, so no log reduction was needed. Only this evidence directory is added to the evidence commit; its parent is c07d3b5204082b9d862091c9e53ce241ab65670d.
