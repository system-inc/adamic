Requested full latent count: **not obtained**; the full latent comparison is blocked.<br>
Separate compiler-owned source scan: **116 sites**: **111 type-only**, **4 real value**, **1 unclassified**.<br>
Draft 90 retypes **69 declarations in seven files**; the source scan falls **116 to 5**, with no new finding from that check.<br>
**0 public changes and 0 statement splits are proposed**; the global noop retype remains unclassified.<br>
Node corpus: **106,366 passing, one existing API baseline failure, zero pending**; hooks have zero corpus hits, with separate Node witnesses returning **42** and **17**.

# Void widening, October 20, 2026

This is a partial unit, not the requested completed full latent measurement.
[The table](void-widening-2026-10-20.json) separates each observation from its
judgment. It has exact compiler refusal text, source lines, one-based UTF-16
columns, expressions, resolved return declarations, public reachability,
checked rewrites and runtime evidence. `measurement.fullSiteCount` is null and
`fullClearanceProven` is false. Do not use 116 as a full latent count.

The branch was cut from fetched origin/main
`165f427eb629dda1b27553bca3c01e0328edcaa2`. Compiler and dependency pins come from
that ref's meter worktree. The production compiler is unchanged.

The compiler has two Refused prefixes in internal/lower/void_widening.go:
`a void value at a return site whose function returns ` and
`a void-returning function at the `. Structural paths and the target type remain
in the complete text. Its separate structural-depth NotYet is not counted as a
void-widening Refused site.

The requested per-ref meter ran with STAGE3_METER_LATENT unset. Early preparation
failed because I started it before setup had completed; a later build exhausted
the separate /tmp mount. After those were corrected, the paired run's 30-minute
bound still yielded no completed full latent file record. Longer compiler-root
runs also yielded only headers. The before process disappeared while the cgroup
recorded one OOM kill. The victim PID and the process exit were not retained, so
attributing that specific process to the OOM is an inference. The after run was
stopped after the baseline was lost, having consumed 3,430.31 CPU seconds.
[evidence/full-blocker.json](evidence/full-blocker.json) records the observations.

A separate 60-second SIGQUIT probe sampled predicateReferenceDirections,
predicateSameReference and checker spelling-suggestion lookup. The stack is in
[evidence/stack-probe.log.gz](evidence/stack-probe.log.gz). This identifies sampled
cost; it does not prove a permanent hang. Fixing that implementation is outside
this unit's territory. No full count, full-set mutant delta, or absence of new
refusals from every compiler check is claimed.

To recover useful compiler observations, make-scan.py adds a scratch-only driver
that invokes the unchanged production refuseVoidWidening on every AST node in
all 79 compiler sources. It retains the existing no-output loader/lowering guards.
Its header explicitly says `void-widening-only-source-scan`, not `full`.
All four source scans exit 0, with 79 file records and no panic. This scan includes
checker-diagnosed bodies but performs no lowering or concrete generic
instantiations. It does not execute other refusal checks. The requested full
latent streams therefore cannot be replaced by this result.

| Form | Source-scan sites |
| --- | ---: |
| argument | 76 |
| return-of-void | 39 |
| assignment | 1 |
| initializer, field, other | 0 |

| File under src/compiler | Sites |
| --- | ---: |
| binder.ts | 36 |
| checker.ts | 43 |
| emitter.ts | 17 |
| factory/nodeFactory.ts | 2 |
| parser.ts | 7 |
| program.ts | 2 |
| symbolWalker.ts | 5 |
| transformers/es2017.ts | 1 |
| utilities.ts | 2 |
| watch.ts | 1 |

| Judgment | Sites |
| --- | ---: |
| type-only adaptation | 111 |
| real value | 4 |
| unclassified | 1 |
| public, needs a statement split | 0 proved |

