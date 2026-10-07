Built: merged nullable references and coverage; typeof lookup acceptance follows main and source Node.
Commits: main 71d7e491, nullable representation 98ad7e10, coverage 49213658; this report is in the reconciliation merge commit.
Commands and outputs: complete lower, native, runner and regexp suites; filtered oracle and Linux counts; raw logs linked below.
Mutants: restored refusal, null classification, lost slot presence and altered count row and wrong schema flag were caught; suite mutants are listed below.
Not covered: full repository gate, full differential oracle, full upstream test262 corpus, opt-in benchmarks and non-Linux counts.

The previous coverage refusal predated typeof-null-2. Main's accepted lookup at
`internal/oracle/testdata/typeof_null_slots.a:3` matches source Node byte for byte
with nullable references' pointer representation in effect. There is no observed
representation miscompile to fix. Keep main's classification and update nullable
coverage to require acceptance.

## Observations

The exact fixture runs three ways: unmodified source on Node, emitted JavaScript
through `oracle/node.mjs`, and compiled native. Each exits zero, prints identical
161-byte stdout and empty stderr. The differential oracle also runs native with
ASan/UBSan and leak checks. Raw output is in [reconciled/](reconciled/).

Generated C uses `adamic_array *` for nullable RegExpExecArray values. Its typeof
calls pass both the nullable pointer and the actual lookup slot's `!= NULL`
presence, distinguishing stored null from a missing lookup. Inference: immediate
classification retains enough information; storing a general R|null|undefined
value still requires a tag and remains refused.

The lower guard allows immediate typeof observations, including parentheses;
main's typeof lowering still requires ArrayIndex, MapGet or ArrayPop IR when both
empty cases occur. Six acceptance witnesses cover indexing, parenthesized indexing,
array.at, array.pop, map.get and parenthesized map.get. Other triple-empty reads
and changed nullable views remain refused.

Linux regenerated all 443 count rows. Existing main rows are unchanged. Fourteen
nullable fixtures and the inherited regexp_literal_freshness fixture add rows.
The typeof_null_slots row remains 14 allocs, 14 frees, 26 retains, 36 releases,
5 peak objects, 0 regions. Main's direct TypeOf IR avoids the incoming helper's
extra ownership operations, so incoming nullable count values were regenerated.

## Merge integration

This is a real merge of 49213658, carrying 98ad7e10 and its inherited regexp and
test262 changes. Preserve main's iterator checks and worker/cache scheduling,
and incoming concrete nullable class handling, timeout setting, original Node
RegExp verdicts and results.jsonl audit records.

The first complete package run found two additional integration failures:

- Main's records C harness lacked the new JSON schema `null_reference` field.
  Initialize all five existing nonnullable schemas with false. This is a harness
  compatibility fix, not a change to nullable value representation.
- Incoming attempt-side diagnostic writes broke serial/parallel log equality.
  Keep main's ordered reducer logging; per-result verdict/reason remains recorded.

The initial failing run is preserved in [packages.log](reconciled/packages.log).
No files were copied from cohere. The existing cohere submodule remains referenced.
No protected internal/native/emit.go, internal/native/native.go,
internal/lower/lower.go or internal/oracle/oracle_test.go was edited.

## Commands

All test stdout and stderr were redirected directly to log files, never piped.
The environment was sourced from `/workspace/adamic-tools/env.sh` and GOPROXY was
`https://proxy.golang.org|direct`.

```sh
bash cloud/setup.sh
# nproc: 5; cgroup cpu.max: 400000 100000
# Go ready .092s; Node .095s; submodules .107s; clang .786s
# markdown deps 1.597s; cache warm 58.503s; done 58.577s
# Go 1.27.1, Node 24.19.0, clang 20.1.8, Linux amd64

go test ./internal/lower ./internal/native ./cmd/adamic-test262 -count=1 -timeout 30m
# lower passed 59.735s; native and runner initially failed, then fixed

go test ./internal/native ./cmd/adamic-test262 -count=1 -timeout 30m
# passed native 295.679s, runner 80.482s; compiled runtime mutants caught

go test ./internal/regexp -count=1 -timeout 30m
# passed 10.744s

ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle \
  -run 'TestNullable|TestTypeOfNull|TestNativeAgreesWithNode/internal/oracle/testdata/(typeof_|nullable_)|TestCountsAreRecorded' \
  -count=1 -timeout 30m
# passed lower 3.884s, oracle 89.308s after restoring all source/table mutants
# 7 typeof fixtures, 14 nullable fixtures including 8 coverage programs

go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
# passed 95.453s, regenerated Linux counts

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_literal_freshness|TestRegExpLiteralFreshness' \
  -count=1 -timeout 30m
# passed 3.173s

go vet ./...
gofmt -l cmd internal
# passed; formatting output empty
```

Setup, commands' raw results, deliberate failures and final green runs are in
[reconciled/](reconciled/). Full staged whitespace checking encounters inherited
CRLF in regexp directories.csv and trailing spaces in its upstream LICENSE;
reconciliation source changes pass whitespace checks.

## Mutants actually run

| Mutant | Check that caught it |
| --- | --- |
| Restore nullable's blanket triple-empty refusal | New positive lower test fails on all six lookup forms, after successful compilation |
| Change TypeOf null classification to undefined | Source Node stdout comparison on typeof_null, compare, switch, slots, roll; clean sanitized exits |
| Force all lookup slots present | typeof_null_slots source Node stdout comparison; missing slots print object, clean sanitized exit |
| Change recorded typeof_null_slots retains from 26 to 27 | TestCountsAreRecorded reports measured 26 versus recorded 27 |
| Mark nonnullable numeric schema nullable | Record semantics source Node comparison; zero wrongly serializes as null, clean counted execution |
| Record indices-in-insertion-order | Record Node stdout comparison |
| Record uint32-max-as-index | Record Node stdout comparison |
| Record deleted-key-iterated | Record Node stdout comparison |
| Record overwrite-key-leaked | LeakSanitizer |
| Record stored-key-freed | ASan heap-use-after-free |
| Record own-slot-null-read | UBSan null member access |
| Record prototype-membership-restored | Exact diagnostic and exit contract, forbidden stdout 1 |
| Record missing-read-silent | Exact diagnostic and exit contract, forbidden stdout 0 |
| Record own-read-checked-as-missing | Own-hit fixture, wrong exit 70 and diagnostic |
| Map old hash | Probe bound 64 exceeded on integers |
| Map zero normalization removed | Zero hashes differ |
| Map NaN normalization removed | NaN hashes differ |
| RegExp adapted false-pass assertion | Independent original-source Node rejects false pass |
| RegExp wrong error constructor | Constructor-identity guard refuses unsupported proof |
| RegExp wrong capture | Runtime verdict fails rather than passing/refusing/crashing |

The restored numeric-schema semantics check passed in 0.684s. The complete native
suite and runner rerun both passed; their raw results are in packages-fixed.log.

The first record mutant run died at -Werror on the schema mismatch and is not
counted as a mutant kill. Only the fixed complete suite's compiled mutant checks
count. Dedicated typeof mutation output is in typeof-mutants.log; source refusal
and count-table deliberate failures have separate logs. All manual mutations
were restored before the final green oracle/counts command. The later schema-flag
mutant was restored and its record semantics check rerun green separately.
