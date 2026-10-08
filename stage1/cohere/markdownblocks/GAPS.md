# Markdown parser and layout: embedding off

The contract is Markdown parsing and layout with `embeddedLanguageFormatting: off`
on both oracles. The port is held to Go cohere. Embedded JSON, YAML, TOML and
other languages compose later from their own slices. Front matter and fenced
contents are raw under this contract. The prior inline printers are available as
a dependency. The complete native parser and block formatter are still unfinished.

## Seat compiler reconciliation, October 7, 2026

On seat `1cd00a262ed08e67c51c1bc368e07bc5a4ec0117`, with main
`71d7e491b3c9724f7a0e2ee754592149e7f9790b` already merged, gap 11 is closed:
the unchanged function-expression probe now runs byte-identically on source
Node, native and the JavaScript backend, with leak detection. `AstPath.map`
now calls `each` with a capturing function expression, removing the duplicated
traversal loop. Result collection still uses push because gap 13 remains open.

Gap 16 moved from NotYet to the exact Refused diagnostic recorded below. Its
unchanged program already calls shift on its receiver; that form is still
refused, so the production task-list removal retains discarded splice(0,1).
The test holds the complete diagnostic, including source line/column and fix.

Main also closed gap 6: its unchanged probe prints `missing` on all three
backends with leak detection, as main's positive test already asserts. The source
decoder retains its explicit initialized match flag: it is how the decoder
tracks whether a reference matched, independently of optional-field initialization.
All other representation probes retain their recorded diagnoses or observations,
including gap 13's fatal array-growth check. Historical unit coverage counts
below describe their original censuses. Gate corpora now use this unit's files
and generated cases; the whole-repository census is opt-in through
`ADAMIC_MARKDOWNBLOCKS_CENSUS=1`. The width oracle has its seat's named skip
without `ADAMIC_MARKDOWNWIDTH_DEPS`; with dependencies installed it runs unchanged.

## Native tokenizer event primitives and full mdast construction

`tokenArena.ts` and `tokenizerEvents.ts` port token identity, copied points,
consume/previous/expected-code state, enter/exit events, reused token fields,
column skips, checkpoints/rollback, token chunk slicing and serialization.
Numeric token/context/construct IDs keep the graph acyclic. Actual Go private
createTokenizer and the pinned fork's unchanged private factory run a small
controlled construct program independently. It exercises successful checks that
roll back, line skips and reused field objects on every special code. The native
probe drives those same effects explicitly; it does not implement the construct
factory or Markdown grammar. The fork's private consumed/stack cells are not
public: the adapter infers their terminal true/empty values from its completed
state program; Go reads them directly. Public events, points, source slices,
previous and currentConstruct come from both actual engines.

All 79,873 cases agree: the 4,943 accumulated documents, all 65,536 UTF-16 units
including lone surrogates, special-code sequences through length five, empty and
column/BOM/NUL/CRLF/tab edges. Native serialization follows Go's UTF-16 decoding,
replacing unmatched surrogates. Original source strings retain them internally,
but direct UTF-8 stdout replaces them too. This is byte parity of the exposed
observations, not proof that the two engines store lone surrogates identically.
Zero-span code-chunk SliceStream is an invalid Go operation; the controlled
program does not call it. Streaming writes, construct attempts/interrupt views,
resolver ordering and subtokenization are still unported.

`mdastCompile.ts`, `mdastNode.ts` and `mdastArena.ts` construct the complete non-MDX
Go mdast from resolved events: list-item insertion and spread inference, all core
enter/exit handlers, buffers/text merging, references and escapes, autolinks,
footnotes, strikethrough, tables/task lists, math, wiki links and liquid.
`parseFrontMatter.ts` recognizes, blanks and inserts front matter, preserving
all UTF-16 lengths, explicit languages, delimiters and raw values. Embedding is
absent from this parsing/construction layer; raw contents stay raw.

The Go adapter transports actual PRE-COMPILATION resolved tokens, event ordering,
all context chunk streams, flags and private chunk indices. It does not supply
constructed nodes, listItem events, inferred list spreads or token text slices.
Native computes slices and the tree. Its 5,008-document gate includes all 876
physical files, all accumulated generated documents, extension cases and 54
additional front-matter contexts. Expected trees come from actual Go ParseMarkdown
and the pinned fork's actual full source parser. Independently, the unchanged Go
and fork event compilers reconstruct trees from the same event transport. Both
routes agree with native/source/backend, sanitizers and leaks.

The canonical protocol compares every Go construction field, null/absent flags,
child order and coordinates. Go byte offsets are converted to UTF-16 for the
comparison. Original HAST data, originalLabelText and open JavaScript object
properties that Go omits are outside this typed Go projection. In particular the
fork's frontMatter.start coordinate object is not Go's list-start integer, and
its raw value belongs in the Go FrontMatter record, not Go's literal Value field.
No Markdown grammar decisions are native yet: actual Go supplies resolved lexical
events. The new full node arena has not replaced the existing layout transport or
been composed with the earlier standalone preprocessing/path arena. This is a
native constructor and event primitive slice, not a complete native source parser
or Markdown formatter.

Seven new zero-exit output mutants are caught: event rollback, virtual-space
serialization, restored construct, text construction, list spread, YAML closing
fallback and the generated identifier case table. Three additional expected-exit70
message mutants are caught solely by stderr. Full coverage and exact observations
are in REPORT.txt. Malformed event messages below are held to Go; fatal panics are
not Go's recoverable error values or the fork's catchable exceptions. Leak checks
apply to successful construction, not fatal-panic probes.

### Identifier normalization witnesses

Go normalizes identifiers with simple lower/upper/lower Unicode mappings. The
fork uses JavaScript full casing and context-sensitive final sigma. The native
`identifier.ts` uses generated Go Unicode17.0.0 tables. All 1,112,064 Unicode
scalars agree with actual cohere NormalizeIdentifier followed by Go ToLower on
native/source/backend, with sanitizers/leaks and a case-table mutant. No mapping
is accepted solely because its generator produced it. The generator is
`testdata/identifier_go.go --generate`, built through the test's Go overlay.

`gaps/identifier_case.jsonl` contains three full Markdown proving programs:
sharp-s, dotted-I and Greek final sigma. The complete canonical Go outputs are
[identifier_case_go.txt](gaps/identifier_case_go.txt), and the complete pinned
fork outputs are [identifier_case_fork.txt](gaps/identifier_case_fork.txt).
The test asserts both complete files and holds native construction to Go on all
three, with source/backend/sanitizers/leaks. Identifier values differ as follows:

| Label | Go identifier | Fork identifier |
| --- | --- | --- |
| straße | straße | strasse |
| İ | i (U+0069) | i + U+0307 |
| ΟΣ | οσ (U+03BF U+03C3) | ος (U+03BF U+03C2) |

These witnesses use identical definition/reference labels, so both grammars
resolve their references; the identifier fields still differ. They do not claim
that different labels which fold together in only one implementation produce the
same lexical events or tree.

### Malformed-event diagnostic witnesses

