# Factory and parser body index

All detailed rows from data/factories.json. Field statuses and evidence are in FIELDS.tsv; every remaining callable body is in data/functions.json. Completeness is for this function boundary under the stock allocator profile.

| Body and source | Made/returned type | Role | Write-complete | Non-undefined complete |
| --- | --- | --- | --- | --- |
| createBaseNodeFactory at src/compiler/factory/baseNodeFactory.ts:26:1 | BaseNodeFactory | non-node-builder | unproven | unproven |
| createBaseSourceFileNode at src/compiler/factory/baseNodeFactory.ts:41:5 | Node | allocating | yes | unproven |
| createBaseIdentifierNode at src/compiler/factory/baseNodeFactory.ts:45:5 | Node | allocating | unproven | unproven |
| createBasePrivateIdentifierNode at src/compiler/factory/baseNodeFactory.ts:49:5 | Node | allocating | yes | unproven |
| createBaseTokenNode at src/compiler/factory/baseNodeFactory.ts:53:5 | Node | allocating | unproven | unproven |
| createBaseNode at src/compiler/factory/baseNodeFactory.ts:57:5 | Node | allocating | yes | unproven |
| createParenthesizerRules at src/compiler/factory/parenthesizerRules.ts:54:1 | ParenthesizerRules | non-node-builder | unproven | unproven |
| <anonymous@99:33> at src/compiler/factory/parenthesizerRules.ts:99:33 | Expression | composed-or-forwarding | unproven | unproven |
| <anonymous@109:33> at src/compiler/factory/parenthesizerRules.ts:109:33 | Expression | composed-or-forwarding | unproven | unproven |
| binaryOperandNeedsParentheses at src/compiler/factory/parenthesizerRules.ts:133:5 | boolean | non-node-builder | unproven | unproven |
| getLiteralKindOfBinaryPlusOperand at src/compiler/factory/parenthesizerRules.ts:265:5 | SyntaxKind | non-node-builder | unproven | unproven |
| parenthesizeBinaryOperand at src/compiler/factory/parenthesizerRules.ts:299:5 | Expression | composed-or-forwarding | yes | unproven |
| parenthesizeLeftSideOfBinary at src/compiler/factory/parenthesizerRules.ts:312:5 | Expression | composed-or-forwarding | unproven | unproven |
| parenthesizeRightSideOfBinary at src/compiler/factory/parenthesizerRules.ts:316:5 | Expression | composed-or-forwarding | unproven | unproven |
| parenthesizeExpressionOfComputedPropertyName at src/compiler/factory/parenthesizerRules.ts:320:5 | Expression | composed-or-forwarding | yes | unproven |
| parenthesizeConditionOfConditionalExpression at src/compiler/factory/parenthesizerRules.ts:324:5 | Expression | composed-or-forwarding | yes | unproven |
| parenthesizeBranchOfConditionalExpression at src/compiler/factory/parenthesizerRules.ts:334:5 | Expression | composed-or-forwarding | yes | unproven |
| parenthesizeExpressionOfExportDefault at src/compiler/factory/parenthesizerRules.ts:355:5 | Expression | composed-or-forwarding | yes | unproven |
| parenthesizeExpressionOfNew at src/compiler/factory/parenthesizerRules.ts:372:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| parenthesizeLeftSideOfAccess at src/compiler/factory/parenthesizerRules.ts:391:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| parenthesizeOperandOfPostfixUnary at src/compiler/factory/parenthesizerRules.ts:412:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| parenthesizeOperandOfPrefixUnary at src/compiler/factory/parenthesizerRules.ts:417:5 | UnaryExpression | composed-or-forwarding | unproven | unproven |
| parenthesizeExpressionForDisallowedComma at src/compiler/factory/parenthesizerRules.ts:427:5 | Expression | composed-or-forwarding | unproven | unproven |
| parenthesizeExpressionOfExpressionStatement at src/compiler/factory/parenthesizerRules.ts:435:5 | Expression | composed-or-forwarding | unproven | unproven |
| parenthesizeConciseBodyOfArrowFunction at src/compiler/factory/parenthesizerRules.ts:463:5 | ConciseBody | composed-or-forwarding | unproven | unproven |
| parenthesizeCheckTypeOfConditionalType at src/compiler/factory/parenthesizerRules.ts:483:5 | TypeNode | composed-or-forwarding | yes | unproven |
| parenthesizeExtendsTypeOfConditionalType at src/compiler/factory/parenthesizerRules.ts:493:5 | TypeNode | composed-or-forwarding | yes | unproven |
| parenthesizeConstituentTypeOfUnionType at src/compiler/factory/parenthesizerRules.ts:506:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeConstituentTypeOfIntersectionType at src/compiler/factory/parenthesizerRules.ts:524:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeOperandOfTypeOperator at src/compiler/factory/parenthesizerRules.ts:544:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeOperandOfReadonlyTypeOperator at src/compiler/factory/parenthesizerRules.ts:552:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeNonArrayTypeOfPostfixType at src/compiler/factory/parenthesizerRules.ts:573:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeElementTypeOfTupleType at src/compiler/factory/parenthesizerRules.ts:617:5 | TypeNode | composed-or-forwarding | yes | unproven |
| hasJSDocPostfixQuestion at src/compiler/factory/parenthesizerRules.ts:622:5 | boolean | non-node-builder | unproven | unproven |
| parenthesizeTypeOfOptionalType at src/compiler/factory/parenthesizerRules.ts:633:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeLeadingTypeArgument at src/compiler/factory/parenthesizerRules.ts:659:5 | TypeNode | composed-or-forwarding | yes | unproven |
| parenthesizeOrdinalTypeArgument at src/compiler/factory/parenthesizerRules.ts:663:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parenthesizeLeftSideOfBinary at src/compiler/factory/parenthesizerRules.ts:678:35 | Expression | input-or-unresolved | unproven | unproven |
| parenthesizeRightSideOfBinary at src/compiler/factory/parenthesizerRules.ts:679:36 | Expression | input-or-unresolved | unproven | unproven |
| parenthesizeExpressionOfNew at src/compiler/factory/parenthesizerRules.ts:684:34 | LeftHandSideExpression | input-or-unresolved | unproven | unproven |
| parenthesizeLeftSideOfAccess at src/compiler/factory/parenthesizerRules.ts:685:35 | LeftHandSideExpression | input-or-unresolved | unproven | unproven |
| parenthesizeOperandOfPostfixUnary at src/compiler/factory/parenthesizerRules.ts:686:40 | LeftHandSideExpression | input-or-unresolved | unproven | unproven |
| parenthesizeOperandOfPrefixUnary at src/compiler/factory/parenthesizerRules.ts:687:39 | UnaryExpression | input-or-unresolved | unproven | unproven |
| createNodeConverters at src/compiler/factory/nodeConverters.ts:41:1 | NodeConverters | non-node-builder | unproven | unproven |
| convertToFunctionBlock at src/compiler/factory/nodeConverters.ts:54:5 | Block | composed-or-forwarding | yes | unproven |
| convertToFunctionExpression at src/compiler/factory/nodeConverters.ts:63:5 | FunctionExpression | composed-or-forwarding | yes | unproven |
| <anonymous@66:40> at src/compiler/factory/nodeConverters.ts:66:40 | boolean | non-node-builder | unproven | unproven |
| convertToClassExpression at src/compiler/factory/nodeConverters.ts:82:5 | ClassExpression | composed-or-forwarding | yes | unproven |
| <anonymous@84:36> at src/compiler/factory/nodeConverters.ts:84:36 | boolean | non-node-builder | unproven | unproven |
| convertToArrayAssignmentElement at src/compiler/factory/nodeConverters.ts:98:5 | Expression | composed-or-forwarding | unproven | unproven |
| convertToObjectAssignmentElement at src/compiler/factory/nodeConverters.ts:118:5 | ObjectLiteralElementLike | composed-or-forwarding | unproven | unproven |
| convertToAssignmentPattern at src/compiler/factory/nodeConverters.ts:135:5 | AssignmentPattern | composed-or-forwarding | unproven | unproven |
| convertToObjectAssignmentPattern at src/compiler/factory/nodeConverters.ts:147:5 | ObjectLiteralExpression | composed-or-forwarding | unproven | unproven |
| convertToArrayAssignmentPattern at src/compiler/factory/nodeConverters.ts:160:5 | ArrayLiteralExpression | composed-or-forwarding | unproven | unproven |
| convertToAssignmentElementTarget at src/compiler/factory/nodeConverters.ts:173:5 | Expression | composed-or-forwarding | unproven | unproven |
| createNodeFactory at src/compiler/factory/nodeFactory.ts:492:1 | NodeFactory | non-node-builder | unproven | unproven |
| <anonymous@500:78> at src/compiler/factory/nodeFactory.ts:500:78 | Mutable<BinaryExpression> | composed-or-forwarding | yes | unproven |
| <anonymous@501:88> at src/compiler/factory/nodeFactory.ts:501:88 | Mutable<PrefixUnaryExpression> | composed-or-forwarding | yes | unproven |
| <anonymous@502:90> at src/compiler/factory/nodeFactory.ts:502:90 | Mutable<PostfixUnaryExpression> | composed-or-forwarding | yes | unproven |
| <anonymous@503:100> at src/compiler/factory/nodeFactory.ts:503:100 | Mutable<T> | composed-or-forwarding | yes | unproven |
| <anonymous@504:141> at src/compiler/factory/nodeFactory.ts:504:141 | T | composed-or-forwarding | yes | unproven |
| <anonymous@505:141> at src/compiler/factory/nodeFactory.ts:505:141 | T | composed-or-forwarding | unproven | unproven |
| <anonymous@506:178> at src/compiler/factory/nodeFactory.ts:506:178 | T | composed-or-forwarding | yes | unproven |
| <anonymous@507:178> at src/compiler/factory/nodeFactory.ts:507:178 | T | composed-or-forwarding | unproven | unproven |
| <anonymous@508:97> at src/compiler/factory/nodeFactory.ts:508:97 | Mutable<T> | composed-or-forwarding | yes | unproven |
| <anonymous@509:97> at src/compiler/factory/nodeFactory.ts:509:97 | T | composed-or-forwarding | unproven | unproven |
| <anonymous@510:143> at src/compiler/factory/nodeFactory.ts:510:143 | Mutable<T> | composed-or-forwarding | yes | unproven |
| <anonymous@511:143> at src/compiler/factory/nodeFactory.ts:511:143 | T | composed-or-forwarding | unproven | unproven |
| <anonymous@636:13> at src/compiler/factory/nodeFactory.ts:636:13 | Mutable<PropertyAccessExpression> | composed-or-forwarding | unproven | unproven |
| <anonymous@640:13> at src/compiler/factory/nodeFactory.ts:640:13 | Mutable<PropertyAccessChain> | composed-or-forwarding | unproven | unproven |
| createJSDocAllType at src/compiler/factory/nodeFactory.ts:802:9 | () => Mutable<JSDocAllType> | non-node-builder | unproven | unproven |
| createJSDocUnknownType at src/compiler/factory/nodeFactory.ts:805:9 | () => Mutable<JSDocUnknownType> | non-node-builder | unproven | unproven |
| createJSDocNonNullableType at src/compiler/factory/nodeFactory.ts:808:9 | (type: TypeNode, postfix?: boolean &#124; undefined) => JSDocNonNullableType | non-node-builder | unproven | unproven |
| updateJSDocNonNullableType at src/compiler/factory/nodeFactory.ts:811:9 | (node: JSDocNonNullableType, type: TypeNode) => JSDocNonNullableType | non-node-builder | unproven | unproven |
| createJSDocNullableType at src/compiler/factory/nodeFactory.ts:814:9 | (type: TypeNode, postfix?: boolean &#124; undefined) => JSDocNullableType | non-node-builder | unproven | unproven |
| updateJSDocNullableType at src/compiler/factory/nodeFactory.ts:817:9 | (node: JSDocNullableType, type: TypeNode) => JSDocNullableType | non-node-builder | unproven | unproven |
| createJSDocOptionalType at src/compiler/factory/nodeFactory.ts:820:9 | (type: TypeNode) => JSDocOptionalType | non-node-builder | unproven | unproven |
| updateJSDocOptionalType at src/compiler/factory/nodeFactory.ts:823:9 | (node: JSDocOptionalType, type: TypeNode) => JSDocOptionalType | non-node-builder | unproven | unproven |
| createJSDocVariadicType at src/compiler/factory/nodeFactory.ts:826:9 | (type: TypeNode) => JSDocVariadicType | non-node-builder | unproven | unproven |
| updateJSDocVariadicType at src/compiler/factory/nodeFactory.ts:829:9 | (node: JSDocVariadicType, type: TypeNode) => JSDocVariadicType | non-node-builder | unproven | unproven |
| createJSDocNamepathType at src/compiler/factory/nodeFactory.ts:832:9 | (type: TypeNode) => JSDocNamepathType | non-node-builder | unproven | unproven |
| updateJSDocNamepathType at src/compiler/factory/nodeFactory.ts:835:9 | (node: JSDocNamepathType, type: TypeNode) => JSDocNamepathType | non-node-builder | unproven | unproven |
| createJSDocTypeTag at src/compiler/factory/nodeFactory.ts:877:9 | (tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<...> | non-node-builder | unproven | unproven |
| updateJSDocTypeTag at src/compiler/factory/nodeFactory.ts:880:9 | (node: JSDocTypeTag, tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocTypeTag | non-node-builder | unproven | unproven |
| createJSDocReturnTag at src/compiler/factory/nodeFactory.ts:883:9 | (tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<...> | non-node-builder | unproven | unproven |
| updateJSDocReturnTag at src/compiler/factory/nodeFactory.ts:886:9 | (node: JSDocReturnTag, tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocReturnTag | non-node-builder | unproven | unproven |
| createJSDocThisTag at src/compiler/factory/nodeFactory.ts:889:9 | (tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<...> | non-node-builder | unproven | unproven |
| updateJSDocThisTag at src/compiler/factory/nodeFactory.ts:892:9 | (node: JSDocThisTag, tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocThisTag | non-node-builder | unproven | unproven |
| createJSDocAuthorTag at src/compiler/factory/nodeFactory.ts:895:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocAuthorTag> | non-node-builder | unproven | unproven |
| updateJSDocAuthorTag at src/compiler/factory/nodeFactory.ts:898:9 | (node: JSDocAuthorTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocAuthorTag | non-node-builder | unproven | unproven |
| createJSDocClassTag at src/compiler/factory/nodeFactory.ts:901:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocClassTag> | non-node-builder | unproven | unproven |
| updateJSDocClassTag at src/compiler/factory/nodeFactory.ts:904:9 | (node: JSDocClassTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocClassTag | non-node-builder | unproven | unproven |
| createJSDocPublicTag at src/compiler/factory/nodeFactory.ts:907:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocPublicTag> | non-node-builder | unproven | unproven |
| updateJSDocPublicTag at src/compiler/factory/nodeFactory.ts:910:9 | (node: JSDocPublicTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocPublicTag | non-node-builder | unproven | unproven |
| createJSDocPrivateTag at src/compiler/factory/nodeFactory.ts:913:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocPrivateTag> | non-node-builder | unproven | unproven |
| updateJSDocPrivateTag at src/compiler/factory/nodeFactory.ts:916:9 | (node: JSDocPrivateTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocPrivateTag | non-node-builder | unproven | unproven |
| createJSDocProtectedTag at src/compiler/factory/nodeFactory.ts:919:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocProtectedTag> | non-node-builder | unproven | unproven |
| updateJSDocProtectedTag at src/compiler/factory/nodeFactory.ts:922:9 | (node: JSDocProtectedTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocProtectedTag | non-node-builder | unproven | unproven |
| createJSDocReadonlyTag at src/compiler/factory/nodeFactory.ts:925:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocReadonlyTag> | non-node-builder | unproven | unproven |
| updateJSDocReadonlyTag at src/compiler/factory/nodeFactory.ts:928:9 | (node: JSDocReadonlyTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocReadonlyTag | non-node-builder | unproven | unproven |
| createJSDocOverrideTag at src/compiler/factory/nodeFactory.ts:931:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocOverrideTag> | non-node-builder | unproven | unproven |
| updateJSDocOverrideTag at src/compiler/factory/nodeFactory.ts:934:9 | (node: JSDocOverrideTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocOverrideTag | non-node-builder | unproven | unproven |
| createJSDocDeprecatedTag at src/compiler/factory/nodeFactory.ts:937:9 | (tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<JSDocDeprecatedTag> | non-node-builder | unproven | unproven |
| updateJSDocDeprecatedTag at src/compiler/factory/nodeFactory.ts:940:9 | (node: JSDocDeprecatedTag, tagName: Identifier &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocDeprecatedTag | non-node-builder | unproven | unproven |
| createJSDocThrowsTag at src/compiler/factory/nodeFactory.ts:943:9 | (tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<...> | non-node-builder | unproven | unproven |
| updateJSDocThrowsTag at src/compiler/factory/nodeFactory.ts:946:9 | (node: JSDocThrowsTag, tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => JSDocThrowsTag | non-node-builder | unproven | unproven |
| createJSDocSatisfiesTag at src/compiler/factory/nodeFactory.ts:949:9 | (tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<JSDocComment> &#124; undefined) => Mutable<...> | non-node-builder | unproven | unproven |
| updateJSDocSatisfiesTag at src/compiler/factory/nodeFactory.ts:952:9 | (node: JSDocSatisfiesTag, tagName: Identifier &#124; undefined, typeExpression?: JSDocTypeExpression &#124; undefined, comment?: NodeArray<...> &#124; undefined) => JSDocSatisfiesTag | non-node-builder | unproven | unproven |
| createComma at src/compiler/factory/nodeFactory.ts:1022:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createAssignment at src/compiler/factory/nodeFactory.ts:1025:9 | { (left: ArrayLiteralExpression &#124; ObjectLiteralExpression, right: Expression): DestructuringAssignment; (left: Expression, right: Expression): AssignmentExpression<...>; } | non-node-builder | unproven | unproven |
| createLogicalOr at src/compiler/factory/nodeFactory.ts:1028:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createLogicalAnd at src/compiler/factory/nodeFactory.ts:1031:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createBitwiseOr at src/compiler/factory/nodeFactory.ts:1034:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createBitwiseXor at src/compiler/factory/nodeFactory.ts:1037:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createBitwiseAnd at src/compiler/factory/nodeFactory.ts:1040:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createStrictEquality at src/compiler/factory/nodeFactory.ts:1043:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createStrictInequality at src/compiler/factory/nodeFactory.ts:1046:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createEquality at src/compiler/factory/nodeFactory.ts:1049:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createInequality at src/compiler/factory/nodeFactory.ts:1052:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createLessThan at src/compiler/factory/nodeFactory.ts:1055:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createLessThanEquals at src/compiler/factory/nodeFactory.ts:1058:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createGreaterThan at src/compiler/factory/nodeFactory.ts:1061:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createGreaterThanEquals at src/compiler/factory/nodeFactory.ts:1064:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createLeftShift at src/compiler/factory/nodeFactory.ts:1067:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createRightShift at src/compiler/factory/nodeFactory.ts:1070:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createUnsignedRightShift at src/compiler/factory/nodeFactory.ts:1073:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createAdd at src/compiler/factory/nodeFactory.ts:1076:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createSubtract at src/compiler/factory/nodeFactory.ts:1079:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createMultiply at src/compiler/factory/nodeFactory.ts:1082:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createDivide at src/compiler/factory/nodeFactory.ts:1085:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createModulo at src/compiler/factory/nodeFactory.ts:1088:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createExponent at src/compiler/factory/nodeFactory.ts:1091:9 | (left: Expression, right: Expression) => Mutable<BinaryExpression> | non-node-builder | unproven | unproven |
| createPrefixPlus at src/compiler/factory/nodeFactory.ts:1094:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createPrefixMinus at src/compiler/factory/nodeFactory.ts:1097:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createPrefixIncrement at src/compiler/factory/nodeFactory.ts:1100:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createPrefixDecrement at src/compiler/factory/nodeFactory.ts:1103:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createBitwiseNot at src/compiler/factory/nodeFactory.ts:1106:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createLogicalNot at src/compiler/factory/nodeFactory.ts:1109:9 | (operand: Expression) => Mutable<PrefixUnaryExpression> | non-node-builder | unproven | unproven |
| createPostfixIncrement at src/compiler/factory/nodeFactory.ts:1112:9 | (operand: Expression) => Mutable<PostfixUnaryExpression> | non-node-builder | unproven | unproven |
| createPostfixDecrement at src/compiler/factory/nodeFactory.ts:1115:9 | (operand: Expression) => Mutable<PostfixUnaryExpression> | non-node-builder | unproven | unproven |
| createNodeArray at src/compiler/factory/nodeFactory.ts:1169:5 | NodeArray<T> | non-node-builder | unproven | unproven |
| createBaseNode at src/compiler/factory/nodeFactory.ts:1209:5 | Mutable<T> | allocating | yes | unproven |
| createBaseDeclaration at src/compiler/factory/nodeFactory.ts:1213:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| finishUpdateBaseSignatureDeclaration at src/compiler/factory/nodeFactory.ts:1220:5 | T | composed-or-forwarding | unproven | unproven |
| createNumericLiteral at src/compiler/factory/nodeFactory.ts:1233:5 | NumericLiteral | composed-or-forwarding | yes | unproven |
| createBigIntLiteral at src/compiler/factory/nodeFactory.ts:1244:5 | BigIntLiteral | composed-or-forwarding | unproven | unproven |
| createBaseStringLiteral at src/compiler/factory/nodeFactory.ts:1251:5 | Mutable<StringLiteral> | composed-or-forwarding | yes | unproven |
| createStringLiteral at src/compiler/factory/nodeFactory.ts:1259:5 | StringLiteral | composed-or-forwarding | yes | unproven |
| createStringLiteralFromNode at src/compiler/factory/nodeFactory.ts:1267:5 | StringLiteral | composed-or-forwarding | yes | unproven |
| createRegularExpressionLiteral at src/compiler/factory/nodeFactory.ts:1274:5 | RegularExpressionLiteral | composed-or-forwarding | unproven | unproven |
| createLiteralLikeNode at src/compiler/factory/nodeFactory.ts:1281:5 | LiteralToken | composed-or-forwarding | unproven | unproven |
| createBaseIdentifier at src/compiler/factory/nodeFactory.ts:1304:5 | Mutable<Identifier> | allocating | unproven | unproven |
| createBaseGeneratedIdentifier at src/compiler/factory/nodeFactory.ts:1313:5 | Mutable<GeneratedIdentifier> | composed-or-forwarding | unproven | unproven |
| createIdentifier at src/compiler/factory/nodeFactory.ts:1326:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createTempVariable at src/compiler/factory/nodeFactory.ts:1349:5 | GeneratedIdentifier | composed-or-forwarding | unproven | unproven |
| createLoopVariable at src/compiler/factory/nodeFactory.ts:1361:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createUniqueName at src/compiler/factory/nodeFactory.ts:1369:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getGeneratedNameForNode at src/compiler/factory/nodeFactory.ts:1377:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createBasePrivateIdentifier at src/compiler/factory/nodeFactory.ts:1388:5 | Mutable<PrivateIdentifier> | allocating | yes | unproven |
| createPrivateIdentifier at src/compiler/factory/nodeFactory.ts:1396:5 | PrivateIdentifier | composed-or-forwarding | yes | unproven |
| createBaseGeneratedPrivateIdentifier at src/compiler/factory/nodeFactory.ts:1401:5 | Mutable<PrivateIdentifier> | composed-or-forwarding | yes | unproven |
| createUniquePrivateName at src/compiler/factory/nodeFactory.ts:1415:5 | PrivateIdentifier | composed-or-forwarding | yes | unproven |
| getGeneratedPrivateNameForNode at src/compiler/factory/nodeFactory.ts:1423:5 | PrivateIdentifier | composed-or-forwarding | yes | unproven |
| createBaseToken at src/compiler/factory/nodeFactory.ts:1436:5 | Mutable<T> | allocating | unproven | unproven |
| createToken at src/compiler/factory/nodeFactory.ts:1453:5 | Mutable<Token<TKind>> | composed-or-forwarding | unproven | unproven |
| createSuper at src/compiler/factory/nodeFactory.ts:1521:5 | SuperExpression | composed-or-forwarding | unproven | unproven |
| createThis at src/compiler/factory/nodeFactory.ts:1526:5 | ThisExpression | composed-or-forwarding | unproven | unproven |
| createNull at src/compiler/factory/nodeFactory.ts:1531:5 | NullLiteral | composed-or-forwarding | unproven | unproven |
| createTrue at src/compiler/factory/nodeFactory.ts:1536:5 | TrueLiteral | composed-or-forwarding | unproven | unproven |
| createFalse at src/compiler/factory/nodeFactory.ts:1541:5 | FalseLiteral | composed-or-forwarding | unproven | unproven |
| createModifier at src/compiler/factory/nodeFactory.ts:1550:5 | ModifierToken<T> | composed-or-forwarding | unproven | unproven |
| createModifiersFromModifierFlags at src/compiler/factory/nodeFactory.ts:1555:5 | Modifier[] &#124; undefined | non-node-builder | unproven | unproven |
| createQualifiedName at src/compiler/factory/nodeFactory.ts:1580:5 | Mutable<QualifiedName> | composed-or-forwarding | yes | unproven |
| updateQualifiedName at src/compiler/factory/nodeFactory.ts:1592:5 | QualifiedName | composed-or-forwarding | unproven | unproven |
| createComputedPropertyName at src/compiler/factory/nodeFactory.ts:1600:5 | Mutable<ComputedPropertyName> | composed-or-forwarding | yes | unproven |
| updateComputedPropertyName at src/compiler/factory/nodeFactory.ts:1610:5 | ComputedPropertyName | composed-or-forwarding | unproven | unproven |
| createTypeParameterDeclaration at src/compiler/factory/nodeFactory.ts:1621:5 | TypeParameterDeclaration | composed-or-forwarding | yes | unproven |
| updateTypeParameterDeclaration at src/compiler/factory/nodeFactory.ts:1635:5 | TypeParameterDeclaration | composed-or-forwarding | unproven | unproven |
| createParameterDeclaration at src/compiler/factory/nodeFactory.ts:1645:5 | Mutable<ParameterDeclaration> | composed-or-forwarding | yes | unproven |
| updateParameterDeclaration at src/compiler/factory/nodeFactory.ts:1680:5 | ParameterDeclaration | composed-or-forwarding | unproven | unproven |
| createDecorator at src/compiler/factory/nodeFactory.ts:1700:5 | Mutable<Decorator> | composed-or-forwarding | yes | unproven |
| updateDecorator at src/compiler/factory/nodeFactory.ts:1711:5 | Decorator | composed-or-forwarding | unproven | unproven |
| createPropertySignature at src/compiler/factory/nodeFactory.ts:1722:5 | PropertySignature | composed-or-forwarding | yes | unproven |
| updatePropertySignature at src/compiler/factory/nodeFactory.ts:1741:5 | PropertySignature | composed-or-forwarding | unproven | unproven |
| finishUpdatePropertySignature at src/compiler/factory/nodeFactory.ts:1756:5 | PropertySignature | composed-or-forwarding | unproven | unproven |
| createPropertyDeclaration at src/compiler/factory/nodeFactory.ts:1765:5 | Mutable<PropertyDeclaration> | composed-or-forwarding | yes | unproven |
| updatePropertyDeclaration at src/compiler/factory/nodeFactory.ts:1794:5 | PropertyDeclaration | composed-or-forwarding | unproven | unproven |
| createMethodSignature at src/compiler/factory/nodeFactory.ts:1813:5 | Mutable<MethodSignature> | composed-or-forwarding | yes | unproven |
| updateMethodSignature at src/compiler/factory/nodeFactory.ts:1838:5 | MethodSignature | composed-or-forwarding | unproven | unproven |
| createMethodDeclaration at src/compiler/factory/nodeFactory.ts:1858:5 | Mutable<MethodDeclaration> | composed-or-forwarding | yes | unproven |
| updateMethodDeclaration at src/compiler/factory/nodeFactory.ts:1914:5 | MethodDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateMethodDeclaration at src/compiler/factory/nodeFactory.ts:1937:5 | MethodDeclaration | composed-or-forwarding | unproven | unproven |
| createClassStaticBlockDeclaration at src/compiler/factory/nodeFactory.ts:1946:5 | ClassStaticBlockDeclaration | composed-or-forwarding | yes | unproven |
| updateClassStaticBlockDeclaration at src/compiler/factory/nodeFactory.ts:1963:5 | ClassStaticBlockDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateClassStaticBlockDeclaration at src/compiler/factory/nodeFactory.ts:1972:5 | ClassStaticBlockDeclaration | composed-or-forwarding | unproven | unproven |
| createConstructorDeclaration at src/compiler/factory/nodeFactory.ts:1981:5 | Mutable<ConstructorDeclaration> | composed-or-forwarding | yes | unproven |
| updateConstructorDeclaration at src/compiler/factory/nodeFactory.ts:2013:5 | ConstructorDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateConstructorDeclaration at src/compiler/factory/nodeFactory.ts:2026:5 | ConstructorDeclaration | composed-or-forwarding | unproven | unproven |
| createGetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2035:5 | Mutable<GetAccessorDeclaration> | composed-or-forwarding | yes | unproven |
| updateGetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2073:5 | GetAccessorDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateGetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2090:5 | GetAccessorDeclaration | composed-or-forwarding | unproven | unproven |
| createSetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2099:5 | Mutable<SetAccessorDeclaration> | composed-or-forwarding | yes | unproven |
| updateSetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2135:5 | SetAccessorDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateSetAccessorDeclaration at src/compiler/factory/nodeFactory.ts:2150:5 | SetAccessorDeclaration | composed-or-forwarding | unproven | unproven |
| createCallSignature at src/compiler/factory/nodeFactory.ts:2160:5 | CallSignatureDeclaration | composed-or-forwarding | yes | unproven |
| updateCallSignature at src/compiler/factory/nodeFactory.ts:2179:5 | CallSignatureDeclaration | composed-or-forwarding | unproven | unproven |
| createConstructSignature at src/compiler/factory/nodeFactory.ts:2193:5 | ConstructSignatureDeclaration | composed-or-forwarding | yes | unproven |
| updateConstructSignature at src/compiler/factory/nodeFactory.ts:2212:5 | ConstructSignatureDeclaration | composed-or-forwarding | unproven | unproven |
| createIndexSignature at src/compiler/factory/nodeFactory.ts:2226:5 | IndexSignatureDeclaration | composed-or-forwarding | yes | unproven |
| updateIndexSignature at src/compiler/factory/nodeFactory.ts:2245:5 | IndexSignatureDeclaration | composed-or-forwarding | unproven | unproven |
| createTemplateLiteralTypeSpan at src/compiler/factory/nodeFactory.ts:2259:5 | Mutable<TemplateLiteralTypeSpan> | composed-or-forwarding | yes | unproven |
| updateTemplateLiteralTypeSpan at src/compiler/factory/nodeFactory.ts:2268:5 | TemplateLiteralTypeSpan | composed-or-forwarding | unproven | unproven |
| createKeywordTypeNode at src/compiler/factory/nodeFactory.ts:2280:5 | KeywordTypeNode<TKind> | composed-or-forwarding | unproven | unproven |
| createTypePredicateNode at src/compiler/factory/nodeFactory.ts:2285:5 | Mutable<TypePredicateNode> | composed-or-forwarding | yes | unproven |
| updateTypePredicateNode at src/compiler/factory/nodeFactory.ts:2295:5 | TypePredicateNode | composed-or-forwarding | unproven | unproven |
| createTypeReferenceNode at src/compiler/factory/nodeFactory.ts:2304:5 | Mutable<TypeReferenceNode> | composed-or-forwarding | yes | unproven |
| updateTypeReferenceNode at src/compiler/factory/nodeFactory.ts:2313:5 | TypeReferenceNode | composed-or-forwarding | unproven | unproven |
| createFunctionTypeNode at src/compiler/factory/nodeFactory.ts:2321:5 | FunctionTypeNode | composed-or-forwarding | yes | unproven |
| updateFunctionTypeNode at src/compiler/factory/nodeFactory.ts:2341:5 | FunctionTypeNode | composed-or-forwarding | unproven | unproven |
| finishUpdateFunctionTypeNode at src/compiler/factory/nodeFactory.ts:2354:5 | FunctionTypeNode | composed-or-forwarding | unproven | unproven |
| createConstructorTypeNode at src/compiler/factory/nodeFactory.ts:2363:5 | ConstructorTypeNode | composed-or-forwarding | yes | unproven |
| createConstructorTypeNode1 at src/compiler/factory/nodeFactory.ts:2369:5 | ConstructorTypeNode | composed-or-forwarding | yes | unproven |
| createConstructorTypeNode2 at src/compiler/factory/nodeFactory.ts:2390:5 | ConstructorTypeNode | composed-or-forwarding | yes | unproven |
| updateConstructorTypeNode at src/compiler/factory/nodeFactory.ts:2399:5 | ConstructorTypeNode | composed-or-forwarding | unproven | unproven |
| updateConstructorTypeNode1 at src/compiler/factory/nodeFactory.ts:2405:5 | ConstructorTypeNode | composed-or-forwarding | unproven | unproven |
| updateConstructorTypeNode2 at src/compiler/factory/nodeFactory.ts:2421:5 | ConstructorTypeNode | composed-or-forwarding | unproven | unproven |
| createTypeQueryNode at src/compiler/factory/nodeFactory.ts:2431:5 | Mutable<TypeQueryNode> | composed-or-forwarding | yes | unproven |
| updateTypeQueryNode at src/compiler/factory/nodeFactory.ts:2440:5 | TypeQueryNode | composed-or-forwarding | unproven | unproven |
| createTypeLiteralNode at src/compiler/factory/nodeFactory.ts:2448:5 | Mutable<TypeLiteralNode> | composed-or-forwarding | yes | unproven |
| updateTypeLiteralNode at src/compiler/factory/nodeFactory.ts:2456:5 | TypeLiteralNode | composed-or-forwarding | unproven | unproven |
| createArrayTypeNode at src/compiler/factory/nodeFactory.ts:2463:5 | Mutable<ArrayTypeNode> | composed-or-forwarding | yes | unproven |
| updateArrayTypeNode at src/compiler/factory/nodeFactory.ts:2471:5 | ArrayTypeNode | composed-or-forwarding | unproven | unproven |
| createTupleTypeNode at src/compiler/factory/nodeFactory.ts:2478:5 | Mutable<TupleTypeNode> | composed-or-forwarding | yes | unproven |
| updateTupleTypeNode at src/compiler/factory/nodeFactory.ts:2486:5 | TupleTypeNode | composed-or-forwarding | unproven | unproven |
| createNamedTupleMember at src/compiler/factory/nodeFactory.ts:2493:5 | Mutable<NamedTupleMember> | composed-or-forwarding | yes | unproven |
| updateNamedTupleMember at src/compiler/factory/nodeFactory.ts:2506:5 | NamedTupleMember | composed-or-forwarding | unproven | unproven |
| createOptionalTypeNode at src/compiler/factory/nodeFactory.ts:2516:5 | Mutable<OptionalTypeNode> | composed-or-forwarding | yes | unproven |
| updateOptionalTypeNode at src/compiler/factory/nodeFactory.ts:2524:5 | OptionalTypeNode | composed-or-forwarding | unproven | unproven |
| createRestTypeNode at src/compiler/factory/nodeFactory.ts:2531:5 | Mutable<RestTypeNode> | composed-or-forwarding | yes | unproven |
| updateRestTypeNode at src/compiler/factory/nodeFactory.ts:2539:5 | RestTypeNode | composed-or-forwarding | unproven | unproven |
| createUnionOrIntersectionTypeNode at src/compiler/factory/nodeFactory.ts:2545:5 | Mutable<UnionTypeNode &#124; IntersectionTypeNode> | composed-or-forwarding | yes | unproven |
| updateUnionOrIntersectionTypeNode at src/compiler/factory/nodeFactory.ts:2552:5 | T | composed-or-forwarding | unproven | unproven |
| createUnionTypeNode at src/compiler/factory/nodeFactory.ts:2559:5 | UnionTypeNode | composed-or-forwarding | yes | unproven |
| updateUnionTypeNode at src/compiler/factory/nodeFactory.ts:2564:5 | UnionTypeNode | composed-or-forwarding | unproven | unproven |
| createIntersectionTypeNode at src/compiler/factory/nodeFactory.ts:2569:5 | IntersectionTypeNode | composed-or-forwarding | yes | unproven |
| updateIntersectionTypeNode at src/compiler/factory/nodeFactory.ts:2574:5 | IntersectionTypeNode | composed-or-forwarding | unproven | unproven |
| createConditionalTypeNode at src/compiler/factory/nodeFactory.ts:2579:5 | Mutable<ConditionalTypeNode> | composed-or-forwarding | yes | unproven |
| updateConditionalTypeNode at src/compiler/factory/nodeFactory.ts:2593:5 | ConditionalTypeNode | composed-or-forwarding | unproven | unproven |
| createInferTypeNode at src/compiler/factory/nodeFactory.ts:2603:5 | Mutable<InferTypeNode> | composed-or-forwarding | yes | unproven |
| updateInferTypeNode at src/compiler/factory/nodeFactory.ts:2611:5 | InferTypeNode | composed-or-forwarding | unproven | unproven |
| createTemplateLiteralType at src/compiler/factory/nodeFactory.ts:2618:5 | Mutable<TemplateLiteralTypeNode> | composed-or-forwarding | yes | unproven |
| updateTemplateLiteralType at src/compiler/factory/nodeFactory.ts:2627:5 | TemplateLiteralTypeNode | composed-or-forwarding | unproven | unproven |
| createImportTypeNode at src/compiler/factory/nodeFactory.ts:2635:5 | ImportTypeNode | composed-or-forwarding | yes | unproven |
| updateImportTypeNode at src/compiler/factory/nodeFactory.ts:2656:5 | ImportTypeNode | composed-or-forwarding | unproven | unproven |
| createParenthesizedType at src/compiler/factory/nodeFactory.ts:2674:5 | Mutable<ParenthesizedTypeNode> | composed-or-forwarding | yes | unproven |
| updateParenthesizedType at src/compiler/factory/nodeFactory.ts:2682:5 | ParenthesizedTypeNode | composed-or-forwarding | unproven | unproven |
| createThisTypeNode at src/compiler/factory/nodeFactory.ts:2689:5 | Mutable<ThisTypeNode> | composed-or-forwarding | yes | unproven |
| createTypeOperatorNode at src/compiler/factory/nodeFactory.ts:2696:5 | TypeOperatorNode | composed-or-forwarding | yes | unproven |
| updateTypeOperatorNode at src/compiler/factory/nodeFactory.ts:2707:5 | TypeOperatorNode | composed-or-forwarding | unproven | unproven |
| createIndexedAccessTypeNode at src/compiler/factory/nodeFactory.ts:2714:5 | Mutable<IndexedAccessTypeNode> | composed-or-forwarding | yes | unproven |
| updateIndexedAccessTypeNode at src/compiler/factory/nodeFactory.ts:2723:5 | IndexedAccessTypeNode | composed-or-forwarding | unproven | unproven |
| createMappedTypeNode at src/compiler/factory/nodeFactory.ts:2731:5 | MappedTypeNode | composed-or-forwarding | yes | unproven |
| updateMappedTypeNode at src/compiler/factory/nodeFactory.ts:2747:5 | MappedTypeNode | composed-or-forwarding | unproven | unproven |
| createLiteralTypeNode at src/compiler/factory/nodeFactory.ts:2759:5 | Mutable<LiteralTypeNode> | composed-or-forwarding | yes | unproven |
| updateLiteralTypeNode at src/compiler/factory/nodeFactory.ts:2767:5 | LiteralTypeNode | composed-or-forwarding | unproven | unproven |
| createObjectBindingPattern at src/compiler/factory/nodeFactory.ts:2778:5 | Mutable<ObjectBindingPattern> | composed-or-forwarding | yes | unproven |
| updateObjectBindingPattern at src/compiler/factory/nodeFactory.ts:2792:5 | ObjectBindingPattern | composed-or-forwarding | unproven | unproven |
| createArrayBindingPattern at src/compiler/factory/nodeFactory.ts:2799:5 | Mutable<ArrayBindingPattern> | composed-or-forwarding | yes | unproven |
| updateArrayBindingPattern at src/compiler/factory/nodeFactory.ts:2809:5 | ArrayBindingPattern | composed-or-forwarding | unproven | unproven |
| createBindingElement at src/compiler/factory/nodeFactory.ts:2816:5 | Mutable<BindingElement> | composed-or-forwarding | yes | unproven |
| updateBindingElement at src/compiler/factory/nodeFactory.ts:2834:5 | BindingElement | composed-or-forwarding | unproven | unproven |
| createArrayLiteralExpression at src/compiler/factory/nodeFactory.ts:2848:5 | Mutable<ArrayLiteralExpression> | composed-or-forwarding | yes | unproven |
| updateArrayLiteralExpression at src/compiler/factory/nodeFactory.ts:2862:5 | ArrayLiteralExpression | composed-or-forwarding | unproven | unproven |
| createObjectLiteralExpression at src/compiler/factory/nodeFactory.ts:2869:5 | Mutable<ObjectLiteralExpression> | composed-or-forwarding | yes | unproven |
| updateObjectLiteralExpression at src/compiler/factory/nodeFactory.ts:2880:5 | ObjectLiteralExpression | composed-or-forwarding | unproven | unproven |
| createBasePropertyAccessExpression at src/compiler/factory/nodeFactory.ts:2886:5 | Mutable<PropertyAccessExpression> | composed-or-forwarding | yes | unproven |
| createPropertyAccessExpression at src/compiler/factory/nodeFactory.ts:2903:5 | Mutable<PropertyAccessExpression> | composed-or-forwarding | yes | unproven |
| updatePropertyAccessExpression at src/compiler/factory/nodeFactory.ts:2919:5 | PropertyAccessExpression | composed-or-forwarding | unproven | unproven |
| createPropertyAccessChain at src/compiler/factory/nodeFactory.ts:2930:5 | Mutable<PropertyAccessChain> | composed-or-forwarding | yes | unproven |
| updatePropertyAccessChain at src/compiler/factory/nodeFactory.ts:2942:5 | PropertyAccessChain | composed-or-forwarding | unproven | unproven |
| createBaseElementAccessExpression at src/compiler/factory/nodeFactory.ts:2953:5 | Mutable<ElementAccessExpression> | composed-or-forwarding | yes | unproven |
| createElementAccessExpression at src/compiler/factory/nodeFactory.ts:2968:5 | Mutable<ElementAccessExpression> | composed-or-forwarding | yes | unproven |
| updateElementAccessExpression at src/compiler/factory/nodeFactory.ts:2984:5 | ElementAccessExpression | composed-or-forwarding | unproven | unproven |
| createElementAccessChain at src/compiler/factory/nodeFactory.ts:2995:5 | Mutable<ElementAccessChain> | composed-or-forwarding | yes | unproven |
| updateElementAccessChain at src/compiler/factory/nodeFactory.ts:3007:5 | ElementAccessChain | composed-or-forwarding | unproven | unproven |
| createBaseCallExpression at src/compiler/factory/nodeFactory.ts:3018:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createCallExpression at src/compiler/factory/nodeFactory.ts:3038:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| updateCallExpression at src/compiler/factory/nodeFactory.ts:3052:5 | CallExpression | composed-or-forwarding | unproven | unproven |
| createCallChain at src/compiler/factory/nodeFactory.ts:3064:5 | Mutable<CallChain> | composed-or-forwarding | yes | unproven |
| updateCallChain at src/compiler/factory/nodeFactory.ts:3077:5 | CallChain | composed-or-forwarding | unproven | unproven |
| createNewExpression at src/compiler/factory/nodeFactory.ts:3088:5 | Mutable<NewExpression> | composed-or-forwarding | yes | unproven |
| updateNewExpression at src/compiler/factory/nodeFactory.ts:3104:5 | NewExpression | composed-or-forwarding | unproven | unproven |
| createTaggedTemplateExpression at src/compiler/factory/nodeFactory.ts:3113:5 | Mutable<TaggedTemplateExpression> | composed-or-forwarding | yes | unproven |
| updateTaggedTemplateExpression at src/compiler/factory/nodeFactory.ts:3132:5 | TaggedTemplateExpression | composed-or-forwarding | unproven | unproven |
| createTypeAssertion at src/compiler/factory/nodeFactory.ts:3141:5 | Mutable<TypeAssertion> | composed-or-forwarding | yes | unproven |
| updateTypeAssertion at src/compiler/factory/nodeFactory.ts:3152:5 | TypeAssertion | composed-or-forwarding | unproven | unproven |
| createParenthesizedExpression at src/compiler/factory/nodeFactory.ts:3160:5 | Mutable<ParenthesizedExpression> | composed-or-forwarding | yes | unproven |
| updateParenthesizedExpression at src/compiler/factory/nodeFactory.ts:3170:5 | ParenthesizedExpression | composed-or-forwarding | unproven | unproven |
| createFunctionExpression at src/compiler/factory/nodeFactory.ts:3177:5 | Mutable<FunctionExpression> | composed-or-forwarding | yes | unproven |
| updateFunctionExpression at src/compiler/factory/nodeFactory.ts:3224:5 | FunctionExpression | composed-or-forwarding | unproven | unproven |
| createArrowFunction at src/compiler/factory/nodeFactory.ts:3246:5 | Mutable<ArrowFunction> | composed-or-forwarding | unproven | unproven |
| updateArrowFunction at src/compiler/factory/nodeFactory.ts:3285:5 | ArrowFunction | composed-or-forwarding | unproven | unproven |
| createDeleteExpression at src/compiler/factory/nodeFactory.ts:3305:5 | Mutable<DeleteExpression> | composed-or-forwarding | yes | unproven |
| updateDeleteExpression at src/compiler/factory/nodeFactory.ts:3313:5 | DeleteExpression | composed-or-forwarding | unproven | unproven |
| createTypeOfExpression at src/compiler/factory/nodeFactory.ts:3320:5 | Mutable<TypeOfExpression> | composed-or-forwarding | yes | unproven |
| updateTypeOfExpression at src/compiler/factory/nodeFactory.ts:3328:5 | TypeOfExpression | composed-or-forwarding | unproven | unproven |
| createVoidExpression at src/compiler/factory/nodeFactory.ts:3335:5 | Mutable<VoidExpression> | composed-or-forwarding | yes | unproven |
| updateVoidExpression at src/compiler/factory/nodeFactory.ts:3343:5 | VoidExpression | composed-or-forwarding | unproven | unproven |
| createAwaitExpression at src/compiler/factory/nodeFactory.ts:3350:5 | Mutable<AwaitExpression> | composed-or-forwarding | yes | unproven |
| updateAwaitExpression at src/compiler/factory/nodeFactory.ts:3361:5 | AwaitExpression | composed-or-forwarding | unproven | unproven |
| createPrefixUnaryExpression at src/compiler/factory/nodeFactory.ts:3368:5 | Mutable<PrefixUnaryExpression> | composed-or-forwarding | yes | unproven |
| updatePrefixUnaryExpression at src/compiler/factory/nodeFactory.ts:3387:5 | PrefixUnaryExpression | composed-or-forwarding | unproven | unproven |
| createPostfixUnaryExpression at src/compiler/factory/nodeFactory.ts:3394:5 | Mutable<PostfixUnaryExpression> | composed-or-forwarding | yes | unproven |
| updatePostfixUnaryExpression at src/compiler/factory/nodeFactory.ts:3412:5 | PostfixUnaryExpression | composed-or-forwarding | unproven | unproven |
| createBinaryExpression at src/compiler/factory/nodeFactory.ts:3419:5 | Mutable<BinaryExpression> | composed-or-forwarding | yes | unproven |
| propagateAssignmentPatternFlags at src/compiler/factory/nodeFactory.ts:3459:5 | TransformFlags | non-node-builder | unproven | unproven |
| updateBinaryExpression at src/compiler/factory/nodeFactory.ts:3464:5 | BinaryExpression | composed-or-forwarding | unproven | unproven |
| createConditionalExpression at src/compiler/factory/nodeFactory.ts:3473:5 | Mutable<ConditionalExpression> | composed-or-forwarding | yes | unproven |
| updateConditionalExpression at src/compiler/factory/nodeFactory.ts:3491:5 | ConditionalExpression | composed-or-forwarding | unproven | unproven |
| createTemplateExpression at src/compiler/factory/nodeFactory.ts:3509:5 | Mutable<TemplateExpression> | composed-or-forwarding | yes | unproven |
| updateTemplateExpression at src/compiler/factory/nodeFactory.ts:3520:5 | TemplateExpression | composed-or-forwarding | unproven | unproven |
| createTemplateLiteralLikeToken at src/compiler/factory/nodeFactory.ts:3561:5 | Mutable<TemplateLiteralLikeNode> | composed-or-forwarding | unproven | unproven |
| createTemplateLiteralLikeDeclaration at src/compiler/factory/nodeFactory.ts:3570:5 | Mutable<NoSubstitutionTemplateLiteral> | composed-or-forwarding | yes | unproven |
| createTemplateLiteralLikeNode at src/compiler/factory/nodeFactory.ts:3580:5 | Mutable<TemplateLiteralLikeNode> | composed-or-forwarding | unproven | unproven |
| createTemplateHead at src/compiler/factory/nodeFactory.ts:3588:5 | TemplateHead | composed-or-forwarding | unproven | unproven |
| createTemplateMiddle at src/compiler/factory/nodeFactory.ts:3594:5 | TemplateMiddle | composed-or-forwarding | unproven | unproven |
| createTemplateTail at src/compiler/factory/nodeFactory.ts:3600:5 | TemplateTail | composed-or-forwarding | unproven | unproven |
| createNoSubstitutionTemplateLiteral at src/compiler/factory/nodeFactory.ts:3606:5 | NoSubstitutionTemplateLiteral | composed-or-forwarding | yes | unproven |
| createYieldExpression at src/compiler/factory/nodeFactory.ts:3612:5 | YieldExpression | composed-or-forwarding | yes | unproven |
| updateYieldExpression at src/compiler/factory/nodeFactory.ts:3626:5 | YieldExpression | composed-or-forwarding | unproven | unproven |
| createSpreadElement at src/compiler/factory/nodeFactory.ts:3634:5 | Mutable<SpreadElement> | composed-or-forwarding | yes | unproven |
| updateSpreadElement at src/compiler/factory/nodeFactory.ts:3644:5 | SpreadElement | composed-or-forwarding | unproven | unproven |
| createClassExpression at src/compiler/factory/nodeFactory.ts:3651:5 | Mutable<ClassExpression> | composed-or-forwarding | yes | unproven |
| updateClassExpression at src/compiler/factory/nodeFactory.ts:3677:5 | ClassExpression | composed-or-forwarding | unproven | unproven |
| createOmittedExpression at src/compiler/factory/nodeFactory.ts:3695:5 | Mutable<OmittedExpression> | composed-or-forwarding | yes | unproven |
| createExpressionWithTypeArguments at src/compiler/factory/nodeFactory.ts:3700:5 | Mutable<ExpressionWithTypeArguments> | composed-or-forwarding | yes | unproven |
| updateExpressionWithTypeArguments at src/compiler/factory/nodeFactory.ts:3711:5 | ExpressionWithTypeArguments | composed-or-forwarding | unproven | unproven |
| createAsExpression at src/compiler/factory/nodeFactory.ts:3719:5 | Mutable<AsExpression> | composed-or-forwarding | yes | unproven |
| updateAsExpression at src/compiler/factory/nodeFactory.ts:3730:5 | AsExpression | composed-or-forwarding | unproven | unproven |
| createNonNullExpression at src/compiler/factory/nodeFactory.ts:3738:5 | Mutable<NonNullExpression> | composed-or-forwarding | yes | unproven |
| updateNonNullExpression at src/compiler/factory/nodeFactory.ts:3747:5 | NonNullExpression | composed-or-forwarding | unproven | unproven |
| createSatisfiesExpression at src/compiler/factory/nodeFactory.ts:3757:5 | Mutable<SatisfiesExpression> | composed-or-forwarding | yes | unproven |
| updateSatisfiesExpression at src/compiler/factory/nodeFactory.ts:3768:5 | SatisfiesExpression | composed-or-forwarding | unproven | unproven |
| createNonNullChain at src/compiler/factory/nodeFactory.ts:3776:5 | Mutable<NonNullChain> | composed-or-forwarding | yes | unproven |
| updateNonNullChain at src/compiler/factory/nodeFactory.ts:3786:5 | NonNullChain | composed-or-forwarding | unproven | unproven |
| createMetaProperty at src/compiler/factory/nodeFactory.ts:3794:5 | Mutable<MetaProperty> | composed-or-forwarding | unproven | unproven |
| updateMetaProperty at src/compiler/factory/nodeFactory.ts:3815:5 | MetaProperty | composed-or-forwarding | unproven | unproven |
| createTemplateSpan at src/compiler/factory/nodeFactory.ts:3826:5 | Mutable<TemplateSpan> | composed-or-forwarding | yes | unproven |
| updateTemplateSpan at src/compiler/factory/nodeFactory.ts:3837:5 | TemplateSpan | composed-or-forwarding | unproven | unproven |
| createSemicolonClassElement at src/compiler/factory/nodeFactory.ts:3845:5 | Mutable<SemicolonClassElement> | composed-or-forwarding | unproven | unproven |
| createBlock at src/compiler/factory/nodeFactory.ts:3856:5 | Block | composed-or-forwarding | yes | unproven |
| updateBlock at src/compiler/factory/nodeFactory.ts:3869:5 | Block | composed-or-forwarding | unproven | unproven |
| createVariableStatement at src/compiler/factory/nodeFactory.ts:3876:5 | Mutable<VariableStatement> | composed-or-forwarding | yes | unproven |
| updateVariableStatement at src/compiler/factory/nodeFactory.ts:3892:5 | VariableStatement | composed-or-forwarding | unproven | unproven |
| createEmptyStatement at src/compiler/factory/nodeFactory.ts:3900:5 | Mutable<EmptyStatement> | composed-or-forwarding | yes | unproven |
| createExpressionStatement at src/compiler/factory/nodeFactory.ts:3907:5 | ExpressionStatement | composed-or-forwarding | yes | unproven |
| updateExpressionStatement at src/compiler/factory/nodeFactory.ts:3918:5 | ExpressionStatement | composed-or-forwarding | unproven | unproven |
| createIfStatement at src/compiler/factory/nodeFactory.ts:3925:5 | Mutable<IfStatement> | composed-or-forwarding | yes | unproven |
| updateIfStatement at src/compiler/factory/nodeFactory.ts:3940:5 | IfStatement | composed-or-forwarding | unproven | unproven |
| createDoStatement at src/compiler/factory/nodeFactory.ts:3949:5 | Mutable<DoStatement> | composed-or-forwarding | yes | unproven |
| updateDoStatement at src/compiler/factory/nodeFactory.ts:3962:5 | DoStatement | composed-or-forwarding | unproven | unproven |
| createWhileStatement at src/compiler/factory/nodeFactory.ts:3970:5 | Mutable<WhileStatement> | composed-or-forwarding | yes | unproven |
| updateWhileStatement at src/compiler/factory/nodeFactory.ts:3983:5 | WhileStatement | composed-or-forwarding | unproven | unproven |
| createForStatement at src/compiler/factory/nodeFactory.ts:3991:5 | Mutable<ForStatement> | composed-or-forwarding | yes | unproven |
| updateForStatement at src/compiler/factory/nodeFactory.ts:4010:5 | ForStatement | composed-or-forwarding | unproven | unproven |
| createForInStatement at src/compiler/factory/nodeFactory.ts:4020:5 | Mutable<ForInStatement> | composed-or-forwarding | yes | unproven |
| updateForInStatement at src/compiler/factory/nodeFactory.ts:4037:5 | ForInStatement | composed-or-forwarding | unproven | unproven |
| createForOfStatement at src/compiler/factory/nodeFactory.ts:4046:5 | Mutable<ForOfStatement> | composed-or-forwarding | yes | unproven |
| updateForOfStatement at src/compiler/factory/nodeFactory.ts:4067:5 | ForOfStatement | composed-or-forwarding | unproven | unproven |
| createContinueStatement at src/compiler/factory/nodeFactory.ts:4077:5 | ContinueStatement | composed-or-forwarding | yes | unproven |
| updateContinueStatement at src/compiler/factory/nodeFactory.ts:4089:5 | ContinueStatement | composed-or-forwarding | unproven | unproven |
| createBreakStatement at src/compiler/factory/nodeFactory.ts:4096:5 | BreakStatement | composed-or-forwarding | yes | unproven |
| updateBreakStatement at src/compiler/factory/nodeFactory.ts:4108:5 | BreakStatement | composed-or-forwarding | unproven | unproven |
| createReturnStatement at src/compiler/factory/nodeFactory.ts:4115:5 | ReturnStatement | composed-or-forwarding | yes | unproven |
| updateReturnStatement at src/compiler/factory/nodeFactory.ts:4129:5 | ReturnStatement | composed-or-forwarding | unproven | unproven |
| createWithStatement at src/compiler/factory/nodeFactory.ts:4136:5 | Mutable<WithStatement> | composed-or-forwarding | yes | unproven |
| updateWithStatement at src/compiler/factory/nodeFactory.ts:4149:5 | WithStatement | composed-or-forwarding | unproven | unproven |
| createSwitchStatement at src/compiler/factory/nodeFactory.ts:4157:5 | SwitchStatement | composed-or-forwarding | yes | unproven |
| updateSwitchStatement at src/compiler/factory/nodeFactory.ts:4171:5 | SwitchStatement | composed-or-forwarding | unproven | unproven |
| createLabeledStatement at src/compiler/factory/nodeFactory.ts:4179:5 | Mutable<LabeledStatement> | composed-or-forwarding | yes | unproven |
| updateLabeledStatement at src/compiler/factory/nodeFactory.ts:4192:5 | LabeledStatement | composed-or-forwarding | unproven | unproven |
| createThrowStatement at src/compiler/factory/nodeFactory.ts:4200:5 | Mutable<ThrowStatement> | composed-or-forwarding | yes | unproven |
| updateThrowStatement at src/compiler/factory/nodeFactory.ts:4211:5 | ThrowStatement | composed-or-forwarding | unproven | unproven |
| createTryStatement at src/compiler/factory/nodeFactory.ts:4218:5 | Mutable<TryStatement> | composed-or-forwarding | yes | unproven |
| updateTryStatement at src/compiler/factory/nodeFactory.ts:4233:5 | TryStatement | composed-or-forwarding | unproven | unproven |
| createDebuggerStatement at src/compiler/factory/nodeFactory.ts:4242:5 | Mutable<DebuggerStatement> | composed-or-forwarding | yes | unproven |
| createVariableDeclaration at src/compiler/factory/nodeFactory.ts:4251:5 | Mutable<VariableDeclaration> | composed-or-forwarding | yes | unproven |
| updateVariableDeclaration at src/compiler/factory/nodeFactory.ts:4266:5 | VariableDeclaration | composed-or-forwarding | unproven | unproven |
| createVariableDeclarationList at src/compiler/factory/nodeFactory.ts:4276:5 | Mutable<VariableDeclarationList> | composed-or-forwarding | yes | unproven |
| updateVariableDeclarationList at src/compiler/factory/nodeFactory.ts:4293:5 | VariableDeclarationList | composed-or-forwarding | unproven | unproven |
| createFunctionDeclaration at src/compiler/factory/nodeFactory.ts:4300:5 | Mutable<FunctionDeclaration> | composed-or-forwarding | yes | unproven |
| updateFunctionDeclaration at src/compiler/factory/nodeFactory.ts:4351:5 | FunctionDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateFunctionDeclaration at src/compiler/factory/nodeFactory.ts:4372:5 | FunctionDeclaration | composed-or-forwarding | unproven | unproven |
| createClassDeclaration at src/compiler/factory/nodeFactory.ts:4383:5 | Mutable<ClassDeclaration> | composed-or-forwarding | yes | unproven |
| updateClassDeclaration at src/compiler/factory/nodeFactory.ts:4418:5 | ClassDeclaration | composed-or-forwarding | unproven | unproven |
| createInterfaceDeclaration at src/compiler/factory/nodeFactory.ts:4436:5 | Mutable<InterfaceDeclaration> | composed-or-forwarding | yes | unproven |
| updateInterfaceDeclaration at src/compiler/factory/nodeFactory.ts:4456:5 | InterfaceDeclaration | composed-or-forwarding | unproven | unproven |
| createTypeAliasDeclaration at src/compiler/factory/nodeFactory.ts:4474:5 | Mutable<TypeAliasDeclaration> | composed-or-forwarding | yes | unproven |
| updateTypeAliasDeclaration at src/compiler/factory/nodeFactory.ts:4494:5 | TypeAliasDeclaration | composed-or-forwarding | unproven | unproven |
| createEnumDeclaration at src/compiler/factory/nodeFactory.ts:4510:5 | Mutable<EnumDeclaration> | composed-or-forwarding | yes | unproven |
| updateEnumDeclaration at src/compiler/factory/nodeFactory.ts:4530:5 | EnumDeclaration | composed-or-forwarding | unproven | unproven |
| createModuleDeclaration at src/compiler/factory/nodeFactory.ts:4544:5 | Mutable<ModuleDeclaration> | composed-or-forwarding | yes | unproven |
| updateModuleDeclaration at src/compiler/factory/nodeFactory.ts:4573:5 | ModuleDeclaration | composed-or-forwarding | unproven | unproven |
| createModuleBlock at src/compiler/factory/nodeFactory.ts:4587:5 | Mutable<ModuleBlock> | composed-or-forwarding | yes | unproven |
| updateModuleBlock at src/compiler/factory/nodeFactory.ts:4597:5 | ModuleBlock | composed-or-forwarding | unproven | unproven |
| createCaseBlock at src/compiler/factory/nodeFactory.ts:4604:5 | CaseBlock | composed-or-forwarding | yes | unproven |
| updateCaseBlock at src/compiler/factory/nodeFactory.ts:4615:5 | CaseBlock | composed-or-forwarding | unproven | unproven |
| createNamespaceExportDeclaration at src/compiler/factory/nodeFactory.ts:4622:5 | Mutable<NamespaceExportDeclaration> | composed-or-forwarding | yes | unproven |
| updateNamespaceExportDeclaration at src/compiler/factory/nodeFactory.ts:4634:5 | NamespaceExportDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateNamespaceExportDeclaration at src/compiler/factory/nodeFactory.ts:4640:5 | NamespaceExportDeclaration | composed-or-forwarding | unproven | unproven |
| createImportEqualsDeclaration at src/compiler/factory/nodeFactory.ts:4649:5 | Mutable<ImportEqualsDeclaration> | composed-or-forwarding | yes | unproven |
| updateImportEqualsDeclaration at src/compiler/factory/nodeFactory.ts:4675:5 | ImportEqualsDeclaration | composed-or-forwarding | unproven | unproven |
| createImportDeclaration at src/compiler/factory/nodeFactory.ts:4691:5 | ImportDeclaration | composed-or-forwarding | yes | unproven |
| updateImportDeclaration at src/compiler/factory/nodeFactory.ts:4711:5 | ImportDeclaration | composed-or-forwarding | unproven | unproven |
| createImportClause at src/compiler/factory/nodeFactory.ts:4727:5 | ImportClause | composed-or-forwarding | yes | unproven |
| updateImportClause at src/compiler/factory/nodeFactory.ts:4746:5 | ImportClause | composed-or-forwarding | unproven | unproven |
| createAssertClause at src/compiler/factory/nodeFactory.ts:4758:5 | AssertClause | composed-or-forwarding | yes | unproven |
| updateAssertClause at src/compiler/factory/nodeFactory.ts:4768:5 | AssertClause | composed-or-forwarding | unproven | unproven |
| createAssertEntry at src/compiler/factory/nodeFactory.ts:4776:5 | AssertEntry | composed-or-forwarding | yes | unproven |
| updateAssertEntry at src/compiler/factory/nodeFactory.ts:4785:5 | AssertEntry | composed-or-forwarding | unproven | unproven |
| createImportTypeAssertionContainer at src/compiler/factory/nodeFactory.ts:4793:5 | ImportTypeAssertionContainer | composed-or-forwarding | yes | unproven |
| updateImportTypeAssertionContainer at src/compiler/factory/nodeFactory.ts:4801:5 | ImportTypeAssertionContainer | composed-or-forwarding | unproven | unproven |
| createImportAttributes at src/compiler/factory/nodeFactory.ts:4811:5 | ImportAttributes | composed-or-forwarding | yes | unproven |
| updateImportAttributes at src/compiler/factory/nodeFactory.ts:4821:5 | ImportAttributes | composed-or-forwarding | unproven | unproven |
| createImportAttribute at src/compiler/factory/nodeFactory.ts:4829:5 | ImportAttribute | composed-or-forwarding | yes | unproven |
| updateImportAttribute at src/compiler/factory/nodeFactory.ts:4838:5 | ImportAttribute | composed-or-forwarding | unproven | unproven |
| createNamespaceImport at src/compiler/factory/nodeFactory.ts:4846:5 | NamespaceImport | composed-or-forwarding | yes | unproven |
| updateNamespaceImport at src/compiler/factory/nodeFactory.ts:4855:5 | NamespaceImport | composed-or-forwarding | unproven | unproven |
| createNamespaceExport at src/compiler/factory/nodeFactory.ts:4862:5 | NamespaceExport | composed-or-forwarding | yes | unproven |
| updateNamespaceExport at src/compiler/factory/nodeFactory.ts:4872:5 | NamespaceExport | composed-or-forwarding | unproven | unproven |
| createNamedImports at src/compiler/factory/nodeFactory.ts:4879:5 | NamedImports | composed-or-forwarding | yes | unproven |
| updateNamedImports at src/compiler/factory/nodeFactory.ts:4888:5 | NamedImports | composed-or-forwarding | unproven | unproven |
| createImportSpecifier at src/compiler/factory/nodeFactory.ts:4895:5 | Mutable<ImportSpecifier> | composed-or-forwarding | yes | unproven |
| updateImportSpecifier at src/compiler/factory/nodeFactory.ts:4907:5 | ImportSpecifier | composed-or-forwarding | unproven | unproven |
| createExportAssignment at src/compiler/factory/nodeFactory.ts:4916:5 | Mutable<ExportAssignment> | composed-or-forwarding | yes | unproven |
| updateExportAssignment at src/compiler/factory/nodeFactory.ts:4935:5 | ExportAssignment | composed-or-forwarding | unproven | unproven |
| createExportDeclaration at src/compiler/factory/nodeFactory.ts:4947:5 | Mutable<ExportDeclaration> | composed-or-forwarding | yes | unproven |
| updateExportDeclaration at src/compiler/factory/nodeFactory.ts:4970:5 | ExportDeclaration | composed-or-forwarding | unproven | unproven |
| finishUpdateExportDeclaration at src/compiler/factory/nodeFactory.ts:4987:5 | ExportDeclaration | composed-or-forwarding | unproven | unproven |
| createNamedExports at src/compiler/factory/nodeFactory.ts:4998:5 | Mutable<NamedExports> | composed-or-forwarding | yes | unproven |
| updateNamedExports at src/compiler/factory/nodeFactory.ts:5007:5 | NamedExports | composed-or-forwarding | unproven | unproven |
| createExportSpecifier at src/compiler/factory/nodeFactory.ts:5014:5 | Mutable<ExportSpecifier> | composed-or-forwarding | unproven | unproven |
| updateExportSpecifier at src/compiler/factory/nodeFactory.ts:5028:5 | ExportSpecifier | composed-or-forwarding | unproven | unproven |
| createMissingDeclaration at src/compiler/factory/nodeFactory.ts:5037:5 | MissingDeclaration | composed-or-forwarding | yes | unproven |
| createExternalModuleReference at src/compiler/factory/nodeFactory.ts:5049:5 | Mutable<ExternalModuleReference> | composed-or-forwarding | yes | unproven |
| updateExternalModuleReference at src/compiler/factory/nodeFactory.ts:5058:5 | ExternalModuleReference | composed-or-forwarding | unproven | unproven |
| createJSDocPrimaryTypeWorker at src/compiler/factory/nodeFactory.ts:5071:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| createJSDocPrePostfixUnaryTypeWorker at src/compiler/factory/nodeFactory.ts:5078:5 | T | composed-or-forwarding | yes | unproven |
| createJSDocUnaryTypeWorker at src/compiler/factory/nodeFactory.ts:5091:5 | T | composed-or-forwarding | yes | unproven |
| updateJSDocPrePostfixUnaryTypeWorker at src/compiler/factory/nodeFactory.ts:5100:5 | T | composed-or-forwarding | unproven | unproven |
| updateJSDocUnaryTypeWorker at src/compiler/factory/nodeFactory.ts:5110:5 | T | composed-or-forwarding | unproven | unproven |
| createJSDocFunctionType at src/compiler/factory/nodeFactory.ts:5117:5 | JSDocFunctionType | composed-or-forwarding | yes | unproven |
| updateJSDocFunctionType at src/compiler/factory/nodeFactory.ts:5132:5 | JSDocFunctionType | composed-or-forwarding | unproven | unproven |
| createJSDocTypeLiteral at src/compiler/factory/nodeFactory.ts:5140:5 | JSDocTypeLiteral | composed-or-forwarding | yes | unproven |
| updateJSDocTypeLiteral at src/compiler/factory/nodeFactory.ts:5148:5 | JSDocTypeLiteral | composed-or-forwarding | unproven | unproven |
| createJSDocTypeExpression at src/compiler/factory/nodeFactory.ts:5156:5 | JSDocTypeExpression | composed-or-forwarding | yes | unproven |
| updateJSDocTypeExpression at src/compiler/factory/nodeFactory.ts:5163:5 | JSDocTypeExpression | composed-or-forwarding | unproven | unproven |
| createJSDocSignature at src/compiler/factory/nodeFactory.ts:5170:5 | JSDocSignature | composed-or-forwarding | yes | unproven |
| updateJSDocSignature at src/compiler/factory/nodeFactory.ts:5183:5 | JSDocSignature | composed-or-forwarding | unproven | unproven |
| getDefaultTagName at src/compiler/factory/nodeFactory.ts:5191:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createBaseJSDocTag at src/compiler/factory/nodeFactory.ts:5199:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| createBaseJSDocTagDeclaration at src/compiler/factory/nodeFactory.ts:5206:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| createJSDocTemplateTag at src/compiler/factory/nodeFactory.ts:5214:5 | JSDocTemplateTag | composed-or-forwarding | yes | unproven |
| updateJSDocTemplateTag at src/compiler/factory/nodeFactory.ts:5222:5 | JSDocTemplateTag | composed-or-forwarding | unproven | unproven |
| createJSDocTypedefTag at src/compiler/factory/nodeFactory.ts:5232:5 | JSDocTypedefTag | composed-or-forwarding | yes | unproven |
| updateJSDocTypedefTag at src/compiler/factory/nodeFactory.ts:5244:5 | JSDocTypedefTag | composed-or-forwarding | unproven | unproven |
| createJSDocParameterTag at src/compiler/factory/nodeFactory.ts:5254:5 | JSDocParameterTag | composed-or-forwarding | yes | unproven |
| updateJSDocParameterTag at src/compiler/factory/nodeFactory.ts:5264:5 | JSDocParameterTag | composed-or-forwarding | unproven | unproven |
| createJSDocPropertyTag at src/compiler/factory/nodeFactory.ts:5276:5 | JSDocPropertyTag | composed-or-forwarding | yes | unproven |
| updateJSDocPropertyTag at src/compiler/factory/nodeFactory.ts:5286:5 | JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| createJSDocCallbackTag at src/compiler/factory/nodeFactory.ts:5298:5 | JSDocCallbackTag | composed-or-forwarding | yes | unproven |
| updateJSDocCallbackTag at src/compiler/factory/nodeFactory.ts:5310:5 | JSDocCallbackTag | composed-or-forwarding | unproven | unproven |
| createJSDocOverloadTag at src/compiler/factory/nodeFactory.ts:5320:5 | JSDocOverloadTag | composed-or-forwarding | yes | unproven |
| updateJSDocOverloadTag at src/compiler/factory/nodeFactory.ts:5327:5 | JSDocOverloadTag | composed-or-forwarding | unproven | unproven |
| createJSDocAugmentsTag at src/compiler/factory/nodeFactory.ts:5336:5 | JSDocAugmentsTag | composed-or-forwarding | yes | unproven |
| updateJSDocAugmentsTag at src/compiler/factory/nodeFactory.ts:5343:5 | JSDocAugmentsTag | composed-or-forwarding | unproven | unproven |
| createJSDocImplementsTag at src/compiler/factory/nodeFactory.ts:5352:5 | JSDocImplementsTag | composed-or-forwarding | yes | unproven |
| createJSDocSeeTag at src/compiler/factory/nodeFactory.ts:5359:5 | JSDocSeeTag | composed-or-forwarding | yes | unproven |
| updateJSDocSeeTag at src/compiler/factory/nodeFactory.ts:5366:5 | JSDocSeeTag | composed-or-forwarding | unproven | unproven |
| createJSDocNameReference at src/compiler/factory/nodeFactory.ts:5375:5 | JSDocNameReference | composed-or-forwarding | yes | unproven |
| updateJSDocNameReference at src/compiler/factory/nodeFactory.ts:5382:5 | JSDocNameReference | composed-or-forwarding | unproven | unproven |
| createJSDocMemberName at src/compiler/factory/nodeFactory.ts:5389:5 | Mutable<JSDocMemberName> | composed-or-forwarding | yes | unproven |
| updateJSDocMemberName at src/compiler/factory/nodeFactory.ts:5399:5 | JSDocMemberName | composed-or-forwarding | unproven | unproven |
| createJSDocLink at src/compiler/factory/nodeFactory.ts:5407:5 | JSDocLink | composed-or-forwarding | yes | unproven |
| updateJSDocLink at src/compiler/factory/nodeFactory.ts:5415:5 | JSDocLink | composed-or-forwarding | unproven | unproven |
| createJSDocLinkCode at src/compiler/factory/nodeFactory.ts:5422:5 | JSDocLinkCode | composed-or-forwarding | yes | unproven |
| updateJSDocLinkCode at src/compiler/factory/nodeFactory.ts:5430:5 | JSDocLinkCode | composed-or-forwarding | unproven | unproven |
| createJSDocLinkPlain at src/compiler/factory/nodeFactory.ts:5437:5 | JSDocLinkPlain | composed-or-forwarding | yes | unproven |
| updateJSDocLinkPlain at src/compiler/factory/nodeFactory.ts:5445:5 | JSDocLinkPlain | composed-or-forwarding | unproven | unproven |
| updateJSDocImplementsTag at src/compiler/factory/nodeFactory.ts:5452:5 | JSDocImplementsTag | composed-or-forwarding | unproven | unproven |
| createJSDocSimpleTagWorker at src/compiler/factory/nodeFactory.ts:5468:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| updateJSDocSimpleTagWorker at src/compiler/factory/nodeFactory.ts:5481:5 | T | composed-or-forwarding | unproven | unproven |
| createJSDocTypeLikeTagWorker at src/compiler/factory/nodeFactory.ts:5494:5 | Mutable<T> | composed-or-forwarding | yes | unproven |
| updateJSDocTypeLikeTagWorker at src/compiler/factory/nodeFactory.ts:5506:5 | T | composed-or-forwarding | unproven | unproven |
| createJSDocUnknownTag at src/compiler/factory/nodeFactory.ts:5515:5 | JSDocUnknownTag | composed-or-forwarding | yes | unproven |
| updateJSDocUnknownTag at src/compiler/factory/nodeFactory.ts:5521:5 | JSDocUnknownTag | composed-or-forwarding | unproven | unproven |
| createJSDocEnumTag at src/compiler/factory/nodeFactory.ts:5529:5 | Mutable<JSDocEnumTag> | composed-or-forwarding | yes | unproven |
| updateJSDocEnumTag at src/compiler/factory/nodeFactory.ts:5539:5 | JSDocEnumTag | composed-or-forwarding | unproven | unproven |
| createJSDocImportTag at src/compiler/factory/nodeFactory.ts:5548:5 | JSDocImportTag | composed-or-forwarding | yes | unproven |
| updateJSDocImportTag at src/compiler/factory/nodeFactory.ts:5557:5 | JSDocImportTag | composed-or-forwarding | unproven | unproven |
| createJSDocText at src/compiler/factory/nodeFactory.ts:5568:5 | JSDocText | composed-or-forwarding | yes | unproven |
| updateJSDocText at src/compiler/factory/nodeFactory.ts:5575:5 | JSDocText | composed-or-forwarding | unproven | unproven |
| createJSDocComment at src/compiler/factory/nodeFactory.ts:5582:5 | Mutable<JSDoc> | composed-or-forwarding | yes | unproven |
| updateJSDocComment at src/compiler/factory/nodeFactory.ts:5590:5 | JSDoc | composed-or-forwarding | unproven | unproven |
| createJsxElement at src/compiler/factory/nodeFactory.ts:5602:5 | Mutable<JsxElement> | composed-or-forwarding | yes | unproven |
| updateJsxElement at src/compiler/factory/nodeFactory.ts:5615:5 | JsxElement | composed-or-forwarding | unproven | unproven |
| createJsxSelfClosingElement at src/compiler/factory/nodeFactory.ts:5624:5 | Mutable<JsxSelfClosingElement> | composed-or-forwarding | yes | unproven |
| updateJsxSelfClosingElement at src/compiler/factory/nodeFactory.ts:5640:5 | JsxSelfClosingElement | composed-or-forwarding | unproven | unproven |
| createJsxOpeningElement at src/compiler/factory/nodeFactory.ts:5649:5 | Mutable<JsxOpeningElement> | composed-or-forwarding | yes | unproven |
| updateJsxOpeningElement at src/compiler/factory/nodeFactory.ts:5665:5 | JsxOpeningElement | composed-or-forwarding | unproven | unproven |
| createJsxClosingElement at src/compiler/factory/nodeFactory.ts:5674:5 | Mutable<JsxClosingElement> | composed-or-forwarding | yes | unproven |
| updateJsxClosingElement at src/compiler/factory/nodeFactory.ts:5683:5 | JsxClosingElement | composed-or-forwarding | unproven | unproven |
| createJsxFragment at src/compiler/factory/nodeFactory.ts:5690:5 | Mutable<JsxFragment> | composed-or-forwarding | yes | unproven |
| updateJsxFragment at src/compiler/factory/nodeFactory.ts:5703:5 | JsxFragment | composed-or-forwarding | unproven | unproven |
| createJsxText at src/compiler/factory/nodeFactory.ts:5712:5 | Mutable<JsxText> | composed-or-forwarding | yes | unproven |
| updateJsxText at src/compiler/factory/nodeFactory.ts:5721:5 | JsxText | composed-or-forwarding | unproven | unproven |
| createJsxOpeningFragment at src/compiler/factory/nodeFactory.ts:5729:5 | Mutable<JsxOpeningFragment> | composed-or-forwarding | yes | unproven |
| createJsxJsxClosingFragment at src/compiler/factory/nodeFactory.ts:5736:5 | Mutable<JsxClosingFragment> | composed-or-forwarding | yes | unproven |
| createJsxAttribute at src/compiler/factory/nodeFactory.ts:5743:5 | Mutable<JsxAttribute> | composed-or-forwarding | yes | unproven |
| updateJsxAttribute at src/compiler/factory/nodeFactory.ts:5754:5 | JsxAttribute | composed-or-forwarding | unproven | unproven |
| createJsxAttributes at src/compiler/factory/nodeFactory.ts:5762:5 | Mutable<JsxAttributes> | composed-or-forwarding | yes | unproven |
| updateJsxAttributes at src/compiler/factory/nodeFactory.ts:5771:5 | JsxAttributes | composed-or-forwarding | unproven | unproven |
| createJsxSpreadAttribute at src/compiler/factory/nodeFactory.ts:5778:5 | Mutable<JsxSpreadAttribute> | composed-or-forwarding | unproven | unproven |
| updateJsxSpreadAttribute at src/compiler/factory/nodeFactory.ts:5787:5 | JsxSpreadAttribute | composed-or-forwarding | unproven | unproven |
| createJsxExpression at src/compiler/factory/nodeFactory.ts:5794:5 | Mutable<JsxExpression> | composed-or-forwarding | yes | unproven |
| updateJsxExpression at src/compiler/factory/nodeFactory.ts:5805:5 | JsxExpression | composed-or-forwarding | unproven | unproven |
| createJsxNamespacedName at src/compiler/factory/nodeFactory.ts:5812:5 | Mutable<JsxNamespacedName> | composed-or-forwarding | yes | unproven |
| updateJsxNamespacedName at src/compiler/factory/nodeFactory.ts:5823:5 | JsxNamespacedName | composed-or-forwarding | unproven | unproven |
| createCaseClause at src/compiler/factory/nodeFactory.ts:5835:5 | Mutable<CaseClause> | composed-or-forwarding | yes | unproven |
| updateCaseClause at src/compiler/factory/nodeFactory.ts:5847:5 | CaseClause | composed-or-forwarding | unproven | unproven |
| createDefaultClause at src/compiler/factory/nodeFactory.ts:5855:5 | Mutable<DefaultClause> | composed-or-forwarding | yes | unproven |
| updateDefaultClause at src/compiler/factory/nodeFactory.ts:5863:5 | DefaultClause | composed-or-forwarding | unproven | unproven |
| createHeritageClause at src/compiler/factory/nodeFactory.ts:5870:5 | Mutable<HeritageClause> | composed-or-forwarding | unproven | unproven |
| updateHeritageClause at src/compiler/factory/nodeFactory.ts:5889:5 | HeritageClause | composed-or-forwarding | unproven | unproven |
| createCatchClause at src/compiler/factory/nodeFactory.ts:5896:5 | Mutable<CatchClause> | composed-or-forwarding | yes | unproven |
| updateCatchClause at src/compiler/factory/nodeFactory.ts:5911:5 | CatchClause | composed-or-forwarding | unproven | unproven |
| createPropertyAssignment at src/compiler/factory/nodeFactory.ts:5923:5 | Mutable<PropertyAssignment> | composed-or-forwarding | yes | unproven |
| updatePropertyAssignment at src/compiler/factory/nodeFactory.ts:5938:5 | PropertyAssignment | composed-or-forwarding | unproven | unproven |
| finishUpdatePropertyAssignment at src/compiler/factory/nodeFactory.ts:5945:5 | PropertyAssignment | composed-or-forwarding | unproven | unproven |
| createShorthandPropertyAssignment at src/compiler/factory/nodeFactory.ts:5957:5 | Mutable<ShorthandPropertyAssignment> | composed-or-forwarding | yes | unproven |
| updateShorthandPropertyAssignment at src/compiler/factory/nodeFactory.ts:5974:5 | ShorthandPropertyAssignment | composed-or-forwarding | unproven | unproven |
| finishUpdateShorthandPropertyAssignment at src/compiler/factory/nodeFactory.ts:5981:5 | ShorthandPropertyAssignment | composed-or-forwarding | unproven | unproven |
| createSpreadAssignment at src/compiler/factory/nodeFactory.ts:5993:5 | Mutable<SpreadAssignment> | composed-or-forwarding | yes | unproven |
| updateSpreadAssignment at src/compiler/factory/nodeFactory.ts:6005:5 | SpreadAssignment | composed-or-forwarding | unproven | unproven |
| createEnumMember at src/compiler/factory/nodeFactory.ts:6016:5 | Mutable<EnumMember> | composed-or-forwarding | yes | unproven |
| updateEnumMember at src/compiler/factory/nodeFactory.ts:6029:5 | EnumMember | composed-or-forwarding | unproven | unproven |
| createSourceFile at src/compiler/factory/nodeFactory.ts:6041:5 | Mutable<SourceFile> | allocating | unproven | unproven |
| createRedirectedSourceFile at src/compiler/factory/nodeFactory.ts:6095:5 | SourceFile | input-or-unresolved | unproven | unproven |
| get at src/compiler/factory/nodeFactory.ts:6099:17 | number &#124; undefined | non-node-builder | unproven | unproven |
| set at src/compiler/factory/nodeFactory.ts:6102:17 | void | non-node-builder | unproven | unproven |
| get at src/compiler/factory/nodeFactory.ts:6107:17 | Symbol | non-node-builder | unproven | unproven |
| set at src/compiler/factory/nodeFactory.ts:6110:17 | void | non-node-builder | unproven | unproven |
| cloneRedirectedSourceFile at src/compiler/factory/nodeFactory.ts:6119:5 | Mutable<SourceFile> | composed-or-forwarding | unproven | unproven |
| cloneSourceFileWorker at src/compiler/factory/nodeFactory.ts:6132:5 | Mutable<SourceFile> | allocating | unproven | unproven |
| cloneSourceFile at src/compiler/factory/nodeFactory.ts:6150:5 | Mutable<SourceFile> | composed-or-forwarding | unproven | unproven |
| cloneSourceFileWithChanges at src/compiler/factory/nodeFactory.ts:6156:5 | Mutable<SourceFile> | composed-or-forwarding | unproven | unproven |
| updateSourceFile at src/compiler/factory/nodeFactory.ts:6178:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| createBundle at src/compiler/factory/nodeFactory.ts:6197:5 | Mutable<Bundle> | composed-or-forwarding | yes | unproven |
| updateBundle at src/compiler/factory/nodeFactory.ts:6207:5 | Bundle | composed-or-forwarding | unproven | unproven |
| createSyntheticExpression at src/compiler/factory/nodeFactory.ts:6218:5 | Mutable<SyntheticExpression> | composed-or-forwarding | yes | unproven |
| createSyntaxList at src/compiler/factory/nodeFactory.ts:6227:5 | Mutable<SyntaxList> | composed-or-forwarding | yes | unproven |
| createNotEmittedStatement at src/compiler/factory/nodeFactory.ts:6244:5 | Mutable<NotEmittedStatement> | composed-or-forwarding | yes | unproven |
| createPartiallyEmittedExpression at src/compiler/factory/nodeFactory.ts:6259:5 | Mutable<PartiallyEmittedExpression> | composed-or-forwarding | yes | unproven |
| updatePartiallyEmittedExpression at src/compiler/factory/nodeFactory.ts:6270:5 | PartiallyEmittedExpression | composed-or-forwarding | unproven | unproven |
| createNotEmittedTypeElement at src/compiler/factory/nodeFactory.ts:6277:5 | Mutable<NotEmittedTypeElement> | composed-or-forwarding | unproven | unproven |
| flattenCommaElements at src/compiler/factory/nodeFactory.ts:6281:5 | Expression &#124; readonly Expression[] | input-or-unresolved | unproven | unproven |
| createCommaListExpression at src/compiler/factory/nodeFactory.ts:6294:5 | Mutable<CommaListExpression> | composed-or-forwarding | yes | unproven |
| updateCommaListExpression at src/compiler/factory/nodeFactory.ts:6302:5 | CommaListExpression | composed-or-forwarding | unproven | unproven |
| createSyntheticReferenceExpression at src/compiler/factory/nodeFactory.ts:6309:5 | Mutable<SyntheticReferenceExpression> | composed-or-forwarding | yes | unproven |
| updateSyntheticReferenceExpression at src/compiler/factory/nodeFactory.ts:6319:5 | SyntheticReferenceExpression | composed-or-forwarding | unproven | unproven |
| cloneGeneratedIdentifier at src/compiler/factory/nodeFactory.ts:6326:5 | GeneratedIdentifier | composed-or-forwarding | unproven | unproven |
| cloneIdentifier at src/compiler/factory/nodeFactory.ts:6335:5 | Identifier | composed-or-forwarding | unproven | unproven |
| cloneGeneratedPrivateIdentifier at src/compiler/factory/nodeFactory.ts:6350:5 | GeneratedPrivateIdentifier | composed-or-forwarding | yes | unproven |
| clonePrivateIdentifier at src/compiler/factory/nodeFactory.ts:6359:5 | PrivateIdentifier | composed-or-forwarding | yes | unproven |
| cloneNode at src/compiler/factory/nodeFactory.ts:6369:5 | T | allocating | unproven | unproven |
| createImmediatelyInvokedFunctionExpression at src/compiler/factory/nodeFactory.ts:6413:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createImmediatelyInvokedArrowFunction at src/compiler/factory/nodeFactory.ts:6431:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createVoidZero at src/compiler/factory/nodeFactory.ts:6446:5 | Mutable<VoidExpression> | composed-or-forwarding | yes | unproven |
| createExportDefault at src/compiler/factory/nodeFactory.ts:6450:5 | Mutable<ExportAssignment> | composed-or-forwarding | yes | unproven |
| createExternalModuleExport at src/compiler/factory/nodeFactory.ts:6458:5 | Mutable<ExportDeclaration> | composed-or-forwarding | yes | unproven |
| createTypeCheck at src/compiler/factory/nodeFactory.ts:6472:5 | BinaryExpression | composed-or-forwarding | unproven | unproven |
| createIsNotTypeCheck at src/compiler/factory/nodeFactory.ts:6478:5 | BinaryExpression | composed-or-forwarding | unproven | unproven |
| createMethodCall at src/compiler/factory/nodeFactory.ts:6484:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createFunctionBindCall at src/compiler/factory/nodeFactory.ts:6501:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createFunctionCallCall at src/compiler/factory/nodeFactory.ts:6505:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createFunctionApplyCall at src/compiler/factory/nodeFactory.ts:6509:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createGlobalMethodCall at src/compiler/factory/nodeFactory.ts:6513:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createArraySliceCall at src/compiler/factory/nodeFactory.ts:6517:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createArrayConcatCall at src/compiler/factory/nodeFactory.ts:6521:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createObjectDefinePropertyCall at src/compiler/factory/nodeFactory.ts:6525:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createObjectGetOwnPropertyDescriptorCall at src/compiler/factory/nodeFactory.ts:6529:5 | Mutable<CallExpression> | composed-or-forwarding | yes | unproven |
| createReflectGetCall at src/compiler/factory/nodeFactory.ts:6533:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createReflectSetCall at src/compiler/factory/nodeFactory.ts:6537:5 | CallExpression | composed-or-forwarding | yes | unproven |
| tryAddPropertyAssignment at src/compiler/factory/nodeFactory.ts:6541:5 | boolean | non-node-builder | unproven | unproven |
| createPropertyDescriptor at src/compiler/factory/nodeFactory.ts:6549:5 | Mutable<ObjectLiteralExpression> | composed-or-forwarding | yes | unproven |
| updateOuterExpression at src/compiler/factory/nodeFactory.ts:6564:5 | ParenthesizedExpression &#124; TypeAssertion &#124; ExpressionWithTypeArguments &#124; AsExpression &#124; NonNullExpression &#124; SatisfiesExpression &#124; PartiallyEmittedExpression | composed-or-forwarding | unproven | unproven |
| isIgnorableParen at src/compiler/factory/nodeFactory.ts:6597:5 | boolean | non-node-builder | unproven | unproven |
| restoreOuterExpressions at src/compiler/factory/nodeFactory.ts:6606:5 | Expression | composed-or-forwarding | unproven | unproven |
| restoreEnclosingLabel at src/compiler/factory/nodeFactory.ts:6616:5 | Statement | composed-or-forwarding | unproven | unproven |
| shouldBeCapturedInTempVariable at src/compiler/factory/nodeFactory.ts:6633:5 | boolean | non-node-builder | unproven | unproven |
| createCallBinding at src/compiler/factory/nodeFactory.ts:6656:5 | CallBinding | non-node-builder | unproven | unproven |
| createAssignmentTargetWrapper at src/compiler/factory/nodeFactory.ts:6725:5 | PropertyAccessExpression | composed-or-forwarding | yes | unproven |
| inlineExpressions at src/compiler/factory/nodeFactory.ts:6751:5 | Expression | composed-or-forwarding | unproven | unproven |
| getName at src/compiler/factory/nodeFactory.ts:6759:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getInternalName at src/compiler/factory/nodeFactory.ts:6784:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getLocalName at src/compiler/factory/nodeFactory.ts:6799:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getExportName at src/compiler/factory/nodeFactory.ts:6813:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getDeclarationName at src/compiler/factory/nodeFactory.ts:6824:5 | Identifier | composed-or-forwarding | unproven | unproven |
| getNamespaceMemberName at src/compiler/factory/nodeFactory.ts:6836:5 | PropertyAccessExpression | composed-or-forwarding | yes | unproven |
| getExternalModuleOrNamespaceExportName at src/compiler/factory/nodeFactory.ts:6857:5 | Identifier &#124; PropertyAccessExpression | composed-or-forwarding | unproven | unproven |
| isUseStrictPrologue at src/compiler/factory/nodeFactory.ts:6876:5 | boolean | non-node-builder | unproven | unproven |
| createUseStrictPrologue at src/compiler/factory/nodeFactory.ts:6880:5 | PrologueDirective | composed-or-forwarding | unproven | unproven |
| liftToBlock at src/compiler/factory/nodeFactory.ts:6959:5 | Statement | composed-or-forwarding | unproven | unproven |
| replaceModifiers at src/compiler/factory/nodeFactory.ts:7067:5 | ConstructorTypeNode &#124; TypeParameterDeclaration &#124; FunctionExpression &#124; ParameterDeclaration &#124; ... 20 more ... &#124; ExportDeclaration | composed-or-forwarding | unproven | unproven |
| replaceDecoratorsAndModifiers at src/compiler/factory/nodeFactory.ts:7104:5 | ParameterDeclaration &#124; PropertyDeclaration &#124; MethodDeclaration &#124; GetAccessorDeclaration &#124; SetAccessorDeclaration &#124; ClassExpression &#124; ClassDeclaration | composed-or-forwarding | unproven | unproven |
| replacePropertyName at src/compiler/factory/nodeFactory.ts:7116:5 | PropertySignature &#124; PropertyDeclaration &#124; MethodSignature &#124; MethodDeclaration &#124; GetAccessorDeclaration &#124; SetAccessorDeclaration &#124; PropertyAssignment | composed-or-forwarding | unproven | unproven |
| asName at src/compiler/factory/nodeFactory.ts:7141:5 | Identifier &#124; T | composed-or-forwarding | unproven | unproven |
| asExpression at src/compiler/factory/nodeFactory.ts:7146:5 | StringLiteral &#124; NumericLiteral &#124; BooleanLiteral &#124; T | composed-or-forwarding | unproven | unproven |
| asInitializer at src/compiler/factory/nodeFactory.ts:7153:5 | Expression &#124; undefined | composed-or-forwarding | unproven | unproven |
| asToken at src/compiler/factory/nodeFactory.ts:7157:5 | Token<TKind> | composed-or-forwarding | unproven | unproven |
| asEmbeddedStatement at src/compiler/factory/nodeFactory.ts:7163:5 | EmptyStatement &#124; T &#124; undefined | composed-or-forwarding | unproven | unproven |
| asVariableDeclaration at src/compiler/factory/nodeFactory.ts:7167:5 | VariableDeclaration &#124; undefined | composed-or-forwarding | yes | unproven |
| update at src/compiler/factory/nodeFactory.ts:7179:5 | T | composed-or-forwarding | unproven | unproven |
| propagateNameFlags at src/compiler/factory/nodeFactory.ts:7286:1 | number | non-node-builder | unproven | unproven |
| propagateIdentifierNameFlags at src/compiler/factory/nodeFactory.ts:7290:1 | number | non-node-builder | unproven | unproven |
| propagatePropertyNameFlagsOfChild at src/compiler/factory/nodeFactory.ts:7295:1 | number | non-node-builder | unproven | unproven |
| propagateChildFlags at src/compiler/factory/nodeFactory.ts:7299:1 | TransformFlags | non-node-builder | unproven | unproven |
| makeSynthetic at src/compiler/factory/nodeFactory.ts:7394:1 | Node | input-or-unresolved | unproven | unproven |
| createBaseSourceFileNode at src/compiler/factory/nodeFactory.ts:7400:31 | Node | allocating | unproven | unproven |
| createBaseIdentifierNode at src/compiler/factory/nodeFactory.ts:7401:31 | Node | allocating | unproven | unproven |
| createBasePrivateIdentifierNode at src/compiler/factory/nodeFactory.ts:7402:38 | Node | allocating | unproven | unproven |
| createBaseTokenNode at src/compiler/factory/nodeFactory.ts:7403:26 | Node | allocating | unproven | unproven |
| createBaseNode at src/compiler/factory/nodeFactory.ts:7404:21 | Node | allocating | unproven | unproven |
| createSourceMapSource at src/compiler/factory/nodeFactory.ts:7414:1 | SourceMapSource | non-node-builder | unproven | unproven |
| setOriginalNode at src/compiler/factory/nodeFactory.ts:7420:1 | T | input-or-unresolved | unproven | unproven |
| getOrCreateEmitNode at src/compiler/factory/emitNode.ts:36:1 | EmitNode | non-node-builder | unproven | unproven |
| disposeEmitNodes at src/compiler/factory/emitNode.ts:62:1 | void | non-node-builder | unproven | unproven |
| removeAllComments at src/compiler/factory/emitNode.ts:81:1 | T | input-or-unresolved | unproven | unproven |
| setEmitFlags at src/compiler/factory/emitNode.ts:92:1 | T | input-or-unresolved | unproven | unproven |
| addEmitFlags at src/compiler/factory/emitNode.ts:102:1 | T | input-or-unresolved | unproven | unproven |
| setInternalEmitFlags at src/compiler/factory/emitNode.ts:113:1 | T | input-or-unresolved | unproven | unproven |
| addInternalEmitFlags at src/compiler/factory/emitNode.ts:123:1 | T | input-or-unresolved | unproven | unproven |
| getSourceMapRange at src/compiler/factory/emitNode.ts:132:1 | SourceMapRange | non-node-builder | unproven | unproven |
| setSourceMapRange at src/compiler/factory/emitNode.ts:139:1 | T | input-or-unresolved | unproven | unproven |
| getTokenSourceMapRange at src/compiler/factory/emitNode.ts:147:1 | SourceMapRange &#124; undefined | non-node-builder | unproven | unproven |
| setTokenSourceMapRange at src/compiler/factory/emitNode.ts:154:1 | T | input-or-unresolved | unproven | unproven |
| getStartsOnNewLine at src/compiler/factory/emitNode.ts:166:1 | boolean &#124; undefined | non-node-builder | unproven | unproven |
| setStartsOnNewLine at src/compiler/factory/emitNode.ts:175:1 | T | input-or-unresolved | unproven | unproven |
| getCommentRange at src/compiler/factory/emitNode.ts:183:1 | TextRange | non-node-builder | unproven | unproven |
| setCommentRange at src/compiler/factory/emitNode.ts:190:1 | T | input-or-unresolved | unproven | unproven |
| getSyntheticLeadingComments at src/compiler/factory/emitNode.ts:195:1 | SynthesizedComment[] &#124; undefined | non-node-builder | unproven | unproven |
| setSyntheticLeadingComments at src/compiler/factory/emitNode.ts:199:1 | T | input-or-unresolved | unproven | unproven |
| addSyntheticLeadingComment at src/compiler/factory/emitNode.ts:204:1 | T | allocating | unproven | unproven |
| getSyntheticTrailingComments at src/compiler/factory/emitNode.ts:208:1 | SynthesizedComment[] &#124; undefined | non-node-builder | unproven | unproven |
| setSyntheticTrailingComments at src/compiler/factory/emitNode.ts:212:1 | T | input-or-unresolved | unproven | unproven |
| addSyntheticTrailingComment at src/compiler/factory/emitNode.ts:217:1 | T | allocating | unproven | unproven |
| moveSyntheticComments at src/compiler/factory/emitNode.ts:221:1 | T | composed-or-forwarding | unproven | unproven |
| getConstantValue at src/compiler/factory/emitNode.ts:233:1 | string &#124; number &#124; undefined | non-node-builder | unproven | unproven |
| setConstantValue at src/compiler/factory/emitNode.ts:240:1 | AccessExpression | input-or-unresolved | unproven | unproven |
| addEmitHelper at src/compiler/factory/emitNode.ts:249:1 | T | input-or-unresolved | unproven | unproven |
| addEmitHelpers at src/compiler/factory/emitNode.ts:258:1 | T | input-or-unresolved | unproven | unproven |
| removeEmitHelper at src/compiler/factory/emitNode.ts:271:1 | boolean | non-node-builder | unproven | unproven |
| getEmitHelpers at src/compiler/factory/emitNode.ts:282:1 | EmitHelper[] &#124; undefined | non-node-builder | unproven | unproven |
| moveEmitHelpers at src/compiler/factory/emitNode.ts:289:1 | void | non-node-builder | unproven | unproven |
| getSnippetElement at src/compiler/factory/emitNode.ts:317:1 | SnippetElement &#124; undefined | non-node-builder | unproven | unproven |
| setSnippetElement at src/compiler/factory/emitNode.ts:326:1 | T | input-or-unresolved | unproven | unproven |
| ignoreSourceNewlines at src/compiler/factory/emitNode.ts:333:1 | T | input-or-unresolved | unproven | unproven |
| setTypeNode at src/compiler/factory/emitNode.ts:339:1 | T | input-or-unresolved | unproven | unproven |
| getTypeNode at src/compiler/factory/emitNode.ts:346:1 | TypeNode &#124; undefined | input-or-unresolved | unproven | unproven |
| setIdentifierTypeArguments at src/compiler/factory/emitNode.ts:351:1 | T | input-or-unresolved | unproven | unproven |
| getIdentifierTypeArguments at src/compiler/factory/emitNode.ts:357:1 | NodeArray<TypeNode &#124; TypeParameterDeclaration> &#124; undefined | non-node-builder | unproven | unproven |
| setIdentifierAutoGenerate at src/compiler/factory/emitNode.ts:362:1 | T | input-or-unresolved | unproven | unproven |
| getIdentifierAutoGenerate at src/compiler/factory/emitNode.ts:368:1 | AutoGenerateInfo &#124; undefined | non-node-builder | unproven | unproven |
| setIdentifierGeneratedImportReference at src/compiler/factory/emitNode.ts:373:1 | T | input-or-unresolved | unproven | unproven |
| getIdentifierGeneratedImportReference at src/compiler/factory/emitNode.ts:379:1 | ImportSpecifier &#124; undefined | input-or-unresolved | unproven | unproven |
| createEmitHelperFactory at src/compiler/factory/emitHelpers.ts:148:1 | EmitHelperFactory | non-node-builder | unproven | unproven |
| <anonymous@150:35> at src/compiler/factory/emitHelpers.ts:150:35 | TrueLiteral | composed-or-forwarding | unproven | unproven |
| <anonymous@151:36> at src/compiler/factory/emitHelpers.ts:151:36 | FalseLiteral | composed-or-forwarding | unproven | unproven |
| getUnscopedHelperName at src/compiler/factory/emitHelpers.ts:202:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createDecorateHelper at src/compiler/factory/emitHelpers.ts:208:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createMetadataHelper at src/compiler/factory/emitHelpers.ts:228:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createParamHelper at src/compiler/factory/emitHelpers.ts:240:5 | CallExpression | composed-or-forwarding | unproven | unproven |
| createESDecorateClassContextObject at src/compiler/factory/emitHelpers.ts:257:5 | ObjectLiteralExpression | composed-or-forwarding | yes | unproven |
| createESDecorateClassElementAccessGetMethod at src/compiler/factory/emitHelpers.ts:267:5 | PropertyAssignment | composed-or-forwarding | yes | unproven |
| createESDecorateClassElementAccessSetMethod at src/compiler/factory/emitHelpers.ts:289:5 | PropertyAssignment | composed-or-forwarding | yes | unproven |
| createESDecorateClassElementAccessHasMethod at src/compiler/factory/emitHelpers.ts:325:5 | PropertyAssignment | composed-or-forwarding | yes | unproven |
| createESDecorateClassElementAccessObject at src/compiler/factory/emitHelpers.ts:351:5 | ObjectLiteralExpression | composed-or-forwarding | yes | unproven |
| createESDecorateClassElementContextObject at src/compiler/factory/emitHelpers.ts:359:5 | ObjectLiteralExpression | composed-or-forwarding | yes | unproven |
| createESDecorateContextObject at src/compiler/factory/emitHelpers.ts:371:5 | ObjectLiteralExpression | composed-or-forwarding | yes | unproven |
| createESDecorateHelper at src/compiler/factory/emitHelpers.ts:376:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createRunInitializersHelper at src/compiler/factory/emitHelpers.ts:392:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAssignHelper at src/compiler/factory/emitHelpers.ts:402:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAwaitHelper at src/compiler/factory/emitHelpers.ts:414:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAsyncGeneratorHelper at src/compiler/factory/emitHelpers.ts:419:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAsyncDelegatorHelper at src/compiler/factory/emitHelpers.ts:437:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAsyncValuesHelper at src/compiler/factory/emitHelpers.ts:447:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createRestHelper at src/compiler/factory/emitHelpers.ts:461:5 | Expression | composed-or-forwarding | unproven | unproven |
| createAwaiterHelper at src/compiler/factory/emitHelpers.ts:503:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createExtendsHelper at src/compiler/factory/emitHelpers.ts:533:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createTemplateObjectHelper at src/compiler/factory/emitHelpers.ts:542:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createSpreadArrayHelper at src/compiler/factory/emitHelpers.ts:551:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createPropKeyHelper at src/compiler/factory/emitHelpers.ts:560:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createSetFunctionNameHelper at src/compiler/factory/emitHelpers.ts:569:5 | Expression | composed-or-forwarding | yes | unproven |
| createValuesHelper at src/compiler/factory/emitHelpers.ts:580:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createReadHelper at src/compiler/factory/emitHelpers.ts:589:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createGeneratorHelper at src/compiler/factory/emitHelpers.ts:602:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createImportStarHelper at src/compiler/factory/emitHelpers.ts:613:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createImportStarCallbackHelper at src/compiler/factory/emitHelpers.ts:622:5 | Identifier | composed-or-forwarding | unproven | unproven |
| createImportDefaultHelper at src/compiler/factory/emitHelpers.ts:627:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createExportStarHelper at src/compiler/factory/emitHelpers.ts:636:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createClassPrivateFieldGetHelper at src/compiler/factory/emitHelpers.ts:648:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createClassPrivateFieldSetHelper at src/compiler/factory/emitHelpers.ts:660:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createClassPrivateFieldInHelper at src/compiler/factory/emitHelpers.ts:672:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createAddDisposableResourceHelper at src/compiler/factory/emitHelpers.ts:677:5 | Expression | composed-or-forwarding | yes | unproven |
| createDisposeResourcesHelper at src/compiler/factory/emitHelpers.ts:686:5 | CallExpression | composed-or-forwarding | yes | unproven |
| createRewriteRelativeImportExtensionsHelper at src/compiler/factory/emitHelpers.ts:691:5 | CallExpression | composed-or-forwarding | yes | unproven |
| isCallToHelper at src/compiler/factory/emitHelpers.ts:1486:1 | boolean | non-node-builder | unproven | unproven |
| isNumericLiteral at src/compiler/factory/nodeTests.ts:235:1 | boolean | non-node-builder | unproven | unproven |
| isBigIntLiteral at src/compiler/factory/nodeTests.ts:239:1 | boolean | non-node-builder | unproven | unproven |
| isStringLiteral at src/compiler/factory/nodeTests.ts:243:1 | boolean | non-node-builder | unproven | unproven |
| isJsxText at src/compiler/factory/nodeTests.ts:247:1 | boolean | non-node-builder | unproven | unproven |
| isRegularExpressionLiteral at src/compiler/factory/nodeTests.ts:251:1 | boolean | non-node-builder | unproven | unproven |
| isNoSubstitutionTemplateLiteral at src/compiler/factory/nodeTests.ts:255:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateHead at src/compiler/factory/nodeTests.ts:261:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateMiddle at src/compiler/factory/nodeTests.ts:265:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateTail at src/compiler/factory/nodeTests.ts:269:1 | boolean | non-node-builder | unproven | unproven |
| isDotDotDotToken at src/compiler/factory/nodeTests.ts:275:1 | boolean | non-node-builder | unproven | unproven |
| isCommaToken at src/compiler/factory/nodeTests.ts:280:1 | boolean | non-node-builder | unproven | unproven |
| isPlusToken at src/compiler/factory/nodeTests.ts:284:1 | boolean | non-node-builder | unproven | unproven |
| isMinusToken at src/compiler/factory/nodeTests.ts:288:1 | boolean | non-node-builder | unproven | unproven |
| isAsteriskToken at src/compiler/factory/nodeTests.ts:292:1 | boolean | non-node-builder | unproven | unproven |
| isExclamationToken at src/compiler/factory/nodeTests.ts:296:1 | boolean | non-node-builder | unproven | unproven |
| isQuestionToken at src/compiler/factory/nodeTests.ts:300:1 | boolean | non-node-builder | unproven | unproven |
| isColonToken at src/compiler/factory/nodeTests.ts:304:1 | boolean | non-node-builder | unproven | unproven |
| isQuestionDotToken at src/compiler/factory/nodeTests.ts:308:1 | boolean | non-node-builder | unproven | unproven |
| isEqualsGreaterThanToken at src/compiler/factory/nodeTests.ts:312:1 | boolean | non-node-builder | unproven | unproven |
| isIdentifier at src/compiler/factory/nodeTests.ts:318:1 | boolean | non-node-builder | unproven | unproven |
| isPrivateIdentifier at src/compiler/factory/nodeTests.ts:322:1 | boolean | non-node-builder | unproven | unproven |
| isExportModifier at src/compiler/factory/nodeTests.ts:329:1 | boolean | non-node-builder | unproven | unproven |
| isDefaultModifier at src/compiler/factory/nodeTests.ts:334:1 | boolean | non-node-builder | unproven | unproven |
| isAsyncModifier at src/compiler/factory/nodeTests.ts:339:1 | boolean | non-node-builder | unproven | unproven |
| isAssertsKeyword at src/compiler/factory/nodeTests.ts:343:1 | boolean | non-node-builder | unproven | unproven |
| isAwaitKeyword at src/compiler/factory/nodeTests.ts:347:1 | boolean | non-node-builder | unproven | unproven |
| isReadonlyKeyword at src/compiler/factory/nodeTests.ts:352:1 | boolean | non-node-builder | unproven | unproven |
| isStaticModifier at src/compiler/factory/nodeTests.ts:357:1 | boolean | non-node-builder | unproven | unproven |
| isAbstractModifier at src/compiler/factory/nodeTests.ts:362:1 | boolean | non-node-builder | unproven | unproven |
| isOverrideModifier at src/compiler/factory/nodeTests.ts:367:1 | boolean | non-node-builder | unproven | unproven |
| isAccessorModifier at src/compiler/factory/nodeTests.ts:372:1 | boolean | non-node-builder | unproven | unproven |
| isSuperKeyword at src/compiler/factory/nodeTests.ts:377:1 | boolean | non-node-builder | unproven | unproven |
| isImportKeyword at src/compiler/factory/nodeTests.ts:382:1 | boolean | non-node-builder | unproven | unproven |
| isCaseKeyword at src/compiler/factory/nodeTests.ts:387:1 | boolean | non-node-builder | unproven | unproven |
| isQualifiedName at src/compiler/factory/nodeTests.ts:393:1 | boolean | non-node-builder | unproven | unproven |
| isComputedPropertyName at src/compiler/factory/nodeTests.ts:397:1 | boolean | non-node-builder | unproven | unproven |
| isTypeParameterDeclaration at src/compiler/factory/nodeTests.ts:403:1 | boolean | non-node-builder | unproven | unproven |
| isParameter at src/compiler/factory/nodeTests.ts:408:1 | boolean | non-node-builder | unproven | unproven |
| isDecorator at src/compiler/factory/nodeTests.ts:412:1 | boolean | non-node-builder | unproven | unproven |
| isPropertySignature at src/compiler/factory/nodeTests.ts:418:1 | boolean | non-node-builder | unproven | unproven |
| isPropertyDeclaration at src/compiler/factory/nodeTests.ts:422:1 | boolean | non-node-builder | unproven | unproven |
| isMethodSignature at src/compiler/factory/nodeTests.ts:426:1 | boolean | non-node-builder | unproven | unproven |
| isMethodDeclaration at src/compiler/factory/nodeTests.ts:430:1 | boolean | non-node-builder | unproven | unproven |
| isClassStaticBlockDeclaration at src/compiler/factory/nodeTests.ts:434:1 | boolean | non-node-builder | unproven | unproven |
| isConstructorDeclaration at src/compiler/factory/nodeTests.ts:438:1 | boolean | non-node-builder | unproven | unproven |
| isGetAccessorDeclaration at src/compiler/factory/nodeTests.ts:442:1 | boolean | non-node-builder | unproven | unproven |
| isSetAccessorDeclaration at src/compiler/factory/nodeTests.ts:446:1 | boolean | non-node-builder | unproven | unproven |
| isCallSignatureDeclaration at src/compiler/factory/nodeTests.ts:450:1 | boolean | non-node-builder | unproven | unproven |
| isConstructSignatureDeclaration at src/compiler/factory/nodeTests.ts:454:1 | boolean | non-node-builder | unproven | unproven |
| isIndexSignatureDeclaration at src/compiler/factory/nodeTests.ts:458:1 | boolean | non-node-builder | unproven | unproven |
| isTypePredicateNode at src/compiler/factory/nodeTests.ts:464:1 | boolean | non-node-builder | unproven | unproven |
| isTypeReferenceNode at src/compiler/factory/nodeTests.ts:468:1 | boolean | non-node-builder | unproven | unproven |
| isFunctionTypeNode at src/compiler/factory/nodeTests.ts:472:1 | boolean | non-node-builder | unproven | unproven |
| isConstructorTypeNode at src/compiler/factory/nodeTests.ts:476:1 | boolean | non-node-builder | unproven | unproven |
| isTypeQueryNode at src/compiler/factory/nodeTests.ts:480:1 | boolean | non-node-builder | unproven | unproven |
| isTypeLiteralNode at src/compiler/factory/nodeTests.ts:484:1 | boolean | non-node-builder | unproven | unproven |
| isArrayTypeNode at src/compiler/factory/nodeTests.ts:488:1 | boolean | non-node-builder | unproven | unproven |
| isTupleTypeNode at src/compiler/factory/nodeTests.ts:492:1 | boolean | non-node-builder | unproven | unproven |
| isNamedTupleMember at src/compiler/factory/nodeTests.ts:496:1 | boolean | non-node-builder | unproven | unproven |
| isOptionalTypeNode at src/compiler/factory/nodeTests.ts:500:1 | boolean | non-node-builder | unproven | unproven |
| isRestTypeNode at src/compiler/factory/nodeTests.ts:504:1 | boolean | non-node-builder | unproven | unproven |
| isUnionTypeNode at src/compiler/factory/nodeTests.ts:508:1 | boolean | non-node-builder | unproven | unproven |
| isIntersectionTypeNode at src/compiler/factory/nodeTests.ts:512:1 | boolean | non-node-builder | unproven | unproven |
| isConditionalTypeNode at src/compiler/factory/nodeTests.ts:516:1 | boolean | non-node-builder | unproven | unproven |
| isInferTypeNode at src/compiler/factory/nodeTests.ts:520:1 | boolean | non-node-builder | unproven | unproven |
| isParenthesizedTypeNode at src/compiler/factory/nodeTests.ts:524:1 | boolean | non-node-builder | unproven | unproven |
| isThisTypeNode at src/compiler/factory/nodeTests.ts:528:1 | boolean | non-node-builder | unproven | unproven |
| isTypeOperatorNode at src/compiler/factory/nodeTests.ts:532:1 | boolean | non-node-builder | unproven | unproven |
| isIndexedAccessTypeNode at src/compiler/factory/nodeTests.ts:536:1 | boolean | non-node-builder | unproven | unproven |
| isMappedTypeNode at src/compiler/factory/nodeTests.ts:540:1 | boolean | non-node-builder | unproven | unproven |
| isLiteralTypeNode at src/compiler/factory/nodeTests.ts:544:1 | boolean | non-node-builder | unproven | unproven |
| isImportTypeNode at src/compiler/factory/nodeTests.ts:548:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateLiteralTypeSpan at src/compiler/factory/nodeTests.ts:552:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateLiteralTypeNode at src/compiler/factory/nodeTests.ts:556:1 | boolean | non-node-builder | unproven | unproven |
| isObjectBindingPattern at src/compiler/factory/nodeTests.ts:562:1 | boolean | non-node-builder | unproven | unproven |
| isArrayBindingPattern at src/compiler/factory/nodeTests.ts:566:1 | boolean | non-node-builder | unproven | unproven |
| isBindingElement at src/compiler/factory/nodeTests.ts:570:1 | boolean | non-node-builder | unproven | unproven |
| isArrayLiteralExpression at src/compiler/factory/nodeTests.ts:576:1 | boolean | non-node-builder | unproven | unproven |
| isObjectLiteralExpression at src/compiler/factory/nodeTests.ts:580:1 | boolean | non-node-builder | unproven | unproven |
| isPropertyAccessExpression at src/compiler/factory/nodeTests.ts:584:1 | boolean | non-node-builder | unproven | unproven |
| isElementAccessExpression at src/compiler/factory/nodeTests.ts:588:1 | boolean | non-node-builder | unproven | unproven |
| isCallExpression at src/compiler/factory/nodeTests.ts:592:1 | boolean | non-node-builder | unproven | unproven |
| isNewExpression at src/compiler/factory/nodeTests.ts:596:1 | boolean | non-node-builder | unproven | unproven |
| isTaggedTemplateExpression at src/compiler/factory/nodeTests.ts:600:1 | boolean | non-node-builder | unproven | unproven |
| isTypeAssertionExpression at src/compiler/factory/nodeTests.ts:604:1 | boolean | non-node-builder | unproven | unproven |
| isParenthesizedExpression at src/compiler/factory/nodeTests.ts:608:1 | boolean | non-node-builder | unproven | unproven |
| isFunctionExpression at src/compiler/factory/nodeTests.ts:612:1 | boolean | non-node-builder | unproven | unproven |
| isArrowFunction at src/compiler/factory/nodeTests.ts:616:1 | boolean | non-node-builder | unproven | unproven |
| isDeleteExpression at src/compiler/factory/nodeTests.ts:620:1 | boolean | non-node-builder | unproven | unproven |
| isTypeOfExpression at src/compiler/factory/nodeTests.ts:624:1 | boolean | non-node-builder | unproven | unproven |
| isVoidExpression at src/compiler/factory/nodeTests.ts:628:1 | boolean | non-node-builder | unproven | unproven |
| isAwaitExpression at src/compiler/factory/nodeTests.ts:632:1 | boolean | non-node-builder | unproven | unproven |
| isPrefixUnaryExpression at src/compiler/factory/nodeTests.ts:636:1 | boolean | non-node-builder | unproven | unproven |
| isPostfixUnaryExpression at src/compiler/factory/nodeTests.ts:640:1 | boolean | non-node-builder | unproven | unproven |
| isBinaryExpression at src/compiler/factory/nodeTests.ts:644:1 | boolean | non-node-builder | unproven | unproven |
| isConditionalExpression at src/compiler/factory/nodeTests.ts:648:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateExpression at src/compiler/factory/nodeTests.ts:652:1 | boolean | non-node-builder | unproven | unproven |
| isYieldExpression at src/compiler/factory/nodeTests.ts:656:1 | boolean | non-node-builder | unproven | unproven |
| isSpreadElement at src/compiler/factory/nodeTests.ts:660:1 | boolean | non-node-builder | unproven | unproven |
| isClassExpression at src/compiler/factory/nodeTests.ts:664:1 | boolean | non-node-builder | unproven | unproven |
| isOmittedExpression at src/compiler/factory/nodeTests.ts:668:1 | boolean | non-node-builder | unproven | unproven |
| isExpressionWithTypeArguments at src/compiler/factory/nodeTests.ts:672:1 | boolean | non-node-builder | unproven | unproven |
| isAsExpression at src/compiler/factory/nodeTests.ts:676:1 | boolean | non-node-builder | unproven | unproven |
| isSatisfiesExpression at src/compiler/factory/nodeTests.ts:680:1 | boolean | non-node-builder | unproven | unproven |
| isNonNullExpression at src/compiler/factory/nodeTests.ts:684:1 | boolean | non-node-builder | unproven | unproven |
| isMetaProperty at src/compiler/factory/nodeTests.ts:688:1 | boolean | non-node-builder | unproven | unproven |
| isSyntheticExpression at src/compiler/factory/nodeTests.ts:692:1 | boolean | non-node-builder | unproven | unproven |
| isPartiallyEmittedExpression at src/compiler/factory/nodeTests.ts:696:1 | boolean | non-node-builder | unproven | unproven |
| isCommaListExpression at src/compiler/factory/nodeTests.ts:700:1 | boolean | non-node-builder | unproven | unproven |
| isTemplateSpan at src/compiler/factory/nodeTests.ts:706:1 | boolean | non-node-builder | unproven | unproven |
| isSemicolonClassElement at src/compiler/factory/nodeTests.ts:710:1 | boolean | non-node-builder | unproven | unproven |
| isBlock at src/compiler/factory/nodeTests.ts:716:1 | boolean | non-node-builder | unproven | unproven |
| isVariableStatement at src/compiler/factory/nodeTests.ts:720:1 | boolean | non-node-builder | unproven | unproven |
| isEmptyStatement at src/compiler/factory/nodeTests.ts:724:1 | boolean | non-node-builder | unproven | unproven |
| isExpressionStatement at src/compiler/factory/nodeTests.ts:728:1 | boolean | non-node-builder | unproven | unproven |
| isIfStatement at src/compiler/factory/nodeTests.ts:732:1 | boolean | non-node-builder | unproven | unproven |
| isDoStatement at src/compiler/factory/nodeTests.ts:736:1 | boolean | non-node-builder | unproven | unproven |
| isWhileStatement at src/compiler/factory/nodeTests.ts:740:1 | boolean | non-node-builder | unproven | unproven |
| isForStatement at src/compiler/factory/nodeTests.ts:744:1 | boolean | non-node-builder | unproven | unproven |
| isForInStatement at src/compiler/factory/nodeTests.ts:748:1 | boolean | non-node-builder | unproven | unproven |
| isForOfStatement at src/compiler/factory/nodeTests.ts:752:1 | boolean | non-node-builder | unproven | unproven |
| isContinueStatement at src/compiler/factory/nodeTests.ts:756:1 | boolean | non-node-builder | unproven | unproven |
| isBreakStatement at src/compiler/factory/nodeTests.ts:760:1 | boolean | non-node-builder | unproven | unproven |
| isReturnStatement at src/compiler/factory/nodeTests.ts:764:1 | boolean | non-node-builder | unproven | unproven |
| isWithStatement at src/compiler/factory/nodeTests.ts:768:1 | boolean | non-node-builder | unproven | unproven |
| isSwitchStatement at src/compiler/factory/nodeTests.ts:772:1 | boolean | non-node-builder | unproven | unproven |
| isLabeledStatement at src/compiler/factory/nodeTests.ts:776:1 | boolean | non-node-builder | unproven | unproven |
| isThrowStatement at src/compiler/factory/nodeTests.ts:780:1 | boolean | non-node-builder | unproven | unproven |
| isTryStatement at src/compiler/factory/nodeTests.ts:784:1 | boolean | non-node-builder | unproven | unproven |
| isDebuggerStatement at src/compiler/factory/nodeTests.ts:788:1 | boolean | non-node-builder | unproven | unproven |
| isVariableDeclaration at src/compiler/factory/nodeTests.ts:792:1 | boolean | non-node-builder | unproven | unproven |
| isVariableDeclarationList at src/compiler/factory/nodeTests.ts:796:1 | boolean | non-node-builder | unproven | unproven |
| isFunctionDeclaration at src/compiler/factory/nodeTests.ts:800:1 | boolean | non-node-builder | unproven | unproven |
| isClassDeclaration at src/compiler/factory/nodeTests.ts:804:1 | boolean | non-node-builder | unproven | unproven |
| isInterfaceDeclaration at src/compiler/factory/nodeTests.ts:808:1 | boolean | non-node-builder | unproven | unproven |
| isTypeAliasDeclaration at src/compiler/factory/nodeTests.ts:812:1 | boolean | non-node-builder | unproven | unproven |
| isEnumDeclaration at src/compiler/factory/nodeTests.ts:816:1 | boolean | non-node-builder | unproven | unproven |
| isModuleDeclaration at src/compiler/factory/nodeTests.ts:820:1 | boolean | non-node-builder | unproven | unproven |
| isModuleBlock at src/compiler/factory/nodeTests.ts:824:1 | boolean | non-node-builder | unproven | unproven |
| isCaseBlock at src/compiler/factory/nodeTests.ts:828:1 | boolean | non-node-builder | unproven | unproven |
| isNamespaceExportDeclaration at src/compiler/factory/nodeTests.ts:832:1 | boolean | non-node-builder | unproven | unproven |
| isImportEqualsDeclaration at src/compiler/factory/nodeTests.ts:836:1 | boolean | non-node-builder | unproven | unproven |
| isImportDeclaration at src/compiler/factory/nodeTests.ts:840:1 | boolean | non-node-builder | unproven | unproven |
| isImportClause at src/compiler/factory/nodeTests.ts:844:1 | boolean | non-node-builder | unproven | unproven |
| isImportTypeAssertionContainer at src/compiler/factory/nodeTests.ts:848:1 | boolean | non-node-builder | unproven | unproven |
| isAssertClause at src/compiler/factory/nodeTests.ts:853:1 | boolean | non-node-builder | unproven | unproven |
| isAssertEntry at src/compiler/factory/nodeTests.ts:858:1 | boolean | non-node-builder | unproven | unproven |
| isImportAttributes at src/compiler/factory/nodeTests.ts:862:1 | boolean | non-node-builder | unproven | unproven |
| isImportAttribute at src/compiler/factory/nodeTests.ts:866:1 | boolean | non-node-builder | unproven | unproven |
| isNamespaceImport at src/compiler/factory/nodeTests.ts:870:1 | boolean | non-node-builder | unproven | unproven |
| isNamespaceExport at src/compiler/factory/nodeTests.ts:874:1 | boolean | non-node-builder | unproven | unproven |
| isNamedImports at src/compiler/factory/nodeTests.ts:878:1 | boolean | non-node-builder | unproven | unproven |
| isImportSpecifier at src/compiler/factory/nodeTests.ts:882:1 | boolean | non-node-builder | unproven | unproven |
| isExportAssignment at src/compiler/factory/nodeTests.ts:886:1 | boolean | non-node-builder | unproven | unproven |
| isExportDeclaration at src/compiler/factory/nodeTests.ts:890:1 | boolean | non-node-builder | unproven | unproven |
| isNamedExports at src/compiler/factory/nodeTests.ts:894:1 | boolean | non-node-builder | unproven | unproven |
| isExportSpecifier at src/compiler/factory/nodeTests.ts:898:1 | boolean | non-node-builder | unproven | unproven |
| isModuleExportName at src/compiler/factory/nodeTests.ts:902:1 | boolean | non-node-builder | unproven | unproven |
| isMissingDeclaration at src/compiler/factory/nodeTests.ts:906:1 | boolean | non-node-builder | unproven | unproven |
| isNotEmittedStatement at src/compiler/factory/nodeTests.ts:910:1 | boolean | non-node-builder | unproven | unproven |
| isSyntheticReference at src/compiler/factory/nodeTests.ts:915:1 | boolean | non-node-builder | unproven | unproven |
| isExternalModuleReference at src/compiler/factory/nodeTests.ts:921:1 | boolean | non-node-builder | unproven | unproven |
| isJsxElement at src/compiler/factory/nodeTests.ts:927:1 | boolean | non-node-builder | unproven | unproven |
| isJsxSelfClosingElement at src/compiler/factory/nodeTests.ts:931:1 | boolean | non-node-builder | unproven | unproven |
| isJsxOpeningElement at src/compiler/factory/nodeTests.ts:935:1 | boolean | non-node-builder | unproven | unproven |
| isJsxClosingElement at src/compiler/factory/nodeTests.ts:939:1 | boolean | non-node-builder | unproven | unproven |
| isJsxFragment at src/compiler/factory/nodeTests.ts:943:1 | boolean | non-node-builder | unproven | unproven |
| isJsxOpeningFragment at src/compiler/factory/nodeTests.ts:947:1 | boolean | non-node-builder | unproven | unproven |
| isJsxClosingFragment at src/compiler/factory/nodeTests.ts:951:1 | boolean | non-node-builder | unproven | unproven |
| isJsxAttribute at src/compiler/factory/nodeTests.ts:955:1 | boolean | non-node-builder | unproven | unproven |
| isJsxAttributes at src/compiler/factory/nodeTests.ts:959:1 | boolean | non-node-builder | unproven | unproven |
| isJsxSpreadAttribute at src/compiler/factory/nodeTests.ts:963:1 | boolean | non-node-builder | unproven | unproven |
| isJsxExpression at src/compiler/factory/nodeTests.ts:967:1 | boolean | non-node-builder | unproven | unproven |
| isJsxNamespacedName at src/compiler/factory/nodeTests.ts:971:1 | boolean | non-node-builder | unproven | unproven |
| isCaseClause at src/compiler/factory/nodeTests.ts:977:1 | boolean | non-node-builder | unproven | unproven |
| isDefaultClause at src/compiler/factory/nodeTests.ts:981:1 | boolean | non-node-builder | unproven | unproven |
| isHeritageClause at src/compiler/factory/nodeTests.ts:985:1 | boolean | non-node-builder | unproven | unproven |
| isCatchClause at src/compiler/factory/nodeTests.ts:989:1 | boolean | non-node-builder | unproven | unproven |
| isPropertyAssignment at src/compiler/factory/nodeTests.ts:995:1 | boolean | non-node-builder | unproven | unproven |
| isShorthandPropertyAssignment at src/compiler/factory/nodeTests.ts:999:1 | boolean | non-node-builder | unproven | unproven |
| isSpreadAssignment at src/compiler/factory/nodeTests.ts:1003:1 | boolean | non-node-builder | unproven | unproven |
| isEnumMember at src/compiler/factory/nodeTests.ts:1009:1 | boolean | non-node-builder | unproven | unproven |
| isSourceFile at src/compiler/factory/nodeTests.ts:1014:1 | boolean | non-node-builder | unproven | unproven |
| isBundle at src/compiler/factory/nodeTests.ts:1018:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTypeExpression at src/compiler/factory/nodeTests.ts:1026:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocNameReference at src/compiler/factory/nodeTests.ts:1030:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocMemberName at src/compiler/factory/nodeTests.ts:1034:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocLink at src/compiler/factory/nodeTests.ts:1038:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocLinkCode at src/compiler/factory/nodeTests.ts:1042:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocLinkPlain at src/compiler/factory/nodeTests.ts:1046:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocAllType at src/compiler/factory/nodeTests.ts:1050:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocUnknownType at src/compiler/factory/nodeTests.ts:1054:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocNullableType at src/compiler/factory/nodeTests.ts:1058:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocNonNullableType at src/compiler/factory/nodeTests.ts:1062:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocOptionalType at src/compiler/factory/nodeTests.ts:1066:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocFunctionType at src/compiler/factory/nodeTests.ts:1070:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocVariadicType at src/compiler/factory/nodeTests.ts:1074:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocNamepathType at src/compiler/factory/nodeTests.ts:1078:1 | boolean | non-node-builder | unproven | unproven |
| isJSDoc at src/compiler/factory/nodeTests.ts:1082:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTypeLiteral at src/compiler/factory/nodeTests.ts:1086:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocSignature at src/compiler/factory/nodeTests.ts:1090:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocAugmentsTag at src/compiler/factory/nodeTests.ts:1096:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocAuthorTag at src/compiler/factory/nodeTests.ts:1100:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocClassTag at src/compiler/factory/nodeTests.ts:1104:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocCallbackTag at src/compiler/factory/nodeTests.ts:1108:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocPublicTag at src/compiler/factory/nodeTests.ts:1112:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocPrivateTag at src/compiler/factory/nodeTests.ts:1116:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocProtectedTag at src/compiler/factory/nodeTests.ts:1120:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocReadonlyTag at src/compiler/factory/nodeTests.ts:1124:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocOverrideTag at src/compiler/factory/nodeTests.ts:1128:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocOverloadTag at src/compiler/factory/nodeTests.ts:1132:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocDeprecatedTag at src/compiler/factory/nodeTests.ts:1136:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocSeeTag at src/compiler/factory/nodeTests.ts:1140:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocEnumTag at src/compiler/factory/nodeTests.ts:1144:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocParameterTag at src/compiler/factory/nodeTests.ts:1148:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocReturnTag at src/compiler/factory/nodeTests.ts:1152:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocThisTag at src/compiler/factory/nodeTests.ts:1156:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTypeTag at src/compiler/factory/nodeTests.ts:1160:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTemplateTag at src/compiler/factory/nodeTests.ts:1164:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTypedefTag at src/compiler/factory/nodeTests.ts:1168:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocUnknownTag at src/compiler/factory/nodeTests.ts:1172:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocPropertyTag at src/compiler/factory/nodeTests.ts:1176:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocImplementsTag at src/compiler/factory/nodeTests.ts:1180:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocSatisfiesTag at src/compiler/factory/nodeTests.ts:1184:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocThrowsTag at src/compiler/factory/nodeTests.ts:1188:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocImportTag at src/compiler/factory/nodeTests.ts:1192:1 | boolean | non-node-builder | unproven | unproven |
| isSyntaxList at src/compiler/factory/nodeTests.ts:1199:1 | boolean | non-node-builder | unproven | unproven |
| getNodeChildren at src/compiler/factory/nodeChildren.ts:14:1 | readonly Node[] &#124; undefined | non-node-builder | unproven | unproven |
| setNodeChildren at src/compiler/factory/nodeChildren.ts:27:1 | readonly Node[] | non-node-builder | unproven | unproven |
| unsetNodeChildren at src/compiler/factory/nodeChildren.ts:44:1 | void | non-node-builder | unproven | unproven |
| createEmptyExports at src/compiler/factory/utilities.ts:183:1 | ExportDeclaration | composed-or-forwarding | yes | unproven |
| createMemberAccessForPropertyName at src/compiler/factory/utilities.ts:188:1 | MemberExpression | composed-or-forwarding | unproven | unproven |
| createReactNamespace at src/compiler/factory/utilities.ts:204:1 | Identifier | composed-or-forwarding | unproven | unproven |
| createJsxFactoryExpressionFromEntityName at src/compiler/factory/utilities.ts:215:1 | Expression | composed-or-forwarding | unproven | unproven |
| createJsxFactoryExpression at src/compiler/factory/utilities.ts:228:1 | Expression | composed-or-forwarding | unproven | unproven |
| createJsxFragmentFactoryExpression at src/compiler/factory/utilities.ts:237:1 | Expression | composed-or-forwarding | unproven | unproven |
| createExpressionForJsxElement at src/compiler/factory/utilities.ts:247:1 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| createExpressionForJsxFragment at src/compiler/factory/utilities.ts:280:1 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| createForOfBindingStatement at src/compiler/factory/utilities.ts:309:1 | Statement | composed-or-forwarding | unproven | unproven |
| createExpressionFromEntityName at src/compiler/factory/utilities.ts:334:1 | Expression | composed-or-forwarding | unproven | unproven |
| createExpressionForPropertyName at src/compiler/factory/utilities.ts:348:1 | Expression | composed-or-forwarding | unproven | unproven |
| createExpressionForAccessorDeclaration at src/compiler/factory/utilities.ts:362:1 | CallExpression &#124; undefined | composed-or-forwarding | unproven | unproven |
| createExpressionForPropertyAssignment at src/compiler/factory/utilities.ts:411:1 | AssignmentExpression<EqualsToken> | composed-or-forwarding | unproven | unproven |
| createExpressionForShorthandPropertyAssignment at src/compiler/factory/utilities.ts:424:1 | AssignmentExpression<EqualsToken> | composed-or-forwarding | unproven | unproven |
| createExpressionForMethodDeclaration at src/compiler/factory/utilities.ts:437:1 | AssignmentExpression<EqualsToken> | composed-or-forwarding | unproven | unproven |
| createExpressionForObjectLiteralElementLike at src/compiler/factory/utilities.ts:465:1 | Expression &#124; undefined | composed-or-forwarding | unproven | unproven |
| expandPreOrPostfixIncrementOrDecrementExpression at src/compiler/factory/utilities.ts:516:1 | Expression | composed-or-forwarding | unproven | unproven |
| isInternalName at src/compiler/factory/utilities.ts:556:1 | boolean | non-node-builder | unproven | unproven |
| isLocalName at src/compiler/factory/utilities.ts:565:1 | boolean | non-node-builder | unproven | unproven |
| isExportName at src/compiler/factory/utilities.ts:575:1 | boolean | non-node-builder | unproven | unproven |
| isUseStrictPrologue at src/compiler/factory/utilities.ts:579:1 | boolean | non-node-builder | unproven | unproven |
| findUseStrictPrologue at src/compiler/factory/utilities.ts:584:1 | Statement &#124; undefined | input-or-unresolved | unproven | unproven |
| isCommaExpression at src/compiler/factory/utilities.ts:607:1 | boolean | non-node-builder | unproven | unproven |
| isCommaSequence at src/compiler/factory/utilities.ts:612:1 | boolean | non-node-builder | unproven | unproven |
| isJSDocTypeAssertion at src/compiler/factory/utilities.ts:617:1 | boolean | non-node-builder | unproven | unproven |
| getJSDocTypeAssertionType at src/compiler/factory/utilities.ts:624:1 | TypeNode | input-or-unresolved | unproven | unproven |
| isOuterExpression at src/compiler/factory/utilities.ts:631:1 | boolean | non-node-builder | unproven | unproven |
| skipOuterExpressions at src/compiler/factory/utilities.ts:660:1 | Node | input-or-unresolved | unproven | unproven |
| walkUpOuterExpressions at src/compiler/factory/utilities.ts:668:1 | Node | input-or-unresolved | unproven | unproven |
| startOnNewLine at src/compiler/factory/utilities.ts:678:1 | T | composed-or-forwarding | unproven | unproven |
| getExternalHelpersModuleName at src/compiler/factory/utilities.ts:683:1 | Identifier &#124; undefined | input-or-unresolved | unproven | unproven |
| hasRecordedExternalHelpers at src/compiler/factory/utilities.ts:690:1 | boolean | non-node-builder | unproven | unproven |
| createExternalHelpersImportDeclarationIfNeeded at src/compiler/factory/utilities.ts:697:1 | ImportEqualsDeclaration &#124; ImportDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| <anonymous@723:42> at src/compiler/factory/utilities.ts:723:42 | ImportSpecifier | composed-or-forwarding | yes | unproven |
| getImportedHelpers at src/compiler/factory/utilities.ts:760:1 | UnscopedEmitHelper[] &#124; undefined | non-node-builder | unproven | unproven |
| getOrCreateExternalHelpersModuleNameIfNeeded at src/compiler/factory/utilities.ts:764:1 | Identifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| getLocalNameForExternalImport at src/compiler/factory/utilities.ts:786:1 | Identifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| getExternalModuleNameLiteral at src/compiler/factory/utilities.ts:814:1 | StringLiteral &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryRenameExternalModule at src/compiler/factory/utilities.ts:829:1 | StringLiteral &#124; undefined | composed-or-forwarding | yes | unproven |
| tryGetModuleNameFromFile at src/compiler/factory/utilities.ts:843:1 | StringLiteral &#124; undefined | composed-or-forwarding | yes | unproven |
| tryGetModuleNameFromDeclaration at src/compiler/factory/utilities.ts:856:1 | StringLiteral &#124; undefined | composed-or-forwarding | unproven | unproven |
| getInitializerOfBindingOrAssignmentElement at src/compiler/factory/utilities.ts:865:1 | Expression &#124; undefined | composed-or-forwarding | unproven | unproven |
| getTargetOfBindingOrAssignmentElement at src/compiler/factory/utilities.ts:910:1 | BindingOrAssignmentElementTarget &#124; undefined | composed-or-forwarding | unproven | unproven |
| getRestIndicatorOfBindingOrAssignmentElement at src/compiler/factory/utilities.ts:987:1 | BindingOrAssignmentElementRestIndicator &#124; undefined | input-or-unresolved | unproven | unproven |
| getPropertyNameOfBindingOrAssignmentElement at src/compiler/factory/utilities.ts:1008:1 | Identifier &#124; StringLiteral &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; ComputedPropertyName &#124; BigIntLiteral &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryGetPropertyNameOfBindingOrAssignmentElement at src/compiler/factory/utilities.ts:1015:1 | Identifier &#124; StringLiteral &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; ComputedPropertyName &#124; BigIntLiteral &#124; undefined | composed-or-forwarding | unproven | unproven |
| isStringOrNumericLiteral at src/compiler/factory/utilities.ts:1065:1 | boolean | non-node-builder | unproven | unproven |
| getElementsOfBindingOrAssignmentPattern at src/compiler/factory/utilities.ts:1076:1 | readonly BindingOrAssignmentElement[] | non-node-builder | unproven | unproven |
| getJSDocTypeAliasName at src/compiler/factory/utilities.ts:1092:1 | Identifier &#124; undefined | input-or-unresolved | unproven | unproven |
| canHaveIllegalType at src/compiler/factory/utilities.ts:1105:1 | boolean | non-node-builder | unproven | unproven |
| canHaveIllegalTypeParameters at src/compiler/factory/utilities.ts:1112:1 | boolean | non-node-builder | unproven | unproven |
| canHaveIllegalDecorators at src/compiler/factory/utilities.ts:1120:1 | boolean | non-node-builder | unproven | unproven |
| canHaveIllegalModifiers at src/compiler/factory/utilities.ts:1142:1 | boolean | non-node-builder | unproven | unproven |
| isQuestionOrExclamationToken at src/compiler/factory/utilities.ts:1151:1 | boolean | non-node-builder | unproven | unproven |
| isIdentifierOrThisTypeNode at src/compiler/factory/utilities.ts:1155:1 | boolean | non-node-builder | unproven | unproven |
| isReadonlyKeywordOrPlusOrMinusToken at src/compiler/factory/utilities.ts:1159:1 | boolean | non-node-builder | unproven | unproven |
| isQuestionOrPlusOrMinusToken at src/compiler/factory/utilities.ts:1163:1 | boolean | non-node-builder | unproven | unproven |
| isModuleName at src/compiler/factory/utilities.ts:1167:1 | boolean | non-node-builder | unproven | unproven |
| isBinaryOperatorToken at src/compiler/factory/utilities.ts:1267:1 | boolean | non-node-builder | unproven | unproven |
| pushStack at src/compiler/factory/utilities.ts:1392:5 | number | non-node-builder | unproven | unproven |
| checkCircularity at src/compiler/factory/utilities.ts:1400:5 | void | non-node-builder | unproven | unproven |
| createBinaryExpressionTrampoline at src/compiler/factory/utilities.ts:1464:1 | (node: BinaryExpression, outerState: TOuterState) => TResult | non-node-builder | unproven | unproven |
| trampoline at src/compiler/factory/utilities.ts:1475:5 | TResult | non-node-builder | unproven | unproven |
| isExportOrDefaultModifier at src/compiler/factory/utilities.ts:1494:1 | boolean | non-node-builder | unproven | unproven |
| getNodeForGeneratedName at src/compiler/factory/utilities.ts:1518:1 | Node &#124; GeneratedIdentifier &#124; GeneratedPrivateIdentifier | input-or-unresolved | unproven | unproven |
| formatIdentifier at src/compiler/factory/utilities.ts:1565:1 | string | non-node-builder | unproven | unproven |
| formatIdentifierWorker at src/compiler/factory/utilities.ts:1570:1 | string | non-node-builder | unproven | unproven |
| formatGeneratedName at src/compiler/factory/utilities.ts:1599:1 | string | non-node-builder | unproven | unproven |
| createAccessorPropertyBackingField at src/compiler/factory/utilities.ts:1611:1 | PropertyDeclaration | composed-or-forwarding | unproven | unproven |
| createAccessorPropertyGetRedirector at src/compiler/factory/utilities.ts:1627:1 | GetAccessorDeclaration | composed-or-forwarding | yes | unproven |
| createAccessorPropertySetRedirector at src/compiler/factory/utilities.ts:1649:1 | SetAccessorDeclaration | composed-or-forwarding | yes | unproven |
| findComputedPropertyNameCacheAssignment at src/compiler/factory/utilities.ts:1673:1 | (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | composed-or-forwarding | unproven | unproven |
| isSyntheticParenthesizedExpression at src/compiler/factory/utilities.ts:1695:1 | boolean | non-node-builder | unproven | unproven |
| flattenCommaListWorker at src/compiler/factory/utilities.ts:1701:1 | void | non-node-builder | unproven | unproven |
| flattenCommaList at src/compiler/factory/utilities.ts:1725:1 | Expression[] | non-node-builder | unproven | unproven |
| containsObjectRestOrSpread at src/compiler/factory/utilities.ts:1739:1 | boolean | non-node-builder | unproven | unproven |
| canHaveModifiers at src/compiler/factory/utilitiesPublic.ts:14:1 | boolean | non-node-builder | unproven | unproven |
| canHaveDecorators at src/compiler/factory/utilitiesPublic.ts:43:1 | boolean | non-node-builder | unproven | unproven |
| createBaseSourceFileNode at src/compiler/parser.ts:433:31 | Node | allocating | yes | unproven |
| createBaseIdentifierNode at src/compiler/parser.ts:434:31 | Node | allocating | unproven | unproven |
| createBasePrivateIdentifierNode at src/compiler/parser.ts:435:38 | Node | allocating | yes | unproven |
| createBaseTokenNode at src/compiler/parser.ts:436:26 | Node | allocating | unproven | unproven |
| createBaseNode at src/compiler/parser.ts:437:21 | Node | allocating | yes | unproven |
| visitNode at src/compiler/parser.ts:443:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| isFileProbablyExternalModule at src/compiler/parser.ts:469:1 | Node &#124; undefined | composed-or-forwarding | unproven | unproven |
| isAnExternalModuleIndicatorNode at src/compiler/parser.ts:476:1 | HasModifiers &#124; undefined | input-or-unresolved | unproven | unproven |
| getImportMetaIfNecessary at src/compiler/parser.ts:484:1 | Node &#124; undefined | composed-or-forwarding | unproven | unproven |
| walkTreeForImportMeta at src/compiler/parser.ts:490:1 | Node &#124; undefined | input-or-unresolved | unproven | unproven |
| hasModifierOfKind at src/compiler/parser.ts:495:1 | boolean | non-node-builder | unproven | unproven |
| <anonymous@496:33> at src/compiler/parser.ts:496:33 | boolean | non-node-builder | unproven | unproven |
| isImportMeta at src/compiler/parser.ts:499:1 | boolean | non-node-builder | unproven | unproven |
| forEachChildInQualifiedName at src/compiler/parser.ts:506:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeParameter at src/compiler/parser.ts:510:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInShorthandPropertyAssignment at src/compiler/parser.ts:517:47 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSpreadAssignment at src/compiler/parser.ts:525:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInParameter at src/compiler/parser.ts:528:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPropertyDeclaration at src/compiler/parser.ts:536:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPropertySignature at src/compiler/parser.ts:544:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPropertyAssignment at src/compiler/parser.ts:551:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInVariableDeclaration at src/compiler/parser.ts:558:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInBindingElement at src/compiler/parser.ts:564:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInIndexSignature at src/compiler/parser.ts:570:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInConstructorType at src/compiler/parser.ts:576:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInFunctionType at src/compiler/parser.ts:582:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInMethodDeclaration at src/compiler/parser.ts:590:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInMethodSignature at src/compiler/parser.ts:601:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInConstructor at src/compiler/parser.ts:609:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInGetAccessor at src/compiler/parser.ts:617:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSetAccessor at src/compiler/parser.ts:625:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInFunctionDeclaration at src/compiler/parser.ts:633:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInFunctionExpression at src/compiler/parser.ts:642:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInArrowFunction at src/compiler/parser.ts:651:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInClassStaticBlockDeclaration at src/compiler/parser.ts:659:47 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeReference at src/compiler/parser.ts:663:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypePredicate at src/compiler/parser.ts:667:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeQuery at src/compiler/parser.ts:672:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeLiteral at src/compiler/parser.ts:676:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInArrayType at src/compiler/parser.ts:679:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTupleType at src/compiler/parser.ts:682:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInConditionalType at src/compiler/parser.ts:687:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInInferType at src/compiler/parser.ts:693:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportType at src/compiler/parser.ts:696:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportTypeAssertionContainer at src/compiler/parser.ts:702:48 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInIndexedAccessType at src/compiler/parser.ts:707:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInMappedType at src/compiler/parser.ts:711:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInLiteralType at src/compiler/parser.ts:719:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNamedTupleMember at src/compiler/parser.ts:722:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInArrayLiteralExpression at src/compiler/parser.ts:730:42 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInObjectLiteralExpression at src/compiler/parser.ts:733:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPropertyAccessExpression at src/compiler/parser.ts:736:44 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInElementAccessExpression at src/compiler/parser.ts:741:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTaggedTemplateExpression at src/compiler/parser.ts:748:44 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeAssertionExpression at src/compiler/parser.ts:754:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInParenthesizedExpression at src/compiler/parser.ts:758:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInDeleteExpression at src/compiler/parser.ts:761:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeOfExpression at src/compiler/parser.ts:764:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInVoidExpression at src/compiler/parser.ts:767:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPrefixUnaryExpression at src/compiler/parser.ts:770:41 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInYieldExpression at src/compiler/parser.ts:773:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInAwaitExpression at src/compiler/parser.ts:777:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPostfixUnaryExpression at src/compiler/parser.ts:780:42 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInBinaryExpression at src/compiler/parser.ts:783:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInAsExpression at src/compiler/parser.ts:788:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNonNullExpression at src/compiler/parser.ts:792:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSatisfiesExpression at src/compiler/parser.ts:795:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInMetaProperty at src/compiler/parser.ts:798:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInConditionalExpression at src/compiler/parser.ts:801:41 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSpreadElement at src/compiler/parser.ts:808:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSourceFile at src/compiler/parser.ts:813:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInVariableStatement at src/compiler/parser.ts:817:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInVariableDeclarationList at src/compiler/parser.ts:821:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInExpressionStatement at src/compiler/parser.ts:824:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInIfStatement at src/compiler/parser.ts:827:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInDoStatement at src/compiler/parser.ts:832:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInWhileStatement at src/compiler/parser.ts:836:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInForStatement at src/compiler/parser.ts:840:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInForInStatement at src/compiler/parser.ts:846:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInForOfStatement at src/compiler/parser.ts:851:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInReturnStatement at src/compiler/parser.ts:859:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInWithStatement at src/compiler/parser.ts:862:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInSwitchStatement at src/compiler/parser.ts:866:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCaseBlock at src/compiler/parser.ts:870:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCaseClause at src/compiler/parser.ts:873:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInDefaultClause at src/compiler/parser.ts:877:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInLabeledStatement at src/compiler/parser.ts:880:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInThrowStatement at src/compiler/parser.ts:884:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTryStatement at src/compiler/parser.ts:887:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCatchClause at src/compiler/parser.ts:892:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInDecorator at src/compiler/parser.ts:896:29 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInInterfaceDeclaration at src/compiler/parser.ts:901:40 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTypeAliasDeclaration at src/compiler/parser.ts:908:40 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInEnumDeclaration at src/compiler/parser.ts:914:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInEnumMember at src/compiler/parser.ts:919:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInModuleDeclaration at src/compiler/parser.ts:923:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportEqualsDeclaration at src/compiler/parser.ts:928:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportDeclaration at src/compiler/parser.ts:933:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportClause at src/compiler/parser.ts:939:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportAttributes at src/compiler/parser.ts:943:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportAttribute at src/compiler/parser.ts:946:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNamespaceExportDeclaration at src/compiler/parser.ts:950:46 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNamespaceImport at src/compiler/parser.ts:954:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNamespaceExport at src/compiler/parser.ts:957:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInExportDeclaration at src/compiler/parser.ts:962:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInExportAssignment at src/compiler/parser.ts:970:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTemplateExpression at src/compiler/parser.ts:974:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTemplateSpan at src/compiler/parser.ts:978:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTemplateLiteralType at src/compiler/parser.ts:982:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInTemplateLiteralTypeSpan at src/compiler/parser.ts:986:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInComputedPropertyName at src/compiler/parser.ts:990:40 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInHeritageClause at src/compiler/parser.ts:993:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInExpressionWithTypeArguments at src/compiler/parser.ts:996:47 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInExternalModuleReference at src/compiler/parser.ts:1000:43 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInMissingDeclaration at src/compiler/parser.ts:1003:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCommaListExpression at src/compiler/parser.ts:1006:39 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxElement at src/compiler/parser.ts:1009:30 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxFragment at src/compiler/parser.ts:1014:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxAttributes at src/compiler/parser.ts:1021:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxAttribute at src/compiler/parser.ts:1024:32 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxSpreadAttribute at src/compiler/parser.ts:1028:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxExpression at src/compiler/parser.ts:1031:33 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxClosingElement at src/compiler/parser.ts:1035:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxNamespacedName at src/compiler/parser.ts:1038:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocFunctionType at src/compiler/parser.ts:1049:37 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDoc at src/compiler/parser.ts:1053:25 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocSeeTag at src/compiler/parser.ts:1057:31 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocNameReference at src/compiler/parser.ts:1062:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocMemberName at src/compiler/parser.ts:1065:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocAuthorTag at src/compiler/parser.ts:1071:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocImplementsTag at src/compiler/parser.ts:1075:38 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocAugmentsTag at src/compiler/parser.ts:1080:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocTemplateTag at src/compiler/parser.ts:1085:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocTypedefTag at src/compiler/parser.ts:1091:35 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocCallbackTag at src/compiler/parser.ts:1102:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocSignature at src/compiler/parser.ts:1115:34 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocTypeLiteral at src/compiler/parser.ts:1123:36 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCallOrConstructSignature at src/compiler/parser.ts:1140:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInUnionOrIntersectionType at src/compiler/parser.ts:1146:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInParenthesizedTypeOrTypeOperator at src/compiler/parser.ts:1150:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInObjectOrArrayBindingPattern at src/compiler/parser.ts:1154:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInCallOrNewExpression at src/compiler/parser.ts:1158:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInBlock at src/compiler/parser.ts:1166:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInContinueOrBreakStatement at src/compiler/parser.ts:1170:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInClassDeclarationOrExpression at src/compiler/parser.ts:1174:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInNamedImportsOrExports at src/compiler/parser.ts:1182:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInImportOrExportSpecifier at src/compiler/parser.ts:1186:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJsxOpeningOrSelfClosingElement at src/compiler/parser.ts:1191:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInOptionalRestOrJSDocParameterModifier at src/compiler/parser.ts:1197:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocParameterOrPropertyTag at src/compiler/parser.ts:1201:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocTypeLikeTag at src/compiler/parser.ts:1209:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocLinkCodeOrPlain at src/compiler/parser.ts:1215:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocTag at src/compiler/parser.ts:1219:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInJSDocImportTag at src/compiler/parser.ts:1224:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildInPartiallyEmittedExpression at src/compiler/parser.ts:1232:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChild at src/compiler/parser.ts:1249:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| forEachChildRecursively at src/compiler/parser.ts:1272:1 | T &#124; undefined | non-node-builder | unproven | unproven |
| gatherPossibleChildren at src/compiler/parser.ts:1311:1 | (Node &#124; NodeArray<Node>)[] | non-node-builder | unproven | unproven |
| addWorkItem at src/compiler/parser.ts:1316:5 | void | non-node-builder | unproven | unproven |
| setExternalModuleIndicator at src/compiler/parser.ts:1340:1 | void | non-node-builder | unproven | unproven |
| createSourceFile at src/compiler/parser.ts:1344:1 | SourceFile | composed-or-forwarding | unproven | unproven |
| <anonymous@1359:90> at src/compiler/parser.ts:1359:90 | void | non-node-builder | unproven | unproven |
| parseIsolatedEntityName at src/compiler/parser.ts:1372:1 | EntityName &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJsonText at src/compiler/parser.ts:1381:1 | JsonSourceFile | composed-or-forwarding | unproven | unproven |
| isExternalModule at src/compiler/parser.ts:1386:1 | boolean | non-node-builder | unproven | unproven |
| updateSourceFile at src/compiler/parser.ts:1399:1 | SourceFile | composed-or-forwarding | unproven | unproven |
| countNode at src/compiler/parser.ts:1455:5 | Node | input-or-unresolved | unproven | unproven |
| createBaseSourceFileNode at src/compiler/parser.ts:1463:35 | Node | allocating | unproven | unproven |
| createBaseIdentifierNode at src/compiler/parser.ts:1464:35 | Node | allocating | unproven | unproven |
| createBasePrivateIdentifierNode at src/compiler/parser.ts:1465:42 | Node | allocating | unproven | unproven |
| createBaseTokenNode at src/compiler/parser.ts:1466:30 | Node | allocating | unproven | unproven |
| createBaseNode at src/compiler/parser.ts:1467:25 | Node | allocating | unproven | unproven |
| parseSourceFile at src/compiler/parser.ts:1603:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| parseIsolatedEntityName at src/compiler/parser.ts:1635:5 | EntityName &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJsonText at src/compiler/parser.ts:1646:5 | JsonSourceFile | composed-or-forwarding | unproven | unproven |
| parseSourceFileWorker at src/compiler/parser.ts:1803:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| withJSDoc at src/compiler/parser.ts:1847:5 | T | input-or-unresolved | unproven | unproven |
| <anonymous@1853:75> at src/compiler/parser.ts:1853:75 | JSDoc &#124; undefined | composed-or-forwarding | unproven | unproven |
| reparseTopLevelAwait at src/compiler/parser.ts:1862:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| containsPossibleTopLevelAwait at src/compiler/parser.ts:1938:9 | boolean | non-node-builder | unproven | unproven |
| currentNode at src/compiler/parser.ts:1961:9 | Node | input-or-unresolved | unproven | unproven |
| fixupParentReferences at src/compiler/parser.ts:1970:5 | void | non-node-builder | unproven | unproven |
| createSourceFile at src/compiler/parser.ts:1978:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| setFields at src/compiler/parser.ts:2003:9 | void | non-node-builder | unproven | unproven |
| parseErrorForMissingSemicolonAfter at src/compiler/parser.ts:2364:5 | void | non-node-builder | unproven | unproven |
| parseSemicolonAfterPropertyName at src/compiler/parser.ts:2454:5 | void | non-node-builder | unproven | unproven |
| parseOptionalToken at src/compiler/parser.ts:2524:5 | Node &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseOptionalTokenJSDoc at src/compiler/parser.ts:2532:5 | Node &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseExpectedToken at src/compiler/parser.ts:2540:5 | Node | input-or-unresolved | unproven | unproven |
| parseExpectedTokenJSDoc at src/compiler/parser.ts:2546:5 | Node | input-or-unresolved | unproven | unproven |
| parseTokenNode at src/compiler/parser.ts:2553:5 | T | composed-or-forwarding | unproven | unproven |
| parseTokenNodeJSDoc at src/compiler/parser.ts:2560:5 | T | composed-or-forwarding | unproven | unproven |
| finishNode at src/compiler/parser.ts:2600:5 | T | input-or-unresolved | unproven | unproven |
| createMissingNode at src/compiler/parser.ts:2619:5 | T | composed-or-forwarding | unproven | unproven |
| createIdentifier at src/compiler/parser.ts:2648:5 | Identifier | composed-or-forwarding | unproven | unproven |
| parseBindingIdentifier at src/compiler/parser.ts:2684:5 | Identifier | composed-or-forwarding | unproven | unproven |
| parseIdentifier at src/compiler/parser.ts:2688:5 | Identifier | composed-or-forwarding | unproven | unproven |
| parseIdentifierName at src/compiler/parser.ts:2692:5 | Identifier | composed-or-forwarding | unproven | unproven |
| parseIdentifierNameErrorOnUnicodeEscapeSequence at src/compiler/parser.ts:2696:5 | Identifier | composed-or-forwarding | unproven | unproven |
| parsePropertyNameWorker at src/compiler/parser.ts:2714:5 | PropertyName | composed-or-forwarding | unproven | unproven |
| parsePropertyName at src/compiler/parser.ts:2729:5 | PropertyName | composed-or-forwarding | unproven | unproven |
| parseComputedPropertyName at src/compiler/parser.ts:2733:5 | ComputedPropertyName | composed-or-forwarding | unproven | unproven |
| parsePrivateIdentifier at src/compiler/parser.ts:2747:5 | PrivateIdentifier | composed-or-forwarding | unproven | unproven |
| parseListElement at src/compiler/parser.ts:3116:5 | T | composed-or-forwarding | unproven | unproven |
| currentNode at src/compiler/parser.ts:3125:5 | Node &#124; undefined | input-or-unresolved | unproven | unproven |
| consumeNode at src/compiler/parser.ts:3179:5 | Node | input-or-unresolved | unproven | unproven |
| canReuseNode at src/compiler/parser.ts:3203:5 | boolean | non-node-builder | unproven | unproven |
| isReusableClassMember at src/compiler/parser.ts:3281:5 | boolean | non-node-builder | unproven | unproven |
| isReusableSwitchClause at src/compiler/parser.ts:3306:5 | boolean | non-node-builder | unproven | unproven |
| isReusableStatement at src/compiler/parser.ts:3318:5 | boolean | non-node-builder | unproven | unproven |
| isReusableEnumMember at src/compiler/parser.ts:3357:5 | boolean | non-node-builder | unproven | unproven |
| isReusableTypeMember at src/compiler/parser.ts:3361:5 | boolean | non-node-builder | unproven | unproven |
| isReusableVariableDeclaration at src/compiler/parser.ts:3376:5 | boolean | non-node-builder | unproven | unproven |
| isReusableParameter at src/compiler/parser.ts:3399:5 | boolean | non-node-builder | unproven | unproven |
| parseEntityName at src/compiler/parser.ts:3589:5 | EntityName | composed-or-forwarding | unproven | unproven |
| createQualifiedName at src/compiler/parser.ts:3609:5 | QualifiedName | composed-or-forwarding | unproven | unproven |
| parseRightSideOfDot at src/compiler/parser.ts:3613:5 | Identifier &#124; PrivateIdentifier | composed-or-forwarding | unproven | unproven |
| parseTemplateExpression at src/compiler/parser.ts:3668:5 | TemplateExpression | composed-or-forwarding | unproven | unproven |
| parseTemplateType at src/compiler/parser.ts:3679:5 | TemplateLiteralTypeNode | composed-or-forwarding | unproven | unproven |
| parseTemplateTypeSpan at src/compiler/parser.ts:3702:5 | TemplateLiteralTypeSpan | composed-or-forwarding | unproven | unproven |
| parseLiteralOfTemplateSpan at src/compiler/parser.ts:3713:5 | TemplateMiddle &#124; TemplateTail | composed-or-forwarding | unproven | unproven |
| parseTemplateSpan at src/compiler/parser.ts:3724:5 | TemplateSpan | composed-or-forwarding | unproven | unproven |
| parseLiteralNode at src/compiler/parser.ts:3735:5 | LiteralExpression | composed-or-forwarding | unproven | unproven |
| parseTemplateHead at src/compiler/parser.ts:3739:5 | TemplateHead | composed-or-forwarding | unproven | unproven |
| parseTemplateMiddleOrTemplateTail at src/compiler/parser.ts:3748:5 | TemplateMiddle &#124; TemplateTail | composed-or-forwarding | unproven | unproven |
| parseLiteralLikeNode at src/compiler/parser.ts:3760:5 | LiteralLikeNode | composed-or-forwarding | unproven | unproven |
| parseEntityNameOfTypeReference at src/compiler/parser.ts:3787:5 | EntityName | composed-or-forwarding | unproven | unproven |
| parseTypeReference at src/compiler/parser.ts:3797:5 | TypeReferenceNode | composed-or-forwarding | unproven | unproven |
| typeHasArrowFunctionBlockingParseError at src/compiler/parser.ts:3809:5 | boolean | non-node-builder | unproven | unproven |
| parseThisTypePredicate at src/compiler/parser.ts:3825:5 | TypePredicateNode | composed-or-forwarding | unproven | unproven |
| parseThisTypeNode at src/compiler/parser.ts:3830:5 | ThisTypeNode | composed-or-forwarding | unproven | unproven |
| parseJSDocAllType at src/compiler/parser.ts:3836:5 | JSDocAllType &#124; JSDocOptionalType | composed-or-forwarding | unproven | unproven |
| parseJSDocNonNullableType at src/compiler/parser.ts:3842:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseJSDocUnknownOrNullableType at src/compiler/parser.ts:3848:5 | JSDocUnknownType &#124; JSDocNullableType | composed-or-forwarding | unproven | unproven |
| parseJSDocFunctionType at src/compiler/parser.ts:3878:5 | JSDocFunctionType &#124; TypeReferenceNode | composed-or-forwarding | unproven | unproven |
| parseJSDocParameter at src/compiler/parser.ts:3889:5 | ParameterDeclaration | composed-or-forwarding | unproven | unproven |
| parseJSDocType at src/compiler/parser.ts:3910:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTypeQuery at src/compiler/parser.ts:3946:5 | TypeQueryNode | composed-or-forwarding | unproven | unproven |
| parseTypeParameter at src/compiler/parser.ts:3955:5 | TypeParameterDeclaration | composed-or-forwarding | unproven | unproven |
| parseNameOfParameter at src/compiler/parser.ts:4001:5 | Identifier &#124; BindingPattern | composed-or-forwarding | unproven | unproven |
| parseParameter at src/compiler/parser.ts:4026:5 | ParameterDeclaration | input-or-unresolved | unproven | unproven |
| parseParameterForSpeculation at src/compiler/parser.ts:4030:5 | ParameterDeclaration &#124; undefined | input-or-unresolved | unproven | unproven |
| parseParameterWorker at src/compiler/parser.ts:4036:5 | ParameterDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseReturnType at src/compiler/parser.ts:4095:5 | TypeNode &#124; undefined | input-or-unresolved | unproven | unproven |
| <anonymous@4142:59> at src/compiler/parser.ts:4142:59 | ParameterDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseSignatureMember at src/compiler/parser.ts:4184:5 | CallSignatureDeclaration &#124; ConstructSignatureDeclaration | composed-or-forwarding | unproven | unproven |
| parseIndexSignatureDeclaration at src/compiler/parser.ts:4260:5 | IndexSignatureDeclaration | composed-or-forwarding | unproven | unproven |
| <anonymous@4261:96> at src/compiler/parser.ts:4261:96 | ParameterDeclaration | composed-or-forwarding | unproven | unproven |
| parsePropertyOrMethodSignature at src/compiler/parser.ts:4268:5 | PropertySignature &#124; MethodSignature | composed-or-forwarding | unproven | unproven |
| parseTypeMember at src/compiler/parser.ts:4330:5 | TypeElement | composed-or-forwarding | unproven | unproven |
| parseTypeLiteral at src/compiler/parser.ts:4373:5 | TypeLiteralNode | composed-or-forwarding | unproven | unproven |
| parseMappedTypeParameter at src/compiler/parser.ts:4402:5 | TypeParameterDeclaration | composed-or-forwarding | unproven | unproven |
| parseMappedType at src/compiler/parser.ts:4410:5 | MappedTypeNode | composed-or-forwarding | unproven | unproven |
| parseTupleElementType at src/compiler/parser.ts:4438:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTupleElementNameOrTupleElementType at src/compiler/parser.ts:4464:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTupleType at src/compiler/parser.ts:4479:5 | TupleTypeNode | composed-or-forwarding | unproven | unproven |
| parseParenthesizedType at src/compiler/parser.ts:4489:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseFunctionOrConstructorType at src/compiler/parser.ts:4508:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseKeywordAndNoDot at src/compiler/parser.ts:4523:5 | TypeNode &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseLiteralTypeNode at src/compiler/parser.ts:4528:5 | LiteralTypeNode | composed-or-forwarding | unproven | unproven |
| parseImportType at src/compiler/parser.ts:4547:5 | ImportTypeNode | composed-or-forwarding | unproven | unproven |
| parseNonArrayType at src/compiler/parser.ts:4589:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parsePostfixTypeOrHigher at src/compiler/parser.ts:4716:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTypeOperator at src/compiler/parser.ts:4752:5 | TypeOperatorNode | composed-or-forwarding | unproven | unproven |
| tryParseConstraintOfInferType at src/compiler/parser.ts:4758:5 | TypeNode &#124; undefined | input-or-unresolved | unproven | unproven |
| parseTypeParameterOfInferType at src/compiler/parser.ts:4767:5 | TypeParameterDeclaration | composed-or-forwarding | unproven | unproven |
| parseInferType at src/compiler/parser.ts:4775:5 | InferTypeNode | composed-or-forwarding | unproven | unproven |
| parseTypeOperatorOrHigher at src/compiler/parser.ts:4781:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseFunctionOrConstructorTypeToError at src/compiler/parser.ts:4794:5 | TypeNode &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseUnionOrIntersectionType at src/compiler/parser.ts:4819:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseIntersectionTypeOrHigher at src/compiler/parser.ts:4839:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseUnionTypeOrHigher at src/compiler/parser.ts:4843:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTypeOrTypePredicate at src/compiler/parser.ts:4912:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTypePredicatePrefix at src/compiler/parser.ts:4924:5 | Identifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseAssertsTypePredicate at src/compiler/parser.ts:4932:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseType at src/compiler/parser.ts:4940:5 | TypeNode | composed-or-forwarding | unproven | unproven |
| parseTypeAnnotation at src/compiler/parser.ts:4961:5 | TypeNode &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseExpression at src/compiler/parser.ts:5041:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseInitializer at src/compiler/parser.ts:5065:5 | Expression &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseAssignmentExpressionOrHigher at src/compiler/parser.ts:5069:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseYieldExpression at src/compiler/parser.ts:5169:5 | YieldExpression | composed-or-forwarding | unproven | unproven |
| parseSimpleArrowFunctionExpression at src/compiler/parser.ts:5197:5 | ArrowFunction | composed-or-forwarding | unproven | unproven |
| tryParseParenthesizedArrowFunctionExpression at src/compiler/parser.ts:5216:5 | Expression &#124; undefined | composed-or-forwarding | unproven | unproven |
| <anonymous@5229:22> at src/compiler/parser.ts:5229:22 | ArrowFunction &#124; undefined | composed-or-forwarding | unproven | unproven |
| parsePossibleParenthesizedArrowFunctionExpression at src/compiler/parser.ts:5381:5 | ArrowFunction &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryParseAsyncSimpleArrowFunctionExpression at src/compiler/parser.ts:5395:5 | ArrowFunction &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseParenthesizedArrowFunctionExpression at src/compiler/parser.ts:5430:5 | ArrowFunction &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseArrowFunctionExpressionBody at src/compiler/parser.ts:5533:5 | Expression &#124; Block | composed-or-forwarding | unproven | unproven |
| <anonymous@5567:32> at src/compiler/parser.ts:5567:32 | Expression | composed-or-forwarding | unproven | unproven |
| <anonymous@5568:39> at src/compiler/parser.ts:5568:39 | Expression | composed-or-forwarding | unproven | unproven |
| parseConditionalExpressionRest at src/compiler/parser.ts:5574:5 | Expression | composed-or-forwarding | unproven | unproven |
| <anonymous@5588:67> at src/compiler/parser.ts:5588:67 | Expression | composed-or-forwarding | unproven | unproven |
| parseBinaryExpressionOrHigher at src/compiler/parser.ts:5598:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseBinaryExpressionRest at src/compiler/parser.ts:5608:5 | Expression | composed-or-forwarding | unproven | unproven |
| makeSatisfiesExpression at src/compiler/parser.ts:5681:5 | SatisfiesExpression | composed-or-forwarding | unproven | unproven |
| makeBinaryExpression at src/compiler/parser.ts:5685:5 | BinaryExpression | composed-or-forwarding | unproven | unproven |
| makeAsExpression at src/compiler/parser.ts:5689:5 | AsExpression | composed-or-forwarding | unproven | unproven |
| parsePrefixUnaryExpression at src/compiler/parser.ts:5693:5 | PrefixUnaryExpression | composed-or-forwarding | unproven | unproven |
| parseDeleteExpression at src/compiler/parser.ts:5698:5 | DeleteExpression | composed-or-forwarding | unproven | unproven |
| parseTypeOfExpression at src/compiler/parser.ts:5703:5 | TypeOfExpression | composed-or-forwarding | unproven | unproven |
| parseVoidExpression at src/compiler/parser.ts:5708:5 | VoidExpression | composed-or-forwarding | unproven | unproven |
| parseAwaitExpression at src/compiler/parser.ts:5726:5 | AwaitExpression | composed-or-forwarding | unproven | unproven |
| parseUnaryExpressionOrHigher at src/compiler/parser.ts:5738:5 | UnaryExpression &#124; BinaryExpression | composed-or-forwarding | unproven | unproven |
| parseSimpleUnaryExpression at src/compiler/parser.ts:5796:5 | UnaryExpression | composed-or-forwarding | unproven | unproven |
| parseUpdateExpression at src/compiler/parser.ts:5875:5 | UpdateExpression | composed-or-forwarding | unproven | unproven |
| parseLeftHandSideExpressionOrHigher at src/compiler/parser.ts:5897:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| parseMemberExpressionOrHigher at src/compiler/parser.ts:5970:5 | MemberExpression | composed-or-forwarding | unproven | unproven |
| parseSuperExpression at src/compiler/parser.ts:6023:5 | MemberExpression | composed-or-forwarding | unproven | unproven |
| parseJsxElementOrSelfClosingElementOrFragment at src/compiler/parser.ts:6048:5 | JsxElement &#124; JsxSelfClosingElement &#124; JsxFragment | composed-or-forwarding | unproven | unproven |
| <anonymous@6114:45> at src/compiler/parser.ts:6114:45 | JsxElement &#124; JsxSelfClosingElement &#124; JsxFragment | composed-or-forwarding | unproven | unproven |
| parseJsxText at src/compiler/parser.ts:6126:5 | JsxText | composed-or-forwarding | unproven | unproven |
| parseJsxChild at src/compiler/parser.ts:6133:5 | JsxChild &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJsxChildren at src/compiler/parser.ts:6164:5 | NodeArray<JsxChild> | non-node-builder | unproven | unproven |
| parseJsxAttributes at src/compiler/parser.ts:6189:5 | JsxAttributes | composed-or-forwarding | unproven | unproven |
| parseJsxOpeningOrSelfClosingElementOrOpeningFragment at src/compiler/parser.ts:6194:5 | JsxOpeningElement &#124; JsxSelfClosingElement &#124; JsxOpeningFragment | composed-or-forwarding | unproven | unproven |
| parseJsxElementName at src/compiler/parser.ts:6234:5 | JsxTagNameExpression | composed-or-forwarding | unproven | unproven |
| parseJsxTagName at src/compiler/parser.ts:6252:5 | Identifier &#124; ThisExpression &#124; JsxNamespacedName | composed-or-forwarding | unproven | unproven |
| parseJsxExpression at src/compiler/parser.ts:6265:5 | JsxExpression &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJsxAttribute at src/compiler/parser.ts:6294:5 | JsxAttribute &#124; JsxSpreadAttribute | composed-or-forwarding | unproven | unproven |
| parseJsxAttributeValue at src/compiler/parser.ts:6303:5 | JsxAttributeValue &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJsxAttributeName at src/compiler/parser.ts:6319:5 | Identifier &#124; JsxNamespacedName | composed-or-forwarding | unproven | unproven |
| parseJsxSpreadAttribute at src/compiler/parser.ts:6331:5 | JsxSpreadAttribute | composed-or-forwarding | unproven | unproven |
| parseJsxClosingElement at src/compiler/parser.ts:6340:5 | JsxClosingElement | composed-or-forwarding | unproven | unproven |
| parseJsxClosingFragment at src/compiler/parser.ts:6356:5 | JsxClosingFragment | composed-or-forwarding | unproven | unproven |
| parseTypeAssertion at src/compiler/parser.ts:6371:5 | TypeAssertion | composed-or-forwarding | unproven | unproven |
| tryReparseOptionalChain at src/compiler/parser.ts:6393:5 | boolean | non-node-builder | unproven | unproven |
| parsePropertyAccessExpressionRest at src/compiler/parser.ts:6415:5 | PropertyAccessExpression | composed-or-forwarding | unproven | unproven |
| parseElementAccessExpressionRest at src/compiler/parser.ts:6432:5 | ElementAccessExpression | composed-or-forwarding | unproven | unproven |
| parseMemberExpressionRest at src/compiler/parser.ts:6453:5 | MemberExpression | composed-or-forwarding | unproven | unproven |
| parseTaggedTemplateRest at src/compiler/parser.ts:6505:5 | TaggedTemplateExpression | composed-or-forwarding | unproven | unproven |
| parseCallExpressionRest at src/compiler/parser.ts:6520:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| parsePrimaryExpression at src/compiler/parser.ts:6608:5 | PrimaryExpression | composed-or-forwarding | unproven | unproven |
| parseParenthesizedExpression at src/compiler/parser.ts:6663:5 | ParenthesizedExpression | composed-or-forwarding | unproven | unproven |
| parseSpreadElement at src/compiler/parser.ts:6672:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseArgumentOrArrayLiteralElement at src/compiler/parser.ts:6679:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseArgumentExpression at src/compiler/parser.ts:6685:5 | Expression | input-or-unresolved | unproven | unproven |
| parseArrayLiteralExpression at src/compiler/parser.ts:6689:5 | ArrayLiteralExpression | composed-or-forwarding | unproven | unproven |
| parseObjectLiteralElement at src/compiler/parser.ts:6699:5 | ObjectLiteralElementLike | composed-or-forwarding | unproven | unproven |
| <anonymous@6737:74> at src/compiler/parser.ts:6737:74 | Expression | composed-or-forwarding | unproven | unproven |
| <anonymous@6745:44> at src/compiler/parser.ts:6745:44 | Expression | composed-or-forwarding | unproven | unproven |
| parseObjectLiteralExpression at src/compiler/parser.ts:6755:5 | ObjectLiteralExpression | composed-or-forwarding | unproven | unproven |
| parseFunctionExpression at src/compiler/parser.ts:6765:5 | FunctionExpression | composed-or-forwarding | unproven | unproven |
| parseOptionalBindingIdentifier at src/compiler/parser.ts:6797:5 | Identifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseNewExpressionOrNewDotTarget at src/compiler/parser.ts:6801:5 | NewExpression &#124; MetaProperty | composed-or-forwarding | unproven | unproven |
| parseBlock at src/compiler/parser.ts:6824:5 | Block | composed-or-forwarding | unproven | unproven |
| parseFunctionBlock at src/compiler/parser.ts:6847:5 | Block | composed-or-forwarding | unproven | unproven |
| parseEmptyStatement at src/compiler/parser.ts:6877:5 | Statement | composed-or-forwarding | unproven | unproven |
| parseIfStatement at src/compiler/parser.ts:6884:5 | IfStatement | composed-or-forwarding | unproven | unproven |
| parseDoStatement at src/compiler/parser.ts:6897:5 | DoStatement | composed-or-forwarding | unproven | unproven |
| parseWhileStatement at src/compiler/parser.ts:6916:5 | WhileStatement | composed-or-forwarding | unproven | unproven |
| parseForOrForInOrForOfStatement at src/compiler/parser.ts:6928:5 | Statement | composed-or-forwarding | unproven | unproven |
| <anonymous@6952:43> at src/compiler/parser.ts:6952:43 | Expression | composed-or-forwarding | unproven | unproven |
| parseBreakOrContinueStatement at src/compiler/parser.ts:6977:5 | BreakOrContinueStatement | composed-or-forwarding | unproven | unproven |
| parseReturnStatement at src/compiler/parser.ts:6991:5 | ReturnStatement | composed-or-forwarding | unproven | unproven |
| parseWithStatement at src/compiler/parser.ts:7000:5 | WithStatement | composed-or-forwarding | unproven | unproven |
| parseCaseClause at src/compiler/parser.ts:7012:5 | CaseClause | composed-or-forwarding | unproven | unproven |
| parseDefaultClause at src/compiler/parser.ts:7022:5 | DefaultClause | composed-or-forwarding | unproven | unproven |
| parseCaseOrDefaultClause at src/compiler/parser.ts:7030:5 | CaseOrDefaultClause | composed-or-forwarding | unproven | unproven |
| parseCaseBlock at src/compiler/parser.ts:7034:5 | CaseBlock | composed-or-forwarding | unproven | unproven |
| parseSwitchStatement at src/compiler/parser.ts:7042:5 | SwitchStatement | composed-or-forwarding | unproven | unproven |
| parseThrowStatement at src/compiler/parser.ts:7053:5 | ThrowStatement | composed-or-forwarding | unproven | unproven |
| parseTryStatement at src/compiler/parser.ts:7078:5 | TryStatement | composed-or-forwarding | unproven | unproven |
| parseCatchClause at src/compiler/parser.ts:7097:5 | CatchClause | composed-or-forwarding | unproven | unproven |
| parseDebuggerStatement at src/compiler/parser.ts:7115:5 | Statement | composed-or-forwarding | unproven | unproven |
| parseExpressionOrLabeledStatement at src/compiler/parser.ts:7123:5 | ExpressionStatement &#124; LabeledStatement | composed-or-forwarding | unproven | unproven |
| parseStatement at src/compiler/parser.ts:7380:5 | Statement | composed-or-forwarding | unproven | unproven |
| isDeclareModifier at src/compiler/parser.ts:7463:5 | boolean | non-node-builder | unproven | unproven |
| parseDeclaration at src/compiler/parser.ts:7467:5 | Statement | composed-or-forwarding | unproven | unproven |
| <anonymous@7484:57> at src/compiler/parser.ts:7484:57 | Statement | composed-or-forwarding | unproven | unproven |
| tryReuseAmbientDeclaration at src/compiler/parser.ts:7491:5 | Statement &#124; undefined | input-or-unresolved | unproven | unproven |
| <anonymous@7492:53> at src/compiler/parser.ts:7492:53 | Statement &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseDeclarationWorker at src/compiler/parser.ts:7502:5 | Statement | composed-or-forwarding | unproven | unproven |
| parseFunctionBlockOrSemicolon at src/compiler/parser.ts:7567:5 | Block &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseArrayBindingElement at src/compiler/parser.ts:7583:5 | ArrayBindingElement | composed-or-forwarding | unproven | unproven |
| parseObjectBindingElement at src/compiler/parser.ts:7594:5 | BindingElement | composed-or-forwarding | unproven | unproven |
| parseObjectBindingPattern at src/compiler/parser.ts:7612:5 | ObjectBindingPattern | composed-or-forwarding | unproven | unproven |
| parseArrayBindingPattern at src/compiler/parser.ts:7620:5 | ArrayBindingPattern | composed-or-forwarding | unproven | unproven |
| parseIdentifierOrPattern at src/compiler/parser.ts:7635:5 | Identifier &#124; BindingPattern | composed-or-forwarding | unproven | unproven |
| parseVariableDeclarationAllowExclamation at src/compiler/parser.ts:7645:5 | VariableDeclaration | composed-or-forwarding | unproven | unproven |
| parseVariableDeclaration at src/compiler/parser.ts:7649:5 | VariableDeclaration | composed-or-forwarding | unproven | unproven |
| parseVariableDeclarationList at src/compiler/parser.ts:7666:5 | VariableDeclarationList | composed-or-forwarding | unproven | unproven |
| parseVariableStatement at src/compiler/parser.ts:7727:5 | VariableStatement | composed-or-forwarding | unproven | unproven |
| parseFunctionDeclaration at src/compiler/parser.ts:7734:5 | FunctionDeclaration | composed-or-forwarding | unproven | unproven |
| parseConstructorName at src/compiler/parser.ts:7753:5 | boolean &#124; LiteralExpression &#124; undefined | input-or-unresolved | unproven | unproven |
| <anonymous@7758:29> at src/compiler/parser.ts:7758:29 | LiteralExpression &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryParseConstructorDeclaration at src/compiler/parser.ts:7765:5 | ConstructorDeclaration &#124; undefined | input-or-unresolved | unproven | unproven |
| <anonymous@7766:25> at src/compiler/parser.ts:7766:25 | ConstructorDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseMethodDeclaration at src/compiler/parser.ts:7782:5 | MethodDeclaration | composed-or-forwarding | unproven | unproven |
| parsePropertyDeclaration at src/compiler/parser.ts:7814:5 | PropertyDeclaration | composed-or-forwarding | unproven | unproven |
| parsePropertyOrMethodDeclaration at src/compiler/parser.ts:7835:5 | PropertyDeclaration &#124; MethodDeclaration | composed-or-forwarding | unproven | unproven |
| parseAccessorDeclaration at src/compiler/parser.ts:7851:5 | AccessorDeclaration | composed-or-forwarding | unproven | unproven |
| parseClassStaticBlockDeclaration at src/compiler/parser.ts:7935:5 | ClassStaticBlockDeclaration | composed-or-forwarding | unproven | unproven |
| parseClassStaticBlockBody at src/compiler/parser.ts:7943:5 | Block | composed-or-forwarding | unproven | unproven |
| parseDecoratorExpression at src/compiler/parser.ts:7958:5 | LeftHandSideExpression | composed-or-forwarding | unproven | unproven |
| tryParseDecorator at src/compiler/parser.ts:7971:5 | Decorator &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryParseModifier at src/compiler/parser.ts:7980:5 | Modifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseClassElement at src/compiler/parser.ts:8068:5 | ClassElement | composed-or-forwarding | unproven | unproven |
| <anonymous@8115:61> at src/compiler/parser.ts:8115:61 | PropertyDeclaration &#124; MethodDeclaration | composed-or-forwarding | unproven | unproven |
| parseDecoratedExpression at src/compiler/parser.ts:8132:5 | PrimaryExpression | composed-or-forwarding | unproven | unproven |
| parseClassExpression at src/compiler/parser.ts:8146:5 | ClassExpression | composed-or-forwarding | unproven | unproven |
| parseClassDeclaration at src/compiler/parser.ts:8150:5 | ClassDeclaration | composed-or-forwarding | unproven | unproven |
| parseClassDeclarationOrExpression at src/compiler/parser.ts:8154:5 | ClassLikeDeclaration | composed-or-forwarding | unproven | unproven |
| parseNameOfClassDeclarationOrExpression at src/compiler/parser.ts:8181:5 | Identifier &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseHeritageClause at src/compiler/parser.ts:8207:5 | HeritageClause | composed-or-forwarding | unproven | unproven |
| parseExpressionWithTypeArguments at src/compiler/parser.ts:8216:5 | ExpressionWithTypeArguments | composed-or-forwarding | unproven | unproven |
| parseInterfaceDeclaration at src/compiler/parser.ts:8239:5 | InterfaceDeclaration | composed-or-forwarding | unproven | unproven |
| parseTypeAliasDeclaration at src/compiler/parser.ts:8249:5 | TypeAliasDeclaration | composed-or-forwarding | unproven | unproven |
| parseEnumMember at src/compiler/parser.ts:8267:5 | EnumMember | composed-or-forwarding | unproven | unproven |
| parseEnumDeclaration at src/compiler/parser.ts:8275:5 | EnumDeclaration | composed-or-forwarding | unproven | unproven |
| parseModuleBlock at src/compiler/parser.ts:8290:5 | ModuleBlock | composed-or-forwarding | unproven | unproven |
| parseModuleOrNamespaceDeclaration at src/compiler/parser.ts:8303:5 | ModuleDeclaration | composed-or-forwarding | unproven | unproven |
| parseAmbientExternalModuleDeclaration at src/compiler/parser.ts:8315:5 | ModuleDeclaration | composed-or-forwarding | unproven | unproven |
| parseModuleDeclaration at src/compiler/parser.ts:8338:5 | ModuleDeclaration | composed-or-forwarding | unproven | unproven |
| parseNamespaceExportDeclaration at src/compiler/parser.ts:8373:5 | NamespaceExportDeclaration | composed-or-forwarding | unproven | unproven |
| parseImportDeclarationOrImportEqualsDeclaration at src/compiler/parser.ts:8384:5 | ImportEqualsDeclaration &#124; ImportDeclaration | composed-or-forwarding | unproven | unproven |
| tryParseImportClause at src/compiler/parser.ts:8425:5 | ImportClause &#124; undefined | composed-or-forwarding | unproven | unproven |
| tryParseImportAttributes at src/compiler/parser.ts:8441:5 | ImportAttributes &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseImportAttribute at src/compiler/parser.ts:8448:5 | ImportAttribute | composed-or-forwarding | unproven | unproven |
| parseImportAttributes at src/compiler/parser.ts:8456:5 | ImportAttributes | composed-or-forwarding | unproven | unproven |
| parseImportEqualsDeclaration at src/compiler/parser.ts:8492:5 | ImportEqualsDeclaration | composed-or-forwarding | unproven | unproven |
| parseImportClause at src/compiler/parser.ts:8501:5 | ImportClause | composed-or-forwarding | unproven | unproven |
| parseModuleReference at src/compiler/parser.ts:8529:5 | EntityName &#124; ExternalModuleReference | composed-or-forwarding | unproven | unproven |
| parseExternalModuleReference at src/compiler/parser.ts:8535:5 | ExternalModuleReference | composed-or-forwarding | unproven | unproven |
| parseModuleSpecifier at src/compiler/parser.ts:8544:5 | Expression | composed-or-forwarding | unproven | unproven |
| parseNamespaceImport at src/compiler/parser.ts:8558:5 | NamespaceImport | composed-or-forwarding | unproven | unproven |
| parseModuleExportName at src/compiler/parser.ts:8572:5 | ModuleExportName | composed-or-forwarding | unproven | unproven |
| parseNamedImportsOrExports at src/compiler/parser.ts:8578:5 | NamedImportsOrExports | composed-or-forwarding | unproven | unproven |
| parseExportSpecifier at src/compiler/parser.ts:8595:5 | ExportSpecifier | composed-or-forwarding | unproven | unproven |
| parseImportSpecifier at src/compiler/parser.ts:8600:5 | ImportSpecifier | composed-or-forwarding | unproven | unproven |
| parseImportOrExportSpecifier at src/compiler/parser.ts:8604:5 | ImportOrExportSpecifier | composed-or-forwarding | unproven | unproven |
| parseNameWithKeywordCheck at src/compiler/parser.ts:8689:9 | Identifier | composed-or-forwarding | unproven | unproven |
| parseNamespaceExport at src/compiler/parser.ts:8697:5 | NamespaceExport | composed-or-forwarding | unproven | unproven |
| parseExportDeclaration at src/compiler/parser.ts:8701:5 | ExportDeclaration | composed-or-forwarding | unproven | unproven |
| parseExportAssignment at src/compiler/parser.ts:8736:5 | ExportAssignment | composed-or-forwarding | unproven | unproven |
| parseJSDocTypeExpression at src/compiler/parser.ts:8809:9 | JSDocTypeExpression | composed-or-forwarding | unproven | unproven |
| parseJSDocNameReference at src/compiler/parser.ts:8822:9 | JSDocNameReference | composed-or-forwarding | unproven | unproven |
| <anonymous@8843:62> at src/compiler/parser.ts:8843:62 | JSDoc &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJSDocComment at src/compiler/parser.ts:8852:9 | JSDoc &#124; undefined | input-or-unresolved | unproven | unproven |
| <anonymous@8857:64> at src/compiler/parser.ts:8857:64 | JSDoc &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJSDocCommentWorker at src/compiler/parser.ts:8885:9 | JSDoc &#124; undefined | input-or-unresolved | unproven | unproven |
| doJSDocScan at src/compiler/parser.ts:8915:13 | JSDoc | composed-or-forwarding | unproven | unproven |
| parseTag at src/compiler/parser.ts:9095:13 | JSDocTag | composed-or-forwarding | unproven | unproven |
| parseJSDocLink at src/compiler/parser.ts:9309:13 | JSDocLink &#124; JSDocLinkCode &#124; JSDocLinkPlain &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseJSDocLinkName at src/compiler/parser.ts:9328:13 | EntityName &#124; JSDocMemberName &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseUnknownTag at src/compiler/parser.ts:9362:13 | JSDocUnknownTag | composed-or-forwarding | unproven | unproven |
| addTag at src/compiler/parser.ts:9366:13 | void | non-node-builder | unproven | unproven |
| tryParseTypeExpression at src/compiler/parser.ts:9380:13 | JSDocTypeExpression &#124; undefined | composed-or-forwarding | unproven | unproven |
| isObjectOrObjectArrayTypeReference at src/compiler/parser.ts:9410:13 | boolean | non-node-builder | unproven | unproven |
| parseParameterOrPropertyTag at src/compiler/parser.ts:9421:13 | JSDocParameterTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| parseNestedTypeLiteral at src/compiler/parser.ts:9446:13 | JSDocTypeExpression &#124; undefined | composed-or-forwarding | unproven | unproven |
| <anonymous@9451:45> at src/compiler/parser.ts:9451:45 | false &#124; JSDocTemplateTag &#124; JSDocTypeTag &#124; JSDocThisTag &#124; JSDocParameterTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| parseReturnTag at src/compiler/parser.ts:9466:13 | JSDocReturnTag | composed-or-forwarding | unproven | unproven |
| parseTypeTag at src/compiler/parser.ts:9475:13 | JSDocTypeTag | composed-or-forwarding | unproven | unproven |
| parseSeeTag at src/compiler/parser.ts:9485:13 | JSDocSeeTag | composed-or-forwarding | unproven | unproven |
| parseThrowsTag at src/compiler/parser.ts:9493:13 | JSDocThrowsTag | composed-or-forwarding | unproven | unproven |
| parseAuthorTag at src/compiler/parser.ts:9499:13 | JSDocAuthorTag | composed-or-forwarding | unproven | unproven |
| parseAuthorNameAndEmail at src/compiler/parser.ts:9513:13 | JSDocText | composed-or-forwarding | unproven | unproven |
| parseImplementsTag at src/compiler/parser.ts:9536:13 | JSDocImplementsTag | composed-or-forwarding | unproven | unproven |
| parseAugmentsTag at src/compiler/parser.ts:9541:13 | JSDocAugmentsTag | composed-or-forwarding | unproven | unproven |
| parseSatisfiesTag at src/compiler/parser.ts:9546:13 | JSDocSatisfiesTag | composed-or-forwarding | unproven | unproven |
| parseImportTag at src/compiler/parser.ts:9552:13 | JSDocImportTag | composed-or-forwarding | unproven | unproven |
| parseExpressionWithTypeArgumentsForAugments at src/compiler/parser.ts:9568:13 | ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | composed-or-forwarding | unproven | unproven |
| parsePropertyAccessEntityNameExpression at src/compiler/parser.ts:9584:13 | Identifier &#124; PropertyAccessEntityNameExpression | composed-or-forwarding | unproven | unproven |
| parseSimpleTag at src/compiler/parser.ts:9594:13 | JSDocTag | composed-or-forwarding | unproven | unproven |
| parseThisTag at src/compiler/parser.ts:9598:13 | JSDocThisTag | composed-or-forwarding | unproven | unproven |
| parseEnumTag at src/compiler/parser.ts:9604:13 | JSDocEnumTag | composed-or-forwarding | unproven | unproven |
| parseTypedefTag at src/compiler/parser.ts:9610:13 | JSDocTypedefTag | composed-or-forwarding | unproven | unproven |
| <anonymous@9624:45> at src/compiler/parser.ts:9624:45 | false &#124; JSDocTemplateTag &#124; JSDocTypeTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| parseJSDocTypeNameWithNamespace at src/compiler/parser.ts:9668:13 | Identifier &#124; JSDocNamespaceDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| <anonymous@9695:41> at src/compiler/parser.ts:9695:41 | JSDocTemplateTag &#124; JSDocParameterTag | composed-or-forwarding | unproven | unproven |
| parseJSDocSignature at src/compiler/parser.ts:9705:13 | JSDocSignature | composed-or-forwarding | unproven | unproven |
| <anonymous@9707:44> at src/compiler/parser.ts:9707:44 | JSDocReturnTag &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseCallbackTag at src/compiler/parser.ts:9718:13 | JSDocCallbackTag | composed-or-forwarding | unproven | unproven |
| parseOverloadTag at src/compiler/parser.ts:9730:13 | JSDocOverloadTag | composed-or-forwarding | unproven | unproven |
| escapedTextsEqual at src/compiler/parser.ts:9741:13 | boolean | non-node-builder | unproven | unproven |
| parseChildPropertyTag at src/compiler/parser.ts:9754:13 | false &#124; JSDocTemplateTag &#124; JSDocTypeTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| parseChildParameterOrPropertyTag at src/compiler/parser.ts:9758:13 | false &#124; JSDocTemplateTag &#124; JSDocTypeTag &#124; JSDocThisTag &#124; JSDocParameterTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| tryParseChildTag at src/compiler/parser.ts:9795:13 | false &#124; JSDocTemplateTag &#124; JSDocTypeTag &#124; JSDocThisTag &#124; JSDocParameterTag &#124; JSDocPropertyTag | composed-or-forwarding | unproven | unproven |
| parseTemplateTagTypeParameter at src/compiler/parser.ts:9828:13 | TypeParameterDeclaration &#124; undefined | composed-or-forwarding | unproven | unproven |
| parseTemplateTag at src/compiler/parser.ts:9866:13 | JSDocTemplateTag | composed-or-forwarding | unproven | unproven |
| parseJSDocEntityName at src/compiler/parser.ts:9891:13 | EntityName | composed-or-forwarding | unproven | unproven |
| parseJSDocIdentifierName at src/compiler/parser.ts:9909:13 | Identifier | composed-or-forwarding | unproven | unproven |
| markAsIncrementallyParsed at src/compiler/parser.ts:9929:1 | void | non-node-builder | unproven | unproven |
| intersectsIncrementalChange at src/compiler/parser.ts:9938:1 | boolean | non-node-builder | unproven | unproven |
| markAsIntersectingIncrementalChange at src/compiler/parser.ts:9942:1 | void | non-node-builder | unproven | unproven |
| updateSourceFile at src/compiler/parser.ts:9947:5 | SourceFile | composed-or-forwarding | unproven | unproven |
| moveElementEntirelyPastChangeRange at src/compiler/parser.ts:10087:5 | void | non-node-builder | unproven | unproven |
| visitNode at src/compiler/parser.ts:10096:9 | void | non-node-builder | unproven | unproven |
| shouldCheckNode at src/compiler/parser.ts:10130:5 | boolean | non-node-builder | unproven | unproven |
| adjustIntersectingElement at src/compiler/parser.ts:10141:5 | void | non-node-builder | unproven | unproven |
| checkNodePositions at src/compiler/parser.ts:10216:5 | void | non-node-builder | unproven | unproven |
| visitNode at src/compiler/parser.ts:10219:31 | void | non-node-builder | unproven | unproven |
| updateTokenPositionsAndMarkElements at src/compiler/parser.ts:10233:5 | void | non-node-builder | unproven | unproven |
| visitNode at src/compiler/parser.ts:10246:9 | void | non-node-builder | unproven | unproven |
| extendToAffectedRange at src/compiler/parser.ts:10308:5 | TextChangeRange | non-node-builder | unproven | unproven |
| findNearestNodeStartingBeforeOrAtPosition at src/compiler/parser.ts:10340:5 | Node | composed-or-forwarding | unproven | unproven |
| getLastDescendant at src/compiler/parser.ts:10355:9 | Node | input-or-unresolved | unproven | unproven |
| visit at src/compiler/parser.ts:10367:9 | true &#124; undefined | non-node-builder | unproven | unproven |
| checkChangeRange at src/compiler/parser.ts:10425:5 | void | non-node-builder | unproven | unproven |
| createSyntaxCursor at src/compiler/parser.ts:10449:5 | SyntaxCursor | non-node-builder | unproven | unproven |
| currentNode at src/compiler/parser.ts:10458:13 | Node | input-or-unresolved | unproven | unproven |
| visitNode at src/compiler/parser.ts:10505:13 | boolean | non-node-builder | unproven | unproven |
| extractPragmas at src/compiler/parser.ts:10714:1 | void | non-node-builder | unproven | unproven |
| addPragmaForMatch at src/compiler/parser.ts:10767:1 | void | non-node-builder | unproven | unproven |
| tagNamesAreEquivalent at src/compiler/parser.ts:10800:1 | boolean | non-node-builder | unproven | unproven |
