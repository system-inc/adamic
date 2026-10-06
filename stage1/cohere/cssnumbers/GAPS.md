# CSS numeric adjustment

`numbers.ts` ports the `adjustNumbers` dependency chain from
`cohere/internal/format/css/print_misc.go`, with the canonical unit list from
`css_units.go` (css-units-list 2.1.0). Its public composition `formatValue` is
`adjustNumbers(adjustStrings(value, single))`, the media-value/type printer's
actual operation. The quote implementation is imported from the previous slice;
it is not copied into the new production code.

The number matcher implements the original unsigned decimal/exponent grammar,
including greedy identifier-prefix backtracking. Quoted spans are left alone.
Unknown ASCII units and identifier-prefixed matches retain their spelling.
Known units use the canonical table; `n` is a special accepted unit. Numeric
normalization uses finite scans, retaining leading integer zeros, removing a
redundant zero exponent, exponent plus/leading zeros, trailing decimal zeros and
a redundant dot. The helper only accepts the scanner's matched number grammar;
it is not a general replacement for upstream printNumber on arbitrary strings.
No CSS parser, TypeScript parser or general regex engine is needed.

## Oracles and corpus

`testdata/bridge.go` exposes the actual Go functions through an overlay. It does
not reimplement their behavior or modify the submodule. The independent oracle
calls the original PostCSS `media-value` printer in npm Prettier 3.9.6, verifying
that version on each run. It receives identical texts and quote preferences.
The driver formats a raw file to stdout, or an escaped batch using the prior
slice's protocol. Raw output preserves terminal newlines through `/dev/stdout`.

The test walks every case-insensitive CSS/SCSS/Less file in the checkout and
initialized submodules, skipping only `.git`. Generated cases include all texts
of length zero through five over an eight-character alphabet, a cross-product
of mantissas, exponents, known/unknown units and identifier/quote/sign/Unicode
prefixes, every canonical unit in uppercase, malformed ends, escapes, NUL,
Unicode whitespace, and long repeated quoted/numeric texts. Both preferences
are checked. Native ASan/UBSan, native leak detection, source Node, the JavaScript
backend and the original library must agree with Go. Every timed result must
agree too. Three mutants must compile, finish and disagree solely by output,
both natively and on Node.

```sh
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/adamic-cssstrings-prettier --save-exact prettier@3.9.6
ADAMIC_CSSNUMBERS_LIBRARY=/tmp/adamic-cssstrings-prettier \
  ADAMIC_CSSNUMBERS_KEEP=/tmp/cssnumbers-cases.txt \
  go test -count=1 -v -timeout=30m ./stage1/cohere/cssnumbers > /tmp/cssnumbers-test.log 2>&1
go run ./cmd/adamic build stage1/cohere/cssnumbers/main.ts -o /tmp/cssnumbers
/tmp/cssnumbers value.css
/tmp/cssnumbers value.css --single
```

## Gaps and boundaries

No new stage 0 lowering or output gap was observed. The inherited multi-value
Array.push gap remains recorded and executable in `../cssstrings/gaps/` and its
gap test; new code uses one-value pushes. This is an observation on the corpus,
not a universal proof. No upstream output discrepancy was found.

Full declaration printing is not self-contained here: it requires child-value
printing, recursive printSequence and document layout, plus AST/path/source
metadata. Those are not part of this slice. Neither CSS parser files nor compiler
files were edited. Arbitrary malformed inputs to printNumber as a standalone
function, invalid UTF-8 byte preservation, non-Linux raw stdout devices, and full
CSS layout are not covered. Throughput includes process startup, input, batch
escaping and output collection, excluding compilation; it is not full-formatting
or steady-state throughput. Prettier and css-units-list notices are in LICENSE.
