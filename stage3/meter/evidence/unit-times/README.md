# Latent unit time proof (#679q9d7)

Built from statementrewrite base `55c1c38fbb0840fbd4c2b93579c8722d4c1df35b` with the timing changes in this commit. No compiler source under `internal/` changed.

The real first-error census lowered the adapted TypeScript 6.0.3 compiler directory once with timing, then once without timing, using the same binary and source tree. Both runs enabled `LATENT_ASSERT_NO_OUTPUT=1`.

All 4,017 attempted units across 79 files have exactly one time record, in ledger order. The 498 deferred generics and 28 checker-body skips have none.

This box's unit times: 580.009824 wall seconds and 631.938860 process CPU seconds. These times are not comparable across boxes.

[Complete time stream](latent-times.jsonl.gz), [box metadata](box.json), and [commands, source/binary hashes, summaries and test output](proof.json).

## Byte identity

`cmp /tmp/meter-unit-proof/latent.jsonl /tmp/meter-unit-proof/untimed.jsonl` exited 0 with no output.

Both findings files have SHA-256 `9c035e9edf81136b7617b97424404e86ac50a41733e40c460a7540870679f2bc`.

## Ten slowest units

| Where | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| src/compiler/moduleSpecifiers.ts:433:1 | 5.653752 | 5.697332 |
| src/compiler/commandLineParser.ts:1722:1 | 5.463495 | 5.501743 |
| src/compiler/moduleSpecifiers.ts:301:1 | 5.445309 | 6.734941 |
| src/compiler/utilities.ts:2638:1 | 5.311620 | 7.428198 |
| src/compiler/utilities.ts:2036:1 | 5.228691 | 5.241222 |
| src/compiler/commandLineParser.ts:1728:1 | 5.167933 | 5.199083 |
| src/compiler/builder.ts:1621:1 | 5.163446 | 5.174607 |
| src/compiler/utilities.ts:6144:1 | 5.161594 | 5.197365 |
| src/compiler/moduleSpecifiers.ts:288:1 | 5.147976 | 5.177219 |
| src/compiler/parser.ts:1386:1 | 5.109005 | 5.107989 |

## Checks

- `go test ./stage3/census/latent/... > /tmp/meter-latent-tests.log 2>&1`: passed.
- From `stage3/meter`, `python3 -E -S -B -m unittest report_test entry_test > /tmp/meter-report-tests.log 2>&1`: OK, 36 tests, 4 optional binary-dependent tests skipped.
- `TestLatentUnitTimerWritesSeparateSequentialRecords` checks the exact four-field schema, sequential records, recovered failure, positive wall and nonnegative CPU time, EOF, and the disabled recorder. It calls no subprocess and reads no repository file. Its recorder is compiled verbatim into the scratch lower overlay.
- The actual scratch driver is exercised outside Go tests: first-error and full smoke runs each produced two records for two eligible units and omitted one checker skip and one generic deferral. The real run above verifies complete driver coverage. A Go integration test would require building the scratch compiler overlay; the ordinary package test keeps compiler subprocesses out of tests.
- `git grep -n 'func TestLatentUnitTimerWritesSeparateSequentialRecords('`: one repository match.
- `bash -n stage3/meter/twice-daily.sh` and `git diff --check`: passed.

## Mutant

`python3 -E -S -B stage3/meter/check-unit-time-mutant.py > /tmp/meter-unit-time-mutant.log 2>&1` removes each tree's `latent_time` block from serialized JSON without breaking report generation. The test goes red at the missing data:

```text
ERROR: test_meter_unit_times_block_sorts_files_and_shortens_units (report_test.ReportTests.test_meter_unit_times_block_sorts_files_and_shortens_units)
report_test.py:132: summary = disk['trees'][label]['latent_time']
KeyError: 'latent_time'
```
