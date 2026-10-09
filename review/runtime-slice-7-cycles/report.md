# Slice 7 cycles: baseline evidence and counted-equality stop

Task #aacs25v, label runtime-slice-7-cycles. Read the previous report at
9651c792 before doing this work. Slice 6 production source is exactly
9674ce3f7854d009ec51212840d9629d81df6daf. Slice 7 starts at
9651c792b09efdc2af549a775980c63a7593e480.

## Decision and stopping reason

Slice 6 accepts all twelve probes in default mode: cwd and the eleven review
programs identified below. This selects path 3, not path 2. Default acceptance
was already present in slice 6; it was not introduced by slice 7's cherry-picks.
No cycle-policy production changes or fixture moves were made.

All twelve match Node's stdout, stderr and exit code under ASan and UBSan and
pass the shared internal/leakcheck helper on Linux (LeakSanitizer). Eleven meet
the explicit allocations == frees requirement. The devirtualization fixture
reports allocations 15, frees 14, statement-region values 1. The shared helper's
counted balance check passes because 15 == 14 + 1. The strict equality assertion
fails. This is evidence of a counted-equality boundary, not evidence of a leak,
sanitizer failure or Node disagreement. A statement-region value is freed with
its region and is recorded separately from individual frees by this runtime.

Named stop: the task's stricter raw allocations == frees condition is unmet for
fxspptb_devirt_fresh_method_keeps_argument.a. Do not relabel this as a leak or
silently replace that condition with allocations == frees + regions. No fixture
was moved while the full acceptance requirement was unmet. No policy question
was re-asked. The Program opt-in ruling is accepted: member cycles are allowed;
legacy refusal suites need not refuse member cycles in that mode.

## Baseline preparation and results

Fetched the requested branches and created an isolated worktree at 9674ce3f.
Pinned submodules fetched recursively through HTTPS under timeout 900; stage3/api
npm ci --ignore-scripts installed the locked Node types under timeout 900.
The first build attempt hit /usr/bin/go (an unrelated executable); the second
needed the missing /tmp/adamic-gate directory. After sourcing
/workspace/adamic-tools/env.sh and creating that directory with mode 1777,
both lower and oracle binaries compiled successfully under timeout 900.
Build time was outside each test's 90-second execution budget.

Every baseline execution used ADAMIC_PROGRAM_REGION=0,
ADAMIC_GATE_UNCACHED=1, external timeout -k 5 90, internal timeout 85s.
Suite executions used -test.parallel=4. Logs are included.

| Binary / selection | Exit | Observed result |
|---|---:|---|
| oracle TestReviewProgramsRefuse | 1 | Eleven accepted review fixtures; 42 refused leaves pass, 2 existing pending leaves skip |
| lower cycle-capable tests | 1 | Only cwd expectation fails because lowering returns nil; all other selected leaves pass |
| oracle GraphRegion\|FreshRefused | 0 | All selected graph tests pass, including million and recorded graph counts |
| oracle TestFreshWriteProbesUseRegions | 0 | All 54 fresh_refused fixtures pass Node/native/JS/release/leak/count checks |

The lower selection was NestedCallbackCycleIsRefused,
AsyncGeneratedIdentityCannotBeClaimedBySource, PromisePayloadCannotHideUserCycles,
NamespaceAmbientHostInitialization, InheritanceRefusesUnsoundOverrides and
InheritanceGenericSoundness. It includes all test sites mentioning
adamic/cycle-capable on this baseline.
FreshRefused matches no test on slice 6, so TestFreshWriteProbesUseRegions was
also run explicitly. The graph pattern also selects the memory-example graph
suite. Counts reported by these tests were checked, not regenerated.

## Accepted-probe safety evidence

baseline-witness.go.txt is the temporary test added only to the slice 6 worktree.
Production source remained at the baseline SHA. It uses the existing oracle Node
runner and native.Options{Sanitize: true}, leaksUncached (the shared helper), and
a separate native.Options{Count: true} build. Each leaf ran independently with a
90-second external kill. The loop stopped at the strict count failure; cwd was
then run independently to complete the twelfth probe's safety evidence. All
commands completed; no verification process remains running.

The eleven review files are still under internal/oracle/testdata/review/refused;
cwd is still stage3/namespace-init-sys/cwd.a. Every row below matches Node and
passes ASan, UBSan, LeakSanitizer and the shared counted-balance helper.

| Program | Allocations | Frees | Statement-region values | Strict equality |
|---|---:|---:|---:|---|
| `cwd.a` | 6 | 6 | 0 | PASS |
| `fxspptb_aee89b2_alias_fill.a` | 603 | 603 | 0 | PASS |
| `fxspptb_aee89b2_alias_mapset.a` | 603 | 603 | 0 | PASS |
| `fxspptb_aee89b2_alias_reverse.a` | 403 | 403 | 0 | PASS |
| `fxspptb_aee89b2_alias_setadd.a` | 603 | 603 | 0 | PASS |
| `fxspptb_aee89b2_alias_sort.a` | 503 | 503 | 0 | PASS |
| `fxspptb_aee89b2_clobber_link_loop.a` | 1002 | 1002 | 0 | PASS |
| `fxspptb_aee89b2_escaped_before.a` | 1007 | 1007 | 0 | PASS |
| `fxspptb_aee89b2_map_entries_pattern.a` | 802 | 802 | 0 | PASS |
| `fxspptb_aee89b2_spread_outside.a` | 10 | 10 | 0 | PASS |
| `fxspptb_aee89b2_throw_keeps_old.a` | 7 | 7 | 0 | PASS |
| `fxspptb_devirt_fresh_method_keeps_argument.a` | 15 | 14 | 1 | FAIL, one value in a statement region |

## Still unrun and pending

Slice 7's native Program|Region|RuntimeStatics bounded units, lower
Program|Membership|Cycle, and oracle Program|GraphRegion|FreshRefused|ReviewPrograms
in both modes were not rerun this turn. The counts generator was not launched;
counts.md is unchanged, with every Linux row retained. Fixture migration and
slice 7's default graph restoration remain unfinished at the named stop.
Host fixture 14 remains pending on the non-null assertion feature on this base,
not a verification failure. No force push, merge or main update was performed.
