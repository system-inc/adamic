# Markdown inline leaf printers

This slice ports the parser-free leaf boundary in Go cohere's
`internal/format/markdown/word.go` and `print.go`: contextual word escaping,
inline code, wiki links, image destinations/titles/alt text, image references,
reference-label normalization and footnote references. The punctuation table
is the BMP projection of `classes_generated.go`. The original library is
Prettier 3.9.6. Its MIT notice is in LICENSE.

This is the largest clean boundary found without importing a parser, recursive
child printing or the document-layout engine. Full Markdown paragraphs, lists,
tables and YAML blocks/collections require those dependencies. CSS declaration
printing also requires recursive value/document printing and AST/source state;
the CSS parser is another worker's territory. No compiler or parser file is
changed here.

## Driver and integration boundary

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/markdowninline/main.ts -o /tmp/markdown-inline
/tmp/markdown-inline input.md w
```

The raw driver treats the complete file as one leaf value and writes its formatted
bytes to stdout. It does not parse a Markdown document. Modes select explicit
node/ancestor contexts:

| Modes | Leaf and context |
| --- | --- |
| w, e | Word inside emphasis with letter or space sibling flanks |
| f, n | First word, with its sentence first or later inside emphasis |
| s, p | Word without emphasis, eligible pseudo setext or plain context |
| c, t | Inline code, preserved newlines, outside/inside a table |
| r, u | Inline code, newlines replaced with spaces, outside/inside a table |
| k, v | Wiki link, preserved or collapsed tabs/newlines |
| h, i | Image, value as alt/URL/title, double/single quote preference |
| o | Image, original alt value with fallback alt and absent title |
| j, l, q | Full, collapsed, shortcut image reference |
| b | Footnote reference |

The exported helpers accept separate strings for sibling values, alt text, original
alt text, URL and title. Missing optional strings map to the empty string, which
has the same output at this boundary. The booleans supplied to `printWord` are
facts a future AST adapter must establish: emphasis ancestor, first word of the
first sentence directly inside that ancestor, and the complete preserved-prose
pseudo-setext condition. The driver modes are fixtures for those facts, not an
AST adapter. `--batch <file>` reads mode-prefixed, escaped lines; the four escapes
are backslash, newline, carriage return and tab. Unknown modes fail explicitly.
Raw stdout uses `/dev/stdout`, verified on Linux.

## Closed delimiter-expression gap

`gaps/1_delimiter_expression.ts` checks the original non-u expression against
Node, native sanitizers, the JavaScript backend and LeakSanitizer:

```text
Node: a\*b
native and JavaScript backend: a\*b
```

`TestDelimiterExpressionMatchesNode` verifies that native regex closes this gap.
`inline.ts` retains its independently verified finite matcher for this specific expression,
including greedy delimiter runs, ordered alternatives, backtracking, escaped
backslash parity, line terminators and non-overlapping matches. It is not a
regular-expression engine. The two simple whitespace expressions are explicit
scans. Output cuts occur at ASCII delimiters, preserving astral UTF-16 pairs.
No output or lowering discrepancy was observed in the supported port itself.

## Oracles, corpus and mutants

The Go adapter is an overlay inside the actual upstream package. It constructs
real `printing.AstPath` stacks and calls `printMdast`, then the actual doc printer;
it does not duplicate the formatting algorithms. The independent JavaScript
adapter invokes `prettier/plugins/markdown`'s original `mdast.print`, supplying
matching path facts and rendering the returned doc with Prettier's doc printer.
It checks the pinned library version on every invocation.

The corpus walks every `.md`, `.markdown`, `.mdown` and `.mkd` file in the checkout
and initialized submodules, skipping only `.git`. Every file is supplied as a
leaf value in every mode, deliberately bypassing document parsing. Generated
cases exhaust length zero through four over eight tokens (including an astral
scalar), cover all 63,488 BMP Unicode scalars on both sides of both emphasis
markers, and add astral punctuation endpoints, line terminators, NUL, malformed
ends, long repeated values, quotes, brackets, pipes and backtick runs. Native
ASan/UBSan, LeakSanitizer, source Node, the JavaScript backend and pinned Prettier
must match Go byte for byte. Every repository file also receives a native raw
stdout check. Three mutants invert escape parity, invert table-pipe handling,
and force a present backtick fence; each must lower, build, exit zero with empty
stderr and be caught solely by its output on native and Node.

```sh
npm install --prefix /tmp/adamic-markdown-prettier --save-exact prettier@3.9.6
ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/adamic-markdown-prettier/node_modules/prettier \
  ADAMIC_MARKDOWNINLINE_KEEP=/tmp/markdown-cases.txt \
  go test -count=1 -v -timeout=30m ./stage1/cohere/markdowninline > /tmp/markdown-test.log 2>&1
```

Measurements are three fresh process runs per implementation, including startup,
file I/O, batch decoding/encoding and captured output, excluding builds. This
mixed-size corpus includes very large documents as single leaves; texts/s is not
full-document formatting throughput or a steady-state small-token benchmark.
Unsupported coverage includes full Markdown parsing/layout, MDX's legacy word
printer, invalid AST shapes such as null wiki targets, isolated surrogate or
invalid UTF-8 file bytes, and non-Linux stdout devices.
