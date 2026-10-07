Built: three continuation claims, updated helpers, and .a probes; no new rule port is certified.
Commits: claim 8ab1afd2 pushed before code; helpers merge 4bc1c523; evidence commit follows.
Commands: original Go rule suites PASS; Node/emitted-JS/sanitized-native substrate controls PASS; setup exits 1, nproc 5.
Mutants: compiling string-count probe mutant caught by Go comparison on all three paths; compiling comment-adapter guard mutant caught by refusal check; no rule-semantic mutants.
Not covered: new rule implementations, full findings/fixes parity over compiler and stage1, native/Node/Go findings per second, full repository gate.

Branch: codex/lint-wave1-04. Everything from the previous unit was pushed before
fetching and selecting this continuation. No pull request was opened.

## Selection and claim

The claim reserves, in order:

1. structure/tailwind-no-physical-direction, position 46 of the helper list.
2. @eslint-community/eslint-comments/require-description, first remaining
   syntax-only inventory entry.
3. @next/next/google-font-display, next remaining syntax-only inventory entry.

The selection searched 297 fetched origin refs and 27 unique Markdown blobs
under stage1/cohere/lint/claims/. Main was ef3d907ecdc4c771b016f7d9c52372def057a340.
Published assignments, including previous skipped assignments, were treated as
occupied. None of these three names appeared in other claim documents or main's
executable lint sources at selection time. The helper REPORT.md still specifies
the original 46-rule list delegated to HELPERS.md. The later comment-helper
readiness increment is a separate ledger. The inventory's rules array was then
filtered to neither needs_type_information nor binding_only, preserving order.

8ab1afd2 was pushed before creating the probe. The newer helper branch
6769b88e876e3797293b34a29dfd02cbd8507256 was then merged cleanly as 4bc1c523.
The new shared comment helpers are available; they were not reimplemented here.
Claims remain reserved, with this blocked status recorded explicitly.

## Measured substrate, not completed ports

Cohere's internal/lint/rules/PortingARule.md says:

> If the substrate is genuinely absent, the deliverable changes from a port to a
> measurement. Prove the absence with a compiling probe that has a control, say
> what is missing and how big it is, and stop.

The owned parser_gap.a is that probe. It uses the existing independent stage1
parser and comment helper, with no Go AST projection. It counts reachable string
literals, JSX opening/self-closing elements and comments. Its oracle.go.txt is
built through a Go overlay inside the unchanged cohere module. Go supplies its
own parser observations and separately executes the actual selected rules.

Each selected rule has a plain TypeScript control and a JSX witness copied from
its original Go tests. All controls agree byte for byte between Go, source Node,
emitted JavaScript and ASan/UBSan native. All six sources are valid Go TypeScript,
with zero parse diagnostics. Go reports the expected positive rule message on
each JSX witness. The findings are observed on Go only, not claimed to have been
ported to Adamic by this probe.

| Rule witness | Go | Source Node, emitted JS and sanitized native |
|---|---|---|
| Tailwind attribute | 1 JSX opening, useLogicalClass | Exit 70: parser expected GreaterThanToken, got Identifier at 22 |
| require-description JSX comment | 1 JSX opening, 1 comment, missingDescription | Exit 0, no stderr, 0 JSX openings, 1 comment |
| Google Fonts link | 1 JSX opening, googleFontDisplayMissing | Exit 70: parser expected GreaterThanToken, got Identifier at 32 |

The require-description observation is deliberately narrower: comment discovery
agrees on its JSX witness despite the invalid AST shape. This is not evidence
that its findings differ. It does show that the independent adapter lacks the
validated JSX tree assumed by the helper inventory. The merged comment-helper
adapter explicitly refuses this case, and its guard mutant proves the underlying
parser would otherwise silently accept it.

The two attribute witnesses directly prevent complete independent-parser parity
for Tailwind and Google Fonts. Google Fonts also depends on JSX helpers and HTML
entity decoding not supplied by the two original foundations. require-description
still needs its directive grammar/defaults/options/reporting implementation and
integration. These measurements do not establish that every non-JSX source is
supported or that a require-description port is intrinsically impossible.

