Built: collapse.parseModifier, one .a helper, with explicit decoder and validator dependencies.
Commits: claim 22bc331b pushed before code; implementation commit is this report's commit.
Commands: modifier go test PASS 13.573s and at-rule go test PASS 5.797s; owned go vet PASS, empty output; modifier compares 1,148 cases and 26,559 Go bytes in all three Adamic modes.
Mutants: accept blank brackets, omit variable prefix, change named kind, decode before wrapping, omit arbitrary validity, omit named validity; every mutant compiles and finishes cleanly before comparison catches it on source Node, emitted JavaScript and sanitized native.
Not covered: implementations of the explicit shared dependencies, whole-rule findings, full gate, arbitrary invalid UTF-8, process concurrency; six dependency occurrences removed, zero completely helper-ready rules.

The oracle uses unmodified parseModifier at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. An oracle-only overlay exposes the private helper and its real decoder/predicates; cohere's worktree is untouched. Capture stores actual Go decodeArbitraryValue, isValidArbitrary, isBlank and isValidNamedValue results for every possible normal or mutant dependency argument used in the comparison. The owned driver refuses an uncaptured dependency call. None of the helper implementation's branch decisions or output is supplied by the oracle, only the independently owned dependency answers. The final verdict comes from actual Go parseModifier. This proves the modifier logic conditional on those explicit dependencies; it does not claim to port the decoder or validators.

The return model has a presence bit to preserve Go nil versus a successful modifier. Brackets decode, validate, then reject blank values. Parentheses first require -- and validate the inner string, then wrap it in var(...) before decoding, preserving underscores in custom-property names. The shorthand has no blank test, exactly as Go. Named values validate and retain their input unchanged. The helper does not invent an absent modifier for invalid text.

Capture includes every Go string literal in all matching test files from all six consuming rule families: canonical 204, class order 316, variant order 52, shorthand 178, conflicting 195 and unknown 184, plus 19 controls. These literals include source, settings and expected output; this is direct helper comparison, not whole-rule finding parity. Controls cover named, bracketed, CSS-variable shorthand, empty, blank/Unicode blank, malformed brackets, invalid punctuation, wrapping and underscore preservation. Output observes nil/presence, kind and UTF-16 value units byte for byte against Go.

Consumers:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

The at-rule and modifier ports together remove twelve dependency occurrences from these same six rules. Neither completes the rules' remaining helper sets. No finding throughput is reported for helpers that produce no lint findings. The NewTheme duplicate remains withdrawn under slot 08's earlier claim and is not counted.

Reproduce with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/from-wave1-13/at-rule ./stage1/cohere/lint/helpers/from-wave1-13/modifier -count=1 -v -timeout=10m > /tmp/wave13-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-13/at-rule ./stage1/cohere/lint/helpers/from-wave1-13/modifier > /tmp/wave13-helpers-vet.log 2>&1
```

The initial oracle collector edit had a syntax error, preserved in evidence/initial-oracle.log. The repaired collector and final comparisons pass; no compiler or shared harness file was edited.
