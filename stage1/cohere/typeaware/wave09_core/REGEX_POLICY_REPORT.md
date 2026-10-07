Built: an explicit requested dynamic-RegExp gap probe, independent check and updated status for the existing regex claims.
Commits: follows pushed a94250624 on codex/typeaware-wave-09; fetched main f8013f0b is already an ancestor; no new claims.
Commands/output: verify_dynamic_gap.py PASS; source Node returns true, native refuses the nonconstant pattern, static-pattern sanitized mutant builds and returns true.
Mutant: replace the dynamic constructor input with a static 'a'; clean build/run removes the required refusal and is caught by the blocker assertion.
Not covered: dynamic native regex compilation, complete regex findings/fixes/suggestions, shared listener integration or the full gate; these claims remain incomplete.

The latest user instruction prohibits hand-rolled matchers and requests JS
RegExp literals for translated fixed Go patterns, with new RegExp(pattern,'u')
for option patterns. The original Go no-invalid-regexp, no-label-var and
no-misleading-character-class files were inspected. None contains a fixed
regexp.MustCompile pattern for this unit to translate. no-invalid-regexp
instead validates source-derived strings through esregexp.Compile. The
misleading-character-class rule uses the production regex syntax parser to
inspect classes, alongside reference/constant/source-map helpers. Substituting
a RegExp.test boolean alone would not expose those parsed class members.

An all-head origin fetch completed. origin/codex/lint-regex is absent; no
shared translation row is available from that named branch. No placeholder
row or replacement hand-written matcher was invented. Existing custom
pattern scanner/rewrite components are retained as historical evidence only.
They are not being presented as completion under the new requirement, and no
new matcher, pattern parser or rewrite was added in this update.

The requested constructor form is blocked concretely by current lowering:
internal/lower/regexp.go accepts only constant patterns and flags, and rejects
a function parameter supplied to new RegExp. The owned dynamic_regexp_gap.a
probe passes the pattern through a function parameter and uses the exact
new RegExp(pattern,'u') form. Source Node exits 0 with true and empty stderr.
Native build exits nonzero with "RegExp with a nonconstant pattern" before
producing an executable. Replacing the parameter with the static string 'a'
builds under sanitizers and exits 0 with true and empty stderr. That successful
mutant is rejected by the required-refusal assertion, demonstrating the check
observes this precise boundary rather than accepting arbitrary failures.

Reproduce with the existing toolchain, with test output in a file:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_dynamic_gap.py > /workspace/wave-09-dynamic-gap-test.log 2>&1
```

Complete stdout/stderr, generated mutant, results and source hashes are in
validation-dynamic-gap. There is no dynamic native lint time to compare with
Go because this constructor is refused. Prior native/Go corpus timings and
released-handle/sanitizer checks are retained in LANDING_F801_REPORT.md.
No checker ABI or complete rule implementation changed, so those suites were
not repeated for this isolated blocker probe. Setup remains 88 seconds,
nproc 5. Shared compiler, parser, harness and generator files were untouched.

The numeric handed-node interface remains separately unavailable: ParseNode
exposes string kinds, and the checker adapter refetches nodes by index. Six
numeric rule.json declarations are ready, but historical execution is not
claimed speed-compliant. No batch-8 Diagnostic SHA was supplied. There was no
macOS leak-helper diff on this unchanged main to resolve or revert.

These are regex claims, not React analysis claims. They are not parked or
counted as finished under the new React exception. no-label-var retains its
previous supported-scope completion; both regex rules remain incomplete with
their shared boundaries named. No additional batch is claimed while they
remain incomplete. Publishing is solely to codex/typeaware-wave-09.
