Built: diagnostic-reporting portions of the three outstanding React rules, each in its own .a file; source analysis explicitly refuses.
Commits: based on current main e8ba3d5d; previous fifteen completed ports were re-greened and pushed at f4bd9c34.
Checks: five real Go findings and one production-message-helper record match in 3563 bytes across native, source Node, emitted JavaScript and ASan/UBSan/LSan; vet and formatting pass.
Mutants: three reporter-ID mutants and one empty-creator branch mutant fail only Go bytes; three removed-analysis-refusal mutants fail the required panic-70 contract.
Not covered: all three native source analyses, full-rule mutants, native range discovery, corpus agreement, released checker handles for new code or comparable native/Go lint time.

Current progress: standalone native source validators now live in each rule.a.
See SOURCE_REPORT.md for their production Go comparisons and current limits.
The remainder records the earlier reporting-only milestone.

The three claimed rules remain unfinished. No new claims were taken. Landing
readiness was checked before adding these files: origin/main still points to
e8ba3d5d, and the prior wave 26 branch is rebased, pushed and green at f4bd9c34.
Only this worker's codex/typeaware-wave-26 branch is pushed. Main and area
branches are left to integration.

The owned reporters are adapted from the reporting portions on
origin/codex/typeaware-wave-06. This reuses existing work transparently rather
than duplicating an attempted HIR analysis. Each reporter produces the canonical
rule ID, message ID, full message, supplied byte range and zero fixes/suggestions.
Set-state-in-render covers both render and useMemo messages. Static-components
covers ordinary creation text, Unicode creation text and the empty-creator branch.
There are no registrations, shared harness edits or compiler changes.

The independent Go oracle runs the three unchanged production rules through
Cohere's registry, with its own compiler, AST walk and checker leases. Five
positive controls produce actual rule findings: effect state, render state,
useMemo state and two static-component creations. The Unicode creation contains
世界🌍, so the comparison checks canonical Unicode escaping too. A sixth fixture
has no actual source finding; an oracle-only overlay calls the unchanged private
production staticComponentsMessage helper with an empty HIR function and no
source file. That separate record tests its exact fallback wording. It is not
claimed as a sixth production finding.

The validator supplies Go-discovered spans and message categories to a generated
native reporting probe. Native code renders its own messages and diagnostic
fields. Static creation text is known fixture input. This proves reporting;
it proves neither native verdicts nor native range discovery. No Go lint verdict
is exported as a checker question or used by an application runner. Calling
analyze() refuses with NotYet and exit 70 before producing findings. The same
refusals run on the original source under Node and match native stderr exactly.

All five streams match complete bytes: Go, native, original .a source on Node,
Adamic-emitted JavaScript on Node and sanitized native. Diagnostic-hashes.json
records their equality; complete streams and command outputs are retained under
validation/. The reporting, Node and sanitizer runs have empty stderr. Each of
four diagnostic mutants compiles, exits 0 and has empty stderr; only its Go byte
comparison fails. The fallback mutant forces the nonempty-creator branch and
changes only the fallback diagnostic. Three mutants remove the analysis panic;
they compile and exit 0 with no output, which violates the required refusal.
These are reporting/refusal mutants, not full-rule mutants.

The initial emitted-JavaScript run used plain Node and failed because its
'adamic' import could not resolve. The validator now uses the existing
oracle/node.mjs adapter for both source and emitted JavaScript. The failure is
preserved separately; the final run passes all 33 commands. No runtime or shared
oracle files were edited to resolve it.

Reproduce:

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace/wave-26-bridge-scratch python3 stage1/cohere/typeaware/wave_26_react_reporting/validate_partial.py --scratch /workspace/wave-26-react-reporting > /tmp/wave-26-react-reporting.log 2>&1
go vet ./... > /tmp/wave-26-react-reporting-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_26_react_reporting/testdata > /tmp/wave-26-react-reporting-format.log
git diff --check
```

All commands pass; vet and formatting logs are empty. Toolchain setup is reused
from validation-wave-26/setup.txt: 135 seconds total, Go ready 0s, clang ready
1s, Node ready 1s, submodules 1s, warm 135s; nproc 5. Source hashes, generated
fixtures, run commands, elapsed times and deterministic compressed outputs are
preserved. New Adamic implementation files are .a; generated TypeScript JSX
fixtures remain .tsx input to the independent Go oracle.

Remaining engineering work is a native React HIR pipeline or an isolated raw-HIR
bridge question and serializer. The rules require SSA phis, closure capture
translation, typed places, compilation-unit selection, manual-memoization erasure
and inlining, and control/post-dominance. The bridge currently provides AST and
checker facts, not that graph. This is not a .a-loading or suggestion-serialization
harness gap. The reporters do not remove that prerequisite. The previous fifteen
ports remain completed; these three are partial and remain reserved.

There are no new native checker handles or bridge questions in this reporting
work, so it does not add a released-handle check. Those checks remain green in
the preceding landing report. No corpus comparison is presented as evidence for
these three analyzers, and no native/Go lint benchmark is quoted: Go performs
source analysis while the native probe only renders supplied records, so comparing
their elapsed times would compare different work. The full root gate was not
rerun for these isolated reporting modules.