The extractor independently reproduces the supplied first-error run's 88 sites.
The source scan shares 70 of them, loses the 18 executeCommandLine sys.exit sites
already repaired by main's adaptation 79, and adds 46: checker 43, program 2 and
es2017 1. This is a net increase of 28, with different coverage explicitly noted.
Lower's narrower 127 becomes 109 after those same 18 exit repairs. The source
scan adds seven return sites: six concise arrow bodies and checker.ts:8231's
return inside a block arrow. Exact added and removed rows are in the table.

The annotation closure uses stock TypeScript 6.0.3. Each credited site becomes
undefined, including its transitive returned-call dependencies. TypeScript checks
all paths rather than trusting void. The manifest records each retype, including
inferred return annotations and four local parser callback type assertions.
The annotation plan introduces zero compiler-project diagnostics: 140 before,
140 after. The complete upstream compiler/harness/test build also passes.
Every source declaration in this table is internal or lexically local; the
receiving forEachChild signature may itself be public and is recorded separately.

The one unclassified site is watch.ts:679:61, the noop timer fallback. noop's empty
body always returns undefined, but globally retyping it breaks four harness/test
typing sites involving typeof noop and existing assertions. The failed full build
is retained. No proof that every possible type-only repair is impossible is
claimed, so this is not mislabeled as a statement split. The choice was to exclude
noop rather than extend the unit into harness type repairs or add trusted casts.

The four real-value rows are the callback argument and returned call at each of
nodeFactory.ts:1164 and utilities.ts:8582. The open internal registration APIs
permit arbitrary external callbacks, so a closed external implementation
inventory cannot be established. No registration call occurs in the compiler,
harness or test source. Counters at these actual calls recorded zero completed
hits and no values in 23 process snapshots during the full Node corpus. Zero hits
do not prove unreachability. Separate Node witnesses register callbacks returning
42 and 17; forEach retains each truthy value and skips the second patcher. Their
four witness implementations are in hooks.cjs. Both return-to-undefined mutants
fail the early-termination assertions.

The counter was reused from void-reads 2670c4e7. Its identity/hit control passes
and its increment-removal mutant is caught. The instrumented corpus has exactly
the control's 106,366 passes and sole existing API-baseline failure. Both actual
baseline diffs are byte-identical: 51,093 bytes, SHA-256
`5333568a661aea24fc8af3149e9f4d2dfca70731842461c677f385b092adc4f0`.
All ten uninstrumented built JavaScript artifacts and both public declaration
artifacts match the control byte for byte. Public typescript.d.ts stays at 11,449
lines; the existing 223 sanctioned declarations are unchanged.

The source-scan mutant adds exactly one row, 116 to 117, at core.ts:2601:50.
Its definitions exist only in scratch. Restoring bindFunctionOrConstructorType
to void adds exactly binder.ts:2995:24, five to six. Independent full-mode
single-root controls also pass: core 0 to 1 and binder 0 to 1. Those bounded
single-root checks are not claimed as complete full-tree mutant runs.

