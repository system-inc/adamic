# Lint slice gaps

## 1. Constructor proof carried into another object's method

Observed at the current branch baseline: `gaps/1_nested_constructor.ts` prints
`1` on Node. Stage 0 returns `lower.Refused` at line 7, column 16:

```
Adamic 0.1 refuses this escaping a constructor before every field is set
```

That location is `Inner.touch` calling `this.read()`. `Inner.value` is already
initialized. The unset field belongs to Outer (`label`), and the inner object
has no reference to Outer. The refusal prevents compiling correct TypeScript;
this is an observed compiler defect, not a reference-counting cycle or an
intended language restriction. `TestNestedConstructorGap` requires Node's exact
answer and the constructor-refusal category. Removing the method delegation
made the reduction compile, which isolates the proof-state interaction.

Inference from the reduction and the inspected `unsetUntil` code: lowering a
callee while a caller's constructor is active carries the caller's initialization
boundary into the other object's method. No compiler or runtime code is changed
in this unit.

The initial linter constructed its parser inside its constructor and encountered
this refusal in `Parser.token`, at its delegated `this.make` call. Workaround:
construct the parser and scanner in the driver, pass the initialized objects to
Linter, and parse/walk from `run()` after Linter's construction. The ordinary
three-way tests compile and execute that workaround.

## Existing parser representation gaps used by this slice

Go AST nodes own parent pointers. Adamic cannot retain those alongside owning
children: reference counting cannot free the cycle. The linter uses child
indexes from the parser and a parallel numeric parent table. The inherited
proving program and test are in
[the parser gap directory](../../typescript/parser/gaps/1_strong_ast_parent.ts)
and `TestStrongAstParentGap`; see
[parser GAPS.md](../../typescript/parser/GAPS.md). Concrete Parser, Scanner and
Finding classes avoid the parser's observed structural-interface method bug;
this slice adds no new interface-method workaround.

## 2. Optional indexed access on an optional value

`gaps/2_optional_index.ts` prints `1` on Node. Stage 0 returned `lower.NotYet`:
`?.[] on a value`. Closed by compiler/area-stack (views slice 1, Oct 8): it lowers, and
native and the JavaScript backend print what Node prints (TestClosedComparatorGaps).
The option reader initially used `this.values.get(name)?.[0] ?? fallback`.
Workaround: bind the optional array, guard `undefined`, then index it.

## 3. Numeric logical-or used as a comparator fallback

`gaps/3_numeric_or.ts` prints `2` on Node. Stage 0 returns `lower.NotYet`:
`a BinaryExpression with a number and a number`. The test requires both.
The comparator initially used `startDifference || endDifference || ruleDifference`.
Workaround: test each numeric difference against zero and return explicitly.
This is a stage 0 lowering limit; it does not change sorting semantics.

Closed by compiler/area-next: numeric `||` lowers with JavaScript's ToBoolean, and native and the
JavaScript backend print what Node prints (TestClosedComparatorGaps). The workaround can go.

## Go regex and Unicode services

Warning-comment matching uses a fixed literal pattern shape rather than a
runtime regular expression object: ASCII word boundaries, an optional
whitespace/decoration prefix with backtracking, and Unicode simple folding.
Decoration matching retains Go's case-insensitive character class, including
hyphen ranges and omitted matchers for reversed ranges.
Go `unicode.SimpleFold` and `strconv.IsPrint` tables are generated once into
`unicode.ts` by `testdata/unicode.go` (the file records Unicode version).
The runtime port executes only Adamic code and data. The real Go rule remains
the oracle; no regex pattern parser is claimed. Comment quotes count UTF-8
bytes and retain Go quoting, including non-ASCII printable characters.

## Limits rather than language gaps

Twenty baseline syntax-only rules plus ten implemented continuation rules;
method-signature-style is not fully covered. Cohere's `no-shadow` and `no-redeclare` declare
`NeedsTypeChecker` and use the binder, so they are excluded. Scope-sensitive
syntax work here is generator ownership, loop repetition boundaries, label
resolution and declaration prefixes. General binding/symbol lookup, config
validation, suppression, JSX, JSDoc node traversal and non-UTF-8 input are
outside coverage. Decoded options from real cohere tests are the wire input;
this is not an ESLint configuration decoder.

Five malformed `no-div-regex` fixtures are explicitly marked findings-only.
They compare recovered rule findings and proposed edits. Go's actual
`edit.FixText` rejects invalid input before it runs rules, so converged fixed
output is deliberately not requested for these five cases. No general error
recovery or diagnostic parity is claimed. Baseline own-rule cases are all held;
the continuation has the explicit parser limits below.
The main valid-source corpus still rejects every Go parse diagnostic.

Cohere's comment collector has an ASCII leading-trivia guard, which can miss
an opening comment behind a BOM. The port retains that behavior. Cohere's
live osvfs path strips a leading BOM, while rule fixtures and this raw-file
oracle retain it; `unicode-bom` covers the raw-source rule API and edit engine,
not live cohere filesystem integration.

## 4. Positioned string lastIndexOf

The method-signature fixer needs the last opening delimiter before the first
parameter. `gaps/4_last_index_position.ts` prints `1` on Node; stage 0 refuses
the two-argument `lastIndexOf` form. Closed by compiler/area-stack (views slice 1, Oct 8):
it lowers, and TestClosedComparatorGaps holds native and the JavaScript backend to Node. The port searches `source.slice(0, firstParameterStart)` using
the supported one-argument form. The three-way fixtures hold the resulting
fixes to Go, including generic and commented signatures.

## Parser recovery dependency exposed by method-signature-style

This is a port coverage defect, not an Adamic language restriction. Cohere's
own tests include four malformed shapes, each in both styles: an interface
without a body, an interface missing its closing brace, a stray opening angle,
and a half-written generic. Go recovers these; for property style it even
reports and proposes fixes on three of them. Native and Node panic on three
shapes, and fail to terminate on the missing closing brace. The test records
Go's actual recovered output and bounds each port run at two seconds. All eight
combinations remain visible in the upstream inventory and log, with explicit
`unsupported-recovery` markers. They are NOT counted as identical bytes.

`gaps/5_parser_recovery.ts` is the standalone EOF-loop reduction. The existing
parser's `typeLiteral()` loops until `CloseBraceToken` with no EOF exit. Its
missing-token recovery is also incomplete in the other shapes. No parser or
internal compiler/runtime source is changed here. The malformed bare-arrow
fixture is handled and compared normally as findings-only. Go's edit engine
refuses malformed input, so fixed-source parity is unavailable even there.

The continuation therefore does not deliver the requested twenty completely
verified rules. Nine new rules have full own-fixture coverage, one has the
valid-source coverage above, and ten selected rules are still unimplemented.
