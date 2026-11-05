# 99t: visible checks for debug formatters and a function-expression claim

Task #b5w3ycg. Original base: `dd57a3c669045b9942d6aa014f955e8d2a483ae7`. After 99n landed, the branch was rebased onto main `a7b6da7f622bd768a99a443706490e80667dd547`, including 99k, 99l and 99m. Upstream is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

JavaScript-changing: true. Eleven accepted sort/caller sites become eleven assertions in two source files and four owners. Every retained check changes behavior only when its existing invariant is broken. There is no conversion or new default. The complete source delta is in [changes.patch](changes.patch).

The follow-up explicitly exempts `type<TypeMapper>(this)` in `DebugTypeMapper.__debugToString`. It is unchanged, and `Debug.type`'s overload and empty implementation are unchanged. TypeMapper is the one caller deliberately left by instruction: callback result contracts cannot be checked by any runtime test. Its lowering stays NotYet and fails closed. The follow-up also permits stopping other callers individually. Five further callers stop on measured failures, and ModuleResolutionHost stops on another callback result contract. Thus seven callers remain, for the distinct reasons below; only the function-expression caller is replaced.

## Retained invariants

CW144 and CW145 check the simple mapper's source and target. CW146 and CW147 check every source and present target element as zipWith consumes it. They add no native-array restriction and no preliminary element traversal. The existing undefined-target `any` fallback is preserved. CW148 and CW149 check the deferred source and the deferred target result. The deferred callback runs once, after formatting its source, as before. CW150 and CW151 check each nested mapper through an object-typed view, avoiding an in-test that narrows a union. Every formatter object must be non-null, have the named debug property, and have a callable value there. Each formatter is read once and invoked with Reflect.apply using the original object as receiver. Array fields are captured in the original read order; the existing fallback still performs its original second sources read. Source calls still precede target calls, and mapper1 still precedes mapper2. No DebugType or DebugTypeMapper cast replaces a check.

CW143 first stores the result in a `Node | undefined` slot, so it does not claim SourceFile before checking. It asserts `ts.isSourceFile(sourceFile)`, whose existing body is exactly `return node.kind === SyntaxKind.SourceFile;`, then passes the narrowed SourceFile to getSourceTextOfNodeFromSourceFile. The explicit parseNode conjunction preserves the original correlation and fallback. An absent source still follows the existing empty-string fallback. A present source with the wrong kind fails; it does not route to that fallback.

CW154 first tests FlowFlags.Label. Non-labels retain the predicate's normal false result. On a label, an object-typed alias keeps the FlowNode parameter guarantees while making the presence check operate on one object view rather than narrowing a union. The antecedent property must exist, and its captured value must be undefined or an array. Its getter is read once, preserving the original predicate read count. Undefined is the existing legitimate absent-list case and returns false. A present array returns true. Missing properties and invalid domains fail. The helper contains no FlowNode-to-FlowLabel cast. 99n's hasNode owner is byte-identical.

The replacement in `createSyntacticTypeNodeBuilder.typeFromExpression` asserts `isFunctionLikeDeclaration(node)` and the exact ArrowFunction or FunctionExpression kind. The existing predicate in utilitiesPublic.ts:1685 delegates to a kind switch; the two exact kind comparisons choose these two members of its closed function-like union. The checked field is kind. These are ordinary AST variants, with no additional operator/operand intersection refinement or new host callback contract. Stock 6.0.3 accepts the resulting narrowed uses without a cast.

## Assertion levels

The CLI adaptation's README:72-74 says:

> shouldAssertFunction must be called exactly six times, once inside each of
> the six assertion functions, as `shouldAssertFunction(AssertionLevel.Normal,
> "<its own name>")`.

This does not gate Debug.assert. In the fully applied base debug.ts:192-200, its complete body is:

```ts
if (!expression) {
    message = message ? `False expression: ${message}` : "False expression.";
    if (verboseDebugInfo) {
        message += "\r\nVerbose Debug Information: " + (typeof verboseDebugInfo === "string" ? verboseDebugInfo : verboseDebugInfo());
    }
    fail(message, stackCrawlMark || assert);
}
```

