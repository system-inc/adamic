Built: a pushed reservation and dependency audit for the next three rules; no new rule implementation.
Commits: prior implementation/evidence pushed at 5ebe19e4; new reservation c96793ae.
Checks: fetched 389 origin refs; audited 33 claim documents and 197 ranked rules; setup PASS 29s, nproc 5; diff check clean.
Mutants: none for this reservation, because native rule code was not written; prior successful mutants remain in WAVE_15_FOURTH_REPORT.md.
Not covered: all three new ports, byte/fix/suggestion comparisons, mutants, sanitizers and native/Go timings are blocked by missing shared native HIR.

The first three remaining names in descending combined volume with lexical
ties, excluding existing ports and all origin claims, are:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Each has combined volume zero. Thirty-three claim documents mention 142 ranked
rules; the existing bridge has 25 checker-dependent ports plus the non-checker
method-signature-style port. Matches for the selected names on origin/main and
origin/codex/tsgo-c-library are inventory/count records only. The reservation
was pushed before any implementation code. The complete ref snapshot and
remaining ranking are in validation-wave-15-fifth.

Prior code and evidence are already pushed. Existing partial regex rules were
checked against the freshly fetched helper branches: no authored .a or .ts
implementation of RegexSyntax, ConstantStringIn or SkipPatternEscape was found.
Their documented blockers remain. The shared harness branch alone cannot fill
these missing semantic helpers.

The new three are blocked by a different dependency. Production Go cohere does
not decide them from a checker type or a simple syntax walk. It consumes the
shared high_level_intermediate_representation package, including lowered
functions, instructions, value identities, captures, phi operands and control
flow. Searches found no native implementation on this branch or the six fetched
lint-helper branches. The production dependency lines and exact search commands
are retained in validation-wave-15-fifth. The existing compiler's Go IR is not
an Adamic module the native lint rules can consume.

Specific prerequisites:

- set-state-in-effect: MayHoldComponentOrHook, ForFunctionWithoutManualMemoization,
  AsCompilationUnit, memoization erasure/inlining, capture translation and
  ControlDominators for the ref-derived control exemption.
- set-state-in-render: compiled-function HIR, capture/setter propagation and
  UnconditionalBlocks to reproduce the unconditional render-path judgment.
- static-components: ForFunction, AsCompilationUnit, function/JSX compilation
  gates, reverse-postorder phi/instruction taint propagation and identifier-node
  source ranges. Actual JSX parsing is also still absent on this branch.

These are shared native implementation prerequisites, rather than the .a loader,
profile compilation or suggestion serialization gaps named in the latest
request. Adding a checker question cannot provide native HIR lowering or its
analysis passes. Calling production Go lint decisions through the bridge would
not constitute a native Adamic port. An always-quiet implementation would not
prove the rules on positive inputs.

Ahra's earlier correction says: "If anything else blocks you, say exactly what
it is and stop, rather than editing shared files." That instruction applies
here. No shared HIR, parser, registration generator, existing test harness or
protected compiler files were edited. No new rule files, checker questions,
mutants or passing rule comparisons are claimed for this reservation. Work stops
at this dependency boundary; no later rules were selected to bypass the ranking.

Commands and observed results:

```
git fetch --recurse-submodules=no origin '+refs/heads/*:refs/remotes/origin/*'
bash cloud/setup.sh > /workspace/wave15-fifth-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
git diff --check
```

Fetch succeeded, nproc printed 5, and diff check was empty. Setup timing:
Go 1.27.1 ready 0s; clang 20.1.8 ready 0s; Node v24.19.0 ready 0s;
submodules ready 0s; build cache warm 29s; setup done 29s.
Cgroup quota is 400000/100000, four cores. No code tests were rerun for this
documentation-only blocked reservation. Prior rule comparisons and timings remain
in the earlier reports; they do not validate these three rules.

Recheck after the next keep-cooking request, 2026-10-07:

Fetched all heads again, now 417 origin refs. The all-origin .a/.ts search for
ForFunctionWithoutManualMemoization, ControlDominators, controlDominators and
high_level_intermediate_representation returned no matches (git grep exit 1).
The pinned refs and exact patterns are retained alongside this report. The
current branch still contains no native HIR modules. No additional claims were
made because these three are unfinished.

The freshly fetched wave-29-third report independently names the same three
rules and the same missing React SSA, capture, post-dominator and memoization
substrate. Its .a files under gaps/ are parser probes, not rule implementations.
Its report additionally records hook-only positive controls, establishing that
JSX parsing alone would not unblock the missing analysis. This is that worker's
reported evidence, not a native comparison newly run by this worker. A copy of
the pinned report is retained for provenance.

The unchanged production Go suites were rerun successfully:

```
source /workspace/adamic-tools/env.sh
(cd cohere && go test ./internal/lint/rules/react -run '^Test(SetStateInEffect|SetStateInRender|StaticComponents)' -count=1 -timeout=10m -v) > /workspace/wave15-hir-recheck-go.log 2>&1
```

PASS, package 0.084s. This validates the production baseline only. No native
ports, byte comparisons, native mutants, sanitizer runs or native timings were
added; those remain blocked by the missing shared HIR. Toolchain setup was not
rerun for this dependency recheck; the existing environment was sourced. Work
stops under Ahra's shared-file boundary instruction.