`gaps/event_unclosed.txt`, `event_not-open.txt` and `event_mismatch.txt` are minimal
native event proving programs. Go's actual compiler is invoked by
`testdata/mdast_go.go --error <name>` through its overlay bridge. The fork's actual
compiler is invoked by `testdata/mdast_library.mjs <bundles> --error <name>`.
Each complete Go/fork output pair is retained and asserted:

Unclosed, Go:
```text
Cannot close document, a token (`paragraph`) is still open
```
Unclosed, fork:
```text
Cannot close document, a token (`paragraph`, 1:1-1:1) is still open
```
Not open, Go:
```text
Cannot close `paragraph`: it’s not open
```
Not open, fork:
```text
Cannot close `paragraph` (1:1-1:1): it’s not open
```
Mismatch, Go:
```text
Cannot close `strong`: a different token (`paragraph`) is open
```
Mismatch, fork:
```text
Cannot close `strong` (1:1-1:1): a different token (`paragraph`, 1:1-1:1) is open
```

Native/source/backend emit `adamic: panic: ` followed by the complete Go text,
with empty stdout and exit70. The test normalizes Go's recovered panic and the
fork's thrown Error to message values before comparing. It does not claim that
all three original process-level exception behaviors are equal. Three native
message mutants retain exit70/empty stdout and are caught only by stderr bytes.

### Additional literal-port compiler gaps

- `gaps/15_string_or.ts`: Node prints `fallback` and `x`; string `||` is NotYet,
  `a BinaryExpression with a string and a string`. Production uses explicit
  comparisons/conditionals for truthy-string defaults.
- `gaps/16_array_shift.ts`: Node prints `1` and `2`. Its direct receiver call
  now receives this exact Refused diagnostic (relative path shown):

  ```text
  gaps/16_array_shift.ts:2:16: Adamic 0.1 refuses inherited library member shift read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)
  ```

  Production task-list removal keeps discarded splice(0,1).
- `gaps/17_long_optional_chain.ts`: Node prints `1`; lowering is NotYet,
  `an optional chain longer than one step`. Production checks the chunk then
  reads its text length. All three exact diagnostics and Node outputs are held.

The context source registry is readonly: a mutable array of chunk-owning source
objects is refused as potentially cyclic, even after using a nominal wrapper.
The readonly registry compiles and leaks nothing. This is the existing structural
array representation issue, not a compiler change. Classes are split into their
own files and all production code passes cohere's full 276-rule checker.

## Native Markdown printer AstPath

`astPath.ts` ports the alternating node/property/array/index path stack, using
numeric AST identities and tagged leaf frames. It implements current/root/parent/
grandparent, siblings and adjacent siblings, names/keys/indexes, array boundaries,
ancestor lookup, match predicates, scoped call/callParent, each and numeric map.
The oracle compares full stacks, callback arguments/results and restoration, not
only the final current node. Negative and out-of-range child indices, missing
properties and nesting through 32 generated quote levels are included.

The gate covers 4,975 documents: the accumulated 4,943 layout documents plus 32
path-depth cases, including all 876 physical Markdown files. Go parses and
preprocesses the input AST. Path decisions are native. Both Go's actual generic
AstPath methods and the unchanged pinned fork's actual AstPath class decide
expected observations independently from identical transported ASTs. The fork
class is exposed by reference inside its actual standalone bundle using a unique
checked anchor; its body is unchanged and the bundle bytes are checked first.
Node identity is normalized to numeric preorder IDs. JavaScript's generic getters
are projected to Go's typed Markdown node API. Children are the traversal field;
this does not claim arbitrary JavaScript object-property traversal.

The class specializes callbacks to string, number and boolean results; map
returns numeric results suitable for document IDs. Exception unwinding is not
claimed: Adamic's fatal panic is not a catchable JavaScript exception. Invalid
callParent counts beyond the root, arbitrary scalar fields, mutating the child
array during traversal and attached comment nodes are not covered. The existing
layout fixture driver still receives its path facts from Go; this independent
path implementation has not replaced that transport. Tokenizer events, grammar,
resolvers/subtokenization and full mdast construction remain unfinished.

Three native output-only mutants compile, exit zero with empty stderr and are
caught by exact path bytes: siblings frame -3 becomes -5 (byte268), the first
nodeStackIndex step -2 becomes -4 (byte191), and call retains two extra stack
frames instead of restoring its checkpoint (byte260). Source Node, JavaScript
backend, sanitized native and leaks pass the unchanged implementation.

### Missing-property key witness

The two actual AstPath oracles differ outside normal Markdown child traversal.
Starting at an empty root, call through `absent`, `children`, then numeric index
0. The complete Go output is:

```text
key="" present=false
```

The complete pinned fork output is:

```text
key=0 type=number
```

The proving programs are `testdata/path_go.go --key-gap` (calling the actual Go
class through `testdata/path_bridge.go`) and `testdata/path_library.mjs
<scratch-bundles> --key-gap` (calling the actual fork class). The gate asserts
both complete outputs. The common-domain path comparison includes a single
absent property, but does not falsely assert that this deeper witness agrees.
Native key uses Go's typed optional-string contract. All original seven embedded
language witnesses and their fourteen complete outputs below remain intact.

### Additional compiler representation witnesses

Each numbered program is executable on source Node and covered by a test:

- `gaps/8_generic_callback_result.ts`: Node prints `value`; lowering a generic
  callback-result method is NotYet, `a function returning Result`. Production
  path callbacks use concrete string, number and boolean specializations.
- `gaps/9_mixed_path_names.ts`: Node prints `children,0`; an array of string or
  number is NotYet, `an array of string | number`. Production names are tagged
  property/index frames, with numeric IDs instead of owning graph edges.
- `gaps/10_multiple_push.ts`: Node prints `1,2`; multiple push arguments are
  NotYet, `push with other than one value`. Production pushes one frame at a time.
- `gaps/11_function_expression.ts` is closed: the unchanged direct function
  expression prints `value` on source Node, native and the JavaScript backend,
  with sanitizer/leak checks. Production map now supplies a capturing function
  expression to each, rather than duplicating traversal in an explicit loop.
- `gaps/12_conditional_empty_array.ts`: Node prints `0`; the inferred empty arm
  is NotYet, `an array of never`. Production empties have explicit array types.
- `gaps/13_array_growth.ts`: Node prints `1`. Native AND the JavaScript backend
  compile but exit70 with empty stdout and the complete stderr below. The
  separate array-growth test records this discrepancy as a gap observation.
  Production map appends with push. No compiler-owned files were changed.

```text
adamic: panic: index 0 is outside an array of length 0
```

## Native tokenizer input chunking

`inputChunks.ts` ports micromark's complete single-input preprocessing call with
end=true: initial BOM removal, NUL replacement, text chunk boundaries, deferred
CR/CRLF handling, column resets, horizontal tabs and virtual spaces. Chunks own
numeric UTF-16 units, so even lone surrogates can be checked without converting
through UTF-8. The probe checks each input number is a UTF-16 unit. EOF is the
Go code 0; the actual original JavaScript oracle's null EOF is serialized as that
same code explicitly. It is not a text NUL, which preprocessing replaces.

