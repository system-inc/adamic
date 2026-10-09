# YAML parser, printer and file driver

The complete Adamic YAML formatting slice is implemented at Go cohere's default options.
Go, sanitized native, source Node and emitted JavaScript compare all 36 repository files and generated cases.
Original Prettier 3.9.6 is compared independently, with 42 proved upstream binary/minification differences.
Thirty-one successful wrong-output port mutants are caught; nine active compiler refusals have proving programs; the former runtime bug and two closed presence gaps have parity regression tests.
Final throughput: native 11,504, source Node 13,819, Go 15,190 texts/s; native is 4.10 times its original baseline.

## Presence gaps closed on library 45d70a05

[stringPresence.ts](gaps/stringPresence.ts) and
[valuePresence.ts](gaps/valuePresence.ts) now lower and agree with Node.
Both closures trace to compiler change `1a9c637cf3556f14462dcf8798cb41ff30bfbb4c`,
"Represent nullable references with one pointer and a static empty case",
which lowers string negation through its length and reference negation through
its empty case. This change is present in the library area.

The lexer uses direct `!unit`, `!nextUnit` and `!character(...)` tests again,
and scalar block-line checks use `!line`. Positive presence comparisons and
string fallback workarounds remain; the separately tracked logical gaps are
still open.
The two original probes remain as `TestClosedLexerPresenceGapsMatchNode`, which
compares Node with sanitized native and emitted JavaScript. `TestLexerGaps`
continues to hold only active refusals. The lexer mutant that waits on an empty
character even at end of input must be caught by token-byte comparisons.

Validation: lexer parity passed on 8,732 complete/chunked cases (4,182,610
answer bytes), and scalar parity passed on 9,272 cases (8,419,050 answer bytes).
Both compare Go, sanitized native, source Node 24.19.0, emitted JavaScript and
yaml 2.9.0. All six lexer and three scalar mutants were caught. The new
empty-character EOF mutant exits successfully with clean stderr; native and
Node token-byte comparisons both catch its missing output at byte 860236.
Vet, gofmt, and cohere lint and formatting checks pass for the changed ports.
The full formatter suite was not rerun for this change.

```sh
ADAMIC_YAML_LIBRARY=/tmp/library-yaml-gap-deps go test ./stage1/cohere/yaml -run 'TestLexerGaps|TestClosedLexerPresenceGapsMatchNode|TestLexerMatchesGo|TestLexerMutants|TestScalarMatchesGo' -count=1 -v -timeout 15m > /tmp/project-ts-imports-yaml-after.log 2>&1
ADAMIC_YAML_LIBRARY=/tmp/library-yaml-gap-deps go test ./stage1/cohere/yaml -run 'TestScalarsMatchGo|TestScalarMutants' -count=1 -v -timeout 15m > /tmp/project-ts-imports-yaml-scalars.log 2>&1
go vet ./stage1/cohere/yaml > /tmp/project-ts-imports-yaml-vet.log 2>&1
```

## Main integration follow-up

On Linux, after merging main, the shared-slice append gap is closed and its
flow-folding workaround is removed. The proving program is retained as
`TestSharedSliceAppendMatchesNode`, covering release and sanitized native
execution at both offsets with `ASAN_OPTIONS=detect_leaks=1`.

