Built six native test partitions and keyed shared products for roadmap step 38.  
Measured base 54cbc125 and final code on main 031a1259; delivery stays on compiler/test-split-native.  
Checks: 90 native top-level selectors, zero JavaScript tests, 192 corpus shards and 137 other executed leaves; maximum invocation 16.28 s.  
Mutants: eleven compiling Go overlays and two independent budget boundary failures were caught.  
Limits: bridge/tsgo belongs to its worker; three opt-in performance probes and other operating systems were not measured.

| Test | Before s | After largest invocation s | Units |
|---|---:|---:|---:|
| TestDecodeASCII | 185.20 | 6.79 | 90 |
| TestNormalizeMatchesNode | 40.18 | 7.53 | 18 |
| TestRecordMutants | 34.20 | 2.56 | 6 |
| TestSplitTSGoAgrees | 42.66 | 16.28 | 2 |
| TestRegExpBytecodeRandomNode | 22.77 | 3.24 | 40 |
| TestWASI | 27.73 | 5.08 | 36 |

Before seconds are the largest observed group or leaf duration. Normalization had a parallel points child that exceeded its reported parent time. The complete before/after inventory is in TABLE.md. The after column measures a fresh Go invocation selecting one unit with its shared products prepared by input hash.

The original top-level tests over 30 s were TestDecodeASCII 185.20 s, TestRecordMutants 34.20 s, TestSplitTSGoAgrees 42.66 s. Normalization also had a 40.18 s points child.

Coverage keeps 92,014 decoder prefixes in both native and WASI, with 1,040 old baseline/oracle comparisons per prefix; 10,000 seeded regex cases in both execution modes; all 1,114,112 normalization code points and the original context corpus; the same 35 WASI fixture names plus requests; both checker cache modes; and all six record mutants. TestSplitUnitCoverage freezes the range totals, fixture membership, backend and mode sets, mutant count and unique ownership. The measurement helper also compares each owning shard leaf to the full run and proves their union is complete.

All 329 executed leaves were selected independently. Their names equal the passing leaves in the complete sweep, with no duplicates or omissions. Go leaf timings and complete invocation timings are both below 30 s. 54 entries in the shared product store were prepared before the standalone runs; none was built after its preparation group finished. Programs owned by a single unit compile during that unit.

The planted valid regex program changes global case 500 after Node supplies its expected answer. All 40 shards ran through the same Go overlay. Only 2/40 failed, at TestRegExpBytecodeRandomNode/cases_000500_000750, with DISAGREEMENT case=0; the other 39 passed.

| Mutant | Catcher |
|---|---|
| missing-range | TestSplitUnitCoverage/decode_prefixes, TestSplitUnitCoverage/normalize_points, TestSplitUnitCoverage/random_regex |
| missing-wasi-fixture | TestSplitUnitCoverage/WASI |
| changed-wasi-membership | TestSplitUnitCoverage/WASI |
| missing-cache-mode | TestSplitUnitCoverage/split_cache |
| duplicate-record-mutant | TestSplitUnitCoverage/record_mutants |
| missing-decode-target | TestSplitUnitCoverage/decode_targets |
| duplicate-shard-owner | TestSplitUnitCoverage/shard_union |
| invalid-shard-accepted | TestShardSelection |
| artifact-input-ignored | TestArtifactInputsAndReuse |
| artifact-reuse-disabled | TestArtifactInputsAndReuse |
| elapsed-over-budget | [{"test": "planted", "action": "pass", "seconds": 30.001}] |
| invocation-over-budget | selector probe invocation took 30.001 seconds |

The input-ignored overlay uses input[:0], so it compiles and fails the input/reuse assertion. An earlier nil-input version failed Go compilation; that attempt was rejected as proof. Both timing boundary checks accept 30.000 s and reject 30.001 s. One checks leaf timing, and the other gives a short leaf an excessive complete-invocation duration.

Shared products use the test executable, source and header bytes, flags, tool versions, explicit compiler inputs, and overlay contents as keys. Native/WASI execution and comparison outcomes run again. Decoder Node goldens are a prepared input keyed by oracle source and Node version. Go dependency builds and the checker archive are prepared tool inputs. Runtime libraries retain their existing byte/flag/compiler keyed helper.

internal/buildcache at devtools commit 64ea92b646560aabf79c182b449f6a68561e562f was absent from origin/main 031a1259. The temporary testbuildcache package exposes its Inputs/Get/Tool API; changing the single aliased import in products_test.go selects the shared helper after it lands. No unlanded branch was merged.

Commands ran with GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/test-split-tsgo.a, ADAMIC_TEST_WASI=1 and ADAMIC_ORACLE_WASI=1. WASI selectors use the SDK clang on PATH. Each preparation group starts with an empty XDG_CACHE_HOME; unit executions fetch those prepared products. Go tests used exact -run selectors, -count=1, -json and -timeout 30m. Each command and exit is recorded in the result ledgers.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/test-split-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go build -buildmode=c-archive -o /tmp/test-split-tsgo.a ./bridge/tsgo/archive > /tmp/test-split-archive.log 2>&1
go vet ./internal/native ./internal/native/testbuildcache ./internal/javascript > /tmp/test-split-native-vet.log 2>&1
python3 internal/native/test-split/measure.py /workspace/test-split-native-final > /tmp/test-split-native-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /workspace/test-split-native-final --output /workspace/test-split-native-final/units.json > /tmp/test-split-native-budget.log 2>&1
python3 internal/native/test-split/measure.py /workspace/test-split-native-shards --shards-from /workspace/test-split-native-final > /tmp/test-split-native-shards-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /workspace/test-split-native-shards --output /workspace/test-split-native-shards/units.json > /tmp/test-split-native-shards-budget.log 2>&1
python3 internal/native/test-split/measure.py /workspace/test-split-native-other-units --units-from /workspace/test-split-native-final --exclude-corpus > /tmp/test-split-native-other-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /workspace/test-split-native-other-units --output /workspace/test-split-native-other-units/units.json > /tmp/test-split-native-other-budget.log 2>&1
python3 internal/native/test-split/budget_policy.py /workspace/test-split-native-budget-policy > /tmp/test-split-native-budget-policy.log 2>&1
python3 internal/native/test-split/mutants.py /workspace/test-split-native-mutants > /tmp/test-split-native-mutants-progress.log 2>&1
```

Setup timing lines: Node 0.023 s, Go 0.024 s, submodules 0.068 s, markdown 0.074 s, clang 0.172 s, build 42.408 s, total 42.575 s. The separate WASI setup took 13.921 s, including SDK 3.404 s. nproc printed 5; cpu.max is 400000 100000, allowing four CPUs. Tool paths, binary hashes, versions and source hashes are in evidence/metadata.json and evidence/inputs.json.

The budget audit rejects timings only with ADAMIC_UNIT_BUDGET=1 on the reference box. Without that flag, it logs slow timings while retaining correctness failures. No unconditional 30-second deadline was added inside any Go unit. budget_policy.py proves flag 1 rejects slow timings, flag 0 and an absent flag only log them, and all modes reject correctness failures. Test output went to logs. No whole-package test run or full gate was used. No production files or Adamic fixture files were changed, so counts.md did not need a new row. Main brought one additional test into the measured inventory; the split adds three helper tests. The ordinary gate skips TestMeasureClangUnits, TestNormalizeLongMeasurements and TestRecordBenchmark; they are opt-in measurements outside this unit, as requested.
