Fixed the four parser sources underlying all eight reported lint failures.
Go, Node and sanitized native agree on their recovered trees and diagnostics.
Three recovery mutants prove diagnostic, child-tree and timeout comparisons.
All 77 compiler files yield 22,497 planned cutoff/removal/duplication inputs.
Bare export and incomplete export clauses agree; wider recovery remains active.

## Built

Branch `codex/parser-recovery`, based on `origin/codex/typescript-scanner`
at `0090256e607c3f2de7d5b67cef680ec010f95c1d`, as the unit explicitly
requested. The checkout's narrow fetch omitted that branch; fetching its
explicit ref succeeded. No compiler, runtime, scanner, lint or submodule
source was changed. No pull request was opened.

The four source shapes are recorded in lint's `VOLUME.md` and `GAPS.md`,
rather than its `REPORT.md` on this branch:

```ts
interface I
interface I { m(a: string): void;
interface I { m<(a: string): void; }
interface I { m<T(a: T): T; }
```

Expected tokens now report diagnostic 1005 without consuming an unexpected
token. Type-member lists terminate at EOF and a missing opening brace creates
an empty member list. Generic parameter lists recognize Go's recovery
delimiters, including an opening parenthesis. Parameter lists also terminate
at EOF and retain an empty missing list when their opener is absent.
Parser speculation rolls diagnostics back with the scanner and node table.

`main.ts --whole --recovery` prints diagnostics and the recovered canonical
tree. The ordinary driver rejects diagnosed input instead of silently printing
an answer without its diagnostics. Diagnostic fields are code, byte start,
byte length, category and escaped English message. UTF-16 positions are
converted through the same byte-offset table used for trees. Tree fields
remain exactly the canonical protocol documented in `WHOLE_REPORT.md`.
This is not serialization of every Go AST field.

The independent Go oracle calls unmodified typescript-go. Recovery mode
allows its parse diagnostics and prints their actual localized messages.
It does not normalize their codes or positions. Twelve parser cases cover
the four originals, Unicode/CRLF prefixes and trailing CRLF/whitespace.
A separate integration test exercises all eight original method/property
lint combinations against real cohere findings and proposed repairs.
Cohere's edit engine refuses malformed source, so this is findings/proposed
repairs parity, not fixed-source parity.

## Wider comparison and remaining failure

`TestIncompleteCompilerAgrees` checks the TypeScript v6.0.3 pin
`050880ce59e30b356b686bd3144efe24f875ebc8` and enumerates every compiler file
before comparing. The real Go parser identifies literal locations for
regex/template rescans in the independent scanner, giving token boundaries
without slash or brace heuristics. Each file is cut after every Nth token,
with N selected for up to 256 cuts. Short files use N=1. An empty cutoff is
also included. Up to 32 regularly spaced tokens per file are removed and
then duplicated with a separating space. No diagnostic input is filtered.

Observed: all 77 files were enumerated, giving **22,497 planned inputs**.
Only **two inputs were compared**, because the test stops at its first
failure after running both ports. The empty cutoff passes. A namespace
wrapper cut after `export` fails on both Node and native with exit 70.
Go completes with diagnostic 1128, `Declaration or statement expected.`,
and a SourceFile containing only its EOF token. The exact input and all
three outputs are retained in `validation/recovery-export.*`.

Each incomplete-input process has an independent two-second deadline.
Input, stdout and stderr are written before reporting a failure and survive
test cleanup under `/tmp/adamic-parser-incomplete`, or the explicit
`ADAMIC_RECOVERY_ARTIFACTS` directory. Token-boundary enumeration is bounded
as well. The EOF-loop mutant proves a nonterminating case fails by this
deadline with its input saved.

General parity was **not achieved**. The next missing mechanism is Go's
outer-list element recognition and resynchronization: a bare export is not
a statement element, so Go reports and skips it; the port consumes it as a
modifier and attempts to parse an EOF expression. That mechanism is not
implemented by expected-token insertion or the signature list terminators.
Other malformed grammar remains unverified. Scanner-error integration,
missing identifier/expression/type nodes, related diagnostic information,
and recovery context flags are also outside this implementation.
The remaining 22,495 planned inputs, including the token edits, have not
been compared. The broader test deliberately remains failing when the
pinned corpus is supplied; its failure is not converted into a skip.

