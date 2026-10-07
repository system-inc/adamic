# Slot 02 second helper batch

Three public helpers, one per `.a` file. The exports consume the numeric node arena or explicit prerequisites; no rule is implemented and no shared harness or registration file is edited.

| File | Contract |
|---|---|
| `class_literals_in.a` | `classLiteralsIn(node, classValuesIn)` invokes the supplied prerequisite once and returns its exact shared, read-only `literals` array. Templates are excluded. The caller represents a Go nil slice with an empty array. Literal contents, node identity, ranges, origins and edges are owned by the supplied values projection. |
| `split_classes.a` | `splitClasses(text)` implements Go `strings.Fields` and drops every entire field containing `${`. The exact 25 Unicode whitespace code points are supported, including U+0085 and U+00A0; U+FEFF and U+200B are retained inside a field. Return order, duplicates and remaining text are unchanged. Input is a decoded Unicode string, not an arbitrary invalid UTF-8 byte sequence. |
| `namespaced_member.a` | `isNamespacedMember(nodes, node, matches)` accepts only a PropertyAccessExpression whose receiver, after skipping parentheses, is the Identifier `React`, and whose member is an Identifier satisfying the supplied predicate. Nil node declines; the outer node is not unwrapped. Optional property access follows Go's kind. Predicate runs exactly once after guards, otherwise never. |

The node projection is `../../slot02_ast.a`: Go kind names, real decoded AST identifier text, numeric expression/name edges and -1 for nil. Parenthesized receiver chains are finite. Missing indices, cycles and a missing property receiver panic explicitly. Go itself panics for a missing receiver inside SkipParentheses; malformed factory graphs and exact internal panic prose are outside parity coverage. The production AST adapter remains rule-worker territory.

The test captures dynamically assembled source inputs from the real consumer tests. Its workflow is adapted from slot 03's capture workflow using temporary Go overlays; no cohere worktree file is modified. Typed and live-engine wrapper sources are captured before external checks. All 20 consumers are present. React and structure package tests pass; the eight expected Tailwind live/corpus tests fail for missing external installations or empty corpora, so that run is not a passing Tailwind rule gate. The regeneration script refuses unexpected failures or missing consumer coverage.

The independent Go oracle parses 559 distinct sources including controls, projects every node and calls actual Go IsNamespacedMember with three predicates, actual SplitClasses on every distinct literal text and 1,114,112 separator candidates, and actual ClassLiteralsIn on a bound reader. Its private prerequisite slices are compared with the public slice for sharing. No expected-output field is passed to the Adamic driver. The same corpus runs on Node source, emitted JavaScript and ASan/UBSan native. Mutants compile and execute on all three, agree with one another and differ from Go. A compiler refusal or runtime error is not credited as a semantic mutant.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch2/testdata/regenerate.py > /tmp/slot02-batch2-regenerate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch2$' -count=1 -v -timeout=20m > /tmp/slot02-batch2-tests.log 2>&1
```

See REPORT.md for commands, timings, mutants and limitations, RULES.md for every consumer and readiness.json for residual dependencies. The original shared readiness ledger is unchanged.
