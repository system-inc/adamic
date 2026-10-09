# Callback widening through intrinsics

Task #1xxqbrk. Original main: 7c9ad7da7b8fe8553e184734fde6c0941952288b. Refusal commit: 9836491bd.

Built a call-site IR closure adapter that captures the original callback once and converts its supplied element, index, array and accumulator slots into the callback parameter representations. Numeric union members are boxed. A nullable reference becomes the null token when widened into a boxed union; ordinary optional references keep undefined. Both backends lower the same adapter, and the existing ownership passes see its cells, boxes, returns and throw edges. Effects returning a closure literal now preserve that proven call target.

Observations on main: eight numeric intrinsics printed no stdout and terminated by signal (the execution harness records exit -1). Their source on Node and their JavaScript output completed correctly. All eight object-element intrinsics and both generic for-of functions already agreed. The added generic-map probe independently uses main compiler files through a Go overlay and reproduces the numeric intrinsic failure; its object case agrees. The overlay main had only stage-1 test changes relative to the original main.

Inference: raw numeric element slots reached a callback whose declared parameter used boxed-union storage. The direct generic for-of route already fits its call arguments. The intrinsic route needs an adapter before calling the callback.

The first commit refuses the scalar-to-union mismatch at the callback, names its parameter and array element type, and asks for an element-type annotation. Its eight Node goldens and exact refusal pins pass. The second implementation retires those refusals for the represented shapes. Array-to-structural callback parameters requiring a different layout remain narrowly refused, naming number[] and asking to annotate x as number[]. Library map adapters keep their existing path, including String and parseInt; flatMap keeps its existing argument fitting.

| Witness | Node stdout | Native before stdout / exit | JavaScript before stdout / exit | Native and JavaScript after stdout / exit |
|---|---|---|---|---|
| number-every | `"true\n"` | `""` / -1 | `"true\n"` / 0 | `"true\n"` / 0 |
| number-filter | `"1,3\n"` | `""` / -1 | `"1,3\n"` / 0 | `"1,3\n"` / 0 |
| number-find | `"2\n"` | `""` / -1 | `"2\n"` / 0 | `"2\n"` / 0 |
| number-forEach | `"1\n2\n3\n"` | `""` / -1 | `"1\n2\n3\n"` / 0 | `"1\n2\n3\n"` / 0 |
| number-generic | `"1,2,3\n"` | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| number-generic-map | `"1,2,3\n"` | `""` / -1 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| number-map | `"1,2,3\n"` | `""` / -1 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| number-reduce | `"123\n"` | `""` / -1 | `"123\n"` / 0 | `"123\n"` / 0 |
| number-some | `"true\n"` | `""` / -1 | `"true\n"` / 0 | `"true\n"` / 0 |
| number-sort | `"1,2,3\n"` | `""` / -1 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| object-every | `"true\n"` | `"true\n"` / 0 | `"true\n"` / 0 | `"true\n"` / 0 |
| object-filter | `"1,3\n"` | `"1,3\n"` / 0 | `"1,3\n"` / 0 | `"1,3\n"` / 0 |
| object-find | `"2\n"` | `"2\n"` / 0 | `"2\n"` / 0 | `"2\n"` / 0 |
| object-forEach | `"1\n2\n3\n"` | `"1\n2\n3\n"` / 0 | `"1\n2\n3\n"` / 0 | `"1\n2\n3\n"` / 0 |
| object-generic | `"1,2,3\n"` | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| object-generic-map | `"1,2,3\n"` | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| object-map | `"1,2,3\n"` | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |
| object-reduce | `"123\n"` | `"123\n"` / 0 | `"123\n"` / 0 | `"123\n"` / 0 |
| object-some | `"true\n"` | `"true\n"` / 0 | `"true\n"` / 0 | `"true\n"` / 0 |
| object-sort | `"1,2,3\n"` | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 | `"1,2,3\n"` / 0 |

