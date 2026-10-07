Built: supplied-SSA static-components kernel in .a plus Go comparison controls; all three full React ports remain blocked.
Commits: previous six complete ports pushed through 7f4eabb8; third claim d0c89a9d and blocker evidence 5b1590fe were already pushed.
Commands: fresh setup 19s, nproc 5; supplied-graph comparison PASS 34 cases, 27 findings, 14419 identical bytes; ASan/UBSan/leaks PASS.
Mutants: store-binding and phi propagation mutants both compile and exit 0; byte comparison catches each.
Not covered: native source-to-SSA, captures, post-dominance and compilation gates; full three-rule corpus agreement, rule mutants, handles and lint timings remain unfinished.

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

## Follow-up: supplied-graph kernel and dependency recheck

The preceding blocker observations describe the earlier branch state. Origin
was refreshed again across all heads (417 refs). Current main is
`e011f8f60899586d6373a5ccb07335ad82cfbf3c`, bridge
`eb6df00e91b07c9fc81b2ec396a6f118be55d172`, and harness
`f4d98cab50048692781da3599131317dc569d466`.

JSX parser work is now published on `origin/codex/stage1-jsx-lint`, tip
`a8a62d62ca49db7415e14c3887dd305022b17309`. Its JSX_REPORT.md was read
in full. Published parser/scanner directories were materialized in scratch
`/workspace/wave29-jsx-dependency`, and the static-components probe was compiled
with its import pointing there. It exits 0 and emits `JsxSelfClosingElement`,
resolving the formerly failing `<C/>` parse in that scratch experiment.
The shared parser was not changed on this branch. Its logged output is
`validation/recheck-jsx.stdout`.

An all-origin source search for `LoadContext`, `ControlDominators`,
`UnconditionalBlocks`, and `ForFunctionWithoutManualMemoization` found only
wave08-react's supplied-HIR effect kernel and its core probe. That kernel was
read in full; its run entry explicitly rejects the absent lowering/SSA pipeline.
This search supports, but does not prove exhaustively, the inference that no
published usable native React lowering entry exists. The Go lowering, capture
translation, manual-memoization handling, and control analysis are still needed.

`static_components.a` now implements the independent forward-taint validator
on supplied graphs. It follows the unchanged Go `reportDynamicComponents`:
function/call/new/method taint creation, local loads and both store outputs,
phi propagation, ordered block traversal, JSX tag filtering, creator-specific
messages, missing-node fallback ranges, and UTF-16-to-byte diagnostics. It
imports no Go lint judgment and adds no bridge predicate. `static_core.a`
is its framed-graph executable, not a source-file lint entry.

The overlay test in `testdata/static_core_test.go` invokes the unchanged private
production Go validator directly. The Go test constructs the HIR and exports
its structural fields for native evaluation, along with the expected complete
diagnostics. No production Go file is replaced. These are supplied-graph
controls, not lowering or compiler-corpus agreement. Each phi control has at
most one tainted operand: Go ranges an unordered map when choosing a creator,
so different tainted creators at the same merge are not covered by this test.

Run with the configured Go/clang toolchain:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave-29-third/check_static_core.py \
  /workspace/wave29-third-static-core-verified \
  > /tmp/wave29-third-static-core-verified.log 2>&1
```

Observed: 34 supplied graphs, 27 findings, 14,419 identical bytes. All findings
have zero fixes and zero suggestions, matching Go. Byte SHA-256 is
`78ce14fae61fcede8e45b0fcc1137e1896855818af4a8770c55bf7b9c8d9fe31`.
The native run exits 0 with empty stderr; its ASan/UBSan/leak-check build produces
identical bytes and empty stderr with halt-on-error options. Two mutations
replace the creator propagated to a store binding or phi with that binding's
own identifier. Both compile and execute normally; comparison catches each
as a differing diagnostic. They prove the graph-kernel checks can fail, not
that an end-to-end rule comparison exists. This kernel uses no bridge handles,
so released-handle checks for a new full rule are not claimed.

Fresh setup output: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready
1s, build cache warm 19s, total 19s; nproc 5, quota 400000/100000 and 17.6 GB.
Only the supplied-graph overlay test and native kernel checks were run in this
follow-up; the previous full production rule-suite result remains above.
The native framed invocation took 2.865 ms; Go test elapsed time includes build
and cannot serve as a native-versus-Go lint comparison. Full lint timings remain
unavailable until lowering is implemented.

Commands, controls, Go/native bytes, empty sanitizer stderr, and mutant output
are stored in `validation/static-core/`. No shared files were edited, no new
rules were claimed, and the three full rules remain blocked rather than marked
complete. Set-state-in-effect and set-state-in-render have no new native
validator here; their blocker controls remain the executable evidence.
