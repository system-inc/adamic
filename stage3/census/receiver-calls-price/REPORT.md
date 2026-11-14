# Split receiver-call price

**Rewrite: 46 parser uses, 52 exact-parameter arrows across 17 transformer files, and 10 readonly printer aliases; 142 lines added / 189 removed across 19 files.** Observed writes: **0** (typed assignments 0, defineProperty writes 0). Direct member identity / `.length` reads: **0 / 0**. Implementation `this` reads: **0 by AST**. Broader function-object identity comparisons: **12,505,584**; broader length reads: **59,188**, with attribution below.

**Cost per complete lane:** parser **18,022,708 property reads**; transformers **189,211 forwarding calls** and **1,601,433 arrow initializations** (calls match the certified 189,211); printer **739,160 alias reads** at 73,916 entries, with **zero forwarding calls**. Against the hybrid’s **30,851,760 additional calls**, the split retains 0.613% of those calls. This is operation counting, not timing of dispatch or a performance forecast.

**Printer result: readonly-view option (i) reaches all ten alias locals with no finding there. Option (ii) was not needed and was not run.**

## Scope and measured price

The AST recipe is [split-codemod.cjs](split-codemod.cjs). Its safety scan collects every unsafe parser use before writing source, including shorthand value uses, shadowed declarations and non-callee references. The verified regeneration reproduces all 19 measured source files byte-for-byte.

The parser adds 18,022,708 per-use lookups and removes 26 cached reads per namespace initialization. The certified workload has 9 such initializations, so the baseline comparison removes 234 reads: net **18,022,474 added parser reads**. The unchanged pre-existing receiver call to `factory.createToken` is excluded from the added-read counter.

The 52 transformer arrow initializers execute **1,601,433 times** and the arrows are invoked **189,211 times**. These are initializer executions, not a claim about allocations surviving V8 optimization. The arrows replace 1,601,433 cached method reads with 189,211 per-call reads (net -1,412,222). The `factory` property bindings stay in place. All original 127 direct uses and 46 value references stay unchanged: 44 callback arguments plus two branches of one conditional callback argument. Every arrow uses the exact declared parameter list and return type, without rest parameters. All 52 arrow `context` identifiers resolve to their transformer’s own parameter.

The printer resolves ten members once per entry: 739,160 member reads, matching the original ten-member destructure, plus one object alias per entry. The two defaults use `??`; the same handlers object and original function values are retained. No printer use is rewritten. The three existing sentinel comparisons remain. Their total execution count is 20,816,897; the broader registered-function observer counts only comparisons involving a registered function value, not all default-sentinel executions.

## Four counter checks and attribution

The final observer unwraps optional types and follows base interfaces, including TextChangesWriter extending PrintHandlers. Its late-registration counter checks that receivers are observed at creation, rather than inferring lifetime coverage from a zero write total. Earlier counting versions with entry-only gaps are not used to support the final lifetime-write claim.

Counters run only in private counting copies. Writes cover all NodeFactory/PrintHandlers members and the taken method-signature members of TransformationContext; its unchanged factory property and mutable transformation hooks are outside that latter scope. Instrumentation installation writes are excluded. In-place data-property accessors preserve receiver objects and function values; typed assignment sites and Object.defineProperty calls are also observed. Function length getters retain the original length value. Equality observations retain the original comparison expression for TypeScript narrowing, and inspect only side-effect-free local/constant operands. The prototype has none of this instrumentation. A separate virtual-source negative control reports all five unsafe references/declarations (shorthand value use, both assignment references, a shadowed local and a shadowed receiver) with file:line, exits 2, and performs zero compiler-source writes.

- Writes: setter 0, typed assignment 0, defineProperty 0. These instruments overlap and must not be summed as distinct writes. Observer descriptor failures: 0. Late receiver registrations: `{}`; creation-time coverage is limited wherever a late registration is reported.

- Identity: direct receiver comparisons 0; broader registered-function comparisons 12,505,584. Per-site counts are in [split-counts.json](evidence/split-counts.json). These include the preserved printer sentinel identity uses; a blanket zero claim for alias identity would be false. A shared implementation such as `noop` also participates in unrelated comparisons after it is registered.

