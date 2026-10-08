# no-throw-literal

Base: `95968dd93ad0876245f181af63931c3134e14b3d`. A typed node listener for `ThrowStatement`, verbatim messages, and an unchanged Go rule adapter that rejects unsupported options. `TestNoThrowLiteral` captures all seven upstream test functions.

- 48 upstream cases match byte for byte on Go, Node, emitted JavaScript and ASan/UBSan native, including findings, fixed source and empty fix/suggestion payloads.
- Three firing witnesses and nine additional controls agree. Controls include imported/shadowed undefined, Unicode, nested scopes, `.d.ts`/`.d.mts`/`.d.cts` aliases and no-program coverage. Declaration-file controls use real filenames separately because the witness copier retains only the final extension.
- `non Error throw suppressed` compiles and is caught by ordinary comparison on sanitized native, Node and emitted JavaScript.
- Only the shared live checker is used (`symbol-origin`, `scope-locals`). No private checker, stored answers, bridge changes or shared edits.
- Supplemental typed corpus: 558 `.ts` files, 14,670,601 identical bytes. Sanitized native 29.962s; Go 3.376s. Upstream totals: native 2.385s; Go 1.594s. Times include whole-process work and concurrent gate load.
- Package corpus: 1,068 files including `.a`, 31,341,014 identical bytes. Its shared typed-rule coverage is syntax-only; the supplemental run supplies live checking for `.ts` inputs.

Full lint package: **147 pass, 0 fail, 1 skip** (top level: 41/0/1), all 94 mutants caught, 1575.671s, nproc 5, ending load 2.63/2.28/2.37. All inputs supplied. The inherited `TestCheckerBridgeRefusalPending` awaits `tsgoInspect` returning `TSGoError`; it remains pending. Registry generation, gofmt and vet pass.

Logs: `/workspace/wave17-clean-unit/clean-certified/`; focused comparisons: `../no-throw-replay/`; held failures: `../failures/REPORT.md`; exclusions: `../OMITTED.md`. Logs and wave suites are absent from this diff.