Lint's old `checkRecoveryRefusal` assertions still describe the earlier
parser behavior. They need migration to positive parity checks by the lint
unit. This unit leaves that worker's files untouched and provides the
positive eight-case integration test under the parser package instead.
The complete repository gate is not claimed green.

## Mutants

| Mutant | Independent check that catches it on both backends |
| --- | --- |
| Expected-token diagnostic 1005 changed to 1006 | Diagnostic bytes differ while the tree remains intact |
| Generic TypeParameter child list emptied | Recovered tree loses the Identifier child for T |
| Type-member EOF terminator removed | Each run reaches the two-second deadline |

The first two compile and finish normally on Node and ASan/UBSan native.
The third compiles and is caught specifically by the external deadline,
not by a compiler refusal or sanitizer crash. Successful executions retain
leak checking through the normal sanitized native build.
The first version of the child mutant used an untyped empty array and met
stage 0's `array of never` refusal. That attempt is **not credited**.
The final mutant uses `children.slice(0, 0)` and completes normally.

## Commands and observations

Every test command wrote its output to a log, never a pipe. Commands ran
from the repository after sourcing `/workspace/adamic-tools/env.sh`.
`nproc` printed **5**. `bash cloud/setup.sh` succeeded and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (21s)
setup: done in 21s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, Node 24.19.0, clang 20.1.8. The actual setup log is retained.

```sh
bash cloud/setup.sh > /tmp/parser-recovery-setup.log 2>&1
source /workspace/adamic-tools/env.sh

go test ./stage1/typescript/parser -run '^TestMethodRecoveryAgrees$' \
  -count=1 -v > /tmp/parser-recovery-baseline.log 2>&1
# Before the fix: FAIL, 18.563s. Three source shapes panic on both backends;
# the missing closing brace reaches both two-second deadlines.

go test ./stage1/typescript/parser \
  -run '^(TestMethodRecoveryAgrees|TestRecoveryMutants|TestRecoveredLintCasesAgree)$' \
  -count=1 -v -timeout 10m > /tmp/parser-recovery-focused-final.log 2>&1
# PASS, 70.908s. Twelve parser inputs, all eight lint combinations,
# and all three mutants on both backends.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -skip '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 30m > /tmp/parser-recovery-regression.log 2>&1
# PASS, 361.047s. All existing parser tests and recovery tests passed.
# 77 whole compiler files: 44,766,682 identical whole-tree bytes.
# The broader incomplete-input test is explicitly excluded here.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 30m > /tmp/parser-recovery-incomplete-final.log 2>&1
# FAIL, 15.026s. Planned 77 files / 22,497 inputs; stopped after two inputs.

(cd cohere && go build -o /workspace/scratch/cohere ./command/cohere) \
  > /tmp/parser-recovery-cohere-build.log 2>&1
/workspace/scratch/cohere --format-only stage1/typescript/parser/*.ts \
  > /tmp/parser-recovery-format.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts \
  > /tmp/parser-recovery-cohere.log 2>&1
# Exit 0, 276 rules, no findings, 100% Adamic-ready.

go vet ./... > /tmp/parser-recovery-vet.log 2>&1
# Exit 0, no output.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' \
  -count=1 -timeout 10m > /tmp/parser-recovery-filtered-oracle.log 2>&1
# PASS, 3.225s.
gofmt -l stage1/typescript/parser
# No output.
git diff --check
# No output.
```

The first lint integration build succeeded but emitted Go dependency-download
chatter on stderr. The existing execute helper treats any stderr as failure;
rerunning after the successful downloads passed all eight combinations in
20.103s. It was a harness/build-output issue, not a lint comparison failure.
The final focused run repeats the checks on formatted source, including the
added Unicode and trailing-trivia variants. Full repository tests and lint's
full package were not run. Parser regressions, the focused integration, vet,
cohere source lint and the named filtered compiler oracle were run.

