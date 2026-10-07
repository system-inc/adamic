# Slot 02 third helper batch

Three public helpers, one per `.a` file. No rule implementation or shared registration/harness file is changed.

| File | Contract |
|---|---|
| `create_class_name.a` | `isCreateClassName(name)` accepts exactly createClass and createReactClass, case sensitive and with no prefix/suffix tolerance. |
| `es5_component_call.a` | `isEs5ComponentCall(nodes, node, isIdentifierNamed)` accepts bare factory calls and React-namespaced factory calls. It skips nested callee parentheses and delegates the namespace predicate to the supplied existing helper. Member names must be Identifiers. Nil and other node kinds decline. |
| `math_function_name.a` | `isMathFunctionName(name)` implements exact membership in Go's 19-entry mathFunctions list. It does not substitute the broader hasMathFunction substring probe. |

The React node arena is `../../slot02_ast.a`: immutable kinds, real Go parser identifier text, numeric expression/name edges and -1 for nil. The namespace predicate takes the exact receiver node and wanted name React. It must implement the other slot's isIdentifierNamed contract, including skipping receiver parentheses. The caller owns the full AST projection. Call expressions require finite callee chains and a real callee; Go SkipParentheses panics for a malformed missing expression. The port explicitly panics for missing/cyclic arena indices. Malformed factory graphs and exact Go internal panic prose are outside normal parity coverage.

The oracle-only react exporter exposes the real private factory-name and identifier predicates. A temporary overlay wraps only IsEs5ComponentCall's dependency call site to observe the predicate count, receiver identity and wanted name, then delegates to the unchanged real predicate. No tracked cohere source is modified. The math exporter calls the actual private helper and derives positive/boundary/case controls from Go's own table, then collects actual function names from Go ParseValue over every captured consumer source and literal text.

There are 583 captured rule/file/source inputs across all twelve consumers. The test refuses missing coverage against the frozen ledger, deduplicates parser inputs by source and extension and adds controls. The oracle parses 580 distinct sources, projects 11,627 nodes, checks 371 factory names and 1,247 math names. All expected-output fields are removed from the JSON runtime input. Actual Go output must match Node source, emitted JavaScript and ASan/UBSan native. Each credited mutant compiles and exits successfully on all backends, then differs from Go.

The capture script, adapted from the earlier slot 03 workflow, records runtime fixture inputs including dynamically assembled strings with temporary Go overlays. React package capture passes; the eight known Tailwind live/corpus checks fail for unavailable external installations or empty corpora. Those live failures are not reported as passing rule gates. Every selected consumer is observed; unknown failures or missing coverage stop regeneration. Repeated capture regenerates both artifacts byte for byte.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch3/testdata/regenerate.py > /tmp/slot02-batch3-regenerate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch3$' -count=1 -v -timeout=20m > /tmp/slot02-batch3-tests.log 2>&1
```

See REPORT.md for exact commands, results and mutants, RULES.md for consumers, and readiness.json for residual dependencies. StringAttributeValue was withdrawn after an earlier slot 01 claim appeared; no duplicate implementation is delivered or counted.
