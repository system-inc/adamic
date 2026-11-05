# 99u: false Debug.type claims leave

Task #b5w3ycg. Amended Node parameter ruling is recorded at the end.
Branch base: main `9a56f1c0760b444f1de85ccb4d44a9713a5d4d1e`.
Combined-stack oracle source base: main `31d7a244808fa8092643926d21ab9d310a54cce3`.
Complete Node and 79-root census measurements use `b0efce6dce1cd7445ed8a570aecac387a4a7d684`.
The a3f94672 advance was Go lowering only. Main then acquired 99r. The combined
99r/99s/99u stack has a separate full oracle and counter verification below.
Earlier complete compiler measurements retain their exact commit bindings;
focused compiler verification is recorded in CENSUS.md.
Upstream: TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.
99t's README.md, references.json and callers.json at
`9b3410c37fd40d75161d628ac75d90740abd4d4e` were read first. The original
stop commit remains on typescript/debug-type-claims-stop; this branch starts
from main and contains a real adapter, not the refusing stop adapter.

javascriptChanging: true. Five empty Debug.type calls and the first owner's
empty assertType call are dropped. Ten new Debug.assert calls perform tsc's
existing case-kind tests. Four remaining empty assertType calls now receive
node.kind, with the generic Exclude argument erased. Those calls therefore
add a kind read in default; they are compile-time kind pins, not runtime checks
of X. There are no new locals, guard functions or casts. The leftover default cast is removed by widening createExpressionError's
parameter from Expression to Node. That follow-up is type-only. changes.patch records
the complete source delta. No other adaptation is changed.

## Measure the kind pins first

kind-pin-measure.json is the preliminary stock 6.0.3 measurement before the
census and adaptation build. Each of the five owners has zero diagnostics.
Its isolated stock profile lacks Node ambient declarations and has 59 unrelated
environment diagnostics; none is in these owners. The later complete stock
check supplies stage3/api's existing pinned Node ambient declarations and has
zero diagnostics over all 79 roots both before and after. Adamic types also
reads zero over the same 79 roots before and after.

| Owner | Pin before | Pin after |
| --- | --- | --- |
| isDeclarationWithTypeParameters | `assertType<DeclarationWithTypeParameterChildren>(node)` | `deleted` |
| isDeclarationWithTypeParameterChildren | `assertType<never>(node)` | `assertType<Exclude<SyntaxKind, DeclarationWithTypeParameterChildren["kind"]>>(node.kind)` |
| isPrimitiveLiteralValue | `assertType<never>(node)` | `assertType<Exclude<SyntaxKind, PrimitiveLiteral["kind"]>>(node.kind)` |
| hasInferredType | `assertType<never>(node)` | `assertType<Exclude<SyntaxKind, HasInferredType["kind"]>>(node.kind)` |
| createGetIsolatedDeclarationErrors.getDiagnostic | `assertType<never>(node)` | `assertType<Exclude<SyntaxKind, WithIsolatedDeclarationDiagnostic["kind"]>>(node.kind)` |

Each Debug.type line is deleted. Case-local narrowing uses only existing tsc
guards through Debug.assert. Return-only cases need no new guard. The first
owner continues to use the existing child predicate on its default path.

## Count before building

The proposal is counted in a scratch copy of the complete applied main tree,
with non-throwing probes at the exact proposed assertion sites. The source
still has its original Debug.type and node pins during this census. Each probe
uses the actual imported tsc guard body. No rejected assertion is built into
the counted tree. Reference, counted and accepted-candidate runs each use the
unmodified stage3/oracle/run.sh: all runners, eight workers, no filter, default
timeout. counter-preload.cjs observes original Mocha events in workers and host
replay and exposes the same counter in new VM contexts, following 99t's method.