The unchanged directory registry still opens rule.ts and rejects .a mutant files;
its copied-source harness omits .a. The previous report contains executable
registration probes. The new direct build demonstrates that .a compilation,
Node execution and emitted JavaScript are available independently of that
registration contract. No incomplete descriptor was installed to break discovery,
and no shared dispatch, copied-file list, parser or protected compiler file was
changed. There is no new .ts Adamic source.

## Size of the original Go corpus

capture.py reruns the original Go test suites with their existing docs capture
enabled. It preserves source, filename, options, outcomes and expected findings,
deduplicating only identical rule/file/source/options configurations. Its Go
oracle decodes the real require-description options, walks the real rule, and
checks every finding ID against the captured asserted verdict.

| Rule | Unique configurations | JSX configurations | Recovery configurations | Go findings |
|---|---:|---:|---:|---:|
| Tailwind | 54 | 1 | 0 | 33 |
| require-description | 109 | 1 | 1 | 55 |
| Google Fonts | 16 | 16 | 0 | 7 |

The recovery case is the deliberately unterminated block comment. Counts are
measurements of these 179 original test configurations, not TypeScript compiler
or stage1 corpus coverage. See asserted-cases.json, go-facts.jsonl and
corpus-size.json under wave1-04-probes/evidence/.

The first capture analysis failed because Go omits the findings field on a clean
case. The script now handles that omission; the retry passes and compares actual
Go findings for every row. Its failure log is preserved.

## Commands, setup and mutants

Source /workspace/adamic-tools/env.sh before every build. Setup's cache warm still
fails at stage1/cohere/lint/profile_test.go:32:

```
cannot range over portFiles (value of type func(t *testing.T) []string): func must be func(yield func(...) bool): unexpected results
```

Timing lines: go ready (0s), clang ready (0s), node ready (0s), submodules ready
(0s). No done/build-cache-ready line. nproc: 5. Tools: Go 1.27.1, clang 20.1.8,
Node 24.19.0. The setup sanitizer probe succeeds. Tests use independently built
drivers to work around the lint harness's compilation failure without editing it.

All test output goes directly to log files; none is piped.

```
python3 stage1/cohere/lint/claims/wave1-04-probes/capture.py > /tmp/lint-wave1-04-cont-capture-retry.log 2>&1
go test -count=1 -v -timeout 10m ./stage1/cohere/lint/claims/wave1-04-probes > /tmp/lint-wave1-04-cont-probes-final.log 2>&1
go test -count=1 -v -timeout 10m ./stage1/cohere/lint/helpers/comments -run '^Test(JsxParserGapIsExplicit|JsxAdapterGuardMutant|CommentsMatchCohere)$' > /tmp/lint-wave1-04-cont-comment-guards.log 2>&1
```

The original Go core, next and tailwind suites pass. capture.py's three separate
suite logs preserve every test outcome. The final probe package requires exact
parser refusals, so a sanitizer failure or arbitrary crash cannot satisfy a gap
check. Its silent-misparse witness must finish normally and produce the exact
known wrong count. The initial probe package passed in 51.467s; the final run's
time is 27.638s, recorded in evidence/probes-final.log. The owned probe package's
go vet passes with an empty log, and git diff --check passes.

The filtered comment-helper package passes in 46.857s: 390 matching output lines
on Go/Node/native; four valid Go JSX cases explicitly refused; the compiling
adapter-guard mutant returns exit 0 instead of the required refusal and is caught.
This existing helper mutant is not a mutant of require-description.

The owned instrumentation mutant changes strings++ to strings += 2. It compiles
and finishes all plain-source controls with exit 0 and empty stderr on source
Node, emitted JavaScript and sanitized native. Only comparison with the
independent Go parser catches its wrong output. This establishes that the probe
can fail, not that any selected rule is correct. No rule-semantic mutant was run,
because no new rule implementation was delivered.

## Remaining work and limits

Enable .a directory registration and the shared copied-source/mutant harness,
repair the pre-existing profile test API mismatch, and provide a validated JSX
adapter before claiming the complete three-rule port. Then implement the rules
and compare exact diagnostics, spans, messages, suggestions and unchanged source
for their no-fix outcomes over all required corpora.

Native, Node and Go findings per second are unmeasured, not zero. The probe counts
AST nodes, not findings; substituting its throughput would misstate the requested
measurement. No full gate or new-rule certification is claimed. No protected
compiler, shared lint source or cohere submodule file was edited.
