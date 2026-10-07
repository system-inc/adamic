Built: no new ports; the fifth batch is blocked on the absent native React HIR substrate, with JSX additionally blocked.
Commits: claim b8be942f was pushed before implementation; this report is a separate evidence commit.
Checks: fresh origin audit, cloud setup PASS 37s, pinned Go reference tests PASS 0.130s (38 top-level tests), native JSX probe exits 70.
Mutants: none for this batch; no native implementation exists to validate, and Go-only tests are not byte agreement.
Uncovered: all three claimed ports, native/Go finding/fix/suggestion agreement, sanitizer/released-handle gates and timings for this batch.

## Claimed rules and state

The first twelve wave 12 implementations and their evidence were pushed before this continuation. Their latest implementation/evidence commits are 974dc275/acb1ff7c. The remote wave branch was verified at acb1ff7c before this claim.

Fetched all origin heads with git fetch origin '+refs/heads/*:refs/remotes/origin/*'. The combined VOLUME_REPORT compiler-all.counts and repository-all.counts contain 197 checker-dependent names, sorted by descending combined volume and lexical name. Audit: 389 origin refs, 142 claimed ranking names, 25 checker-dependent ports on main/the bridge branch, 30 remaining names. Claim bodies under stage1/cohere/typeaware/claims on every fetched origin ref were read; the first three available names were:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

All three have zero counts in both recorded corpora. Exact matching rule module filenames across origin refs found no existing ports. Claim b8be942f was pushed before any implementation. These remain claimed and UNFINISHED; this report does not release them or count them as ports. No further rules were claimed.

## Exact blockers

All three production rules consume cohere's React high_level_intermediate_representation rather than a local AST predicate. The Go package contains 134 Go files. No native HIR implementation was found in this working branch, current origin/main or origin/codex/lint-harness-dot-a. A path audit of all 389 origin refs found no stage1 .a/.ts path containing /hir/, high_level_intermediate_representation, postdominator, react_compiler or /ssa.; that is a path observation, not proof about every implementation under arbitrary filenames.

| Rule | Required substrate missing from the native rule layer |
| --- | --- |
| set-state-in-effect | ForFunctionWithoutManualMemoization; erasure/inlining; per-function SSA identities; capture-to-context translation; ref-derived phi propagation; ControlDominators |
| set-state-in-render | ForFunction; AsCompilationUnit; capture translation; UnconditionalBlocks computed from post-dominators; useMemo callback handling |
| static-components | ForFunction and AsCompilationUnit; reverse-postorder SSA/phi taint; JSX tag places with source identities |

The current native parser also has no JSX productions. Every positive static-components finding requires JSX, so a positive native comparison for that rule is blocked independently of HIR. The existing source-has-jsx bridge question detects syntax and refuses it; it supplies no lowering or graph.

The shared harness branch's .a loading, profile compilation and suggestion serialization do not supply HIR. The existing compiler's internal/flow is Go over Adamic's compiler IR, not native code over a lint source AST with React lowering semantics. Reusing it would not provide the graph these validators consume.

An AST shortcut cannot be certified as equivalent: the production controls cover conditional assignments joined by phis, ref-controlled setter exemptions, transitive captured setter calls and manual memoization inlining. Asking Go cohere's HIR to perform that analysis through the bridge would move the substantive rule work into Go, rather than keep native Adamic responsible for it. A checker question or one-line registration cannot fill this substrate gap.

Your earlier correction explicitly says: “If anything else blocks you, say exactly what it is and stop, rather than editing shared files.” This is the absent analysis-substrate blocker, beyond the shared harness gap described in the continuation. I stopped here, preserved the claims and did not edit shared files. No zero-finding placeholder modules were added, and no Go-only pass is represented as a native port.

## Commands and observations

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*' > /tmp/wave-12-fetch-fifth.log 2>&1
bash cloud/setup.sh > /tmp/wave-12-setup-fifth.log 2>&1
source /workspace/adamic-tools/env.sh
# From /workspace/adamic/cohere:
go test -count=1 -v ./internal/lint/rules/react \
  -run '^(TestSetStateInEffect|TestSetStateInRender|TestStaticComponents)' \
  > /tmp/wave-12-fifth-go-controls.log 2>&1
# Existing fourth-batch native binary and preserved JSX input:
/workspace/wave-12/fourth/attempt7/native \
  /workspace/wave-12/fourth/attempt7/tsconfig.json \
  /workspace/wave-12/fourth/attempt7/jsx.manifest \
  > /tmp/wave-12-fifth-jsx-probe.stdout 2> /tmp/wave-12-fifth-jsx-probe.stderr
```

Setup timings: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm 37s, total 37s. nproc: 5; cgroup quota: four cores; memory 17.6GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Pinned production Go test package: PASS 0.130s; 38 matching top-level tests ran, with positive and silent subcases. This establishes that the reference rules and their positive controls run here. It does not establish native agreement.

The native JSX probe exited 70 with exactly adamic: panic: wave 12 shared parser does not support JSX. Its stdout is empty. This reruns the already preserved fourth-batch probe, not a new rule implementation.

Compressed fetch/setup/test/probe and claim-push logs are in evidence/. No test output was piped. There are no new native ports, per-rule mutants, timing measurements or sanitizer/released-handle results for this batch. The preceding batch's results remain separate in wave12_fourth/REPORT.md.

## Work required to resume

Provide or port native React HIR lowering with SSA, source identities, nested capture mappings, memoization preparation and post-dominator/control-dependence analyses; provide native JSX parsing for static-components. Then implement these three validators in their own .a modules and run independent pinned production Go byte comparisons, meaningful mutants and normal/sanitized corpora. Until that exists, these three claims remain blocked and must not be reported as completed.
