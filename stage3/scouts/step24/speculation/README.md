# Step 24: speculation as written

Pinned input: TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. The ruling keeps lookAhead,
tryParse, scanner rewinds and JavaScript truthiness directly. No adaptation or
compiler option was changed by this scout.

The complete AST inventory is [inventory.md](inventory.md), with callback
signatures, body return expressions, assertion inputs and consumer contexts in
[inventory.json](inventory.json). It records 94 calls: 60 lookAhead, 20 tryParse,
five speculationHelper, two tryScan, two scanRange and five resetTokenState.
There are also 21 rescan calls and five setText calls, listed separately.
setTextPos has zero calls. Declarations and exported aliases are not call sites.
Stock TypeScript 6.0.3 reports zero semantic diagnostics over the compiler project.
The meter checks the pin and exact parser/scanner bytes before recording results.

## Result proof and specialization

67 callback results have one proven runtime representation family. Numeric
Tristate and SyntaxKind results stay numbers, including zero; boolean literals
stay booleans. One family does not prove one object shape. The checker result
column honors explicit return annotations; `bodyReturns` separately records
return-expression types, and `assertions` records types before and after casts.

22 sites are marked as checked specialization boundaries: six generic forwarding
calls with unresolved T, 11 object/undefined callbacks, two string/undefined
callbacks, two object/false callbacks, and one assertion hiding alternatives.
This marking is conservative: a known closed union can be represented directly.
It needs a check only when a specialization projects an arm without a proof.
A proven control-flow narrowing can discharge that check. Generic forwarding is
instantiated at its callers, never permanently coerced to boolean.

At parser.ts:9695 the callback checker result is
`JSDocTemplateTag | JSDocParameterTag`, but the asserted expression can also be
`false`, JSDocPropertyTag, JSDocTypeTag or JSDocThisTag. The while condition must
observe false before any checked projection; checking for a node before the
truthiness decision would turn ordinary parse failure into a panic.

The scanner helper (scanner.ts:3930) restores six fields on `!result || isLookahead`.
The parser helper (parser.ts:2256) additionally restores currentToken, diagnostic
length and the pending parse-error flag on `!result || kind !== TryParse`.
Reparse keeps its diagnostics, and asserts unchanged context flags. scanRange
(scanner.ts:3952) unconditionally restores eight fields, including end and comment
directives. resetTokenState (scanner.ts:4019) resets position and token state.
Numeric result consumers at parser.ts:5238 and :5398 compare Tristate; :7243 and
:7757 compare SyntaxKind. None may substitute a boolean for the returned value.

## Four trims and their semantic mutants

| Fixture | Real site | Trim | Mutant and catcher |
|---|---|---|---|
| lookahead.a | parser.ts:5238, worker :5251 | Keep the closed-parenthesis worker branch and scanner helper; token enums become numeric witnesses | Remove unconditional lookahead rewind; source Node stdout differs |
| tryparse-boolean.a | parser.ts:2755, callback const arm :2767 | Keep next-token equality and helper; lexical production is deterministic | Remove false-result rewind; source Node, native and backend stdout differ |
| tryparse-node.a | parser.ts:7758 | Keep literal text comparison and node/undefined callback; literal parsing is one small object | Reverse constructor text comparison; source Node stdout differs |
| scan-range.a | parser.ts:8911, scanner.ts:3952 | Keep unconditional restoration; JSDoc is a comment object and directives a scalar witness | Omit directive restoration; source Node, native and backend stdout differ |

All four fixtures preserve the helpers' direct calls and result tests. The first
three embed the scanner helper; tryParse's forwarding uses scanner tryScan's
helper semantics. They do not model the parser's diagnostic array or full grammar.
The lookahead fixture returns valid zero, and boolean parsing returns valid false.
scan-range also instantiates its real generic helper with empty string, proving
that the result is preserved. That extra instantiation is a helper contract probe,
not a claim that upstream doJSDocScan returns a string. Token production, node
payloads and unused enum cases are trimmed. resetTokenState's unused undefined
token-value slot is represented by an empty string witness in the range trim.

## Run

Use an external pinned checkout, never add upstream source to Adamic. Generate
its diagnostics and install its locked dependencies before inventorying it:

```sh
source /workspace/adamic-tools/env.sh
# Set tree to your external checkout at the pinned commit.
npm ci --prefix stage3/api > /tmp/step24-api.log 2>&1
export STAGE3_TYPESCRIPT="$PWD/stage3/api/node_modules/typescript"
(cd "$tree" && npm ci && node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json) > /tmp/step24-source.log 2>&1
node stage3/scouts/step24/speculation/inventory.cjs "$tree" > /tmp/step24-inventory.log 2>&1
go build -o /tmp/step24-adamic ./cmd/adamic > /tmp/step24-build.log 2>&1
node stage3/scouts/step24/speculation/check.cjs /tmp/step24-adamic > /tmp/step24-check.log 2>&1
```

The runner typechecks every baseline and mutant with stock 6.0.3 under strict,
exactOptionalPropertyTypes, noUncheckedIndexedAccess, verbatimModuleSyntax and
erasableSyntaxOnly. It compares stock-erased source and direct .a source on Node,
then sanitized counted native and backend Node where compilation succeeds.
Mutants live only in scratch, must typecheck and run successfully, and must differ
in stdout. The two compilable fixtures also run their mutants through both backends.
Unexpected compiler failures fail the runner; only the two measured NotYet
messages are accepted as scout gaps. A future successful build is compared to Node.

## Observed coverage and gaps

[observations.json](observations.json) records raw results; [counts.md](counts.md)
is the refreshed local fixture registry. The four source baselines pass and all
four mutants are killed. Boolean tryParse and scanRange pass under ASan/UBSan,
with balanced allocation/free counts, and through backend Node. Numeric lookahead
and node/undefined tryParse cannot currently compile: stage 0 reports
`can't lower a PrefixUnaryExpression on a number yet` and
`can't lower a PrefixUnaryExpression on a value yet`, respectively, at `!result`.
The native promise is therefore partial, not four compiled fixtures. Closing
those production lowering gaps or implementing checked result specialization is
outside this scout's territory. No compiler file, central oracle fixture registry,
upstream suite, whole package test or full gate was changed or run.

Setup used GOPROXY `https://proxy.golang.org|direct`. Timing lines: Go ready
0.083s, Node ready 0.090s, clang ready 0.452s, markdown installed 0.698s,
markdown ready 0.845s, submodules ready 2.881s, Go build ready 164.693s,
test binaries deferred 164.795s, cache warm 164.797s, done 164.828s.
`nproc=5`; cgroup cpu.max=`400000 100000` (four CPUs). Base Adamic commit:
`45487a809f89885a3fc651cd590e7dabf31362dc`. Setup and all measurements used
`/workspace/adamic-tools/env.sh`. Recursive fetch was stopped while fetching
unneeded nested submodules; `git fetch --no-recurse-submodules origin` completed,
and setup selected the main-pinned submodules. Full command output is in
`/tmp/step24-{setup,source,generate,npm,build,inventory,check}.log` for this session.
