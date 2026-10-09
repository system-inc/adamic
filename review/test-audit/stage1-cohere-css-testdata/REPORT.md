Unit u081 stopped on a red baseline; no test binary ran.
Starting origin/main: 12e77e8972a2e606cab6db05d84f428246a85339; nproc 5.
The requested directory mixes package postcss and package css.
No mutants or probes were planted; no test verdicts or medians are established.
Evidence branch: test-audit/stage1-cohere-css-testdata.

```json
[]
```

The array is empty because the required scope command could not list tests. This does not claim the directory has no Test declarations.

Code under test and oracle

The directory contains overlay adapters for Go cohere, not standalone Adamic-port checks. They call Go cohere Parse/ParseSCSS, parseWithParserName, formatWithParser and dump helpers to produce expected answers. Their callers in stage1/cohere/css compile these adapters separately into the appropriate cohere packages. No Adamic code-under-test entry can be established for the requested package because it does not compile. Mutating these adapters or Go cohere would mutate the oracle, forbidden by the brief.

Commands and failure

`go test -list . ./stage1/cohere/css/testdata/` exited 1 in 1.009 command seconds.
`timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/testdata/ -run . > baseline.log 2>&1` exited 1 in 0.031 command seconds.
Both report:

```text
found packages postcss (cohere_side_test.go) and css (compose_side_test.go) in /workspace/adamic/stage1/cohere/css/testdata
```

Static declarations, not measured test rows

- cohere_side_test.go:92: TestAdamicPortCases, package postcss.
- composition_corpus_side_test.go:89: TestAdamicPortCases, package postcss.
- compose_side_test.go:18: TestAdamicCompositionCases, package css.
- print_side_test.go:21: TestAdamicPrinterCases, package css.
- print_side_test.go:96: TestAdamicPrinterThroughput, package css.

Both postcss adapters also define adamicCohereTexts, so merely making package names uniform would not make this a valid package. The bodies reference private cohere functions supplied only by their overlay destinations.

Mutant table

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|

No mutants, probes or survivors exist in this audit. No sacred, subsumed, untrue or vacuity judgment is justified.

Brief feedback

1. The named unit is an adapter directory, not a buildable Go package. Explicit go test includes testdata even though recursive ./... normally omits it. This prevents the required scope enumeration and every baseline/matrix/timing command.
2. These files declare two different Go packages. Two files also define the same TestAdamicPortCases and helper. They are separate overlays, not one test family in one package.
3. The actual port checks live in the parent stage1/cohere/css package. Switching to that scope without authorization would change the unit; it is a candidate corrected brief, not a completed audit here.
4. The supplied directory's functions produce oracle answers. Treating them as production code under test would violate the prohibition against mutating Go cohere or its oracle harness.
5. The required JSON schema does not represent scope discovery failure. I use an empty array plus explicit unavailable scope, rather than invent rows, timings or oracles. Static declarations above preserve what was read.
6. The brief instructs both pushing evidence initially and permits skipping push later. I followed the explicit initial request and pushed failure evidence.
7. My first timing wrapper used /usr/bin/time, absent here. Those commands did not execute; I reran all three with Python monotonic timing and preserved the actual successful invocation records. This is a tooling mistake, not a baseline result.
8. npm ci in stage3/api was required even though compilation fails during package discovery before Node is reached. It succeeded. Setup was skipped because the warm env.sh gives go1.27.1.
9. The red-baseline stop rule is decisive here. Installing optional tools, measuring adapters in cohere, or running the parent package would not fix the requested package and would expand scope.

Timing and limits

Setup: 0 seconds. npm ci: 0.968 command seconds. Scope enumeration: 1.009 command seconds. Baseline build/discovery failure: 0.031 command seconds. Test binary execution: none. No three-run medians, native builds, standalone mutation compilation, matrices, probes or repo-wide replay were performed. No production source changed. No main push or pull request.
