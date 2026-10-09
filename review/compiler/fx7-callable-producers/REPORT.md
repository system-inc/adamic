# Tasks aecx10a and 1fk58py: callable producer identities, items 133 and 145

Base: compiler/fx7-scalar-union-view at 32b8583facc362307049d182acc987dc42a02d35. Lowering now certifies only non-receiver closures from closureRecords; a named function is certified through the existing functionValue closure thunk, not its direct-call record. Native certifyUntaggedCallableRecorded independently excludes direct functions and receiver methods before generating code comparisons, matching the producer registry.

The original p21 function declaration, p60 named function, and p61 captured class-method call all compile and agree with Node through lowersAndAgreesWithNodeNative. The original integration review witnesses are active, with pending sidecars removed. Each prints true. An injected stale direct-function certificate is ignored by native emission and still agrees with Node.

Two independent mutants, restored in finally blocks:
- unfiltered-native.diff restores comparison against non-closure functions. TestCallableProducerRegistryFiltersDirectFunctions fails in clang: comparison of distinct pointer types adamic_code and double (*)(double), exactly the requested mutant. This is the explicitly requested compiler rejection proof, not a claimed runtime-check mutant.
- direct-producer-certificate.diff restores direct records in lowering. TestCallableProducerCertificatesUseClosureThunks detects a direct increment function in the certificate. This metadata assertion is necessary because the separate native defensive filter masks the stale certificate at runtime; the test first runs Node agreement.

Setup: GOPROXY=https://proxy.golang.org|direct, timeout 240 bash cloud/setup.sh, source /workspace/adamic-tools/env.sh. Go 0.023s, Node 0.024s, submodules 0.075s, clang 0.194s, build cache 41.415s, total 41.443s. nproc 5, four-CPU cgroup quota.

Initial new lower leaf seconds: P21 0.43, P60 0.48, P61 0.46, RegistryFiltersDirectFunctions 0.46. All top-level tests use t.Parallel; the certificate leaf's mutant timing is in direct-producer-certificate.log.

Commands, with output archived here:
- timeout 180 go test ./internal/lower -run '^TestCallableProducer' -count=1 -v -timeout 90s: PASS 0.496s.
- timeout 400 python3 review/compiler/fx7-callable-producers/run-mutants.py: PASS runner, both test exits 1.
- timeout 300 go test ./internal/lower -count=1 -timeout 240s: PASS 73.011s, lower.log.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: PASS 28.570s, readers.log.
- timeout 360 go test ./internal/oracle -run 'TestCountsAreRecorded|TestNativeAgreesWithNode/internal/lower/testdata/callable_producers|TestReviewProgramsAgreeWithNode/fxspptb_oct9_(views_p21|native_p60|native_p61)_' -count=1 -v -timeout 5m -args -update-counts: PASS 65.900s, oracle-counts.log; three counts rows added.
- Integration lane command after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -.

This first delivery fixes items 133 and 145. Item 135's assignable but non-identical producers are evaluated separately; no fix for their runtime stops is claimed in this receipt. No protected orchestration files were edited, no cohere source was copied, no full oracle package or repository gate was run, and no PR was opened.

Restored focused run passed in 0.424s; certificate leaf 0.29s, P61 0.36s, P60 0.37s, defensive filter 0.40s, P21 0.41s.

Lane checks PASS 1.7s: gofmt/tools on 23 Go files, parallel rule on 2 test packages, vet 2 packages.
