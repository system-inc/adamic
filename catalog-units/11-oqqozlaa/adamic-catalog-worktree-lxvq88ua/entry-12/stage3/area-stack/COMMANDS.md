# Current group commands

```
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/adamic-scratch GOMAXPROCS=2 GOFLAGS='-p=1 -trimpath'
go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(function_values_|new_expression_|notyet_library_|map_union|host_map_union)|TestNativeAgreesWithNode/stage3/fixtures/taste/13_void_callback|TestFunctionValueBoundaryBoxing|TestNewExpression|TestNotYetLibrary.*Mutants' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
bash verify/catalog/check.sh <recorded-group-sha> --jobs 2
```

Every test and catalog command writes directly to its own log file. Review worktrees add -buildvcs=false to GOFLAGS for their referenced SDK submodule.