- Length: direct receiver reads 0; new-arrow length reads **0**; broader function-object length reads 59,188. Breakdown and first stack samples are in [split-metadata-attribution.json](evidence/split-metadata-attribution.json). All 59,188 hits are implicit Function.bind reads on shared primitives: 59,176 noop calls in host synchronization and 12 notImplemented calls in host.readFile binding. None is a new-arrow length read. A control proves that registering the null context’s `noop` and later binding the same global `noop` for an unrelated host produces a broad-counter hit; function-object hits alone do not establish a read through a changed binding.

- This: AST over Function.prototype.toString for every observed selected receiver implementation: 47 unique bodies, 0 parse failures, 0 this reads. Nested arrows are included; nested ordinary functions are excluded because their this is independent. The two default implementations are separately AST-checked, also zero. This is not an instrumented count of executed ThisKeyword expressions.

Positive controls execute the actual AST instrumenter and demonstrate reads, writes, defineProperty, direct identity, alias identity, direct and alias length, parser calls, transformer calls and printer aliases. A deliberate method using this.flag yields one AST this read. Controls and failed attempts are outside lane snapshots. Printer controls additionally verify once-only reads in original member order, the original getter receiver, bare callback this, explicit sentinels and printing const answer = 42.

## Rewrite size and per-file price

Added/removed counts compare logical source lines with the adapted baseline. “Parser uses” counts replaced callee references; transformer and printer uses are retained. Per-member signatures, use counts and all binding checks are in [split-rewrite.json](evidence/split-rewrite.json).

| File | Added | Removed | Parser uses | Arrows | Arrow calls / lane |
| --- | ---: | ---: | ---: | ---: | ---: |
| src/compiler/parser.ts | 47 | 74 | 46 | 0 | 0 |
| src/compiler/transformers/ts.ts | 7 | 8 | 0 | 5 | 16,794 |
| src/compiler/transformers/classFields.ts | 8 | 9 | 0 | 6 | 12,423 |
| src/compiler/transformers/typeSerializer.ts | 2 | 4 | 0 | 1 | 38 |
| src/compiler/transformers/legacyDecorators.ts | 4 | 5 | 0 | 2 | 1,627 |
| src/compiler/transformers/esDecorators.ts | 6 | 7 | 0 | 4 | 6,469 |
| src/compiler/transformers/es2017.ts | 6 | 7 | 0 | 4 | 7,645 |
| src/compiler/transformers/es2018.ts | 6 | 7 | 0 | 4 | 55,851 |
| src/compiler/transformers/es2020.ts | 3 | 4 | 0 | 1 | 432 |
| src/compiler/transformers/es2021.ts | 3 | 4 | 0 | 1 | 36 |
| src/compiler/transformers/esnext.ts | 6 | 7 | 0 | 4 | 3,387 |
| src/compiler/transformers/jsx.ts | 3 | 4 | 0 | 1 | 148 |
| src/compiler/transformers/es2016.ts | 3 | 4 | 0 | 1 | 228 |
| src/compiler/transformers/es2015.ts | 7 | 8 | 0 | 5 | 36,312 |
| src/compiler/transformers/generators.ts | 7 | 8 | 0 | 5 | 4,142 |
| src/compiler/transformers/module/module.ts | 6 | 7 | 0 | 4 | 25,281 |
| src/compiler/transformers/module/system.ts | 4 | 6 | 0 | 3 | 2,493 |
| src/compiler/transformers/module/esnextAnd2015.ts | 3 | 4 | 0 | 1 | 15,905 |
| src/compiler/emitter.ts | 11 | 12 | 0 | 0 | 0 |

### Every parser member

