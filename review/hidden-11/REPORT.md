Built: a derived-binding regression fixture and unsafe-capture refusal pin; the hidden boundary remains blocked by __String.
Commit: this feature commit on codex/hidden-11-derived-binding, directly from dcdbb9098f77f30ad41790c56df1bd63ad462b63.
Checks: exact head replay, focused Node/native/JavaScript oracle, negative pin, counts refresh and clean formatting.
Mutants: default-valued derived read changes both backends from 7 to 0; removing self-initializer guard fails its diagnostic pin.
Uncovered: no production compiler fix, no whole package/full gate, __String remains unsupported; the full gate is left to integration.

This serves roadmap step 30 by replacing the proposed closure witness with a checked-in semantic regression and identifying the actual prerequisite behind hidden-11. The short witness already lowers on the selected area tip. No production source is changed by this unit. The delivery is a regression/evidence unit, not completion of the hidden-site burndown.

The recorded head still replays exactly, exit 0: checker.ts:47546:28, NotYet, reading derived. The attempted owner remains checkKindsOfPropertyMemberOverrides at checker.ts:47508:5, under createTypeChecker. All 82 adapted file SHA-256 hashes match the 388096e6 RESULT.json pin. The replay uses the full compiler project, no reduced source or deleted ancestors, and guarded scratch overlays only. replay.json.gz and replay-summary.json retain the actual selected unit and findings. An initial replay started while adaptation was still running; it is discarded. Only the hash-verified replay is reported.

The first finding in this attempted unit is checker.ts:15068:50: a value of type __String. It prevents lowering the initializer of baseSymbol at checker.ts:47539:13. The failed statement is rolled back. The next if reads baseSymbol, and derived's initializer at checker.ts:47543:45 then reads the absent baseSymbol. That statement is also rolled back. The reported derived read is therefore a secondary census failure, not an enclosing capture that the compiler defaulted incorrectly. The independently attempted parent createTypeChecker is checker-diagnosed and split; the regional attempt remains eligible.

A sound implementation must first lower the __String prerequisite. Creating a default local after failed initialization would contradict the census rollback contract and would invent a runtime binding if applied to production. Seeding this same-function failed declaration as an ancestor would also misrepresent the attempt. The existing latent ancestor-binding hook is retained unchanged. Dependency: the phantom-brand __String work on codex/phantom-brands, last delivered d90994daa68bdba300f22acdad849a22ad5192a2, is absent from this area tip. No other worker's unlanded branch was merged, no cohere code was copied, and no language rule was weakened.

The fixture is the exact proposed source:

```typescript
function enclosing(): () => number { const derived = 7; return () => derived; }
console.log(`${enclosing()()}`);
```

Source Node after node:module.stripTypeScriptTypes prints `7\n`, stderr empty, exit 0. The explicit observation suppresses Node's experimental API warning, as the oracle does. The differential oracle matches source Node with release native, ASan/UBSan/leak checking, and generated JavaScript on Node. Its registry has checked=false, so source Node remains the truth. Fixture allocations/frees/retains/releases/peak/regions are 4/4/1/4/4/0.

The separate self-initializing-derived.a witness captures the function variable its own initializer is declaring. TestHiddenDerivedSelfInitializerStaysNotYet pins `a function value that captures the variable its own initializer declares`. It remains NotYet. This negative fixture is not claimed as a successful runtime oracle program.

Mutants were applied independently to real compiler code and restored in finally:

| Mutant | Observed intended catcher |
| --- | --- |
| Replace the captured numeric derived cell read with its default zero | Native and JavaScript differential assertions: Node exit 0 stdout "7\n" stderr "", both backends exit 0 stdout "0\n" stderr "" |
| Remove the self-initializer capture guard | Exact NotYet assertion fails; the separate cycle-capable refusal still prevents compilation |

Both exit 1 with test assertion failures, without Go/clang build errors. The second mutant is a diagnostic catcher; it does not demonstrate an accepted leaking cycle. The restored positive oracle and negative test pass. A final guarded replay record also equals the base record exactly, including ancestor context and ordered findings. Runner, JSON outcomes and raw failure logs are committed in evidence/.

