Built: reporting portions of three React rules, with explicit source-analysis refusal; earlier twelve ports remain complete.
Commits: earlier completion d0dc8c2c; claim 3c3f5f4b53fb063461f38d8051add1f056a16527 pushed before this partial implementation.
Checks: four actual production Go findings match native reporting in 2355 bytes, normal and ASan/UBSan; JSX parser refusal reproduced.
Mutants: three diagnostic-ID mutations caught by Go bytes; three removed-analysis-refusal mutations caught by the required panic exit 70.
Not covered: native source analysis, corpus parity, full-rule mutants, checker ownership and comparable native/Go lint timing for these three.

The claims are `react-hooks/set-state-in-effect`, `react-hooks/set-state-in-render` and `react-hooks/static-components`. They remain unfinished and reserved. No additional rules were claimed. Selection fetched all origin heads and inspected 389 refs, 33 unique claim documents, 142 claimed ranked rules and 25 ranked baseline/main ports. These were the first three remaining entries in the combined volume ranking, descending count with lexical ties, each with zero recorded corpus volume.

Each owned `.a` file implements the reporting portion, including both render-state messages, the static-component creation-site message and canonical zero-fix/zero-suggestion fields. Calling `analyze()` refuses with `NotYet` before any finding output. This is deliberately not a functioning source-analysis port, and it cannot silently pretend to have analyzed a source file. No shared harness, registration, parser, bridge or compiler file changed.

The primary blocker is the missing native React HIR substrate. The production Go rules consume `cohere/internal/lint/ecmascript/high_level_intermediate_representation`:

- Set-state-in-effect requires manual-memoization erasure and inlining, typed SSA places, phi values, setter propagation across captures, ref-derived value taint and control dominance. A nested function inside an effect behaves differently from one outside it, and translating identifier spaces across a capture is essential to the verdict and span.
- Set-state-in-render requires SSA setter propagation, manual-memoization inlining, recognition of memo regions and post-dominance over the lowered graph. A lexical parent test cannot distinguish early returns, loop exits and unconditional render calls correctly.
- Static-components requires React compilation-unit selection, JSX lowering, one reverse-postorder pass over SSA values and phis, and creation-site provenance. Its outermost-unit and loop-back-edge behavior must be preserved rather than approximated.

The current native syntax CFG helpers supply edges, not those typed value/capture spaces or React lowering passes. Scoped searches for the required native substrate return exit 1 with empty output; commands and output are preserved in `validation/`. Other origin workers also document this blocker, including `origin/codex/typeaware-wave-04:stage1/cohere/typeaware/wave_04_react/REPORT.md` and the wave-08 React report. Origin contains isolated JSX parser work, including wave 16, but that does not supply the missing React HIR pipeline.

There is also a directly reproduced parser gap on this branch. The owned probe feeds the unchanged shared native parser:

```
function Component(){const Inner=()=>null;return <Inner/>;}
```

It exits 70 with:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 55 in fixture.tsx
```

The Go production static-components rule reports a positive finding on the corresponding TSX control. Thus a zero-finding compiler/repository comparison would not demonstrate a port. No such comparison is claimed for these three. Sending Go HIR validators or lint findings through a checker question would also violate the native-judgment boundary; no shortcut was added.

Ahra instructed: "If anything else blocks you, say exactly what it is and stop, rather than editing shared files." Work stops with these dependencies recorded and the portable reporting portion pushed. This is a substrate blocker, not an automatic approval rejection or a permission request.

The owned Go oracle loads four generated controls and runs the unmodified production rules. They produce one effect finding, one render finding, one memo-render finding and one static-component finding. The native reporting probe receives their span coordinates; it renders its own messages and fields rather than copying the Go message bytes. The creation spelling is the known fixture expression `() => null`. Complete streams match, retaining full record fields. This proves reporting only, not native range discovery or verdicts. All four production records have no fixes or suggestions.

Normal and ASan/UBSan reporting match with empty sanitizer stderr. Each reporter ID mutant compiles, exits normally and differs only through the independent Go byte comparison. Each analyzer's unmodified refusal exits 70; removing it exits normally with no output and fails the required refusal contract. These are reporting/refusal mutants, not full-rule mutants. The parser probe separately reproduces the missing JSX path.

Reproduce from the repository:

```
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_react_state/validate_partial.py --scratch /workspace/wave-06-react-final > /tmp/wave-06-react-final.log 2>&1
```

All output goes directly to files. Complete compressed stdout/stderr, commands, exits and elapsed times, input sources as JSON, origin refs and hashes are committed under `validation/`. Elapsed reporting time and Go loading/analysis time measure different work and are not presented as a native/Go lint benchmark.

The toolchain was reused from the original successful 77-second setup: nproc 5, CPU quota 4, Go 1.27.1, clang 20.1.8 and Node 24.19.0. Cohere remains pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`; no pins changed. The new oracle driver was formatted, Python source compiled successfully and `git diff --check` passed. No root test gate or released-checker test was run for this reporting-only work, which makes no checker bridge calls.

The later [prepared-HIR continuation](CORE_REPORT.md) adds native validator cores
and post-dominance reused from wave 21, with semantic mutants and Go/native/Node
comparisons. This report describes the earlier reporting-only stage. Current
source analyses remain unfinished because the native source-to-HIR adapter,
source unit gates and memo-scope resolution are still absent.