Debug.fail creates its Error at line 178 and unconditionally executes `throw e` at line 182. Neither body consults an assertion level. Therefore these unconditional Debug.assert calls fail at every level. The Node witness also verifies identical failure messages at levels 0, 1, 2 and 3. Level-gated assertNode and its five companions are not used for these checks. R-T80 in docs/rulings.md was read before editing; empty type assertions are not treated as proof.

## Count first

The method follows `typescript/debug-assertion-level-census`'s `stage3/census/debug-assertion-level/REPORT.md`: a complete applied base, a scratch-only instrumented copy, a startup preload exposing the same emitter in new VM contexts, active Mocha titles, and complete pass/fail/pending title/outcome multisets. The original-base uncounted run and proposal census both ran the unmodified stage3/oracle/run.sh, exactly the oracle invoked by stage3/lane/run.sh: all runners, eight workers, no filter, default timeout. Instrumentation observes conditions without throwing, so rejected proposals cannot erase later measurements. All injections remain in scratch copies, not the shipped source transformation.

The final main-stack oracle counts the exact eleven retained assertion conditions in a scratch copy of the accepted candidate, together with fresh non-throwing probes for the six individually stopped callers. Its assertions remain active. The ordinary candidate is separately built without instrumentation for declarations and Node witnesses. [counter-conditions.json](counter-conditions.json) records the proposal predicates; [shipped-counter-conditions.json](shipped-counter-conditions.json) records the actual retained predicates. [counter-results.json](counter-results.json) retains the final main-stack counts and every failing title. [stopped-titles.json](stopped-titles.json) provides the stopped titles without compression.

| Check | Proposal oracle runs | Would fail | Direct startup control runs/failures |
| --- | ---: | ---: | ---: |
| CW143, SourceFile kind | 0 | 0 | 1 / 1 |
| CW144, simple source formatter | 0 | 0 | 1 / 1 |
| CW145, simple target formatter | 0 | 0 | 1 / 1 |
| CW146, each source formatter | 0 | 0 | 1 / 1 |
| CW147, each target formatter | 0 | 0 | 1 / 1 |
| CW148, deferred source formatter | 0 | 0 | 1 / 1 |
| CW149, deferred result formatter | 0 | 0 | 1 / 1 |
| CW150, nested mapper1 formatter | 0 | 0 | 1 / 1 |
| CW151, nested mapper2 formatter | 0 | 0 | 1 / 1 |
| CW154, label antecedent property/domain | 0 | 0 | 1 / 1 |
| ArrowFunction or FunctionExpression | 3,435 | 0 | 1 / 1 |

All ten retained debug-only conditions run zero times in the full oracle. Two extra native-array identity probes also recorded 0 runs and 0 failures, with 1/1 startup controls each, but are not shipped: native-array identity is more than the DebugType element casts add. An array-like witness preserves the original result without that restriction. Zero failures there are not oracle coverage of those paths. Separate startup controls run through the actual instrumented functions, using extracted exact owners for closure-local functions. Each direct control fires its own counter once and fails once. The declaration-with-type-parameters control also reaches the child predicate, which produces one additional, separately titled child failure. Startup events are excluded from every oracle count. [startup-controls.jsonl.gz](startup-controls.jsonl.gz) and [shipped-startup-controls.jsonl.gz](shipped-startup-controls.jsonl.gz) retain these events. Pure condition-evaluation controls were also run before the stronger actual-path controls; they are not substituted for the latter.

## Every Debug.type caller

AST `findReferences`, using stock TypeScript 6.0.3 against the complete applied base, returns one symbol group, its two declarations, and eight callers. See [references.json](references.json) and [callers.json](callers.json).

