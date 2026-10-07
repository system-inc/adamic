# Slot 02 fifth retained helper batch

| File | Contract |
|---|---|
| design_system_has_utility.a | hasUtility(system, root, kind) gives repository root presence precedence over framework fallback, including empty kind maps and false flags. Static framework lookup uses key presence; functional lookup uses boolean value. Other kinds decline unless the repository explicitly declares them. |
| find_roots.a | findRoots(input, exists) eagerly collects an exact match, registered dash prefixes longest-first and a final @ fallback. Presence is separate from empty value. It preserves callback order and short circuit behavior. |
| property_sort.a | propertySort(nodes, compute) delegates once with the original node list to the separately owned private propertySort computation. It copies the count and the order slice header (length and capacity), sharing the underlying values, preserving Go struct/slice semantics. |

HasUtility receives generated framework maps explicitly: static values describe whether the declaration sequence is nonempty, but only presence matters; functional boolean flags are consumed as values. The loader owns repository maps. A declared nil Go kind map is represented by an empty map, which has the same read-only observations. The helper reads the supplied maps at each call. No guessed framework roots or stale snapshots are built into the implementation.

findRoots uses string slicing only at ASCII hyphens and the leading @, so valid Unicode text stays intact. It preserves Go’s actual eager behavior even where its comments describe a generator or suggest that @ fallback is suppressed after a dash stop. The port does not replace the actual final fallback with that comment’s promise. Stateful predicates are supported: repeated @ queries and short circuit order remain observable.

PropertySort is the public adapter, not the broader private breadth-first traversal. compute receives the caller’s exact numeric node-id list and returns {sort: {order: {values, length, capacity}, count}, visited}. The slice projection must satisfy 0 <= length <= capacity <= values.length; callers index only inside length, and use capacity for append semantics. The caller owns AST projection and private traversal implementation. The oracle supplies actual Go dependency results for independent testing. Count mutation of the dependency result does not change the returned count; element mutation of its order backing array remains visible. Reslicing the dependency result to length/capacity zero does not change the returned header. Nil and empty zero-capacity slices share the empty projection; exact pointer identity of empty Go slices is not observable here.

The Go oracle uses temporary overlays and invokes the actual helpers. A wrapper at the public PropertySort call site records call count and input-slice identity, then calls the unchanged real private traversal. No tracked cohere file changes. 157 actual runtime rule/file/source inputs cover every consumer; the TypeScript parser supplies decoded literal/template text. Supplementary Go candidate/design-system fixture literals are static controls. Actual ParseCandidate supplies more roots. All framework registry entries, five repository states, eager/stateful predicate modes and actual parsed CSS/framework static declaration trees are compared. Expected output is removed before Adamic reads the corpus.

Baseline and compiling semantic mutants must agree on Node source, emitted JavaScript and ASan/UBSan native, then compare with actual Go. Compile failures and crashes are not credited as semantic mutant kills. Eight external Tailwind live/corpus capture gates remain unavailable; capture is not counted as a passing integrated rule gate. Missing consumers and unexpected failures reject regeneration. Artifacts reproduce byte for byte.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch5/testdata/regenerate.py > /tmp/slot02-batch5-capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch5$' -count=1 -v -timeout=20m > /tmp/slot02-batch5-tests.log 2>&1
```

See REPORT.md for observations and limits, RULES.md for each consumer and readiness.json for residual dependencies. The earlier claimed HasVariant and VariantKind were withdrawn before any implementation; neither is delivered here.
