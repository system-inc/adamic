Ported the census onto main and added bounded, resumable whole-project streaming.
Compiler base 946a8f09; census implementation commits 039b4994 and 5eda5fb9; no production compiler edits.
Attempted all 79 files: 29 completed, 50 incomplete; 1,323,621 / 10,009,820 TypeScript bytes covered.
Independent depth/coverage audits, required mutants, Node/native mapper fixture and C/JS byte identity pass.
Timeouts are precise partial coverage; main mapper behavior is already present in 6b33ec61.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 819 | 498 |
| 1 | 169 | 79 |
| 2 | 52 | 23 |
| 3 | 21 | 20 |
| 4+ | 8 | 4 |
| Total | 1,069 | 624 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading SyntaxKind | 342 | 9 | 0 | 0 | 0 | 351 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 228 | 0 | 0 | 0 | 0 | 228 |
| NotYet | a generic function as a value | 78 | 15 | 6 | 0 | 0 | 99 |
| Refused | an object refinement using an open numeric enum as a literal tag | 76 | 14 | 0 | 0 | 0 | 90 |
| NotYet | an overloaded function as a value | 40 | 7 | 1 | 0 | 0 | 48 |
| NotYet | reading ScriptTarget | 39 | 0 | 0 | 0 | 0 | 39 |
| Refused | a cast the runtime can't check | 9 | 14 | 5 | 5 | 2 | 35 |
| NotYet | an assignment value to a member | 13 | 11 | 5 | 1 | 1 | 31 |
| NotYet | a value of type Path | 24 | 2 | 2 | 2 | 0 | 30 |
| Refused | inherited library member push read as an own field | 12 | 16 | 0 | 0 | 0 | 28 |
| NotYet | this outside a method | 2 | 16 | 3 | 5 | 0 | 26 |
| NotYet | a value of type T | 17 | 6 | 0 | 0 | 0 | 23 |
| NotYet | reading CharacterCodes | 23 | 0 | 0 | 0 | 0 | 23 |
| NotYet | a call through ?. (an optional call) | 7 | 8 | 6 | 0 | 0 | 21 |
| NotYet | a value of type __String | 19 | 0 | 0 | 0 | 0 | 19 |
| NotYet | a value of type any | 11 | 6 | 1 | 0 | 0 | 18 |
| Refused | inherited library member charCodeAt read as an own field | 16 | 0 | 0 | 0 | 0 | 16 |
| Refused | inherited library member slice read as an own field | 9 | 6 | 0 | 0 | 1 | 16 |
| Refused | inherited library member substring read as an own field | 11 | 1 | 3 | 0 | 0 | 15 |
| Refused | inherited library member get read as an own field | 8 | 2 | 3 | 1 | 0 | 14 |

Examined **1,323,621 / 10,009,820 TypeScript source bytes (13.223225%)**, in 29 files. Including non-TypeScript inputs, that is **1,323,621 / 10,615,807 bytes (12.468397%)**. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:

- `binder.ts`: 195,107 bytes; per-file measurement incomplete; see stream inventory.
- `builder.ts`: 117,298 bytes; per-file measurement incomplete; see stream inventory.
- `builderState.ts`: 26,209 bytes; per-file measurement incomplete; see stream inventory.
- `checker.ts`: 3,154,156 bytes; per-file measurement incomplete; see stream inventory.
- `commandLineParser.ts`: 183,733 bytes; per-file measurement incomplete; see stream inventory.
- `core.ts`: 92,680 bytes; per-file measurement incomplete; see stream inventory.
- `debug.ts`: 57,713 bytes; per-file measurement incomplete; see stream inventory.
- `diagnosticMessages.generated.json`: 306,771 bytes; non-TypeScript input; no lowering sites.
- `diagnosticMessages.json`: 299,061 bytes; non-TypeScript input; no lowering sites.
- `emitter.ts`: 274,627 bytes; per-file measurement incomplete; see stream inventory.
- `executeCommandLine.ts`: 54,177 bytes; per-file measurement incomplete; see stream inventory.
- `expressionToTypeNode.ts`: 70,354 bytes; per-file measurement incomplete; see stream inventory.
- `factory/emitHelpers.ts`: 69,184 bytes; per-file measurement incomplete; see stream inventory.
- `factory/emitNode.ts`: 12,800 bytes; per-file measurement incomplete; see stream inventory.
- `factory/nodeFactory.ts`: 336,646 bytes; per-file measurement incomplete; see stream inventory.
- `factory/parenthesizerRules.ts`: 34,072 bytes; per-file measurement incomplete; see stream inventory.
- `factory/utilities.ts`: 76,615 bytes; per-file measurement incomplete; see stream inventory.
- `moduleNameResolver.ts`: 183,618 bytes; per-file measurement incomplete; see stream inventory.
- `moduleSpecifiers.ts`: 80,070 bytes; per-file measurement incomplete; see stream inventory.
- `parser.ts`: 541,260 bytes; per-file measurement incomplete; see stream inventory.
- `program.ts`: 271,300 bytes; per-file measurement incomplete; see stream inventory.
- `resolutionCache.ts`: 82,223 bytes; per-file measurement incomplete; see stream inventory.
- `scanner.ts`: 219,390 bytes; per-file measurement incomplete; see stream inventory.
- `sourcemap.ts`: 32,210 bytes; per-file measurement incomplete; see stream inventory.
- `sys.ts`: 86,254 bytes; per-file measurement incomplete; see stream inventory.
- `tracing.ts`: 15,322 bytes; per-file measurement incomplete; see stream inventory.
- `transformer.ts`: 29,875 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/classFields.ts`: 156,676 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/declarations/diagnostics.ts`: 46,404 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/declarations.ts`: 99,818 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/destructuring.ts`: 30,894 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2015.ts`: 231,789 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2017.ts`: 50,181 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2018.ts`: 69,980 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2020.ts`: 13,128 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/esDecorators.ts`: 127,380 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/esnext.ts`: 32,645 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/generators.ts`: 124,036 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/jsx.ts`: 36,207 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/legacyDecorators.ts`: 36,740 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/module/esnextAnd2015.ts`: 18,686 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/module/module.ts`: 111,793 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/module/system.ts`: 85,481 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/ts.ts`: 116,741 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/utilities.ts`: 35,851 bytes; per-file measurement incomplete; see stream inventory.
- `tsbuildPublic.ts`: 114,655 bytes; per-file measurement incomplete; see stream inventory.
- `tsconfig.json`: 155 bytes; non-TypeScript input; no lowering sites.
- `utilities.ts`: 514,740 bytes; per-file measurement incomplete; see stream inventory.
- `utilitiesPublic.ts`: 104,608 bytes; per-file measurement incomplete; see stream inventory.
- `visitorPublic.ts`: 88,114 bytes; per-file measurement incomplete; see stream inventory.
- `watch.ts`: 45,924 bytes; per-file measurement incomplete; see stream inventory.
- `watchPublic.ts`: 61,201 bytes; per-file measurement incomplete; see stream inventory.
- `watchUtilities.ts`: 35,634 bytes; per-file measurement incomplete; see stream inventory.