The complete package passed in 250.076s with no skips and all 30 wrong-output
mutants caught. Comparisons retain their exact byte checks and named upstream
Prettier differences. The single npm prefix contains yaml 2.9.0, Prettier 3.9.6
and yaml-unist-parser 3.2.0; this package does not require the TypeScript source
oracle. Vet and gofmt are clean; cohere lint and formatting pass for the changed
source. [Full package log](audit/yaml-2-suite.log).

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=30m ./stage1/cohere/yaml
go vet ./stage1/cohere/yaml
gofmt -l stage1/cohere/yaml
```

The initial cold-cache run failed because the strict empty-stderr check saw Go
oracle dependency download messages. After caching the dependencies, the full
package was rerun successfully without changing that check.

## Speed follow-up

The native formatter now shares its compiled schemas and avoids impossible
schema/emoji matches. [SPEED.md](SPEED.md) records the Linux sampling profile,
all new mutants, exhaustive Unicode widths and minimal runtime/compiler cost
programs. The throughput section below retains the original formatter baseline;
the final speed comparison is reported in SPEED.md. Native clears the original
Node target but still takes 20.1% longer than the optimized Node port. The six
new qualifying speed mutants and three cost programs are documented there.
[negativeCase.ts](gaps/negativeCase.ts) adds an eleventh compiler-refusal proof:
Node prints `1`; Adamic refused `case -1` as a nonconstant case until
compiler/area-next, which lowers it; the lexer still uses an explicit numeric
predicate instead.

## Original formatter throughput

Five interleaved fresh-process rounds reproduce all 992,282 Go answer bytes.
Native release measures 2,768.29 texts/s, source Node 7,300.26 and Go 14,894.45;
native trailed both at that checkpoint. [PERFORMANCE.md](PERFORMANCE.md) gives full methodology,
commands, limitations and setup timings; [timing log](audit/format-timing.log)
contains every sample. No concurrent tests ran during measurement.

## Complete formatter checkpoint

`format.ts` composes lexer, CST, scalar/property resolution, schema and document
composition with the unist tree and `printer.ts`. `layout.ts` ports the YAML-reachable
Go document algebra: groups, fills, conditional breaks, suffixes, indentation,
Unicode display widths and break propagation. Numeric arenas keep ownership acyclic.
`main.ts <file>` writes exact formatted bytes to stdout, including BOM-only files,
empty output and keep-chomping output without a final newline. Its `--cases` mode
provides the escaped batch comparison protocol. The printer's document entry can
be composed by a future Markdown front-matter formatter.

The Go adapter calls the public `native.Formatter.Format` with `PrettierDefaults`,
rather than independently reimplementing file normalization. An independent BOM-only
Prettier probe caught a mistake in both the first port wrapper and its initial Go
adapter. Both were corrected before this checkpoint: a BOM is restored even when
the formatted body is empty. The original-library adapter always calls Prettier,
including empty and whitespace inputs. The direct-driver controls hold this rule.

The formatter corpus contains 10,026 cases: all 36 repository/submodule YAML files,
parser cases and generated printer cases covering width boundaries, Unicode,
quotes, comments, ignore directives, mapping styles, flow punctuation, block
indentation/chomping, directives, CRLF/CR and BOM. Direct-file comparisons add
14 stdout controls to the 36 repository files. Native uses ASan, UBSan and leak
detection; source Node and emitted JavaScript must produce the same Go bytes.

Published Prettier matches 9,984 cases. Exactly six explicit binary cases differ
because its bundled browser decoder rejects strings Go and yaml 2.9.0 accept;
36 prototype-tag cases differ solely in the thrown variable name (`c.resolve`
versus `tag.resolve`). `gaps/bundledParser.mjs` independently proves both categories
at pinned versions. Every allowed input and both outputs are held explicitly;
any other difference fails. This is an observed upstream discrepancy, not a claim
of byte identity to published Prettier on those 42 cases.

The three printer mutations are root newline suppression, loss of the colon's
following space, and loss of broken-flow trailing commas. Each compiles and exits
zero with empty stderr on native and Node; only byte comparison catches it.
Earlier sections retain every parser-layer mutant and the compiler gap programs.

Final corrected formatter/gap/driver/mutant suite: PASS, 139.079s;
992,282 identical format answer bytes, baseline 64.49s, driver 36.40s,
mutants 38.07s. Printer mutant differences begin at bytes 28293, 1694 and
783469, respectively, on both native and Node. The full YAML package passed
in 384.721s before the BOM-only correction; the focused suite above revalidated
all corrected formatting paths afterward. The uncached filtered oracle passed
in 2.421s, with 27 native and 18 Node cache misses and zero hits. Vet and gofmt
produced empty logs; cohere reported 276 rules, 42 files, 100% Adamic-ready,
and the formatting check passed.

Logs: [corrected formatter suite](audit/format-suite.log),
[full YAML suite](audit/final-suite.log), [uncached oracle](audit/final-oracle.log),
[lint](audit/format-lint.log), [format check](audit/format-check.log),
[vet](audit/final-vet.log), [gofmt](audit/final-gofmt.log).


Validation commands (output redirected to the linked audit logs):

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-format-artifacts go test -v -count=1 -timeout=30m ./stage1/cohere/yaml
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-format-artifacts go test -v -count=1 -timeout=20m ./stage1/cohere/yaml -run 'TestFormatter|TestFileDriver|TestBundledParser'
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=15m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$'
go vet ./stage1/cohere/yaml
gofmt -l stage1/cohere/yaml
/tmp/stage1-yaml-cohere --no-fix --no-format stage1/cohere/yaml/*.ts
/tmp/stage1-yaml-cohere --no-fix --format-only --format-all stage1/cohere/yaml/*.ts
```