Commands (all test output written to log files):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_derived_binding.a$' -count=1
go test ./internal/lower -run '^TestHiddenDerivedSelfInitializerStaysNotYet$' -count=1
python3 /tmp/hidden-11-mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
gofmt -l internal/oracle/hidden_boundary_derived_binding_test.go internal/lower/hidden_boundary_derived_binding_test.go
git diff --check
```

The restored oracle reports ok 0.595s; restored lower pin ok 0.052s; counts refresh ok 56.095s. Formatting and diff checks produce no output. Only the focused oracle selector was run, except the explicitly required full counts refresh. No whole packages or full gate ran.

The counts refresh initially failed because stage3/api lacked the host fixtures' @types/node 25.3.3 dependency. `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund` installs the existing lockfile (three packages), and the rerun passes. No manifest or lockfile changed. The refreshed table adds the new fixture row; moves logical_and_reference_maybe.a without changing its 8/8/11/22/4/0 numbers; removes the stale stage3/fixtures/taste/17_binder_flow.a row because the selected base's taste_stage3_test.go already registers it with lowers=false. No other row or numeric value changed.

Toolchain setup used the requested GOPROXY=https://proxy.golang.org|direct and passed: Go 0.060s, Node 0.070s, markdown 0.186s (lock validation 0.015s), clang 0.402s, submodules 8.224s, build 202.398s, deferred tests 202.648s, cache warm 202.650s, total 202.682s. nproc=5, cgroup CPU quota=4; Go1.27.1, Node24.19.0, clang20.1.8. The environment file printed by setup is sourced in each build/test shell. Setup itself needed no workaround; the additional host type install was needed for counts.

Base assumption: the explicitly requested compiler/area-next-fixtures tip overrides the old main-start rule. The repository tracked only main, so the area ref was fetched explicitly. The feature branch has only this unit's commits over that tip; evidence pins are detached scratch worktrees, never merged. No main or area branch was pushed or merged into. Push destination is only codex/hidden-11-derived-binding.

Regional measurement completed with the guarded full census and independent byte-mask audit. These are lowering-ledger byte intersections on a checker-rejected program, not whole-program compilation or runtime coverage. The historical 8,143 assigned bytes must not be reported as revealed merely because the small closure fixture passes.


| Assigned checker.ts region | Old hidden intersection | New hidden intersection | Byte difference |
| --- | --- | --- | ---: |
| [2778143, 2786286) | [2778143, 2786286), 8143 bytes | [2778143, 2786286), 8143 bytes | 0 revealed |

Old is the hash-pinned historical 388096e6 result. New is the complete fresh 79-TypeScript-file full census (80 JSONL records including the checker header), using the selected area compiler; the denominator manifest includes all 82 regular compiler files. Since production lowering/loading are unchanged from dcdbb909, the completed fixture-only unit uses that same compiler measurement. Base and final replay records are identical. No other corpus or group total is attributed to this unit as revealed bytes.

The next head boundary is unchanged: checker.ts:47546:28, NotYet, reading derived. Its failed-statement span is [2778143, 2778264), owned by checker.ts:47508:5. The following if statement remains blocked by reading derived at checker.ts:47551:17, span [2778264, 2786286). The prior failure in the same attempted owner remains __String at checker.ts:15068:50. This unit therefore does not satisfy the implementation burndown condition; the sound binding behavior is pinned, and the actual prerequisite is named for integration.

Measurement commands, all exit 0:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-11-before-overlay
go build -buildvcs=false -overlay=/tmp/hidden-11-before-overlay/overlay.json -o /tmp/hidden-11-before-replay ./stage3/census/latent/replay/worker
/tmp/hidden-11-before-replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/checker.ts:47546:28 -kind NotYet -reason 'reading derived'
go build -buildvcs=false -overlay=/tmp/hidden-11-before-overlay/overlay.json -o /tmp/hidden-11-census ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-11-census /tmp/hidden-adapted/src/compiler /tmp/hidden-11-full.jsonl
python3 /tmp/hidden-11-measure.py
python3 /tmp/hidden-11-pin/stage3/census/hidden/hidden.py /tmp/hidden-adapted/src/compiler /tmp/hidden-11-full.jsonl /tmp/hidden-11-pin/stage3/census/hidden/evidence/stock.json.gz /tmp/hidden-11-result.json --commit dcdbb9098f77f30ad41790c56df1bd63ad462b63
python3 /tmp/hidden-11-pin/stage3/census/hidden/audit.py /tmp/hidden-11-full.jsonl /tmp/hidden-11-pin/stage3/census/hidden/evidence/stock.json.gz /tmp/hidden-11-result.json
```

The replay prints reproduced NotYet with the exact position/kind/reason. The regional calculation's independent byte mask passes, with zero source hash mismatches. The full audit prints PASS: independent byte-mask oracle; 82 files. Measurement files and output are under evidence/: full.jsonl.gz, stock.json.gz, hidden-result.json, region.json and measure.py. production-source-hashes.json verifies every production lower/load source against the area base and pins the measurement binary. No production census hooks or permissive loader were committed.
