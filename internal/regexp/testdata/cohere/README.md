# Cohere regex corpus

The TypeScript 6.0.3 compiler API visits every `.ts`, `.tsx`, `.mts`, and `.cts`
file below `cohere/` and `stage1/cohere/`, including declarations, vendored
TypeScript sources, and intentionally invalid compiler fixtures. It does not
search comments, Go regexp declarations, JavaScript bundles, embedded source
strings, or non-TypeScript files. `node_modules` and `.git` are excluded.
`inventory.json` preserves each occurrence, original literal token, exact UTF-16
pattern, flags, file, line, column, per-file SHA-256, and checkout revisions.
Malformed files are still visited; every TypeScript parse diagnostic is recorded.
The original literal token is also checked with Node's parser, so an unterminated
source token cannot accidentally be treated as an ordinary valid constructor.

The constant evaluator handles string literals, no-substitution templates,
parentheses, assertions, `+`, constant template interpolation, file-local `const`
aliases resolved by the TypeScript binder, regex arguments to constructors, and
Node's deterministic `RegExp.escape` of a constant string. It does not execute
source during inventory extraction. Imported constants, computed properties,
function parameters, arbitrary calls, mutable aliases, and dynamic flags remain
explicit unresolved entries. Constructor aliases and calls without `new` are
outside the requested syntactic `new RegExp(...)` scope. Constant evaluation
assumes the standard global RegExp constructor, not a user-shadowed replacement.
The `dynamic` list means unresolved by this evaluator, not a proof of dynamism.

The fixture evidence has two distinct kinds:

- Constant string arguments to `exec`/`test` and receivers of string regex
  methods are found with the TypeScript AST and binder. Their location and
  method are recorded. That proves the syntactic input relationship, not that
  the enclosing fixture executes or its branch is taken.
- The Go AST parses cohere's selector fixtures and literal `mustParse` inputs
  without grep. A TypeScript AST transform wraps only the two regex literals
  in the selector port. Running the original parser on those fixtures records
  the actual substrings reaching `String.prototype[Symbol.split]`. The trace
  preserves the original fixture text and file:line in the consumed inventory.
  It does not claim the full selector reaches a substring regex or use
  fabricated fixture strings.

The corpus does not trace all cohere APIs, resolve cross-file calls, or recover
all dynamic fixture inputs. Most implementation patterns consequently have
only generated inputs. Go regexes rewritten as explicit character tests in the
stage-1 ports are not regex literals and are outside this inventory.

## Deterministic generation and observations

`generate.cjs` uses regexpp 4.12.2 (ECMAScript 2025) to derive a finite sample of
the pattern's own alphabet: literal characters, range endpoints/midpoints, and
representatives of sets selected by Node. Every alphabet is recorded as UTF-16
units. Witnesses walk the regex syntax tree, including alternatives, captures,
backreferences, and positive assertions. They are candidates, not guaranteed
matches: lookaround, negation, and backtracking still belong to Node.

Xorshift32 starts at `0xc0ae2026 XOR (pattern id + 1)`. Each pattern gets 24
bounded syntax-tree witnesses, 4 repeated witnesses, 32 alphabet strings of
0..16 code points, alphabet singletons, and separately labelled fixed boundary
probes, including line terminators, supplementary points and lone surrogates.
Quantifier witnesses are bounded to 8 repeats and witness strings to 128 UTF-16
units; this is finite sampling, not exhaustive language coverage. Duplicate
inputs retain all provenance. Invalid Node patterns have rejection observations
instead of invented match results. If a literal's original token is invalid
but its recovered body and flags are valid, the report marks a source-only
rejection, not an Adamic regex compiler defect.

Node 24 adds `d` only to expose indices. It records match index, all capture
spans and UTF-16 values, named span/value dictionaries including unmatched
entries, and lastIndex. Nonstateful patterns are probed with both zero and
out-of-range lastIndex. Global/sticky patterns start at 0, 1, length and
length+1, then exec repeatedly through the final null result. Empty matches
record their unchanged lastIndex, then the sequence driver explicitly applies
AdvanceStringIndex, as a string iterator does. It does not claim exec itself
advances empty matches. Supplementary code points advance by two only in u/v.
No `matchAll` lastIndex semantics are inferred from this exec corpus.

`observations.json.gz` contains all Node results, seed, algorithm, runtime
versions, inventory hash, inputs, sequence/step, and provenance. Go and C execute
stateful sequences on the same regex object; each advances empty matches from
its own observed offsets. The C harness prints its observations independently
of expected data. It reads native match index, actual strings, indices and
named objects, rather than reconstructing all values from expected spans.
Three runs exercise bounded execution, unlimited automatic dispatch, and the
forced VM, with AddressSanitizer, UndefinedBehaviorSanitizer and LeakSanitizer.
The boolean-only `test` fast path is not exercised by this exec corpus.

All completed calls are compared. `issues.json` pins compiler refusals and every
observed disagreement, with first counterexample, count and SHA-256 of *all*
disagreeing case IDs, Node results, and Adamic results. No pattern is skipped
because it disagrees. New, changed, or fixed failures require explicit review
and recording; changing one expected capture cannot hide inside an existing
failure. An optional report exported outside the checkout contains every refusal and
disagreement, plus triage inferences. Runtime
interruptions or native build/sanitizer failures fail immediately, rather than
being recorded as matches.

## Regenerate with Node

From the repository, after sourcing the environment printed by `cloud/setup.sh`:

```sh
bash internal/regexp/testdata/cohere/regenerate.sh > /workspace/regex-cohere-regenerate.log 2>&1
ADAMIC_COHERE_RECORD=1 ADAMIC_COHERE_REPORT=/workspace/regex-cohere-report.md go test ./internal/regexp -run '^TestCohere' -count=1 -timeout 10m -v > /workspace/regex-cohere-record.log 2>&1
go test ./internal/regexp -run '^TestCohere' -count=1 -timeout 10m -v > /workspace/regex-cohere-check.log 2>&1
```

The first command installs the two exact npm versions in a scratch directory,
extracts all patterns, traces selector fixtures, and regenerates Node data. It
never silently blesses Adamic's results. The explicit second command writes
`issues.json` as a test expectation and the optional external Markdown report. Review both; it is a characterization
suite containing known failures, not a claim that all regexes agree. CI runs the
third command with Go and clang; no Node or npm is invoked by these tests.
Corpus source hashes preserve the extraction snapshot but do not require the
cohere submodule to be present to run the comparison itself.

## Mutants

```sh
python3 internal/regexp/testdata/cohere/run-mutants.py /workspace/regex-cohere-mutants > /workspace/regex-cohere-mutants.log 2>&1
```

The runner changes one participating recorded capture in the compressed corpus
and runs the full comparison. It then removes the shared named dictionary
comparison and runs the named-only guard seeded by a real recorded named match.
Both must fail for the specified reason. Four further comparator deletions prove
that match index, capture values, lastIndex, and named capture values are each
held independently by a corpus-based guard. Files are restored in `finally` blocks.
No matcher, compiler or runtime source is changed. The comparison guards
also run in ordinary CI. Logs preserve the actual failures, not only a claim
that mutation would be detected.