Scope: default width 80, tab width 2, double quotes and preserved prose wrapping.
Repository option discovery, other formatting options, malformed UTF-8 files,
Markdown front matter and YAML semantic object loading are not covered. Timestamp
and binary semantic payloads are not materialized; the formatter does not read
them. No compiler/runtime implementation changed. The complete repository gate
was not run; the touched package and uncached filtered oracle were run instead.
Historical checkpoint sections below describe coverage at their commit dates;
their statements about unfinished later stages are superseded by this checkpoint.

## Green printer-tree step

`unistContext.ts` ports yaml-unist-parser 3.2.0's scalar, collection and document
transforms, comment attachment, parent assignment and position expansion.
Numeric node, point and position arenas retain shared position and point
identity while keeping ownership acyclic. Parsing uses whole text, as the
original unist API does; incremental CST parsing is tested separately.

36 repository files and 9,216 cases produced 5,940,046 identical answer bytes
on Go, native with ASan/UBSan/LeakSanitizer, source Node, emitted JavaScript,
and original yaml-unist-parser 3.2.0. Output includes tree fields, attached and
root comments, parent spans, parse failures and scalar metadata. UTF-16 offsets
are compared through Go's documented byte-offset mapping: an error ending
inside an astral pair maps to that pair's beginning; line and column stay in
UTF-16 units. This conversion is explicit in both comparison adapters.

The final tree/gap/mutation suite passed in 72.994s. All three mutations compile
and exit zero with empty stderr on native and Node; only comparison catches
comment prefix loss (byte 9609), shifted columns (14), and a lost document end
marker (1974423). Cohere lint/types reported 276 rules and 100% Adamic-ready.

[structuralPosition.ts](gaps/structuralPosition.ts) prints `1` on Node and
gets `lower.Refused` before clang. A structural `{ start, end }` position type
also admits objects with diagnostic arrays reaching another such position.
The cycle check consequently refuses copying caller-owned diagnostic arrays.
The port uses distinct `startPoint` and `endPoint` names for point indices.
This is a conservative 0.1 structural refusal, not an observed memory bug.
`TestStructuralPositionRefusal` holds the exact refusal and Node result.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run 'TestUnist|TestStructuralPosition' > /tmp/stage1-yaml-unist-final.log 2>&1
```

Logs: [suite](audit/unist-suite.log), [lint](audit/unist-lint.log).
The printer and file driver are written and undergoing formatted-byte checks;
they are not yet a green checkpoint. No compiler or runtime file changed.

## Green composer step

`composer.ts`, `directives.ts`, `composedNode.ts` and `composedDocument.ts`
compose document streams, block and flow collections, pairs, aliases, tag and
anchor properties, schema tests, and exact errors and warnings. `compose_main.ts`
compares the formatter-facing tree fields, presence and ranges. Its comparison
intentionally excludes scalar semantic values: date and binary payloads are not
implemented, and the printer does not read them. This is not a general YAML
object loader.

36 repository files and 9,216 cases produced 13,235,527 identical answer bytes
on Go, native under ASan/UBSan/LeakSanitizer, source Node, emitted JavaScript,
and original yaml 2.9.0. The added 484 cases cover explicit core and YAML 1.1
tags, sets, pairs, ordered maps, prototype properties, tag URI decoding,
document boundaries and the 1,024-character implicit key limit. Baseline passed
in 50.985s; the mutation suite passed in 24.322s. All three mutations compiled
and exited zero with empty stderr on native and Node; only comparison caught
lost document start markers (byte 241655), lost mapping values (6896), and
skipped implicit key limits (13235125).

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run TestComposeMatchGo > /tmp/stage1-yaml-compose-test.log 2>&1
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run TestComposeMutants > /tmp/stage1-yaml-compose-mutants.log 2>&1
go vet ./stage1/cohere/yaml > /tmp/stage1-yaml-compose-vet.log 2>&1
```

Logs: [comparison](audit/compose-suite.log), [mutants](audit/compose-mutants.log),
[lint](audit/compose-lint.log). Unist conversion is being implemented; printer,
formatting driver and formatter throughput are not yet complete.

## Green schema-test step

`schemaPattern.ts` implements only the fixed anchored tests that core and YAML
1.1 schema resolution require: alternatives, concatenation, ASCII character
classes and bounded/unbounded repetition. `schemaTags.ts` carries the exact
patterns and ordering from pinned Go cohere. This avoids dependence on RegExp,
which is outside the current Adamic library, without changing the tag tests.

