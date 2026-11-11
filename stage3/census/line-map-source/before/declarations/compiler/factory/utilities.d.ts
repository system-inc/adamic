import { type BinaryOperatorToken, type ExclamationToken, type Identifier, type MinusToken, type ModuleName, type Node, type PlusToken, type QuestionToken, type ReadonlyKeyword, type ThisTypeNode } from "../_namespaces/ts.js";
export declare function isQuestionOrExclamationToken(node: Node): node is QuestionToken | ExclamationToken;
export declare function isIdentifierOrThisTypeNode(node: Node): node is Identifier | ThisTypeNode;
export declare function isReadonlyKeywordOrPlusOrMinusToken(node: Node): node is ReadonlyKeyword | PlusToken | MinusToken;
export declare function isQuestionOrPlusOrMinusToken(node: Node): node is QuestionToken | PlusToken | MinusToken;
export declare function isModuleName(node: Node): node is ModuleName;
export declare function isBinaryOperatorToken(node: Node): node is BinaryOperatorToken;
