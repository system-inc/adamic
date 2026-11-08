# Graph adoption boundary census on bcbe5d12

Task jawqwv7. Report only. Counts cover lowered unit results only, including retained successful statements after full-mode rollback. They do not measure accepted whole programs.

Compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`. Reference report tree: `30057eca`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `755daee0c6ba`; cgroup CPU quota `400000 100000`; memory limit `17179869184` bytes; 5 visible CPUs. Environment: `ccarenv_b64_Y2NhcmVudl85M2MyY2I4NzA3ZTA4MTkxODMwYTI0NDZlY2ZhNjAyZg`.

## Four-way split, lowered IR only

| Kind | Observations | Distinct reported locations |
| --- | ---: | ---: |
| views | 19443 | 251 |
| record_literals | 0 | 0 |
| dynamic_reads | 12 | 6 |
| checked_object_calls | 0 | 0 |

Views combine 634 ViewOrigins entries and 18809 CheckedFields keys, as requested; these are metadata observations, not that many distinct adapter expressions.

The 79 compiler-file subtotal is `{'views': 19443, 'record_literals': 0, 'dynamic_reads': 12, 'checked_object_calls': 0}`. The two entry-file subtotal is `{'views': 0, 'record_literals': 0, 'dynamic_reads': 0, 'checked_object_calls': 0}`. Both cover lowered unit results only.

Long pole by observed count: **views, 19443**. Coverage caveat: lowered IR only; failed statements, checker-skipped bodies, deferred generics and uncompleted units are absent. Inference: this ranking may differ after those gaps lower.

## Coverage and interpretation

Bounded-pass stops: first: child exit -15, configured elapsed bound, 1790.787 seconds; full-01: child exit -15, configured elapsed bound, 1791.304 seconds. Interrupted active units are excluded from completed observation counts.

Entry checker diagnostics: 0. The latent observations still do not establish whole-program lowering or native behavior.

Observed 4923 unit results: {'lowered': 3384, 'stopped': 1539}. First-error pass observed 3447 results; full pieces observed 1634. The summed table uses one result per unit, preferring full mode over first-error mode. Repeated sites in different units are repeated observations, not independent source sites.

Stock unit inventory has 559 units in 7 emitted files out of 81 resolved files. Units without an observer result, by stock status: `{'deferred_generic_instantiation': 316}`. Missing stock file records after an interrupted pass are not evidence of zero units.

In first-error mode a unit with a lowering finding contributes zero sites: its body is discarded. In full mode the stock hook rolls failed statements back and continues. A `stopped` full-mode observation can therefore contain retained successful statements; its failed statements contribute no sites. Panicked unit results contribute zero sites. `lowered` means no new lowering finding in that unit, not whole-program acceptance. File-level refusal-scan findings do not suppress independent-unit examination.

Every listed expression and metadata entry lacks a source-position field at this base. All positions are explicitly labeled unit fallbacks in the raw stream. Distinct reported locations deduplicate `(kind, unit file:line:column)`, and are not a bound on distinct actual source sites. Several expressions can share a fallback, while dependency bodies repeated in different units receive different owner fallbacks. Exact source-site deduplication cannot be recovered from this IR. The observer does not use the boundary's last-seen graphViewSite or graphDynamicSite.

GraphTypes non-empty in 0 observed results. The latent hook never runs `findCycles`; GraphTypes is not computed per unit. Empty flags here do not establish acyclicity. No graph membership is inferred.

## Per file, lowered IR only

Attribution is to the unit's file, not an independent root. A zero with zero observed unit results is a coverage gap, not evidence of an empty file. Dependency function bodies or metadata can be included in that unit's result.

| Unit file | Views | Record literals | Dynamic reads | Checked object calls | Observed unit results | Any observed GraphTypes |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 1 | no, not computed |
| _namespaces/ts.performance.ts | 0 | 0 | 0 | 0 | 1 | no, not computed |
| _namespaces/ts.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| binder.ts | 0 | 0 | 0 | 0 | 16 | no, not computed |
| builder.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| builderPublic.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| builderState.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| builderStatePublic.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| checker.ts | 10465 | 0 | 0 | 0 | 1402 | no, not computed |
| commandLineParser.ts | 0 | 0 | 0 | 0 | 159 | no, not computed |
| core.ts | 0 | 0 | 4 | 0 | 59 | no, not computed |
| corePublic.ts | 0 | 0 | 0 | 0 | 9 | no, not computed |
| debug.ts | 0 | 0 | 0 | 0 | 85 | no, not computed |
| diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| emitter.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| executeCommandLine.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| expressionToTypeNode.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| factory/emitHelpers.ts | 0 | 0 | 0 | 0 | 45 | no, not computed |
| factory/emitNode.ts | 0 | 0 | 0 | 0 | 18 | no, not computed |
| factory/nodeChildren.ts | 0 | 0 | 0 | 0 | 6 | no, not computed |
| factory/nodeConverters.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| factory/nodeFactory.ts | 0 | 0 | 1 | 0 | 25 | no, not computed |
| factory/nodeTests.ts | 0 | 0 | 0 | 0 | 228 | no, not computed |
| factory/parenthesizerRules.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| factory/utilities.ts | 637 | 0 | 0 | 0 | 100 | no, not computed |
| factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| hostErrors.ts | 0 | 0 | 0 | 0 | 3 | no, not computed |
| moduleNameResolver.ts | 0 | 0 | 0 | 0 | 148 | no, not computed |
| moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 48 | no, not computed |
| parser.ts | 0 | 0 | 0 | 0 | 53 | no, not computed |
| path.ts | 0 | 0 | 0 | 0 | 77 | no, not computed |
| performance.ts | 0 | 0 | 0 | 0 | 24 | no, not computed |
| performanceCore.ts | 0 | 0 | 0 | 0 | 12 | no, not computed |
| program.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| programDiagnostics.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| resolutionCache.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| scanner.ts | 0 | 0 | 0 | 0 | 74 | no, not computed |
| semver.ts | 0 | 0 | 0 | 0 | 29 | no, not computed |
| sourcemap.ts | 0 | 0 | 0 | 0 | 30 | no, not computed |
| symbolWalker.ts | 0 | 0 | 0 | 0 | 2 | no, not computed |
| sys.ts | 0 | 0 | 0 | 0 | 60 | no, not computed |
| tracing.ts | 0 | 0 | 0 | 0 | 10 | no, not computed |
| transformer.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/classFields.ts | 192 | 0 | 0 | 0 | 23 | no, not computed |
| transformers/classThis.ts | 0 | 0 | 0 | 0 | 5 | no, not computed |
| transformers/declarations.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/declarations/diagnostics.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/destructuring.ts | 193 | 0 | 0 | 0 | 22 | no, not computed |
| transformers/es2015.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/es2016.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/es2017.ts | 0 | 0 | 0 | 0 | 6 | no, not computed |
| transformers/es2018.ts | 0 | 0 | 0 | 0 | 4 | no, not computed |
| transformers/es2019.ts | 0 | 0 | 0 | 0 | 2 | no, not computed |
| transformers/es2020.ts | 0 | 0 | 0 | 0 | 2 | no, not computed |
| transformers/es2021.ts | 0 | 0 | 0 | 0 | 2 | no, not computed |
| transformers/esDecorators.ts | 0 | 0 | 0 | 0 | 9 | no, not computed |
| transformers/esnext.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/generators.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/jsx.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/legacyDecorators.ts | 0 | 0 | 0 | 0 | 2 | no, not computed |
| transformers/module/esnextAnd2015.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/module/module.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/module/system.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| transformers/namedEvaluation.ts | 0 | 0 | 0 | 0 | 19 | no, not computed |
| transformers/taggedTemplate.ts | 0 | 0 | 0 | 0 | 5 | no, not computed |
| transformers/ts.ts | 0 | 0 | 0 | 0 | 5 | no, not computed |
| transformers/typeSerializer.ts | 0 | 0 | 0 | 0 | 6 | no, not computed |
| transformers/utilities.ts | 1130 | 0 | 0 | 0 | 48 | no, not computed |
| tsbuild.ts | 0 | 0 | 0 | 0 | 5 | no, not computed |
| tsbuildPublic.ts | 0 | 0 | 0 | 0 | 34 | no, not computed |
| tsc/_namespaces/ts.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| tsc/tsc.ts | 0 | 0 | 0 | 0 | 6 | no, not computed |
| types.ts | 0 | 0 | 0 | 0 | 875 | no, not computed |
| utilities.ts | 4668 | 0 | 7 | 0 | 834 | no, not computed |
| utilitiesPublic.ts | 2158 | 0 | 0 | 0 | 242 | no, not computed |
| visitorPublic.ts | 0 | 0 | 0 | 0 | 28 | no, not computed |
| watch.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| watchPublic.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |
| watchUtilities.ts | 0 | 0 | 0 | 0 | 0 | no, not computed |

## Top 10 files by record_literals, lowered IR only

| Unit file | Count |
| --- | ---: |
| no sites observed in covered unit results | 0 |

Examples below are unit fallback positions and adapted unit source text, not exact expression locations.

## Top 10 files by views, lowered IR only

| Unit file | Count |
| --- | ---: |
| checker.ts | 10465 |
| utilities.ts | 4668 |
| utilitiesPublic.ts | 2158 |
| transformers/utilities.ts | 1130 |
| factory/utilities.ts | 637 |
| transformers/destructuring.ts | 193 |
| transformers/classFields.ts | 192 |

Examples below are unit fallback positions and adapted unit source text, not exact expression locations.

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

`utilities.ts:2037:1` (unit fallback):
```ts
function canUseOriginalText(node: LiteralLikeNode, flags: GetLiteralTextFlags): boolean {
    if (nodeIsSynthesized(node) || !node.parent || (flags & GetLiteralTextFlags.TerminateUnterminatedLiterals && node.isUnterminated)) {
        return false;
```

`utilitiesPublic.ts:615:1` (unit fallback):
```ts
export function isParameterPropertyDeclaration(node: Node, parent: Node): node is ParameterPropertyDeclaration {
    return isParameter(node) && hasSyntacticModifier(node, ModifierFlags.ParameterPropertyModifier) && parent.kind === SyntaxKind.Constructor;
}
```

`utilitiesPublic.ts:870:1` (unit fallback):
```ts
export function symbolName(symbol: Symbol): string {
    if (symbol.valueDeclaration && isPrivateIdentifierClassElementDeclaration(symbol.valueDeclaration)) {
        return idText(symbol.valueDeclaration.name);
```

`utilitiesPublic.ts:946:1` (unit fallback):
```ts
export function isNamedDeclaration(node: Node): node is NamedDeclaration & { name: DeclarationName; } {
    return !!(node as NamedDeclaration).name; // A 'name' property should always be a DeclarationName.
}
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

`transformers/destructuring.ts:218:1` (unit fallback):
```ts
function bindingOrAssignmentElementContainsNonLiteralComputedName(element: BindingOrAssignmentElement): boolean {
    const propertyName = tryGetPropertyNameOfBindingOrAssignmentElement(element);
    if (propertyName && isComputedPropertyName(propertyName) && !isLiteralExpression(propertyName.expression)) {
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
| entry-build-fixed | 0 | 33.1 | completed |
| entry-build | 1 | 10.017 | nonzero exit; cause not inferred |
| entry-fixture | 0 | 1.007 | completed |
| export-audit | 0 | 2.003 | completed |
| export-report | 0 | 1.003 | completed |
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
| full-03 | 0 | 119.289 | completed |
| instrument-entry-fixed | 0 | 1.001 | completed |
| instrument-entry | 0 | 1.002 | completed |
| interim-audit | 0 | 3.002 | completed |
| interim-report | 0 | 1.002 | completed |
| npm-ci | 0 | not timed | completed |
| overlay-entry-fixed | 0 | 2.011 | completed |
| overlay-entry | 0 | 3.006 | completed |
| overlay | 0 | 6.029 | completed |
| selftest | 0 | 1.028 | completed |
| setup | 0 | 758.331 | completed |
| source | 0 | 12.004 | completed |
| verify-apply | 0 | 1.003 | completed |

## Method and reproduction

Read `graph_types.go:460` and its callers at `graph_types.go:355` and `program_region.go:140`. Count ViewOrigins entries plus CheckedFields keys without deduplicating their overlap; walk Main and every function body with the boundary's own walk for record ObjectLiteral, both dynamic-read nodes, and checked ObjectCall. These are hypothetical boundary candidates before the rest of ordinary lowering, not demonstrated graph-program refusals.

The entry resolves 81 implementation files: the 79 compiler files from the independent-root report plus `src/tsc/tsc.ts` and `src/tsc/_namespaces/ts.ts`. The sums cover this resolved entry program. `instrument.py` uses the meter entry overlay to keep the entry first and expand Files to its implementation dependencies. It removes only the meter driver's full-mode-only checker-rejection gate so the requested first-error latent pass can retain diagnostics and proceed. `instrument.py` extends the existing guarded latent overlay in scratch only. It retains successful top-level statement returns in the measurement result's Main, which the stock hook discards. Ordinary Load and Lower remain disabled and every real pass asserts those guards. `overlay/` contains all generated Go inputs as `.go.txt`; no production lower source is changed. The deterministic self-test exercises Main, a function body, both dynamic reads, both view sources, GraphTypes presence, and failed-unit exclusion.

An initial entry build omitted the observer mapping and failed; the corrected build and entry fixture passed before the real census. Both attempts are retained. The flattened selected-site stream is [`sites.jsonl.gz`](sites.jsonl.gz), beside this report; each line includes kind, reported position, unit, owner file, GraphTypes flag and original stream. It is derived without changing the preserved raw pass streams. Raw boundary streams, original latent streams, timing streams, logs and each child's own exit status are compressed separately beside this report in `raw/`. `source-manifest.json`, `provenance.json` and `SHA256SUMS` retain source and binary identities. Piece selections and measurement environment variables are retained in raw JSON. Full pieces use Go's `GOMEMLIMIT=10GiB` soft memory limit and a 60-minute aggregate budget; each child has at most 1790 seconds. Each process group is killed and reaped by the supervisor on completion or stopping.

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
