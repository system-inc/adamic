The native test fleet uses zero-based `ADAMIC_TEST_SHARD=i/n`. Each shard owns pieces whose ordinal modulo n equals i. Use the recorded counts: decoder 90, random regexp 40, normalization 18, WASI 36, checker cache modes 2, record mutants 6. An unset variable runs the complete workload.

Build products are separate inputs. The tests fetch them by source, flags, toolchain and test-executable hashes. `ADAMIC_GATE_UNCACHED=1` still reruns unit execution and comparison outcomes. Decoder Node goldens are a prepared corpus input keyed by oracle source and Node version. The temporary `testbuildcache` store has the `Inputs`, `Get` and `Tool` API of devtools/buildcache at 64ea92b6. Once that helper lands, change the single aliased import in products_test.go to internal/buildcache.

The budget audit fails on elapsed time only with `ADAMIC_UNIT_BUDGET=1` on the reference box. Elsewhere it logs slow timings and still rejects correctness failures. Go units have no new 30-second deadlines. The three opt-in performance probes retain their existing skips.

After cloud/setup.sh, source its printed environment, enable the WASI SDK, and set ADAMIC_CLANG_TSGO_ARCHIVE to the prepared checker archive. Test output goes directly to files:

```sh
python3 internal/native/test-split/measure.py /absolute/path/groups > /absolute/path/groups-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /absolute/path/groups --output /absolute/path/group-units.json > /absolute/path/group-budget.log 2>&1
python3 internal/native/test-split/measure.py /absolute/path/shards --shards-from /absolute/path/groups > /absolute/path/shard-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /absolute/path/shards --output /absolute/path/shard-units.json > /absolute/path/shard-budget.log 2>&1
python3 internal/native/test-split/measure.py /absolute/path/other-units --units-from /absolute/path/groups --exclude-corpus > /absolute/path/other-unit-progress.log 2>&1
ADAMIC_UNIT_BUDGET=1 python3 internal/native/test-split/budget.py /absolute/path/other-units --output /absolute/path/other-unit-budget.json > /absolute/path/other-unit-budget.log 2>&1
python3 internal/native/test-split/budget_policy.py /absolute/path/policy > /absolute/path/policy.log 2>&1
python3 internal/native/test-split/mutants.py /absolute/path/mutants > /absolute/path/mutant-progress.log 2>&1
```

The group run begins with empty Adamic caches per top-level test. The shard run fetches only the shared products prepared by that run. Each shard must execute exactly its owning leaf, the union must equal the complete group, and the reference audit with ADAMIC_UNIT_BUDGET=1 holds both the leaf elapsed time and complete invocation to 30 seconds. Go dependency builds are prepared tool inputs. The mutant driver changes one valid regex program after Node supplies its answers and runs all 40 shards; only shard 2/40 may fail.
