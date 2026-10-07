# CFG composition helpers, slot 03 batch seventeen

Three helpers port the original cohere Builder method bodies, each in its own `.a` file. The checker/parser supplies named AST views and integer node handles; `-1` represents a nil node. Block handles preserve identity. These handles are not numeric node kinds.

`conditionalExpression` evaluates the condition before remembering `state.cur`, allocates the join and true branch, links and evaluates that branch, then allocates the false branch and links it from the remembered test end. Both branch ends reach the join, which becomes current.

`bindWithDefault` forks around an existing fallback, evaluates that fallback and rejoins before binding the target. A missing fallback binds the target directly. Target and fallback handles are passed unchanged, including nil.

`memberHeader` calls decorators for the member and its function-like parameters, evaluates a computed name when present, then evaluates a property type or an index signature's parameter types and return type. It does not walk a method body, property initializer or method return type here.

`GraphDependencies` supplies graph allocation, edge storage, entry, expression evaluation, binding and decorators. Each callback must update the shared cursor exactly as the corresponding Go operation does. `MemberView` comes from the parser's named fields and predicates; it must preserve parameter order, nil types and the distinction between property and index signature. No graph backend, hook lookup, parser or rule listener is supplied by these three files.

The oracle executes the real unmodified method bodies with temporary Go overlays that record dependency calls. The driver replays their returned block handles and cursor changes while comparing the calls, arguments, before/after cursor and final cursor. This holds helper composition, not the dependency implementation or complete rule findings. All four consumer corpora are captured; some do not exercise conditional expressions. See REPORT.md for observed counts and every mutant.

From the repository root with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/slot03/batch17/testdata/regenerate.py > /tmp/slot03-batch17-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch17' -count=1 -v -timeout=20m > /tmp/slot03-batch17.log 2>&1
```
