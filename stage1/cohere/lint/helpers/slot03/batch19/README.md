# Expression, binding and switch CFG helpers

Three original cohere Builder methods, each in one `.a` file. These helpers compose explicit backends over named, immutable parsed views. Integer identities name arena nodes, statement lists and CFG blocks; they are not numeric AST kinds. The category strings describe adapter operations. There are no new lint rules or registry changes.

`expression` preserves the optional expression hook before identifier reads, lazy throwable classification, parenthesized recursion, binary/conditional dispatch, prefix/postfix update handling, access/call dispatch, yield operand order, class headers, lazy root classification and ordered child traversal. A missing node runs nothing. Hooks and dependencies must preserve the parsed view; arbitrary AST mutation during a callback is outside this contract.

`patternBind` preserves read/fork/write ordering, holes, wrappers, binding elements, computed keys and defaults, object assignment/shorthand/spread targets, array elements, spread operands and equals defaults. Binding lists and typed binding metadata have separate presence flags; absent fields ignore parked payloads. Object/array literal lists must be valid parsed lists, as the Go methods assume. A wrong Go metadata type is rejected by its assertion before it becomes a view.

`switchStatement` evaluates the discriminant, allocates the after block and all clause bodies, evaluates non-default tests in source order, routes the final test to default or after, then emits clause bodies and fallthrough. Cycle barriers combine cumulative switch breaks and clauses before the first case. The shared state holds the current block and each jump frame's broken flag. Backend callbacks must update both exactly as Go does, preserving enclosing frames and balanced push/pop operations. Clause body callbacks can set the current switch's broken flag through nested statements.

Dependencies supply recursive expression/binding walks, identifier operations and classification, graph allocation/links/entry, cycle-barrier storage, jump-stack push/pop and statement traversal. The helpers do not implement those dependencies or a complete rule runner.

The oracle calls unchanged Go helper bodies in temporary overlays over all four consumers' upstream corpora. Direct callbacks run the real Go implementations; their return values, cursor and jump-stack updates are recorded and replayed by source Node, emitted JavaScript and sanitized native. The compared trace includes each operation, arguments, callback return and before/after cursor. Replay proves the three compositions, while dependency semantics remain outside this unit.

Controls cover nil nodes and lists, hooks present/absent, updates/yield/root dispatch, wrappers and object/array/default binding, switches with empty/default/ordered cases, breaks, labelled abrupt paths and short-circuit tests. Direct per-node controls expose branches otherwise nested beneath a parent trace. One oracle-only method in a temporary TypeScript AST overlay gives a shallow-copied binding node typed nil metadata, the value the original Go guard explicitly accepts; the real parsed node is unchanged. Parked payloads make ignoring that guard observable.

With `/workspace/adamic-tools/env.sh` sourced:

```
python3 stage1/cohere/lint/helpers/slot03/batch19/testdata/regenerate.py > /tmp/slot03-batch19-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch19' -count=1 -v -timeout=20m > /tmp/slot03-batch19.log 2>&1
```

See REPORT.md for the exact coverage, every mutant, ownership and remaining readiness dependencies.
