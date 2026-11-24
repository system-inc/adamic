# SyntacticResult's first member

Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 22, casts cause 4); Compiler accepts the head after its review and census.

The private alias's first member records `reportFallback: true`. At adapted `expressionToTypeNode.ts:167`, `syntacticResult` constructs and casts one object. Its overloads allow a second argument only when `type` is undefined. Every call that can produce a TypeNode omits that argument, so its default is true. The constructor, cast and two undefined-type members keep their source bytes.

Both tables use fully adapted coordinates from the adapter's `-before` snapshot at Adamic base `26d185e56443f95d24df42d4de7be290bd6a8091`. Calls with two arguments always pass undefined for the first argument.

| Line | Arguments |
| --- | --- |
| 169 | `undefined, false` |
| 170 | `undefined, false` |
| 171 | `undefined, true` |
| 774 | `serializeTypeAnnotationOfDeclaration(declaredType, context, node, symbol)` |
| 792 | `serializeTypeAnnotationOfDeclaration(declaredType, context, node, symbol, addUndefined)` |
| 819 | `serializeTypeAnnotationOfDeclaration(declaredType, context, node, symbol, requiresAddingUndefined)` |
| 890 | `serializeExistingTypeNodeWithFallback(type, context, requiresAddingUndefined)` |
| 901 | `createUndefinedTypeNode()` |
| 906 | `addUndefinedIfNeeded(factory.createLiteralTypeNode(factory.createNull()), requiresAddingUndefined, node, context)` |
| 909 | `factory.createKeywordTypeNode(SyntaxKind.AnyKeyword)` |
| 936 | `inferExpressionType(node as ClassExpression, context, /*reportFallback*/ true, requiresAddingUndefined)` |
| 939 | `factory.createKeywordTypeNode(SyntaxKind.StringKeyword)` |
| 974 | `factory.createFunctionTypeNode( typeParameters, parameters, returnType, )` |
| 1000 | `inferExpressionType(arrayLiteral, context, /*reportFallback*/ false, requiresAddingUndefined)` |
| 1057 | `inferExpressionType(objectLiteral, context, /*reportFallback*/ false, requiresAddingUndefined)` |
| 1230 | `addUndefinedIfNeeded(result, requiresAddingUndefined, node, context)` |
| 1273 | `serializeTypeAnnotationOfDeclaration(returnTypeNode, context, fn, symbol)` |
| 1313 | `serializeExistingTypeNode(type, context)` |

Every read is in the false branch of the following condition, so the result's type is undefined when its flag is observed.

| Line | Read | Condition |
| --- | --- | --- |
| 671 | `result.reportFallback` | `result.type !== undefined` |
| 781 | `resultType.reportFallback` | `resultType.type !== undefined` |
| 797 | `resultType.reportFallback` | `resultType.type !== undefined` |
| 828 | `resultType.reportFallback` | `resultType.type !== undefined` |
| 1015 | `expressionType.reportFallback` | `expressionType.type !== undefined` |
| 1102 | `expressionResult.reportFallback` | `expressionResult.type !== undefined` |
| 1278 | `returnType.reportFallback` | `returnType.type !== undefined` |

The adapter pins producer statements, guarded reads and occurrence counts before writing or replaying, and `test.cjs` exercises the clean and drift cases against the adapter's own snapshots.

Rows 12 and 96 keep mutable widening refused. Row 18 governs an untagged downcast's checked-view road. Row 72 describes fresh literals. This patch establishes the target's truthful type before a separate compiler ruling can change how the cast is classified.
