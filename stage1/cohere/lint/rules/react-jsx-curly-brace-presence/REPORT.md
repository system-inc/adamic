# react/jsx-curly-brace-presence

The port calls the unchanged Go rule at cohere
`7945d102a6c18dd36adf9114a758ce646e8b2359` through its typed oracle adapter.
It uses the shared `OptionsJson`, `comments.forFile`, and runtime `utf8Length`
helpers. Options are initialized in the factory after construction.

`evidence/upstream.jsonl` preserves 159 unique captured upstream
source/file/options combinations (307 capture records before deduplication).
`evidence/selected.log` records successful upstream/inherited corpus comparison,
owned witnesses, and the compiling mutant across source Node, emitted
JavaScript, and ASan/UBSan native. The combined syntax corpus comparison held
13,815,695 identical bytes. All five witnesses cause Go findings; option
sidecars exercise shorthand always, element props, and declined fixes.

For inherited all-rule settings, the Go adapter follows
`../no-underscore-dangle/oracle.go` and
`../no-unsafe-optional-chaining/oracle.go`: decode into the typed options value
with `json.Unmarshal`, leaving unknown inherited keys alone. Own witness wire
options go through upstream `DecodeJsxCurlyBracePresenceOptions`, which decides
strictness and shorthand/default behavior. Captured upstream rows already carry
Go's decoded struct shape. Disabled listeners do not decode another rule's bag.

Go behavior preserved even when surprising:

- Corpus I35: a quote in a string child reports, while the same quote in a
  no-substitution template declines.
- Corpus I44/I45: with `propElementValues: "never"`, element-valued props can
  oscillate between braces and bare markup under repeated fixing.
- Multiline JSX text: the diagnostic starts at the token range, while its fix
  replaces the full text span including leading indentation.

The `empty-string-is-whitespace` mutant compiles and runs on all three sides.
It wrongly treats an empty attribute string as whitespace, suppressing the
finding on `testdata/empty.tsx.txt`. The caught lines are in
`evidence/selected.log` (Node, emitted JavaScript, and native).

## Whole-package run

`evidence/whole.log` and `evidence/whole-summary.json` record the single complete
package run with every external input supplied: clean TypeScript checkout at
`050880ce59e30b356b686bd3144efe24f875ebc8`, WASI SDK 27, benchmarks enabled, and
one fresh directory shared by both profile variables. Counts include named
Go test events (top-level tests and subtests): **112 pass, 2 fail, 1 skip**.
Wall time was **1157.958 seconds**; `nproc` was **5**. Load before was
`0.89 1.68 1.52`, and load after was `1.71 2.42 2.10`.

The rule is certified against the required oracle comparisons; the complete
package is not green because of these shared assertions:

- `TestJsxLintReleaseAndThroughput` and `TestJsxLintTrees` both fail at the frozen
  map in `../../jsx_integration_test.go:51` (asserted at line 61). They expect the
  prior five-rule JSX inventory and reject the additional 159 cases for this
  registered rule. No finding, range, fix, or runtime comparison caused these
  failures. Updating that shared test is outside this rule-directory-only scope.
- `TestCheckerBridgeRefusalPending` skips at `../../checker_pending_test.go:49`
  because the base lacks the `TSGoError` bridge declaration. This is unrelated
  to missing external inputs and to this rule.

The 874-file compiler/stage1 comparison passed on Go, source Node, emitted
JavaScript and sanitized native, holding 30,500,309 identical bytes. The full
upstream/inherited comparison, all owned witnesses, every compiling rule mutant,
shards, non-JSX throughput, and all three profile tests passed. This rule's mutant
was caught on all three runtimes in both selected and whole-package runs.

No rule algorithm or shared-helper blocker remains. `BLOCKED.md` was removed.
Only this rule directory changed. Log line-end whitespace is normalized for git;
the harness's comparisons themselves used the original bytes.
