# nexus/correctness-no-discarded-pure-result

Upstream: `cohere/internal/lint/rules/nexus/correctness_no_discarded_pure_result.go:86`. Captured upstream cases: 22. Messages are copied verbatim. New Adamic source is `.a`.

Program reads are ReadsCompilerOptions and ReadsDefaultLibrary, exactly as upstream. The listener is handed ExpressionStatement nodes. It checks every method declaration and each argument union part using existing node-symbol-details, type-shape and call-count facts. It reports the call range without automatic fixes or suggestions.

The unified upstream comparison, witnesses and compiling/running mutant passed on Go, source Node, emitted JavaScript and ASan/UBSan native. `validation/mutant.log` records each runtime kill; `validation/upstream-capture.jsonl.gz` preserves the asserted upstream sources and options.

Common full-package evidence and remaining claims are recorded in the sibling `react-jsx-no-undef` directory. The first full run exposed the facts branch's stale closed JSX inventory; area already supplies discovery. The final merged run is recorded separately. No harness, checker, parser or shared helper is edited by this unit.