9,388 inputs produced 271,142 identical answer bytes on Go, native under
ASan/UBSan/LeakSanitizer, source Node, emitted JavaScript and original yaml 2.9.0.
The inputs include all 36 repository files, the earlier generated corpus and
656 targeted schema spellings. The standalone pattern driver trims its input;
this specifically tests trimmed candidate values. It does not establish parity
for untrimmed quoted scalar values or numeric/date/binary value resolution.

The schema comparison and three successful wrong-output mutants passed in
7.390s. Native and Node both caught class-range endpoint loss at byte 260024,
plus repetition accepting empty at byte 968, and optional repetition accepting
twice at byte 262914. Cohere reported 276 rules and 100% Adamic-ready; package
vet passed. No compiler or cohere implementation file changed.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run TestSchema > /tmp/stage1-yaml-schema-final.log 2>&1
```

Logs: [suite](audit/schema-suite.log), [lint](audit/schema-lint.log).
Document composition passed the later checkpoint above.
The printer, comment attachment and formatting driver remain unfinished.

## Green property step

`propsResolver.ts` and `props.ts` port property resolution: tag and anchor
presence, whitespace requirements, tab indentation, comment and newline
folding, document and collection indicators, comma presence, blank-line flags,
and exact starts and ends. Ambiguous anchors retain the original warning.
`props_main.ts` resolves the property runs of every CST input with document,
block collection and flow collection contexts. This does not compose documents.

All 8,732 parser inputs, including the 36 repository YAML files, produce
1,182,227 identical answer bytes on Go, native under ASan/UBSan/LeakSanitizer,
source Node, emitted JavaScript and pinned yaml 2.9.0. The property/gap/mutant
suite passed in 20.930s. Cohere passed 276 rules on all three new files, 100%
Adamic-ready; `go vet ./stage1/cohere/yaml` passed.

Three more mutants compile and exit zero with empty stderr on native and Node.
Only the byte comparisons catch them:

| Mutation | First differing output byte |
| --- | ---: |
| Collection indicator presence lost | 63 |
| Anchor and tag separation unchecked | 309370 |
| Tab indentation accepted | 318861 |

[dynamicCase.ts](gaps/dynamicCase.ts) prints `1` on Node. It refused before
clang with `lower.NotYet: a case that isn't a constant` until compiler/area-next,
which lowers it. The port still compares the token type to the contextual
indicator before switching over constant types.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run 'TestProps|TestLexerGaps' > /tmp/stage1-yaml-props-final.log 2>&1
/tmp/stage1-yaml-cohere --no-fix stage1/cohere/yaml/props.ts stage1/cohere/yaml/propsResolver.ts stage1/cohere/yaml/props_main.ts > /tmp/stage1-yaml-props-lint.log 2>&1
go vet ./stage1/cohere/yaml > /tmp/stage1-yaml-props-vet.log 2>&1
```

Logs: [suite](audit/props-suite.log), [lint](audit/props-lint.log).
No compiler or cohere implementation files changed. Formatting, its file driver
and its throughput remain unfinished; the existing speed measurements are
explicitly lexer/scalar comparison-driver measurements.

## Green scalar step

`scalar.ts` resolves plain, single-quoted, double-quoted, literal and folded
scalars. Its supporting classes have separate files. Resolution preserves
UTF-16 values, lone surrogates from escapes, CRLF behavior, indentation,
chomping, comments, all three range positions, and exact error codes, messages
and positions. `scalar_main.ts` walks each CST and resolves scalars under both
root contexts. This is a comparison driver, not a formatter.

The 8,732 parser inputs plus 540 targeted scalar texts produced 8,415,462
identical answer bytes on Go, native with ASan/UBSan/LeakSanitizer, source Node,
emitted JavaScript and original yaml 2.9.0. The targeted cases include every
escape, invalid hex and code points, lone surrogates, folded whitespace,
header permutations, more-indented content, blank lines and all chomping modes.
The focused suite passed in 28.285s. The complete lexer/CST/scalar/gap/mutant
package passed in 68.537s; its log is [all parser layers](audit/parser-layers-suite.log). Cohere passed 276 rules on seven files;
`go vet ./stage1/cohere/yaml` passed.

Three further mutants compile, exit zero and have empty stderr on native and
Node. Only exact answer comparison catches them:

| Mutation | First differing output byte |
| --- | ---: |
| Escaped line feed becomes carriage return | 1431265 |
| Strip chomping adds a final newline | 1390653 |
| Flow line break becomes newline rather than space | 1395121 |

Additional compiler gap programs, held by `TestLexerGaps`:

| Program | Node stdout | Observed `lower.NotYet` | Workaround |
| --- | --- | --- | --- |
| [stringFallback.ts](gaps/stringFallback.ts) | one space | Closed by compiler/area-next | Explicit empty-string comparison |
| [valuePresence.ts](gaps/valuePresence.ts) | `false` | Closed by `cdddfee346e70a5cba3e982f808b9ebe426a6dfb` | Direct optional-object truthiness restored in scalar block-line iteration |
| [valueConjunction.ts](gaps/valueConjunction.ts) | `true` | Lowers on compiler/area-next, but emitting its C panics (`native: no conversion from 4 to 9`) | Separate presence and value branches |
| [multiplePush.ts](gaps/multiplePush.ts) | `2` | `push with other than one value` | Push one value per call |

### Closed native runtime gap

[sharedSliceAppend.ts](gaps/sharedSliceAppend.ts) previously printed `x\nx\n`
on release native instead of Node's `a\nx\n`; the end-of-owner slice also
triggered an ASan heap-buffer-overflow. Main now preserves the owner and safely
appends to both slices. `TestSharedSliceAppendMatchesNode` retains the proving
program as a regression test, requiring exact Node bytes and clean sanitized
execution for offsets 0 and 48, with leak detection enabled.

The flow-folding workaround (collecting pieces in an array and joining once)
is removed from `scalarText.ts`; folding now appends directly to its initial
string slice. This gap is closed; no compiler or runtime implementation is
changed by the YAML follow-up.

### Scalar driver throughput

Five interleaved fresh-process rounds. Every measured answer equals Go, with
zero exit and empty stderr. Startup, reading, chunked CST parsing, resolving
both contexts and writing 8,415,462 answer bytes are included; compilation is
excluded. These numbers do not measure the unfinished formatter.

| Driver | Median texts/s |
| --- | ---: |
| Native release | 6780.09 |
| Port source on Node | 12923.48 |
| Go cohere | 22352.18 |
| Original yaml 2.9.0 on Node | 9123.67 |

Native is 3.30 times slower than Go and 1.91 times slower than source Node.
Go uses buffered output; the port uses `console.log`.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-artifacts go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run 'TestScalar|TestSharedSliceAppendGap|TestLexerGaps' > /tmp/stage1-yaml-scalars-final.log 2>&1
go run ./cmd/adamic build stage1/cohere/yaml/scalar_main.ts -o /tmp/stage1-yaml-artifacts/native-scalar > /tmp/stage1-yaml-scalar-release-build.log 2>&1
python3 stage1/cohere/yaml/benchmark_lexer.py /tmp/stage1-yaml-artifacts /tmp/stage1-yaml-library --layer scalar > /tmp/stage1-yaml-scalar-timing.log 2>&1
```

