# @typescript-eslint/consistent-indexed-object-style: blocked

Source ba85c35ba; facts d845dccde are already an ancestor.
The private symbol decoder is replaced by shared SymbolDetails. programReads is empty, matching cohere/internal/lint/rules/typescript/consistent_indexed_object_style.go:113.
The missing accessor is type-reference-graph, required by the circular type walk at cohere/internal/lint/rules/typescript/consistent_indexed_object_style.go:235 and :244, with anchor symbol resolution at :338.
Small input: type T = { [k: string]: number };
Live Go exits 0; sanitized native exits 70 with unsupported checker question: type-reference-graph. Reproducer and outputs: /workspace/wave-23/facts-wave07-evidence/blocked/.
The aggregate upstream comparison matches 21 of 121 cases before this refusal. This is partial evidence, not complete certification.
The message mutant compiles, then exits on the same missing fact before the comparison; it is not counted as caught.
Two upstream recovery cases also remain blocked: type Foo = { [] }; and interface Foo { []; }, at cohere/internal/lint/rules/typescript/consistent_indexed_object_style_test.go:112 and :114. The shared parser panics at stage1/typescript/parser/parser.ts:176.
The private bindings/graph adapters remain while these shared contracts are missing. No shared helper or substitute checker fact is added.
Complete lint log: /workspace/wave-23/facts-wave07-lint-complete.jsonl.

Full package: 26 top-level passes, 8 shared failures, one TestCheckerBridgeRefusalPending skip; 84 registered mutants caught and the indexed mutant blocked. Wall 2400.15s, nproc 5. No input skips or relaxed checks.
