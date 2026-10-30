# Every new or growing reason

| Reason | Before | After |
|---|---:|---:|
| Array.isArray on a tuple or an erased object/any/unknown view | 0 | 1 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 2 | 3 |
| a call through ?. (an optional call) | 24 | 25 |
| a function returning T &#124; undefined | 2 | 3 |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 4 | 5 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 2 |
| a union or optional field read in a program with record storage | 69 | 73 |
| a value of type NodeArray<ClassElement> | 3 | 5 |
| a value of type NodeArray<Modifier> &#124; NodeArray<ModifierLike> &#124; undefined | 0 | 1 |
| a value of type NodeArray<ObjectLiteralElementLike> | 0 | 1 |
| a value of type NodeArray<ObjectLiteralElementLike> &#124; undefined | 0 | 1 |
| a value of type NodeArray<Statement> | 4 | 5 |
| a value of type ResolvedConfigFileName | 0 | 1 |
| an unproven relation from Node &#124; undefined to T &#124; undefined: the source is not assignable to the target | 0 | 1 |
| assigning a field of a value | 3 | 4 |
| optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 0 | 1 |
| optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 1 | 2 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 3 | 32 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 0 | 7 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 0 | 6 |
| optional property id in BinaryExpression absent from structural source never, which can hide fields | 0 | 1 |
| optional property id in Expression absent from structural source never, which can hide fields | 0 | 5 |
| optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 0 | 5 |
| optional property id in Identifier absent from structural source never, which can hide fields | 0 | 4 |
| optional property id in Node absent from structural source never, which can hide fields | 0 | 10 |
| optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 0 | 1 |
| optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 0 | 1 |
| optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 0 | 2 |
| optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 0 | 1 |
| optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 0 | 1 |
| optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 0 | 1 |
| optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 0 | 2 |
| optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 1 | 2 |
| overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] &#124; undefined | 0 | 1 |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 0 | 1 |

## Examples

- Array.isArray on a tuple or an erased object/any/unknown view: `src/compiler/utilities.ts:8765:54`
- a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions: `src/compiler/moduleNameResolver.ts:1284:31`
- a call through ?. (an optional call): `src/compiler/builder.ts:985:5`
- a function returning T | undefined: `src/compiler/parser.ts:1249:17`
- a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation: `src/compiler/transformers/es2017.ts:1106:25`
- a template interpolating an object, an array, a map, a function or undefined: `src/compiler/core.ts:1786:59`
- a union or optional field read in a program with record storage: `src/compiler/commandLineParser.ts:3275:21`, `src/compiler/utilities.ts:8632:29`, `src/compiler/utilities.ts:8702:29`
- a value of type NodeArray<ClassElement>: `src/compiler/transformers/classThis.ts:91:47`, `src/compiler/transformers/namedEvaluation.ts:157:50`
- a value of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined: `src/compiler/parser.ts:496:17`
- a value of type NodeArray<ObjectLiteralElementLike>: `src/compiler/utilities.ts:3954:20`
- a value of type NodeArray<ObjectLiteralElementLike> | undefined: `src/compiler/utilities.ts:3096:20`
- a value of type NodeArray<Statement>: `src/compiler/parser.ts:472:20`
- a value of type ResolvedConfigFileName: `src/compiler/program.ts:1290:19`
- an unproven relation from Node | undefined to T | undefined: the source is not assignable to the target: `src/compiler/utilitiesPublic.ts:774:16`
- assigning a field of a value: `src/compiler/core.ts:307:5`
- optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields: `src/compiler/binder.ts:3426:124`
- optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields: `src/compiler/transformers/namedEvaluation.ts:288:9`
- optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields: `src/compiler/emitter.ts:1444:66`, `src/compiler/emitter.ts:3961:112`, `src/compiler/emitter.ts:3961:180`
- optional property id in ArrayBindingPattern absent from structural source never, which can hide fields: `src/compiler/binder.ts:738:67`, `src/compiler/transformers/namedEvaluation.ts:314:9`, `src/compiler/transformers/namedEvaluation.ts:315:9`
- optional property id in BigIntLiteral absent from structural source never, which can hide fields: `src/compiler/factory/utilities.ts:363:96`, `src/compiler/factory/utilities.ts:368:58`, `src/compiler/moduleSpecifiers.ts:881:141`
- optional property id in BinaryExpression absent from structural source never, which can hide fields: `src/compiler/transformers/namedEvaluation.ts:437:9`
- optional property id in Expression absent from structural source never, which can hide fields: `src/compiler/binder.ts:3293:68`, `src/compiler/binder.ts:3313:64`, `src/compiler/binder.ts:3427:55`
- optional property id in ExternalModuleReference absent from structural source never, which can hide fields: `src/compiler/transformers/module/module.ts:1566:55`, `src/compiler/transformers/module/module.ts:1588:63`, `src/compiler/transformers/module/module.ts:1618:73`
- optional property id in Identifier absent from structural source never, which can hide fields: `src/compiler/transformers/namedEvaluation.ts:311:46`, `src/compiler/transformers/namedEvaluation.ts:341:46`, `src/compiler/transformers/namedEvaluation.ts:373:46`
- optional property id in Node absent from structural source never, which can hide fields: `src/compiler/binder.ts:3705:86`, `src/compiler/factory/nodeFactory.ts:7105:28`, `src/compiler/factory/nodeFactory.ts:7302:56`
- optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields: `src/compiler/parser.ts:1854:40`
- optional property modifiers in ExportAssignment absent from structural source never, which can hide fields: `src/compiler/transformers/namedEvaluation.ts:459:9`
- optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields: `src/compiler/utilities.ts:6839:25`, `src/compiler/utilities.ts:6841:27`
- optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields: `src/compiler/transformers/namedEvaluation.ts:396:9`
- optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields: `src/compiler/utilities.ts:6844:27`
- optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields: `src/compiler/emitter.ts:1408:32`
- optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields: `src/compiler/utilities.ts:12192:121`, `src/compiler/utilities.ts:4369:20`
- optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields: `src/compiler/expressionToTypeNode.ts:438:64`
- overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] | undefined: `src/compiler/core.ts:342:1`
- overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor: `src/compiler/visitorPublic.ts:587:119`

## Other measurement events

- SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) 1 → 3; new examples: `src/compiler/tsbuildPublic.ts:295:1`, `src/compiler/watch.ts:843:1`
