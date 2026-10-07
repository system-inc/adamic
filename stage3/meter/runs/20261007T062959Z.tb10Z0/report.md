Whole program: main: 0/78; area: 1/79
Own file: main: 22/78; area: 25/79

Stage 3 meter 20261007T062959Z

Main: origin/main at b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Area: origin/area/stage3 at e39a299323cad76aea44ec71ca7b740130323d4c.
Both measured with Adamic e39a299323cad76aea44ec71ca7b740130323d4c and ordinary stage 0 options.

| File | Main whole program | Main own file | Area whole program | Area own file |
| --- | --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.ts | fail | pass | fail | pass |
| src/compiler/binder.ts | fail | fail | fail | fail |
| src/compiler/builder.ts | fail | fail | fail | fail |
| src/compiler/builderPublic.ts | fail | pass | fail | pass |
| src/compiler/builderState.ts | fail | fail | fail | fail |
| src/compiler/builderStatePublic.ts | fail | pass | fail | pass |
| src/compiler/checker.ts | fail | fail | fail | fail |
| src/compiler/commandLineParser.ts | fail | fail | fail | fail |
| src/compiler/core.ts | fail | fail | fail | fail |
| src/compiler/corePublic.ts | fail | fail | fail | fail |
| src/compiler/debug.ts | fail | fail | fail | fail |
| src/compiler/diagnosticInformationMap.generated.ts | fail | fail | fail | pass |
| src/compiler/diagnosticMessages.generated.json | fail | fail | fail | fail |
| src/compiler/diagnosticMessages.json | fail | fail | fail | fail |
| src/compiler/emitter.ts | fail | fail | fail | fail |
| src/compiler/executeCommandLine.ts | fail | fail | fail | fail |
| src/compiler/expressionToTypeNode.ts | fail | fail | fail | fail |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | fail | pass |
| src/compiler/factory/emitHelpers.ts | fail | fail | fail | fail |
| src/compiler/factory/emitNode.ts | fail | fail | fail | fail |
| src/compiler/factory/nodeChildren.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeConverters.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeFactory.ts | fail | fail | fail | fail |
| src/compiler/factory/nodeTests.ts | fail | pass | fail | pass |
| src/compiler/factory/parenthesizerRules.ts | fail | fail | fail | fail |
| src/compiler/factory/utilities.ts | fail | fail | fail | fail |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | fail | pass |
| src/compiler/hostErrors.ts | absent | absent | pass | pass |
| src/compiler/moduleNameResolver.ts | fail | fail | fail | fail |
| src/compiler/moduleSpecifiers.ts | fail | fail | fail | fail |
| src/compiler/parser.ts | fail | fail | fail | fail |
| src/compiler/path.ts | fail | fail | fail | fail |
| src/compiler/performance.ts | fail | pass | fail | pass |
| src/compiler/performanceCore.ts | fail | fail | fail | fail |
| src/compiler/program.ts | fail | fail | fail | fail |
| src/compiler/programDiagnostics.ts | fail | pass | fail | pass |
| src/compiler/resolutionCache.ts | fail | fail | fail | fail |
| src/compiler/scanner.ts | fail | fail | fail | fail |
| src/compiler/semver.ts | fail | pass | fail | pass |
| src/compiler/sourcemap.ts | fail | fail | fail | fail |
| src/compiler/symbolWalker.ts | fail | pass | fail | pass |
| src/compiler/sys.ts | fail | fail | fail | fail |
| src/compiler/tracing.ts | fail | fail | fail | fail |
| src/compiler/transformer.ts | fail | fail | fail | fail |
| src/compiler/transformers/classFields.ts | fail | fail | fail | fail |
| src/compiler/transformers/classThis.ts | fail | pass | fail | pass |
| src/compiler/transformers/declarations.ts | fail | fail | fail | fail |
| src/compiler/transformers/declarations/diagnostics.ts | fail | fail | fail | fail |
| src/compiler/transformers/destructuring.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2016.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2017.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2018.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2019.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2020.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2021.ts | fail | pass | fail | pass |
| src/compiler/transformers/esDecorators.ts | fail | fail | fail | fail |
| src/compiler/transformers/esnext.ts | fail | fail | fail | fail |
| src/compiler/transformers/generators.ts | fail | fail | fail | fail |
| src/compiler/transformers/jsx.ts | fail | fail | fail | fail |
| src/compiler/transformers/legacyDecorators.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | fail | pass |
| src/compiler/transformers/module/module.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/system.ts | fail | fail | fail | fail |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | fail | pass |
| src/compiler/transformers/taggedTemplate.ts | fail | fail | fail | fail |
| src/compiler/transformers/ts.ts | fail | fail | fail | fail |
| src/compiler/transformers/typeSerializer.ts | fail | pass | fail | pass |
| src/compiler/transformers/utilities.ts | fail | fail | fail | fail |
| src/compiler/tsbuild.ts | fail | fail | fail | fail |
| src/compiler/tsbuildPublic.ts | fail | fail | fail | fail |
| src/compiler/tsconfig.json | fail | fail | fail | fail |
| src/compiler/types.ts | fail | fail | fail | fail |
| src/compiler/utilities.ts | fail | fail | fail | fail |
| src/compiler/utilitiesPublic.ts | fail | fail | fail | fail |
| src/compiler/visitorPublic.ts | fail | fail | fail | pass |
| src/compiler/watch.ts | fail | fail | fail | fail |
| src/compiler/watchPublic.ts | fail | fail | fail | fail |
| src/compiler/watchUtilities.ts | fail | fail | fail | fail |

