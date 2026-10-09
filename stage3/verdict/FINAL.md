Built: final.sh runs the complete Node verdict before permitting native performance measurement.
Base: 98d2269a; performance commit afb0651c merged through d39b22e7.
Observed: A 301/301, 1/1, 13,693/13,693; six workloads timed, 60 native-slot samples.
Mutants: B changes one diagnostic byte; C exits 1 silently; timing-permission and artifact mutants have focused checks.
Not covered: the API-only baseline floor and a real native tsc, which still does not link.

Use the day the native binary links:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/final-setup.log 2>&1
source /workspace/adamic-tools/env.sh # use the printed path
stage3/verdict/final.sh --tsc /absolute/native-tsc /tmp/new-final > /tmp/final.log 2>&1
```

The output directory must be new and outside package metadata. The compiler
argument is one executable, including paths with spaces. Correctness runs first:
301 acceptance projects, tiny, and all 13,693 selected upstream configurations.
Node's observed goldens and upstream summary expectations remain unchanged.
No new baseline selection or expected output is introduced by this unit.
The 614 excluded inputs, including the 608 diagnostic-collection API floor,
remain as recorded in REMAINING.md.

The command validates the child verdict exit, success flag, exact suite
population, total/pass/fail consistency, failure records and zero deferred
cases. Any mismatch refuses the performance invocation. Missing or malformed
artifacts are infrastructure errors. The supplied executable's SHA256 must
remain unchanged through correctness and performance. This detects replacing
the executable after its correctness run; it does not hash arbitrary shared
libraries or interpreter dependencies of a stand-in wrapper.

Only after correctness passes does it invoke exactly:

```sh
stage3/performance/run.sh --native /absolute/native-tsc /tmp/new-final/performance
```

The merged performance runner is unchanged from afb0651c. It pins Node 24.19.0,
TypeScript 6.0.3 and typescript-go 7.0.0-dev.20260707.2, checks identical stock
library bytes, and performs its own global native preflight. A performance
input mismatch also refuses all warmups, timed runs and profiles. Inputs with
Node/Go disagreements are recorded as performance exclusions. Every accepted
input has two warmups and ten fresh-process samples for native, Node,
Go single-threaded and Go default, plus a separate peak-RSS run. See the merged
performance README for exact workloads and shared-machine measurement limits.

Final summary.json and summary.md combine correctness and timing in one table.
The src/compiler row must be present with all four modes and ten finite,
positive samples each. Recorded mean and sample SD are checked against those
samples. Missing native results cannot look like successful measurement.
The **1.78 s** historical compiler bar is separate from the freshly measured
Go default mean: the table reports both. A native mean below the bar prints
"met"; otherwise "missed". Exit 0 means all correctness passed and measurement
completed, even when the speed bar is missed. Exit 1 means compiler differences;
exit 2 means infrastructure or inconsistent artifacts. Performance failure
prints a failed row instead of partial timings as a complete result.

The stand-ins use the existing adapted Node CLI and diagnostic mutant wrappers.
B targets argument.ts (tiny and acceptance) and ArrowFunctionExpression1.ts
(upstream), preserving stderr, length and exit. C exits 1 without output.
The proof requires B/C to have no performance directory, log or command file,
not merely no recorded result. The A performance artifacts must contain ten
real native timing argv per accepted input. Node-forwarder measurements are
plumbing evidence, not native speed claims.

Executed commands and complete observations are recorded in evidence/final/.
The reproducible commands are:

```sh
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-auto-upstream/upstream
stage3/verdict/final.sh --tsc "$PWD/stage3/verdict/standins/node.sh" /tmp/verdict-final-A > /tmp/verdict-final-A.log 2>&1
STAGE3_VERDICT_MUTANT_TARGET=argument.ts,ArrowFunctionExpression1.ts \
  stage3/verdict/final.sh --tsc "$PWD/stage3/verdict/standins/mutant.sh" /tmp/verdict-final-B > /tmp/verdict-final-B.log 2>&1
stage3/verdict/final.sh --tsc "$PWD/stage3/verdict/standins/empty.sh" /tmp/verdict-final-C > /tmp/verdict-final-C.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove_final.py \
  /tmp/verdict-final-A /tmp/verdict-final-B /tmp/verdict-final-C \
  /tmp/new-final-proof > /tmp/final-proof.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s stage3/verdict \
  -p 'test_*.py' > /tmp/verdict-final-tests.log 2>&1