The Go test overlay adds an adapter to an unchanged Go source snapshot, calling
cohere's actual private scalar resolvers. The original library oracle imports
its private resolver modules at the installed pinned path; no port code is
shared with either oracle. Logs: [suite](audit/scalar-suite.log),
[throughput](audit/scalar-timing.log), [lint](audit/scalar-lint.log).

## Green CST step

`cstParser.ts`, `cst.ts` and `collectionItem.ts` port the CST parser, token
classification and collection item representation. `cst_main.ts` serializes the
entire tree with field presence, UTF-16 sources and offsets, and line starts.
Children are numeric indexes into a parser-owned token arena. Optional arrays
and keys have separate presence flags; empty separator arrays remain present,
and an explicit null key remains distinct from an absent key.

The same 8,732 cases produced 8,198,809 exact bytes from Go, native with
ASan/UBSan/LeakSanitizer, raw Node source, emitted JavaScript and yaml 2.9.0.
The combined lexer/CST/gap/mutant suite passed in 41.094s; the CST comparison
itself took 13.60s. Three additional mutants compile and exit zero with empty
stderr on native and Node, and the byte comparison catches each:

| Mutation | First differing output byte |
| --- | ---: |
| Key presence lost, including explicit null keys | 8086 |
| Source token offset advanced by one | 15 |
| Recorded line start shifted back by one | 180490 |

[emptyAlternative.ts](gaps/emptyAlternative.ts) proves the additional
`lower.NotYet` observation `an array of never`: an untyped empty array in a
conditional branch did not lower, though Node prints `1`; it lowers on compiler/area-stack
(views slice 1, Oct 8) and `TestClosedLexerGaps` holds it. The workaround is a
separately initialized, explicitly typed array followed by conditional assignment.
The earlier GraphQL-era forward class method limitation is closed on this main:
a scratch forward-method probe compiled successfully. No new refusal is claimed
for that behavior.

