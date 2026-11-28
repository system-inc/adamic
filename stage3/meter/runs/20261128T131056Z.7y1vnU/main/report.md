Whole program: 80/80
Own file: 80/80
Module format: CommonJSBundle

Stage 3 meter 20261128T131056Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | pass | pass | fail |
| src/compiler/_namespaces/ts.performance.ts | pass | pass | fail |
| src/compiler/_namespaces/ts.ts | pass | pass | fail |
| src/compiler/binder.ts | pass | pass | fail |
| src/compiler/builder.ts | pass | pass | fail |
| src/compiler/builderPublic.ts | pass | pass | fail |
| src/compiler/builderState.ts | pass | pass | fail |
| src/compiler/builderStatePublic.ts | pass | pass | fail |
| src/compiler/checker.ts | pass | pass | fail |
| src/compiler/commandLineParser.ts | pass | pass | fail |
| src/compiler/core.ts | pass | pass | fail |
| src/compiler/corePublic.ts | pass | pass | pass |
| src/compiler/debug.ts | pass | pass | fail |
| src/compiler/diagnosticInformationMap.generated.ts | pass | pass | fail |
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked |
| src/compiler/diagnosticMessages.json | fail | fail | blocked |
| src/compiler/emitter.ts | pass | pass | fail |
| src/compiler/executeCommandLine.ts | pass | pass | fail |
| src/compiler/expressionToTypeNode.ts | pass | pass | fail |
| src/compiler/factory/baseNodeFactory.ts | pass | pass | fail |
| src/compiler/factory/emitHelpers.ts | pass | pass | fail |
| src/compiler/factory/emitNode.ts | pass | pass | fail |
| src/compiler/factory/nodeChildren.ts | pass | pass | fail |
| src/compiler/factory/nodeConverters.ts | pass | pass | fail |
| src/compiler/factory/nodeFactory.ts | pass | pass | fail |
| src/compiler/factory/nodeTests.ts | pass | pass | fail |
| src/compiler/factory/parenthesizerRules.ts | pass | pass | fail |
| src/compiler/factory/utilities.ts | pass | pass | fail |
| src/compiler/factory/utilitiesPublic.ts | pass | pass | fail |
| src/compiler/hostErrors.ts | pass | pass | pass |
| src/compiler/jsDocCache.ts | pass | pass | pass |
| src/compiler/moduleNameResolver.ts | pass | pass | fail |
| src/compiler/moduleSpecifiers.ts | pass | pass | fail |
| src/compiler/parser.ts | pass | pass | fail |
| src/compiler/path.ts | pass | pass | fail |
| src/compiler/performance.ts | pass | pass | fail |
| src/compiler/performanceCore.ts | pass | pass | fail |
| src/compiler/program.ts | pass | pass | fail |
| src/compiler/programDiagnostics.ts | pass | pass | fail |
| src/compiler/resolutionCache.ts | pass | pass | fail |
| src/compiler/scanner.ts | pass | pass | fail |
| src/compiler/semver.ts | pass | pass | fail |
| src/compiler/sourcemap.ts | pass | pass | fail |
| src/compiler/symbolWalker.ts | pass | pass | fail |
| src/compiler/sys.ts | pass | pass | fail |
| src/compiler/tracing.ts | pass | pass | fail |
| src/compiler/transformer.ts | pass | pass | fail |
| src/compiler/transformers/classFields.ts | pass | pass | fail |
| src/compiler/transformers/classThis.ts | pass | pass | fail |
| src/compiler/transformers/declarations.ts | pass | pass | fail |
| src/compiler/transformers/declarations/diagnostics.ts | pass | pass | fail |
| src/compiler/transformers/destructuring.ts | pass | pass | fail |
| src/compiler/transformers/es2015.ts | pass | pass | fail |
| src/compiler/transformers/es2016.ts | pass | pass | fail |
| src/compiler/transformers/es2017.ts | pass | pass | fail |
| src/compiler/transformers/es2018.ts | pass | pass | fail |
| src/compiler/transformers/es2019.ts | pass | pass | fail |
| src/compiler/transformers/es2020.ts | pass | pass | fail |
| src/compiler/transformers/es2021.ts | pass | pass | fail |
| src/compiler/transformers/esDecorators.ts | pass | pass | fail |
| src/compiler/transformers/esnext.ts | pass | pass | fail |
| src/compiler/transformers/generators.ts | pass | pass | fail |
| src/compiler/transformers/jsx.ts | pass | pass | fail |
| src/compiler/transformers/legacyDecorators.ts | pass | pass | fail |
| src/compiler/transformers/module/esnextAnd2015.ts | pass | pass | fail |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | pass | pass | fail |
| src/compiler/transformers/module/module.ts | pass | pass | fail |
| src/compiler/transformers/module/system.ts | pass | pass | fail |
| src/compiler/transformers/namedEvaluation.ts | pass | pass | fail |
| src/compiler/transformers/taggedTemplate.ts | pass | pass | fail |
| src/compiler/transformers/ts.ts | pass | pass | fail |
| src/compiler/transformers/typeSerializer.ts | pass | pass | fail |
| src/compiler/transformers/utilities.ts | pass | pass | fail |
| src/compiler/tsbuild.ts | pass | pass | fail |
| src/compiler/tsbuildPublic.ts | pass | pass | fail |
| src/compiler/tsconfig.json | fail | fail | blocked |
| src/compiler/types.ts | pass | pass | fail |
| src/compiler/utilities.ts | pass | pass | fail |
| src/compiler/utilitiesPublic.ts | pass | pass | fail |
| src/compiler/visitorPublic.ts | pass | pass | fail |
| src/compiler/watch.ts | pass | pass | fail |
| src/compiler/watchPublic.ts | pass | pass | fail |
| src/compiler/watchUtilities.ts | pass | pass | fail |

