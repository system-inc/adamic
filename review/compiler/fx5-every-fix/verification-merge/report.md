Merged refusal verification and current main into compiler/fx5-every-fix, retaining all six mixed fixture forms.
Merge commits: b4b82ebf (84915995), dee5161f (main cf735d9f).
Full lower passed 158.326s; reader guard passed 2.968s; lane checks passed 5.4s; count refresh passed 172.746s with no row changes.
Both real-fix mutants compiled and failed external-oracle stdout comparisons; both were reverted before verification.
No new compiler behavior or fixtures added; existing filter/write conversion limits remain.

There were no textual merge conflicts. All six verification fixture forms were retained because they isolate post-narrowing reads from literal construction. The extra mixed member is popped before the narrowing operation:

| Fixture | Kept form and path exercised |
| --- | --- |
| cat | [1, 2, { name: 'discard' }], pop; Cat-or-number predicate then numeric reduction. |
| forward | [1, 2, 'discard'], pop; narrowed union storage passed to a number[] parameter. |
| inferred | [1, 2, 'discard'], pop; inferred every predicate. |
| property | obj.items built mixed, then obj.items.pop(); narrowing through a property binding. |
| readonly | Mutable mixed built array, pop, then readonly union-array binding; readonly every path. |
| some_negated | ['a', 'b', 1], pop; negated some and subsequent string reads. |

The all-number and nested literal fixtures remain separate and continue to exercise destination-layout construction. Existing every, early-return, find, findLast and stored indexed/map/loop reads also passed. Per-leaf seconds are in leaf-seconds.json; all sixteen leaves finished below ten seconds.

Commands, each bounded and logged:
- go test ./internal/lower -json -count=1 -timeout 5m: passed 158.326s, full log lower.jsonl.gz.
- go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: passed 2.968s, readers.log.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 9m -args -update-counts: passed 172.746s, counts.log.
- Required fetch plus integration lane-checks.py after the merge commits: passed 5.4s; gofmt/tools eight Go files, t.Parallel one test package, vet one package.

Stored-layout mutant: bypass native arrayCallbackSlot adaptation. TestArrayNarrowingEveryAgrees compiled and failed: native "total 9.3079274943955e-310\n", Node "total 3\n". Source-type mutant: exclude union from destination contextual layout selection. TestFX5NumberLiteralAgrees compiled and failed: native stdout empty, Node "number 1\nnumber 2\nnumber 1\nnumber 2\nstring a\n". Patches and complete failure logs are saved here. An initial source mutant command used an unmatched test name; it was corrected and the reported run executes TestFX5NumberLiteralAgrees.
