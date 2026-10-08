Whole program: 2/79
Own file: 56/79

Roots (target 79/79): tree: 79 total, 78 tsc original
Tsc original roots, tree: whole program 1/78; own file 55/78.
Adaptation-created root, tree: src/compiler/hostErrors.ts (adaptation 47-host-errors).

Stage 3 meter 20261008T203110Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.ts | fail | pass | blocked |
| src/compiler/binder.ts | fail | pass | blocked |
| src/compiler/builder.ts | fail | fail | blocked |
| src/compiler/builderPublic.ts | fail | pass | blocked |
| src/compiler/builderState.ts | fail | pass | blocked |
| src/compiler/builderStatePublic.ts | fail | pass | blocked |
| src/compiler/checker.ts | fail | fail | blocked |
| src/compiler/commandLineParser.ts | fail | fail | blocked |
| src/compiler/core.ts | fail | fail | blocked |
| src/compiler/corePublic.ts | pass | pass | fail |
| src/compiler/debug.ts | fail | fail | blocked |
| src/compiler/diagnosticInformationMap.generated.ts | fail | pass | blocked |
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked |
| src/compiler/diagnosticMessages.json | fail | fail | blocked |
| src/compiler/emitter.ts | fail | pass | blocked |
| src/compiler/executeCommandLine.ts | fail | pass | blocked |
| src/compiler/expressionToTypeNode.ts | fail | pass | blocked |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | blocked |
| src/compiler/factory/emitHelpers.ts | fail | pass | blocked |
| src/compiler/factory/emitNode.ts | fail | pass | blocked |
| src/compiler/factory/nodeChildren.ts | fail | pass | blocked |
| src/compiler/factory/nodeConverters.ts | fail | pass | blocked |
| src/compiler/factory/nodeFactory.ts | fail | pass | blocked |
| src/compiler/factory/nodeTests.ts | fail | pass | blocked |
| src/compiler/factory/parenthesizerRules.ts | fail | pass | blocked |
| src/compiler/factory/utilities.ts | fail | pass | blocked |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | blocked |
| src/compiler/hostErrors.ts | pass | pass | pass |
| src/compiler/moduleNameResolver.ts | fail | fail | blocked |
| src/compiler/moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/parser.ts | fail | pass | blocked |
| src/compiler/path.ts | fail | pass | blocked |
| src/compiler/performance.ts | fail | pass | blocked |
| src/compiler/performanceCore.ts | fail | pass | blocked |
| src/compiler/program.ts | fail | fail | blocked |
| src/compiler/programDiagnostics.ts | fail | pass | blocked |
| src/compiler/resolutionCache.ts | fail | fail | blocked |
| src/compiler/scanner.ts | fail | pass | blocked |
| src/compiler/semver.ts | fail | pass | blocked |
| src/compiler/sourcemap.ts | fail | fail | blocked |
| src/compiler/symbolWalker.ts | fail | pass | blocked |
| src/compiler/sys.ts | fail | fail | blocked |
| src/compiler/tracing.ts | fail | pass | blocked |
| src/compiler/transformer.ts | fail | fail | blocked |
| src/compiler/transformers/classFields.ts | fail | fail | blocked |
| src/compiler/transformers/classThis.ts | fail | pass | blocked |
| src/compiler/transformers/declarations.ts | fail | fail | blocked |
| src/compiler/transformers/declarations/diagnostics.ts | fail | pass | blocked |
| src/compiler/transformers/destructuring.ts | fail | pass | blocked |
| src/compiler/transformers/es2015.ts | fail | fail | blocked |
| src/compiler/transformers/es2016.ts | fail | pass | blocked |
| src/compiler/transformers/es2017.ts | fail | fail | blocked |
| src/compiler/transformers/es2018.ts | fail | pass | blocked |
| src/compiler/transformers/es2019.ts | fail | pass | blocked |
| src/compiler/transformers/es2020.ts | fail | pass | blocked |
| src/compiler/transformers/es2021.ts | fail | pass | blocked |
| src/compiler/transformers/esDecorators.ts | fail | fail | blocked |
| src/compiler/transformers/esnext.ts | fail | fail | blocked |
| src/compiler/transformers/generators.ts | fail | fail | blocked |
| src/compiler/transformers/jsx.ts | fail | fail | blocked |
| src/compiler/transformers/legacyDecorators.ts | fail | pass | blocked |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | blocked |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | blocked |
| src/compiler/transformers/module/module.ts | fail | pass | blocked |
| src/compiler/transformers/module/system.ts | fail | pass | blocked |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | blocked |
| src/compiler/transformers/taggedTemplate.ts | fail | pass | blocked |
| src/compiler/transformers/ts.ts | fail | pass | blocked |
| src/compiler/transformers/typeSerializer.ts | fail | pass | blocked |
| src/compiler/transformers/utilities.ts | fail | pass | blocked |
| src/compiler/tsbuild.ts | fail | pass | blocked |
| src/compiler/tsbuildPublic.ts | fail | fail | blocked |
| src/compiler/tsconfig.json | fail | fail | blocked |
| src/compiler/types.ts | fail | pass | blocked |
| src/compiler/utilities.ts | fail | pass | blocked |
| src/compiler/utilitiesPublic.ts | fail | pass | blocked |
| src/compiler/visitorPublic.ts | fail | pass | blocked |
| src/compiler/watch.ts | fail | fail | blocked |
| src/compiler/watchPublic.ts | fail | fail | blocked |
| src/compiler/watchUtilities.ts | fail | pass | blocked |

