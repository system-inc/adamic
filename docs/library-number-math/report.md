Built: immediate Number slots, optional formatting arguments, global predicates, and intrinsic metadata; 44 newly passing tests.
Commits: implementation 1cd6780, based on main ef3d907; this report is the following documentation commit.
Commands and outputs: official validity survey 194 Number / 120 Math passes, zero disagreements; touched packages, filtered oracle, counts, formatting, and vet pass.
Mutants: fourteen executed; each caught by Node stdout disagreement with clean sanitizer execution, detailed in mutants.md.
Uncovered: 28 Number and 36 Math refusals remain; shared language requirements, Math.random, and four diagnostic-code mismatches are documented.

## Measurements

Linux is the gate of record. Both tables use the unchanged runner built separately from `origin/codex/test262-ts-validity` at `af128990588aab7ee52ea3ab8527d638009c4979`, with TypeScript 6.0.3 and test262 at `c8c798898646638cd0c24879f8e0374e847e7d74`. The implementation starts from main `ef3d907ecdc4c771b016f7d9c52372def057a340`. The owned Number slot change and four fixtures were brought forward from `735a7026ab98778ffa9b97cf85861ac1a593f3b7`, without merging that branch's unrelated changes.

| Before | Pass | Disagreement | Refused | Crashed | Skipped | Not TypeScript | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| built-ins/Number | 152 | 0 | 70 | 0 | 74 | 44 | 340 |
| built-ins/Math | 118 | 0 | 38 | 0 | 161 | 10 | 327 |

| After | Pass | Disagreement | Refused | Crashed | Skipped | Not TypeScript | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| built-ins/Number | 194 | 0 | 28 | 0 | 74 | 44 | 340 |
| built-ins/Math | 120 | 0 | 36 | 0 | 161 | 10 | 327 |

Observed: every old pass is retained; all 44 new passes come from the refused column; every new pass agrees with Node. Machine-readable aggregates are [before.json](before.json) and [after.json](after.json). Exact newly passing test paths are [new-passes.txt](new-passes.txt).

The observation-only copy of the runner emits per-test result records; its baseline aggregates, refusal reasons, and pass lists were checked equal to the unchanged runner. Those 667 baseline records are [before-results.jsonl](before-results.jsonl). The unchanged runner determines both published tables.

[refusals.md](refusals.md) lists every original reason and count, sorted by size, and the library work completed. [language-handoff.md](language-handoff.md) gives one-line TypeScript-valid reproducers for @system_adamic. Stock tsc and Adamic outputs for those probes are preserved in [reproducers.json](reproducers.json) and [reproducers.log](reproducers.log). Four Math.sumPrecise tests remain refused because Adamic emits TS2550 while stock tsc emits TS2339: the prescribed runner matches diagnostic codes. This is observed rejection by stock tsc, not inferred library work.

## Implementation and limits

Lowering uses existing exact runtime conversions and formatters. Immediate Number constructions need no persistent allocation, and preserve argument evaluation. Metadata observations do not make detached functions generally callable. Own-property keys are converted once in a normal IR helper. A three-line typeof dispatch hook and separate oracle registration are the only shared code additions; the count table records the eight new fixtures.

No runtime algorithm was changed or newly ported. The protected ieee754.c, radix.c, parse.c, and hypot.c remain untouched, as does library_math_number.c; no THIRD_PARTY_NOTICES entry was needed. Stored boxes, dynamic ToPrimitive, observable prototype operations, growing arrays, and catchable formatting errors require shared language/value machinery. Treating those as simple numeric fields would lose observable behavior, so they remain refused. Math.random remains refused under the independent deterministic stdout oracle contract.

## Verification

Tool setup: `bash cloud/setup.sh` completed in 105 seconds: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node v24.19.0 ready 0s, submodules 0s, cache warm 105s. Environment file: `/workspace/adamic-tools/env.sh`. `nproc` reports 5; CPU quota is 4 CPUs. Full setup output is [setup.log](setup.log).

Survey command, before and after, after building the implementation compiler:

```sh
PATH=/workspace/test262-typescript/node_modules/.bin:$PATH GOMAXPROCS=2 /workspace/adamic-test262-ts-validity -root /workspace/adamic -test262 /workspace/test262 -work /workspace/number-math-final -adapt -json built-ins/Number built-ins/Math
```

The runner builds the checkout compiler into its work directory. The baseline compiler was built from main and the final compiler from this implementation. The command above includes the TypeScript executable search path and final reduced concurrency setting.

Successful Linux checks, with stdout/stderr redirected directly to log files:

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=2 ADAMIC_GATE_UNCACHED=1 go test -p1 -parallel2 -v -count1 -timeout30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_(number|math)_|TestLibraryNumberMathOracleCatchesMutants|TestMathNumberOracleCatchesMutants'
GOMAXPROCS=2 go test -p1 -parallel2 -count1 -timeout30m ./internal/lower ./internal/native
GOMAXPROCS=2 go test -p1 -parallel2 -count1 -timeout30m ./internal/oracle -run TestCountsAreRecorded
gofmt -l cmd internal
GOMAXPROCS=2 go vet -p1 ./...
git diff --check
```

The filtered oracle passes 11 fixtures and all fourteen mutants in 4.576s, with uncached native execution. Lower passes in 13.039s; native in 79.953s; recorded counts in 19.105s. Formatting and vet outputs are empty, with exit 0. See [unit-gate.log](unit-gate.log), [packages.log](packages.log), [counts.log](counts.log), [format.log](format.log), and [vet.log](vet.log). [mutants.md](mutants.md) identifies every executed mutation and its catch.

The attempted uncached full `go test -count1 -timeout30m ./...` did not complete: the execution server disconnected and its sessions disappeared. [full-gate-interrupted.log](full-gate-interrupted.log) is incomplete and is not a passing gate. No OOM kill was recorded; the cause remains unknown. The touched-package and filtered-oracle checks above are the completed fallback requested by the unit rules.