| ID | Owner and case group | Existing guard condition | Runs | Would fail | Decision |
| --- | --- | --- | ---: | ---: | --- |
| G01 | isPrimitiveLiteralValue: PrefixUnaryExpression | `isPrefixUnaryExpression(node)` | 288 | 0 | ship |
| G02 | getDiagnostic: GetAccessor / SetAccessor | `isGetAccessor(node) || isSetAccessor(node)` | 29 | 0 | ship |
| G03 | getDiagnostic: ComputedPropertyName / ShorthandPropertyAssignment / SpreadAssignment | `isComputedPropertyName(node) || isShorthandPropertyAssignment(node) || isSpreadAssignment(node)` | 126 | 0 | ship |
| G04 | getDiagnostic: ArrayLiteralExpression / SpreadElement | `isArrayLiteralExpression(node) || isSpreadElement(node)` | 23 | 0 | ship |
| G05 | getDiagnostic: MethodDeclaration / ConstructSignature / FunctionExpression / ArrowFunction / FunctionDeclaration | `isMethodDeclaration(node) || isConstructSignatureDeclaration(node) || isFunctionExpression(node) || isArrowFunction(node) || isFunctionDeclaration(node)` | 311 | 0 | ship |
| G06 | getDiagnostic: BindingElement | `isBindingElement(node)` | 24 | 0 | ship |
| G07 | getDiagnostic: PropertyDeclaration / VariableDeclaration | `isPropertyDeclaration(node) || isVariableDeclaration(node)` | 308 | 0 | ship |
| G08 | getDiagnostic: Parameter | `isParameter(node)` | 88 | 0 | ship |
| G09 | getDiagnostic: PropertyAssignment | `isPropertyAssignment(node)` | 61 | 0 | ship |
| G10 | getDiagnostic: ClassExpression | `isClassExpression(node)` | 18 | 0 | ship |
| G11 | getDiagnostic: default | `isExpression(node)` | 26 | 12 | no assertion; widen helper parameter to Node |

Grouped case labels share one condition and one counter; their runs are the
combined group total. The other three owners have no new Debug.assert site.
Every shipped condition has nonzero oracle coverage and zero failures.

The default isExpression proposal fails 12 of its 26 evaluations, so it is not
shipped. The original 99u retained `return createExpressionError(node as Expression);`.
The amended ruling instead accepts Node in createExpressionError and calls it
with `node`, without any new check. The kind pin stays.

| Default census | Runs | Non-Expressions | Kinds |
| --- | ---: | ---: | --- |
| Original 99u | 26 | 12 | six MethodSignature, six TypeReference |
| Amended Node parameter | 26 | 12 | six MethodSignature, six TypeReference |

The amended recount has the identical title-and-kind multiset, including all
twelve occurrences listed below. See amended-default-recount.json. Its full
oracle has 106,366 passes and the sanctioned failure; its title/outcome multiset
exactly matches main (amended-recount-v-main-identity.json). These are all failing titles, including hook and test
multiplicities rather than deduplicating them away:

- 3: `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts"`
- 3: `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts Correct JS output for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts`
- 3: `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts"`
- 3: `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts Correct JS output for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts`

All eleven planted counter controls fire exactly once and fail exactly once.
They execute the exact instrumented lexical owner with a changing-kind getter:
the switch sees its case kind and the guard sees Unknown. Existing tsc guards
are used; unrelated closure dependencies are stubbed. The default control uses
Unknown directly. Startup events are isolated from all oracle counts.
startup-controls.jsonl.gz and startup-control-paths.json retain the evidence.

The separate guard-runtime-witness.json executes every shipped assertion through
its accepted owner, with the actual built compiler's Debug and guard functions.
All ten fail on the planted input at each assertion level 0, 1, 2 and 3, forty
expected failures total. This uses main's existing assertion-level adaptation.
It does not treat the planted changing-kind object as an ordinary AST input.

## Guard bodies for Language

Every shipped guard below reads only node.kind. No operator, operand, parent,
callback or other field is read by a shipped guard. Existing operand checks in
isPrimitiveLiteralValue remain unchanged. Exact bodies, source positions and
the unshipped isExpression dependency chain are also in guard-bodies.json.

### isSetAccessor

`src/compiler/utilitiesPublic.ts:2538`. Reads beyond kind: no.

```ts
export function isSetAccessor(node: Node): node is SetAccessorDeclaration {
    return node.kind === SyntaxKind.SetAccessor;
}
```

### isGetAccessor

`src/compiler/utilitiesPublic.ts:2542`. Reads beyond kind: no.

```ts
export function isGetAccessor(node: Node): node is GetAccessorDeclaration {
    return node.kind === SyntaxKind.GetAccessor;
}
```

### isComputedPropertyName

`src/compiler/factory/nodeTests.ts:397`. Reads beyond kind: no.

