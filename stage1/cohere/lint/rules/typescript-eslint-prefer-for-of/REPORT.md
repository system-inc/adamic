# @typescript-eslint/prefer-for-of: waiting on lint-checker/facts

Ported from codex/typeaware-wave-07 at 2a15504f. There, with that branch's own checker bridge, it
matched unchanged Go cohere on 97 captured `TestPreferForOf` cases.

This area's checker does not answer the questions below yet, so every run of this rule stops with
`unsupported checker question`. The facts land only through lint-checker/facts; the rule is
certified when they do. Checker questions it asks:

- `node-symbol-details` (symbol.a): `ctx.TypeChecker.GetSymbolAtLocation(identifier)` and its `symbol.Declarations`, cohere/internal/lint/rules/typescript/prefer_for_of.go:313 and :317 (does a body identifier name the loop index).
