# ESTree gaps

## Observed Go behavior: trailing multibyte whitespace

The pinned Go `lastWhitespaceSize` returns after its first one-byte attempt.
A UTF-8 continuation byte therefore ends trimming even when the character
is NBSP or U+2028. For the input `x` followed by NBSP and `;`, Go's
ExpressionStatement ContentEnd is byte 3. JavaScript `trimEnd` would give
byte 1. The port uses an explicit Go-compatible tail trim for ContentEnd.
The generated agreement corpus includes both NBSP and U+2028 before a
semicolon, so using JavaScript trimEnd fails byte comparison.

This is an observation of this cohere pin, not an inferred language rule.
No runtime helper was changed. `gaps/nbsp.ts` and `gaps/lineSeparator.ts` are minimal source inputs.
`TestOriginalLibraries` proves the Go/Prettier ContentEnd difference,
allowing exactly the corresponding field delta and no other change.

## Parser dependency boundaries

The parser's `GAPS.md` and `WHOLE_REPORT.md` explicitly exclude diagnostic
parity, JSX and AST list ranges beyond the canonical fields they compare.
The converter's source driver currently refuses unrepresented conversion
kinds rather than outputting a partial tree. Generic argument wrappers now reconstruct positions from child spans and tokens;
ordinary, nested and trailing-comma cases are held to Go and the original library.
These are port coverage limits, not compiler or runtime failures.

## Observed wrapper behavior: CRLF normalization

`gaps/crlf.ts` contains `x`, CRLF, and a semicolon. The raw original converter
agrees with Go. Prettier's full parse wrapper normalizes CRLF before parsing: its
Program and ExpressionStatement ranges end at UTF-16 position 3, whereas this
Go package's direct ParseTypeScript API preserves the original four-byte source.
The port preserves the Go API boundary. `TestOriginalLibraries` requires exactly
this two-range delta. This does not claim the complete format pipeline disagrees.

## Stage 0 NotYet: postfix increment as a value

`gaps/postfixValue.ts` indexes an array with `index++`, then prints the index.
Node prints `7` and `1`. Loading passes; lowering refuses a PostfixUnaryExpression
at the value use. `TestPostfixValueGap` requires both observations. The converter
splits each index read and increment into statements. No compiler fix was made.

## Input API cannot preserve malformed UTF-8

`TestRawInputGap` writes two eight-byte files with hex bytes
`2f 2f f0 90 80 0a 78 3b` and `2f 2f ef bf bd 0a 78 3b`.
Both are a comment and `x;`. `gaps/rawInput.ts` prints fileStatus.size,
utf8Length(readTextFile.text) and the decoded text. Source Node, sanitized native
and emitted JS give the identical `8:8://�\nx;\n` for both files.
Go accepts both and its canonical comment values differ: three replacement
characters versus one. This proves information loss; file size cannot repair it.
Replacement decoding is the documented text-input contract, not an inferred
runtime defect. No raw-byte input function is exposed by the current prelude.

The pipeline conservatively refuses literal U+FFFD, including a valid replacement
character written directly in UTF-8. ASCII escape spellings are still usable.
Nine inventory files contain malformed UTF-8. Complete Go byte parity requires a
raw input boundary before this port can represent those files faithfully.

## Parser recovery and unsupported grammar

Current follow-up: the ESTree-local parser explicitly stops after more than 32
repeated scan positions. All 13 recorded stalls and three minimal EOF cases
refuse before 2s on Node, sanitized native and emitted JS; disabling that guard
hits a 500ms deadline on Node/native. The audit no longer injects a guard.
`gaps/portParserRecovery.ts` uses the local parser. Seven of the recorded stalls
still yield Go trees the port cannot recover; acceptance parity remains a gap.
The following shared-parser observations describe the original dependency,
which is intentionally unchanged, rather than the current local driver.