Checker: 2/79 source files. Lowering: 1/79 source files.
All files: 82; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.

tsc entry lowering, main: measured on a checker-rejected entry-root program.
Resolved source files: 81; NotYet: 1551; Refused: 8412.
Errors: 0; panics: 10558; skipped dependencies: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | OWNER BLANK | 0 | 2776 | 2776 |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 2141 | 2141 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | 01a1143b-d691 | 0 | 414 | 414 |
| an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) | OWNER BLANK | 0 | 314 | 314 |
| checked view field expression of type LeftHandSideExpression | OWNER BLANK | 138 | 0 | 138 |
| checked view field expression of type Expression | OWNER BLANK | 108 | 0 | 108 |
| checked view field left of type Expression | OWNER BLANK | 108 | 0 | 108 |
| an unproven predicate argument for parameter test (argument &quot;isTypeNode&quot;) | OWNER BLANK | 0 | 84 | 84 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 83 | 83 |
| checked view field modifiers of type NodeArray&lt;ModifierLike&gt; &#124; undefined | OWNER BLANK | 82 | 0 | 82 |

Counts cover the entry root and its resolved implementation dependencies, excluding declarations.
Errors and panics are retained separately in JSON. This census does not establish native output.

Latent lowering, main: measured on a checker-rejected program.
NotYet: 1551; Refused: 8410.
Recovery boundaries: 10550; named nested checker exclusions: 110.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | OWNER BLANK | 0 | 2776 | 2776 |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 2141 | 2141 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | 01a1143b-d691 | 0 | 414 | 414 |
| an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) | OWNER BLANK | 0 | 314 | 314 |
| checked view field expression of type LeftHandSideExpression | OWNER BLANK | 138 | 0 | 138 |
| checked view field expression of type Expression | OWNER BLANK | 108 | 0 | 108 |
| checked view field left of type Expression | OWNER BLANK | 108 | 0 | 108 |
| an unproven predicate argument for parameter test (argument &quot;isTypeNode&quot;) | OWNER BLANK | 0 | 84 | 84 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 83 | 83 |
| checked view field modifiers of type NodeArray&lt;ModifierLike&gt; &#124; undefined | OWNER BLANK | 82 | 0 | 82 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.
