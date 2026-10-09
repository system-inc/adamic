Published audited partial speculative depth tables for four compiler directories.
Base `a5630a90` includes `ed6e2975` and mapper fix `69501280`; production compiler files are unchanged.
Coverage: `_namespaces/`, `factory/`, `transformers/declarations/`, `transformers/module/`, each loaded separately.
Stock TypeScript ancestry, exact depth/top-20 recount and no-stubs baselines pass for each directory.
The whole-compiler run was interrupted in `checker.ts`; remaining files are explicitly unexamined below.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 4,548 | 959 |
| 1 | 991 | 322 |
| 2 | 186 | 19 |
| 3 | 22 | 4 |
| 4+ | 5 | 11 |
| Total | 5,752 | 1,315 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading SyntaxKind | 971 | 42 | 62 | 0 | 0 | 1,075 |
| NotYet | an assignment value to a member | 973 | 57 | 16 | 1 | 0 | 1,047 |
| NotYet | a generic function as a value | 505 | 192 | 2 | 0 | 0 | 699 |
| Refused | an object refinement using an open numeric enum as a literal tag | 268 | 242 | 2 | 1 | 1 | 514 |
| NotYet | reading TransformFlags | 43 | 228 | 9 | 0 | 0 | 280 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 246 | 0 | 0 | 0 | 0 | 246 |
| NotYet | a union of differently held members variable a function value captures | 87 | 56 | 0 | 0 | 0 | 143 |
| NotYet | reading Diagnostics | 103 | 0 | 29 | 0 | 0 | 132 |
| NotYet | a namespace member without a lowered binding | 79 | 10 | 3 | 0 | 0 | 92 |
| NotYet | reading Debug | 0 | 79 | 10 | 1 | 0 | 90 |
| NotYet | a BinaryExpression with a value and a value | 87 | 0 | 0 | 0 | 0 | 87 |
| Refused | a cast the runtime can't check | 41 | 24 | 6 | 2 | 6 | 79 |
| Refused | inherited library member push read as an own field | 70 | 3 | 0 | 0 | 0 | 73 |
| NotYet | reading visitNode | 65 | 3 | 0 | 0 | 0 | 68 |
| NotYet | a boolean &#124; undefined variable a function value captures | 50 | 15 | 0 | 0 | 0 | 65 |
| NotYet | reading setTextRange | 49 | 2 | 0 | 0 | 0 | 51 |
| NotYet | a value of type T | 8 | 39 | 0 | 0 | 0 | 47 |
| NotYet | reading EmitFlags | 39 | 4 | 1 | 0 | 0 | 44 |
| NotYet | reading ModifierFlags | 42 | 2 | 0 | 0 | 0 | 44 |
| NotYet | reading isExpression | 42 | 2 | 0 | 0 | 0 | 44 |

| Covered directory | Files | Speculative sites | Full/no-stubs sites |
|---|---:|---:|---:|
| `_namespaces/` | 3 | 0 | 0 |
| `factory/` | 10 | 5,160 | 1,371 |
| `transformers/declarations/` | 1 | 448 | 108 |
| `transformers/module/` | 4 | 1,459 | 640 |

Examined **851,207 / 10,009,820 TypeScript bytes (8.503719%)**, in **18 / 79 files**. Including non-TypeScript files: **851,207 / 10,615,807 bytes (8.018298%)**. All covered AST children were visited. Comments and whitespace are included in parsed source bytes; type and structural nodes are traversed rather than passed to a value lowerer. This is speculative examination, not successful production lowering.

These are four independent directory-root projects on the original adapted bytes. Imports remain available to the checker, but registration and the census walk use each directory's roots. Isolated binding context can change observed reasons and depths compared with a whole-compiler project; these counts must not be presented as the final full-corpus totals. Attempts attributed outside a project directory are excluded from its site tables.