All 79,873 cases agree with actual Go private preprocess and the pinned original
fork's unchanged micromark preprocess factory, exposed by reference in memory.
Coverage includes every one of the 65,536 units, all special-code sequences of
length one through five, every accumulated layout document, and column/BOM/NUL/
CRLF/tab cases. Native/source/backend, ASan/UBSan and leaks agree. Three native
output mutants change initial BOM, tab stops and CRLF recognition and are caught.
The original bundle bytes are checked first, including the private-function
anchor. Original regular expressions remain in that oracle, not the native port.

This primitive accepts one complete input. Streaming incremental calls, the
state/event engine, rollback, resolvers, subtokenization, full mdast construction
and printer AstPath are still unported. The existing layout driver continues to
use Go's parser/preprocess/path fixtures. No full native Markdown formatter or
full-parser speedup is claimed.

## Native six-pass AST preprocessing

`astArena.ts`, `astWalk.ts`, `astSource.ts`, `astLists.ts` and `astPreprocess.ts`
implement all six non-MDX preprocessing passes: raw text capture, continuous-text
merging, indented-code detection, original image alt text, list alignment and
sentence splitting with accidental wiki-link protection. The replacement-aware
preorder walker uses numeric child edges and shared position identities, so the
native graph has no owning pointer cycle or recursive closure. It is the mapAst
walker, not the printer's alternating node/name/array AstPath stack.

The standalone oracle starts from Go's parsed but UNPREPROCESSED AST. Native
preprocessing decides the changes. Comparison projects every field these passes
read or write, including all source coordinates and shared position identity.
It does not yet store every other mdast field, parse source Markdown, supply the
printer path predicates, or replace preprocessing in the existing layout driver.
The original oracle calls the unchanged pinned fork's exported mdast preprocess.
Its unused originalLabelText field is outside this projection, as Go drops it.
Core BOM and CR/CRLF normalization occurs before parsing, as in the layout gate.
Offsets are converted from Go's UTF-8 bytes to the source's UTF-16 code units.

Coverage includes the accumulated 4,943 layout documents, 155 generated cases
across five tab widths, and an AST with three adjacent text children. On that
synthetic AST both original preprocessors lose raw during merging and throw;
the native result preserves that observed error. The protocol compares errors
as values, prefixing the original JavaScript exception message with Go's
`markdown: ` prefix. It does not claim identical process-exit behavior.

`gaps/7_postfix_property.ts` proves that a postfix property increment used as an
expression is NotYet: Node prints 0 and 1; Adamic reports
`a PostfixUnaryExpression`. The production pass instead assigns the old counter
and then increments it in a separate statement. Compiler-owned files are untouched.
The native tokenizer/event engine, resolvers, rollback/subtokenization, full mdast
construction and printer AstPath remain unfinished.

## Native whitespace and preserved reference labels

`whitespace.ts` now makes whitespace document decisions natively: CJ/non-CJK
spacing style, Hangul pairs, punctuation, all three prose-wrap policies, link-label
mode, single-line ancestors and syntax-leading/setext/fake-whitespace exceptions.
`preservedLabel.ts` composes native splitText and document fills for the preserved
raw reference-label branch. No Go whitespace document or width service remains.
The composed fixture stream still supplies parsed/preprocessed AST facts and
path/ancestor predicates; this is not a source-file native Markdown formatter.

All 4,943 accumulated documents (953 original plus generated layout cases) match
Go, source Node, native, backend, original doc printer and original full-source
fork formatter with embedding off. Separately, 205,931 whitespace nodes in six
modes each match actual Go private routines and the unchanged original fork's
private dispatch/policy functions, exposed in memory. Three policy and three
layout output-only native mutants fail. All sanitize and leak checks pass.
Only three preserved-label and three footnote-definition frames occur in this
corpus. These counts describe coverage, not proof of every AST/options context.
The full pipeline uses preserved prose; always/never have isolated policy checks,
and compact never-mode tables remain a printer option gap.

`gaps/5_conditional_panic.ts` proves that a nested conditional ending in panic
is NotYet: Node prints T, Adamic reports `reading panic`. The whitespace probe
now uses explicit if/else statements and passes every policy oracle.
Fixture transport initially repeated every parent token per whitespace, producing
302 MB. It now transports only actual whitespace samples and neighbor kinds that
can contribute to CJ/non-CJK spacing counts, 43 MB; native still counts them.
Superseded long runs were stopped. The final complete gate passes; no native
printer decision was replaced by an oracle decision to reduce fixture size.

The native tokenizer/event engine, rollback/resolvers/subtokenization, mdast
construction and printer AST path walker remain unfinished. AST preprocessing is
now independently native, as recorded above. The
remaining-printer wording in historical sections below describes earlier slices.

## Native source escape and character-reference decoding

`decodeString.ts` ports micromark's complete source-string decoder: ASCII
punctuation escapes, semicolon-terminated named references (31-character bound),
decimal references (7 digits), hexadecimal references (6 digits), and the exact
invalid-control/surrogate/noncharacter/out-of-range replacement rules. The 2,125
case-sensitive entities are generated into separate uppercase/lowercase key and
value tables, with pinned source hashes and byte-checked regeneration.
There is no runtime regular expression or Go service. All 2,242,268 source
strings match actual Go DecodeString and the unchanged original fork function,
including every numeric code point in decimal and hexadecimal, all physical
Markdown files and generated contexts. Three output-only native mutants fail.
This closes a parser primitive; it does not implement tokenizer events or mdast.

`gaps/6_uninitialized_optional_string.ts` caught a compiler miscompile before
integration: Node printed `missing`, while both compiled backends printed
`present` because the declaration used an empty string storage placeholder.
Integration 15 initializes optional declarations with an explicit undefined IR
value. `optional_gap_test.go` now requires Node/native/backend byte parity and
leak checks; removing that initialization must fail this regression. The decoder
retains its separately verified explicit match flag.
The initial combined entity file exceeded cohere's 2,000-line limit. Splitting
case-sensitive keys and values into their own data concerns passes all 276 rules.

## Native source text splitting

`splitText.ts`, `textTokens.ts` and generated `textClasses.ts` port cohere's
source-driven prose splitting. This receives source strings directly: normalized
ASCII whitespace runs, CJK code-point/selector splitting, Korean/CJ/non-CJK
classification, punctuation flags and synthetic empty whitespace are native.
Go's actual private splitText and the actual fork's private function agree with
native/source/backend on 1,180,938 source cases, including every Unicode scalar.
The fork function is exposed in memory without changing its body or constants;
the bundle bytes are checked against the pinned repository copy first.

This is a completed preprocessing primitive. It has not yet replaced the full
AST preprocessing walk or micromark/mdast parser. Initial million-line source
Node output exited 70 with empty stderr; batching probe stdout avoided that
failure, and all complete oracle/sanitizer/leak/mutant checks then passed.
The underlying stream failure was not isolated; no semantic mismatch was observed.

## Native dispatch composed with prior inline leaves

