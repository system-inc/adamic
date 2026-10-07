# CFG expression, class and branch helpers

Three cohere Builder methods ported to separate `.a` files. Each takes named parser fields and classifications, integer node/block identity handles and explicit backend callbacks. The handles are not numeric AST kinds. Node `-1` is nil; valid block handles are provided by the backend.

`binaryExpression` preserves the ordered categories: `&&`/`||`/`??`, logical/coalescing assignment, plain equals, other assignment, then other binary expressions. The first two construct the short branch and right branch before joining. Plain equals calls `isDestructuringTarget` lazily; a destructuring assignment evaluates the right side before binding, while other assignments read the left side, evaluate the right and write the left. Other operators evaluate left before right.

`classLike` returns immediately for missing class metadata, then visits decorators, type parameters, heritage types in source order and member headers. Presence flags control whether associated flattened payloads are read. Payloads can contain parked arena entries when a field is absent; those entries are ignored. A heritage view must come from a typed Go heritage node: Go rejects an unrelated node via its type assertion. The view's `present` field represents a missing typed value, not an arbitrary wrong node kind.

`ifStatement` allocates the after and then blocks, allocates a separate else block only when an else statement is present, asks the condition backend for the two continuations, evaluates and links the then branch, optionally evaluates and links the else branch, then enters the after block.

`FlowDependencies` supplies expressions, allocation, edges, entry, reads/writes/binding, destructuring classification, decorators, type parameters, member headers, condition branching and statement traversal. Views describe one immutable parsed AST; callbacks must preserve the view payloads and update the shared cursor exactly as Go does. These three files port composition; they do not implement these dependencies or a full CFG builder.

The owned oracle executes actual Go bodies in temporary overlays, capturing their direct callbacks and before/after cursor changes on all four consumer corpora. The driver replays dependency results and cursor changes and compares its complete operation trace and final cursor against Go on source Node, emitted JavaScript and sanitized native. Controls cover all operator categories, nested/abrupt/absent else branches, class headers, and absent class/list/type fields with parked payloads. No shared harness or compiler file changes.

```
python3 stage1/cohere/lint/helpers/slot03/batch18/testdata/regenerate.py > /tmp/slot03-batch18-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch18' -count=1 -v -timeout=20m > /tmp/slot03-batch18.log 2>&1
```

Source /workspace/adamic-tools/env.sh before these commands. Exact consumers, actual live-call counts, every mutant and integration limits are in REPORT.md.
