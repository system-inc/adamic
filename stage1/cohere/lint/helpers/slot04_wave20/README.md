# JSX and hook search helpers

These three `.a` files port structure.HasJsxOrReactHookCalls, searchForJsxOrHook and descendsForJsxSearch. The wrapper searches from depth zero. The recursive search permits depth 20, recognizes JSX elements, self-closing elements and fragments, checks actual hook calls, and searches children in order until success. The descent predicate rejects nil and switch-statement children. A directly supplied switch root can still be searched; its exclusion applies when entering a child.

SearchNode is a common AST adapter view: presence, named JSX/call/switch classifications and ordered child identities. Numeric identities identify arena slots, not AST kinds. No invented numeric kind registry or rule dispatch is introduced. isHookCall is an explicit dependency on cohere's existing ecmascript/react.IsHookCall and must retain its exact behavior. The helper handles traversal, stopping and the depth boundary; the caller supplies AST access and hook classification. Arbitrary cyclic adapters are outside the tree contract.

The private oracle parses real TSX with pinned typescript-go, projects actual child order and classifications, and calls the real Go methods on the parsed AST nodes and nil. Consumer source texts are parsed directly. Go supplies hook-call answers independently; this unit does not hand-roll a hook matcher. Controls include depth boundaries, all JSX forms, switch/if/loop/try/class/function nesting and hook/non-hook spellings.

From the repository root, source the setup environment and run:

```
python3 stage1/cohere/lint/helpers/slot04_wave20/testdata/regenerate.py > /tmp/slot04-wave20-regeneration.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave20/testdata/capture.py > /tmp/slot04-wave20-capture.log 2>&1
go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave20 > /tmp/slot04-wave20-tests.log 2>&1
```

All adaptations and capture instrumentation use temporary Go overlays. The capture asserts all four inventory consumers have fixtures and all three helpers are reached by the actual Go structure suite. Reachability records are observations of helper entry kinds/depths; they are not complete execution traces or per-rule finding comparisons. The fresh parity oracle independently invokes actual Go on every selected AST node. REPORT.md names commands, mutants, dependency removals and limits.
