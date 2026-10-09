Built: self-comparison representation helpers rebuilt on main for step 04, task #7c6b4pq.
Commits: c0b206949dfc7c7122149f0e1f734c5e0e9c5b34 rebuilds 10ad5d57; the delivery commit adds this evidence.
Commands and outputs: focused fixture and mutant tests pass; counts PASS (54.184s), regression PASS (52.628s), lane checks PASS.
Mutant: folding emitted numeric self-equality to true compiles and runs; sweep.a catches NaN stdout differing from Node.
Not covered: optional WASI execution and the full gate; no coercion semantics were added.

Number, boolean and reference comparisons use helpers with distinct parameters, avoiding clang's self-comparison warning. IEEE numeric equality preserves NaN and signed zero. Existing string, optional and union helpers keep their semantics. No c2-only dependency was needed. The three fixture tests are separate top-level parallel tests.

Commands source /workspace/adamic-tools/env.sh and use GOPROXY=https://proxy.golang.org|direct. Tests write directly to the named logs. Focused commands have an outer timeout of 150 seconds and go test -timeout 90s. Count regeneration uses timeout 210s and go test -timeout 180s after the cold aggregate run hit 90 seconds.

- go test ./internal/oracle -run '^TestSelfCompare' -count=1 -v -timeout 90s: PASS, 9.324s. Leaves: ArrayFill 0.57s, ConstantMutant 6.13s, Closure 9.31s, Sweep 9.31s.
- go test ./internal/oracle -run '^TestSelfCompare|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closure_convention_.*|closures.a|method_closures.a|unions.a|numbers.a)$' -count=1 -v -timeout 90s: PASS (52.628s), regression.log.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts: PASS (54.184s), counts.log.

The first counts and related oracle attempts found missing @types/node 25.3.3, required by host fixtures. They failed before count regeneration. timeout 180s npm ci --prefix stage3/api installed the pinned dependencies; those logs are retained with the successful reruns. Counts are regenerated once successfully, adding only the three self-comparison rows.

Setup: node 0.026s; go 0.028s; markdown dependency step 0.007s, ready 0.077s; submodules 0.091s; clang 0.206s; go build 38.669s; test binaries deferred 38.803s; cache warm 38.804s; total 38.830s. nproc=5, cpu.max=400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8. setup.log records the complete output.

The conservative scope is the requested commit and its witnesses, rebuilt directly on current main. The historical fixture report was omitted because evidence belongs under review/compiler/self-compare-main/.

Final main-base run leaves: ConstantMutant 0.88s, Sweep 1.54s, Closure 1.65s, ArrayFill 1.87s. The aggregate cold counts attempt reached its hard limit without writing counts; counts-timeout.log records it. The successful bounded retry completed after the checker/setup work stopped competing for CPUs.

Lane command: timeout 60s git fetch -q origin main devtools/fast-gate cloud/merge-tree, then the fetched cloud/integration/lane-checks.py runs under timeout 150s. Output: lane checks 1.6 s: gofmt and tools on 3 Go files, t.Parallel on 1 test packages; vet 1 packages. git diff --check passes.
