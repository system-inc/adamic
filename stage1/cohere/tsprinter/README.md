# TypeScript printer: document layout and expression core

This is a **partial port** of cohere's `internal/format/javascript`, composed with the Adamic
TypeScript parser on `codex/typescript-scanner`. The document engine is ported; the expression core
is the first printer increment. A separate statement driver formats complete files whose statements are supported.
Control flow, typed declarations and types remain later families. Remaining expression families are listed in
[GAPS.md](GAPS.md), with proving inputs. Nothing falls back to printing the input unchanged.

`doc.ts` implements cohere's shared Prettier document interpreter, including group, indent, line,
softline, ifBreak and fill. It also implements conditional groups, keyed and shared groups,
indentIfBreak, alignment, root indentation, hard and literal lines, suffixes, suffix boundaries,
trim, labels and break propagation. Docs are table indexes; shared groups retain their identity.
Unknown doc kinds and forward or cyclic edges panic. `width.ts` and `widthTables.ts` are the already
proven JSON slice's Unicode width implementation, copied from `origin/codex/stage1-json-format`.

`expressions.ts` reads the real parser's indexed nodes. It implements literals, ordinary and logical
binary operators, sequence expressions, assignments, conditionals, object literals, precedence and parentheses, unary and update operators, arrays including holes
and spreads, property and computed accesses including optional access, non-null assertions, and
calls, member-call chains and constructors without type arguments, including expanded object and array arguments. It
includes cohere's logical-tree rebalancing and protection against accidentally making a string
expression a directive. Formatting defaults to width 120, four spaces, single quotes, semicolons,
and trailing commas. The first differential expression corpus uses width 80.

## Build and run

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/tsprinter/main.ts -o /tmp/ts-expression-printer
/tmp/ts-expression-printer /path/to/single-expression.ts
```

The file must contain one supported expression, with an optional semicolon. The formatted expression
statement goes to stdout. An unsupported shape is a loud failure. `--cases <file> [width]` is the test
batch protocol; it returns `ok` with escaped formatting, or `notyet` with its reason, for each input.
`docMain.ts` is the independent document-protocol driver.

## Independent oracles

The initial printer branch started at scanner/parser commit `ed2477e538f54e772c62593de9bab4ddeca0d4ab`.
That slice used cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`; the tsc-corpus repair uses
`7945d102a6c18dd36adf9114a758ce646e8b2359`. Both embed Prettier **3.9.6**.
The TypeScript source pin is 6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
Install upstream Prettier in a scratch directory and point the test at it:

```sh
mkdir -p /tmp/ts-printer-prettier
npm install --prefix /tmp/ts-printer-prettier --no-audit --no-fund prettier@3.9.6
export ADAMIC_TS_PRETTIER=/tmp/ts-printer-prettier
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript
# Optional: retain corpus metadata and a release binary outside the repository.
export ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-corpus
export ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-artifacts
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter > /tmp/ts-printer-test.log 2>&1
```

The tests overlay helpers into cohere without changing its checkout. Go selects maximal supported
fragments using its original TypeScript AST and expression-context predicate, then formats each
fragment independently. The corpus is the git-tracked `.ts` files under the named roots in
[corpus_test.go](corpus_test.go): `bench`, `bridge`, `cmd`, `internal`, `stage1`,
`stage3/drivers/tsc/corpus`, and the pinned TypeScript checkout's `src/compiler`.
Every root logs its count and fails if missing or empty. Untracked files and unnamed roots do not
enter the corpus. The initial slice had 120 repository files and 77 compiler files; those historical
counts below are not the current gate's counts. A rejected parent is traversed for supported children;
it is never silently counted as formatted. Invalid standalone contexts, such as `delete` extracted
from a property name, are counted separately.

The initial selected fragments plus 1,203 generated cases total **149,852**. Generated cases include 625
operator pairs, 52 literal/operator/access/array edges, 60 long argument lists, 60 numeric fills,
and 406 supported recursively generated expressions (seed 20261006). Ninety-four generated
expanded-argument shapes are recorded as outside this core. Counts are fragments, **not complete
files**. Every initial file parsed in Go; the deliberate tsc corpus now records one Go parse
refusal in [TSC_CORPUS.md](TSC_CORPUS.md). Coverage and rejected
candidate counts are saved as `coverage.json`. Counts naturally change when repository sources do.

