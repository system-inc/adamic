package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// utilities/is-flow-keyword-type.js, utilities/is-ts-keyword-type.js,
// utilities/is-simple-type-annotation.js, utilities/is-simple-type.js,
// utilities/is-type-annotation-a-function.js, utilities/is-as-const-expression.js.

// isFlowKeywordType is upstream's isFlowKeywordType.
func isFlowKeywordType(node Node) bool {
	return node.Is(
		"AnyTypeAnnotation",
		"ThisTypeAnnotation",
		"NumberTypeAnnotation",
		"VoidTypeAnnotation",
		"BooleanTypeAnnotation",
		"BigIntTypeAnnotation",
		"SymbolTypeAnnotation",
		"StringTypeAnnotation",
		"NeverTypeAnnotation",
		"UndefinedTypeAnnotation",
		"UnknownTypeAnnotation",
		// FLow only
		"EmptyTypeAnnotation",
		"MixedTypeAnnotation",
	)
}

// isTsKeywordType is upstream's isTsKeywordType.
func isTsKeywordType(node Node) bool {
	return strings.HasPrefix(node.Type(), "TS") && strings.HasSuffix(node.Type(), "Keyword")
}

// isSimpleTypeAnnotation is upstream's isSimpleTypeAnnotation.
func isSimpleTypeAnnotation(node Node) bool {
	return node.Is(
		"TSThisType",
		// literals
		"NullLiteralTypeAnnotation",
		"BooleanLiteralTypeAnnotation",
		"StringLiteralTypeAnnotation",
		"BigIntLiteralTypeAnnotation",
		"NumberLiteralTypeAnnotation",
		"TSLiteralType",
		"TSTemplateLiteralType",
	)
}

// isSimpleType is upstream's isSimpleType.
func isSimpleType(node Node) bool {
	return isTsKeywordType(node) ||
		isFlowKeywordType(node) ||
		isSimpleTypeAnnotation(node) ||
		node.Is("GenericTypeAnnotation") && !node.Truthy("typeParameters") ||
		node.Is("TSTypeReference") && !node.Truthy("typeArguments")
}

// isTypeAnnotationAFunction is upstream's isTypeAnnotationAFunction.
//
// Hack to differentiate between the following two which have the same AST
// declare function f(a): void;
// var f: (a) => void;
func isTypeAnnotationAFunction(node Node) bool {
	return node.Is("TypeAnnotation", "TSTypeAnnotation") &&
		node.Child("typeAnnotation").Is("FunctionTypeAnnotation") &&
		!node.Truthy("static") &&
		!estree.HasSameLocStart(node, node.Child("typeAnnotation"))
}

// isTsAsConstExpression is upstream's isTsAsConstExpression.
func isTsAsConstExpression(node Node) bool {
	return node.Is("TSAsExpression") &&
		node.Child("typeAnnotation").Is("TSTypeReference") &&
		node.Child("typeAnnotation").Child("typeName").Is("Identifier") &&
		node.Child("typeAnnotation").Child("typeName").String("name") == "const"
}
