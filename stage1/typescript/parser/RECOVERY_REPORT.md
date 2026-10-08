Rebased codex/parser-recovery onto origin/main at 5d4c801 without conflicts.
Parser regression passes in 612.959s, including five recovery mutants.
77 compiler files match Go, Node and native expression and whole-tree bytes.
Vet, formatting, cohere and the filtered oracle pass; setup 86s, nproc 5.
The wider gate fails at input 2597 after 2,603 comparisons, with no timeouts.

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

## Initial comparison before continuation (historical)

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

## Original mutant runs (historical)

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

## Continued recovery: missing element-access arguments

Inputs 1162/1163 expose empty element access while Go resynchronizes malformed
interface members. Empty [] now produces Go's missing identifier argument and
zero-width diagnostic 1011, also for optional chains. The type-start token
table moves to the existing lookahead module to preserve the repository's
source-size check. Three focused additions and all preceding probes pass
after extraction in 27.142s; cohere passes. The continuation next retains
a duplicated union operator and a missing parameter-list closing parenthesis.

## Continued recovery: type operator precedence and binding lists

Inputs 1186/1187 expose duplicated union operators and a reserved binding
property after a missing parameter-list closing parenthesis. Leading type
operators now respect the parsing precedence, so doubled operators retain
a missing type and diagnostic 1110. Object bindings now require a colon for
non-identifier property names. Object and array bindings use Go's delimited
list recovery and allow-in context, preserving missing commas and invalid
object separators. Eight focused additions and preceding probes pass in
28.578s; cohere passes.

## Continued recovery: function declaration binding names

Duplicated function keywords at inputs 1190/1192 now retain a missing binding
name and diagnostic 1359; non-default functions always parse a binding name,
while anonymous default exports keep their optional name. Four focused
additions and preceding probes pass in 29.184s; cohere passes. Its suggested
logical assignment shorthand is refused by Adamic 0.1, so the modifier flag
uses a plain if. The continuation reaches input 1236, a namespace missing
its opening brace.

## Continued recovery: missing namespace blocks

A namespace cut after its name (input 1236) now retains an empty ModuleBlock
and the missing-opening-brace diagnostic. Nested dotted namespaces use the
same recovery; ambient external modules retain their allowed semicolon.
Four focused additions and all preceding probes pass in 30.714s; cohere passes.

## Continued recovery: stranded let expressions

At input 1294 a cutoff after let is an identifier expression in Go. Statement
parsing now treats let as a declaration only when followed by a binding
identifier or destructuring opener, preserving exported declarations and
Go's distinct for-initializer rule. Six focused probes and preceding cases
pass in 30.458s; cohere passes.

## Continued recovery: enum lists

Inputs 1470/1471 expose missing and duplicated enum commas. Enum member
lists now recover in Go's active context, diagnose missing separators with
1357 and invalid members with 1132, and retain empty lists after a missing
opening brace. Enum expressions use Go's allow-in and cleared await/yield
contexts. Four focused probes and all preceding cases pass in 30.015s;
cohere passes.

## Continued recovery: contextual keyword labels

The duplicated undefined at input 1499 makes undefined: a recovered label.
Label recognition now follows expression parsing, as in Go, so contextual
keywords can label statements and await-context expressions cannot. Label
names are removed from the port's separate expression-root collection. Four
focused additions and preceding cases pass in 30.164s; cohere passes.

## Continued recovery: keyword-specific statement diagnostics

Removing the interface name at input 1597 now matches Go's diagnostic 1438
and current-token range. The related interface, namespace, type-alias,
variable-declaration and type-predicate diagnostics follow Go's missing
semicolon recovery. Stranded declare/abstract tokens use declaration lookahead
and unknown scanner tokens retain their scanner error without a fabricated
identifier error. Ten focused additions and preceding probes pass in
30.901s; cohere passes after its else-if style correction.

## Continued recovery: missing finally blocks

At input 1644, a checker.ts cutoff inside try, Go retains a missing finally
Block after reporting the missing closing brace. Try recovery now always
parses finally when no catch exists, reporting 1472 unless position
deduplication suppresses it, and retaining the empty block. Six focused
additions and preceding cases pass in 33.092s; cohere passes.

## Continued comparison: large-output deadlines

Input 1725 (checker.ts cut after token 117218, 1,046,858 source bytes) failed
the initial two-second native deadline while printing: native had emitted
5,175,461 of Go's 5,201,071 output bytes. Retesting with a fixed ten-second
parse-and-print deadline passes exact Go/Node/native bytes. The corpus uses
that same deadline for all three runtimes; the small probes and loop mutants
retain two seconds. The shared deadline helper still kills the seeded EOF
loop in both runtimes: filtered mutant test passes in 27.906s. Progress and
slowest per-runtime durations are recorded without changing the corpus.
The saved timeout input is retained separately at
/tmp/adamic-parser-timeout-evidence/checker-cut-117218.ts, SHA-256
ebca5fcd4795f2c463b828006461a8a101bb52d40223a05ad2cf3a69ec0b4281.
The wider continuation remains in progress; this is not a full corpus pass.

