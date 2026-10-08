# no-throw-literal

Landing base: `ad7bd06632f119abc7680719ad3a7d3b71100f58`. A typed node listener for `ThrowStatement`, verbatim messages, and an unchanged Go rule adapter that rejects unsupported options. `TestNoThrowLiteral` captures all seven upstream test functions.

- 48 upstream cases match byte for byte on Go, Node, emitted JavaScript and ASan/UBSan native, including findings, fixed source and empty fix/suggestion payloads.
- Three firing witnesses and nine additional controls agree. Controls include imported/shadowed undefined, Unicode, nested scopes, `.d.ts`/`.d.mts`/`.d.cts` aliases and no-program coverage. Declaration-file controls use real filenames separately because the witness copier retains only the final extension.
- `non Error throw suppressed` compiles and is caught by ordinary comparison on sanitized native, Node and emitted JavaScript.
- Only the shared live checker is used (`symbol-origin`, `scope-locals`). No private checker, stored answers, bridge changes or shared edits.
- Supplemental typed corpus: 558 `.ts` files, 14,670,601 identical bytes. Sanitized native 29.962s; Go 3.376s. Upstream totals: native 2.385s; Go 1.594s. Times include whole-process work and concurrent gate load.
- Package corpus: 1,068 files including `.a`, 31,341,014 identical bytes. Its shared typed-rule coverage is syntax-only; the supplemental run supplies live checking for `.ts` inputs.

The previous standalone certification passed 48 upstream cases and caught the rule mutant on all three port engines. This landing reruns the full lint package with all inputs and a three-hour timeout. Current logs: `/workspace/wave17-area-unit/clean/`; prior supplemental live-checker comparisons: `/workspace/wave17-clean-unit/no-throw-replay/`. Logs and wave suites are absent from the landing diff.
