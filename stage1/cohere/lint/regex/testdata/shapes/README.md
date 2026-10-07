# Per-site string-pattern fixtures

Base: area/stage1-lint db2ecc00447f9ebe8adecb190f71ac222e5db860.
Table: unchanged from codex/lint-regex 071fb012848ce0408428c61aba0857cca472236f.
This unit changes testdata only after bringing that table and its existing gate forward.
No runtime/compiler code, rule migration, submodule bump or shared harness edit.

fixtures.json contains one fixture for every table row, in table order: rule and Go
source line, source expression, shape, Go pattern, JavaScript pattern and flags as
strings, bounded input corpus, and both engines' whole matches and UTF-16 spans.
Go's byte offsets become UTF-16 at the Go observation boundary. Empty matches use
Go's adjacent-empty enumeration rule, as in the existing table gate. Captured groups
and replacement/split behavior are not covered. The `g` flag is an observation flag
for enumerating matches, not a claim about a rule's upstream method or flags.

Shapes: plain 76, (?i) 4, (?s) 1, (?m) 1, dynamic 25; total 107. Of the table's 88
MustCompile sites, six are dynamic expressions; the other 19 dynamic sites use Compile.
Every selected instance has a positive input. Dynamic rows include a documented
binding for their option or source expression. These 25 representatives do not
exhaust the infinite option-pattern language and do not change option semantics.
They hold the selected Go-compatible table translations to the pinned Go oracle.

Rows outside the port's shapes: none of these 107 selected instances. Every constructor
argument is a string. No instance contains a RegExp object argument, Unicode-set
q strings, scoped i alternatives exhibiting the named V8 leaks, or a quantifier
bound over 2^31-1. All selected translations compile through the existing static
compiler gate. This classification does not assert support for every future option.

Each numbered .a fixture constructs new RegExp from constant strings and runs today
on source Node and the area JavaScript backend. Its _runtime.a companion gets both
strings from programArguments, so constant folding cannot hide runtime compilation.
All runtime companions run on Node today. The area still refuses these before
backend selection, including emitted JS. The optional --runtime-js-compiler argument
checks the real runtime-string JS path using the unmodified library compiler from
codex/regex-runtime-compiler 6e47d677ebcd6957d37a99bed99206b6abe434ea, built in a
separate checkout. This dependency is not merged into the fixture branch.

The native leg attempts each runtime companion and accepts only the precise
nonconstant-pattern refusal, labeled `awaits codex/regex-runtime-compiler`. Any other
failure fails the gate. As soon as lowering accepts the companion, sanitized native
and area runtime emitted JS must agree; there is no skip or fixed deadline bypass.
-force-native forces the first fixture to pass native, and therefore fails today.
check.py asserts that expected failure is the named refusal; after the compiler
lands it instead requires success. The translation mutant changes PaginationInput
to PaginationInpuX and must execute cleanly, then fail its match comparison naming
base/consistency_require_pagination_argument_name.go:27.

Run from the repository root, with the toolchain environment sourced:

```
go run stage1/cohere/lint/regex/testdata/shapes/build.go stage1/cohere/lint/regex/testdata/shapes > /tmp/regex-shapes-build.log 2>&1
node stage1/cohere/lint/regex/testdata/shapes/generate.mjs > /tmp/regex-shapes-node.log 2>&1
python3 stage1/cohere/lint/regex/testdata/shapes/check.py --runtime-js-compiler /tmp/regex-library-adamic > /tmp/regex-shapes-check.log 2>&1
go vet stage1/cohere/lint/regex/testdata/shapes/gate.go > /tmp/regex-shapes-vet.log 2>&1
```

Generation uses Go regexp, a regexp/syntax witness and up to three matches from the
recorded corpus per row. generate.mjs independently observes Node and refuses a
mismatch before emitting fixtures. gate.go rechecks every stored Go observation
against Go regexp on every run. Patterns are strings throughout the fixture data;
only the executing constructor makes a RegExp. See evidence for observed results.