`leaves.ts` composes the previously verified inline-code, wiki-link, URL, title,
reference and image routines and constructs emphasis, links, definitions,
footnotes, thematic breaks, explicit breaks, math, liquid and table-cell docs.
It also prints ignored nodes, including the quote-list trailing-marker rule.
All 4,813 source contexts match the complete Go and original fork formatters
with embedding off. The captured corpus has only two footnote definitions;
this is recorded coverage, not a claim that every nesting/options combination
has been exercised. Prose remains preserved with common options.

Whitespace dispatch and preserved reference-label content are now native, as
recorded above. Parent/ancestor facts
for emphasis and thematic breaks still come from the fixture adapter. The printer
AST path walker and tokenizer/mdast remain unported; AST preprocessing is now
independently native as recorded above.

## Native root and ignore layout boundary

`root.ts` prints root children, exact blank-line predicates, trailing hardline,
paired ignore ranges and raw single-node ignores. Go byte offsets become UTF-16
source offsets in the fixture adapter; native slices and prints source itself.
All 4,724 accumulated source contexts match both original full-source oracles.
Go still parses/preprocesses the AST and supplies unported children.

A two-field structural range array was rejected before code generation because
other structural subtypes can carry child arrays. The minimal witness is
`gaps/4_structural_ranges.ts`: Node prints 1; Adamic reports Refused with the
cycle reference-counting diagnostic. Root ignore ranges now use numeric index
pairs; the complete root gate compiles, runs and leaks nothing. This is a
representation workaround, not a compiler change or a reason to stop parsing.

## Native heading, sentence and paragraph boundary

`structure.ts` constructs ATX/setext headings, sentence fills and flattened
paragraph fills. Its stack of document IDs replaces the recursive flattening
closure with owned acyclic data. The bridge supplies parsed/preprocessed children
and source spans, and the native implementation constructs these three document
forms. All 4,598 source contexts match Go and the original fork with embedding
off, including 3,942 headings, 32,904 sentences and 10,639 paragraphs. This
closes these printer components under preserved prose; the tokenizer, mdast,
AST preprocessing, whitespace decisions and other child printers remain pending.

## Unicode width gap closed

`width.ts` computes cohere's StringWidth natively. It uses the pinned East Asian
ranges and finite ordered emoji matcher instructions generated from the exact Go
tables, with no runtime RegExp. Matching preserves UTF-16 code units, priority,
greedy optional sequences, narrow emojis, controls, combining marks, selectors,
and cohere's printable-ASCII shortcut, including DEL. Go supplies no text widths
to the native document printer or the list, quote, table, code and HTML APIs.
The fixture test poisons all legacy Go text-width protocol slots to prove this.

The width check covers every Unicode scalar, all repository documents and their
lines, plus generated emoji/selector/control sequences. The pinned original
emoji-regex 10.6.0, get-east-asian-width 1.6.0 and narrow-emojis 0.0.3 packages
are installed only in scratch. The actual bundled fork independently decides
width through group-fit boundaries. This closes the width-service gap, not the
full parser or the remaining printer gaps. Source-file decoding replaces lone
surrogates; this corpus does not claim a direct lone-surrogate Go StringWidth
oracle. Historical references to supplied widths below describe earlier slices.

## Native source preprocessing and quote-prefix recognition

`preprocess.ts` now ports the complete single-call micromark preprocessing stage:
UTF-16 text chunks, first-unit BOM, NUL replacement, tabs/virtual spaces, CR/LF/
CRLF and EOF. `quotePrefix.ts` recognizes quote start and continuation prefixes,
with open-container and code-indented-disabled contexts. It reports consumed
code count and UTF-16 offset/column after the attempt; failed attempts consume
nothing. Native receives actual source text, not Go AST or document fixtures.

Actual private Go preprocessing and quote constructs agree with native, source
Node and the backend on 2,091 sources and 16,728 context combinations. Tests
compare chunk boundaries/text/code values plus match/consumption/position/open
facts; they do not compare tokenizer events or claim a complete tokenizer.
The Go adapter drives real construct attempts, then drains remaining input. The
native primitive recognizes the prefix and leaves token/event emission to its
future caller. Their timings therefore measure different workloads. The bundled
fork does not expose these private chunk/prefix functions; the original public
Markdown parser/layout remains checked by the full-source embedding-off audit.

The document/container/flow dispatchers, attempt/rollback and event emission,
resolvers/subtokenization, mdast tree construction and Markdown AST preprocessing
remain unported. The literal closure representation is still rejected as below;
these native numeric/context primitives show that a redesign can make progress.
This is unfinished work, not evidence that a native Markdown parser is impossible.

## Native HTML layout boundary

`htmlblocks.ts` prints HTML values with comment hardlines and marked-root literal
lines, and scans final-root JavaScript whitespace without regex. Its inputs still
come from Go AST/path facts and exact display widths. All five requested layout
components now compose natively in the fixture driver. This does not close the
native tokenizer/mdast, AST preprocessing, Unicode-width or complete-block-printer gaps.

## Native code block layout boundary

`codeblocks.ts` prints fenced and indented values with embedding off and computes
canonical fences without regular expressions. Go supplies AST fields and explicit
text display widths; parsing and preprocessing remain unported. Embedded JSON,
YAML, TOML and other languages remain raw, as required by this unit's contract.

## Native table layout boundary

`tables.ts` builds preserved-prose table rows and renders its child documents in
native code. Cell and full-row display widths are explicit inputs from Go's
Unicode width service. The compact `proseWrap: never` alternative is not ported.
The shared test now checks the original fork's complete Markdown parser/layout
with embedding off for all accumulated source cases, as well as doc printers.
Parsing/preprocessing and other unported children remain supplied by Go fixtures.

## Native quote layout boundary

`quotes.ts` constructs quotes and their child spacing from explicit AST facts.
Its native children include lists and eligible words. The fixture bridge supplies
Go-parsed/preprocessed trees, unported child docs and display widths. Tests cover
all 953 original contexts plus list and quote cases. This is a native layout
component, not yet a native Markdown-file parser or formatter. Parser state-frame
work and Unicode widths remain as described below. No new compiler gap was found.

## The default full-document oracles disagree

Go cohere's real `native.Formatter` invokes the Markdown parser/printer and its
native embedded-language dispatch. The JavaScript library it follows is its
vendored Prettier fork, based on version 3.9.6, not unmodified npm Prettier.
`testdata/library.mjs` calls those exact standalone/plugin bundles on Node.
The audit also measured stock npm Prettier 3.9.6 separately.

On the initial census of 873 physical repository/submodule Markdown files and
75 generated block cases, Go and the pinned fork matched 941/948 documents with
embedding enabled. Seven real files disagreed: one JSONC fence and six Flow
fixtures. Stock npm Prettier matched 788/948. With embedding disabled, both
JavaScript libraries matched Go on all 948. The test repeats the census and
checks a named list of observed default-mode gaps; closing or adding a gap fails
that assertion so the report cannot silently become stale.

