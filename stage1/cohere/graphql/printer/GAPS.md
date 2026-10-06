# GraphQL printer gaps

## Scope and oracle results

The printer ports cohere's `internal/format/graphql/print.go`,
`print_helpers.go`, `visitor_keys.go` and formatting entry point, with the shared
comment attachment and document rendering operations reachable from GraphQL.
It composes with the existing Adamic lexer and parser without changing them.
All 45 visitor kinds are asserted covered by the independent Go AST walk.

On 3,564 generated texts, four option sets and three implementations (native,
Node source, JavaScript backend), every output is byte-identical to Go cohere.
The corpus generator uses seed 20261006. It includes descriptions, escaped and
block strings, comments in each attachment position, prettier-ignore, every
extension, fragment arguments, blank lines, BOMs, CRLFs, tabs, Unicode display
width and long argument lists. Deep list/object/selection/type cases reach depths
500/300/300/400. Recursive discovery found zero .graphql files in this checkout
and its initialized submodules. This is generated coverage, not a claim about a
nonexistent file corpus.

## 1. Nested constructor initialization boundary

**Observation:** `gaps/nestedConstructor.ts` imports `gaps/nestedInner.ts` and
constructs Inner while initializing Holder. Node prints `ready`. Stage 0 refuses
Inner.show's call to Inner.read, at nestedInner.ts:5:29, as this escaping a
constructor before all fields are set. Inner.label was already assigned. The
fixture's padding comment preserves the source offsets that expose the refusal.

**Inference from lowering:** the outer constructor's numeric `unsetUntil`
boundary survives nested class instantiation and is consulted while lowering an
inner method, without belonging to that method's source/function. See
`internal/lower/class_inheritance.go`, `class.go` and initialization checks. This
is a conservative false refusal, not evidence of an escaping partially
initialized object.

**Workaround:** construct Documents at the call site and pass it into Printer's
constructor. The parser remains unchanged and no compiler or runtime code is
modified. `TestPrinterConstructorGap` requires Node's success and the exact
refusal category; removing the gap must update this report and the workaround.

```sh
go test -v -count=1 ./stage1/cohere/graphql/printer -run TestPrinterConstructorGap > /tmp/graphql-printer-constructor-gap.log 2>&1
```

## 2. Go cohere and Prettier disagree on whitespace-only files

**Observation:** npm Prettier 3.9.6 and cohere's embedded fork 3.9.6 return the
empty string for the five corpus entries below. Go cohere rejects each. The port
preserves Go's exact errors, so universal parity with both upstreams is currently
impossible. No valid formatted document differs from either Prettier build.

| Corpus index (zero-based) | Input (escaped) | Go error |
|---|---|---|
| 717 | empty string | Syntax Error: Unexpected <EOF>. (1:1) |
| 1482 | one ASCII space | Syntax Error: Unexpected <EOF>. (1:2) |
| 2381 | \n\r | Syntax Error: Unexpected <EOF>. (3:1) |
| 2934 | \r | Syntax Error: Unexpected <EOF>. (2:1) |
| 3526 | one ASCII space, duplicate | Syntax Error: Unexpected <EOF>. (1:2) |

The file wrapper converts CRLF and lone CR to LF before parsing. Consequently
\n\r is parsed as two LFs. It removes and restores a UTF-8 BOM on successful
formatting, matching cohere's native.Formatter rather than calling graphql.Format
on raw file bytes.

`gaps/whitespace-cases.json` independently contains these five inputs. The test
encodes them into the batch transport. `TestPrinterWhitespaceGap` holds native and Node to Go refusals and
requires empty successful output from both Prettier builds. The full preflight
also checks these exact input/error pairs and requires exactly five differences
in each option set. Every other acceptance or formatted-byte mismatch fails.

```sh
ADAMIC_GRAPHQL_PRETTIER=/tmp/graphql-printer-prettier go test -v -count=1 ./stage1/cohere/graphql/printer -run 'TestPrinter(WhitespaceGap|UpstreamPreflight)' > /tmp/graphql-printer-whitespace-gap.log 2>&1
```

## Mutants

`TestPrinterMutants` changes only a temporary copy of port source. All three
mutants lower and run successfully, then fail the byte comparison on native and
Node. A compile failure is not credited as a caught mutant.

| Mutation | First caught output line | Evidence |
|---|---|---|
| Ignore available line width in fits | 3 | A long emoji argument remains flat instead of wrapping. |
| Attach an end-of-line comment as a leading comment | 547 | Comments after braces, fields and arguments move to the next line/node. |
| Trim each block string line | 541 | Two semantic indentation spaces are lost from the string. |

See results/printer-tests.log for the differing bytes and successful sanitizer
runs. These checks prove width, comment placement and block-string preservation
can fail independently.

## Not covered

No general-purpose Prettier document library, cursor/range formatting, embedded
languages or arbitrary parser options are added. Only operations reachable from
GraphQL are ported. Parser limitations documented in ../GAPS.md still apply,
including non-UTF-8 byte inputs and non-document parser entry points. The cooked
parser string transport is decoded explicitly using String.fromCodePoint;
original parser behavior is unchanged. The fixture corpus cannot establish
correctness for all possible documents or platform-dependent Unicode versions.
