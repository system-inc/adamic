# react/jsx-no-undef

Upstream: `cohere/internal/lint/rules/react/jsx_no_undef.go:81`. Captured upstream cases: 45. Messages are copied verbatim. New Adamic source is `.a`.

No program reads, exactly as upstream. The adapter decodes JsxNoUndefOptions rather than dropping field 5. The listener is handed opening and self-closing JSX nodes and follows the tag reference structurally. The node-local symbol query has no foreign-file selector; declarations are compared against the current file path, with allowGlobals and .cjs retaining upstream behavior.

The unified upstream comparison, witnesses and compiling/running mutant passed on Go, source Node, emitted JavaScript and ASan/UBSan native. `validation/mutant.log` records each runtime kill; `validation/upstream-capture.jsonl.gz` preserves the asserted upstream sources and options.

Common full-package evidence and remaining claims are recorded in the sibling `react-jsx-no-undef` directory. The first full run exposed the facts branch's stale closed JSX inventory; area already supplies discovery. The final merged run is recorded separately. No harness, checker, parser or shared helper is edited by this unit.
