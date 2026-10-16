# Required packed class fields

Rebuilt `801452d6` in source order. Required boolean-or-undefined fields use MaybeBoolean slots with undefined initial values. Inheritance, readonly constructor assignment, private slots and parameter properties agree with Node. Selector gap 3 and values gap 2 are marked closed; their port workarounds remain.

Commands, all after sourcing `/workspace/adamic-tools/env.sh`, with output in class-*.log:

- `timeout 180 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/scout_(class_boolean|selector_optional|values_optional)' -count=1 -v -timeout 90s`: PASS, 1.310s. New fixture leaves 1.21s, 1.25s and 1.25s, Node and both backends, release builds and sanitizers.
- `timeout 180 go test ./stage1/cohere/selector ./stage1/cohere/values -run 'TestEachGapStandsWhereGapsMdSaysItDoes/gaps/(3_optional_boolean|2_boolean_or_undefined_field)' -count=1 -v -timeout 90s`: PASS, changed leaves 4.12s and 4.07s.
- `timeout 150 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s`: PASS, 71.357s.
- `timeout 150 go test ./internal/lower -run '^TestScoutClassBooleanKeepsUnsafeViewsRefused$' -count=1 -v -timeout 90s`: PASS, new top-level leaf 1.05s.
- Full `internal/lower` with `-count=1 -timeout 90s` hit its aggregate timeout under concurrent counts/build workload; no individual failure was reported before the timeout. `timeout 240 go test ./internal/lower -count=1 -timeout 180s`: PASS, 166.998s. The package deadline is aggregate; new leaves are below 60s.
- `timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 570s -args -update-counts`: PASS, 308.479s. Three new rows, no existing count changes.

Mutants ran with `timeout 150 go test -overlay review/compiler/scout-unions-main/<name>.overlay.json <package> -run <pattern> -count=1 -v -timeout 90s`. `class-zero` (oracle class fixture) changes undefined initialization to present false; both backend outputs disagree with Node, exit 1. `class-optional` (lower refusal test) removes the postfix presence guard; the optional neighbor unexpectedly lowers, exit 1. Complete mutant sources are .go.txt.

Conservative scope: optional class presence, boxed mixed-union class layouts and mutable unsafe inheritance remain refused. No parser-corpus workaround removal or corpus byte credit is claimed. This advances step 17's required class union storage only.

Integration lane checks passed: `lane checks 3.0 s: gofmt and tools on 12 Go files, t.Parallel on 4 test packages; vet 4 packages`.
