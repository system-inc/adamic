Merged origin/main 0bf6186d into compiler/chain-slice-4; only counts.md conflicted, resolved by taking main and regenerating once.
Merge commit: c13ab59e; delivery SHA is reported in the handoff.
Counts passed in 47.676s; focused members 15.381s; stage3 fixtures 4.237s; generic-body fixtures 0.447s; project fixtures 0.503s; readers 11.627s; lane checks 7.3s.
No implementation or test edits; no new mutants requested or run beyond mutants exercised by the focused member tests.
Scope: requested counts reconciliation, focused fixtures, reader guard and lane checks only.

All test shells sourced /workspace/adamic-tools/env.sh with ADAMIC_GOCACHE_OFF=1. Focused oracle tests additionally used ADAMIC_GATE_UNCACHED=1. Test output is retained in sibling logs.

Commands:
```text
export GOPROXY='https://proxy.golang.org|direct'; ADAMIC_GOCACHE_OFF=1 timeout 240 bash cloud/setup.sh
Setup passed: Go 0.025s, Node 0.025s, markdown 0.075s, submodules 0.089s, clang 0.199s, shared cache off 0.200s, go build 10.475s, warm 10.621s, done 10.647s. nproc=5; cgroup cpu.max=400000 100000.
git merge --no-commit origin/main
git checkout origin/main -- internal/oracle/counts.md
timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 180s -args -update-counts
python3 review/compiler/chain-slice-4/main-refresh/attribute-counts.py
timeout 110 go test ./internal/oracle -run '^Test(ElementAccess|GenericBody|Step16|Generator|Step20|IterationDispatch|LoweringChainMixed)' -count=1 -timeout 90s -v
timeout 110 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/stage3/fixtures/(generics|iteration|iteration-dispatch)' -count=1 -timeout 90s -v
timeout 110 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/generic_body' -count=1 -timeout 90s -v
timeout 110 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/stage3/project-references-source' -count=1 -timeout 90s -v
timeout 110 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v
timeout 120 bash -o pipefail -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
lane checks 7.3 s: gofmt and tools on 71 Go files, t.Parallel on 6 test packages; a-check 13 .a files; vet 6 packages
```

The regenerated table has 1,069 rows. Against the prior slice tip 1fc069f0, the only differences are main's three new getters rows (A/F/R/L/P/G):

| Fixture | Counts |
| --- | --- |
| getters_census_class_map.a | 13/13/4/15/6/0 |
| getters_census_throwing_object.a | 3/3/5/8/3/0 |
| getters_census_binary_inline.a | 37/37/29/57/17/0 |

Against current main, the slice retains exactly its 55 previously attributed new rows and four previously attributed existing-row changes:

| Fixture | Main | Slice | Owner |
| --- | --- | --- | --- |
| internal/oracle/testdata/regexp.a | 437/437/382/417/58/0 | 437/437/383/418/58/0 | generators-main |
| internal/fresh/testdata/regexp_tree.ts | 91/91/81/85/35/0 | 91/91/85/89/35/0 | generators-main |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 1050/1050/677/994/97/0 | iteration-main |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5/0/5/3/5/0 | 11/3/7/3/8/0 | iteration-main |

The attribution script asserts exact old-slice values for every slice-owned delta, exact main values for getters, no deleted rows, and no unattributed changes. Full changed-row ownership and values are in counts-attribution.json. No unrelated count moved. Main was still 0bf6186d after the lane fetch.