Stock TypeScript independently matched the ancestry depth of 2,193 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: {"error": 2}. See method.md for their interpretation.

All retained observations come from one whole-compiler project on identical historical adapted bytes. Dependencies reached during completed walks contribute observations, but do not claim whole-file coverage. Incomplete files hit their 30-second wall limit; no inference is made about completion under a larger budget. Each job has a hard 12GiB address-space cap and a sampled 6GiB RSS watchdog.

The isolated checker probe timed out at 90.099s, peak RSS 940,780 KiB. Corpus timings below include up to two concurrent measurement jobs.

| Five largest inputs | Source bytes | Wall seconds | Peak RSS KiB | Exit |
|---|---:|---:|---:|---:|
| checker.ts | 3,154,156 | 30.050 | 790,988 | 124 |
| diagnosticInformationMap.generated.ts | 579,194 | 2.295 | 380,176 | 0 |
| parser.ts | 541,260 | 30.137 | 1,035,120 | 124 |
| utilities.ts | 514,740 | 30.113 | 811,660 | 124 |
| types.ts | 491,930 | 4.161 | 1,024,728 | 0 |

Full and no-stubs baselines attempted all 79 files under a separate 5-second limit, completed the same 21 files, and have byte-identical records for all 21. Remaining baseline files have no equality claim.

Compared identical f6bb0b41 adapted source bytes against main.
Only completed file records enter the common-file comparison; missing files have no measured delta.
Every added/removed unique observation and both top-20 tables are in DIFFERENCES.json.
Main changed lowering and registration since a5630a90. Port controls pass, but they do not establish the cause of each individual changed compiler site.
Individual attribution remains an inference unless backed by a compiler-change replay; these differences must not all be claimed as proven main changes.

| Directory | Completed / old files | Old common NY / Ref | Main common NY / Ref | Removed / added observations |
|---|---:|---:|---:|---:|
| _namespaces | 3 / 3 | 0 / 0 | 0 / 0 | 0 / 0 |
| factory | 7 / 10 | 627 / 342 | 612 / 340 | 21 / 4 |
| transformers/declarations | 1 / 1 | 417 / 31 | 396 / 31 | 22 / 1 |
| transformers/module | 1 / 4 | 16 / 3 | 16 / 3 | 0 / 0 |

Factory emitHelpers.ts, nodeFactory.ts and utilities.ts, and module esnextAnd2015.ts, module.ts and system.ts exceeded the directory rerun limits. No table difference is claimed for them. No same-site depth changes occur in the 12 completed historical files. Main source-history explanations are recorded as inferences in the ledger; individual compiler changes were not replayed.

See [every comparison delta](evidence/main-stream/comparison/DIFFERENCES.json), [streaming method](streaming.md), [controls and commands](validation.md), and [baseline parity](evidence/main-stream/baseline-comparison.json). Raw records, stock ancestry, all per-file metrics and named logs are losslessly retained in evidence/main-stream; SHA256.json inventories the archive.