The document oracle uses cohere's own spec generator, not docs assembled by this port: 5,000 seeded
docs plus 72 boundary docs, exercising widths 8 through 47, tabs and two/four-space indentation,
Unicode, shared groups and every document kind. Native with ASan/UBSan and leaks, source on Node,
and the JavaScript backend must match Go byte for byte. The actual npm Prettier doc engine and
TypeScript printer are separate comparisons. Port output must always match Go. The tsc-corpus
repair separately pins 29 external text/error outcomes in `testdata/tsc-upstream-differences.json`;
all unpinned external outcomes must match Go too. The separate unported proving corpus records
one anonymous-function spacing difference in `testdata/prettier-differences.json`.
Fifteen unported shapes are checked for `NotYet` on all three executions, while Go and Prettier
prove that the same source texts are formatable. Native release output is also compared.

Three mutants must compile and finish normally before a byte mismatch counts as a catch: a group
ignoring width, fill refusing to pack pairs, and required expression parentheses disappearing.
See [VALIDATION.md](VALIDATION.md) for commands, results, limitations and timings.

## Sequence expression increment

The printer now includes comma expressions in every supported expression context. Cohere's ESTree
adapter flattens the left comma spine but keeps explicitly parenthesized sub-sequences. The port
records those boundaries before removing parser wrappers. Top-level expression statements keep
sequence parentheses; wrapped lines after the first item are indented. Nested sequences use their
parent context's indentation. Calls, array items, unary operands and member receivers are covered.

The expanded corpus contains 151,139 fragments from the same 197 source files, including 2,415
generated cases. Of those, 1,212 are sequence cases: 12 boundaries, 240 width cases, and 960
compositions with the existing expression families. See the sequence increment in VALIDATION.md.

## Assignment increment

The expression driver now formats assignments and compound assignments, including all 16
operators, short and long chains, and assignment-sensitive member and binary layout. Destructuring
targets are declined explicitly. The assignment corpus has 150,713 maximal fragments from the
same 197 files, including 9,286 generated cases. The 6,871 added assignment cases cover every
operator, five target layouts, 15 right-hand sides, five outer contexts, and chains up to 20
segments. A lower total fragment count reflects newly accepted larger parents, not omitted files.

## Conditional expression increment

Ternaries now compose with every supported expression family, including nested tests, consequent
and alternate chains, member receivers, assignments and nullish branches. The corpus adds 5,240
generated combinations and nesting cases up to 20 levels. Flattened binary printing preserves its
ancestor stack so nested ternaries receive the same grouping as Go. Counts and execution evidence
are recorded in VALIDATION.md and `results/conditional-coverage.json`.

## Object expression increment

Object values, shorthand properties, computed keys and spreads compose with the existing core.
Object wrapping preserves a newline after the opening brace; arrays of multi-property objects
follow cohere's forced-break rule. String keys use cohere's Unicode-category identifier test and
retain escaped spellings when unquoting is unsafe. Methods and destructuring are explicit gaps.
Extracted object literals are parenthesized to retain expression context in the fragment oracle.

## Call argument increment

Plain-identifier calls and constructors now include object/array argument expansion, conditional
layout states, numeric-array packing, broken-argument propagation, trailing commas and the
CommonJS/AMD special layouts. The corpus contains 154,822 maximal fragments from 198 files,
including 19,084 generated cases. Member-call chains remain the next distinct printer family.

## Member-call increment

Calls now accept every supported callee expression. Member-call chains follow cohere's segmentation
and factory/short-receiver heuristics; curried calls prioritize the correct argument group.
Constructor callees retain parentheses when their left chain contains a call. An optional token is
printed only at the link that owns it, separately from the parser's inherited optional-chain flag.
The corpus contains 137,637 maximal fragments from 198 files; the drop reflects larger accepted
call parents. The generator adds 2,533 member/curried-call compositions, width and boundary cases.

## Template increment

Interpolated and multiline templates now compose with every accepted expression family. Raw
quasis keep their bytes and indentation; interpolation docs follow cohere's source-newline and
absolute-indent rules. Literal newlines break containing groups. Calls with one multiline template
retain the printer's single-template layout. Parenthesized expressions containing optional arguments
or branches are accepted; parentheses stopping an active optional chain remain a proving gap.
The corpus includes 1,538 new template cases and 138,241 maximal fragments from 198 files.

## Arrow and body increment

