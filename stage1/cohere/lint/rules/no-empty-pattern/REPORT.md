Ported `no-empty-pattern` from batch 2 onto the unified registry harness.
Tested area base: `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`. Batch source tip: `6700c572ee3810524e356af7f92973e38ee5265f`.
Pinned cohere oracle: `715ba94f3608a6500086b1076ce5cb7e51b836db`; upstream implementation unchanged.

The descriptor uses `node: true`, named ObjectBindingPattern/ArrayBindingPattern
subscriptions, and a handed-node visitor. No relevance filtering or shared
registration lines are added. The object/array distinction selects the message
and the narrow parameter exemption. Both messages were moved verbatim to
messages.a. New Adamic modules are .a; the Go adapter decodes manifest field 5
into upstream NoEmptyPatternOptions and returns its strict zero defaults when
options are absent. No shared harness, helper, parser or checker edits were needed.

The TestNoEmptyPattern prefix covers every real upstream test:

- `TestNoEmptyPatternReportsEmptyPatterns`
- `TestNoEmptyPatternStaysSilentOnRealBindings`
- `TestNoEmptyPatternWithoutOptionsIsStrict`
- `TestNoEmptyPatternAllowsParameterObjectPatternsWhenConfigured`
- `TestNoEmptyPatternKeepsReportingDespiteOption`

Those five tests contain 45 source/options cases, including nil options, strict
empty object and array patterns, nested patterns, real bindings, allowed direct
object parameters, empty-object defaults, non-empty defaults and methods.
The captured shared corpus reports 2,062 unique source/rule/options combinations.
The owned witness fires under defaults and includes both message kinds plus real
bindings. It does not require the pending witness-options companion feature.

Validation:

- lint-registry: PASS, 41 descriptors, including this rule.
- gofmt: empty output; owned adapter formatted.
- go vet ./...: PASS, empty output.
- TestRulesAgree: PASS, 13,094,247 bytes identical on Go, source Node,
  emitted JavaScript and ASan/UBSan native.
- TestMutants: PASS, all 41 registered mutants caught on Node, emitted
  JavaScript and sanitized native.
- TestOwnedWitnesses: PASS, 97,874 identical bytes in selected-rule and
  all-rule modes; each witness causes upstream findings.
- Combined requested test command: PASS, 1379.913 s; zero test skips.

The owned mutant replaces `node.children.length !== 0` with
`node.children.length < 0`, so real bindings incorrectly report. It compiles
and exits normally; all three executions differ from the independent Go oracle
at witness case 145, line 2746 (outer binding at column 7 versus nested empty
pattern at column 17). Only diagnostic comparison catches it. The full log
records every catcher and all registered mutant assertions.

The shared upstream corpus has existing malformed method-signature cases whose
parser recovery is explicitly unsupported. The gate verifies refusals for those
cases instead of claiming finding parity; no no-empty-pattern input hit a parser,
type-checker, dynamic-RegExp or Tailwind blocker. No recovery condition, option
guard or test assertion was edited. The full repository gate and external-input
stage 1 comparisons were not run by this focused lint unit.

Setup: Go, clang, Node and submodules 0 s each; cache warm 89 s, total 89 s.
nproc=5, quota=4 cores, memory=17.6 GB.

Exact commands are in evidence/commands.txt. Test, registry, formatting, vet,
setup and source-fetch logs are preserved as deterministic gzip files in evidence/.
Only this rule directory is committed.
