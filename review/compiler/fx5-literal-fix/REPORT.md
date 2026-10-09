Built declared union layout conversion for nonempty array literals toward task #8ng1j76, item 130.
Base: fd27bb0e on origin/compiler/fx5-refusals; implementation and evidence commits are on compiler/fx5-literal-fix.
Validation: focused FX5 leaves, full internal/lower, TestCallTargetReaders, counts refresh and integration lane checks; logs are beside this report.
Mutant: restore source layout by excluding ir.Union from contextual selection; flat, assignment and nested native stdout comparisons fail against Node.
Not covered: const tuple-to-array conversion, existing number[] value variance, empty union literals or spread representation conversion.

The object.go nonempty contextual-array path now accepts ir.Union as the destination element representation. arrayLiteral already calls fit on every ordinary element; fit boxes a narrower number as ir.Box. The native array therefore holds union references from allocation onward, exactly as a mixed literal does. Removed the temporary narrower-literal refusal for these converted cases. No backend code changed.

Flat initialization, later assignment and nested all-number literals print each element's typeof and value before mutation, push a string, then read both numeric elements and the new string. Each acceptance leaf calls lowersAndAgreesWithNode and lowersAndAgreesWithNodeNative. Independent source Node observations and witnesses are saved as all-number.a, nested-number.a, source-node.log and nested-source-node.log. Output on Node and both backends is:

```
number 1
number 2
number 1
number 2
string a
```

The const tuple-shaped probe `const items: readonly (string | number)[] = [1, 2] as const` does not take this path. tupleType selects tupleLiteral before elementType, and lowering refuses the tuple-to-array view at main.a:1:45 with NotYet "a tuple where an array goes (as readonly (string | number)[])". TestFX5ConstTupleUsesSeparatePath pins that exact boundary. A push is not valid through its readonly binding. The number[] value variant remains refused by mutable invariance at main.a:2:36, before literal lowering. These are observed separate paths, not claims of conversion support.

source-layout-mutant.diff changes the contextual condition back to !slotless(declared), so a narrower literal is laid out by its source element type. Apply it with git apply, run the mutant command below, and restore with git apply -R. All three acceptance leaves compile and reach the native stdout comparison, failing with native stdout "" versus source Node "number 1\nnumber 2\nnumber 1\nnumber 2\nstring a\n". Generated JavaScript still agrees. The failures are behavioral oracle failures, not clang warnings. The diff is noncompilable review evidence and passed git apply --check.

Commands, with every test's output redirected directly to a log:

- export GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh
- timeout 120 go test ./internal/lower -run '^TestFX5' -v -count=1 -timeout 90s
- timeout 120 go test ./internal/lower -run '^TestFX5(AssignedNumberLiteral|NumberLiteral|NestedNumberLiteral)Agrees$' -v -count=1 -timeout 90s with the source-layout mutant
- timeout 360 go test ./internal/lower -count=1 -timeout 5m -json
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
- timeout 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 9m -args -update-counts
- git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -

Setup succeeded. Timing lines: Node 0.035s, Go 0.038s, submodules 0.118s, markdown dependencies 0.134s, clang 0.308s, Go build 66.329s, cache warm 66.508s, done 66.552s. nproc is 5; cgroup cpu.max is 400000 100000 (four CPUs). setup.log preserves the tool versions and all timing lines.

New fixtures are lowering test sources, not registered oracle fixtures. The counts refresh checks the existing corpus; no new oracle rows are required. No code was copied from cohere. The prescribed fd27bb0e base is retained, including the separate every-narrowing refusals for 127 and 128.

Full internal/lower: PASS, 115.897s. Final TestCallTargetReaders: PASS, 4.500s. Changed leaf times: flat initialization 1.09s, later assignment 0.94s, nested literal 1.10s, mixed control 0.92s, existing number[] value refusal 0.08s, separate const-tuple boundary 0.09s. Full test output is in lower-full.json.gz.
