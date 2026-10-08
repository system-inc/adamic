# @typescript-eslint/consistent-indexed-object-style: waiting on lint-checker/facts

Ported from codex/typeaware-wave-07 at 2a15504f. There, with that branch's own checker bridge, it
matched unchanged Go cohere on 119 of 121 captured `TestConsistentIndexedObjectStyle` cases.

This area's checker does not answer the questions below yet, so every run of this rule stops with
`unsupported checker question`. The facts land only through lint-checker/facts; the rule is
certified when they do. Checker questions it asks:

- `node-symbol-details` (symbol.a): `ctx.TypeChecker.GetSymbolAtLocation(declarationName)` for the anchor symbol, cohere/internal/lint/rules/typescript/consistent_indexed_object_style.go:338.
- `type-reference-graph` (type_reference_graph.a): the type-syntax walk that resolves each identifier with `ctx.TypeChecker.GetSymbolAtLocation(node)` and follows `symbol.Declarations` into other files, cohere/internal/lint/rules/typescript/consistent_indexed_object_style.go:235 and :244.

Two upstream recovery cases also still fail in the stage 1 parser, not the checker:
`interface Foo {\n  [];\n}\n` and `type Foo = { [] };\n` reach primary() in
stage1/typescript/parser/parser.ts and panic on CloseBracketToken, where Go recovers.
