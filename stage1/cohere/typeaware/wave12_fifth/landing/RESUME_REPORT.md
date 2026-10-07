Built: landing readiness verified against freshly fetched main; no new ports or claims.
Commits: validated and published tip 413559587fa66666013f505a38ba25146f2b4a47; this continuation adds evidence only.
Checks: current main remains e8ba3d5d81de4d3773c723914fccd4c76248b965; existing complete native gate remains applicable; blocked-rule Go reference controls PASS 0.137s.
Mutants: no new mutants; twelve rule, five raw-question, four lifetime, JSX and Node catches remain recorded in REPORT.md.
Uncovered: three claimed source ports remain blocked on native React source lowering/SSA/captures/memo preparation, plus JSX for static-components.

Fetched every origin head: 466 refs. Remote wave 12 matches the previously validated tip, and current main is its ancestor. No source or compiler changes require another rebase or repeat of the unchanged native gate. Never pushed main or an area branch.

The previous blanket HIR-path observation is now superseded: origin/codex/typeaware-wave-21 at e9ec024c3213ec7c9ad967d39029e4dd004abb9a contains wave21_react/hir.a, postdominator.a and three validator cores. Read those modules and WAVE_21_REACT_CORE_REPORT.md. The model explicitly says producing the graph from source is a separate adapter. The report explicitly leaves native source Lower/Construct, capture mapping, memo preparation and source integration unimplemented. Its passing comparisons consume Go-prepared HIR; they do not establish native source ports. These existing cores reduce future validator work but do not remove the missing source frontend. Main and the shared harness branch still show no JsxElement/JsxSelfClosingElement parser matches.

Wave 21 also reserves these same three rules. No new reservation was made here. Integration should reconcile the overlapping claims before duplicating its validator cores. Its report records two distinct production Go outputs across 24 identical two-creator phi runs; deterministic byte parity for that case is another explicit outstanding issue, not an observation independently reproduced here.

Reference-only command, with output redirected directly to a log:
source /workspace/adamic-tools/env.sh
go test -count=1 -v ./internal/lint/rules/react -run '^(TestSetStateInEffect|TestSetStateInRender|TestStaticComponents)'
PASS 0.137s. This is not native agreement. Required setup PASS: Go/clang/Node/submodules 0s, cache and total 128s, nproc 5, four-core quota, 17.6 GB.

The earlier instruction to stop on other blockers still applies. Claims remain unfinished, no placeholders or shared edits were added, and no additional rules were claimed. Full repository gate and inherited 26-rule regression remain unrun. Previous native/Go timing and complete mutant evidence remain in REPORT.md.
