Built: ported removed-index reads and callback admission checks onto compiler/after-chain for task #b75jjs3.
Commits: source 0167f5b4213c093fe568e2854598d36e81b33800; base 2391c655a4287fee747af74d31f5781e99be19e5.
Commands and outputs: targeted oracle 32.248s PASS, fourteen-row counts 6.736s PASS, fuzz 0.008s PASS; lane checks recorded separately.
Mutants: skip-check and old-stop caught in both backends by exit-code assertions; old fuzz matcher caught by four classification assertions.
Not covered: whole packages, full gate, other array-method policies, or expanded callback representation support.

The chain still had the blanket search shrink panic. The port preserves its callback adapter and discarded forEach result handling. Search admission is computed from the original checker signature before adaptation; the adapter receives the search argument representation to avoid converting it twice. Omitted and defaulted value parameters retain the source commit's admission policy. Four supplied fixtures and ten search_shrink fixtures run on source Node, release native, ASan/UBSan native and backend JavaScript. Successful sanitized runs also pass LeakSanitizer. Checked calls exit 70 naming method, index and element type.

The skip-check native mutant exits 0 and prints 0 1, 1 2, 2 0, -1 on separate lines; scalar native storage has no undefined tag after the check is removed. JavaScript exits 0 and prints 0 1, 1 2, 2 hole, 2, proving undefined entered the number callback. Both are caught by the required exit 70 assertion, rather than warnings or sanitizer failures. Both old-stop mutants exit 70 instead of Node's 0. The old diagnostic matcher mutant exits test status 1 with four new classifications failing.

Exact successful commands, each redirected to the matching log in this directory:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 timeout 300 go test ./internal/oracle -run '^TestSearchShrink|^TestCallbackWidening(NumberFind|NumberFindLast|NumberFindIndex|ObjectFind|NumberForEach|ObjectForEach)$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(visits|map_shrinks|find_shrinks|find_index_shrinks|library_array_find_last|maybe_booleans)\.a$' -count=1 -v -timeout 90s
timeout 180 go test -overlay=review/compiler/search-shrink-chain/counts-overlay.json ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout 90s -args -update-counts
timeout 120 go test ./internal/fuzz -run '^TestJudgeReadsThePanicLine$' -count=1 -v -timeout 90s
timeout 120 go test -overlay=review/compiler/search-shrink-chain/fuzz-overlay.json ./internal/fuzz -run '^TestJudgeReadsThePanicLine$' -count=1 -v -timeout 90s
```

The counts overlay keeps the production counted helper, measures exactly the fourteen requested rows, and replaces only those rows in the union table. All fourteen equal the source commit's measurements; every pre-existing chain row is unchanged. The overlay sources are .go.txt evidence, not compiled repository packages. The source report is preserved as original-report.md and describes the old branch, not this port.

Cold compilation caused the first setup command (timeout 240) and two first test commands (timeout 150) to exit 124 before results. The bounded retries passed. Setup uses GOPROXY=https://proxy.golang.org|direct. nproc is 5, cgroup cpu.max is 400000 100000. The retry timing lines are:

```text
setup: go ready (0.017s)
setup: node ready (0.018s)
setup: submodules ready (0.062s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.068s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.216s)
setup: go build ready (36.455s)
setup: test binaries deferred (use --warm-tests) (36.553s)
setup: build cache warm (36.554s)
setup: build-flags commit=2391c655a4287fee747af74d31f5781e99be19e5 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=8.58 6.22 2.68 2/169 3931 load-after=10.66 7.12 3.14 2/171 5544
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (36.581s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.iiuAZe
```

Every ported fixture and mutant is now a top-level test. Observed leaf durations, including the oracle's tool setup:

```text
--- PASS: TestSearchShrinkDefault (22.65s)
--- PASS: TestSearchShrinkFindLastIndexMaybe (22.71s)
--- PASS: TestSearchShrinkMutantSkipCheck (1.19s)
--- PASS: TestSearchShrinkMutantOldStop (1.31s)
--- PASS: TestSearchShrinkFindIndexChecked (1.09s)
--- PASS: TestSearchShrinkFindLastIndexChecked (1.37s)
--- PASS: TestSearchShrinkFindLastChecked (1.28s)
--- PASS: TestSearchShrinkFound (1.24s)
--- PASS: TestSearchShrinkFindChecked (1.39s)
--- PASS: TestSearchShrinkBoolean (1.79s)
--- PASS: TestSearchShrinkFindLastMaybe (2.23s)
--- PASS: TestSearchShrinkFindMaybe (2.04s)
--- PASS: TestSearchShrinkOmitted (2.50s)
--- PASS: TestSearchShrinkUnknown (2.87s)
--- PASS: TestSearchShrinkNumberElementMaybe (1.98s)
--- PASS: TestSearchShrinkReferenceElementMaybe (3.00s)
```

The six selected existing array oracle fixtures and six chain callback widening leaves passed. This advances the search-shrink ruling on the lowering chain; no numbered roadmap step was supplied in this unit.