Additional passing controls: named sort, toSorted, findIndex, findLast, widened index and reduce accumulator parameters, callback factory evaluation count and captures, ordinary function arguments.length, thrown callback cleanup, library map callbacks, boolean/string/optional-number storage, and nullable/optional object storage. The nullable control caught a draft conversion that confused null with undefined; explicit normalization fixes it. The source on Node remains the golden.

Unrepresented control: internal/lower/testdata/callback_widening/array_structural.a prints "1,2|3,4\n" on Node. The compiler refuses its Slice parameter receiving number[], with the fix to annotate x as number[]. Its first-line a-check pin is asserted by TestCallbackWideningUnrepresentedParameter.

Validation: 35 executable top-level oracle tests each observe source Node, release native, ASAN/UBSAN native, JavaScript on Node, and LeakSanitizer. One additional top-level oracle test pins the unsupported representation refusal. Fresh and flow have one independent top-level test for each of the 35 executable programs, including SSA, mutation ranges, graph paths and liveness. IR target tests include an unknown-result control. No test is skipped. No whole package or full gate is run.

All runs use affinity [0,1,2,3] and GOMAXPROCS=4 on this one instance. nproc is 5; cgroup cpu.max is 400000 100000. Each new top-level test, including its setup, is under 60 seconds. Individual seconds are in results.json and the table below.