Cohere lint/type passed on all four new files, 276 rules, 100% Adamic-ready.
In-place list and token updates have explicit `@mutates` ownership contracts;
the output builder owns its parts. No cohere, compiler or runtime source changed.
Logs: [combined suite](audit/cst-suite.log), [CST lint](audit/cst-lint.log).
The existing throughput section still measures only lexical drivers.
Document composition and printing have not been implemented or measured yet.

## Green lexer step

`lexer.ts` ports yaml 2.9.0's lexer using the same UTF-16 units as Go cohere.
Eager token arrays replace generators. Scalar calls remain eager: a block header
at end of input must emit an empty scalar even after the last source unit has
been consumed. A first rewrite incorrectly made those calls deferred states;
the generated comparison caught the missing token and the final port restores it.

`lex_main.ts` reads an escaped batch and prints token units in hexadecimal.
This is a lexer comparison driver, not the requested file-formatting driver.
Each generated input is also parsed in chunks of one, two and seven UTF-16 units;
these exercise splits through quoted escapes, markers and astral characters.
The deterministic seed is 20261006. The corpus walk excludes only `.git`,
includes submodules, and fails on invalid UTF-8 rather than replacing it.

The final suite passed in 19.029s: 8,732 lexical cases, 4,181,796 output bytes.
Native ran with ASan, UBSan and LeakSanitizer enabled; stderr was empty.
The exact comparisons also passed on raw source Node, emitted JavaScript, and
scratch-installed yaml 2.9.0. The external oracle imports only the pinned
library. Go's lexer driver is built through an overlay; cohere is unmodified.

### Mutants

Each mutant compiles, exits zero, and has empty stderr on native and Node.
Only the wrong token bytes catch it:

| Mutation | First differing output byte | Control reached |
| --- | ---: | --- |
| Keep chomping becomes clip | 859077 | `a: |+` with trailing blank lines |
| Tab is removed from whitespace classification | 864210 | Tab after a mapping colon |
| BOM is no longer split as its own token | 858065 | BOM-only input |

### Stage 0 gap programs

Both active programs print the recorded behavior on Node and refuse with
`lower.NotYet` before clang. `TestLexerGaps` holds these observations so a
closed gap requires updating the workaround.

| Program | Node stdout | Observed refusal | Port workaround |
| --- | --- | --- | --- |
| [prefixIncrement.ts](gaps/prefixIncrement.ts) | `1` | Closed by compiler/area-stack (views slice 1) | Increment in a separate statement |
| [assignmentValue.ts](gaps/assignmentValue.ts) | `1` | Closed by compiler/area-next | Assign before reading the result |
| [stringPresence.ts](gaps/stringPresence.ts) | `false` | Closed by `cdddfee346e70a5cba3e982f808b9ebe426a6dfb` | Lexer sentinel tests use native string truthiness and negation |

## Lexer gaps compiler/area-next closed

Area-next lowers five of `TestLexerGaps`' programs: assignmentValue, dynamicCase,
negativeCase, stringFallback and valueConjunction. `TestClosedLexerGaps` keeps each
proving program and requires Node's output from native under ASan, UBSan and
LeakSanitizer and from emitted JavaScript, as the presence gaps do. valueConjunction
lowers but its C emission panics, `native: no conversion from 4 to 9`; the test names
that as its failure rather than letting the panic end the package, and it is the
compiler's to fix or refuse. The port's workarounds still stand; retiring each is
its own change against the corpus. compiler/area-stack (views slice 1) closes two more,
prefixIncrement and emptyAlternative, and `TestClosedLexerGaps` holds them the same way.
`TestLexerGaps` keeps the one that still stops, multiplePush.

Labels and logical assignment operators are explicit 0.1 refusals, not newly
observed language gaps. The port uses ordinary loop control and assignments.
The npm library's ISC license accompanies its adapted lexer in `LICENSE-yaml`.

### Lexer throughput

Five interleaved fresh-process rounds, after tests and builds had finished.
Each reads the same 1,189,637-byte input and writes 4,181,796 token bytes.
Each measured answer is compared to Go; exit zero and empty stderr are required.
Startup, file input, chunk handling and output are included; compilation is excluded.
Chunked cases count as texts. This does not measure the unfinished formatter.

| Driver | Median texts/s |
| --- | ---: |
| Native release | 8184.87 |
| Port source on Node | 14155.37 |
| Go cohere | 40714.98 |
| Original yaml 2.9.0 on Node | 11382.30 |

