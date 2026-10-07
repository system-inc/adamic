Built native react/jsx-fragments source analysis with three handed-node listeners and existing raw binding facts.
Source commit: 7f646aa90e014a4287718f95683261413717a905; based on current main b8fb957aa and named harness 41eb6eab2.
Checks: 57 controls/34 findings, both frozen corpora, native sanitizers, released handle, formatting, lint and shared JSX tree oracle PASS.
Mutant: React-to-Preact pragma-object comparison compiles, exits 0 with empty stderr, and fails only independent Go finding bytes.
Not covered: shared checker-context registration, own emitted-JavaScript comparison, full repository gate and the constructed-context-values source analysis.

The private typed runner has one kind selection per handed node. The analysis
exposes visitElement, visitSelfClosing and visitFragment; none tests the entry
node's kind or refetches that entry. Child tag/declaration kinds are inspected
for semantic shape. rule.json already declares the named ast.Kind listeners.
No shared dispatch, registry generator or harness source was edited.

The default mode reports a named fragment without attributes; element mode
reports shorthand fragments. The named React.Fragment spelling is structural,
so comments between its parts do not change it. Bare names use the existing
binding-origin question and inspect every declaration, reproducing Go's merged
symbol behavior. Imported aliases require the Fragment export and exact react
module. Variable and object-binding declarations inspect the initializer as Go
does, including its wider acceptance of a bare React initializer and of an
object binding whose property name is not Fragment. A parsed declaration file
is cached once; mismatched checker spans panic explicitly. The foreign .d.ts
witnesses include astral trivia, an uninitialized variable and a positive
React.Fragment initializer. They exercise actual foreign-file parsing.

The independent Go adapter calls the unchanged cohere rule. The gate extracts
45 production table controls plus 12 additional source witnesses, reproducing
both modes. All 57 controls and 34 findings match full wire bytes, including
message text, complete element spans, zero fixes and zero suggestions. Go
withholds the fixer intentionally, including type-argument shapes; this port
reproduces that behavior. Automatic edits and suggestions are not invented.

Both frozen manifests retain the branch's 287 repository roots and 77 pinned
TypeScript compiler roots. Native/Go bytes match under ordinary and
ASan/UBSan/LSan execution: 18,485 repository bytes and 5,241 compiler bytes.
Both corpora have zero findings; positive controls supply semantic coverage.
The released binding-origin question panics 70 with invalid/released-handle
text. No new bridge question or registration line was needed.

Observed native/Go median seconds: repository .365365/.215161 (1.70x),
compiler 2.604120/.521185 (5.00x). The shared JSX tree build overlapped portions
of this gate, so these are whole-process observations, not isolated benchmarks.
No speed improvement is claimed.

The named harness 41eb6eab2 was merged without conflict. It consolidates shared
batch tests and adds the dedup ledger; it does not expose a checker program or
lease in RuleContext. That is the remaining registration boundary for both
jsx-fragments and jsx-no-undef. Their private source analyses are tested, but
they are not complete registered shared-driver modules. The current claims
remain unchanged; jsx-no-constructed-context-values source analysis is still
unfinished, and the three older React compiler claims remain parked. The
constructed-context Go source has also grown memo-dependency and escape/capture
analysis; no incomplete default-only handler is registered as that full rule.

Main remained b8fb957aa at the fetched snapshot, already an ancestor of this
branch. The previous nine-rule landing gate remains recorded in
LANDING_B8FB957AA_REPORT.md; it was not repeated since its source algorithms
and main have not changed. Fresh shared tests: registry PASS .053s; JSX tree
oracle PASS 105.923s, 54 captured sources and 45,527 identical whole-tree bytes.
The full repository/shared lint gates and own emitted JavaScript were not run.

The existing cohere formatter checked eight files for idempotence, including
all three new .a modules; the other five owned modules were unchanged. The
configured cohere lint adapter reported findings 0 on all three new modules.
Toolchain reused: Go 1.27.1, clang 20.1.8, Node 24.19.0. Recorded setup timings:
Go 0s, clang 1s, Node 1s, submodules 1s, cache 121s, total 121s; nproc 5,
quota four cores. No fresh setup timing is asserted. The cloud-environment
runtime skill verified the resumed network policy; no secrets were read.

Commands source /workspace/adamic-tools/env.sh. Final comparison:

```sh
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/fragments-fourth --compiler /workspace/wave20-validation/h41-adamic --checker /workspace/wave20-validation/b8-third/checker.a --sanitized-checker /workspace/wave20-validation/b8-third/checker-asan.a > /tmp/wave20-fragments-fourth.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-h41-registry.log 2>&1
go test ./stage1/cohere/lint -run '^TestJsxLintTrees$' -count=1 -v -timeout 10m > /tmp/wave20-h41-jsx.log 2>&1
/workspace/wave20-validation/listeners/linter /workspace/adamic/tsconfig.json /workspace/wave20-validation/fragments-lint.manifest > /tmp/wave20-fragments-lint.stdout 2> /tmp/wave20-fragments-lint.stderr
```

Initial attempts are preserved. Two constructor builds were correctly refused
before field initialization could be proven; no failing build is counted as a
semantic mutant. The next comparison caught an unintended default Diagnostic
namespace prefix; passing an empty namespace fixed that React wire mismatch.
The final compiling pragma-object mutant is the one counted above. No Go
regular-expression compile site exists in this rule; no matcher was invented.
