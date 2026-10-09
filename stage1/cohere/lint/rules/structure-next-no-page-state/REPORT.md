# structure/next-no-page-state

Ports pinned Go cohere's `structure.NextNoPageState` without changing the upstream algorithm or oracle. The oracle has no options; its typed adapter returns nil. The descriptor captures all structure test prefixes so `TestNoRuleCrashesOnAbsentOptionalNodes` is included as well as the rule-specific tests. The capture filters by the registered rule name.

## Observations

`TestPageStateUnit` compared 47 captured cases: 14 rule-specific cases and 33 absent-node runs (11 shapes at three filenames). All matched Go byte-for-byte on source Node, emitted JavaScript and ASan/UBSan native. The malformed missing-initializer shape is replayed through the harness's findings-only recovery mode; Go's edit engine refuses malformed input, so no converged fixed-source parity is claimed for that shape.

The three owned witnesses also matched on all three backends, both selecting the rule and selecting all rules. Go reports two findings in `testdata/jsx-page/page.jsx.txt`, four in `testdata/optional/page.tsx.txt`, and four in `testdata/parenthesized/page.tsx.txt`. The witnesses cover both hooks, parenthesized callees and receivers, nested functions, optional member/call syntax, the JSX suffix, other receivers, computed accesses and the deliberately excluded `useRef`.

The shared `structureFileContext`, `project`, and `isNamespacedMember` helpers are reused. The optional-member adapter maps the parser's final child to Go's `Name` field; the parser's intervening `QuestionDotToken` is not a Go name. It leaves the shared helper and projection unchanged. Policy text is generated from `helpers/testdata/catalog.json`, Go's validated resolved catalog; this message has no dynamic substitutions.

Go's lexical matching is retained: a bare hook name is checked without resolving its binding, including the undeclared bare hooks in the owned witnesses. The receiver check accepts `(React).useState`, while computed `React['useState']` stays silent. `IsReactFile` is redundant with `IsPageFile`, as upstream documents, and retained.

The semantic mutant replaces the `useReducer` arm with `useRef`. It compiles to emitted JavaScript, runs, and disagrees with Go on Node and emitted JavaScript. See `evidence/mutant.txt` for the exact caught lines and `evidence/selected.log.gz` for the full log. The harness's separate sanitized native canary also passes. An own-rule native mutant is not claimed.

The first selected build rejected a mutable helper projection under Adamic's cycle check. The final port uses the ordinary supported read-only equivalent, and its successful compiled agreement is recorded. No language gap or missing shared helper was needed.

## Validation

Commands and environment are recorded in `evidence/commands.txt`. `evidence/unit_test.go.txt` is the scratch overlay test, with `t.Parallel()`, not a shared harness edit. Compressed logs retain original bytes. Only this rule directory is changed.

The unscoped whole-package command passed with 159 test/subtest passes, zero failures and one skip; see `evidence/whole-summary.txt` and the raw `evidence/whole.jsonl.gz`. All 100 registered mutants were caught. The run took 1,192s wall time (`go test` reported 1,189.766s); `nproc` was 5. Both profile variables named one fresh directory, the TypeScript corpus came from a clean checkout at the required pin, the WASI SDK was supplied, and throughput tests were enabled.

The sole skip, `TestCheckerBridgeRefusalPending`, is the existing `checker_pending_test.go:51` dependency on `codex/tsgo-errors-as-values`: `tsgoInspect` must return `TSGoError` from the C error buffer. It is not an input skip. No rule implementation blocker remains. The full repository gate was not run; the complete requested lint package was run once.