Implementation commit: `54024b1` (full SHA recorded by `git rev-parse` in
`validation/recovery-commits.log`). The subsequent documentation commit
retains this report, raw logs and the first broader failure. The requested
branch is pushed without rewriting history.

## Continued recovery: stranded export

Declaration lookahead now follows Go's modifier scan. A stranded export in
source elements reports 1128 and advances without constructing a statement.
Six additional inputs cover EOF, CRLF, duplicate exports, an exported
variable, a following expression, and a Unicode prefix. The focused parser
comparison passes in 14.686s. The new stranded-export mutant changes 1128 to
1129, compiles and finishes on both backends, and is caught by diagnostic
bytes (11.656s). Cohere formatting and all parser source rules pass.

The broader comparison advanced to input three, `export *`: Go's missing
module expression is an empty Identifier, while the port retains the
scanner's stale value `export` at EOF. This is the next recovery mechanism
to implement. The full corpus remains red; no failure is filtered.

Logs: `validation/recovery-export-step.log`,
`validation/recovery-export-mutant.log`, and
`validation/recovery-incomplete-next.log`.

## Continued recovery: missing module expressions

Missing identifiers now have empty text and zero width without consuming the
unexpected token. Missing expressions report Go's 1109, at the full token
start for EOF. Import/export module specifiers use expression descent.
Focused tree and diagnostic comparison passes in 13.884s. All five token
cuts of the first namespace wrapper now pass. The wider run reaches input
seven, removing export, and exposes the port's semicolon refusal.

## Continued recovery: missing semicolons

Semicolon recovery now reports 1005 and leaves the next token for the next
statement instead of refusing. The saved removed-export namespace input is
included in the focused comparison. That comparison passes in 14.322s.

## Continued recovery: modified export clauses

A recovered export clause now retains preceding modifiers, including a
duplicated export token. The focused comparison passes in 14.255s. The
broader comparison advances to input nine, removing the asterisk; Go uses
its expression-statement-specific diagnostic 1434 for the following from
identifier, which is the next difference under repair.

## Continued recovery: expression-statement diagnostics

A missing semicolon after a nonempty Identifier expression now reports 1434
at that identifier's range, while other expressions retain the expected
semicolon diagnostic. Explicit-range diagnostics preserve Go's suppression
of consecutive diagnostics at the same start. The focused comparison passes
in 14.756s and cohere reports no findings. The wider run has passed both
five-token namespace wrappers and reached the main namespace wrapper.

## Continued recovery: source element recognition

Source-list recognition now handles stranded import tokens and invalid
punctuation with Go's 1128 diagnostic and token advancement. Declaration
lookahead distinguishes import declarations from import expressions and
handles contextual modifiers. Focused comparison passes in 14.596s. The
wider comparison reaches input 220: an import specifier list cut at EOF
needs an empty remaining list and a missing closing-brace diagnostic.

## Continued recovery: EOF in open lists

All bracketed lists now stop at EOF instead of inventing a final element or
looping. Missing types produce diagnostic 1110 and an empty TypeReference
name. Focused comparison passes in 16.280s across import/export lists,
arguments, arrays, objects, bindings, classes and function bodies. The wider
run passes 286 inputs and reaches a duplicated import token at input 287;
Go leaves that second reserved word for the next import declaration.

## Continued recovery: duplicated imports

Reserved import tokens are no longer consumed as default-import names.
Import in a primary-expression position is accepted only for calls and
meta-properties; otherwise it creates a missing expression without advance.
Focused comparison passes in 16.526s; cohere passes. The wider continuation
checks inputs 287 through 297 and first fails in binder.ts on a conditional
expression cut after its question mark.

The harness can resume at an explicitly logged ADAMIC_RECOVERY_START while
investigating, but reports such runs as partial. The default gate still
checks every planned input. Successful input/output artifacts are removed;
failed inputs and both output streams remain saved.

