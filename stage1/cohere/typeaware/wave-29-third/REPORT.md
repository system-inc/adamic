Built: executable blocker controls and an independent Go oracle for the three newly claimed React rules; no native rule port is marked complete.
Commits: previous six ports are pushed through 7f4eabb8; d0c89a9d claimed this batch and was pushed before any implementation.
Commands: setup 29s, nproc 5; production Go suites PASS 0.082s; Go positive controls produce three JSX findings and two hook-only findings.
Mutants: no new native rule mutants were run because the native validators are not implemented; blocker probes are not counted as lint agreement.
Not covered: native React SSA lowering, JSX parsing, end-to-end findings/fixes/suggestions, sanitizer rule agreement or native-versus-Go lint timings for this batch.

The third batch is:

1. `react-hooks/set-state-in-effect`.
2. `react-hooks/set-state-in-render`.
3. `react-hooks/static-components`.

All origin heads were refreshed with an explicit all-heads fetch. The scan
combined the 197-rule count tables linked by VOLUME_REPORT.md, sorted descending
by combined compiler/repository volume and lexically for ties, excluded the 26
baseline ports, and examined claim Markdown blobs on every origin ref.
There were 389 origin refs, 33 unique claim blobs, 141 additional claimed names
and 30 available names. These were the first three available names, each with
zero default volume. Claim d0c89a9d was pushed before the blocker work.
The previous implementation commit 7f4eabb8 is an ancestor of the origin branch.

The production Go sources for all three rules were read in full. They do not
judge these inputs directly from syntax. Their dependencies include:

- Set-state-in-effect: `ForFunctionWithoutManualMemoization`, SSA capture
  translation, manual memoization erasure and inlining, ref value taint, and
  `ControlDominators` for ref-controlled branches.
- Set-state-in-render: `ForFunction`, SSA local/context propagation, capture
  translation and `UnconditionalBlocks` for post-dominance. A syntactic ancestor
  walk cannot distinguish its early-return and loop cases.
- Static-components: `ForFunction`, SSA phis and forward dynamic-value taint,
  with a JSX tag as the reporting site.

The Go HIR shelf is
`cohere/internal/lint/ecmascript/high_level_intermediate_representation`.
The current branch has no Adamic entry for that React graph pipeline.
`stage1/cohere/typeaware/flow.ts` handles type/value relationships; it is not this
SSA graph, function/capture arena or post-dominator substrate. The refreshed
main does not provide that pipeline either. The shared harness branch adds
module/profile/comparison support, not the missing native React graph.
The examined origin tips were main e011f8f60899586d6373a5ccb07335ad82cfbf3c,
bridge 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 and harness
f4d98cab50048692781da3599131317dc569d466.

There is an independently executable shared parser blocker as well. Each `.a`
probe passes the exact source text and `input.tsx` to the native Parser. The Go
oracle independently loads the equivalent reference TypeScript input, calls all
three unchanged production rules and serializes complete diagnostics:

| Input | Go | Native parser observation |
| --- | --- | --- |
| effect with paired `<div>` | one setStateInEffect finding | exits 0, emits TypeAssertionExpression and no JSX nodes |
| render with paired `<div>` | one setStateInRender finding | exits 0, emits TypeAssertionExpression and no JSX nodes |
| dynamically created `<C/>` | one staticComponents finding | exits 70 at slash, expecting GreaterThanToken |

The paired-tag result is a misclassification, not JSX support. Inspecting kinds
caught it; checking only process success would have treated it as a passing
parse. Go reports spans 101:105, 77:81 and 82:83 respectively, with zero fixes
and zero suggestions. The self-closing probe's exact error is:

```text
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 83 in input.tsx
```

Two additional positive controls contain no JSX at all. `useThing` calls a
Dispatch-typed setter inside an effect or during render. Go reports one finding
for each and the native parser succeeds on each. They establish that merely
adding JSX parsing would leave a separate native SSA dependency outstanding.
Reference `.tsx` inputs are generated in scratch for Go; all executable Adamic
probes are `.a`.

The unchanged production suites pass:

```sh
source /workspace/adamic-tools/env.sh
(cd cohere && go test ./internal/lint/rules/react \
  -run '^Test(SetStateInEffect|SetStateInRender|StaticComponents)' \
  -count=1 -timeout=10m -v) > /tmp/wave29-third-production-tests.log 2>&1
# PASS; package 0.082s.
python3 stage1/cohere/typeaware/wave-29-third/prove_gaps.py \
  /workspace/wave29-third-gaps > /tmp/wave29-third-gaps.log 2>&1
# Exit 0; verifies the independent Go findings and each blocker observation.
```

Setup succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready
1s, build cache warm 29s, done 29s. `nproc` is 5, with cgroup quota
`400000 100000` and 17.6 GB memory. All tests wrote to logs without pipelines.
The reproduction uses the previously built stage-0 compiler at
`/workspace/wave29-regex-controls/adamic`. All build/run commands and diagnostic
streams are preserved under `validation/` and `/workspace/wave29-third-gaps`.

Stopped at this dependency boundary under the instruction, "If anything else
blocks you, say exactly what it is and stop, rather than editing shared files."
No shared parser, harness, registration generator, bridge dispatcher, compiler
or submodule was edited. No Go lint judgment was moved into the bridge and no
no-op rule stub was presented as a port. The three claims remain owned and
blocked, and no additional rules were claimed. Native rule mutants, byte
agreement, sanitizer checks and lint performance measurements remain unfinished
for this batch; zero default corpus counts do not satisfy those requirements.