Untyped arrow parameters (including defaults and rest), nested arrow chains and arrow-sensitive
argument expansion are implemented. Supported block bodies compose expression statements,
return/throw arguments, empty statements and debugger/jump statements. Directive prologues retain
their original status; parenthesized strings do not become directives. This still is an expression
driver, not a claim to complete statement or whole-file printing.

Stateful layout stays in `Expressions`. `syntax.ts` contains pure AST classification helpers, and
`Documents` owns document analysis. This avoids callback allocation and interface dispatch in the
hot path. An earlier callback split exposed two stage-0 gaps; their failing programs and working
forms remain independently tested in GAPS.md.

## Optional-chain boundary increment

Parentheses that stop an optional chain now compose with member lookups, calls, non-null
assertions, assignments and arrows. Normalization records the boundary before removing syntax
wrappers. An optional outer link can remove redundant parentheses while retaining the wrapper's
call-layout distinction. The generated corpus adds 480 combinations, including long chains and
nested parentheses. Context-sensitive await/yield values remain an expression gap.

## Named function increment

Untyped named function expressions now compose ordinary, async and generator prefixes, defaults,
rest and supported block bodies with calls, constructors, objects, arrays and assignments. First
and last function arguments follow their own signature expansion rules. Anonymous functions remain
the separately recorded upstream npm spacing difference; accepted cases retain the strict npm oracle.

## Object method and accessor increment

Untyped object methods, getters and setters compose with the existing expression families and
simple bodies. Keys reuse the existing Unicode and quoted-key rules; async/generator prefixes and
computed names retain their own syntax. Computed short keys now follow Go's cleaned-doc overlap
rule for long values. Typed method signatures remain an explicit proving gap.

## Tagged-template increment

Ordinary tagged templates compose with supported tags, multiline raw quasis, interpolation,
assignments, arrows, calls, member chains and constructors. Required tag parentheses and optional
chain stopping boundaries are retained. Tagged assignment values follow Go's rule that avoids a
break after the operator. Jest `each` tables and type arguments are explicit proving gaps.
Syntax classification is now pure code in `syntax.ts`; document state remains on the concrete printer.

## Await/yield increment

Contextual await, bare yield and delegated yield compose with supported async arrows, named
functions, generators and object methods. Receiver and callee await groups, nesting, precedence,
assignments and long operands follow Go. The input parser retains its actual async/generator
context; the formatter does not force those words into operators globally.

The current expression gate additionally covers named functions, arrows and basic bodies,
object methods/accessors, ordinary tags, contextual await/yield and identifier variable statements.
Incremental coverage, exact commands and mutation results are recorded in VALIDATION.md;
the original corpus totals above describe the initial slice rather than the latest increment.

## Composed statement files

`statementsMain.ts` formats complete files containing supported statements: expression statements,
blocks, return/throw, empty/debugger/break/continue, identifier variable declarations and named
untyped function declarations, including async and generators. Files with unsupported children
are refused as a whole. Directives, empty files and program statement separators are preserved.

```sh
go run ./cmd/adamic build stage1/cohere/tsprinter/statementsMain.ts -o /tmp/ts-statement-printer
/tmp/ts-statement-printer /path/to/supported-file.ts
```

The independent statement oracle walks the same source files and records both maximal fragments
and the number of complete files. `ADAMIC_TS_STATEMENT_KEEP` and `ADAMIC_TS_STATEMENT_ARTIFACTS`
retain its corpus and release executable. Complete-file coverage does not imply that any whole
TypeScript compiler file is supported.

The deliberate tsc-driver corpus gate and its complete disagreement audit are documented in
[TSC_CORPUS.md](TSC_CORPUS.md). Its focused test reports every differing fragment with the originating
file, on source Node, sanitized native and emitted JavaScript. Go cohere remains the formatting oracle.

## Step 26 imports

Side-effect, default, namespace, named and type-only imports, including type specifiers and `with` attributes, compose with the supported statement printer. The method-shorthand source-context fix (`06c6e7cbf`) closes both prior expression regressions. The same 152,660-file corpus has 140 whole-file Go matches (114 before clauses), nine prior text differences and 21 acceptance conflicts, including four newly exposed invalid import-local parser cases. All 61 new fixtures match Go and Prettier, with seven successful-output mutants; retained side-effect and source-context mutants remain independent. [SCOUT.md](SCOUT.md) records exact counts, evidence, shared parser handoffs and current validation. Comments, trivia, exports, import-equals and general type printing remain gaps. Legacy `assert` is refused as it is by the pinned Go parser.