| Member | Rewritten uses | Reads / lane |
| --- | ---: | ---: |
| createNodeArray | 2 | 4,068,227 |
| createNumericLiteral | 2 | 377,494 |
| createStringLiteral | 2 | 375,394 |
| createLiteralLikeNode | 1 | 2,671 |
| createIdentifier | 5 | 8,254,453 |
| createPrivateIdentifier | 1 | 4,176 |
| createToken | 7 | 3,399,650 |
| createArrayLiteralExpression | 2 | 81,897 |
| createObjectLiteralExpression | 1 | 67,014 |
| createPropertyAccessExpression | 4 | 151,152 |
| createPropertyAccessChain | 2 | 812 |
| createElementAccessExpression | 1 | 40,053 |
| createElementAccessChain | 1 | 217 |
| createCallExpression | 1 | 106,406 |
| createCallChain | 1 | 394 |
| createNewExpression | 1 | 16,513 |
| createParenthesizedExpression | 1 | 44,643 |
| createBlock | 2 | 127,614 |
| createVariableStatement | 1 | 227,665 |
| createExpressionStatement | 2 | 184,328 |
| createIfStatement | 1 | 12,746 |
| createWhileStatement | 1 | 1,250 |
| createForStatement | 1 | 3,007 |
| createForOfStatement | 1 | 2,908 |
| createVariableDeclaration | 1 | 239,240 |
| createVariableDeclarationList | 1 | 232,784 |

### Every transformer member

