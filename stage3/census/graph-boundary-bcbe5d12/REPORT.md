# Graph adoption boundary census on bcbe5d12

Task jawqwv7. Report only. Counts cover lowered unit results only, including retained successful statements after full-mode rollback. They do not measure accepted whole programs.

Compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`. Reference report tree: `30057eca`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `755daee0c6ba`; cgroup CPU quota `400000 100000`; memory limit `17179869184` bytes; 5 visible CPUs. Environment: `ccarenv_b64_Y2NhcmVudl85M2MyY2I4NzA3ZTA4MTkxODMwYTI0NDZlY2ZhNjAyZg`.

## Four-way split, lowered IR only

| Kind | Observations | Distinct reported locations |
| --- | ---: | ---: |
| views | 37886 | 391 |
| record_literals | 0 | 0 |
| dynamic_reads | 28 | 21 |
| checked_object_calls | 0 | 0 |

Views combine 1314 ViewOrigins entries and 36572 CheckedFields keys, as requested; these are metadata observations, not that many distinct adapter expressions.

The 79 compiler-file subtotal is `{'views': 37886, 'record_literals': 0, 'dynamic_reads': 28, 'checked_object_calls': 0}`. The two entry-file subtotal is `{'views': 0, 'record_literals': 0, 'dynamic_reads': 0, 'checked_object_calls': 0}`. Both cover lowered unit results only.

Long pole by observed count: **views, 37886**. Coverage caveat: lowered IR only; failed statements, checker-skipped bodies, deferred generics and uncompleted units are absent. Inference: this ranking may differ after those gaps lower.

## Coverage and interpretation

Coverage continuation: 16/51 previously unstarted files now have a single-file full-mode attempt; 35 remain. executeCommandLine.ts is also retried because the earlier emitter piece ended before reaching it. Every continuation child is capped at 1490 seconds, leaving shutdown time below 25 minutes. A nonzero or timed-out single-file attempt is recorded and the runner moves on. 0 continuation files exited nonzero so far.

Bounded-pass stops: first: child exit -15, configured elapsed bound, 1790.787 seconds; full-01: child exit -15, configured elapsed bound, 1791.304 seconds; full-07: child exit -15, configured elapsed bound, 571.096 seconds. Interrupted active units are excluded from completed observation counts.

Entry checker diagnostics: 0. The latent observations still do not establish whole-program lowering or native behavior.

Observed 6798 unit results: {'lowered': 4220, 'stopped': 2578}. First-error pass observed 3447 results; full pieces observed 6283. The summed table uses one result per unit, preferring full mode over first-error mode and later full observations over earlier ones. Earlier completed unit results remain when a later attempt does not reach that unit. Repeated sites in different units are repeated observations, not independent source sites.

Stock unit inventory has 5053 units in 38 emitted files out of 81 resolved files. Units without an observer result, by stock status: `{'deferred_generic_instantiation': 634}`. Missing stock file records after an interrupted pass are not evidence of zero units.

In first-error mode a unit with a lowering finding contributes zero sites: its body is discarded. In full mode the stock hook rolls failed statements back and continues. A `stopped` full-mode observation can therefore contain retained successful statements; its failed statements contribute no sites. Panicked unit results contribute zero sites. `lowered` means no new lowering finding in that unit, not whole-program acceptance. File-level refusal-scan findings do not suppress independent-unit examination.

Every listed expression and metadata entry lacks a source-position field at this base. All positions are explicitly labeled unit fallbacks in the raw stream. Distinct reported locations deduplicate `(kind, unit file:line:column)`, and are not a bound on distinct actual source sites. Several expressions can share a fallback, while dependency bodies repeated in different units receive different owner fallbacks. Exact source-site deduplication cannot be recovered from this IR. The observer does not use the boundary's last-seen graphViewSite or graphDynamicSite.

GraphTypes non-empty in 0 observed results. The latent hook never runs `findCycles`; GraphTypes is not computed per unit. Empty flags here do not establish acyclicity. No graph membership is inferred.

## Per file, lowered IR only

Attribution is to the unit's file, not an independent root. Complete traversal means the census visited the file's candidates; findings, rollback and deferred generic gaps still apply. A zero with zero observed unit results is a coverage gap, not evidence of an empty file. Dependency function bodies or metadata can be included in that unit's result.

| Unit file | Views | Record literals | Dynamic reads | Checked object calls | Observed unit results | Full-mode traversal | Any observed GraphTypes |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 1 | not yet attempted | no, not computed |
| _namespaces/ts.performance.ts | 0 | 0 | 0 | 0 | 1 | not yet attempted | no, not computed |
| _namespaces/ts.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| binder.ts | 0 | 0 | 0 | 0 | 16 | not yet attempted | no, not computed |
| builder.ts | 0 | 0 | 0 | 0 | 124 | complete traversal | no, not computed |
| builderPublic.ts | 0 | 0 | 0 | 0 | 16 | complete traversal | no, not computed |
| builderState.ts | 0 | 0 | 0 | 0 | 33 | complete traversal | no, not computed |
| builderStatePublic.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| checker.ts | 10465 | 0 | 0 | 0 | 1402 | partial | no, not computed |
| commandLineParser.ts | 0 | 0 | 0 | 0 | 159 | not yet attempted | no, not computed |
| core.ts | 0 | 0 | 4 | 0 | 59 | complete traversal | no, not computed |
| corePublic.ts | 0 | 0 | 0 | 0 | 9 | complete traversal | no, not computed |
| debug.ts | 0 | 0 | 0 | 0 | 85 | complete traversal | no, not computed |
| diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 | 3 | not yet attempted | no, not computed |
| emitter.ts | 3433 | 0 | 9 | 0 | 346 | partial | no, not computed |
| executeCommandLine.ts | 160 | 0 | 0 | 0 | 68 | complete traversal | no, not computed |
| expressionToTypeNode.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 | 8 | partial | no, not computed |
| factory/emitHelpers.ts | 0 | 0 | 0 | 0 | 83 | partial | no, not computed |
| factory/emitNode.ts | 0 | 0 | 0 | 0 | 18 | partial | no, not computed |
| factory/nodeChildren.ts | 12 | 0 | 0 | 0 | 6 | partial | no, not computed |
| factory/nodeConverters.ts | 204 | 0 | 0 | 0 | 12 | partial | no, not computed |
| factory/nodeFactory.ts | 621 | 0 | 1 | 0 | 502 | complete traversal | no, not computed |
| factory/nodeTests.ts | 0 | 0 | 0 | 0 | 228 | complete traversal | no, not computed |
| factory/parenthesizerRules.ts | 224 | 0 | 0 | 0 | 42 | complete traversal | no, not computed |
| factory/utilities.ts | 789 | 0 | 0 | 0 | 101 | complete traversal | no, not computed |
| factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| hostErrors.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| moduleNameResolver.ts | 66 | 0 | 1 | 0 | 168 | complete traversal | no, not computed |
| moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 52 | complete traversal | no, not computed |
| parser.ts | 0 | 0 | 0 | 0 | 53 | not yet attempted | no, not computed |
| path.ts | 0 | 0 | 0 | 0 | 77 | not yet attempted | no, not computed |
| performance.ts | 0 | 0 | 0 | 0 | 24 | not yet attempted | no, not computed |
| performanceCore.ts | 0 | 0 | 0 | 0 | 12 | not yet attempted | no, not computed |
| program.ts | 4452 | 0 | 5 | 0 | 235 | complete traversal | no, not computed |
| programDiagnostics.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| resolutionCache.ts | 29 | 0 | 0 | 0 | 75 | complete traversal | no, not computed |
| scanner.ts | 0 | 0 | 0 | 0 | 74 | not yet attempted | no, not computed |
| semver.ts | 0 | 0 | 0 | 0 | 29 | complete traversal | no, not computed |
| sourcemap.ts | 0 | 0 | 0 | 0 | 58 | complete traversal | no, not computed |
| symbolWalker.ts | 0 | 0 | 0 | 0 | 14 | complete traversal | no, not computed |
| sys.ts | 6 | 0 | 0 | 0 | 137 | complete traversal | no, not computed |
| tracing.ts | 0 | 0 | 0 | 0 | 23 | complete traversal | no, not computed |
| transformer.ts | 0 | 0 | 0 | 0 | 15 | complete traversal | no, not computed |
| transformers/classFields.ts | 3176 | 0 | 0 | 0 | 127 | complete traversal | no, not computed |
| transformers/classThis.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| transformers/declarations.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/declarations/diagnostics.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/destructuring.ts | 193 | 0 | 0 | 0 | 22 | not yet attempted | no, not computed |
| transformers/es2015.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/es2016.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/es2017.ts | 0 | 0 | 0 | 0 | 6 | not yet attempted | no, not computed |
| transformers/es2018.ts | 0 | 0 | 0 | 0 | 4 | not yet attempted | no, not computed |
| transformers/es2019.ts | 0 | 0 | 0 | 0 | 2 | not yet attempted | no, not computed |
| transformers/es2020.ts | 0 | 0 | 0 | 0 | 2 | not yet attempted | no, not computed |
| transformers/es2021.ts | 0 | 0 | 0 | 0 | 2 | not yet attempted | no, not computed |
| transformers/esDecorators.ts | 0 | 0 | 0 | 0 | 9 | not yet attempted | no, not computed |
| transformers/esnext.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/generators.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/jsx.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/legacyDecorators.ts | 0 | 0 | 0 | 0 | 2 | not yet attempted | no, not computed |
| transformers/module/esnextAnd2015.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/module/module.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/module/system.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| transformers/namedEvaluation.ts | 0 | 0 | 0 | 0 | 19 | not yet attempted | no, not computed |
| transformers/taggedTemplate.ts | 0 | 0 | 0 | 0 | 5 | not yet attempted | no, not computed |
| transformers/ts.ts | 0 | 0 | 0 | 0 | 5 | not yet attempted | no, not computed |
| transformers/typeSerializer.ts | 0 | 0 | 0 | 0 | 6 | not yet attempted | no, not computed |
| transformers/utilities.ts | 1130 | 0 | 0 | 0 | 48 | complete traversal | no, not computed |
| tsbuild.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| tsbuildPublic.ts | 0 | 0 | 0 | 0 | 34 | complete traversal | no, not computed |
| tsc/_namespaces/ts.ts | 0 | 0 | 0 | 0 | 1 | complete traversal | no, not computed |
| tsc/tsc.ts | 0 | 0 | 0 | 0 | 6 | complete traversal | no, not computed |
| types.ts | 0 | 0 | 0 | 0 | 875 | complete traversal | no, not computed |
| utilities.ts | 10398 | 0 | 8 | 0 | 867 | complete traversal | no, not computed |
| utilitiesPublic.ts | 2528 | 0 | 0 | 0 | 243 | complete traversal | no, not computed |
| visitorPublic.ts | 0 | 0 | 0 | 0 | 28 | complete traversal | no, not computed |
| watch.ts | 0 | 0 | 0 | 0 | 36 | complete traversal | no, not computed |
| watchPublic.ts | 0 | 0 | 0 | 0 | 18 | complete traversal | no, not computed |
| watchUtilities.ts | 0 | 0 | 0 | 0 | 44 | complete traversal | no, not computed |

## Top 10 files by record_literals, lowered IR only

| Unit file | Count |
| --- | ---: |
| no sites observed in covered unit results | 0 |

Examples, when available, use unit fallback positions and adapted unit source text; exact expression locations are unavailable.

## Top 10 files by views, lowered IR only

| Unit file | Count |
| --- | ---: |
| checker.ts | 10465 |
| utilities.ts | 10398 |
| program.ts | 4452 |
| emitter.ts | 3433 |
| transformers/classFields.ts | 3176 |
| utilitiesPublic.ts | 2528 |
| transformers/utilities.ts | 1130 |
| factory/utilities.ts | 789 |
| factory/nodeFactory.ts | 621 |
| factory/parenthesizerRules.ts | 224 |

Examples, when available, use unit fallback positions and adapted unit source text; exact expression locations are unavailable.

`checker.ts:2437:5` (unit fallback):
```ts
    function getJsxNamespace(location: Node | undefined): __String {
        if (location) {
            const file = getSourceFileOfNode(location);
```

`checker.ts:3112:9` (unit fallback):
```ts
        function isUsedInFunctionOrInstancePropertyWorker(usage: Node, declaration: Node): boolean {
            return !!findAncestor(usage, current => {
                if (current === declContainer) {
```

`checker.ts:3299:5` (unit fallback):
```ts
    function onSuccessfullyResolvedSymbol(
        errorLocation: Node | undefined,
        result: Symbol,
```

`utilities.ts:1097:1` (unit fallback):
```ts
export function isGrammarError(parent: Node, child: Node | NodeArray<Node>): boolean {
    if (isTypeParameterDeclaration(parent)) return child === parent.expression;
    if (isClassStaticBlockDeclaration(parent)) return child === parent.modifiers;
```

`utilities.ts:1143:1` (unit fallback):
```ts
function isAnyPrologueDirective(node: Node) {
    return isPrologueDirective(node) || !!(getEmitFlags(node) & EmitFlags.CustomPrologue);
}
```

`utilities.ts:1296:1` (unit fallback):
```ts
export function isExportNamespaceAsDefaultDeclaration(node: Node): boolean {
    return !!(isExportDeclaration(node) && node.exportClause && isNamespaceExport(node.exportClause) && moduleExportNameIsDefault(node.exportClause.name));
}
```

`program.ts:937:1` (unit fallback):
```ts
function getModeForUsageLocationWorker(file: Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">, usage: StringLiteralLike, compilerOptions?: CompilerOptions) {
    if (isImportDeclaration(usage.parent) || isExportDeclaration(usage.parent) || isJSDocImportTag(usage.parent)) {
        const isTypeOnly = isExclusivelyTypeOnlyImportOrExport(usage.parent);
```

`program.ts:959:1` (unit fallback):
```ts
function getEmitSyntaxForUsageLocationWorker(file: Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">, usage: StringLiteralLike, compilerOptions?: CompilerOptions): ResolutionMode {
    if (!compilerOptions) {
        // This should always be provided, but we try to fail somewhat
```

`program.ts:986:1` (unit fallback):
```ts
export function getResolutionModeOverride(node: ImportAttributes | undefined, grammarErrorOnNode?: (node: Node, diagnostic: DiagnosticMessage) => void): ResolutionMode | undefined {
    if (!node) return undefined;
    if (length(node.elements) !== 1) {
```

`emitter.ts:752:1` (unit fallback):
```ts
export function emitFiles(
    resolver: EmitResolver,
    host: EmitHost,
```

`emitter.ts:970:5` (unit fallback):
```ts
    function collectLinkedAliases(node: Node): undefined {
        if (isExportAssignment(node)) {
            if (node.expression.kind === SyntaxKind.Identifier) {
```

`emitter.ts:984:5` (unit fallback):
```ts
    function markLinkedReferences(file: SourceFile) {
        if (ts.isSourceFileJS(file)) return; // JS files don't use reference calculations as they don't do import ellision, no need to calculate it
        ts.forEachChildRecursively(file, n => {
```

`transformers/classFields.ts:3343:1` (unit fallback):
```ts
function isReservedPrivateName(node: PrivateIdentifier) {
    return !isGeneratedPrivateIdentifier(node) && node.escapedText === "#constructor";
}
```

`transformers/classFields.ts:3349:1` (unit fallback):
```ts
function isPrivateIdentifierInExpression(node: BinaryExpression): node is PrivateIdentifierInExpression {
    return isPrivateIdentifier(node.left)
        && node.operatorToken.kind === SyntaxKind.InKeyword;
```

`transformers/classFields.ts:3354:1` (unit fallback):
```ts
function isStaticPropertyDeclaration(node: Node): node is PropertyDeclaration {
    return isPropertyDeclaration(node) && hasStaticModifier(node);
}
```

`utilitiesPublic.ts:615:1` (unit fallback):
```ts
export function isParameterPropertyDeclaration(node: Node, parent: Node): node is ParameterPropertyDeclaration {
    return isParameter(node) && hasSyntacticModifier(node, ModifierFlags.ParameterPropertyModifier) && parent.kind === SyntaxKind.Constructor;
}
```

`utilitiesPublic.ts:629:1` (unit fallback):
```ts
export function isEmptyBindingElement(node: BindingElement | ArrayBindingElement): boolean {
    if (isOmittedExpression(node)) {
        return true;
```

`utilitiesPublic.ts:870:1` (unit fallback):
```ts
export function symbolName(symbol: Symbol): string {
    if (symbol.valueDeclaration && isPrivateIdentifierClassElementDeclaration(symbol.valueDeclaration)) {
        return idText(symbol.valueDeclaration.name);
```

`transformers/utilities.ts:179:1` (unit fallback):
```ts
export function collectExternalModuleInfo(context: TransformationContext, sourceFile: SourceFile): ExternalModuleInfo {
    const resolver = context.getEmitResolver();
    const compilerOptions = context.getCompilerOptions();
```

`transformers/utilities.ts:389:1` (unit fallback):
```ts
export class IdentifierNameMap<V> {
    private readonly _map = new Map<string, V>();

```

`transformers/utilities.ts:441:1` (unit fallback):
```ts
class IdentifierNameMultiMap<V> extends IdentifierNameMap<V[]> {
    add(key: Identifier, value: V): V[] {
        let values = this.get(key);
```

`factory/utilities.ts:579:1` (unit fallback):
```ts
function isUseStrictPrologue(node: ExpressionStatement): boolean {
    return isStringLiteral(node.expression) && node.expression.text === "use strict";
}
```

`factory/utilities.ts:584:1` (unit fallback):
```ts
export function findUseStrictPrologue(statements: readonly Statement[]): Statement | undefined {
    for (const statement of statements) {
        if (isPrologueDirective(statement)) {
```

`factory/utilities.ts:599:1` (unit fallback):
```ts
export function startsWithUseStrict(statements: readonly Statement[]): boolean {
    const firstStatement = firstOrUndefined(statements);
    return firstStatement !== undefined
```

`factory/nodeFactory.ts:2848:5` (unit fallback):
```ts
    function createArrayLiteralExpression(elements?: readonly Expression[], multiLine?: boolean) {
        const node = createBaseNode<ArrayLiteralExpression>(SyntaxKind.ArrayLiteralExpression);
        // Ensure we add a trailing comma for something like `[NumericLiteral(1), NumericLiteral(2), OmittedExpresion]` so that
```

`factory/nodeFactory.ts:2919:5` (unit fallback):
```ts
    function updatePropertyAccessExpression(node: PropertyAccessExpression, expression: Expression, name: Identifier | PrivateIdentifier) {
        if (isPropertyAccessChain(node)) {
            return updatePropertyAccessChain(node, expression, node.questionDotToken, cast(name, isIdentifier));
```

`factory/nodeFactory.ts:2930:5` (unit fallback):
```ts
    function createPropertyAccessChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, name: string | Identifier | PrivateIdentifier) {
        const node = createBasePropertyAccessExpression(
            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),
```

`factory/parenthesizerRules.ts:265:5` (unit fallback):
```ts
    function getLiteralKindOfBinaryPlusOperand(node: Expression): SyntaxKind {
        node = skipPartiallyEmittedExpressions(node);

```

`factory/parenthesizerRules.ts:372:5` (unit fallback):
```ts
    function parenthesizeExpressionOfNew(expression: Expression): LeftHandSideExpression {
        const leftmostExpr = getLeftmostExpression(expression, /*stopAtCallExpressions*/ true);
        switch (leftmostExpr.kind) {
```

`factory/parenthesizerRules.ts:412:5` (unit fallback):
```ts
    function parenthesizeOperandOfPostfixUnary(operand: Expression): LeftHandSideExpression {
        // TODO(rbuckton): Verifiy whether `setTextRange` is needed.
        return isLeftHandSideExpression(operand) ? operand : setTextRange(factory.createParenthesizedExpression(operand), operand);
```

## Pass exits and limits

| Pass | Child exit | Seconds | Stopping reason |
| --- | ---: | ---: | --- |
| apply-add | 0 | 6.017 | completed |
| apply-commit | 0 | 1.006 | completed |
| apply-push | 0 | 3.002 | completed |
| apply-report | 0 | 1.004 | completed |
| apply-snapshot-final | 0 | 1.002 | completed |
| apply-snapshot | 0 | 1.001 | completed |
| apply | 0 | 917.51 | completed |
| census-build | 0 | 473.132 | completed |
| census-final-build | 0 | 1.006 | completed |
| coverage-preflight-audit | 0 | 2.002 | completed |
| coverage-preflight-report | 0 | 1.001 | completed |
| coverage-preflight-snapshot | 0 | 2.006 | completed |
| entry-build-fixed | 0 | 33.1 | completed |
| entry-build | 1 | 10.017 | nonzero exit; cause not inferred |
| entry-fixture | 0 | 1.007 | completed |
| export-audit | 0 | 2.003 | completed |
| export-report | 0 | 1.003 | completed |
| final-audit | 0 | 2.01 | completed |
| final-checksums-corrected | 0 | 1.002 | completed |
| final-checksums | 1 | 1.004 | nonzero exit; cause not inferred |
| final-report | 0 | 2.003 | completed |
| final-snapshot-pre | 0 | 1.001 | completed |
| final-snapshot | 0 | 2.007 | completed |
| first-add | 0 | 1.008 | completed |
| first-commit | 0 | 1.007 | completed |
| first-push | 0 | 3.006 | completed |
| first-report | 0 | 1.009 | completed |
| first-snapshot-final | 0 | 1.004 | completed |
| first-snapshot | 0 | 1.005 | completed |
| first | -15 | 1790.787 | configured elapsed bound |
| full-00-add | 0 | 1.003 | completed |
| full-00-commit | 0 | 1.002 | completed |
| full-00-push | 0 | 3.006 | completed |
| full-00-report | 0 | 1.002 | completed |
| full-00-snapshot-final | 0 | 1.001 | completed |
| full-00-snapshot | 0 | 1.001 | completed |
| full-00 | 0 | 24.088 | completed |
| full-01-add | 0 | 1.021 | completed |
| full-01-commit | 0 | 1.009 | completed |
| full-01-push | 0 | 3.015 | completed |
| full-01-report | 0 | 1.01 | completed |
| full-01-snapshot-final | 0 | 1.005 | completed |
| full-01-snapshot | 0 | 1.007 | completed |
| full-01 | -15 | 1791.304 | configured elapsed bound |
| full-02-add | 0 | 1.003 | completed |
| full-02-commit | 0 | 1.002 | completed |
| full-02-push | 0 | 3.006 | completed |
| full-02-report | 0 | 1.005 | completed |
| full-02-snapshot-final | 0 | 1.002 | completed |
| full-02-snapshot | 0 | 1.005 | completed |
| full-02 | 0 | 79.359 | completed |
| full-03-add | 0 | 1.004 | completed |
| full-03-commit | 0 | 1.002 | completed |
| full-03-push | 0 | 2.002 | completed |
| full-03-report | 0 | 1.001 | completed |
| full-03-snapshot-final | 0 | 1.002 | completed |
| full-03-snapshot | 0 | 1.003 | completed |
| full-03 | 0 | 119.289 | completed |
| full-04-add | 0 | 1.003 | completed |
| full-04-commit | 0 | 1.003 | completed |
| full-04-push | 0 | 3.011 | completed |
| full-04-report | 0 | 1.007 | completed |
| full-04-snapshot-final | 0 | 1.002 | completed |
| full-04-snapshot | 0 | 1.002 | completed |
| full-04 | 0 | 177.329 | completed |
| full-05-add | 0 | 1.007 | completed |
| full-05-commit | 0 | 1.005 | completed |
| full-05-push | 0 | 3.004 | completed |
| full-05-report | 0 | 1.006 | completed |
| full-05-snapshot-final | 0 | 1.003 | completed |
| full-05-snapshot | 0 | 1.003 | completed |
| full-05 | 0 | 328.32 | completed |
| full-06-add | 0 | 1.009 | completed |
| full-06-commit | 0 | 1.002 | completed |
| full-06-push | 0 | 3.005 | completed |
| full-06-report | 0 | 1.002 | completed |
| full-06-snapshot-final | 0 | 1.001 | completed |
| full-06-snapshot | 0 | 1.003 | completed |
| full-06 | 0 | 454.859 | completed |
| full-07-add | 0 | 1.003 | completed |
| full-07-commit | 0 | 1.002 | completed |
| full-07-push | 0 | 3.004 | completed |
| full-07-report | 0 | 1.005 | completed |
| full-07-snapshot-final | 0 | 1.002 | completed |
| full-07-snapshot | 0 | 1.003 | completed |
| full-07 | -15 | 571.096 | configured elapsed bound |
| full-20-000-add | 0 | 1.006 | completed |
| full-20-000-audit | 0 | 2.007 | completed |
| full-20-000-commit | 0 | 1.004 | completed |
| full-20-000-push | 0 | 3.007 | completed |
| full-20-000-report | 0 | 1.002 | completed |
| full-20-000-snapshot-final | 0 | 1.002 | completed |
| full-20-000-snapshot | 0 | 1.003 | completed |
| full-20-000 | 0 | 580.765 | completed |
| full-20-001-add | 0 | 1.005 | completed |
| full-20-001-audit | 0 | 1.003 | completed |
| full-20-001-commit | 0 | 1.003 | completed |
| full-20-001-push | 0 | 3.006 | completed |
| full-20-001-report | 0 | 1.021 | completed |
| full-20-001-snapshot-final | 0 | 1.001 | completed |
| full-20-001-snapshot | 0 | 1.002 | completed |
| full-20-001 | 0 | 75.047 | completed |
| full-20-002-add | 0 | 1.002 | completed |
| full-20-002-audit | 0 | 1.002 | completed |
| full-20-002-commit | 0 | 1.005 | completed |
| full-20-002-push | 0 | 3.006 | completed |
| full-20-002-report | 0 | 1.003 | completed |
| full-20-002-snapshot-final | 0 | 1.004 | completed |
| full-20-002-snapshot | 0 | 1.001 | completed |
| full-20-002 | 0 | 18.023 | completed |
| full-20-003-add | 0 | 1.003 | completed |
| full-20-003-audit | 0 | 1.001 | completed |
| full-20-003-commit | 0 | 1.002 | completed |
| full-20-003-push | 0 | 3.005 | completed |
| full-20-003-report | 0 | 1.001 | completed |
| full-20-003-snapshot-final | 0 | 1.023 | completed |
| full-20-003-snapshot | 0 | 1.002 | completed |
| full-20-003 | 0 | 42.029 | completed |
| full-20-004-add | 0 | 1.003 | completed |
| full-20-004-audit | 0 | 1.001 | completed |
| full-20-004-commit | 0 | 1.024 | completed |
| full-20-004-push | 0 | 3.006 | completed |
| full-20-004-report | 0 | 1.001 | completed |
| full-20-004-snapshot-final | 0 | 1.004 | completed |
| full-20-004-snapshot | 0 | 1.001 | completed |
| full-20-004 | 0 | 17.014 | completed |
| full-20-005-add | 0 | 1.004 | completed |
| full-20-005-audit | 0 | 2.004 | completed |
| full-20-005-commit | 0 | 1.004 | completed |
| full-20-005-push | 0 | 2.008 | completed |
| full-20-005-report | 0 | 1.004 | completed |
| full-20-005-snapshot-final | 0 | 1.002 | completed |
| full-20-005-snapshot | 0 | 1.001 | completed |
| full-20-005 | 0 | 41.043 | completed |
| full-20-006-add | 0 | 1.003 | completed |
| full-20-006-audit | 0 | 1.003 | completed |
| full-20-006-commit | 0 | 1.003 | completed |
| full-20-006-push | 0 | 3.004 | completed |
| full-20-006-report | 0 | 1.005 | completed |
| full-20-006-snapshot-final | 0 | 1.004 | completed |
| full-20-006-snapshot | 0 | 1.002 | completed |
| full-20-006 | 0 | 71.166 | completed |
| full-20-007-add | 0 | 1.005 | completed |
| full-20-007-audit | 0 | 2.029 | completed |
| full-20-007-commit | 0 | 1.005 | completed |
| full-20-007-push | 0 | 3.009 | completed |
| full-20-007-report | 0 | 1.002 | completed |
| full-20-007-snapshot-final | 0 | 1.002 | completed |
| full-20-007-snapshot | 0 | 1.002 | completed |
| full-20-007 | 0 | 86.22 | completed |
| full-20-008-add | 0 | 1.005 | completed |
| full-20-008-audit | 0 | 2.011 | completed |
| full-20-008-commit | 0 | 1.003 | completed |
| full-20-008-push | 0 | 3.011 | completed |
| full-20-008-report | 0 | 2.01 | completed |
| full-20-008-snapshot-final | 0 | 1.001 | completed |
| full-20-008-snapshot | 0 | 2.006 | completed |
| full-20-008 | 0 | 524.788 | completed |
| full-20-009-add | 0 | 1.004 | completed |
| full-20-009-audit | 0 | 2.005 | completed |
| full-20-009-commit | 0 | 1.004 | completed |
| full-20-009-push | 0 | 3.004 | completed |
| full-20-009-report | 0 | 2.004 | completed |
| full-20-009-snapshot-final | 0 | 2.008 | completed |
| full-20-009-snapshot | 0 | 2.002 | completed |
| full-20-009 | 0 | 68.568 | completed |
| full-20-010-add | 0 | 1.004 | completed |
| full-20-010-audit | 0 | 2.003 | completed |
| full-20-010-commit | 0 | 1.006 | completed |
| full-20-010-push | 0 | 3.012 | completed |
| full-20-010-report | 0 | 2.007 | completed |
| full-20-010-snapshot-final | 0 | 2.004 | completed |
| full-20-010-snapshot | 0 | 2.003 | completed |
| full-20-010 | 0 | 70.571 | completed |
| full-20-011-add | 0 | 1.006 | completed |
| full-20-011-audit | 0 | 3.004 | completed |
| full-20-011-commit | 0 | 1.001 | completed |
| full-20-011-push | 0 | 2.001 | completed |
| full-20-011-report | 0 | 2.005 | completed |
| full-20-011-snapshot-final | 0 | 1.003 | completed |
| full-20-011-snapshot | 0 | 2.006 | completed |
| full-20-011 | 0 | 54.11 | completed |
| full-20-012-add | 0 | 1.002 | completed |
| full-20-012-audit | 0 | 2.003 | completed |
| full-20-012-commit | 0 | 1.001 | completed |
| full-20-012-push | 0 | 3.007 | completed |
| full-20-012-report | 0 | 2.002 | completed |
| full-20-012-snapshot-final | 0 | 1.004 | completed |
| full-20-012-snapshot | 0 | 1.003 | completed |
| full-20-012 | 0 | 117.241 | completed |
| full-20-013-add | 0 | 1.002 | completed |
| full-20-013-audit | 0 | 2.009 | completed |
| full-20-013-commit | 0 | 1.006 | completed |
| full-20-013-push | 0 | 3.005 | completed |
| full-20-013-report | 0 | 2.002 | completed |
| full-20-013-snapshot-final | 0 | 1.003 | completed |
| full-20-013-snapshot | 0 | 2.003 | completed |
| full-20-013 | 0 | 42.068 | completed |
| full-20-014-add | 0 | 1.006 | completed |
| full-20-014-audit | 0 | 2.002 | completed |
| full-20-014-commit | 0 | 1.005 | completed |
| full-20-014-push | 0 | 3.006 | completed |
| full-20-014-report | 0 | 2.007 | completed |
| full-20-014-snapshot-final | 0 | 2.004 | completed |
| full-20-014-snapshot | 0 | 1.001 | completed |
| full-20-014 | 0 | 38.079 | completed |
| full-20-015 | 0 | 197.235 | completed |
| full-20-016-add | 0 | 1.003 | completed |
| full-20-016-audit | 0 | 3.024 | completed |
| full-20-016-commit | 0 | 1.002 | completed |
| full-20-016-push | 0 | 3.006 | completed |
| full-20-016-report | 0 | 2.006 | completed |
| full-20-016-snapshot-final | 0 | 1.001 | completed |
| full-20-016-snapshot | 0 | 2.007 | completed |
| full-20-016 | 0 | 31.135 | completed |
| instrument-entry-fixed | 0 | 1.001 | completed |
| instrument-entry | 0 | 1.002 | completed |
| interim-audit | 0 | 3.002 | completed |
| interim-report | 0 | 1.002 | completed |
| npm-ci | 0 | not timed | completed |
| overlay-entry-fixed | 0 | 2.011 | completed |
| overlay-entry | 0 | 3.006 | completed |
| overlay | 0 | 6.029 | completed |
| publish-add | 0 | 1.001 | completed |
| publish-branch | 0 | 1.001 | completed |
| publish-checkpoint-commit | 0 | 1.001 | completed |
| publish-checkpoint-push | 0 | 3.007 | completed |
| publish-commit | 0 | 1.002 | completed |
| publish-count | 0 | 1.002 | completed |
| publish-final-scope | 0 | 1.002 | completed |
| publish-parent | 0 | 1.003 | completed |
| publish-push | 0 | 4.005 | completed |
| publish-remote | 0 | 1.002 | completed |
| publish-scope | 0 | 1.001 | completed |
| publish-staged-scope | 0 | 1.001 | completed |
| publish-switch | 0 | 1.002 | completed |
| publish-tree | 0 | 1.003 | completed |
| selftest | 0 | 1.028 | completed |
| setup | 0 | 758.331 | completed |
| source | 0 | 12.004 | completed |
| verify-apply | 0 | 1.003 | completed |

## Method and reproduction

Read `graph_types.go:460` and its callers at `graph_types.go:355` and `program_region.go:140`. Count ViewOrigins entries plus CheckedFields keys without deduplicating their overlap; walk Main and every function body with the boundary's own walk for record ObjectLiteral, both dynamic-read nodes, and checked ObjectCall. These are hypothetical boundary candidates before the rest of ordinary lowering, not demonstrated graph-program refusals.

The entry resolves 81 implementation files: the 79 compiler files from the independent-root report plus `src/tsc/tsc.ts` and `src/tsc/_namespaces/ts.ts`. The sums cover this resolved entry program. `instrument.py` uses the meter entry overlay to keep the entry first and expand Files to its implementation dependencies. It removes only the meter driver's full-mode-only checker-rejection gate so the requested first-error latent pass can retain diagnostics and proceed. `instrument.py` extends the existing guarded latent overlay in scratch only. It retains successful top-level statement returns in the measurement result's Main, which the stock hook discards. Ordinary Load and Lower remain disabled and every real pass asserts those guards. `overlay/` contains all generated Go inputs as `.go.txt`; no production lower source is changed. The deterministic self-test exercises Main, a function body, both dynamic reads, both view sources, GraphTypes presence, and failed-unit exclusion.

An initial entry build omitted the observer mapping and failed; the corrected build and entry fixture passed before the real census. Both attempts are retained. The flattened selected-site stream is [`sites.jsonl.gz`](sites.jsonl.gz), beside this report; each line includes kind, reported position, unit, owner file, GraphTypes flag and original stream. It is derived without changing the preserved raw pass streams. Raw boundary streams, original latent streams, timing streams, logs and each child's own exit status are compressed separately beside this report in `raw/`. `source-manifest.json`, `provenance.json` and `SHA256SUMS` retain source and binary identities. Piece selections and measurement environment variables are retained in raw JSON. The initial full pieces used Go's `GOMEMLIMIT=10GiB` soft memory limit and a 60-minute aggregate budget, with at most 1790 seconds per child. Continuation pieces select one file each, use `GOMEMLIMIT=4GiB` and at most 1490 seconds per child. Priority and files larger than 300000 bytes run alone; at most two smaller files run concurrently. Only completed child streams enter a checkpoint. The continuation has no aggregate 60-minute cutoff. Each process group is killed and reaped by the supervisor on completion or stopping.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/graph-boundary-adapted > apply.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
python3 stage3/census/graph-boundary-bcbe5d12/instrument.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
go build -overlay /workspace/scratch/jawqwv7/overlay/overlay.json -o /workspace/scratch/jawqwv7/census ./stage3/census/latent/tool
# run-pieces.py records the exact entry invocation and file selections
```