| Package | Test | Wall seconds |
|---|---|---:|
| oracle | TestCallbackWideningNumberEvery | 0.617 |
| oracle | TestCallbackWideningNumberToSorted | 0.628 |
| oracle | TestCallbackWideningNumberNamedSort | 0.663 |
| oracle | TestCallbackWideningNumberGenericMap | 0.668 |
| oracle | TestCallbackWideningNumberFindIndex | 1.142 |
| oracle | TestCallbackWideningNumberFindLast | 1.18 |
| oracle | TestCallbackWideningLibraryMap | 1.211 |
| oracle | TestCallbackWideningNumberThrow | 1.257 |
| oracle | TestCallbackWideningNumberCapture | 1.76 |
| oracle | TestCallbackWideningOptionalNumberMap | 1.76 |
| oracle | TestCallbackWideningNumberReduceAccumulator | 1.765 |
| oracle | TestCallbackWideningNumberIndex | 1.79 |
| oracle | TestCallbackWideningUnrepresentedParameter | 1.906 |
| oracle | TestCallbackWideningOptionalObjectMap | 2.325 |
| oracle | TestCallbackWideningNullableObjectMap | 2.337 |
| oracle | TestCallbackWideningObjectEvery | 2.359 |
| oracle | TestCallbackWideningObjectSort | 2.511 |
| oracle | TestCallbackWideningObjectReduce | 2.899 |
| oracle | TestCallbackWideningObjectSome | 2.919 |
| oracle | TestCallbackWideningObjectMap | 2.926 |
| oracle | TestCallbackWideningObjectGeneric | 3.055 |
| oracle | TestCallbackWideningObjectForEach | 3.391 |
| oracle | TestCallbackWideningObjectFind | 3.455 |
| oracle | TestCallbackWideningObjectFilter | 3.484 |
| oracle | TestCallbackWideningBooleanMap | 3.586 |
| oracle | TestCallbackWideningStringMap | 3.924 |
| oracle | TestCallbackWideningNumberMap | 4.026 |
| oracle | TestCallbackWideningNumberArguments | 4.039 |
| oracle | TestCallbackWideningNumberSort | 4.133 |
| oracle | TestCallbackWideningNumberSome | 4.437 |
| oracle | TestCallbackWideningNumberReduce | 4.58 |
| oracle | TestCallbackWideningObjectGenericMap | 4.593 |
| oracle | TestCallbackWideningNumberForEach | 4.962 |
| oracle | TestCallbackWideningNumberGeneric | 5.068 |
| oracle | TestCallbackWideningNumberFilter | 5.079 |
| oracle | TestCallbackWideningNumberFind | 5.079 |
| fresh | TestCallbackWideningObjectEvery | 0.093 |
| fresh | TestCallbackWideningNumberIndex | 0.096 |
| fresh | TestCallbackWideningLibraryMap | 0.102 |
| fresh | TestCallbackWideningNumberGenericMap | 0.107 |
| fresh | TestCallbackWideningObjectSome | 0.162 |
| fresh | TestCallbackWideningObjectMap | 0.169 |
| fresh | TestCallbackWideningObjectReduce | 0.17 |
| fresh | TestCallbackWideningObjectSort | 0.183 |
| fresh | TestCallbackWideningObjectFind | 0.214 |
| fresh | TestCallbackWideningObjectForEach | 0.219 |
| fresh | TestCallbackWideningObjectGeneric | 0.225 |
| fresh | TestCallbackWideningObjectFilter | 0.236 |
| fresh | TestCallbackWideningNumberFindIndex | 0.268 |
| fresh | TestCallbackWideningNumberForEach | 0.268 |
| fresh | TestCallbackWideningNumberFindLast | 0.276 |
| fresh | TestCallbackWideningNumberGeneric | 0.277 |
| fresh | TestCallbackWideningStringMap | 0.308 |
| fresh | TestCallbackWideningOptionalNumberMap | 0.309 |
| fresh | TestCallbackWideningNullableObjectMap | 0.309 |
| fresh | TestCallbackWideningOptionalObjectMap | 0.311 |
| fresh | TestCallbackWideningNumberArguments | 0.357 |
| fresh | TestCallbackWideningNumberFind | 0.358 |
| fresh | TestCallbackWideningNumberFilter | 0.361 |
| fresh | TestCallbackWideningBooleanMap | 0.365 |
| fresh | TestCallbackWideningNumberSome | 0.39 |
| fresh | TestCallbackWideningNumberToSorted | 0.392 |
| fresh | TestCallbackWideningNumberThrow | 0.4 |
| fresh | TestCallbackWideningObjectGenericMap | 0.408 |
| fresh | TestCallbackWideningNumberReduceAccumulator | 0.425 |
| fresh | TestCallbackWideningNumberReduce | 0.437 |
| fresh | TestCallbackWideningNumberEvery | 0.441 |
| fresh | TestCallbackWideningNumberSort | 0.455 |
| fresh | TestCallbackWideningNumberNamedSort | 0.464 |
| fresh | TestCallbackWideningNumberMap | 0.468 |
| fresh | TestCallbackWideningNumberCapture | 0.468 |
| flow | TestCallbackWideningLibraryMap | 0.243 |
| flow | TestCallbackWideningObjectEvery | 0.244 |
| flow | TestCallbackWideningNumberIndex | 0.249 |
| flow | TestCallbackWideningOptionalObjectMap | 0.274 |
| flow | TestCallbackWideningNullableObjectMap | 0.431 |
| flow | TestCallbackWideningOptionalNumberMap | 0.451 |
| flow | TestCallbackWideningBooleanMap | 0.456 |
| flow | TestCallbackWideningStringMap | 0.458 |
| flow | TestCallbackWideningNumberArguments | 0.629 |
| flow | TestCallbackWideningNumberGenericMap | 0.642 |
| flow | TestCallbackWideningObjectGenericMap | 0.667 |
| flow | TestCallbackWideningObjectSort | 0.677 |
| flow | TestCallbackWideningObjectSome | 0.812 |
| flow | TestCallbackWideningObjectReduce | 0.833 |
| flow | TestCallbackWideningObjectMap | 0.868 |
| flow | TestCallbackWideningObjectGeneric | 0.893 |
| flow | TestCallbackWideningObjectForEach | 1.01 |
| flow | TestCallbackWideningObjectFind | 1.045 |
| flow | TestCallbackWideningObjectFilter | 1.105 |
| flow | TestCallbackWideningNumberSome | 1.126 |
| flow | TestCallbackWideningNumberToSorted | 1.204 |
| flow | TestCallbackWideningNumberThrow | 1.277 |
| flow | TestCallbackWideningNumberSort | 1.326 |
| flow | TestCallbackWideningNumberReduceAccumulator | 1.34 |
| flow | TestCallbackWideningNumberReduce | 1.406 |
| flow | TestCallbackWideningNumberFindIndex | 1.474 |
| flow | TestCallbackWideningNumberGeneric | 1.553 |
| flow | TestCallbackWideningNumberForEach | 1.56 |
| flow | TestCallbackWideningNumberFindLast | 1.58 |
| flow | TestCallbackWideningNumberNamedSort | 1.677 |
| flow | TestCallbackWideningNumberMap | 1.76 |
| flow | TestCallbackWideningNumberFind | 1.806 |
| flow | TestCallbackWideningNumberFilter | 1.814 |
| flow | TestCallbackWideningNumberEvery | 1.852 |
| flow | TestCallbackWideningNumberCapture | 1.929 |
| ir | TestCallbackAdapterTargets | 0.0 |

