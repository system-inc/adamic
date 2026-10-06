# TypeScript printer: document layout and expression core

This is a **partial port** of cohere's `internal/format/javascript`, composed with the Adamic
TypeScript parser on `codex/typescript-scanner`. The document engine is ported; the expression core
is the first printer increment. This is not yet a formatter for complete TypeScript source files.
Statements, declarations and types are later slices. Remaining expression families are listed in
[GAPS.md](GAPS.md), with proving inputs. Nothing falls back to printing the input unchanged.

`doc.ts` implements cohere's shared Prettier document interpreter, including group, indent, line,
softline, ifBreak and fill. It also implements conditional groups, keyed and shared groups,
indentIfBreak, alignment, root indentation, hard and literal lines, suffixes, suffix boundaries,
trim, labels and break propagation. Docs are table indexes; shared groups retain their identity.
Unknown doc kinds and forward or cyclic edges panic. `width.ts` and `widthTables.ts` are the already
proven JSON slice's Unicode width implementation, copied from `origin/codex/stage1-json-format`.

`expressions.ts` reads the real parser's indexed nodes. It implements literals, ordinary and logical
binary operators, sequence expressions, assignments, precedence and parentheses, unary and update operators, arrays including holes
and spreads, property and computed accesses including optional access, non-null assertions, and
plain-identifier calls and constructors without type arguments or expanded array arguments. It
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

The branch starts at scanner/parser commit `ed2477e538f54e772c62593de9bab4ddeca0d4ab`.
The cohere submodule pin is `715ba94f3608a6500086b1076ce5cb7e51b836db`; its embedded Prettier is **3.9.6**.
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
fragment independently. All 120 `.ts` files in this checkout, including other stage-1 ports and gap
programs, and all 77 `src/compiler` files are walked. The root cohere submodule is excluded from the
repository walk: it is a separate repository, not the requested Adamic source corpus. A rejected
parent is traversed for supported children; it is never silently counted as formatted. Invalid
standalone contexts, such as `delete` extracted from a property name, are counted separately.

The selected fragments plus 1,203 generated cases total **149,852**. Generated cases include 625
operator pairs, 52 literal/operator/access/array edges, 60 long argument lists, 60 numeric fills,
and 406 supported recursively generated expressions (seed 20261006). Ninety-four generated
expanded-argument shapes are recorded as outside this core. Counts are fragments, **not complete
files**. Every file parses in Go; no file is dropped for a parse error. Coverage and rejected
candidate counts are saved as `coverage.json`. Counts naturally change when repository sources do.

The document oracle uses cohere's own spec generator, not docs assembled by this port: 5,000 seeded
docs plus 72 boundary docs, exercising widths 8 through 47, tabs and two/four-space indentation,
Unicode, shared groups and every document kind. Native with ASan/UBSan and leaks, source on Node,
and the JavaScript backend must match Go byte for byte. The actual npm Prettier doc engine and
TypeScript printer are separate comparisons; every selected case must agree with them too. The
separate unported proving corpus records one anonymous-function spacing difference in `testdata/prettier-differences.json`; accepted cases have
no exceptions.
Seventeen unported shapes are checked for `NotYet` on all three executions, while Go and Prettier
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