| Caller and claim | Runs | Would fail | Decision and fields examined |
| --- | ---: | ---: | --- |
| debug.ts, TypeMapper | not counted | not counted | Explicitly left. Function mapper func and deferred targets promise Type results; recursive mapper fields also carry this claim. Callability cannot establish the callback result contracts. |
| utilities.ts, DeclarationWithTypeParameters | 5,100 | 4,230 | Stopped. The exact AST kind union, including JSDoc typedef/callback/signature kinds, is checked in the proposal census. |
| utilities.ts, DeclarationWithTypeParameterChildren | 5,120 | 4,230 | Stopped. The exact signature, class-like, interface, type-alias and JSDoc template kind union is checked. |
| utilities.ts, PrimitiveLiteral | 7,027 | 1,788 | Stopped. Literal kind, or PrefixUnaryExpression kind plus operator and operand kind. Plus requires NumericLiteral; minus permits NumericLiteral or BigIntLiteral. These are the extra intersection refinements in PrimitiveLiteral. |
| utilities.ts, HasInferredType | 399,823 | 19,422 | Stopped. The exact variable-like/access/binary/export-assignment kind union is checked. |
| expressionToTypeNode.ts, ArrowFunction or FunctionExpression | 3,435 | 0 | Retained. Function-like kind and the two exact kinds, as explained above. |
| transformers/declarations/diagnostics.ts, WithIsolatedDeclarationDiagnostic | 1,014 | 26 | Stopped. The exact eighteen-kind union from the local type alias is checked. |
| moduleSpecifiers.ts, ModuleResolutionHost | 1,170 | 0 | Stopped after a passing shape probe. It checks required readFile callability and optional getDirectories absence/undefined/callability. The source ModuleSpecifierResolutionHost already declares readFile's result signature, but does not declare getDirectories. X adds its optional `(path: string) => string[]` contract, which no callability test proves. No narrower substitute claim or receiver-changing host copy is introduced. |

All five nonzero sites remain byte-for-byte unchanged and their proposed checks were never built into the candidate. Their inputs are legitimate calls to general predicates or diagnostic fallback logic. The full failing-title maps are retained, including repeated evaluations within the same test. The ModuleResolutionHost counts describe only its shape probe, not proof of its full claim. TypeMapper remains the explicit exception; the other six remaining callers are individually stopped under the follow-up's rule.

## Adapter contract and proofs

The stock 6.0.3 parser finds named lexical owners, including the DebugTypeMapper class method. Each complete before/after owner text has its own SHA-256 pin, and each old site is also checked by exact text within that owner. No whole-file hash gates an edit. All owners must be wholly before or wholly after. Changed, missing, duplicate, partial or mixed owners fail before writes. Every planned file is parsed and all after-owner pins checked, then every input is reread before the first write. CRLF is preserved. Replay makes zero edits and writes. Unrelated text outside these owners is accepted. [contract-results.json](contract-results.json) records twenty contract cases, including sixteen no-write refusal cases.

[witness-before.json](witness-before.json) and [witness-after.json](witness-after.json) are byte-identical. Both enable debug info before formatting a real checker type, all mapper variants, a flow graph, a source file and a statement. Both also emit declarations containing arrow and function expressions. The deferred callback executes once. Additional witnesses preserve one array-source read, one formatter-getter read, the normal two label antecedent reads across predicate and consumer, and receiver identity even when a function has an overridden call property. [mutant-control.json](mutant-control.json) is green; [mutant.json](mutant.json) is red. Removing only the hasAntecedents assertion lets an iterable object masquerade as an antecedent array and returns a graph instead of the required Debug Failure.

[comparison.json](comparison.json) proves that only debug.ts and expressionToTypeNode.ts change among the complete source set. Both public declaration files are byte-identical: typescript.d.ts is 593,502 bytes and tsserverlibrary.d.ts is 866 bytes. Stock TypeScript checks all 79 roots with zero diagnostics. The final main-stack oracle has 106,366 passes and the one sanctioned API failure, with an exact title/outcome match to main and the uncounted reference. The remaining apply, lane, census and final branch evidence is recorded in PROOF.md and CENSUS.md.

The immediate pre-push fetch advanced main to `df44f6f5c8597d851ec5d742c1e86bebb9da5746`. The branch was rebased again. These two commits change only rulings documentation and lane test wrappers, not compiler or adaptation inputs. The full apply, keyed replay and lane checks are repeated after this rebase; compiler measurements remain bound to their recorded a7b6da7f build.
