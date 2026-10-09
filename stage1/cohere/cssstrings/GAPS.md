# CSS quote sub-printer

This slice ports `adjustStrings` and its `printString`, `getPreferredQuote`, and
`makeString` dependencies from `cohere/internal/format/css/print_misc.go`.
`strings.ts` uses finite scanners, with no TypeScript parser or regex runtime.
The Go scanner and upstream's two fixed regexes choose the same spans: a quote
opens a string, a backslash skips the following character, and the first
unescaped matching quote closes it. Unmatched openers remain unchanged unless a
later opener matches. Both single and double quote preferences are covered.
Quote counts include escaped quotes. When the original enclosing quote wins,
the original spelling, including unnecessary escapes, is retained.

The mandatory oracle is the actual unexported Go implementation, exposed only
through a build overlay by `testdata/bridge.go`. The independent oracle is the
`selector-string` branch of the original PostCSS printer in **Prettier 3.9.6**,
installed in scratch. Calling that printer directly isolates this sub-printer
from unrelated CSS parsing and layout. No expected differences were observed.
The upstream implementation is MIT licensed; its notice is in `LICENSE`.

## Layout and reproduction

- `strings.ts`: the sub-printer.
- `main.ts`: file-to-stdout driver; `<file> [--single]` formats that text with this
  sub-printer. It preserves all bytes outside matched strings, including final
  newlines. This does not implement full CSS formatting or CSS parsing.
- `--batch <file>`: one escaped text per line, prefixed with `d` or `s`. Backslash,
  newline, carriage return, and tab use `\\`, `\n`, `\r`, and `\t`; other UTF-8
  text is literal. Output is one equivalently escaped result per line. This is a
  trusted test protocol, not an input-validation interface.
- `port_test.go`: walks the entire checkout and initialized submodules, skipping
  only `.git`, for every `.css`, `.scss`, and `.less` file. It tests both quote
  preferences, exhaustively generates lengths zero through five over letters,
  both quotes, backslashes, newlines and emoji, and adds long and boundary texts.
  The native ASan/UBSan build, leak check, Node source, JavaScript backend, pinned
  library, three output mutants and three-round throughput checks use that corpus.
- `testdata/library.mjs`: invokes the pinned original printer and checks its version.
- `gaps/` and `gaps_test.go`: minimal stage 0 proving program and its observed status.

```sh
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/adamic-cssstrings-prettier --save-exact prettier@3.9.6
ADAMIC_CSSSTRINGS_LIBRARY=/tmp/adamic-cssstrings-prettier \
  ADAMIC_CSSSTRINGS_KEEP=/tmp/cssstrings-cases.txt \
  go test -count=1 -v -timeout=30m ./stage1/cohere/cssstrings > /tmp/cssstrings-test.log 2>&1
go run ./cmd/adamic build stage1/cohere/cssstrings/main.ts -o /tmp/cssstrings
/tmp/cssstrings example.css
/tmp/cssstrings example.css --single
```

`ADAMIC_CSSSTRINGS_KEEP` also writes a `.files.json` manifest. All timed answers
must still agree with Go. Measurements include process startup, file reading,
batch escaping and stdout collection; compilation is excluded. They are mixed
short-text throughput, not a steady-state full-CSS-formatting benchmark.

## Gap 1: multi-value Array.push: closed

Closed on library/area-on-main-4 (`d767edf5`). The unchanged
`gaps/1_multi_push.ts` prints `ab\n` on source Node, native
ASan/UBSan/LeakSanitizer and emitted JavaScript. `TestMultiPushGap` now
requires those exact bytes instead of `push with other than one value`.
The sub-printer combines adjacent pushes into multi-value calls; no compiler
files or oracle expectations were changed.

## Boundaries

The corpus uses valid UTF-8 decoded as Node decodes files. Invalid byte sequences
are not an exact arbitrary-byte contract. The raw driver uses `/dev/stdout` because
Adamic's current prelude exposes `console.log` and file writes but no raw stdout
write; the driver is verified on Linux, including empty output, no final newline,
CRLF, multiple final newlines and Unicode. Other operating systems and redirected
stdout with unusual device semantics were not checked. CSS numbers, units,
attributes, AST parsing, layout, YAML, Markdown and TOML remain outside this slice.