| File / local | Exact parameters → return | Retained direct / value references | Calls / lane |
| --- | --- | ---: | ---: |
| src/compiler/transformers/ts.ts: emitHelpers | ` → EmitHelperFactory` | 4 / 0 | 666 |
| src/compiler/transformers/ts.ts: startLexicalEnvironment | ` → void` | 2 / 0 | 6,788 |
| src/compiler/transformers/ts.ts: resumeLexicalEnvironment | ` → void` | 1 / 0 | 1,214 |
| src/compiler/transformers/ts.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 3 / 0 | 8,002 |
| src/compiler/transformers/ts.ts: hoistVariableDeclaration | `node: Identifier → void` | 1 / 0 | 124 |
| src/compiler/transformers/classFields.ts: emitHelpers | ` → EmitHelperFactory` | 7 / 0 | 1,891 |
| src/compiler/transformers/classFields.ts: hoistVariableDeclaration | `node: Identifier → void` | 2 / 16 | 3,402 |
| src/compiler/transformers/classFields.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 2 / 0 | 3,555 |
| src/compiler/transformers/classFields.ts: startLexicalEnvironment | ` → void` | 1 / 0 | 974 |
| src/compiler/transformers/classFields.ts: resumeLexicalEnvironment | ` → void` | 1 / 0 | 2,581 |
| src/compiler/transformers/classFields.ts: addBlockScopedVariable | `node: Identifier → void` | 2 / 1 | 20 |
| src/compiler/transformers/typeSerializer.ts: hoistVariableDeclaration | `node: Identifier → void` | 0 / 2 | 38 |
| src/compiler/transformers/legacyDecorators.ts: emitHelpers | ` → EmitHelperFactory` | 3 / 0 | 1,583 |
| src/compiler/transformers/legacyDecorators.ts: hoistVariableDeclaration | `node: Identifier → void` | 1 / 0 | 44 |
| src/compiler/transformers/esDecorators.ts: emitHelpers | ` → EmitHelperFactory` | 11 / 0 | 3,162 |
| src/compiler/transformers/esDecorators.ts: startLexicalEnvironment | ` → void` | 2 / 0 | 1,503 |
| src/compiler/transformers/esDecorators.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 2 / 0 | 1,503 |
| src/compiler/transformers/esDecorators.ts: hoistVariableDeclaration | `node: Identifier → void` | 1 / 7 | 301 |
| src/compiler/transformers/es2017.ts: emitHelpers | ` → EmitHelperFactory` | 2 / 0 | 2,513 |
| src/compiler/transformers/es2017.ts: resumeLexicalEnvironment | ` → void` | 1 / 0 | 2,513 |
| src/compiler/transformers/es2017.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 1 / 0 | 2,513 |
| src/compiler/transformers/es2017.ts: hoistVariableDeclaration | `node: Identifier → void` | 1 / 0 | 106 |
| src/compiler/transformers/es2018.ts: emitHelpers | ` → EmitHelperFactory` | 9 / 0 | 2,215 |
| src/compiler/transformers/es2018.ts: resumeLexicalEnvironment | ` → void` | 2 / 0 | 26,640 |
| src/compiler/transformers/es2018.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 2 / 0 | 26,640 |
| src/compiler/transformers/es2018.ts: hoistVariableDeclaration | `node: Identifier → void` | 2 / 2 | 356 |
| src/compiler/transformers/es2020.ts: hoistVariableDeclaration | `node: Identifier → void` | 0 / 4 | 432 |
| src/compiler/transformers/es2021.ts: hoistVariableDeclaration | `node: Identifier → void` | 0 / 2 | 36 |
| src/compiler/transformers/esnext.ts: emitHelpers | ` → EmitHelperFactory` | 3 / 0 | 2,081 |
| src/compiler/transformers/esnext.ts: hoistVariableDeclaration | `node: Identifier → void` | 2 / 0 | 626 |
| src/compiler/transformers/esnext.ts: startLexicalEnvironment | ` → void` | 1 / 0 | 340 |
| src/compiler/transformers/esnext.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 1 / 0 | 340 |
| src/compiler/transformers/jsx.ts: emitHelpers | ` → EmitHelperFactory` | 1 / 0 | 148 |
| src/compiler/transformers/es2016.ts: hoistVariableDeclaration | `node: Identifier → void` | 0 / 3 | 228 |
| src/compiler/transformers/es2015.ts: emitHelpers | ` → EmitHelperFactory` | 4 / 0 | 779 |
| src/compiler/transformers/es2015.ts: startLexicalEnvironment | ` → void` | 3 / 0 | 7,183 |
| src/compiler/transformers/es2015.ts: resumeLexicalEnvironment | ` → void` | 3 / 0 | 10,368 |
| src/compiler/transformers/es2015.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 6 / 0 | 17,551 |
| src/compiler/transformers/es2015.ts: hoistVariableDeclaration | `node: Identifier → void` | 2 / 3 | 431 |
| src/compiler/transformers/generators.ts: emitHelpers | ` → EmitHelperFactory` | 2 / 0 | 1,029 |
| src/compiler/transformers/generators.ts: resumeLexicalEnvironment | ` → void` | 1 / 0 | 1,011 |
| src/compiler/transformers/generators.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 1 / 0 | 1,011 |
| src/compiler/transformers/generators.ts: hoistFunctionDeclaration | `node: FunctionDeclaration → void` | 1 / 0 | 6 |
| src/compiler/transformers/generators.ts: hoistVariableDeclaration | `node: Identifier → void` | 8 / 3 | 1,085 |
| src/compiler/transformers/module/module.ts: emitHelpers | ` → EmitHelperFactory` | 9 / 0 | 2,845 |
| src/compiler/transformers/module/module.ts: startLexicalEnvironment | ` → void` | 2 / 0 | 11,206 |
| src/compiler/transformers/module/module.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 2 / 0 | 11,206 |
| src/compiler/transformers/module/module.ts: hoistVariableDeclaration | `node: Identifier → void` | 0 / 2 | 24 |
| src/compiler/transformers/module/system.ts: startLexicalEnvironment | ` → void` | 1 / 0 | 632 |
| src/compiler/transformers/module/system.ts: endLexicalEnvironment | ` → Statement[] \| undefined` | 1 / 0 | 632 |
| src/compiler/transformers/module/system.ts: hoistVariableDeclaration | `node: Identifier → void` | 5 / 1 | 1,229 |
| src/compiler/transformers/module/esnextAnd2015.ts: emitHelpers | ` → EmitHelperFactory` | 2 / 0 | 15,905 |

### Every printer alias

