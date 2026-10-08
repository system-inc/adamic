Built: boxed String observations, six String-to-RegExp entry points, constant match/search patterns, catchable repeat/fromCodePoint RangeErrors and shared TypeError identity.
Commits: claim 63d737a, regex-protocol merge de0c4e9 (ffb3ae0), implementation f5f66da; branch codex/library-string-3 from e911e46.
Commands and outputs: mandated Linux survey 257 to 288 pass, 217 to 187 refused, zero disagreements/crashes; focused packages, uncached oracles, counts and vet pass.
Mutants: 12 new and 22 inherited automated behavior mutants caught only by Node stdout; a separate deleted representation guard is caught by refusal controls.
Uncovered: 16 String library first obstructions, 168 language obstructions for @system_adamic, and 3 stock-TypeScript rejections still counted refused by the runner.

# String third slice report

| Survey | Pass | Disagreement | Refused | Crashed | Skipped | Not TypeScript | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Before, e911e46 | 257 | 0 | 217 | 0 | 316 | 433 | 1223 |
| After, f5f66da | 288 | 0 | 187 | 0 | 316 | 432 | 1223 |

All 257 baseline passes are retained. Each of the 31 new passes was independently rerun against the frozen baseline and was refused there. Stock TypeScript 6.0.3 accepts every exact adapted new-pass program. Every pass agrees with independently executed Node; no disagreement or crash remains. Before/after JSON and progress logs are committed alongside this report.

The net refusal reduction is 30, rather than 31, because replace/S15.5.4.11_A1_T17.js moves from not-typescript to refused. Its exact adapted source is rejected by stock tsc with TS2769 using both baseline and current preludes. The imported checker declarations now let Adamic reach a nonconstant RegExp pattern refusal; the mandated runner consults stock tsc only for checker diagnostic refusals. This is a measured classifier limitation, not implementation credit or library work. diagnostic-audit-3.jsonl and replace-stock-3.jsonl record the finding. The two pre-existing different-diagnostic-code refusals also remain stock-TypeScript rejections. The runner was not changed to improve the table.

The first commit is the complete claim and library/language grouping in claim-3.md. inventory-3.md lists every remaining observed reason with a one-line language reproducer for @system_adamic. The 16 remaining String library rows are locale casing/collation (11), hidden ToPrimitive members (4), and a boxed String viewed through erased fields (1). Prototype mutation, mixed binary coercion, first-class methods, arguments, other constructors, hoisting and the other language features were not built. Dynamic RegExp patterns, custom protocols/exec/species/getters and richer callback signatures remain dependencies. Locale support requires exact Node-compatible ICU behavior; Node reports ICU 78.3, and neither icu-uc nor icu-i18n is available through pkg-config here. No approximate conversion or collation is substituted.

implementation-3.md describes the admitted boxed representation and the reuse boundaries. The original fromCodePoint disagreement, previously refused on String-2, now passes with a catchable, cleanup-aware RangeError. Existing null-receiver TypeErrors now use the RegExp dependency's exact builtin constructor identity. The old range refusal fixture is now a passing oracle fixture; no fixture was removed to improve results. New source fixtures use .a.

## Measurement command

Runner source: unmodified origin/codex/test262-ts-validity af12899 in the separate /workspace/adamic-validity-string-3 checkout. Build command was `go build -buildvcs=false -o /tmp/library-string-3-runner ./cmd/adamic-test262`. The worktree's cohere directory points at the existing checkout; disabling the VCS stamp handles that filesystem arrangement. Compiler baseline is a separate frozen e911e46 worktree; final compiler is built from this branch. Test262 pin: c8c798898646638cd0c24879f8e0374e847e7d74. Linux, Node v24.19.0, clang 20.1.8, Go 1.27.1, stock tsc 6.0.3.

```sh
source /workspace/adamic-tools/env.sh
export GOFLAGS=-buildvcs=false
export PATH=/workspace/string-tsc/node_modules/.bin:$PATH
/tmp/library-string-3-runner -adapt -json -root /workspace/library-string-3-baseline -test262 /workspace/test262 -work /tmp/library-string-3-before-valid-work built-ins/String > /tmp/library-string-3-before.json 2> /tmp/library-string-3-before.log
/tmp/library-string-3-runner -adapt -json -root /workspace/adamic -test262 /workspace/test262 -work /tmp/library-string-3-final-work built-ins/String > /tmp/library-string-3-after.json 2> /tmp/library-string-3-after.log
```