The native driver is 4.97 times slower than Go and 1.73 times slower than Node
on these inputs. This includes different output implementations: buffered Go
`fmt.Fprintf` versus token-by-token `console.log` on native and Node.

### Commands and evidence

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-artifacts go test -v -count=1 -timeout 30m ./stage1/cohere/yaml > /tmp/stage1-yaml-lexer-final.log 2>&1
go run ./cmd/adamic build stage1/cohere/yaml/lex_main.ts -o /tmp/stage1-yaml-artifacts/native-lexer > /tmp/stage1-yaml-lexer-build.log 2>&1
python3 stage1/cohere/yaml/benchmark_lexer.py /tmp/stage1-yaml-artifacts /tmp/stage1-yaml-library > /tmp/stage1-yaml-lexer-timing.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$' > /tmp/stage1-yaml-filtered-oracle.log 2>&1
/tmp/stage1-yaml-cohere --no-fix --format-only --format-all stage1/cohere/yaml/lexer.ts stage1/cohere/yaml/lex_main.ts > /tmp/stage1-yaml-format-check.log 2>&1
/tmp/stage1-yaml-cohere --no-fix --no-format stage1/cohere/yaml/lexer.ts stage1/cohere/yaml/lex_main.ts > /tmp/stage1-yaml-lint.log 2>&1
```

The filtered oracle passed in 4.097s. Cohere reported 276 rules, two files
checked, 100% Adamic-ready. No compiler or runtime implementation changed.
Logs: [suite](audit/lexer-suite.log), [throughput](audit/lexer-timing.log),
[filtered oracle](audit/lexer-oracle.log), [lint](audit/lexer-lint.log).
The complete repository gate was not run. Formatter parity and formatter
throughput remain pending; this checkpoint must not be treated as a finished YAML slice.

# Baseline audit before the lexer port

No Adamic YAML parser, printer, or stdout driver was built.
This directory records baseline evidence, not a completed stage 1 slice.
Go matches its formatter oracle on all 36 YAML files in this checkout.
Pinned JavaScript parser comparisons passed fixtures and generated cases.
Native parity, performance measurements, and three port mutants remain unrun.

## Revisions and setup

Branch `codex/stage1-yaml` starts at main
`1a4e29984b0b1727b063e76fe01c2aa11dfa56c7`.
Cohere is pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`.
The checkout's 36 YAML files contain 200,885 source bytes: 35 under
TypeScript and one under cohere. The inventory excludes `.git` only.

`bash cloud/setup.sh` exited zero. Its reported timings were Go 0s,
clang 1s, Node 1s, submodules 1s, build cache 78s, total 78s.
`nproc` reported 5; CPU quota was four CPUs (`cpu.max: 400000 100000`).
Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0.
The environment file is `/workspace/adamic-tools/env.sh`.
The complete setup output is [audit/setup.log](audit/setup.log).

## External baseline

Scratch npm dependencies were installed with exact versions:
Prettier 3.9.6, yaml 2.9.0, yaml-unist-parser 3.2.0.
The formatter tests use cohere's vendored fork bundles, not the npm
Prettier package, as their printer oracle. Lexer, composer, and unist
comparisons use the pinned npm packages under Node.
No direct npm Prettier formatting comparison is claimed.

The first Go test run passed but skipped external comparisons because
no Prettier fork path existed. After npm installation, a second run
failed: setting `COHERE_PRETTIER_FORK` also overrides the formatter's
bundle path, and npm does not provide `dist/prettier/standalone.js`.
A scratch symlink to cohere's vendored bundles supplied that path.
The corrected run passed all four packages:

| Package | Seconds | Comparisons |
| --- | ---: | --- |
| yaml | 11.202 | Formatting fixtures; 36/36 YAML files, full stack and tree loader |
| compose | 5.278 | 735 fixtures, 30,000 generated texts, 36 corpus files |
| cst | 2.191 | 586 fixtures, 2,930 chunked parses, 20,000 generated texts, 36 corpus files |
| unist | 5.110 | 444 fixtures, 20,000 generated texts, 42 corpus inputs |

The extra six corpus inputs are Markdown front matter selected by
cohere's existing tests. Their Go comparisons passed; this does not
constitute an Adamic front matter implementation.

`TestFormatOracleCanFail`, the three parser-layer `TestOracleCanFail`
tests, and `TestComparisonSeesEveryField` passed. These are existing
Go oracle checks. They are not the requested three Adamic port mutants.

