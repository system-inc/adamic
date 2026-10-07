Added opt_3 as a permanent oracle fixture; the optional Error message is fixed-before on this branch.
Fix attribution: be5a509; coverage commit: cb196e5, followed by this audit commit.
Commands and outputs: focused and full affected uncached gates pass; counts, Cohere, vet and formatting pass.
Mutant: removing Error_initialize's empty-message fallback compiles and is caught by UBSan null-string access (Node exit 0, mutant exit 1).
Not covered: full repository gate, performance benchmarks or coercion of non-string message types.

Reproduction came first, on the unchanged compiler at b61f9b7. The supplied `failed(message?: string)` program passes source Node, generated JavaScript, ASan/UBSan native, release native and leak checking, printing `[] [m]\n` with exit 0. The reported integration failure does not reproduce on this branch.

Classification: **fixed-before**, in be5a509. The constructor path is now newError -> errorMessage -> Error_new -> Error_initialize. Optional strings retain their nullable string representation at the call boundary; Error_initialize coalesces its message parameter with the empty string before storing the message field. `git show be5a509 -- internal/lower/error_classes.go` shows that default already in the original nominal-error implementation. No new orDefault is needed at the call site, and no compiler implementation file changes in this unit. This attribution is from the source diff and present-branch execution, not a rebuilt execution at every historical commit.

The exact probe is `internal/oracle/testdata/4ddd17f_opt_3.a`, registered in catchability_test.go for normal backend, sanitizer, release and leak checks. Its permanent TestIntegrationCatchabilityMutants entry removes the Coalesce in Error_initialize, passing the undefined/null string directly into the message field. Native compilation succeeds. UBSan reports `member access within null pointer of type 'adamic_string'` in string_build_impl.h:39, and the mutant exits 1 instead of Node's 0. The test requires the sanitizer report as well as the oracle disagreement, so a build failure or unrelated diagnostic cannot count as this mutant's kill.

Counts regenerated successfully (17.024s): all 274 existing fixture rows remain byte-for-byte unchanged. The new row has allocations 3, frees 3, retains 10, releases 7, peak live 1, in regions 0. Scoped Cohere --no-fix on a byte-identical .ts mirror passes: 276 rules, one checked, 100% Adamic-ready. No lint suppression is needed.

Commands, with output written directly to logs:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/adamic-opt-setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/4ddd17f_ -count=1 -v > /tmp/adamic-opt-before.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/4ddd17f_|TestIntegrationCatchabilityMutants/undefined_Error' -count=1 -v > /tmp/adamic-opt-unit.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestIntegrationCatchabilityMutants/undefined_Error' -count=1 -v > /tmp/adamic-opt-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-opt-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql > /tmp/adamic-opt-gate.log 2>&1
go vet ./... > /tmp/adamic-opt-vet.log 2>&1
gofmt -l cmd internal > /tmp/adamic-opt-gofmt.log
git diff --check > /tmp/adamic-opt-diffcheck.log 2>&1
/tmp/adamic-catchability-cohere --directory /tmp/adamic-link-cohere --no-fix 4ddd17f_opt_3.ts > /tmp/adamic-opt-cohere.log 2>&1
```

The initial comparison preceded the mutant addition. The final focused mutant run adds the specific UBSan assertion. All compiler code stays unchanged throughout reproduction and verification.

Setup timing lines, nproc = 5:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (67s)
setup: done in 67s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Final affected gate, exit 0:

```text
?   	github.com/system-inc/adamic/internal/ir	[no test files]
ok  	github.com/system-inc/adamic/internal/lower	59.419s
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/flow	204.879s
ok  	github.com/system-inc/adamic/internal/fresh	86.559s
ok  	github.com/system-inc/adamic/internal/native	317.057s
ok  	github.com/system-inc/adamic/internal/oracle	294.692s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	135.162s
exit 0
```

The gate includes the new permanent mutant and the existing exception mutants. ASan, UBSan and leak checking remain enabled. go vet, gofmt and git diff --check finish with no diagnostics.

Coverage commit: `cb196e5` (`Hold optional Error message defaults against Node and UBSan`). The following audit commit records these observations. Both are pushed on codex/error-classes-counts; no pull request.
