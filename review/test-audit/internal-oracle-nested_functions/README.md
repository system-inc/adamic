# u064: internal/oracle nested functions audit

Starting origin/main: 553ad06abb069736161b2d46532d9036e92f30b3.
All six requested tests exist and remain in their two named files. No family grouping applies to these six. The comparison fixture table remains one row, TestNativeAgreesWithNode, restricted here to 46 fixture subtests. The scope and all top-level comparators are listed in matrix-scope.json; actual selected members are in control.json.

Read code-under-test.txt for the complete 429-function reached inventory, functions-covered.txt and coverage.out for block evidence, callers.txt for static caller evidence, and plan.json for the menu frozen before any mutant result. The frozen SHA256 is a8da1f2255bc04fcea73b28393cf66a15910a5c20d5f06c36b35f37923828158. replay-plan.json contains only two compile repairs described in diff-repairs.json. Every production mutant changes Adamic code. W01 and W02 are the specifically permitted harness edits for the two witnesses. The other witness rows in the matrix are comparators only, not audited production defenders.

## Results and scope

results.json contains the six deliverable rows. matrix.json normalizes failed subtests to top-level rows. mutants.md includes every standalone diff's original line and exact change. No package-unique kill is proven. Three cannot-judge verdicts mean that useful catches are proven but whole-package uniqueness is unresolved, not that the tests cannot fail. The overlapping verdict rests on three production mutant catches: M08 is caught by Uint16, M09 by the fixture comparison row, M16 by Uint16. No single selected other row catches all three. This is a bounded-set finding and no deletion recommendation.

The complete clean package attempt reached its 90-second binary timeout (92.197 seconds process wall) with no failed-test events before the panic. The clean six-row baseline and clean bounded comparison set both passed. No mutation began before those clean checks. M08 and E00 abort the binary with a panic, so the six unit rows were rerun individually; results outside those six are incomplete for those columns. Neither aborted column supports uniqueness. Other complete columns have unknown kills outside the selected set. The full package and repo-wide replay were not completed.

## Survivors

M11: len(initializer arguments)==0 becomes <=1. cache-one-argument.a changes Lower's output from the additional-capture NotYet at line 17 to a successful IR program. Exact outputs are M11-probe-before.log and M11-probe-after.log. This is changed lowering behavior unguarded by the bounded matrix, not a repo-wide claim.

M15: drops sibling Environment unification at nestedDeclarations. No selected test failed. Final Lower output for nested_mutual.a is byte-identical before/after, SHA256 f5a2b1c2512aa6696be41e5c81439de7aeb6377a34236688bb8036cb9847d799. finishNestedEnvironment later unifies the layouts (functions.go:321; nested_functions.go:308 onward). No changed final-answer witness was established. This is an equivalent-candidate survivor, not demonstrated unguarded behavior. Evidence: M15-probe-before.log and M15-probe-after.log. This fails the brief's requested changed-behavior witness honestly; central replay can examine other programs.

## Empty answer

E00 returns an empty IR Program from Lower at entry. Five unit rows fail; TestNestedRebindingCheckerRefusal passes. That row does not call Lower; its actual entry is Load. Its vacuous=true records the requested probe observation only, and does not demonstrate vacuity of its loader check. It catches all three loader mutants. Exact isolated observations are E00-Test*.json/log. Witness failures under E00 are mutation precondition failures, not evidence about the production check they guard.

## Oracles

Node was actually executed uncached. The two NotYet tests additionally pin self-authored diagnostic labels; a changed label can fail them while source Node stays correct. The cycle refusal's normal successful path does not run Node, so its expected refusal is self. The rebinding code TS2630 was checked with independent TypeScript 6.0.3; tsc-authority.log prints Cannot assign to 'inner' because it is a function. Witness W01 suppresses leakSanitizer's nonempty report; W02 returns an empty disagreement. Both guarded witnesses then fail, as recorded in their logs.

## Commands and replay

All test output was redirected to log files, without pipelines. run.py records actual go test argv, binary seconds, wall seconds, exits, passed and failed tests. It bounds the process at 100 seconds, stricter than the requested outer 120, while the Go test binary has timeout 90s. Every run sourced /workspace/adamic-tools/env.sh. Matrix calls set ADAMIC_GATE_UNCACHED=1 and select ADAMIC_MUTANT. matrix.py is the actual matrix command; it begins with an unmutated control. selector-tracked.diff preserves the switched scratch source, including helper files. It is evidence only and was reverted before commit.

To replay a standalone mutant on the starting commit, apply diffs/MNN.diff without the selector and run the whole package's go tests centrally. W01/W02 are witness probes, not production mutations. All 19 standalone diffs were checked against the unchanged origin/main index. standalone-builds.json records compile verification: initial full oracle test binary compiles were narrowed to the mutated production package for budget; witness probes still compile the oracle test binary. No tests in other packages were run.

## Brief friction and limits

1. The clean-baseline requirement says stop on red, while the big-package rule permits narrowing after 90 seconds. A timeout is a failing go command, but not an observed assertion failure. I used the explicit timeout narrowing exception, documented the incomplete package baseline, and required green slice and matrix controls before mutation.
2. A package-unique verdict cannot be proven by a restricted matrix. I leave unique_kills empty and use cannot-judge for the three rows with single-row bounded kills. Coverage establishes reached functions, but these shared compiler functions have callers across much of the package. Caller searches alone cannot prove the omitted rows irrelevant.
3. Several rows have mixed oracles: real Node results plus self-authored refusal labels. The required single oracle_kind cannot express both; the oracle text states both, and the kind follows the executed external comparison.
4. The main checker is not shared by all six rows. Five use Lower and one directly uses Load. The one required empty-answer probe therefore makes the loader-only row pass without demonstrating weakness in that row.
5. Witnesses must be judged by their guarded harness checks. Their failures on production mutants or empty IR often mean their built-in mutant preconditions broke. Those failures are retained in the raw matrix but excluded from witness verdict reasoning.
6. Panics terminate a package test binary and leave the matrix incomplete. Individual reruns recover only the requested six observations; they do not turn the original matrix into a complete package run.
7. The statement-drop menu can create uncompilable standalone Go patches when it leaves a range variable unused. Two diffs needed whole-loop deletion, preserving the original effect. The repairs were not designed from which tests failed.
8. A surviving deletion can be masked by a later pass. M15 had no changed final-answer witness, so labelling every survivor unguarded would overstate the evidence.
9. Repeated full oracle linking cost about 16 to 24 seconds per standalone validation. Production-package compilation was sufficient to check most replay patches and fit the budget. The interrupted validation and initial failed repair logs are preserved.
10. Warm tools do not supply current Node dependencies. npm ci was required and completed successfully. The loader's only configured node_modules path is stage3/api; no additional dependency directory was found.

No sacred, slow-worthy, untrue or subsumed verdict is asserted. No main push or pull request. Evidence only is committed on the requested audit branch.
