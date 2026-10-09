# structure/react-element-no-horizontal-rule

Port of cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`,
`internal/lint/rules/structure/react_no_intrinsic_element.go`.
The oracle adapter returns the unchanged upstream rule; it has no options.

The rule uses the shared `structureFileContext` and JSX `elementParts` helpers.
Its numeric projection supplies parser facts, without implementing either helper.
It reports the whole opening or self-closing element only when its tag name is
an identifier whose exact text is `hr`. The filename exemption comes from the
shared helper, including Go's backslash normalization and substring semantics.
There are no fixes or suggestions.

## Observations

Six unique asserted upstream cases were captured from
`TestReactNoHorizontalRuleElementFires` and
`TestReactNoHorizontalRuleElementStaysSilent`. All match Go on source Node,
Adamic's emitted JavaScript and ASan/UBSan native, byte for byte, including
messages, ranges, finding order, fixes, suggestions and fixed source.
Four witnesses cover self-closing siblings, an opening element with children,
identifier/member/component boundaries and non-ASCII UTF-16/UTF-8 ranges.
The selected comparison produced 4,529 identical output bytes.

The 11 sources and three filenames in
`TestNoRuleCrashesOnAbsentOptionalNodes` contribute another 33 upstream inputs.
Those tests call Run without assertions, so the ordinary assertion capture
omits them. The owned selected-test fragment copies the source strings verbatim
and replays each at the same path suffix. All 33 match across the same three
backends, producing 4,682 identical output bytes. This includes the intentionally
malformed binding-pattern input and uses the inherited recovery-row contract.

The upstream silent case named `an hr inside the Link implementation is still
reported by the other rule only` actually contains an `a`, not an `hr`.
The port matches the source and Go's clean result; the case name does not prove
an exemption for Link or describe a tested horizontal rule.

The semantic mutant changes the accepted tag from `hr` to `a`. It compiles and
runs on all three backends, and Go comparison catches it on each. It is not a
compiler-refusal or sanitizer-crash control.
See [selected.txt](evidence/selected.txt), [absent.txt](evidence/absent.txt) and
the corresponding lossless compressed logs for the observed output.

## Reproduction

Run `go run ./cmd/lint-registry` from the repository root before a manual build.
`testdata/selected.go.txt` contains two top-level parallel tests. Append it to a
scratch copy of `stage1/cohere/lint/harness_test.go`, then use a Go overlay that
maps the original absolute path to the scratch file. Neither shared tests nor
cohere source are changed. Run the selected tests with
`go test -overlay=<overlay.json> ./stage1/cohere/lint -run '^TestStructureReactElementNoHorizontalRule' -count=1 -v -timeout=30m`,
sending output to a log file.

The whole lint package runs without this overlay, so the inherited corpus,
owned selected/all-rule witnesses, discovered mutant and shared package checks
remain unchanged. Set `ADAMIC_TYPESCRIPT_SOURCE` to a clean v6.0.3 checkout,
`WASI_SYSROOT` to the WASI SDK 27 sysroot, `ADAMIC_LINT_BENCH=1`, and both
`ADAMIC_LINT_PROFILE_DIR` and `ADAMIC_LINT_PROFILE_SNAPSHOTS` to one fresh scratch
directory. Run `go test ./stage1/cohere/lint -count=1 -json -timeout=45m` to a log.
The complete-gate result and [setup timing](evidence/setup.txt) are recorded under `evidence/`.

The first whole-package attempt reached its 30-minute timer during the discovered
mutant sweep. It had 132 passing test events, one known skip, and unfinished
checks; the package exited 1. It was not a passing gate. The complete raw JSON
log and metadata are retained as `evidence/timeout-full.jsonl.gz` and
`evidence/timeout-full-meta.json.gz`. The rerun uses a 45-minute test limit and a
new directory shared by the two profile variables. No rule, compiler, runtime
or shared test code changes were needed for this execution-limit failure.

## Complete gate result

The rerun exited 0: 159 passing tests including subtests,
0 failures, and 1 skip.
`TestCheckerBridgeRefusalPending` is the sole skip; it awaits the base's missing
`TSGoError` result API. No optional-input check skipped. Wall time was
1664.411 seconds, `nproc` was 5, and one-minute load moved
from 2.15 to 3.61.
The six captured own-rule cases, inherited corpus, owned selected/all-rule
witnesses and the discovered wrong-tag mutant passed the package's checks.
See [full metadata](evidence/full-meta.json), [full summary](evidence/full.txt)
and `evidence/full.jsonl.gz` for the lossless raw JSON test log.
There are no rule-specific helper or language blockers.

## Limits

This is a syntax rule port, with the upstream Go implementation as authority.
It adds no compiler, runtime, parser or shared-helper behavior. The captured
cases, guard inputs and witnesses are observations, not an exhaustive proof for
all malformed JSX or arbitrary filenames. Shared package skips, if any, are
listed in the complete-gate metadata rather than counted as successful checks.