`gaps/embedded_jsonc.md` is a minimal proving document. Go preserves the contents
because `native/json.go` registers only `.json`, while embedded `jsonc` maps to
`embedded.jsonc`. The fork formats it and inserts a trailing comma. One output
cannot equal both. `gaps/embedded_flow.md` proves the Flow embedding difference:
the native JavaScript path does not format that Flow program as the Babel/Flow
printer in the fork does. These dependencies are outside this unit's territory.
These auto-mode differences are historical witnesses, outside the agreed
embedding-off contract. Front matter is deliberately raw with embedding off.

The audit checks both `auto` and `off` without filtering out embedded files.
The off baseline calls Go's actual `markdown.Format` with a nil embedding
callback, and normalizes/restores BOM/line endings exactly at the core boundary.
The auto baseline calls its actual `native.Formatter`. Neither adapter duplicates
Markdown parsing or printing. The fixed options on both sides are tab width 4,
print width 120, single quotes, spaces, semicolons and preserved prose, with the
remaining Prettier defaults. This is an explicit common option set, not each
nested repository's configuration or ignore rules. Every physical Markdown file
is included, even files formatting enumeration would skip.

## The literal parser representation also needs a redesign

`micromark/types.go` represents states as `func(code Code) State`, with nil meaning
finished. The document/container/flow machines are mutually referring closures
with captured mutable frames. These are additional prerequisites for a direct
Adamic port, not evidence that Markdown cannot ever be implemented in Adamic:

- `gaps/1_recursive_state.ts` prints `35` twice on Node. Stage 0 returns exactly
  `lower.NotYet("a first-class nested function reference from another nested function")`
  for returning the captured state as a first-class recursive callback.
- `gaps/2_state_arrow_cycle.ts` prints `35` twice on Node. Rewriting the nested
  declaration as a mutable arrow slot is refused by Adamic 0.1 as a closure/cell
  reference-counting cycle. This is a `Refused`, not a `NotYet`.
- `gaps/3_recursive_callable.ts` is closed: a noncapturing, top-level recursive
  callable state is supported. Source Node, native under ASan/UBSan, the
  JavaScript backend and LeakSanitizer agree on `1` then `0`. The recursive
  callable type itself is therefore not the blocker.

A full port can instead use explicit state/frame IDs and owned arenas, with
carefully bounded weak back references. That requires porting and checking the
micromark engine, every construct/extension and mdast compilation, followed by
preprocessing, AstPath and document layout. This branch does not pretend that a
line-oriented approximation is the same parser. No compiler files were modified.

## Reproduction and scope of the checks

```sh
source /workspace/adamic-tools/env.sh
(cd cohere && go build -o /tmp/cssstrings-cohere ./command/cohere)
ADAMIC_MARKDOWNBLOCKS_CENSUS=1 \
  ADAMIC_MARKDOWNWIDTH_DEPS=/tmp/adamic-markdown-width \
  ADAMIC_MARKDOWNBLOCKS_FORK=/tmp/adamic-markdown-blocks-fork \
  ADAMIC_MARKDOWNBLOCKS_KEEP=/tmp/markdown-blocks-audit \
  go test -count=1 -v -timeout=30m ./stage1/cohere/markdownblocks > /tmp/markdown-blocks-test.log 2>&1
```

`ADAMIC_MARKDOWNBLOCKS_CENSUS` widens the corpus from this unit's own Markdown files to every
Markdown file in the repository and its submodules, which is the census this file reports. It is
opt-in because that set moves with every commit and cohere bump, so the gate runs without it.
`ADAMIC_MARKDOWNWIDTH_DEPS` names an npm install of emoji-regex 10.6.0, get-east-asian-width 1.6.0
and narrow-emojis 0.0.3; without it the width oracle skips, until #xq2ecw6 installs it on the gate.

Install the original fork in that scratch directory by copying the exact
`cohere/internal/format/prettier/bundles/` tree from the pinned submodule. The
adapter checks `prettier.version === '3.9.6'`. The unit report records its pin and
bundle digests. This is an installation of the vendored original, not a guessed
patch to npm Prettier. The stock comparison uses the separately installed exact
npm version in `/tmp/adamic-markdown-prettier/node_modules/prettier`.

Set `ADAMIC_MARKDOWN_BENCH=1` for the repeated throughput measurements. The
ordinary gate still runs every corpus comparison, sanitizer/leak check, native
mutant and byte-identical data regeneration; only repeated timings are opt-in.
Layout tests admit at most two fixture streams concurrently. Output-only layout
and whitespace-policy mutants retain ASan/UBSan at `-O0`; unchanged ports retain
the normal `-O1` sanitizer run and an `-O2` release comparison. These controls
keep the expanded suite within the gate's deadline without reducing its corpus.
Identical unchanged drivers share compiled artifacts within one package run,
keyed by generated C bytes and sanitizer mode. Executions and observations are
never reused, including in an uncached gate; the temporary binaries disappear
when the package exits. A source-key mutant must fail the existing parity tests.
The layout milestones all use the same final driver on cumulative corpora. Their
unchanged baseline now runs once on the final full superset, comparing fresh Go,
source Node, backend, sanitizer, release, leak and original-library observations.
Every area's native mutants still run against that full oracle answer. Selecting
an individual layout test also runs this complete baseline; no case is omitted.

The census includes `.md`, `.markdown`, `.mdown` and `.mkd`, skipping only `.git`.
Generated cases cover heading forms, all list marker types and five nesting
depths, fenced/indented code, block quotes, tables, HTML, thematic breaks, front
matter and typed code fences. Three Go-printer mutants alter an ATX heading
prefix, unordered list marker and fence length through overlays. Each must build,
exit zero, produce no stderr or formatting errors, and fail the off baseline's
byte comparison. These prove the preflight comparison catches real output
changes; they are Go mutants, not mutants of an unbuilt Adamic block printer.

The list, quote, preserved-prose table, raw code and HTML printers and the doc
operations Markdown uses are now ported. The prior inline word printer composes
with them in the fixture driver. Source preprocessing and quote-prefix recognition
are now native primitives. The complete tokenizer/mdast, AST preprocessing,
remaining printers and Unicode display-width service are unfinished. The block
unit remains incomplete: there is no native Markdown-file formatting driver yet.
The embedding-off contract remains fixed throughout this work.

## Native front-matter parser stage

`frontmatter.ts` ports the complete `mdast.ParseFrontMatter` stage. It recognizes
both YAML and TOML delimiters, explicit languages and YAML's `...` terminator,
retains the exact raw text, and blanks the prefix one space per UTF-16 unit while
preserving LF. The upstream closing-delimiter suffix check accepts all suffixes;
that behavior is retained. ECMAScript whitespace is scanned explicitly: U+FEFF
is trimmed and U+0085 is retained. This stage receives the parser's text, before
any micromark events; it does not normalize line endings or remove BOM itself.

`testdata/frontmatter_probe.ts` exercises this stage, not full Markdown layout.
Its test walks all 876 physical Markdown files, the 77 existing generated block
cases and 3,116 new front-matter cases. Go's actual `ParseFrontMatter`, the pinned
fork's real Markdown parser, source Node, Adamic's JavaScript backend and native
agree on all 4,069 observations. Native runs under ASan/UBSan and LeakSanitizer.
The fork parser's first `frontMatter` node supplies the original library fields;
its blanked prefix is checked with the library's specified non-Unicode JavaScript
replacement. No original parsing is delegated to an oracle in production code.