```

All test output went to files. No whole-package tests, repository full gate or
native compiler build was run. Fixtures for orchestration use disposable JSON
artifacts and a process-call trace. The timing-permission mutant changes the
permission result to true on failed artifacts; the trace catches the forbidden
second invocation. Deferred/count mutations, child-exit inconsistency, missing
summary, changed executable bytes and missing native samples are rejected by
their own checks. The performance-preflight fixture is a distinct failure path.
The real A/B/C command runs establish the stand-in behavior beyond those focused
fixtures. No Adamic compiler/oracle fixture was added; its counts table is
unchanged. stage3/verdict/counts.md records these harness fixtures instead.

Setup succeeded: Go 0.019s, Node 0.020s, submodules 0.078s, markdown 0.083s
(dependency step 0.008s), clang 0.193s, Go build 31.426s, tests deferred 31.531s,
cache warm 31.532s, total 31.561s. nproc: 5; cgroup cpu.max: 400000 100000.
A's orchestration parent was paused while its correctness child continued,
then resumed after B completed, preventing overlapping correctness work during
A's performance measurement. C raw captures were archived after completion to
free scratch space. These scheduling/storage steps do not bypass any suite.

Observed full command exits: **A 0, B 1, C 1**. A passes 301/301 acceptance,
1/1 tiny and 13,693/13,693 upstream. B passes 300/301, 0/1 and 13,692/13,693;
one stdout-only byte mutant fails each suite. C passes 0/301, 0/1 and
0/13,693. B/C create no performance directory, log or command file.
The proof script exits 0, and the final focused run reports **34 tests, OK**.

A's complete one-table result (seconds, mean ± sample SD; ten samples per cell):

| Phase / input | Pass | Fail | Native s | Node s | Go single s | Go default s | 1.78 s bar |
|---|---:|---:|---:|---:|---:|---:|---|
| correctness: acceptance | 301 | 0 | | | | | |
| correctness: tiny | 1 | 0 | | | | | |
| correctness: baselines | 13693 | 0 | | | | | |
| timing: 001_varianceCantBeStrictWhileStructureIsnt | | | 1.0344 ± 0.0471 | 1.0134 ± 0.0457 | 0.2850 ± 0.0186 | 0.2864 ± 0.0329 |  |
| timing: 024_genericTypeParameterEquivalence2 | | | 0.9986 ± 0.0355 | 1.0158 ± 0.0342 | 0.2780 ± 0.0128 | 0.2779 ± 0.0141 |  |
| timing: 056_genericCallInferenceInConditionalTypes1 | | | 1.0690 ± 0.1161 | 1.0117 ± 0.0582 | 0.2905 ± 0.0333 | 0.2742 ± 0.0278 |  |
| timing: mitt | | | 0.3827 ± 0.0185 | 0.3770 ± 0.0312 | 0.0545 ± 0.0059 | 0.0572 ± 0.0046 |  |
| timing: zod | | | 4.3152 ± 0.1434 | 4.2964 ± 0.1689 | 1.2729 ± 0.0579 | 0.9882 ± 0.0694 |  |
| timing: typescript-compiler | | | 9.1687 ± 0.4466 | 8.9089 ± 0.2447 | 2.9664 ± 0.1193 | 1.9559 ± 0.1169 | missed |

The native slot is deliberately a Node forwarder. Its compiler mean is
9.1687 ± 0.4466 s; fresh Node 8.9089 ± 0.2447 s, Go single 2.9664 ± 0.1193 s,
Go default 1.9559 ± 0.1169 s. The fixed 1.78 s bar is missed. These are shared
box observations, not a claim about Adamic native speed or a stable slowdown.
The six accepted workloads matched every preflight, warmup, timed and RSS
invocation. The const-enum workload was excluded only for Go's exit-status
difference; native matched Node on it too.

Two additional mutants changed copies of real recorded artifact contents in
place, with exact-byte restoration in finally blocks. A second byte in B's
raw ArrowFunction stdout is caught by the raw-byte witness check while its
projected capture remains unchanged. A forged native compiler mean of 1.0 s,
leaving its ten actual samples untouched, is caught by sample-statistics
validation. Both assertions pass; artifact-mutants.log.gz retains the output.
The production A/B/C proof is rerun against the restored artifacts before
packaging. Neither mutation runs or changes a compiler.