## Continued recovery: incomplete expression tails

Conditionals insert a zero-width ColonToken and missing false Identifier
when the colon is absent. Optional chains insert a missing member name.
Template expressions and template types insert an empty TemplateTail when
the interpolation's closing brace is absent. The focused comparison passes
in 16.820s. The wider continuation advances to input 307 in binder.ts,
where an incomplete function return type must be recognized before its
arrow token exists. Mutant anchors remain restricted to one exact site.

## Continued recovery: function type recognition

Function types are recognized from Go's unambiguous parameter starts, not
from the eventual presence of an arrow. Speculation rolls back nodes and
diagnostics. Incomplete named, rest and generic type signatures now recover
normally. Focused comparison passes in 16.720s. The wider continuation
advances through 31 more inputs and reaches input 338: a variable list cut
immediately after const must stay empty at EOF.

## Continued recovery: variable and arrow cutoffs

A variable list at EOF remains empty. Typed/rest arrow signatures are
recognized before the arrow exists and insert a missing EqualsGreaterThanToken
instead of an EOF token. Focused comparison passes in 17.941s. All remaining
binder.ts cutoffs pass in the continuation, which checks 208 inputs in
77.350s before reaching a removed interface property name at input 545.
That failure needs type-member recognition and outer-list resynchronization.

## Continued recovery: type member resynchronization

Type-member lookahead now follows Go's modifier and property scan. Invalid
members report 1131 and resume the outer statement list when its element
starts; member separators report missing semicolons. Focused comparison
passes in 19.997s and cohere passes. The continuation checks four inputs and
reaches input 548: a duplicated call argument needs missing-comma recovery.

All four current mutants pass in 54.694s on Node and sanitized native.
The termination mutant now makes EOF appear as an Identifier; removing only
the old type-member loop guard no longer causes nontermination because
resynchronization also recognizes EOF. Both mutant runs hit the two-second
deadline specifically. Diagnostic and child mutants still finish normally.
Current raw logs are retained under validation/recovery-*.log.

## Continued recovery: delimited call lists and block contexts

Call arguments now use Go's list algorithm: recognize elements, insert a
missing comma, resynchronize against active outer lists, and advance only
when no active context accepts the token. Blocks use statement recognition
and stop recovery at their own closer. Source and type-member lists share
those contexts. The diagnostic and terminator tables live in recovery.ts,
a separate grammar concern required by cohere's 2000-line file limit.

Focused comparison passes in 21.386s and cohere passes after its style fixes.
The continuation advances to input 552, where a duplicated switch-case
colon still stalls the old switch-clause statement loop. This is the next
list to migrate to the same recovery mechanism.

## Continued recovery: switch and module statement lists

Switch clauses and module blocks now use active statement-list recovery.
Duplicated colons report the same statement diagnostic and advance without
inventing an extra statement or stalling. Focused comparison passes in
22.667s. The continuation advances to input 554, duplicating const, where
Go reports 1389 and leaves the keyword to start another variable statement.

## Continued recovery: variable declaration lists

Variable declarations now use the same Go list recognition, comma insertion
and outer-context recovery. A duplicated const reports 1389 and remains
available to start the next statement. Focused comparison passes in 21.109s;
cohere passes. The continuation reaches input 555, removing a parameter comma.

An attempted higher-order callback through StatementContextInterface stalled
native compilation in internal/fresh.ProveWrites before any input ran. It
was stopped with SIGQUIT after 155.786s; the stack is retained in
validation/recovery-variable-focused.log. Expressing the identical list
algorithm through existing first-order callbacks builds and passes normally.
No compiler source was edited, and that stalled build is not a mutant kill.

## Continued recovery: parameters and invalid types

Parameter lists now recognize elements, insert commas and resynchronize
against active contexts. Binding names reject reserved words with diagnostic
1359; this parameters remain supported. Invalid type tokens create an empty
TypeReference name with diagnostic 1110. Missing function bodies retain an
empty Block with diagnostic 1144 where required. Focused comparison passes
in 22.130s and cohere passes. The continuation advances to input 560,
duplicating break, which requires a missing label Identifier.