The previous whole-compiler run stopped during checker.ts without emitting its buffered records. Completed progress lines are not depth evidence and are not included. The earlier PARTIAL.json and legacy raw evidence from e46fa73d have unverified depth tags and are superseded only for these measured directory projects.

Unexamined files, including all remaining TypeScript sources:

- `binder.ts`: 195,107 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `builder.ts`: 117,298 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `builderPublic.ts`: 11,609 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `builderState.ts`: 26,209 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `builderStatePublic.ts`: 392 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `checker.ts`: 3,154,156 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `commandLineParser.ts`: 183,733 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `core.ts`: 92,680 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `corePublic.ts`: 1,320 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `debug.ts`: 57,713 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `diagnosticInformationMap.generated.ts`: 579,194 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `diagnosticMessages.generated.json`: 306,771 bytes; non-TypeScript input; no lowering sites.
- `diagnosticMessages.json`: 299,061 bytes; non-TypeScript input; no lowering sites.
- `emitter.ts`: 274,627 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `executeCommandLine.ts`: 54,177 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `expressionToTypeNode.ts`: 70,354 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `hostErrors.ts`: 447 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `moduleNameResolver.ts`: 183,618 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `moduleSpecifiers.ts`: 80,070 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `parser.ts`: 541,260 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `path.ts`: 44,181 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `performance.ts`: 6,161 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `performanceCore.ts`: 3,279 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `program.ts`: 271,300 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `programDiagnostics.ts`: 21,401 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `resolutionCache.ts`: 82,223 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `scanner.ts`: 219,390 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `semver.ts`: 18,012 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `sourcemap.ts`: 32,210 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `symbolWalker.ts`: 8,267 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `sys.ts`: 86,254 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `tracing.ts`: 15,322 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformer.ts`: 29,875 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/classFields.ts`: 156,676 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/classThis.ts`: 5,565 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/declarations.ts`: 99,818 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/destructuring.ts`: 30,894 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2015.ts`: 231,789 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2016.ts`: 4,519 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2017.ts`: 50,181 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2018.ts`: 69,980 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2019.ts`: 1,563 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2020.ts`: 13,128 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/es2021.ts`: 4,144 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/esDecorators.ts`: 127,380 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/esnext.ts`: 32,645 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/generators.ts`: 124,036 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/jsx.ts`: 36,207 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/legacyDecorators.ts`: 36,740 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/namedEvaluation.ts`: 22,318 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/taggedTemplate.ts`: 5,489 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/ts.ts`: 116,741 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/typeSerializer.ts`: 28,972 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `transformers/utilities.ts`: 35,851 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `tsbuild.ts`: 5,332 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `tsbuildPublic.ts`: 114,655 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `tsconfig.json`: 155 bytes; non-TypeScript input; no lowering sites.
- `types.ts`: 491,930 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `utilities.ts`: 514,740 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `utilitiesPublic.ts`: 104,608 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `visitorPublic.ts`: 88,114 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `watch.ts`: 45,924 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `watchPublic.ts`: 61,201 bytes; outside covered directories; full census interrupted before buffered observations were emitted.
- `watchUtilities.ts`: 35,634 bytes; outside covered directories; full census interrupted before buffered observations were emitted.

Normal C and JS output with the overlay off remains byte-identical to base a5630a90 on the isolation witness. Nested controls prove depth 1 and 2, checker-clean controls reach depth 3 and 4, and the equal-span witness distinguishes identifier/parameter ancestry. All six .a controls pass local a-check with zero mismatches; eight header mutants are caught. The original 13:00 MDT deadline was missed.

See [validation and mutants](validation.md), [method](method.md), [fixture counts](counts.md), [header validation](header-validation.md), and [evidence manifest](evidence/manifest.json). Reproduce this partial report with `partial.py ADAPTED_COMPILER_DIRECTORY RUN_DIRECTORY OUTPUT_DIRECTORY BINARY`, after running the four directory speculative measurements. Per-directory depth tables are in `directories/`; raw observations and stock ancestry are retained under `evidence/directories/`.