Three native output-only mutants change YAML fallback selection, add one blank
space to each prefix line, and remove BOM from the language whitespace class.
Each lowers, builds and exits zero with no stderr; only its bytes fail the oracle
comparison. These are native front-matter mutants, not full block-printer mutants.
The inherited Go heading/list/fence mutants still test the whole-document audit.
At the front-matter milestone, inline printers had not yet been wired in. The
later layout harness now composes native words; other inline leaves remain
supplied by Go child documents.

## Native list layout component

`lists.ts` ports the list and list-item printers, ordered-marker scanning,
Git-friendly numbering and list alignment. It covers all marker families,
999,999,999 capping, alternating sibling markers, task boxes, tight/loose spacing,
blank lines before nested lists, required indentation before code, and the HTML
column exception. `document.ts` and `commandStack.ts` port the document operations
used by Markdown with owned arenas and numeric IDs for children and indentation
roots. Captured mutually referring state closures are not used.

The test collects real Go parser/preprocessor trees and child documents for
blocks not ported yet. Labels are transparent to Go's actual formatter; the
collector checks that fact on every input. In the native fixture, every list
label is replaced by native list construction, and every eligible word label by
the prior native word printer. The document renderer lays out the whole file.
Fixtures are serialized before Go propagates breaks, so native break propagation
is exercised. All text display widths are explicit inputs to this component;
Unicode width is still a caller service, not a newly ported native implementation.
The native component never invokes Go or Node, but its fixture driver is not a
Markdown-file formatter. Source parsing, non-list child printing, path/context
facts and inherited alignment still come from Go in the test harness. No native
CommonMark list recognition/container machine is claimed by this layout slice.

The current corpus is all 953 existing document cases plus 1,218 generated list
cases, 2,171 whole-document contexts. Generated inputs cross ten markers, six
space/tab prefixes, four task forms and five body forms, plus numbering, cap,
blank-line, nested code/quote/table/HTML, definitions and ignore edges. Native,
source Node, the JavaScript backend and the original fork's document printer
agree with actual Go Markdown formatting. Native ASan/UBSan and leaks pass.
Three native mutants change the unordered marker, checked task box and ordered
cap; each compiles, exits zero without stderr and fails only the byte comparison.
This proves list layout on Go trees, not completion of native Markdown parsing.

## Seven upstream embedding witnesses, with complete outputs

These are the seven existing repository files from the original census. Both
outputs use embedding **auto**, tab width 4, print width 120, preserved prose,
spaces, semicolons and single quotes. The Go pin is
`715ba94f3608a6500086b1076ce5cb7e51b836db`; the other output comes from its exact
vendored Prettier 3.9.6 fork. With embedding **off**, each pair agrees. The JSON
string literals below preserve every byte, including the final newline. Input
SHA-256 identifies the witness independently of later edits.

### 1. `cohere/TypeScript/packages/vscode-typescript/README.md`

Input SHA-256: `fdb86d928c1f51034f590027c1e158aaf092ddd0b587fac4f4f80ef46d2c8822`.

Go cohere, complete output as a JSON string:

````text
"# TypeScript 7\n\nThis extension provides the native implementation of the TypeScript language service. It provides features like go-to-definition, completions, errors and diagnostics, quick info/tooltip hovers, and more.\n\n## Usage\n\n1. Install the extension from the marketplace.\n2. Open a TypeScript or JavaScript file (`.ts`) in your editor.\n3. Activate the extension with the command `TypeScript: Enable TypeScript 7`, or update your settings below:\n\n## Configuration\n\nYou can enable this extension by modifying the following settings:\n\n```jsonc\n{\n    // UI Setting:\n    // TypeScript 7 > Experimental: Use Tsgo\n    \"js/ts.experimental.useTsgo\": true,\n\n    // Optional: use a local TypeScript package directory.\n    \"js/ts.tsdk.path\": \"./node_modules/typescript\"\n}\n```\n\n## Feedback\n\nIf you encounter any issues or have suggestions for improvement, please open an issue on the [GitHub repository](https://github.com/microsoft/TypeScript/tsc).\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"# TypeScript 7\n\nThis extension provides the native implementation of the TypeScript language service. It provides features like go-to-definition, completions, errors and diagnostics, quick info/tooltip hovers, and more.\n\n## Usage\n\n1. Install the extension from the marketplace.\n2. Open a TypeScript or JavaScript file (`.ts`) in your editor.\n3. Activate the extension with the command `TypeScript: Enable TypeScript 7`, or update your settings below:\n\n## Configuration\n\nYou can enable this extension by modifying the following settings:\n\n```jsonc\n{\n    // UI Setting:\n    // TypeScript 7 > Experimental: Use Tsgo\n    \"js/ts.experimental.useTsgo\": true,\n\n    // Optional: use a local TypeScript package directory.\n    \"js/ts.tsdk.path\": \"./node_modules/typescript\",\n}\n```\n\n## Feedback\n\nIf you encounter any issues or have suggestions for improvement, please open an issue on the [GitHub repository](https://github.com/microsoft/TypeScript/tsc).\n"
````

### 2. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_fn_type_mismatch_2.expect.md`