```ts
export function isComputedPropertyName(node: Node): node is ComputedPropertyName {
    return node.kind === SyntaxKind.ComputedPropertyName;
}
```

### isParameter

`src/compiler/factory/nodeTests.ts:408`. Reads beyond kind: no.

```ts
export function isParameter(node: Node): node is ParameterDeclaration {
    return node.kind === SyntaxKind.Parameter;
}
```

### isPropertyDeclaration

`src/compiler/factory/nodeTests.ts:422`. Reads beyond kind: no.

```ts
export function isPropertyDeclaration(node: Node): node is PropertyDeclaration {
    return node.kind === SyntaxKind.PropertyDeclaration;
}
```

### isMethodDeclaration

`src/compiler/factory/nodeTests.ts:430`. Reads beyond kind: no.

```ts
export function isMethodDeclaration(node: Node): node is MethodDeclaration {
    return node.kind === SyntaxKind.MethodDeclaration;
}
```

### isConstructSignatureDeclaration

`src/compiler/factory/nodeTests.ts:454`. Reads beyond kind: no.

```ts
export function isConstructSignatureDeclaration(node: Node): node is ConstructSignatureDeclaration {
    return node.kind === SyntaxKind.ConstructSignature;
}
```

### isBindingElement

`src/compiler/factory/nodeTests.ts:570`. Reads beyond kind: no.

```ts
export function isBindingElement(node: Node): node is BindingElement {
    return node.kind === SyntaxKind.BindingElement;
}
```

### isArrayLiteralExpression

`src/compiler/factory/nodeTests.ts:576`. Reads beyond kind: no.

```ts
export function isArrayLiteralExpression(node: Node): node is ArrayLiteralExpression {
    return node.kind === SyntaxKind.ArrayLiteralExpression;
}
```

### isFunctionExpression

`src/compiler/factory/nodeTests.ts:612`. Reads beyond kind: no.

```ts
export function isFunctionExpression(node: Node): node is FunctionExpression {
    return node.kind === SyntaxKind.FunctionExpression;
}
```

### isArrowFunction

`src/compiler/factory/nodeTests.ts:616`. Reads beyond kind: no.

```ts
export function isArrowFunction(node: Node): node is ArrowFunction {
    return node.kind === SyntaxKind.ArrowFunction;
}
```

### isPrefixUnaryExpression

`src/compiler/factory/nodeTests.ts:636`. Reads beyond kind: no.

```ts
export function isPrefixUnaryExpression(node: Node): node is PrefixUnaryExpression {
    return node.kind === SyntaxKind.PrefixUnaryExpression;
}
```

### isSpreadElement

`src/compiler/factory/nodeTests.ts:660`. Reads beyond kind: no.

```ts
export function isSpreadElement(node: Node): node is SpreadElement {
    return node.kind === SyntaxKind.SpreadElement;
}
```

### isClassExpression

`src/compiler/factory/nodeTests.ts:664`. Reads beyond kind: no.

```ts
export function isClassExpression(node: Node): node is ClassExpression {
    return node.kind === SyntaxKind.ClassExpression;
}
```

### isVariableDeclaration

`src/compiler/factory/nodeTests.ts:792`. Reads beyond kind: no.

```ts
export function isVariableDeclaration(node: Node): node is VariableDeclaration {
    return node.kind === SyntaxKind.VariableDeclaration;
}
```

### isFunctionDeclaration

`src/compiler/factory/nodeTests.ts:800`. Reads beyond kind: no.

```ts
export function isFunctionDeclaration(node: Node): node is FunctionDeclaration {
    return node.kind === SyntaxKind.FunctionDeclaration;
}
```

### isPropertyAssignment

`src/compiler/factory/nodeTests.ts:995`. Reads beyond kind: no.

```ts
export function isPropertyAssignment(node: Node): node is PropertyAssignment {
    return node.kind === SyntaxKind.PropertyAssignment;
}
```

### isShorthandPropertyAssignment

`src/compiler/factory/nodeTests.ts:999`. Reads beyond kind: no.

```ts
export function isShorthandPropertyAssignment(node: Node): node is ShorthandPropertyAssignment {
    return node.kind === SyntaxKind.ShorthandPropertyAssignment;
}
```

### isSpreadAssignment

`src/compiler/factory/nodeTests.ts:1003`. Reads beyond kind: no.