Commands ran from /workspace/adamic with output directed to logs. Important exact
commands are below; the scratch paths identify the retained artifacts.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 840 bash cloud/setup.sh > /tmp/wks9251-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
STAGE3_METER_COMPILER=per-ref STAGE3_METER_MAIN_REF=165f427eb629dda1b27553bca3c01e0328edcaa2 ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-gate/stage3-meter.0XrKU8/upstream timeout --kill-after=30s 1800 bash stage3/meter/twice-daily.sh > /tmp/wks9251-meter-final.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout --kill-after=30s 7200 /workspace/wks9251-scratch/tmp/stage3-meter.DnDghJ/main-binaries/latent-census /tmp/wks9251-before/src/compiler /workspace/wks9251-scratch/before-full.jsonl > /tmp/wks9251-before-full.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout --kill-after=30s 7200 /workspace/wks9251-scratch/tmp/stage3-meter.DnDghJ/main-binaries/latent-census /tmp/wks9251-after/src/compiler /workspace/wks9251-scratch/after-long.jsonl > /tmp/wks9251-after-long.log 2>&1
python3 stage3/census/void-widening/make-scan.py /workspace/wks9251-scratch/tmp/stage3-meter.DnDghJ/main-source /workspace/wks9251-scratch/tmp/stage3-meter.DnDghJ/main-binaries/latent-overlay/overlay.json /workspace/wks9251-scratch/void-scan-overlay
timeout 180 go -C /workspace/wks9251-scratch/tmp/stage3-meter.DnDghJ/main-source build -buildvcs=false -overlay=/workspace/wks9251-scratch/void-scan-overlay/overlay.json -o /workspace/wks9251-scratch/void-scan ./stage3/census/latent/tool
LATENT_ASSERT_NO_OUTPUT=1 timeout 300 /workspace/wks9251-scratch/void-scan /tmp/wks9251-before/src/compiler /workspace/wks9251-scratch/scan-before.jsonl > /tmp/wks9251-scan-before.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 timeout 300 /workspace/wks9251-scratch/void-scan /tmp/wks9251-after/src/compiler /workspace/wks9251-scratch/scan-after.jsonl > /tmp/wks9251-scan-after.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 timeout 300 /workspace/wks9251-scratch/void-scan /workspace/wks9251-scratch/mutant/src/compiler /workspace/wks9251-scratch/scan-mutant.jsonl > /tmp/wks9251-scan-mutant.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 timeout 300 /workspace/wks9251-scratch/void-scan /workspace/wks9251-scratch/reversion/src/compiler /workspace/wks9251-scratch/scan-reversion.jsonl > /tmp/wks9251-scan-reversion.log 2>&1
timeout 240 node stage3/census/void-widening/plan.cjs /tmp/wks9251-before /workspace/wks9251-scratch/scan-inspected.json /workspace/wks9251-scratch/scan-plan.json > /tmp/wks9251-scan-plan.log 2>&1
timeout 300 node stage3/census/void-widening/verify.cjs /tmp/wks9251-before /workspace/wks9251-scratch/adapter-proof /tmp/wks9251-plan3.json /workspace/wks9251-scratch/adapt.cjs /workspace/wks9251-scratch/verify /tmp/wks9251-verify.json > /tmp/wks9251-verify.log 2>&1
VOID_READS_COVERAGE=/workspace/wks9251-scratch/coverage NODE_OPTIONS='--require=/workspace/adamic/stage3/census/void-widening/observe.cjs' timeout --kill-after=30s 1200 bash stage3/oracle/run.sh /workspace/wks9251-scratch/runtime /workspace/wks9251-scratch/oracle-runtime --workers 4 --limit-seconds 900 > /tmp/wks9251-oracle-runtime.log 2>&1
```

STAGE3_METER_LATENT was unset for every meter attempt. TMPDIR was changed to
/workspace/wks9251-scratch/tmp after /tmp's 8.8 GiB limit caused ENOSPC.
The duplicate same-SHA area entry measurement was stopped to reserve CPU; no
paired meter pass is claimed. Raw run directories are not committed.

The first setup hit its 840-second bound during the filtered submodule fetch.
The retry was stopped and the interrupted checkout preserved in scratch; the
exact d92d9bfee114c80be2c375d72edae966176e3a4f submodule was restored from the
already completed meter worktree. An ENOSPC build then failed before the final
successful setup. Final timing lines: Node 0.027s, Go 0.035s, markdown 0.085s,
submodules 0.116s, clang 0.264s, shared cache off 0.266s, Go build 69.559s,
test binaries deferred 69.725s, build cache warm 69.727s, done 69.757s.
nproc=5; cpu.max=400000 100000; Node 24.19.0, Go 1.27.1, clang 20.1.8.

The missing full measurement, complete full-tree mutant comparison and no-new-
refusal proof remain explicit blockers. Draft 90 is not a completed landing
proof. No PR is opened, and nothing under internal/ or another adaptation is
changed by either commit.