Input SHA-256: `8686b393bc7969930ea787b1a735c652ce2b029a544cd44cab152282614d0cf5`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n * @flow strict-local\n * @format\n */\n\n'use strict';\n\nimport type {SimpleTooltipMessageTypesType} from 'SimpleTooltipMessageTypes';\nimport type {Alignment, Position, Width} from 'AmbientTooltip.react';\n\nimport * as SimpleTooltipMessage from 'SimpleTooltipMessage';\nimport AmbientTooltip from 'AmbientTooltip.react';\n\nimport * as React from 'react';\n\nexport type NUXProps = {\n  alignment?: Alignment,\n  children: React.Node,\n  customWidth?: number,\n  disabled?: boolean,\n  hideOnXout?: boolean,\n  position: Position,\n  showOnce?: boolean,\n  type: SimpleTooltipMessageTypesType,\n  width?: Width,\n};\n\ntype CurrentState = {\n  showNux: boolean,\n};\n\ntype ExposedProps<Props extends {...}> = {\n  ...$Exact<Props>,\n  nuxProps: NUXProps,\n};\n\nexport default function WidgetWithTooltip<\n  Props extends {...},\n  WidgetWithTooltipComponent extends React.ComponentType<Props>,\n>(\n  WrappedComponent: WidgetWithTooltipComponent\n): Class<\n  React.Component<\n    ExposedProps<React.ElementConfig<WidgetWithTooltipComponent>>,\n    CurrentState,\n  >,\n> {\n  class WithNux extends React.PureComponent<ExposedProps<Props>, CurrentState> {\n    state: CurrentState = {\n      showNux:\n        this.props.nuxProps.disabled !== true &&\n        !SimpleTooltipMessage.hasUserSeenMessage_LEGACY(\n          this.props.nuxProps.type\n        ),\n    };\n\n    wrappedRef: {\n      current: HTMLSpanElement | null,\n      ...\n    } = React.createRef();\n\n    componentDidMount(): void {\n      if (this.props.nuxProps.showOnce === true && this.state.showNux) {\n        SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n      }\n    }\n\n    #onNuxClose = (): void => {\n      if (this.props.nuxProps.hideOnXout === true && this.state.showNux) {\n        SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n      }\n      this.setState({\n        showNux: false,\n      });\n    };\n\n    #getRef = (): null | HTMLSpanElement => this.wrappedRef.current;\n\n    render(): React.MixedElement {\n      const {nuxProps, ...passProps} = this.props;\n      return (\n        <>\n          <span className=\"uiContextualLayerParent\" ref={this.wrappedRef}>\n            <WrappedComponent {...passProps} />\n          </span>\n          {this.state.showNux ? (\n            <AmbientTooltip\n              alignment={nuxProps.alignment}\n              children={nuxProps.children}\n              contextRef={this.#getRef}\n              customwidth={nuxProps.customWidth}\n              onCloseButtonClick={this.#onNuxClose}\n              position={nuxProps.position}\n              shown={this.state.showNux}\n              width={nuxProps.width}\n            />\n          ) : null}\n        </>\n      );\n    }\n  }\n  return WithNux;\n}\n\n```\n\n## Error\n\n```\nUnexpected token (32:33)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n * @flow strict-local\n * @format\n */\n\n'use strict';\n\nimport type { SimpleTooltipMessageTypesType } from 'SimpleTooltipMessageTypes';\nimport type { Alignment, Position, Width } from 'AmbientTooltip.react';\n\nimport * as SimpleTooltipMessage from 'SimpleTooltipMessage';\nimport AmbientTooltip from 'AmbientTooltip.react';\n\nimport * as React from 'react';\n\nexport type NUXProps = {\n    alignment?: Alignment,\n    children: React.Node,\n    customWidth?: number,\n    disabled?: boolean,\n    hideOnXout?: boolean,\n    position: Position,\n    showOnce?: boolean,\n    type: SimpleTooltipMessageTypesType,\n    width?: Width,\n};\n\ntype CurrentState = {\n    showNux: boolean,\n};\n\ntype ExposedProps<Props: { ... }> = {\n    ...$Exact<Props>,\n    nuxProps: NUXProps,\n};\n\nexport default function WidgetWithTooltip<Props: { ... }, WidgetWithTooltipComponent: React.ComponentType<Props>>(\n    WrappedComponent: WidgetWithTooltipComponent,\n): Class<React.Component<ExposedProps<React.ElementConfig<WidgetWithTooltipComponent>>, CurrentState>> {\n    class WithNux extends React.PureComponent<ExposedProps<Props>, CurrentState> {\n        state: CurrentState = {\n            showNux:\n                this.props.nuxProps.disabled !== true &&\n                !SimpleTooltipMessage.hasUserSeenMessage_LEGACY(this.props.nuxProps.type),\n        };\n\n        wrappedRef: {\n            current: HTMLSpanElement | null,\n            ...\n        } = React.createRef();\n\n        componentDidMount(): void {\n            if(this.props.nuxProps.showOnce === true && this.state.showNux) {\n                SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n            }\n        }\n\n        #onNuxClose = (): void => {\n            if(this.props.nuxProps.hideOnXout === true && this.state.showNux) {\n                SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n            }\n            this.setState({\n                showNux: false,\n            });\n        };\n\n        #getRef = (): null | HTMLSpanElement => this.wrappedRef.current;\n\n        render(): React.MixedElement {\n            const { nuxProps, ...passProps } = this.props;\n            return (\n                <>\n                    <span className=\"uiContextualLayerParent\" ref={this.wrappedRef}>\n                        <WrappedComponent {...passProps} />\n                    </span>\n                    {this.state.showNux ? (\n                        <AmbientTooltip\n                            alignment={nuxProps.alignment}\n                            children={nuxProps.children}\n                            contextRef={this.#getRef}\n                            customwidth={nuxProps.customWidth}\n                            onCloseButtonClick={this.#onNuxClose}\n                            position={nuxProps.position}\n                            shown={this.state.showNux}\n                            width={nuxProps.width}\n                        />\n                    ) : null}\n                </>\n            );\n        }\n    }\n    return WithNux;\n}\n```\n\n## Error\n\n```\nUnexpected token (32:33)\n```\n"
````

### 3. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_loc_diff_1.expect.md`

Input SHA-256: `f40f1db9e8a46fc284bee4df1ac6bccb6cfefdec0b9df25159bb7c8211514fb6`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n *  * @flow strict\n * @format\n */\n\n/**\n * creates a cache for Component props so we prevent rendering a component\n * sequentially if the props didn't change. Useful to wrap FluxContainer\n * with pure calculateState and getStores functions.\n */\n\n'use strict';\n\nimport * as React from 'react';\nimport {PureComponent} from 'react';\n\nexport default function createPureComponent<\n  DefaultProps extends {...} | void,\n  Props extends {...},\n>(\n  Component: React.ComponentType<Props> & {\n    defaultProps?: DefaultProps,\n    displayName?: string,\n  }\n): React.ComponentType<Props> {\n  class PureComponentCache extends PureComponent<Props, void> {\n    static defaultProps: DefaultProps;\n\n    render(): React.MixedElement {\n      return <Component {...this.props} />;\n    }\n  }\n\n  if (Component.defaultProps) {\n    PureComponentCache.defaultProps = Component.defaultProps;\n  }\n  PureComponentCache.displayName = `PureComponentCache(${\n    /* $FlowFixMe[incompatible-type] (>=0.66.0 site=www) This comment\n     * suppresses an error found when Flow v0.66 was deployed. To see the\n     * error delete this comment and run Flow. */\n    Component.displayName\n  })`;\n  // $FlowFixMe[incompatible-type]\n  return PureComponentCache;\n}\n\n```\n\n## Error\n\n```\nUnexpected token (18:24)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n *  * @flow strict\n * @format\n */\n\n/**\n * creates a cache for Component props so we prevent rendering a component\n * sequentially if the props didn't change. Useful to wrap FluxContainer\n * with pure calculateState and getStores functions.\n */\n\n'use strict';\n\nimport * as React from 'react';\nimport { PureComponent } from 'react';\n\nexport default function createPureComponent<DefaultProps: { ... } | void, Props: { ... }>(\n    Component: React.ComponentType<Props> & {\n        defaultProps?: DefaultProps,\n        displayName?: string,\n    },\n): React.ComponentType<Props> {\n    class PureComponentCache extends PureComponent<Props, void> {\n        static defaultProps: DefaultProps;\n\n        render(): React.MixedElement {\n            return <Component {...this.props} />;\n        }\n    }\n\n    if(Component.defaultProps) {\n        PureComponentCache.defaultProps = Component.defaultProps;\n    }\n    PureComponentCache.displayName = `PureComponentCache(${\n        /* $FlowFixMe[incompatible-type] (>=0.66.0 site=www) This comment\n         * suppresses an error found when Flow v0.66 was deployed. To see the\n         * error delete this comment and run Flow. */\n        Component.displayName\n    })`;\n    // $FlowFixMe[incompatible-type]\n    return PureComponentCache;\n}\n```\n\n## Error\n\n```\nUnexpected token (18:24)\n```\n"
````

### 4. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-pattern3_type_to_poly.expect.md`

Input SHA-256: `09e9ebd9520468174c33b4479402d511f08d097ebe2611efd5db961961b4da05`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Pattern 3: shapeId null→generated + return Type→Poly\n// Generic function with Object.entries().reduce()\n// Divergence: TS has shapeId:null, return:Type(32); Rust has shapeId:\"<generated_1>\", return:Poly\n\n/**\n * @flow strict\n */\nexport default function flipAndAggregateObject<TValue extends string>(obj: {\n  +[key: TKey]: TValue,\n  ...\n}): {} {\n  return Object.entries(obj).reduce((acc, [currKey, currVal]) => {\n    return {};\n  }, {});\n}\n\n```\n\n## Error\n\n```\nUnexpected token (9:2)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Pattern 3: shapeId null→generated + return Type→Poly\n// Generic function with Object.entries().reduce()\n// Divergence: TS has shapeId:null, return:Type(32); Rust has shapeId:\"<generated_1>\", return:Poly\n\n/**\n * @flow strict\n */\nexport default function flipAndAggregateObject<TValue: string>(obj: { +[key: TKey]: TValue, ... }): {} {\n    return Object.entries(obj).reduce((acc, [currKey, currVal]) => {\n        return {};\n    }, {});\n}\n```\n\n## Error\n\n```\nUnexpected token (9:2)\n```\n"
````

### 5. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_identifier_diff.expect.md`

Input SHA-256: `53c0cd7ff1745679ace6a7ee83e9470eea3f6f2b9b82cccfaeeccbec8f4d8be5`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: IDENTIFIER_DIFF (11 files)\n// Extra/different context identifiers — class components with this/setState\n/**\n * @flow strict-local\n */\nexport default function withRemountOnChange<OuterProps extends {}>(\n  shouldRemount: () => boolean\n): () => React.ComponentType<OuterProps> {\n  return function withRemountOnChangeInner(WrappedComponent) {\n    return class Wrapper extends React.Component<OuterProps, WrapperState> {\n      static displayName: ?string = `withRemountOnChange(${getDisplayName()})`;\n      state: WrapperState = {};\n      componentDidUpdate(prevProps: OuterProps, prevState: WrapperState) {\n        if (shouldRemount()) {\n          this.setState(({keyId}) => {});\n        }\n      }\n      render(): React.MixedElement {}\n    };\n  };\n}\n\n```\n\n## Error\n\n```\nUnexpected token (11:26)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: IDENTIFIER_DIFF (11 files)\n// Extra/different context identifiers — class components with this/setState\n/**\n * @flow strict-local\n */\nexport default function withRemountOnChange<OuterProps: {}>(\n    shouldRemount: () => boolean,\n): () => React.ComponentType<OuterProps> {\n    return function withRemountOnChangeInner(WrappedComponent) {\n        return class Wrapper extends React.Component<OuterProps, WrapperState> {\n            static displayName: ?string = `withRemountOnChange(${getDisplayName()})`;\n            state: WrapperState = {};\n            componentDidUpdate(prevProps: OuterProps, prevState: WrapperState) {\n                if(shouldRemount()) {\n                    this.setState(({ keyId }) => {});\n                }\n            }\n            render(): React.MixedElement {}\n        };\n    };\n}\n```\n\n## Error\n\n```\nUnexpected token (11:26)\n```\n"
````

### 6. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_severity_diff.expect.md`

Input SHA-256: `4e4b4b0a786f3ae7de1266c8e4786b107cb0aa2edc753debe2f755aa445f764b`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: SEVERITY_DIFF (1 file from OTHER)\n// delete on optional chain — TS: severity Error/Syntax, Rust: severity Hint/Todo\n/**\n * @flow strict-local\n */\nimport {} from 'PreloadingTTL';\nconst preloadedRequests: Map<> = new Map();\nexport function execute(): Promise<{error?: APIErrorEventArgs['error'], ...}> {\n  if (request.params != null && !(request.params instanceof FormData)) {\n    delete request.params?.__entryPointPreloaded;\n  }\n  if (!consumers) {\n    if (APIRequestMatchingUtils.areRequestsEquivalent()) {\n    }\n  }\n}\n\n```\n\n## Error\n\n```\nType argument list cannot be empty. (7:28)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: SEVERITY_DIFF (1 file from OTHER)\n// delete on optional chain — TS: severity Error/Syntax, Rust: severity Hint/Todo\n/**\n * @flow strict-local\n */\nimport {} from 'PreloadingTTL';\nconst preloadedRequests: Map<> = new Map();\nexport function execute(): Promise<{ error?: APIErrorEventArgs['error'], ... }> {\n    if(request.params != null && !(request.params instanceof FormData)) {\n        delete request.params?.__entryPointPreloaded;\n    }\n    if(!consumers) {\n        if(APIRequestMatchingUtils.areRequestsEquivalent()) {\n        }\n    }\n}\n```\n\n## Error\n\n```\nType argument list cannot be empty. (7:28)\n```\n"
````

### 7. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-update-expression-context-variable-via-type-annotation.expect.md`

Input SHA-256: `cd4a073e669a886fc32f65b5ec66b11ae6fb3f2972e9074936ca0c9bc222ccea`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// @flow @compilationMode(infer)\nfunction Component(props: {data: Array<[string, mixed]>}) {\n  let id = 0;\n  for (const [key, value] of props.data) {\n    const item = {\n      key,\n      id: '' + id++,\n    };\n  }\n  const getIndex = ((): ((id: string) => number) => {\n    return (id: string): number => 0;\n  })();\n  return <div />;\n}\n\n```\n\n## Error\n\n```\nFound 1 error:\n\nTodo: (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n\n   5 |     const item = {\n   6 |       key,\n>  7 |       id: '' + id++,\n     |                ^^^^ (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n   8 |     };\n   9 |   }\n  10 |   const getIndex = ((): ((id: string) => number) => {\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// @flow @compilationMode(infer)\nfunction Component(props: { data: Array<[string, mixed]> }) {\n    let id = 0;\n    for(const [key, value] of props.data) {\n        const item = {\n            key,\n            id: '' + id++,\n        };\n    }\n    const getIndex = ((): ((id: string) => number) => {\n        return (id: string): number => 0;\n    })();\n    return <div />;\n}\n```\n\n## Error\n\n```\nFound 1 error:\n\nTodo: (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n\n   5 |     const item = {\n   6 |       key,\n>  7 |       id: '' + id++,\n     |                ^^^^ (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n   8 |     };\n   9 |   }\n  10 |   const getIndex = ((): ((id: string) => number) => {\n```\n"
````
