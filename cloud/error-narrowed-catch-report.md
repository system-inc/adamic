Added the 047cb0d narrowing probe as a permanent Node oracle fixture; classified fixed-before.
Fix attribution: be5a509; regression coverage commit: 1d739f0, followed by this audit commit.
Commands and outputs: focused and full affected uncached gates passed; counts, Cohere, go vet and formatting passed.
Mutant: replacing error_defined's TypeError throw with panic built successfully and failed the Node exit comparison (0 versus 70).
Not covered: the complete repository gate, performance benchmarks or the other worker's number-field extension.

The fixture `internal/oracle/testdata/047cb0d_narrowed_in_try.a` keeps the supplied program: narrowing and drop happen outside try, while the invalidated property read is inside it. All ordinary oracle backends print `caught\nafter\n`, exit 0. This includes source Node, generated JavaScript, sanitized and release native, and leak checking. The built label also exercises releasing the original object and its string when drop clears chain.

Classification: **fixed-before**. `git show be5a509 -- internal/lower/narrowed.go` identifies the conversion from the old invariant check to errorDefined for TypeError property reads. Execution on this branch's unchanged compiler confirms this particular probe. Historical commit attribution comes from the source diff; this unit does not rebuild that historical commit. narrowed.go and all compiler implementation files remain unchanged.

TestIntegrationCatchabilityMutants now runs the throw-to-panic mutant on this exact fixture too. The mutant successfully compiles native code, then exits 70 where Node exits 0, proving the fixture distinguishes catchability from merely detecting invalidated narrowing. Standard fixture registration also gives it permanent count and backend checks.

Cohere checks a byte-identical .ts mirror with the repository's compiler options and prelude. Three explicit rule suppressions preserve the integration probe's interface name, built template label, and unused named catch binding; no runtime expression is simplified away. Scoped --no-fix result: exit 0, 276 rules, one checked, 100% Adamic-ready.

Commands (output goes directly to logs):

```sh
bash cloud/setup.sh > /tmp/adamic-link-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/047cb0d_|TestIntegrationCatchabilityMutants/top-level' -count=1 -v > /tmp/adamic-link-probe.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-link-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql > /tmp/adamic-link-gate.log 2>&1
go vet ./... > /tmp/adamic-link-vet.log 2>&1
gofmt -l cmd internal > /tmp/adamic-link-gofmt.log
git diff --check > /tmp/adamic-link-diffcheck.log 2>&1
/tmp/adamic-catchability-cohere --directory /tmp/adamic-link-cohere --no-fix 047cb0d_narrowed_in_try.ts > /tmp/adamic-link-cohere.log 2>&1
```

Count regeneration passed (89.901s). All 273 prior rows are byte-for-byte unchanged; the only addition is this fixture: allocations 4, frees 4, retains 6, releases 9, peak live 3, in regions 0.

Setup timing lines, with nproc = 5:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (119s)
setup: done in 119s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Final affected gate: exit 0. This includes the existing exception mutants and new fixture mutant. ASan, UBSan and leak checks stay enabled.

```text
?   	github.com/system-inc/adamic/internal/ir	[no test files]
ok  	github.com/system-inc/adamic/internal/lower	75.319s
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/flow	207.431s
ok  	github.com/system-inc/adamic/internal/fresh	105.745s
ok  	github.com/system-inc/adamic/internal/native	317.847s
ok  	github.com/system-inc/adamic/internal/oracle	303.090s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	133.771s
exit 0
```

Coverage commit: `1d739f0` (`Hold top-level invalidated narrowing catchability against Node`). The following audit commit records these results. Both are pushed on codex/error-classes-counts; no pull request.
