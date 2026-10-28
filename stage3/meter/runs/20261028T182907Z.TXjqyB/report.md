Whole program: main: 2/79; area: 2/79
Own file: main: 58/79; area: 58/79
tsc entry: main: fail; area: fail
tsc entry diagnostics: main: 144; area: 144

Stage 3 meter 20261028T182907Z

Main: origin/main at 727c02215960333047539dd05fc17775e5a77d1a. Area compatibility label at 727c02215960333047539dd05fc17775e5a77d1a.
Main measured with Adamic 727c02215960333047539dd05fc17775e5a77d1a; area measured with Adamic 727c02215960333047539dd05fc17775e5a77d1a. Each compiler is built from its pinned ref with ordinary stage 0 options.

| File | Main whole program | Main own file | Area whole program | Area own file |
| --- | --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.ts | fail | pass | fail | pass |
| src/compiler/binder.ts | fail | pass | fail | pass |
| src/compiler/builder.ts | fail | pass | fail | pass |
| src/compiler/builderPublic.ts | fail | pass | fail | pass |
| src/compiler/builderState.ts | fail | pass | fail | pass |
| src/compiler/builderStatePublic.ts | fail | pass | fail | pass |
| src/compiler/checker.ts | fail | fail | fail | fail |
| src/compiler/commandLineParser.ts | fail | fail | fail | fail |
| src/compiler/core.ts | fail | fail | fail | fail |
| src/compiler/corePublic.ts | pass | pass | pass | pass |
| src/compiler/debug.ts | fail | pass | fail | pass |
| src/compiler/diagnosticInformationMap.generated.ts | fail | pass | fail | pass |
| src/compiler/diagnosticMessages.generated.json | fail | fail | fail | fail |
| src/compiler/diagnosticMessages.json | fail | fail | fail | fail |
| src/compiler/emitter.ts | fail | pass | fail | pass |
| src/compiler/executeCommandLine.ts | fail | pass | fail | pass |
| src/compiler/expressionToTypeNode.ts | fail | pass | fail | pass |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | fail | pass |
| src/compiler/factory/emitHelpers.ts | fail | pass | fail | pass |
| src/compiler/factory/emitNode.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeChildren.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeConverters.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeFactory.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeTests.ts | fail | pass | fail | pass |
| src/compiler/factory/parenthesizerRules.ts | fail | pass | fail | pass |
| src/compiler/factory/utilities.ts | fail | pass | fail | pass |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | fail | pass |
| src/compiler/hostErrors.ts | pass | pass | pass | pass |
| src/compiler/moduleNameResolver.ts | fail | fail | fail | fail |
| src/compiler/moduleSpecifiers.ts | fail | pass | fail | pass |
| src/compiler/parser.ts | fail | pass | fail | pass |
| src/compiler/path.ts | fail | pass | fail | pass |
| src/compiler/performance.ts | fail | pass | fail | pass |
| src/compiler/performanceCore.ts | fail | pass | fail | pass |
| src/compiler/program.ts | fail | fail | fail | fail |
| src/compiler/programDiagnostics.ts | fail | pass | fail | pass |
| src/compiler/resolutionCache.ts | fail | fail | fail | fail |
| src/compiler/scanner.ts | fail | pass | fail | pass |
| src/compiler/semver.ts | fail | pass | fail | pass |
| src/compiler/sourcemap.ts | fail | fail | fail | fail |
| src/compiler/symbolWalker.ts | fail | pass | fail | pass |
| src/compiler/sys.ts | fail | fail | fail | fail |
| src/compiler/tracing.ts | fail | pass | fail | pass |
| src/compiler/transformer.ts | fail | fail | fail | fail |
| src/compiler/transformers/classFields.ts | fail | fail | fail | fail |
| src/compiler/transformers/classThis.ts | fail | pass | fail | pass |
| src/compiler/transformers/declarations.ts | fail | fail | fail | fail |
| src/compiler/transformers/declarations/diagnostics.ts | fail | pass | fail | pass |
| src/compiler/transformers/destructuring.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2016.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2017.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2018.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2019.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2020.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2021.ts | fail | pass | fail | pass |
| src/compiler/transformers/esDecorators.ts | fail | fail | fail | fail |
| src/compiler/transformers/esnext.ts | fail | fail | fail | fail |
| src/compiler/transformers/generators.ts | fail | fail | fail | fail |
| src/compiler/transformers/jsx.ts | fail | fail | fail | fail |
| src/compiler/transformers/legacyDecorators.ts | fail | pass | fail | pass |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | fail | pass |
| src/compiler/transformers/module/module.ts | fail | pass | fail | pass |
| src/compiler/transformers/module/system.ts | fail | pass | fail | pass |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | fail | pass |
| src/compiler/transformers/taggedTemplate.ts | fail | pass | fail | pass |
| src/compiler/transformers/ts.ts | fail | pass | fail | pass |
| src/compiler/transformers/typeSerializer.ts | fail | pass | fail | pass |
| src/compiler/transformers/utilities.ts | fail | pass | fail | pass |
| src/compiler/tsbuild.ts | fail | pass | fail | pass |
| src/compiler/tsbuildPublic.ts | fail | fail | fail | fail |
| src/compiler/tsconfig.json | fail | fail | fail | fail |
| src/compiler/types.ts | fail | pass | fail | pass |
| src/compiler/utilities.ts | fail | pass | fail | pass |
| src/compiler/utilitiesPublic.ts | fail | pass | fail | pass |
| src/compiler/visitorPublic.ts | fail | pass | fail | pass |
| src/compiler/watch.ts | fail | fail | fail | fail |
| src/compiler/watchPublic.ts | fail | fail | fail | fail |
| src/compiler/watchUtilities.ts | fail | pass | fail | pass |

