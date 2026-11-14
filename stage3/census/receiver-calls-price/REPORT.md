# Receiver calls: stopped at the binding-use gate

**Rewrite size: 0 uses, 0 lines changed in every compiler file.** Writes after creation: **unmeasured**; identity/`.length` reads: **unmeasured**; implementation `this` reads: **unmeasured**; added property reads per lane: **unmeasured**, against the certified hybrid’s **30,851,760 calls**. Printer readonly-view/default-local census: **not run**. The per-use fallback was not attempted. These are missing measurements, never zero counts.

The symbol-resolved AST preflight stopped with **50 findings** across all **19 public sites**. No prototype compiler was emitted. The brief explicitly requires non-direct-callee uses and shadowing to fail loud and permits stopping with a reason. Building a forwarding callback, retaining an unsafe method local, or renaming/capturing a receiver would change that gate and was not attempted.

The artifact is the stopped codemod’s preflight, **not a completed rewrite implementation**. Its emission/default-view phase is unimplemented. It resolves references by binding symbol, checks every reference of each non-defaulted method local, reports all failures before returning exit 2, and checks receiver resolution at each use. Property bindings (including `factory`) and defaulted method locals are retained for the planned readonly-view exception. A write/reassignment is a non-call reference and fails the same gate.

Examples include `factory.createTempVariable(hoistVariableDeclaration)` (callback value), `isFileLevelUniqueName(currentSourceFile, name, hasGlobalName)` (callback value), and `!isEmitNotificationEnabled` (presence test). Two uses of the `getEmitHelperFactory` alias in esDecorators resolve `context` to a nearer local, so spelling the receiver call there would address the wrong object.

## Per-file inventory

Counts below are eligible direct-call syntax occurrences in the original AST, **not rewrites or runtime calls**. Defaulted members are excluded from these candidate counts. Each listed file has zero rewritten uses and zero changed source lines.

| File | Direct-call candidates | Stop findings |
| --- | ---: | ---: |
| src/compiler/parser.ts | 46 | 0 |
| src/compiler/transformers/ts.ts | 11 | 0 |
| src/compiler/transformers/classFields.ts | 15 | 17 |
| src/compiler/transformers/typeSerializer.ts | 0 | 2 |
| src/compiler/transformers/legacyDecorators.ts | 4 | 0 |
| src/compiler/transformers/esDecorators.ts | 16 | 9 |
| src/compiler/transformers/es2017.ts | 5 | 0 |
| src/compiler/transformers/es2018.ts | 15 | 2 |
| src/compiler/transformers/es2020.ts | 0 | 4 |
| src/compiler/transformers/es2021.ts | 0 | 2 |
| src/compiler/transformers/esnext.ts | 7 | 0 |
| src/compiler/transformers/jsx.ts | 1 | 0 |
| src/compiler/transformers/es2016.ts | 0 | 3 |
| src/compiler/transformers/es2015.ts | 18 | 3 |
| src/compiler/transformers/generators.ts | 13 | 3 |
| src/compiler/transformers/module/module.ts | 13 | 2 |
| src/compiler/transformers/module/system.ts | 7 | 1 |
| src/compiler/transformers/module/esnextAnd2015.ts | 2 | 0 |
| src/compiler/emitter.ts | 1 | 2 |

Per-member counts, including zero-use method bindings, and full use expressions are in `evidence/preflight.json`.

## Every stop finding

