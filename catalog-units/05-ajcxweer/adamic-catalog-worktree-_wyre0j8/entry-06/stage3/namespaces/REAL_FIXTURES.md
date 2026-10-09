# Real TypeScript namespace fixture matrix

Merged fixture commit a75ba0023a70e01829f1b17850240f60446c1fb9 with merge commit 82024fe, whose parents are ce8a2ac and a75ba00. This imports the fixture branch and its newer main ancestry. Four conflicts were reconciled: class override checks retain parameter-property and enum checks alongside newer accessor/static/nominal checks; Lower retains namespace initialization alongside accessor-name discovery; refusals retain enum/namespace and definite-assignment checks; counts retain both sets of fixtures. The first merged oracle caught a non-conflicting initialization interaction: newer main initialized an optional declaration to undefined even for a namespace var whose storage was already hoisted, resetting a previously assigned value on redeclaration. Commit 2ba6a10 excludes NamespaceVar from that initialization path, preserving ordinary optional-declaration behavior from main. The final matrix row uses that corrected commit. No additional namespace form was admitted during this follow-up.

All twelve unchanged .a fixtures were built against the base, each of the six ordered commits, and the merge. They cover all eleven real compiler namespace declarations: ten runtime declarations and the type-only Status declaration. They retain selected original executable bodies with reduced external contracts and drivers, as described in the fixture author's README. No reopening exists in these fixtures. The three merges are BuilderState/interface, Debug.log/function, and BinaryExpressionState/callable type alias.

## Count moved by each step

| Point | Commit | Compiles / 12 | Delta | Newly compiling fixture |
| --- | --- | ---: | ---: | --- |
| base | adc45ca | 1 | baseline | 12_builder_release_cache.a |
| 1 state | aff39ff | 2 | +1 | 07_parser_singleton.a |
| 2 enums | b0e9b15 | 3 | +1 | 11_status_type_only.a |
| 3 bodies | f5161d8 | 3 | +0 | None |
| 4 exports | 083e7a9 | 4 | +1 | 04_debug_state.a |
| 5 escape refusal | 2f40563 | 4 | +0 | None |
| 6 merge refusals | ce8a2ac | 4 | +0 | None |
| fixture branch merge + hoisting fix | 2ba6a10 | 4 | +0 | None |

## All twelve on the merged branch

| Fixture | Outcome |
| --- | --- |
| 01_builder_state.a | NotYet |
| 02_jsx_names.a | NotYet |
| 03_react_names.a | NotYet |
| 04_debug_state.a | Compiles; native and sanitized native match Node |
| 05_debug_log_merge.a | NotYet |
| 06_binary_expression_state.a | Refused |
| 07_parser_singleton.a | Compiles; native and sanitized native match Node |
| 08_parser_jsdoc_nested.a | NotYet |
| 09_incremental_parser.a | Refused |
| 10_tracing_escape.a | NotYet |
| 11_status_type_only.a | Compiles; native and sanitized native match Node |
| 12_builder_release_cache.a | Compiles; native and sanitized native match Node |

All Node observations were regenerated from the original source using oracle/node.mjs and Node's independent transform, and matched the fixture author's status.json byte for byte. Every successful historical native run matches stdout, stderr and exit status. The final four also match under ASan/UBSan with detect_leaks=1 and halt_on_error=1. No silent miscompile was observed. Exact outputs, diagnostics, commit IDs and outcomes are in [fixture-matrix.json](fixture-matrix.json). The historical fixture author's status.json and validation.json are preserved.

Four successful slices do not mean four complete compiler namespaces compile. In particular, Parser's original countNode runs natively, but unchanged parser.ts still does not. The earlier five accepted normalized declaration shapes are a different experiment: the original-source fixtures retain branded strings, generic signatures and original bodies that expose further blockers.

## Reproduction and checks

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/namespaces/fixture-matrix.py --scratch /tmp/namespaces-real-matrix --output stage3/namespaces/fixture-matrix.json > /tmp/namespaces-real-matrix.log 2>&1
python3 stage3/namespaces/fixture-matrix.py --scratch /tmp/namespaces-real-matrix --output stage3/namespaces/fixture-matrix.json --mutant > /tmp/namespaces-real-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/namespaces-merge-counts.log 2>&1
go test -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/oracle > /tmp/namespaces-merge-packages.log 2>&1
go vet ./... > /tmp/namespaces-merge-vet.log 2>&1
gofmt -l cmd internal > /tmp/namespaces-merge-format.log
```

The matrix extracts each historical compiler from git into scratch and references the existing cohere submodule through a symlink and Go workspace, without copying its code. Build errors are classified only by known checker/NotYet/Refused diagnostics; unexpected errors stop the experiment. Every process observation is saved separately in scratch. The merged compiler is rebuilt for its row. The matrix bypasses oracle caches.

The mutant changes the original Parser countNode increment from nodeCount++ to nodeCount += 2 in a scratch .a source file. It builds and runs cleanly under sanitizers, but prints 6 2 instead of Node's original 3 1. The same matrix equality check raises Node disagreement and exits 1. This proves the check rejects a real semantic mutation, not a compiler error or sanitizer failure. The original fixture and compiler remain unchanged.

Initial merge package results (the namespace-var failure was then corrected):

```text
ok  	github.com/system-inc/adamic/internal/load	2.505s
ok  	github.com/system-inc/adamic/internal/lower	84.090s
ok  	github.com/system-inc/adamic/internal/native	369.614s
--- FAIL: TestNativeAgreesWithNode (0.02s)
    --- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/namespaces_parser_state.a (1.44s)
        oracle_test.go:590: gate cache node hit
        oracle_test.go:590: gate cache node miss
        oracle_test.go:479: gate cache native miss
        oracle_test.go:472: gate cache native miss
        oracle_test.go:611: stdout differs
            node:   exit 0, stdout "7\n-9:nodenode:3\n-9:nodenode:-1\n", stderr ""
            native: exit 0, stdout "-1\n-9:nodenode:3\n-9:nodenode:-1\n", stderr ""
        oracle_test.go:617: JavaScript backend: stdout differs
            node:    exit 0, stdout "7\n-9:nodenode:3\n-9:nodenode:-1\n", stderr ""
            backend: exit 0, stdout "-1\n-9:nodenode:3\n-9:nodenode:-1\n", stderr ""
        oracle_test.go:486: gate cache native hit
FAIL
gate cache: native hits=284 misses=970
gate cache: node hits=62 misses=638
gate cache: probe hits=0 misses=24
FAIL	github.com/system-inc/adamic/internal/oracle	333.643s
FAIL
```

After correction, the focused lower tests passed in 3.321s and the uncached namespace oracle passed in 7.121s, with zero native or Node cache hits. The full oracle package rerun produced:

```text
ok  	github.com/system-inc/adamic/internal/oracle	26.393s
```

Counts regeneration passed in 50.276s. Vet, formatting and git diff --check are clean. The full repository gate was not run; this follow-up ran the four relevant packages plus the independent 96-build fixture matrix.