```ts
export function isSpreadAssignment(node: Node): node is SpreadAssignment {
    return node.kind === SyntaxKind.SpreadAssignment;
}
```

### Existing operand guards, unchanged

These are existing calls, not new assertions or additional counter sites. Both
read only node.kind.

`src/compiler/factory/nodeTests.ts:235`:

```ts
export function isNumericLiteral(node: Node): node is NumericLiteral {
    return node.kind === SyntaxKind.NumericLiteral;
}
```

`src/compiler/factory/nodeTests.ts:239`:

```ts
export function isBigIntLiteral(node: Node): node is BigIntLiteral {
    return node.kind === SyntaxKind.BigIntLiteral;
}
```

### isExpression, measured but not shipped

`src/compiler/utilitiesPublic.ts:2086`:

```ts
export function isExpression(node: Node): node is Expression {
    return isExpressionKind(skipPartiallyEmittedExpressions(node).kind);
}
```

This guard is not kind-only in the presence of partially emitted wrappers.
It delegates through skipPartiallyEmittedExpressions to skipOuterExpressions,
which can read node.expression while unwrapping. isExpressionKind then tests
the resulting kind through isUnaryExpressionKind/isLeftHandSideExpressionKind.
The complete helper bodies and locations are in guard-bodies.json. The broad
isOuterExpression helper has other flag-dependent paths, but this call passes
only PartiallyEmittedExpressions. The guard's twelve observed failures, not
its name or apparent intended contract, explain why no isExpression assertion is shipped. The amended helper accepts Node.

## Default failures for Adamic: all twelve occurrences

These are twelve failing evaluations, not twelve distinct title strings. There
are four distinct hook/test titles, each evaluated three times. All occurrences
are retained below, with the actual node.kind recorded at the failing default
probe. Numeric kinds are independently resolved with pinned TypeScript 6.0.3:
174 is MethodSignature and 184 is TypeReference. There are six of each kind.
The amended fallback is `return createExpressionError(node);`, with the same kind pin.
No isExpression assertion is shipped.

