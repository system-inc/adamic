# TypeScript scanner

This slice ports the ordinary skip-trivia scanner from
`cohere/TypeScript/tsc/internal/scanner/scanner.go`, pinned by the submodule at
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. The Go scanner, not this port,
determines the expected answers. Only files in this directory change.

Build and run after sourcing the setup environment:

```sh
go run ./cmd/adamic build stage1/typescript/scanner/main.ts -o /tmp/adamic-scanner
/tmp/adamic-scanner example.ts
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/scanner/main.ts example.ts
```

`main.ts <file> [scan|regex|greater|template|jsx]` prints one token per line:
`kind byteStart byteEnd flags`, followed by a tab and the cooked value for
identifiers, private identifiers, keywords and literals. Positions are UTF-8
byte offsets and ends are exclusive. Flags are the complete numeric Go
TokenFlags bitset. Values escape non-ASCII, backslash and controls as UTF-16
`\uXXXX` units, preserving lone surrogates. Diagnostics print
`error code byteStart byteLength` before the affected token. Diagnostic text
and substitution arguments are outside this protocol.

`scan` preserves division and the Go scanner's initial single `>` token.
`regex` asks for `ReScanSlashToken(false)` at each slash token, `greater` asks
for `ReScanGreaterThanToken`, and `template` asks for the untagged template
continuation at each closing brace. These explicit modes exercise scanner
operations; they do not implement a parser's contextual choice. `jsx` drives
`ScanJsxToken` over text. The driver also supports a tab-delimited
`--manifest` of mode and path, and `--count` for timing without token output.

## Gap 1: multi-value push: closed

Closed on library/area-on-main-4 (`d767edf5`). The unchanged
`gaps/1_push.ts` prints `ab` on source Node, native ASan/UBSan/LeakSanitizer
and emitted JavaScript. `TestGapStandsWhereGapsMdSays` requires those bytes
instead of the former `NotYet: push with other than one value`.
The UTF-16 byte-offset table now appends both entries for a supplementary
character in one call, preserving their original order and byte offsets.

## Gap 2: arbitrary precision bigint values

Observation: `gaps/2_bigint.ts` runs under Node and prints
`1237940039285380274899124223`. Stage 0 refuses with
`NotYet: a value of type 1237940039285380274899124223n`.
The gaps test requires both the Node result and that exact refusal.

Workaround: `decimalDigits` performs arbitrary precision binary/octal
conversion in decimal strings, in place of Go's `math/big.Int` values.
Only the token's textual value is needed. Generated inputs include
100-digit radix literals and their bigint counterparts, compared against
Go on both Node and native. Hex bigint tokens keep their Go hex spelling.

## Representation choices

Adamic strings use UTF-16 indices; Go scanner strings use byte positions.
When printing tokens, the driver builds a linear UTF-16-to-byte table and checks its final size
with `utf8Length`. Unmappable reported positions panic. Supplementary
identifier characters use code points. Unicode 15.1 range triples are
copied from the pinned Go generated table, with the same stride semantics.

Binary and octal bigint values are converted with decimal string arithmetic,
so there is no arbitrary-precision value requirement. Hex bigint spellings
stay hex, as in Go. Legacy octal literals reproduce Go's signed-64-bit
saturation. Other numeric literal values use the existing stage 0 number
conversion and are checked against Go, including 100-digit radix forms.
The string arithmetic is gap 2's workaround; the remaining choices are
representation differences without an observed compiler refusal.

## Comparison and limits

`ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript go test -count=1 -v
./stage1/typescript/scanner` verifies the checkout is v6.0.3 commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Without the variable, the test
clones that tag into a temporary directory. The corpus includes every
`src/compiler/*.ts` recursively, every repository stage1 `.ts`, and
18,236 generated inputs. Punctuators, keywords and Unicode range boundaries
are generated from the Go source independently of the port's tables.
The comparison includes token kinds, byte ranges, all flags, values and
diagnostic codes/ranges. Native uses ASan, UBSan and LeakSanitizer.

Three independently built mutants must successfully run and disagree with
Go under both Node and native: `!=` becomes `==`, a repeated decimal
separator loses its error and invalid-separator flag, and regex rescanning
is skipped. A crash or compilation refusal is not a killed mutant here.

Not covered: trivia-returning mode, alternate language versions/variants,
tagged template continuations, JSX identifier/attribute rescans, JSDoc
scanner APIs, scanner state/try-scan/lookahead APIs, and the regexp syntax
validator reached by `ReScanSlashToken(true)`. The implemented regex
boundary scan uses `false`, matching the oracle call, and still reports
unterminated literals. No full parser or automatic regex/division context
classification is claimed. Inputs are UTF-8 text; malformed UTF-8 bytes
are outside this slice's corpus.

Timing and gate results are recorded in REPORT.md.