### Unowned

measured on a checker-rejected program.
OWNER BLANK rows go to @system_adamic. Counts are grouped by exact reason.

| Reason | Owner | Main NotYet | Main Refused | Area NotYet | Area Refused |
| --- | --- | ---: | ---: | ---: | ---: |
| a BinaryExpression with a value and a value | OWNER BLANK | 32 | 0 | 32 | 0 |
| a value of type T | OWNER BLANK | 30 | 0 | 30 | 0 |
| a value of type any | OWNER BLANK | 30 | 0 | 30 | 0 |
| a field of type boolean &#124; undefined | OWNER BLANK | 25 | 0 | 25 | 0 |
| a cast the runtime can&#x27;t check | OWNER BLANK | 0 | 23 | 0 | 23 |
| a parameter that isn&#x27;t a plain name | OWNER BLANK | 22 | 0 | 22 | 0 |
| reading ModifierFlags | OWNER BLANK | 22 | 0 | 22 | 0 |
| reading Extension | OWNER BLANK | 20 | 0 | 21 | 0 |
| reading NodeFlags | OWNER BLANK | 20 | 0 | 20 | 0 |
| a value of type Path | OWNER BLANK | 19 | 0 | 19 | 0 |
| a BinaryExpression with a value and a boolean | OWNER BLANK | 18 | 0 | 18 | 0 |
| a value of type ResolvedConfigFilePath | OWNER BLANK | 18 | 0 | 18 | 0 |
| a value of type __String | OWNER BLANK | 18 | 0 | 18 | 0 |
| &#124;&#124;= | OWNER BLANK | 0 | 15 | 0 | 18 |
| a function returning U &#124; undefined | OWNER BLANK | 16 | 0 | 16 | 0 |
| a string as a condition | OWNER BLANK | 0 | 14 | 0 | 14 |
| a generic function as a value | OWNER BLANK | 13 | 0 | 13 | 0 |
| a value of type unknown | OWNER BLANK | 9 | 0 | 13 | 0 |
| a boolean &#124; undefined as a condition | OWNER BLANK | 0 | 11 | 0 | 12 |
| reading Comparison | OWNER BLANK | 12 | 0 | 12 | 0 |
| a namespace | OWNER BLANK | 0 | 11 | 0 | 11 |
| a value of type T &#124; undefined | OWNER BLANK | 11 | 0 | 11 | 0 |
| an index signature | OWNER BLANK | 0 | 8 | 0 | 11 |
| reading ModuleKind | OWNER BLANK | 11 | 0 | 11 | 0 |
| a call through ?. (an optional call) | OWNER BLANK | 9 | 0 | 10 | 0 |
| a function returning __String | OWNER BLANK | 10 | 0 | 10 | 0 |

319 more unowned reasons, 1293 sites in all

Latent lowering, main: measured on a checker-rejected program.
NotYet: 2059; Refused: 2140.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a type predicate | 01a1143b-d691 | 0 | 588 | 588 |
| reading SyntaxKind | compiler/stage3-front | 484 | 0 | 484 |
| the non-null assertion ! | 01a1130a | 0 | 484 | 484 |
| a function without a body | 01a1143b-f5d4 | 191 | 0 | 191 |
| enum | compiler/stage3-front | 0 | 162 | 162 |
| an EnumDeclaration | compiler/stage3-front | 154 | 0 | 154 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 85 | 0 | 85 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 80 | 0 | 80 |
| an ExportDeclaration | 01a113e3-a058 | 0 | 77 | 77 |
| a function returning T &#124; undefined | 01a1143c | 66 | 0 | 66 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Latent lowering, area: measured on a checker-rejected program.
NotYet: 2075; Refused: 2104.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a type predicate | 01a1143b-d691 | 0 | 588 | 588 |
| the non-null assertion ! | 01a1130a | 0 | 520 | 520 |
| reading SyntaxKind | compiler/stage3-front | 484 | 0 | 484 |
| a function without a body | 01a1143b-f5d4 | 191 | 0 | 191 |
| enum | compiler/stage3-front | 0 | 162 | 162 |
| an EnumDeclaration | compiler/stage3-front | 154 | 0 | 154 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 86 | 0 | 86 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 82 | 0 | 82 |
| an ExportDeclaration | 01a113e3-a058 | 0 | 77 | 77 |
| a function returning T &#124; undefined | 01a1143c | 66 | 0 | 66 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Whole program includes imported diagnostics. Own file uses the primary diagnostic location.
Global and external diagnostics are counted separately in JSON; they have no compiler-file location.
Non-source inputs fail the extension gate and are excluded from source denominators.
