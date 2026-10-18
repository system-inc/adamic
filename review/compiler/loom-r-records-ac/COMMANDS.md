Validation commands

Each Go test uses GOMAXPROCS=4, -parallel 4 and -timeout 90s. Toolchain env.sh was sourced; ADAMIC_GOCACHE_OFF=1, GOCACHEPROG unset. Output is in the named log.

oracle-refusal-readers: exit 0, 6.674s
```sh
timeout 150s go test ./internal/oracle -run '^(TestRecordRefusal.*|TestRecordPrototypeMutant|TestRecordNarrowingMutant|TestNamedRecordTypeGuardMutant)$' -count=1 -timeout 90s -v -parallel 4
```

readers: exit 0, 2.777s
```sh
timeout 150s go test ./internal/ir -run '^(TestProduct_CallTargetExports|TestCallTargetReaders)$' -count=1 -timeout 90s -v -parallel 4
```

vet: exit 0, 0.415s
```sh
timeout 150s go vet ./internal/oracle ./internal/ir
```

old-line-pin-mutant: exit 1, 4.838s
```sh
timeout 150s go test -overlay /workspace/adamic/review/compiler/loom-r-records-ac/old-line-pins.json ./internal/oracle -run '^TestRecordRefusalPrototypeRead$' -count=1 -timeout 90s -v -parallel 4
```

fixture-hook: exit 0, 11.777s
```sh
timeout 150s go test ./stage3/fixtures -run '^TestProduct_FixtureOracleHook$' -count=1 -timeout 90s -v -parallel 4
```

archived-fixture-readers: exit 0, 7.501s
```sh
timeout 150s go test ./stage3/fixtures -run '^TestFixturesRecords$/^records$/^(04_own_values[.]a|08_option_boolean[.]a|09_option_undefined[.]a|10_option_wrong_type[.]a|11_option_string[.]a|12_nested_entries[.]a|14_group_by[.]a|15_inherited_read[.]a|17_compare_missing_object[.]a|18_compare_missing_scalar[.]a|19_compare_empty_objects[.]a)$' -count=1 -timeout 90s -v -parallel 4 -args -fixtures /workspace/adamic/review/compiler/loom-records-3/stage3-mutant
```

counts-table: exit 0, 2.228s
```sh
timeout 150s go test ./internal/oracle -run '^TestCountsAreRecordedTable$' -count=1 -timeout 90s -v -parallel 4
```

vet-fixture-reader: exit 0, 0.266s
```sh
timeout 90s go vet ./stage3/fixtures
```

Compiler and real a-check scans:
```sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_GOCACHE_OFF=1 timeout 600s bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
timeout 120s go build -o /workspace/scratch/records-ac-adamic ./cmd/adamic
timeout 600s python3 review/compiler/loom-r-records-ac/scan.py before
python3 review/compiler/loom-r-records-ac/headers.py
timeout 600s python3 review/compiler/loom-r-records-ac/scan.py after
timeout 420s python3 review/compiler/loom-r-records-ac/header-mutants.py
timeout 600s python3 review/compiler/loom-r-records-ac/scan.py final
python3 review/compiler/loom-r-records-ac/verify.py
```

The first scan exits 1 for exactly 35 missing headers. After and final exit 0. The two inner header mutant scans exit 1 as expected, with actual outcome refused; header-mutants.json gives their exact commands and times. old-line-pin-mutant exits 1 at diagnostic changed, as intended.

Archived stock Node input reader, each invocation bounded at 180 seconds by subprocess.run:
```sh
STOCK_TYPESCRIPT=/workspace/adamic/stage3/api/node_modules/typescript python3 review/compiler/loom-records-3/stage3-mutant/records/probe-inputs.py --observe
STOCK_TYPESCRIPT=/workspace/adamic/stage3/api/node_modules/typescript python3 review/compiler/loom-records-3/stage3-mutant/records/probe-inputs.py
```

Both exit 0. A structural comparison proves only main.ts diagnostic line numbers changed in five Node records; every behavior, stdout apart from line numbers, stderr and exit is preserved.

Final Git operations:
```sh
timeout 90s git fetch origin
git checkout -b compiler/loom-r-records-ac dbdf977f5b8b0979d414ebe6f20e7aadd312e309
git cherry-pick --no-commit 611cbc943a28e88567bcb5fb9a43031518a69d3d
timeout 90s git fetch --recurse-submodules=no origin devtools/fast-gate
git add -u
git add review/compiler/loom-r-records-ac
git diff --cached --check
timeout 60s git commit -F review/compiler/loom-r-records-ac/commit-message.txt
timeout 90s git push origin HEAD:refs/heads/compiler/loom-r-records-ac
timeout 30s git ls-remote --heads origin refs/heads/compiler/loom-r-records-ac
```