| Occurrence | Exact active title | Node kind | Numeric kind |
| ---: | --- | --- | ---: |
| 1 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts"` | MethodSignature | 174 |
| 2 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts"` | MethodSignature | 174 |
| 3 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts"` | MethodSignature | 174 |
| 4 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts Correct JS output for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts` | MethodSignature | 174 |
| 5 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts Correct JS output for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts` | MethodSignature | 174 |
| 6 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts Correct JS output for tests/cases/compiler/isolatedDeclarationErrorsClasses.ts` | MethodSignature | 174 |
| 7 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts"` | TypeReference | 184 |
| 8 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts"` | TypeReference | 184 |
| 9 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts "before all" hook for "Correct errors for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts"` | TypeReference | 184 |
| 10 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts Correct JS output for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts` | TypeReference | 184 |
| 11 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts Correct JS output for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts` | TypeReference | 184 |
| 12 | `compiler tests compiler tests for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts Correct JS output for tests/cases/compiler/isolatedDeclarationsAddUndefined.ts` | TypeReference | 184 |

default-expression-failures.json and the compressed raw event stream retain
the twelve observations. combined-stack-counts.json records 26 default runs,
12 failures, and the unchanged zero-failure counts at all ten shipped sites.

## Oracle identity and Node witness

All three full oracle runs finish with 106,366 passes, one sanctioned API
failure and zero pending tests. Their only baseline difference is
api/typescript.d.ts; the failure title is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.

Both the counted and built-candidate raw title/outcome multisets match the
uncounted main reference exactly, with zero additions and removals. Each event
occurs twice through worker and host replay: 212,734 raw events normalize to
106,367 occurrences. Every multiplicity is even; duplicate titles retain their
multiplicity. The normalized multiset SHA-256 is
`1de5705e5a5982dc8172377c138f37e6c65f6e8a60dbe76f160a85a181152371`.
The raw streams, comparison script, summaries and complete oracle reports are
retained. This full oracle is the Node witness on the accepted source, not only
a small fixture or total-count comparison.

## Controls

kind-pin-mutant.json drops only the Parameter case label from hasInferredType.
Stock TypeScript 6.0.3 then reports exactly one TS2345 diagnostic at the kind
pin, naming SyntaxKind.Parameter. The unmutated full 79-root stock check has
zero diagnostics. The existing case guards and all other owner code stay fixed.

For the lowering control, exactly one deleted line is restored at its original
site in isDeclarationWithTypeParameters:
`Debug.type<DeclarationWithTypeParameters>(node);`. Stock checks all 79 roots
with zero diagnostics. The meter's direct owner attempt records a failed
statement at that restored call, carrying exactly the text of a NotYet rooted
in binder.ts:1034:17, "writing a checked object field without its source-slot
type certificate". The NotYet is dependency-located; it is not relabeled as a
diagnostic whose own coordinate is the Debug.type call. debug-type-mutant.json
retains both coordinates and the full correlated trace. mutant-owner.go.txt
uses the existing guarded full attempt path and checks both no-output guards.
The adapted owner lacks that restored-call boundary; its later child predicate
still has a separate dependency wall. No claim is made that the full owner or
compiler is generally accepted after this adaptation.

The HasInferredType restoration was tried first. Its signature narrows node to
never after the exhaustive switch, so the new pin's node.kind is ill-typed:
TS2339, Property 'kind' does not exist on type 'never'. It never reached lowering
and is not used as the lowering control. That exact checker result is retained
in has-inferred-restoration-mutant-types.log.gz. The first owner has no moved
kind pin, so its single-line restoration isolates lowering instead.

## Adapter contract

Stock 6.0.3 locates the six named lexical owners, checks exact old site text
and pins each complete before/after owner with SHA-256. Hashes are owner-scoped,
never whole-file edit gates; CRLF and LF are preserved. Required guard imports
are validated structurally in the named namespace import, including duplicates,
aliases, conflicts and mixed before/after import states. They are added only
with an entirely before-state set of owners.

All owners must be wholly before or wholly after. Missing, changed, duplicate,
partial or mixed owners and partial/mismatched imports are refused before any
write. All planned files are reparsed and after pins checked; every input is
reread before the first write. Unrelated comments outside the pinned owners are
accepted. The original contract-results.json records 43 passing cases, including 38 no-write
refusals, two line-ending profiles, unrelated edits and a no-write replay that
preserves mtime and bytes. Direct replay reports files:0 and edits:0.

The real stage3/apply.py adaptation_inputs() sort places 99u after every
99-correctness-axolotl directory on this main tip; sort.json retains the order.
Full apply passes. The keyed second apply restores an identical source payload,
and the direct adapter replay is a no-op. The executable adapter and pins used
by the measurements are frozen in provenance.json. Evidence archives and prose
added afterward affect the product key but do not change the transformation.

## Latest main compiler verification

After the Go-only main advance, full apply was repeated and all 746 source
files matched the oracle-tested tree exactly. The latest ordinary compiler
again reads zero over all 79 roots before and after. Both edited files were
recensused before and after, and the existing full-body meter selector attempted
all five owners on each side with all 79 roots loaded and no checker diagnostics.
The findings match the earlier owner observations. The exact single-line
Debug.type mutant again fails through a dependency-located NotYet at the
restored call boundary. The lane Go fixtures pass again. The complete 79-root
ordinary censuses remain explicitly bound to their b0efce6d compiler; they are
not relabeled as having used the later compiler.

## API, checker, lane and census proofs

Only utilities.ts and transformers/declarations/diagnostics.ts change among all
746 source files, solely by the planned owners and required guard imports.
The independent measured source copy is byte-identical to full apply. Debug.type's
own declaration, debug.ts's TypeMapper call, moduleSpecifiers.ts and
expressionToTypeNode.ts are unchanged.

All four utility owners are exported internal functions. Their signatures are
unchanged. Both public d.ts files remain byte-identical:

- typescript.d.ts: 593502 bytes, SHA-256 `c4edc2ad37cf1afeb4704ccb4bacf7c76b9f538a8e5685e4b276e2e7dd52796b`.
- tsserverlibrary.d.ts: 866 bytes, SHA-256 `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`.

Both 79-root Adamic types checks exit 0. Both complete stock 6.0.3 checks also
have zero diagnostics when supplied with the pinned Node ambient declarations.
The complete stock options and mutant diagnostic are retained. generate_shards
--check and the stage3/lane Go fixture command pass. The full oracle is run
separately as described above.

CENSUS.md records all 79 ordinary per-file results before and after, the edited
files' exact first-stop lines, and complete before/after findings for each of
the five owners through the meter's existing full-body selector. The production
compiler and refusal policy are unmodified. Remaining dependency walls are not
hidden by total counts or described as newly accepted compiler roots.

PROOF.md records commands, artifacts, pilots and landing verification. Every
Go source added by this directory ends in .go.txt.

## Combined stack after 99r landed

The local commit was rebased onto 31d7a2448. Full apply and zero-edit replay
passed with 99r, 99s and 99u in the real sort order. The main reference is the
completed full apply with only 99u's five exact pinned owner edits and guard
imports reversed. All other main changes remain. A separate scratch copy keeps
the accepted 99u assertions active and observes each condition without throwing
from the probes, including the rejected default proposal. No rejected assertion
is added to the application source.

Both use the unmodified lane oracle: all runners, eight workers, no filter and
default timeout. The combined accepted stack has 106,366 passes, one sanctioned
API failure and no pending tests, with only api/typescript.d.ts as a baseline
difference. Its raw and normalized title/outcome multisets exactly match the
combined main reference, with zero additions/removals. The ten shipped guard
counts and the default 26/12 count are unchanged from the count-before-build
measurement. The guard code and owner pins are unchanged.

The workspace disconnected after the reference test process completed but
before the wrapper wrote report.json. Its completed test log, all 212,734 raw
identity events and all baseline files were recovered and checked. The sole
failure and sole API diff are confirmed. No missing wrapper exit status or
wall time is invented; combined-main-recovered-report.json marks this explicitly.
The accepted-stack wrapper completed normally and retains its full report.

## Amended ruling: createExpressionError accepts Node

The adapted body was read in full. Its uses of the parameter are
findNearestDeclaration(node: Node), node.parent, and createDiagnosticForNode(node, ...).
It reads no Expression-only member. Its body is unchanged; its first parameter
is now node: Node. getDiagnostic's default passes node without a cast, and the
Exclude kind pin remains. No guard or check is added.

createGetIsolatedDeclarationErrors.createExpressionError now has its own exact
before/after owner text and SHA-256 pins in pins.json. getDiagnostic's after pin
also includes the cast removal. The adapter validates all six owners before
any write and refuses partially applied old 99u trees. expression-parameter-review.json
records both helper forms and the body review. The added helper is internal.

The original count above was 26 default nodes with 12 non-Expressions: six
MethodSignature and six TypeReference. The amended-tree recount and its exact
identity comparisons are recorded below with the final verification artifacts.

The amendment adds no JavaScript change: the stock diagnostics.ts emission is
byte-identical to original 99u (41,145 bytes, SHA-256
4dd16539f49992214e35f796178b9e1706d01ac826651ba49564014116ad4209).
The actual built compiler JavaScript also matches byte-for-byte. Both public
d.ts files remain identical. See amended-emitted.json and amended-build-identity.json.
Adamic types reads zero over all 79 roots; stock 6.0.3 also reports zero.
All 50 owner-contract cases pass, including 45 no-write refusals. Full apply,
zero-edit replay, generate_shards --check and the documented lane Go test pass.
Earlier count and census artifacts retain their original version bindings.

The completed amended recount is still **26 runs, 12 non-Expressions**, with
exactly the same six MethodSignature and six TypeReference observations and
the same titles/multiplicities. No node moved between branches. Raw counters,
failing nodes, titles and oracle report are retained in amended-default-* and
amended-recount-* artifacts. The default counter's planted failing input fires
once. The proposed isExpression assertion remains unshipped.

The separate uncounted amended oracle also completes with 106,366 passes,
one sanctioned failure and no pending tests. Its raw and normalized title/outcome
multisets exactly match both main and the recount, with zero additions/removals.
See amended-uncounted-oracle.json, amended-uncounted-v-main-identity.json and
amended-recount-v-uncounted-identity.json.

Final landing verification on main 9a56f1c07 repeats full apply and zero-edit
replay. All 81,500 tracked upstream build/test inputs match the oracle-tested
amended tree byte-for-byte, including all 746 source files. No stage3 adaptation
or oracle harness changed in the main advance. The fresh landing compiler reads
zero over 79 roots before and after, and the shard and lane Go checks pass again.
See amended-landing.json and amended-current-* artifacts.