YAML test-suite comparisons and the broader prose-wrap oracle test
skipped because the upstream test-suite fixture directory was absent.
All test output was redirected to log files. The successful external
run is [audit/go-upstream.log](audit/go-upstream.log).

## Reproduction

Run from the repository root, substituting its absolute path for
`/workspace/adamic` if necessary:

```sh
bash cloud/setup.sh > /tmp/stage1-yaml-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/stage1-yaml-library --save-exact prettier@3.9.6 yaml@2.9.0 yaml-unist-parser@3.2.0 > /tmp/stage1-yaml-npm.log 2>&1
mkdir -p /tmp/stage1-yaml-library/dist
ln -s /workspace/adamic/cohere/internal/format/prettier/bundles /tmp/stage1-yaml-library/dist/prettier
cd cohere
COHERE_PRETTIER_FORK=/tmp/stage1-yaml-library COHERE_YAML_CORPUS=/workspace/adamic go test -v -count=1 -timeout 30m ./internal/format/yaml/... > /tmp/stage1-yaml-external-fixed.log 2>&1
```

## Unfinished work

The formatter depends on the CST lexer/parser, document composer,
unist transformation and comment attachment, then the YAML printer and
document layout engine. The survey covered representative source from
these layers; it did not finish reading every implementation file.
The GraphQL reference was read. The JSON reference reading remains
incomplete, including portions of its generated tables and auxiliary
performance tooling. No implementation was edited before those reads.

No language blocker is established by this audit. There are no new
Adamic gap claims or proving programs. No file-formatting driver,
Adamic corpus comparison, native sanitizers, Go/Node/native texts-per-second
measurements, three port mutants, filtered Adamic oracle or full repository
gate was run. This audit must not be treated as a merge-ready YAML port.

## String presence: closed on the compiler area

`TestClosedStringPresenceGap` keeps the original proving program and requires
`false` from source Node, native ASan/UBSan/LeakSanitizer and emitted JavaScript.
`git log -S censusBooleanCondition -- internal/lower` identifies
`cdddfee346e70a5cba3e982f808b9ebe426a6dfb` as the ToBoolean/negation change.
The lexer now tests its empty-string sentinels directly or with `!`.
Numeric code-unit scanning is retained as an optimization, and `member` keeps
its explicit nonempty boolean operand because mixed string/boolean `&&` is
outside this gap (see `valueConjunction.ts`).

String closure validation: `ADAMIC_YAML_LIBRARY=/tmp/area-gaps-yaml-library
go test -v ./stage1/cohere/yaml -run
'^(TestClosedStringPresenceGap|TestLexerMatchesGo|TestLexerMutants)$' -count=1
-timeout 30m` passed (35.223s). The 8,732 complete/chunked cases from 36
repository files agree byte-for-byte with Go, source Node, both backends and
yaml 2.9.0, including sanitizer/leak checks. Five existing lexer mutants
(keep chomping, tab whitespace, BOM splitting, cached NUL, numeric whitespace)
execute successfully with wrong bytes and are caught on native and Node.
Changing the closed fixture from a nonempty to an empty string made its
normal Node `true` output fail the expected `false` check; it was restored.

## Value presence: closed on compiler area

`cdddfee346e70a5cba3e982f808b9ebe426a6dfb` (Lower unary findings 85 to 0
and overload signatures 191 to 0) lowers unary value truthiness. The original
`gaps/valuePresence.ts` is retained as `TestClosedValuePresenceGap`, matching
Node's `false` in both backends with ASan/UBSan and leak detection.
Scalar block-line iteration now tests indexed optional objects directly with
`!line` and `line`, retiring its explicit undefined-comparison workaround.
Each BlockLine object is truthy; an absent array element is undefined.

Value closure validation: `ADAMIC_YAML_LIBRARY=/tmp/area-gaps-yaml-library
go test -v ./stage1/cohere/yaml -run
'^(TestClosedValuePresenceGap|TestScalarsMatchGo|TestScalarMutants)$' -count=1
-timeout 30m` passed (33.009s). Across 36 repository files and 9,272 cases,
8,419,050 scalar answer bytes agree on Go, native, source Node, emitted
JavaScript and yaml 2.9.0; sanitizer/leak checks pass. The existing escaped-LF,
strip-chomping and flow-folding mutants execute normally with wrong bytes and
are caught on native and Node. Removing the item from the gap fixture's array
made its normal Node `true` output fail the expected `false` check; it was
restored. No full YAML package or formatter corpus was run.
