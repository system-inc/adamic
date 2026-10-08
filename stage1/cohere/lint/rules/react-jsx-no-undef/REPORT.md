# react/jsx-no-undef: waiting on lint-checker/facts

Ported from codex/typeaware-wave-07 at 2a15504f. There, with that branch's own checker bridge, it
matched unchanged Go cohere on 45 captured `TestJsxNoUndef` cases.

This area's checker does not answer the questions below yet, so every run of this rule stops with
`unsupported checker question`. The facts land only through lint-checker/facts; the rule is
certified when they do. Checker questions it asks:

- `node-symbol-details` (symbol.a): `ctx.TypeChecker.GetSymbolAtLocation(reference)` and its `symbol.Declarations` files, cohere/internal/lint/rules/react/jsx_no_undef.go:101 and :129.