| Local | Default | Reads / lane | Census at local |
| --- | --- | ---: | --- |
| hasGlobalName | none | 73,916 | no finding |
| onEmitNode | noEmitNotification | 73,916 | no finding |
| isEmitNotificationEnabled | none | 73,916 | no finding |
| substituteNode | noEmitSubstitution | 73,916 | no finding |
| onBeforeEmitNode | none | 73,916 | no finding |
| onAfterEmitNode | none | 73,916 | no finding |
| onBeforeEmitNodeArray | none | 73,916 | no finding |
| onAfterEmitNodeArray | none | 73,916 | no finding |
| onBeforeEmitToken | none | 73,916 | no finding |
| onAfterEmitToken | none | 73,916 | no finding |

## Full single-file censuses

All 38 runs select one exact file with LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, and a 3,600-second per-file limit. Measurement jobs do not overlap. Each accepted run exits zero, has one selected-file record, zero checker diagnostics, and a relation trailer. The overlay’s historical “checker-rejected program” label is fixed text; the actual checker_rejected header is false. No usable IR or ordinary output program can escape the guarded tool.

| File | Main findings | Prototype findings | Added / changed | Removed / replaced | Main / prototype seconds | Main / prototype peak summed RSS KiB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| src/compiler/parser.ts | 3184 | 3182 | 23 | 25 | 691.907 / 665.501 | 2,156,092 / 2,244,936 |
| src/compiler/transformers/ts.ts | 678 | 678 | 0 | 0 | 146.109 / 141.147 | 1,137,276 / 1,177,680 |
| src/compiler/transformers/classFields.ts | 639 | 639 | 0 | 0 | 143.486 / 143.546 | 1,140,920 / 1,131,048 |
| src/compiler/transformers/typeSerializer.ts | 97 | 95 | 0 | 2 | 50.6 / 49.06 | 993,560 / 997,964 |
| src/compiler/transformers/legacyDecorators.ts | 300 | 298 | 0 | 2 | 87.943 / 89.193 | 1,012,732 / 1,049,836 |
| src/compiler/transformers/esDecorators.ts | 572 | 572 | 0 | 0 | 111.682 / 105.88 | 1,058,760 / 1,086,980 |
| src/compiler/transformers/es2017.ts | 392 | 390 | 0 | 2 | 110.523 / 106.636 | 1,126,440 / 1,134,848 |
| src/compiler/transformers/es2018.ts | 497 | 497 | 0 | 0 | 115.344 / 117.06 | 1,127,932 / 1,134,616 |
| src/compiler/transformers/es2020.ts | 152 | 150 | 0 | 2 | 50.24 / 49.845 | 941,972 / 953,228 |
| src/compiler/transformers/es2021.ts | 27 | 25 | 0 | 2 | 22.0 / 22.611 | 879,056 / 895,152 |
| src/compiler/transformers/esnext.ts | 256 | 254 | 0 | 2 | 74.622 / 77.89 | 1,052,028 / 1,035,544 |
| src/compiler/transformers/jsx.ts | 211 | 209 | 0 | 2 | 72.206 / 78.867 | 1,039,464 / 1,054,896 |
| src/compiler/transformers/es2016.ts | 57 | 55 | 0 | 2 | 26.568 / 29.061 | 899,388 / 914,112 |
| src/compiler/transformers/es2015.ts | 1118 | 1118 | 0 | 0 | 201.207 / 205.496 | 1,351,540 / 1,378,376 |
| src/compiler/transformers/generators.ts | 528 | 528 | 0 | 0 | 115.938 / 116.798 | 1,141,716 / 1,145,608 |
| src/compiler/transformers/module/module.ts | 411 | 411 | 0 | 0 | 94.748 / 96.348 | 1,135,692 / 1,124,744 |
| src/compiler/transformers/module/system.ts | 419 | 417 | 0 | 2 | 138.973 / 141.39 | 1,282,420 / 1,318,800 |
| src/compiler/transformers/module/esnextAnd2015.ts | 122 | 120 | 0 | 2 | 73.241 / 64.184 | 1,026,564 / 1,051,476 |
| src/compiler/emitter.ts | 1455 | 1453 | 0 | 2 | 724.741 / 722.635 | 1,613,280 / 1,643,548 |