Raw continuation commands and their pass/fail output are retained under
validation/recovery-*.log. Failed development probes and resumed comparisons
are evidence of the work in progress, not additional green gates or mutants.

## Continued recovery: array and type-parameter lists

Input 1950 (a duplicated array member in checker.ts) now expects a comma and
continues the array list. The for-in probe caught an incorrect allow-in reset;
arrays retain Go's inherited restriction, and expression-start recovery
excludes in when that restriction applies. At input 2545, duplicated type
parameter defaults recover as additional parameters with the expected comma
diagnostic. Generic lists now use Go's contexts, modifier lookahead, invalid
constraint expression handling, and separator recovery. Ten added probes and
all preceding cases pass in 35.150s; the focused-plus-generic-child-mutant
verification passes in 52.623s, and cohere passes. The continuation advances
to input 2575, exposing parameter modifier recovery.

## Continued recovery: parameter modifiers

At input 2575, a duplicated function-type opener makes Go recover through
a parameter beginning with export. Parameters and generic parameters now
share Go's contextual modifier lookahead, including export/default/static
rules, const permission, duplicate-static stopping, and decorator ordering.
The recovered export modifier is retained before the reserved function-name
diagnostic. Four additional probes and preceding cases pass in 34.965s;
cohere passes. The next failure (input 2578) is a recovered list range flag.

## Continued recovery: trailing list spans

Input 2578 skipped tokens after the last recovered object member. Go computes
HasTrailingComma from the last node end versus the list end, rather than from
the last literal comma. Delimited lists now use the same comparison. Four
additional recovery probes and preceding cases pass in 34.917s; cohere passes.

## Continued recovery: missing method bodies

Inputs 2578 through 2581 exposed empty method-body nodes after malformed
signatures. Methods now distinguish automatic semicolons from required
bodies and retain the missing Block, with the object/class diagnostic choice.
Four added probes and preceding cases pass in 36.039s; cohere passes.

## Continued recovery: generic arrow speculation

Input 2585 duplicated the generic opener in a function declaration. Typed
parameters in a generic expression do not alone commit an arrow parse;
without its arrow Go falls back to a type assertion. Three added probes and
preceding cases pass in 36.045s; cohere passes. The user's main rebase now
interrupts the wider continuation, which has not passed its complete gate.


## Rebase onto current main

The user redirected this turn to rebasing, rerunning validation and pushing.
All 53 branch commits were replayed without conflicts onto origin/main
5d4c8012a0877094134e6c6bac367ff68f9313e8. Main's hot-file split and positive
class/interface dispatch test are retained. There are no changes under
internal/ relative to main. Older commit SHAs in this report refer to the
pre-rebase history; validation/rebase-commit-map.txt records the mapping.
The rebased recovery tip before the additional fixes is
056de2075c89e0a204e5b0edd9dd42b2b2246957.

The first regression rerun exposed existing continuation regressions:
nested parentheses could commit a typed arrow incorrectly, and module
specifier strings were registered as expression roots. The arrow head now
requires an identifier-like first token; string module specifiers use their
literal path while missing specifiers still receive expression recovery.
Parser speculation now rolls expression roots back along with nodes and
diagnostics. Its seeded mutant independently produces an extra expression
on both runtimes. The fixed expression-start token inventory moved into
grammar.ts to keep parser.ts under cohere's 2,000-line limit. Vet also caught
Fatal calls in comparison workers; those workers now report and return.
The interrupted initial regression and failed intermediate verification are
retained as development evidence, not green runs.

Every command below wrote output directly to its named log. These logs are
retained in validation/. Commands used /workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/parser-recovery-rebase-setup.log 2>&1
# PASS: Go, clang, Node and submodules ready in 0s each.
# setup: build cache warm (86s)
# setup: done in 86s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc > /tmp/parser-recovery-rebase-nproc.log
# 5

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -skip '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 30m > /tmp/parser-recovery-rebase-final-regression.log 2>&1
# PASS, 612.959s. The incomplete-input test is run separately.
# 1,676 generated expression inputs: 1,432,520 identical bytes.
# 58 expression probes: 17,463 identical bytes.
# 77 compiler files: 28,836,875 identical expression tree bytes.
# 77 compiler files: 44,766,682 identical whole-tree bytes.
# Focused recovery PASS, 73.29s; five recovery mutants PASS, 106.46s.
# All eight original lint combinations PASS, 35.91s.
# 42 type-node kinds covered, 250 identical doc-type bytes.
# 68 generated whole files: 93,323 identical bytes.
# Obsolete assert forms: diagnostic 2880 and 1,021 identical tree bytes.
# Main's positive class/interface dispatch probe passes under sanitizers.

