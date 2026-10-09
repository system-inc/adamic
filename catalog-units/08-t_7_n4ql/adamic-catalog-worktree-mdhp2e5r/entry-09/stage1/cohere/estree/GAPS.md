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

The local driver includes progress guards and now recovers the Go-accepted
members of the thirteen baseline stalls. Go-refused members and three minimal
EOF cases explicitly refuse with parser diagnostics and empty stdout. Disabling
the class-member progress guard reaches a 500ms deadline on source Node/native.
The audit contains no instrumentation. `gaps/portParserRecovery.ts` calls the
actual driver, which checks parser diagnostics before conversion.

The unchanged shared-parser gap `gaps/parserRecovery.ts` still times out after
one second on source Node and sanitized native. That observation concerns the
shared dependency, not the local driver. The original modifier, cooked surrogate
and JSX refusals are repaired and held to Go on all three builds. The baseline
92 output mismatches are also resolved. Current frozen-corpus acceptance still
differs on five files, all refused because the text input boundary loses raw
UTF-8 information. Every disposition is retained; whole-checkout
success is not claimed. See FOLLOWUP.md for checkpoint commands and counts.

## Compiler gap: interface call and concrete default argument

`gaps/interfaceDefault.ts` views a concrete `next(value, step = 1)` method through
an interface exposing only `next(value)`. Source Node prints 5. Main now refuses
during lowering with a typed `*lower.NotYet` at `gaps/interfaceDefault.ts:10:12`:
"stage 0 can't lower a class method through a view that erases its prototype origin yet".
`TestInterfaceDefaultGap` checks the exact type, location and `What` before native emission.
The former native stack-buffer-overflow path is blocked, but dispatch preserving the
concrete method's prototype origin remains a compiler gap for @system_adamic.
The proving program is unchanged; no compiler or runtime fix was made.

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

## Follow-up cooked surrogate representation

The canonical driver now emits the Go-observed three replacement characters
for each unpaired WTF-8 surrogate, rather than refusing cooked surrogate values.
Eight generated cases and all 32 previously refused repository cases match Go
on Node, sanitized native and emitted JS. TestCookedSurrogateLibraryGap proves
the original library instead preserves UTF-16 surrogates. The one-replacement
mutant finishes with wrong bytes and is caught on both runtimes. This repairs
canonical serialization; it does not change the earlier raw-file information
loss, which remains proved and explicitly refused. The old surrogate refusal
above describes the initial checkpoint, not the current driver.

## Follow-up empty generic recovery

`gaps/emptyGenerics.ts` formats f<>(); and class C<> {} through the actual driver.
Go accepts both with an empty type parameter/argument wrapper and its angle
bracket range. The pinned typescript-estree library refuses both. The focused
library test holds those acceptance differences; generated and repository byte
checks hold their exact Go AST and range. No blanket allowance is made for
other library disagreements. Primitive implements is accepted by the original
library and matches Go in both raw and postprocessed trees.

## Go recovery differs from typescript-estree grammar checks

Go accepts `interface I {x:number=5;}`. Pinned typescript-estree refuses with
“A property signature cannot have an initializer.” The actual-driver proving
program is gaps/typeMemberInitializer.ts. TestTypeMemberLibraryGap requires Go's
acceptance and the pinned library's refusal; the generated native/Node/JS agreement
cases include this recovery. The port follows Go, including recovered mapped-type
members and accessor bodies that later TypeScript grammar checks would reject.

## Compiler gap: interface dispatch to a defaulted method

The 13-line `gaps/interfaceTypeMethod.ts` supplies both arguments when calling a
concrete `type(minimum = 0, conditional = true)` through an interface. Source Node
prints 1. Main now refuses during lowering with a typed `*lower.NotYet` at
`gaps/interfaceTypeMethod.ts:10:12`: "stage 0 can't lower a class method through a view
that erases its prototype origin yet". `TestInterfaceTypeMethodGap` checks the exact
type, location and `What`. The former native missing-method panic is blocked; the
origin-preserving dispatch gap remains open for @system_adamic.
The former default-free control is now refused too. Its workaround removes the
concrete defaults, renames the concrete method to `typeValue`, and exposes an explicit
callback property through the interface, retaining both supplied arguments. It
yields 1 on Node, sanitized native and emitted JS. Renaming follows the ts-printer
proof and avoids main's conservative compatible-shape origin guard. Successful
lowering of that workaround proves the diagnostic check rejects a control. The port uses direct
calls and pure helpers instead. Both interface proofs now share the same lowering
refusal, while their proving programs remain unchanged.

## Deep binary expressions in the frozen corpus

The three binderBinaryExpressionStress files formerly exceeded Node's call stack.
Conversion now walks ordinary left binary spines explicitly, postprocessing uses
preorder/postorder work lists, and the canonical serializer uses explicit frames.
All three files match Go on source Node, sanitized native and emitted JS, including
5,333,598 canonical bytes. TestDeepGrammar checks numeric, quoted and logical
4096-operand chains on all builds. gaps/deepBinary.ts calls the actual driver.
This does not claim unbounded stack safety for every possible source nesting shape;
other recursive grammar paths remain. No compiler or runtime files were edited.