tsc entry lowering, main: blocked by checker diagnostics.

tsc entry lowering, area: blocked by checker diagnostics.

### Unowned

measured on a checker-rejected program.
OWNER BLANK rows go to @system_adamic. Counts are grouped by exact reason.

| Reason | Owner | Main NotYet | Main Refused | Area NotYet | Area Refused |
| --- | --- | ---: | ---: | ---: | ---: |
| a union or optional field read in a program with record storage | OWNER BLANK | 70 | 0 | 70 | 0 |
| a record value of type TsConfigSourceFile &#124; CompilerOptionsValue | OWNER BLANK | 58 | 0 | 58 | 0 |
| a checked field alias of an optional field or through ?. | OWNER BLANK | 51 | 0 | 51 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 32 | 0 | 32 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 27 | 0 | 27 | 0 |
| a call through ?. (an optional call) | OWNER BLANK | 24 | 0 | 24 | 0 |
| an ESM namespace object used as a value; use qualified live bindings | OWNER BLANK | 19 | 0 | 19 | 0 |
| record values with an unsupported slot representation | OWNER BLANK | 15 | 0 | 15 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | OWNER BLANK | 0 | 14 | 0 | 14 |
| Object enumeration with getter or accessor descriptors | OWNER BLANK | 0 | 13 | 0 | 13 |
| a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | OWNER BLANK | 0 | 12 | 0 | 12 |
| node:process.Process.stdout | OWNER BLANK | 10 | 0 | 10 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | OWNER BLANK | 0 | 10 | 0 | 10 |

440 more unowned reasons, 1546 sites in all

Latent lowering, main: measured on a checker-rejected program.
NotYet: 843; Refused: 2411.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 536 | 536 |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 509 | 509 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | #1w5bve2 (Optimizer) | 0 | 255 | 255 |
| a union or optional field read in a program with record storage | OWNER BLANK | 70 | 0 | 70 |
| a record value of type TsConfigSourceFile &#124; CompilerOptionsValue | OWNER BLANK | 58 | 0 | 58 |
| a checked field alias of an optional field or through ?. | OWNER BLANK | 51 | 0 | 51 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 45 | 0 | 45 |
| a type predicate whose return is not proven (there is no body proving this parameter) | #1w5bve2 (Optimizer) | 0 | 33 | 33 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 32 | 32 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 27 | 0 | 27 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Latent lowering, area: measured on a checker-rejected program.
NotYet: 843; Refused: 2411.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 536 | 536 |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 509 | 509 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | #1w5bve2 (Optimizer) | 0 | 255 | 255 |
| a union or optional field read in a program with record storage | OWNER BLANK | 70 | 0 | 70 |
| a record value of type TsConfigSourceFile &#124; CompilerOptionsValue | OWNER BLANK | 58 | 0 | 58 |
| a checked field alias of an optional field or through ?. | OWNER BLANK | 51 | 0 | 51 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 45 | 0 | 45 |
| a type predicate whose return is not proven (there is no body proving this parameter) | #1w5bve2 (Optimizer) | 0 | 33 | 33 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 32 | 32 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 27 | 0 | 27 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Whole program includes imported diagnostics. Own file uses the primary diagnostic location.
Global and external diagnostics are counted separately in JSON; they have no compiler-file location.
Non-source inputs fail the extension gate and are excluded from source denominators.