## Continued recovery: jump labels and throw expressions

Break and continue retain a missing Identifier label and reserved-word
diagnostic instead of dropping the label. Throw always retains an expression,
including a missing Identifier at EOF or after a line break. Its semicolon
recovery shares Go's identifier-expression diagnostic. Focused comparison
passes in 23.829s and cohere passes. The wider continuation advances to
input 598: Go suggests 'set Parent' for a duplicated setParent identifier,
requiring its keyword spelling/spacing suggestion algorithm.

## Continued recovery: keyword spelling and spacing

Identifier-expression diagnostics use Go's weighted edit distance, candidate
length limits and lexical tie breaker, then its keyword-prefix spacing
suggestion. Unicode case probes use scalar characters. Focused comparison
passes in 23.029s. The initial commit still had 28 cohere findings (ambiguous
loop identifiers and a string template preference); the following correction
renames the indices and uses a template, and passes cohere. All remaining binder.ts edits pass in
the continuation, which checks 102 inputs before reaching input 699 in
builder.ts: an indexed type cut after its bracket is recovered as an ArrayType.

## Continued recovery: incomplete array types

An opening square bracket without a following type is an ArrayType in Go,
with a missing closing bracket diagnostic. The port now checks Go's type-start
condition before choosing IndexedAccessType. Three focused cases cover EOF,
a parameter type, and a semicolon. Focused comparison passes in 24.776s;
source cohere passes. Wider continuation starts at input 699.

## Continued recovery: object literal lists and bounded workers

The removed comma at builder.ts token 4981 makes a type declaration recover
into an object expression. Object members now use Go's delimited-list recovery,
including consuming a same-line semicolon and retaining its trailing separator
flag. Three focused additions cover semicolons, missing commas, and this type
recovery path. Focused Go/Node/native comparison passes in 23.585s and cohere
passes. The same deterministic 22,497-input corpus now runs with four workers,
unique position-prefixed artifacts, and unchanged independent two-second
process deadlines. Failures stop new work and preserve all failing inputs.

## Continued recovery: standalone arrows and malformed property names

The builder.ts continuation retains failures at inputs 902 and 904. The former
exposes Go's property-assignment recovery when the property name is missing,
and its 1443 diagnostic for a tagged template without a statement separator.
The latter exposes standalone => recovery into an arrow with a missing parameter
list. The port now preserves these exact shapes and diagnostic ranges. Five
focused additions pass with all preceding probes in 25.038s; cohere passes.

## Continued recovery: scanner diagnostic delivery

The same malformed-template input exposes the missing delivery of scanner
errors to parser diagnostics. Normal scans and regex/template rescans now
deliver new errors immediately through the existing position deduplication,
and parser lookahead rolls back diagnostics together with scanner state.
English messages are ported from typescript-go; diagnostics requiring message
arguments still fail explicitly until implemented rather than fabricating text.
Four focused additions cover unterminated templates, strings, comments and
regex literals. All focused comparisons pass in 26.186s; cohere passes.

## Continued recovery: permissive parameter starts

At input 913 a duplicated colon before a union type made the port leave the
parameter list early. Go also recognizes type starts as parameter recovery
elements, except function, minus and opening-parenthesis tokens. The port now
uses that predicate, preserving missing parameter nodes and consuming invalid
tokens through Go's zero-progress recovery. Three focused probes pass with
all prior cases in 24.708s; cohere passes.

## Continued recovery: import and export specifier separators

The continuation advances through builder.ts and builderPublic.ts cuts, then
retains missing/duplicated import commas at inputs 1153/1154. Specifier lists
now use Go's list contexts, separator expectations and invalid-token recovery,
including an empty list when the opening brace is missing. Statement lookahead
scans without delivering errors, then restores scanner state, so lexical
diagnostics remain speculative. Four focused additions pass with preceding
probes in 27.286s; cohere passes.