The stock-tSC audits used temporary tests in the separate runner checkout, calling its unchanged classify/startTypescript helpers on exact adapted sources. Temporary audit tests were removed. They do not change the recorded survey outcomes. new-pass-baseline-3.json records all 31 isolated baseline refusals; new-pass-stock-3.jsonl records all 31 independent stock-tSC acceptances.

## Mutants

Every behavior mutant below still emits valid C, exits zero, and passes sanitizer and leak checks. Only the independent source-on-Node stdout comparison rejects it. TestLibraryStringThirdMutants runs these twelve:

| Family | Mutation | What catches it |
|---|---|---|
| Box construction | Report length one unit too large | Node stdout |
| Box indexing | Increase numeric index by one UTF-16 unit | Node stdout |
| Box own properties | Include the length boundary as an indexed own property | Node stdout |
| match | Drop the first input unit, changing match.index | Node stdout |
| matchAll | Drop the first unit, splitting the leading surrogate pair | Node stdout |
| replace | Drop the first input unit before the callback protocol | Node stdout |
| replaceAll | Drop the first input unit before the callback protocol | Node stdout |
| search | Drop the first input unit, changing the result index | Node stdout |
| split | Drop the first input unit, changing the leading segment | Node stdout |
| RegExpCreate fallback | Drop the first unit before the constant-pattern match | Node stdout |
| fromCodePoint | Reject 0x10ffff by changing > to >= | Node stdout |
| repeat | Reject zero by changing < 0 to <= 0 | Node stdout |

The twelve inherited first-slice mutants (conversion, prototype, charAt, at, codePointAt, substring, concat, raw, rawTemplate, fromCharCode, normalize, replace) and ten inherited second-slice mutants (identity, primitive, arrayConversion, mapConversion, startsWith, endsWith, isWellFormed, toWellFormed, nullReceiver, rawLimit) also run uncached and are rejected only by Node stdout. The nullReceiver mutant now changes builtin TypeError identity to RangeError identity, rather than changing a synthetic name.

A separate mutation in a detached worktree inserts an immediate `return nil` into stringBoxViewRefusal. TestLibraryStringRefusals rejects it; primitive-to-String and spread probes incorrectly lower with no error, while other probes lose their explicit representation refusal. guard-mutant-3.log records the failures. The actual implementation was never mutated for this control. The current guard controls, including readonly nested views and casts, pass in the full lower package.

## Validation and setup

- `bash cloud/setup.sh`: success; Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 59s, total 59s. `nproc`: 5; cgroup cpu.max 400000 100000; memory 17.6 GB. Printed environment path: /workspace/adamic-tools/env.sh. setup-3.log records the timings.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/flow ./internal/fresh ./cmd/adamic-test262 -count=1`: load 1.428s, native 253.903s, flow 97.424s, fresh 41.297s and runner 110.375s passed. The command initially failed only on the obsolete lower assertion requiring repeat in try to refuse. After replacing it with independent caught-error/finally fixtures, `go test ./internal/lower -count=1` passed in 10.931s. packages-3.log preserves the initial failure; lower-3.log records the complete corrected lower-package pass.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestLibraryString.*Mutants|TestNativeAgreesWithNode/internal/oracle/testdata/(library_string_|regexp_|from_code|from_codes|move_throw|long_literals|search_from_sweep|string_too_long)|TestRegExpProtocolCounts' -count=1 -v`: pass in 16.105s, zero cache hits; 178 native and 106 Node executions. oracle-3.log records every fixture and mutant.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`: pass in 16.662s, regenerated the full ownership table. `go test ./internal/oracle -run 'TestCountsAreRecorded|TestRegExpProtocolCounts' -count=1`: pass in 15.105s. The full table is retained; no filtered update drops unrelated rows.
- `go vet ./internal/lower ./internal/ir ./internal/oracle`: exit 0, empty output. `git diff --check`: clean. New/changed Go files are gofmt formatted.

This is a focused Linux gate, not a claimed full `go test ./...` run. No PR is opened. The branch is pushed for @system_adamic to merge.
