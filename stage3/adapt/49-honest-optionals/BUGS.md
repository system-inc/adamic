# TypeScript 6.0.3 nonliteral nullish assertion bug log

All 22 expressions below evaluated to `undefined` on Node. Stock tsc continued and produced the recorded acceptance output. Locations use the pinned stock source, not adapted line numbers. Input sets resolve in [input-sets.json](../../scouts/step09/nonnull/input-sets.json).

| Stock file:line:column | Expression | Node value | Input set / first input | Adaptation 49 |
| --- | --- | --- | --- | --- |
| src/compiler/checker.ts:14118:44 | `source.typeParameters!` | `undefined` | INPUTS31 / `003_commentOnClassAccessor1` | repaired |
| src/compiler/checker.ts:16437:66 | `signature.declaration!` | `undefined` | INPUTS41 / `029_enumLiteralUnionNotWidened` | repaired |
| src/compiler/checker.ts:19889:118 | `mapper!` | `undefined` | INPUTS53 / `002_conditionalEqualityTestingNullability` | repaired |
| src/compiler/checker.ts:20277:50 | `symbol!` | `undefined` | ALL301 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |
| src/compiler/checker.ts:28392:41 | `node.initializer!` | `undefined` | INPUTS80 / `006_arrayDestructuringInSwitch1` | repaired |
| src/compiler/checker.ts:32024:19 | `contextFlags!` | `undefined` | INPUTS103 / `014_keyofObjectWithGlobalSymbolIncluded` | blocked |
| src/compiler/checker.ts:32724:37 | `contextFlags!` | `undefined` | INPUTS104 / `010_nestedTypeVariableInfersLiteral` | blocked |
| src/compiler/checker.ts:38571:41 | `flags!` | `undefined` | INPUTS124 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |
| src/compiler/checker.ts:38572:35 | `flags!` | `undefined` | INPUTS124 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |
| src/compiler/checker.ts:43525:30 | `getTypeOfPropertyOfType(type, "then" as __String)!` | `undefined` | INPUTS140 / `030_asyncYieldStarContextualType` | repaired |
| src/compiler/checker.ts:46972:59 | `type.localTypeParameters!` | `undefined` | INPUTS148 / `019_superHasMethodsFromMergedInterface` | repaired |
| src/compiler/checker.ts:5553:61 | `symbol!` | `undefined` | ALL301 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |
| src/compiler/checker.ts:5568:62 | `symbol!` | `undefined` | ALL301 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |
| src/compiler/checker.ts:5790:30 | `symbols.get(symbol!.escapedName)!` | `undefined` | INPUTS165 / `088_baseConstraintOfDecorator` | repaired |
| src/compiler/checker.ts:6191:44 | `meaning!` | `undefined` | INPUTS168 / `062_interfaceDeclaration1` | blocked |
| src/compiler/core.ts:1244:26 | `initial!` | `undefined` | INPUTS177 / `045_discriminantUsingEvaluatableTemplateExpression` | repaired |
| src/compiler/core.ts:2474:16 | `lastResult!` | `undefined` | ALL301 / `001_varianceCantBeStrictWhileStructureIsnt` | repaired |
| src/compiler/emitter.ts:1417:18 | `_writer!` | `undefined` | INPUTS170 / `015_decoratorReferences` | repaired |
| src/compiler/emitter.ts:4954:16 | `pos!` | `undefined` | INPUTS181 / `090_constraintWithIndexedAccess` | blocked |
| src/compiler/emitter.ts:4954:34 | `pos!` | `undefined` | INPUTS181 / `090_constraintWithIndexedAccess` | blocked |
| src/compiler/scanner.ts:1034:16 | `textInitial!` | `undefined` | ALL301 / `001_varianceCantBeStrictWhileStructureIsnt` | repaired |
| src/compiler/utilities.ts:11868:13 | `result.isReferenced!` | `undefined` | INPUTS192 / `001_varianceCantBeStrictWhileStructureIsnt` | blocked |

## Disposition

- `src/compiler/checker.ts:14118:44`: concatenate already accepts an undefined list; remove the false unwrap.
- `src/compiler/checker.ts:16437:66`: Widen private getSignatureOfTypeTag.node; isInJSFile rejects undefined before the moved checked read.
- `src/compiler/checker.ts:19889:118`: Widen instantiateTypes.mapper and parameterize instantiateList by its mapper type M; instantiateType already accepts undefined.
- `src/compiler/checker.ts:20277:50`: Undefined is assigned through createTypeWithSymbol to public Type.symbol: Symbol. An honest owner widening changes the public API snapshot; keeping the required public type would retain the lie.
- `src/compiler/checker.ts:28392:41`: Widen private getTypeWithDefault.defaultExpression; its existing conditional returns the original type for undefined.
- `src/compiler/checker.ts:32024:19`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/checker.ts:32724:37`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/checker.ts:38571:41`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/checker.ts:38572:35`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/checker.ts:43525:30`: Infer thenFunction as Type | undefined; existing optional handling accepts an absent then property.
- `src/compiler/checker.ts:46972:59`: Widen private areTypeParametersIdentical.targetParameters; existing length/count guard rejects nonempty source parameters for an absent target before checked indexing.
- `src/compiler/checker.ts:5553:61`: Undefined is assigned through createTypeWithSymbol to public Type.symbol: Symbol. An honest owner widening changes the public API snapshot; keeping the required public type would retain the lie.
- `src/compiler/checker.ts:5568:62`: Undefined is assigned through createTypeWithSymbol to public Type.symbol: Symbol. An honest owner widening changes the public API snapshot; keeping the required public type would retain the lie.
- `src/compiler/checker.ts:5790:30`: Widen private isAccessible.symbolFromSymbolTable; the existing equality test rejects an absent entry before checked member reads.
- `src/compiler/checker.ts:6191:44`: The private builder chain forwards meaning through symbolToName/symbolToExpression, lookupSymbolChain and getSymbolChain into needsQualification, which computes flags & meaning. Honest optional handling at that bit operation changes JavaScript bytes. Public required parameter types can remain narrower than the private implementation; they are not by themselves a proof that this repair must change the API.
- `src/compiler/core.ts:1244:26`: Widen reduceLeft implementation accumulator and callback accumulator parameter; existing generic overload supplies its truthful optional accumulator type.
- `src/compiler/core.ts:2474:16`: Widen or implementation lastResult and general return to U | undefined; a nonempty tuple overload retains precision for fixed nonempty calls.
- `src/compiler/emitter.ts:1417:18`: Own printer writer as EmitTextWriter | undefined; preserve optional saved writers and guards, check active emitting reads after entrypoints install an output writer.
- `src/compiler/emitter.ts:4954:16`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/emitter.ts:4954:34`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
- `src/compiler/scanner.ts:1034:16`: Own scanner text as string | undefined; preserve optional setText input and check reads after setText normalizes the text to a string before methods are returned.
- `src/compiler/utilities.ts:11868:13`: Node deliberately coerces undefined in a numeric operation: bitwise flags treat it as zero; position arithmetic/comparison can produce NaN. An honest optional operand needs explicit runtime coercion/handling, which changes emitted JavaScript bytes.