- `src/compiler/transformers/classFields.ts:931:57` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1231:118` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1232:110` — hoistVariableDeclaration: local used other than as a direct callee. `expandPreOrPostfixIncrementOrDecrementExpression(factory, node, expression, hoistVariableDeclaration, temp)`
- `src/compiler/transformers/classFields.ts:1283:69` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1291:89` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1292:114` — hoistVariableDeclaration: local used other than as a direct callee. `expandPreOrPostfixIncrementOrDecrementExpression(factory, node, expression, hoistVariableDeclaration, temp)`
- `src/compiler/transformers/classFields.ts:1333:59` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1346:84` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(node.expression, hoistVariableDeclaration, languageVersion)`
- `src/compiler/transformers/classFields.ts:1394:77` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(node.tag, hoistVariableDeclaration, languageVersion)`
- `src/compiler/transformers/classFields.ts:1588:73` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1606:89` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:1782:53` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration, /*reservedInNestedScopes*/ true)`
- `src/compiler/transformers/classFields.ts:1880:57` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration, /*reservedInNestedScopes*/ true)`
- `src/compiler/transformers/classFields.ts:1984:103` — hoistVariableDeclaration: local used other than as a direct callee. `requiresBlockScopedVar ? addBlockScopedVariable : hoistVariableDeclaration`
- `src/compiler/transformers/classFields.ts:2161:57` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/classFields.ts:2995:51` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration, /*reservedInNestedScopes*/ true)`
- `src/compiler/transformers/classFields.ts:1984:78` — addBlockScopedVariable: local used other than as a direct callee. `requiresBlockScopedVar ? addBlockScopedVariable : hoistVariableDeclaration`
- `src/compiler/transformers/typeSerializer.ts:500:57` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/typeSerializer.ts:581:49` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1356:46` — getEmitHelperFactory: receiver is shadowed.
- `src/compiler/transformers/esDecorators.ts:1379:46` — getEmitHelperFactory: receiver is shadowed.
- `src/compiler/transformers/esDecorators.ts:1596:61` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1806:69` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1826:85` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1873:65` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1902:85` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/esDecorators.ts:1903:110` — hoistVariableDeclaration: local used other than as a direct callee. `expandPreOrPostfixIncrementOrDecrementExpression(factory, node, expression, hoistVariableDeclaration, temp)`
- `src/compiler/transformers/esDecorators.ts:2267:79` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(expression, hoistVariableDeclaration, languageVersion, /*cacheIdentifiers*/ true)`
- `src/compiler/transformers/es2018.ts:787:50` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2018.ts:832:49` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2020.ts:121:54` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2020.ts:173:55` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2020.ts:185:66` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2020.ts:247:48` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2021.ts:60:44` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2021.ts:79:48` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2016.ts:67:63` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2016.ts:68:71` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2016.ts:86:63` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2015.ts:3334:49` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/es2015.ts:4530:84` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(node.expression, hoistVariableDeclaration)`
- `src/compiler/transformers/es2015.ts:4610:132` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(factory.createPropertyAccessExpression(node.expression, "bind"), hoistVariableDeclaration)`
- `src/compiler/transformers/generators.ts:1227:84` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(node.expression, hoistVariableDeclaration, languageVersion, /*cacheIdentifiers*/ true)`
- `src/compiler/transformers/generators.ts:1257:132` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createCallBinding(factory.createPropertyAccessExpression(node.expression, "bind"), hoistVariableDeclaration)`
- `src/compiler/transformers/generators.ts:2093:49` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/module/module.ts:1167:59` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/module/module.ts:1258:53` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/transformers/module/system.ts:1717:59` — hoistVariableDeclaration: local used other than as a direct callee. `factory.createTempVariable(hoistVariableDeclaration)`
- `src/compiler/emitter.ts:5521:83` — hasGlobalName: local used other than as a direct callee. `isFileLevelUniqueName(currentSourceFile, name, hasGlobalName)`
- `src/compiler/emitter.ts:1502:60` — isEmitNotificationEnabled: local used other than as a direct callee. `!isEmitNotificationEnabled`

## Census coverage and unrun checks

No new full census was run on main or a prototype; there is no changed compiler to compare. No claim is made that the destructuring refusals are gone. Stock tsc, `adamic types`, declaration equivalence and the unfiltered lane were not run on a prototype. No counting copy or positive counter controls were built.

For orientation only, the certified cfee4cd3 inventory says its census reaches these 10 transformer sites: typeSerializer, legacyDecorators, es2016, es2017, es2020, es2021, esnext, jsx, module/system, module/esnextAnd2015. Its seven earlier signature/prologue stops hide ts, classFields, esDecorators, es2018, es2015, generators, and module/module. This is inherited evidence, not a fresh census result.

Likewise emitter’s `var` gate at 1215:5 prevents its original census from reaching the method-destructuring check. It cannot establish whether the requested readonly property view and default locals lower successfully. No inference from sys.ts:1385 substitutes for that experiment.

## Reproduction and provenance

Base: `8e3a784aa5a601f57342c07b68f145fe56252c21`, fresh detached worktree `/workspace/scratch/receiver-calls-price`. No adaptation directory or compiler-rule edits. Submodules initialized recursively/shallow: cohere `5360753e19e4417675a90e92a1823bea66b13918`, compiler fork `d92d9bfee114c80be2c375d72edae966176e3a4f`. Upstream cloned at v6.0.3, verified `050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci` completed in stage3/api; AST API 6.0.3.

`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/receiver-typescript`; `TMPDIR` and `GOTMPDIR` use `/workspace/scratch/receiver-tmp`; `GOCACHE=/workspace/scratch/receiver-gocache`. GNU time was unavailable; build accounting uses Python resource, definitive preflight uses sampled process-tree RSS. See setup.json and preflight-measure.json.

An exploratory preflight ran before adaptation finished and is discarded. Only the rerun after the completed build is published. No overlapping heavy work occurred in the definitive preflight.

```sh
git show cfee4cd3:stage3/census/destructure-price/evidence/heat.json > /workspace/scratch/receiver-heat.json
node stage3/census/receiver-calls-price/codemod.cjs /workspace/scratch/receiver-baseline /workspace/scratch/receiver-heat.json stage3/census/receiver-calls-price/evidence/preflight.json
```

The final program reads 79 compiler roots. The exact pinned heat inventory is used to identify all 19 bindings; symbol identity determines references. No regular expression matches or rewrites TypeScript syntax.