`gaps/parserRecovery.ts` parses `type X = {`. `TestParserRecoveryGap` externally
kills the unchanged dependency after one second on source Node and sanitized
native. That is a reproducible recovery stall, not a port mutant. The initial
unbounded corpus audit exhausted Node's heap at `asiAbstract.ts`; the final audit
adds an audit-only repeated-scan guard, detecting 13 stalls. It throws after more
than 32 identical scanner positions. The source and native driver do not contain
that instrumentation and are not safe for arbitrary invalid input.

`TestParserBoundaryRefusals` holds three minimal Go-accepted inputs against explicit
source Node/native/emitted JS refusal: `class C { readonly!: number; }`,
`'\ud800a\udc00';`, and `const node = <A/>;` with extension `.tsx`.
The modifier example produces a recovered Identifier at the `!` token, with
stale text `readonly` and no ExclamationToken child. The converter rejects a
recovered identifier whose source token is not an identifier or keyword.
Go's cooked surrogate strings contain WTF-8 bytes; Go canonical serialization
replacement-decodes those bytes, whereas JavaScript preserves UTF-16 surrogates.
The driver explicitly refuses that unrepresented string case rather than
normalizing a wrong value. All `.tsx`/`.jsx` input is conservatively refused.
Diagnostic acceptance parity and JSX grammar are still missing.

The final audit has 92 remaining wrong-output files and 984 acceptance
disagreements. These include converter gaps as well as parser dependency gaps;
no claim is made that all remaining work belongs to the compiler or runtime.
Every file and first differing line is recorded in the compressed audit log.
Whole-checkout success has not been achieved.

## Compiler gap: interface call and concrete default argument

`gaps/interfaceDefault.ts` views a concrete `next(value, step = 1)` method through
an interface exposing only `next(value)`. Type checking and lowering accept the
program. Source Node and emitted JS print 5. Sanitized native reads one argument
past the interface call's storage and reports an AddressSanitizer
stack-buffer-overflow. A release run printed 4; its result is undefined behavior.
`TestInterfaceDefaultGap` requires the two 5 answers and the sanitizer diagnostic.
No compiler or runtime fix was attempted. A larger converter interface refactor
also triggered the explicit missing-method panic; it was discarded in favor of
concrete model classes and tables.

`gaps/methodReplacement.ts` is another minimal compiler boundary. Node permits
replacement of one instance's method, printing replacement then original.
Loading succeeds; lowering refuses unbound-method. `TestMethodReplacementGap`
holds both observations. It prevents installing a progress wrapper by replacing
a method on the parser instance. This does not establish that every possible
progress-wrapper design is impossible.

## Observed Go numeric range differences from the pinned original library

`TestPinnedNumericGaps` uses `1e999; 0x10000000000000000;`. Go's converted Literal
values are both NaN. The independently installed typescript-estree 8.65.0 converter
returns Infinity and 18446744073709552000. The test calls that original parser,
not the port's numeric helper, and holds exactly both answers. Go reports range
failure for decimal overflow and a radix integer above uint64; the port follows
that Go behavior. This is an observed conversion difference, not a JavaScript
numeric rule. `TestScalarEdges` exercises decimal/subnormal thresholds, base
limits, range errors and arbitrary precision bigint spelling on all four drivers.
Five finite numeric/bigint files additionally match the raw original converter
and the postprocessed original wrapper without any allowed deltas.

## Follow-up JSX observations

The previous blanket JSX refusal is removed by the local parser extension.
`gaps/jsxEmptyArguments.tsx` is accepted by Go with an empty params array, while
typescript-estree 8.65.0 and Prettier 3.9.6 both reject it with "Type argument
list cannot be empty." `gaps/jsxEntities.tsx` proves postprocessed ampersand
spelling differences: the Go Literal value is `&😀`, the Prettier value is
`&amp;😀`; Go JSXText decodes `&amp;`, while the Prettier wrapper preserves its
spelling. Raw conversion agrees for these entities. CRLF retains the existing
normalization difference. TestJSXOriginalLibraries asserts each exact delta and
compares the rest of the trees; it does not blanket-allow different JSX trees.