Checker: 80/80 source files. Lowering: 3/80 source files.
All files: 83; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.


Lowering, file roots: 3/80; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |
Relation budget: ok; assignable 536,331 of 1,000,000 pairs; assignable cache 536,331 pairs, within 1,000,000.
tsc entry body lowering, main: measured on a checker-clean entry-root program.
Module format: CommonJSBundle.
3722 units lower without a body stop; 0 have their whole-program passes measured
Scoped tail: 9 units stopped (checkRecordGuardedReads: 9).
Resolved source files: 82; NotYet: 611; Refused: 589.
Errors: 0; panics: 0; skipped dependencies: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 25 | 25 |
| a value of type SyntacticTypeNodeBuilderContext seen as NodeBuilderContext, whose readonly field tracker becomes writable: a readonly field may hold something narrower than Required&lt;Pick&lt;SymbolTracker, &quot;reportInferenceFallback&quot;&gt;&gt;, which a write of SymbolTrackerImpl would replace | #k8ed1f2 (unchecked) | 0 | 18 | 18 |
| a cast the runtime can&#x27;t check | #b5w3ycg (unchecked) | 0 | 14 | 14 |
| Object enumeration of a nullable spread with synthetic absent slots | OWNER BLANK | 11 | 0 | 11 |
| a checked view of FlowLabel over the union FlowNode, whose members no literal tag tells apart | OWNER BLANK | 10 | 0 | 10 |
| overload 1 of visitNode result TOut &#124; (TIn &amp; undefined) &#124; (TVisited &amp; undefined) cannot be served by implementation result Node &#124; undefined (return at /home/ahra/scratch/system_adamic/01M8H4EZVDQWGXFSAG50BRVB3J.meter/tmp/stage3-meter.giGgkC/main-adapted/src/compiler/visitorPublic.ts:156:9 isn&#x27;t the value its predicate checked) | OWNER BLANK | 0 | 10 | 10 |
| a cast that makes a viewed readonly array writable | #b5w3ycg (unchecked) | 9 | 0 | 9 |
| a cast the runtime can&#x27;t check: a cast over TypeSystemEntity, whose writable field flags some member can&#x27;t hold every value of, so a write through the cast could break that member | #b5w3ycg (unchecked) | 0 | 9 | 9 |
| a checked view of FlowCall over the union FlowNode, whose members no literal tag tells apart | OWNER BLANK | 8 | 0 | 8 |
| a checked view of VariableDeclarationList over the union ForInitializer, whose members no literal tag tells apart | OWNER BLANK | 8 | 0 | 8 |

Counts cover the entry root and its resolved implementation dependencies, excluding declarations.
Errors and panics are retained separately in JSON. This census does not establish native output.

Latent body lowering, main: measured on a checker-rejected program. Module format: CommonJSBundle.
3715 units lower without a body stop; 0 have their whole-program passes measured
Scoped tail: 9 units stopped (checkRecordGuardedReads: 9).
NotYet: 611; Refused: 589.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 25 | 25 |
| a value of type SyntacticTypeNodeBuilderContext seen as NodeBuilderContext, whose readonly field tracker becomes writable: a readonly field may hold something narrower than Required&lt;Pick&lt;SymbolTracker, &quot;reportInferenceFallback&quot;&gt;&gt;, which a write of SymbolTrackerImpl would replace | #k8ed1f2 (unchecked) | 0 | 18 | 18 |
| a cast the runtime can&#x27;t check | #b5w3ycg (unchecked) | 0 | 14 | 14 |
| Object enumeration of a nullable spread with synthetic absent slots | OWNER BLANK | 11 | 0 | 11 |
| a checked view of FlowLabel over the union FlowNode, whose members no literal tag tells apart | OWNER BLANK | 10 | 0 | 10 |
| overload 1 of visitNode result TOut &#124; (TIn &amp; undefined) &#124; (TVisited &amp; undefined) cannot be served by implementation result Node &#124; undefined (return at /home/ahra/scratch/system_adamic/01M8H4EZVDQWGXFSAG50BRVB3J.meter/tmp/stage3-meter.giGgkC/main-adapted/src/compiler/visitorPublic.ts:156:9 isn&#x27;t the value its predicate checked) | OWNER BLANK | 0 | 10 | 10 |
| a cast that makes a viewed readonly array writable | #b5w3ycg (unchecked) | 9 | 0 | 9 |
| a cast the runtime can&#x27;t check: a cast over TypeSystemEntity, whose writable field flags some member can&#x27;t hold every value of, so a write through the cast could break that member | #b5w3ycg (unchecked) | 0 | 9 | 9 |
| a checked view of FlowCall over the union FlowNode, whose members no literal tag tells apart | OWNER BLANK | 8 | 0 | 8 |
| a checked view of VariableDeclarationList over the union ForInitializer, whose members no literal tag tells apart | OWNER BLANK | 8 | 0 | 8 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

This box's times, not comparable across boxes: 2784.556 wall seconds; 3277.888 CPU seconds.

| Slowest files | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| src/compiler/utilities.ts | 1030.306 | 1189.962 |
| src/compiler/factory/utilities.ts | 398.945 | 427.585 |
| src/compiler/factory/nodeTests.ts | 286.493 | 338.958 |
| src/compiler/utilitiesPublic.ts | 270.053 | 333.108 |
| src/compiler/transformers/declarations/diagnostics.ts | 53.511 | 66.413 |