Across all 19 pairs, 23 parser records are added/changed and 47 records are removed/replaced: net 24 fewer. The parser audit finds zero new unit/phase/kind/reason combinations; its 23 changed records have changed site/location details at rewritten expressions. The other 18 pairs have zero added/changed records.

**Every new/changed and removed/replaced finding, in full, is in [FINDINGS.md](FINDINGS.md).** Unchanged source lines are mapped to baseline positions before the multiset comparison, so line-number shifts alone are not called new findings. Full raw offsets and fields are preserved in the appendix.

The ten transformer sites whose binding checks are reachable are typeSerializer, legacyDecorators, es2016, es2017, es2020, es2021, esnext, jsx, module/system, and module/esnextAnd2015. Seven sites remain hidden behind earlier signature/prologue stops: ts, classFields, esDecorators, es2018, es2015, generators, and module/module. Zero method-destructuring findings at a hidden site is not proof that its body lowered. The stock AST audit independently confirms all 19 public method-binding sites are removed, while the ten internal/private sites remain.

The original emitter binding is behind var at 1215:5, so its main census cannot diagnose the destructured methods. The prototype replaces that statement with the readonly view and ten const aliases; its census reaches those locals with no findings there. Later var and other lowering stops remain; this does not certify native compilation of createPrinter.

## Checks and workload

Peak summed process-tree RSS: counting lane 11,339,000 KiB; pristine prototype lane 9,769,792 KiB. Sampling interval: 0.2 seconds. These instrumented timings are not a dispatch benchmark.

- Stock TypeScript 6.0.3: zero diagnostics over 79 explicit compiler roots. Adamic types: exit zero over the same 79 roots.

- Normal hereby dts: typescript.d.ts (593,536 bytes) and tsserverlibrary.d.ts (866 bytes) byte-identical to the adapted baseline. SHA-256 values are in [split-declarations.json](evidence/split-declarations.json).

- Uninstrumented prototype whole lane: 106,366 passing, one failing, zero pending. Sole failure: unittests:: Public APIs for typescript.d.ts should be acknowledged when they change; sole baseline diff api/typescript.d.ts. The separate counting lanes have the same sanctioned result. All use all runners, eight workers, no test filter and the unchanged 40,000 ms hook timeout.

- All 19,585 tests/cases files and 60,931 reference files are byte-identical in the counting copy, including nested node_modules and built fixtures; inventory hashes are recorded. Only root dependencies/build outputs/history are omitted when making a counting copy.

## Provenance, failed attempts and raw artifacts

Base: Adamic 8e3a784aa5a601f57342c07b68f145fe56252c21 in a fresh private detached worktree before the report branch. Certified comparison evidence: typescript/destructure-price cfee4cd3e20a084eef0aecb747f16d311409a6dc. TypeScript v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8. Recursive shallow submodules are cohere 5360753e19e4417675a90e92a1823bea66b13918 and compiler fork d92d9bfee114c80be2c375d72edae966176e3a4f. Node v24.19.0; Go go1.27.1. Temporary directories and Go caches are scratch. No adaptation directory, compiler-rule change, pull request or submission is part of this measurement.

An exploratory scan during incomplete adaptation was discarded. An early census with missing type-only imports was stopped and discarded. Counting-build attempts with an incorrect namespace path, lost equality narrowing, or erased arrow local types never reached tests and contribute no snapshots. A workspace reconnect interrupted a checker process before a result was written; completed census records were retained and only unfinished work resumed. The successful counting recipe preserves guards and exact arrow local types; its hashed source is retained alongside attribution evidence.

Large raw traces, implementation source strings, source patches, generated Go overlays and process logs stay in private scratch. Sizes and SHA-256 are in [split-raw-manifest.json](evidence/split-raw-manifest.json). All newly added Go code would use .go.txt under stage3; this report adds no Go files. Reproduce the AST rewrite with split-codemod.cjs using the certified heat.json, then use the sequential census/check runners. These are private experiments, not an adaptation.
