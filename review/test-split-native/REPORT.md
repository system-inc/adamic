Merged main into the native test split branch for step 79 and kept the remaining workloads independently selectable.
Commits: fdf48ff2 resolves the merge; 0a5be9f3 adds the retained-unit coverage audit; final delivery tip is reported separately.
Verification uses exact Go selectors, four CPU equivalents, fresh Adamic caches, Node comparisons and the required lane checks.
The wrong-regexp-unit and wrong-width mutants fail the coverage audits; existing record mutants must reach their intended Node or sanitizer check.
No whole package run, full gate or new fixture counts are included; main's unchanged decoder and normalization units were not remeasured.

The conflicts in decode_ascii_test.go and normalize_test.go use main's complete versions. Main already exposes 90 decoder units and 18 normalization units as parallel top-level tests. The topic's environment-selected decoder and normalization shards were dropped. split_units_test.go retains main's build-cache helper, wrapper audit, timing helper and parallel declarations, with the topic's remaining coverage checks added separately. regexp_test.go retains main's parallel declarations and the topic's 250-case partitions, now exposed through 40 top-level functions. The topic's record, checker-cache and WASI partitions likewise use top-level functions rather than an environment variable selecting subtests. No other merge conflict occurred.

The branch remains needed: main does not contain the topic's regexp, record-mutant, checker-cache or WASI partitions. The preserved workloads contain 10,000 random regexp cases, six record mutants, both checker-cache modes, 35 WASI fixtures and the WASI request probe. TestRetainedTopLevelCoverage checks the actual wrapper names, helpers and unit arguments. Its mutant changes regexp unit 01 to run unit 00 and fails with "wrong unit in TestRegExpBytecodeRandomNodeUnit01". The mutant source is wrong-regexp-unit.go.txt and its output is coverage-mutant.log. A second mutant changes the partition width from 250 to 251; TestRetainedSplitCoverage rejects it with "random regex units must partition all 10000 cases into 40 units of 250". Its source and output are wrong-regexp-width.go.txt and width-mutant.log. The Node corpus must also contain exactly 10000 cases before selecting a partition.

The temporary internal/native/testbuildcache package and products_test.go were removed in favor of main's internal/buildcache helper. No test needs the topic's compiler-dependencies.json declaration after that removal, so that file was removed. The former internal/native/test-split directory, including budget.py, is archived under original-evidence/. The topic's historical skip census was also moved there. Historical measurements are retained as evidence, not used as the current budget proof. Relative to main, all remaining changes are test files or review evidence, making this a test-only candidate.

Cold measurement keeps Go dependency builds and the checker archive as prepared tool inputs. Each invocation receives empty XDG_CACHE_HOME and ADAMIC_BUILD_CACHE_DIR, ADAMIC_GATE_UNCACHED=1, GOMAXPROCS=4, -parallel=4, -count=1 and an exact -run selector. Both test elapsed time and complete invocation wall time are recorded. The three record-read leaves are selected individually because their build helper changed. Initial opt-in skips are retained in the ledger but replaced by enabled measurements in the final table. See cold/results.jsonl for exact commands and cold/*.jsonl for test output.

Tool preparation:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/native-setup.log 2>&1
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 go build -buildmode=c-archive -o /tmp/native-split-tsgo.a ./bridge/tsgo/archive > /tmp/native-archive.log 2>&1
```

Setup reported node 0.026s, Go 0.026s, submodules 0.071s, markdown dependencies 0.079s, clang 0.173s, WASI SDK 6.337s, Go build 55.181s, build cache 55.375s and done 55.410s. nproc is 5; cpu.max is 400000 100000, or four CPU equivalents. Preparation succeeded. An initial lane invocation lacked the installed Go PATH and failed to find gofmt; sourcing env.sh resolved it. The subsequent lane checks passed.

Measurement commands:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/native-split-tsgo.a
python3 review/test-split-native/measure.py > /tmp/native-measure.log 2>&1
```

measure.py enables the WASI SDK clang only for WASI selectors. The measurements performed during this merge used the same selectors in a regular phase and an enabled opt-in phase, with the exact commands preserved in the ledger. No timing result relies on cached test observations.

The final audit passes 91 exact selectors covering 95 test leaves. Every invocation, including its cold Adamic setup, is below 60 seconds. Maximum invocation is 40.31s and maximum root test elapsed is 38.46s. TIMINGS.md lists every leaf's seconds. The final verification reran regexp unit 00 after retaining the corpus-size assertion and reran both coverage audits.

Required lane command:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Final lane output is recorded in lane-final.log. No counts refresh was needed because no fixture was added. The destination branch remained at d80001f0 when checked before delivery.