Commands:

```text
source /workspace/adamic-tools/env.sh
GOFLAGS=-buildvcs=false GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1
go test -json -count=1 -timeout=5m ./internal/oracle -run ^TestCallbackWidening
go test -json -count=1 -timeout=5m ./internal/fresh -run ^TestCallbackWidening
go test -json -count=1 -timeout=5m ./internal/flow -run ^TestCallbackWidening
go test -json -count=1 -timeout=5m ./internal/ir -run "^(TestCallbackAdapterTargets|TestClosureTargetsBoundOnlyProvenValues)$"
python3 review/compiler/callback-widening/run-mutants.py
go test -json -count=1 -timeout=10m ./internal/oracle -run ^TestCountsAreRecorded$ -args -update-counts
```

All named commands passed. Five mutants: refusal revert is caught by all eight retired pins; adaptation omission is caught by all eight numeric intrinsics plus generic-map; omission of the null token is caught by the nullable object map; omission of the fit refusal is caught by the structural-array refusal pin; omission of known adapter-target handling is caught by TestCallbackAdapterTargets. No mutant verdict is a compiler build failure. The runner reconstructs the first commit with overlays; generated sources have .go.txt names outside the repository.

Counts refreshed: 35 new rows, zero existing rows changed, zero removed. Each executable fixture is registered in the oracle catalog. The refused fixture is not counted.

Tool setup completed: Go 0.023s; Node 0.024s; markdown step 0.010s/ready 0.077s; submodules 0.079s; clang 0.159s; go build 234.263s; test binaries deferred 234.488s; cache warm 234.489s; done 234.518s. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Environment corrections: one initial mutation invocation used the box default Go launcher and did not run tests; its evidence is replaced by the verified toolchain invocation. Disk filled from generated Go cache variants; older entries generated during this unit were evicted, retaining source files, module downloads and installed tools.

Not covered: new rest-parameter or overloaded-callback admissions, tuple-to-array conversion, or array-to-structural callable readers. The attempted rest probe hits the existing function-relation refusal before this adapter and was not admitted. No protected compiler orchestration file or native runtime file changed. Evidence contains no compilable Go source.

The concrete conservative choice is to keep unsupported representation conversions refused, with an element-type annotation as the fix. No open language ruling is required for the represented cases. Current-main merge and committed lane-check output follow.

The adaptation commit is `fc02927ba`; merge `c40c1af07` incorporates current main `c1aa2c9bf`. On that merged tip, the exact targeted package commands above passed again: 110 top-level results, zero failures and zero skips, including two additional main IR controls. Largest top-level wall duration: 5.261 seconds on four CPUs. Individual durations are in `merged-results.json`. All 36 changed `.a` files passed the gate’s a-check rules; zero failures.

The required lane command ran after the merge was committed. Its output:

```text
lane checks 4.6 s: gofmt and tools on 7 Go files, t.Parallel on 4 test packages; vet 4 packages
```

`go build ./...` also passed on the merged tip; output is in `merged-build.log`. Counts add 35 rows against the merged main with no existing row moved.