/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts \
  > /tmp/parser-recovery-rebase-final-cohere.log 2>&1
# PASS, 276 rules, no findings; 13 of 13 Adamic-ready.
go vet ./... > /tmp/parser-recovery-rebase-final-vet.log 2>&1
# PASS, no output.
gofmt -l stage1/typescript/parser > /tmp/parser-recovery-rebase-final-gofmt.log
# PASS, no output.
git diff --check > /tmp/parser-recovery-rebase-final-diff-check.log 2>&1
# PASS, no output before adding this report's raw logs.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v \
  > /tmp/parser-recovery-rebase-filtered-oracle.log 2>&1
# PASS, 10.114s. Both native and Node cache misses.
```

The final regression reran every existing parser mutant. All compile and
finish normally on Node and sanitized native, except the intentional EOF
loop, which must reach its two-second deadline:

| Mutant | What catches it on both runtimes |
| --- | --- |
| Expected-token code 1005 becomes 1006 | Exact diagnostic bytes |
| Bare export source-list diagnostic incremented | Exact diagnostic bytes |
| TypeParameter Identifier child removed | Recovered tree bytes |
| EOF exposed as Identifier | Two-second deadline, input saved |
| Speculative roots retained | Extra expression versus Go |
| Multiplication precedence lowered | Expression tree bytes |
| Optional-chain flag cleared | Expression tree flag bytes |
| Parenthesized expression becomes ArrowFunction | Expression node kind |
| ForOfStatement becomes ForInStatement | Whole-tree node kind |
| Type-only import phase lost | Whole-tree semantic field |
| keyof operator becomes readonly | Whole-tree operator field |
| countTree starts each node at zero, expression mode | 0 versus Go's 18, unchanged trees |
| Same counter mutant, whole mode | 0 versus Go's 12, unchanged trees |

The filtered compiler oracle's one-byte control also passes. No full
repository test gate or performance benchmark is claimed. Canonical tree
fields and diagnostic fields are those documented above; related diagnostic
information and every internal Go AST field are not serialized. The wider
22,497-input comparison was restarted with ADAMIC_RECOVERY_START unset and
has not yet passed its complete gate.


## Final wider rerun after the rebase

```sh
unset ADAMIC_RECOVERY_START
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 2h > /tmp/parser-recovery-rebase-incomplete.log 2>&1
# FAIL, 712.404s. All 77 files / 22,497 selected inputs were planned.
# Started at input 1 with no continuation skip or diagnostic filtering.
# Stopped after 2,603 comparisons, including the failing input and in-flight jobs.
```

Observed: input 2597, core.ts duplicate token 15780, fails on both Node and
sanitized native. The changed signature starts:

```ts
function cartesianProductWorker<T>(arrays: readonly (readonly T[]) )[], result: (readonly T[])[], outer: readonly T[] | undefined, index: number) {
```

The first differing diagnostic is port code 1005 at byte 90216, length 1,
message `')' expected.`, versus Go code 1005 at byte 90217, length 1,
message `',' expected.`. The recovered trees also differ. Both ports finish
normally; this is a recovery mismatch, not a timeout. No process in this
rerun timed out. Slowest complete parse-and-print observations were Go
0.850014998s, Node 2.421945107s, and native 8.154023398s, all within the
shared ten-second corpus deadline. The EOF mutant still uses two seconds.

The exact input and all three stdout/stderr pairs survive under
/tmp/adamic-parser-incomplete/02597-core.ts-duplicate-15780.ts. Copies are
committed under validation/rebase-incomplete/; the source uses the .input
extension so malformed TypeScript is retained as evidence rather than
checked as an Adamic program. This archive also records its SHA-256.
No claim of parity for all 22,497 inputs is made. The wider recovery work
remains unfinished; this turn ends after the user's requested rebase,
validation and push. The complete repository gate was not run.

The green rebase step was committed and pushed as
d7949bb (Preserve expression byte parity and validate the main rebase).
Its lease-protected rebase push required the remote branch still to name
8ce437f114721b0f997a80fbabecde05e98defbb. The report/evidence follow-up
uses an ordinary fast-forward push. The final report commit is the branch
tip printed in the five-line handoff.

## Case 2597: modifier-led arrow recovery

The discrepancy survives removal of the surrounding declaration and reduces
to `(readonly T)`. Go's arrow lookahead commits when the first parameter token
is a modifier other than async and the following identifier is not as. The
port lacked that recognition, so its ordinary parenthesized expression
parsing produced a different tree and diagnostics. The duplicated closing
parenthesis only exposed the subsequent readonly tuple expression; it was
not itself the missing recovery mechanism. The lookahead now follows Go,
including the async and as exceptions. Five additional probes and preceding
focused cases pass on Go, Node and sanitized native in 35.615s; cohere passes.
