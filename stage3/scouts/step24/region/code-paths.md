# Observed creation paths

Exact constructor expression sites are in sites.json. Function-definition sites below identify the deepest parser creation path. Every full dynamic parser/factory path and every source-mapped synthetic caller stack is in paths.json.gz. “Detached” here is the requested literal union: no parent OR not graph-reachable from a SourceFile. It includes legitimate parentless SourceFile roots; summary.json also reports the root-exempt union. Counts do not estimate bytes or membership.

## Parser and factory paths with detached candidates

| Origin | Creation function site | Scanner detached | Acceptance detached | Unreachable across both | Kinds |
|---|---|---:|---:|---:|---|
| parser | src/compiler/parser.ts:2648:createIdentifier | 10109 | 1197756 | 1207865 | Identifier |
| parser | src/compiler/parser.ts:3797:parseTypeReference | 2528 | 29261 | 31789 | TypeReference |
| parser | src/compiler/parser.ts:2553:parseTokenNode | 61 | 16780 | 16841 | AnyKeyword, AsteriskToken, DotDotDotToken, EndOfFileToken, FalseKeyword, NumberKeyword, PlusToken, QuestionToken, StringKeyword, SymbolKeyword, ThisKeyword, TrueKeyword, UndefinedKeyword, UnknownKeyword |
| parser | src/compiler/parser.ts:1978:createSourceFile | 81 | 7167 | 301 | SourceFile |
| parser | src/compiler/parser.ts:3760:parseLiteralLikeNode | 1541 | 3686 | 5227 | FirstLiteralToken, StringLiteral |
| parser | src/compiler/parser.ts:6699:parseObjectLiteralElement | 0 | 2595 | 2595 | PropertyAssignment |
| parser | src/compiler/parser.ts:4528:parseLiteralTypeNode | 1541 | 157 | 1698 | LiteralType, PrefixUnaryExpression |
| parser | src/compiler/parser.ts:4508:parseFunctionOrConstructorType | 1340 | 0 | 1340 | FunctionType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1326:createIdentifier | 0 | 922 | 922 | Identifier |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2280:createKeywordTypeNode | 0 | 917 | 917 | AnyKeyword, BooleanKeyword, NeverKeyword, NumberKeyword, StringKeyword, SymbolKeyword, UndefinedKeyword, UnknownKeyword, VoidKeyword |
| parser | src/compiler/parser.ts:3589:parseEntityName | 778 | 2 | 780 | FirstNode |
| parser | src/compiler/parser.ts:6689:parseArrayLiteralExpression | 0 | 613 | 613 | ArrayLiteralExpression |
| parser | src/compiler/parser.ts:6755:parseObjectLiteralExpression | 0 | 605 | 605 | ObjectLiteralExpression |
| parser | src/compiler/parser.ts:6453:parseMemberExpressionRest | 548 | 26 | 574 | ExpressionWithTypeArguments |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2304:createTypeReferenceNode | 0 | 520 | 520 | Identifier, TypeReference |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:6369:cloneNode | 0 | 302 | 302 | AnyKeyword, ArrayBindingPattern, BindingElement, BooleanKeyword, FunctionType, Identifier, LiteralType, NumberKeyword, Parameter, PropertySignature, ReadonlyKeyword, StringKeyword, StringLiteral, TrueKeyword, TypeLiteral, TypeQuery, TypeReference, VoidKeyword |
| parser | src/compiler/parser.ts:1646:parseJsonText | 0 | 301 | 301 | ExpressionStatement |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2759:createLiteralTypeNode | 0 | 275 | 275 | LiteralType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1259:createStringLiteral | 0 | 125 | 125 | StringLiteral |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1722:createPropertySignature | 0 | 121 | 121 | PropertySignature |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2321:createFunctionTypeNode | 0 | 117 | 117 | FunctionType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2448:createTypeLiteralNode | 0 | 107 | 107 | TypeLiteral |
| parser | src/compiler/parser.ts:4819:parseUnionOrIntersectionType | 8 | 96 | 104 | UnionType |
| parser | src/compiler/parser.ts:4036:parseParameterWorker | 46 | 52 | 98 | Parameter |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2559:createUnionTypeNode | 0 | 94 | 94 | UnionType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1233:createNumericLiteral | 0 | 92 | 92 | FirstLiteralToken |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1645:createParameterDeclaration | 0 | 86 | 86 | Identifier, Parameter |
| parser | src/compiler/parser.ts:7594:parseObjectBindingElement | 67 | 17 | 84 | BindingElement |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1453:createToken | 0 | 73 | 73 | DotDotDotToken, QuestionToken, ReadonlyKeyword |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2714:createIndexedAccessTypeNode | 0 | 71 | 71 | IndexedAccessType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:6218:createSyntheticExpression | 0 | 66 | 66 | SyntheticExpression |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2478:createTupleTypeNode | 0 | 58 | 58 | TupleType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2431:createTypeQueryNode | 0 | 55 | 55 | TypeQuery |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1621:createTypeParameterDeclaration | 0 | 55 | 55 | TypeParameter |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1531:createNull | 0 | 53 | 53 | NullKeyword |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2463:createArrayTypeNode | 0 | 52 | 52 | ArrayType |
| parser | src/compiler/parser.ts:7612:parseObjectBindingPattern | 28 | 21 | 49 | ObjectBindingPattern |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2731:createMappedTypeNode | 0 | 49 | 49 | MappedType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2696:createTypeOperatorNode | 0 | 47 | 47 | TypeOperator |
| parser | src/compiler/parser.ts:4716:parsePostfixTypeOrHigher | 30 | 1 | 31 | ArrayType, IndexedAccessType, JSDocNonNullableType, JSDocNullableType |
| parser | src/compiler/parser.ts:2619:createMissingNode | 27 | 3 | 30 | Identifier |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2160:createCallSignature | 0 | 30 | 30 | CallSignature |
| parser | src/compiler/parser.ts:4479:parseTupleType | 0 | 24 | 24 | TupleType |
| parser | src/compiler/parser.ts:3955:parseTypeParameter | 14 | 2 | 16 | TypeParameter |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2968:createElementAccessExpression | 0 | 16 | 16 | ElementAccessExpression |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1741:updatePropertySignature | 0 | 15 | 15 | PropertySignature |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1536:createTrue | 0 | 13 | 13 | TrueKeyword |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2569:createIntersectionTypeNode | 0 | 13 | 13 | IntersectionType |
| parser | src/compiler/parser.ts:6520:parseCallExpressionRest | 8 | 4 | 12 | CallExpression |
| parser | src/compiler/parser.ts:6415:parsePropertyAccessExpressionRest | 9 | 3 | 12 | PropertyAccessExpression |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2456:updateTypeLiteralNode | 0 | 9 | 9 | TypeLiteral |
| parser | src/compiler/parser.ts:4489:parseParenthesizedType | 8 | 0 | 8 | ParenthesizedType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1541:createFalse | 0 | 8 | 8 | FalseKeyword |
| parser | src/compiler/parser.ts:7980:tryParseModifier | 7 | 0 | 7 | ReadonlyKeyword |
| parser | src/compiler/parser.ts:7620:parseArrayBindingPattern | 7 | 0 | 7 | ArrayBindingPattern |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1580:createQualifiedName | 0 | 7 | 7 | FirstNode |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1526:createThis | 0 | 5 | 5 | ThisKeyword |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2903:createPropertyAccessExpression | 0 | 5 | 5 | PropertyAccessExpression |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2313:updateTypeReferenceNode | 0 | 5 | 5 | TypeReference |
| parser | src/compiler/parser.ts:4752:parseTypeOperator | 4 | 0 | 4 | TypeOperator |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2767:updateLiteralTypeNode | 0 | 4 | 4 | LiteralType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2834:updateBindingElement | 0 | 4 | 4 | BindingElement |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2809:updateArrayBindingPattern | 0 | 4 | 4 | ArrayBindingPattern |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2226:createIndexSignature | 0 | 4 | 4 | IndexSignature |
| parser | src/compiler/parser.ts:6432:parseElementAccessExpressionRest | 3 | 0 | 3 | ElementAccessExpression |
| parser | src/compiler/parser.ts:2733:parseComputedPropertyName | 0 | 3 | 3 | ComputedPropertyName |
| parser | src/compiler/parser.ts:7583:parseArrayBindingElement | 2 | 0 | 2 | BindingElement |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:1680:updateParameterDeclaration | 0 | 2 | 2 | Parameter |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2440:updateTypeQueryNode | 0 | 2 | 2 | TypeQuery |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2363:createConstructorTypeNode | 0 | 2 | 2 | ConstructorType |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:3856:createBlock | 0 | 2 | 2 | Block |
| parser | src/compiler/parser.ts:5693:parsePrefixUnaryExpression | 1 | 0 | 1 | PrefixUnaryExpression |
| parser | src/compiler/parser.ts:5685:makeBinaryExpression | 1 | 0 | 1 | BinaryExpression |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2193:createConstructSignature | 0 | 1 | 1 | ConstructSignature |
| factory outside parsing | src/compiler/factory/nodeFactory.ts:2341:updateFunctionTypeNode | 0 | 1 | 1 | FunctionType |

