# ts object key audit

Scanned 39,470 JavaScript/TypeScript/JSX source and fixture files from the tracked TypeScript checkout plus generated src files. Root ts-object key enumerations: **0**. The scan includes harness, test runners, test cases and checked-in baselines; it recognizes namespace aliases, Object key/value/property enumeration, Object.assign, Reflect.ownKeys, for-in, spreads and object-rest.

The following six related reads operate on nested values:

| File and line | Operation | Reads |
| --- | --- | --- |
| src/harness/harnessLanguageService.ts:761 | spread | ts.server.nullTypingsInstaller |
| src/testRunner/unittests/config/commandLineParsing.ts:117 | spread | ts.compilerOptionsDidYouMeanDiagnostics.optionDeclarations |
| src/testRunner/unittests/config/commandLineParsing.ts:128 | spread | ts.compilerOptionsDidYouMeanDiagnostics |
| src/testRunner/unittests/services/textChanges.ts:25 | spread | ts.testFormatSettings |
| src/testRunner/unittests/tsserver/session.ts:215 | Object.values | (ts.server.protocol as any).CommandTypes |
| src/tsserver/common.ts:7 | for-in | ts.server.LogLevel |

Generated bundling machinery is separate from existing source consumers. __export iterates the generated descriptor literal named all to define namespace getters. These descriptor tables include the new functions:

| Bundle | Descriptor table | Call line |
| --- | --- | ---: |
| run.js | ts_exports2 | 197081 |
| run.js | ts_exports3 | 200396 |
| run.js | ts_exports4 | 209562 |
| typescript.js | typescript_exports | 48 |
| typescript.js | ts_exports2 | 184151 |

The generated __copyProps helper reads arbitrary module own-property names for module interop. The TypeScript bundle uses the modified __toCommonJS helper that returns its module object directly. No existing TypeScript source consumer enumerates the root ts object. All generated namespace/export additions are covered by the complete emitted JavaScript diff and AST reversion proof.

The proof script compare.cjs deliberately calls Object.keys on the before and after runtime objects to assert exactly two added properties, orPredicate2 and orPredicate3, and zero removed properties. Those two proof reads are outside the TypeScript source checkout being audited.

Three eslint.config.mjs spreads belong to typescript-eslint's tseslint.configs namespace and are excluded from the TypeScript runtime-object count.