## Synthetic allocations by calling code site

These are the first non-factory compiler frames in source-mapped stacks, with line and column capture preserved in the compressed full paths. Modules, CLI launchers and instrumentation frames are omitted from this site summary.

| Caller site | Created | Unreachable | Kinds |
|---|---:|---:|---|
| src/compiler/checker.ts:8806 | 572 | 572 | Identifier |
| src/compiler/checker.ts:8751 | 513 | 513 | TypeReference |
| src/compiler/checker.ts:6798 | 263 | 263 | AnyKeyword |
| src/compiler/checker.ts:6809 | 235 | 235 | NumberKeyword |
| src/compiler/checker.ts:6853 | 218 | 218 | LiteralType, StringLiteral |
| src/compiler/checker.ts:6805 | 195 | 195 | StringKeyword |
| src/compiler/checker.ts:6858 | 184 | 184 | FirstLiteralToken, LiteralType |
| src/compiler/checker.ts:8930 | 176 | 176 | Identifier |
| src/compiler/checker.ts:6519 | 141 | 141 | AnyKeyword, BooleanKeyword, FunctionType, Identifier, LiteralType, NumberKeyword, Parameter, PropertySignature, ReadonlyKeyword, StringKeyword, StringLiteral, TrueKeyword, TypeLiteral, TypeQuery, TypeReference, VoidKeyword |
| src/compiler/utilities.ts:10948 | 116 | 116 | Identifier |
| src/compiler/checker.ts:7863 | 116 | 116 | PropertySignature |
| src/compiler/checker.ts:6986 | 107 | 107 | IntersectionType, UnionType |
| src/compiler/checker.ts:6891 | 106 | 106 | LiteralType, NullKeyword |
| src/compiler/checker.ts:8051 | 100 | 100 | FunctionType |
| src/compiler/parser.ts:437 | 98 | 98 | ElementAccessExpression, StringLiteral, SyntheticExpression |
| src/compiler/checker.ts:6887 | 95 | 95 | UndefinedKeyword |
| src/compiler/checker.ts:7414 | 91 | 91 | TypeLiteral |
| src/compiler/checker.ts:8394 | 78 | 78 | Parameter |
| src/compiler/checker.ts:8408 | 75 | 75 | Identifier |
| src/compiler/checker.ts:7028 | 71 | 71 | IndexedAccessType |
| src/compiler/checker.ts:7345 | 67 | 67 | BooleanKeyword, FunctionType, Identifier, LiteralType, NumberKeyword, Parameter, PropertySignature, ReadonlyKeyword, StringKeyword, StringLiteral, TrueKeyword, TypeLiteral, VoidKeyword |
| src/compiler/checker.ts:8898 | 57 | 57 | Identifier |
| src/compiler/checker.ts:8745 | 55 | 55 | TypeQuery |
| src/compiler/checker.ts:8348 | 55 | 55 | TypeParameter |
| src/compiler/checker.ts:7428 | 52 | 52 | ArrayType |
| src/compiler/checker.ts:6817 | 51 | 51 | BooleanKeyword |
| src/compiler/checker.ts:7157 | 49 | 49 | MappedType |
| src/compiler/checker.ts:7457 | 43 | 43 | TupleType |
| src/compiler/checker.ts:6866 | 42 | 42 | FalseKeyword, LiteralType, TrueKeyword |
| src/compiler/checker.ts:44068 | 32 | 32 | AnyKeyword, FunctionType |
| src/compiler/checker.ts:6801 | 31 | 31 | UnknownKeyword |
| src/compiler/checker.ts:8042 | 30 | 30 | CallSignature |
| src/compiler/checker.ts:8390 | 30 | 30 | DotDotDotToken |
| src/compiler/checker.ts:7859 | 22 | 22 | ReadonlyKeyword |
| src/compiler/checker.ts:7429 | 17 | 17 | TypeOperator |
| src/compiler/visitorPublic.ts:671 | 15 | 15 | PropertySignature |
| src/compiler/checker.ts:7462 | 15 | 15 | TupleType |
| src/compiler/checker.ts:7458 | 15 | 15 | TypeOperator |
| src/compiler/checker.ts:7372 | 13 | 13 | TypeLiteral |
| src/compiler/checker.ts:6883 | 12 | 12 | VoidKeyword |
| src/compiler/checker.ts:8393 | 12 | 12 | QuestionToken |
| src/compiler/checker.ts:8431 | 12 | 12 | ArrayBindingPattern, BindingElement, Identifier |
| src/compiler/checker.ts:47809 | 10 | 10 | PropertyAccessExpression, ThisKeyword |
| src/compiler/checker.ts:7004 | 10 | 10 | TypeOperator |
| src/compiler/visitorPublic.ts:830 | 9 | 9 | TypeLiteral |
| src/compiler/checker.ts:7981 | 8 | 8 | Identifier, Parameter |
| src/compiler/expressionToTypeNode.ts:1229 | 7 | 7 | BooleanKeyword, NumberKeyword, StringKeyword |
| src/compiler/checker.ts:7833 | 6 | 6 | QuestionToken |
| src/compiler/checker.ts:6830 | 6 | 6 | Identifier, TypeReference |
| src/compiler/expressionToTypeNode.ts:196 | 5 | 5 | Identifier |
| src/compiler/expressionToTypeNode.ts:1105 | 5 | 5 | PropertySignature |
| src/compiler/checker.ts:6899 | 5 | 5 | SymbolKeyword |
| src/compiler/checker.ts:6895 | 4 | 4 | NeverKeyword |
| src/compiler/visitorPublic.ts:959 | 4 | 4 | LiteralType |
| src/compiler/visitorPublic.ts:997 | 4 | 4 | BindingElement |
| src/compiler/visitorPublic.ts:990 | 4 | 4 | ArrayBindingPattern |
| src/compiler/checker.ts:7996 | 4 | 4 | IndexSignature |
| src/compiler/checker.ts:6879 | 4 | 4 | SymbolKeyword, TypeOperator |
| src/compiler/checker.ts:8815 | 4 | 4 | FirstNode |
| src/compiler/expressionToTypeNode.ts:1090 | 3 | 3 | TypeLiteral |
| src/compiler/checker.ts:7595 | 3 | 3 | FirstNode |
| src/compiler/checker.ts:7597 | 3 | 3 | TypeReference |
| src/compiler/checker.ts:7997 | 2 | 2 | ReadonlyKeyword |
| src/compiler/checker.ts:7463 | 2 | 2 | TypeOperator |
| src/compiler/checker.ts:9267 | 2 | 2 | Identifier |
| src/compiler/expressionToTypeNode.ts:290 | 2 | 2 | TypeReference |
| src/compiler/expressionToTypeNode.ts:271 | 2 | 2 | TypeQuery |
| src/compiler/checker.ts:8052 | 2 | 2 | ConstructorType |
| src/compiler/checker.ts:8296 | 2 | 2 | Block |
| src/compiler/checker.ts:7715 | 1 | 1 | AnyKeyword |
| src/compiler/checker.ts:8659 | 1 | 1 | Identifier |
| src/compiler/checker.ts:6792 | 1 | 1 | TypeReference |
| src/compiler/checker.ts:7136 | 1 | 1 | TypeOperator |
| src/compiler/expressionToTypeNode.ts:1119 | 1 | 1 | QuestionToken |
| src/compiler/expressionToTypeNode.ts:1114 | 1 | 1 | Parameter |
| src/compiler/expressionToTypeNode.ts:975 | 1 | 1 | FunctionType |
| src/compiler/checker.ts:8043 | 1 | 1 | ConstructSignature |
| src/compiler/visitorPublic.ts:651 | 1 | 1 | Parameter |
| src/compiler/visitorPublic.ts:803 | 1 | 1 | FunctionType |
