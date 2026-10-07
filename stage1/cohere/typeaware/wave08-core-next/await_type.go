package wave08core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
)

// AwaitTypeFields returns raw type edges and call signatures at the supplied
// source location. Thenable and promise-contract decisions remain native.
func AwaitTypeFields(c *checker.Checker, node *ast.Node, t *checker.Type, property string, propertyPresent bool, id func(*checker.Type) uint64) []string {
	out := []string{"1", "wave08-await-type"}
	number := func(n int) { out = append(out, strconv.Itoa(n)) }
	typ := func(t *checker.Type) { out = append(out, strconv.FormatUint(id(t), 10)) }
	number(int(t.Flags()))
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		number(len(t.Types()))
		for _, part := range t.Types() {
			typ(part)
		}
	} else {
		number(0)
	}
	typ(checker.Checker_getApparentType(c, t))
	tuple := checker.Checker_isArrayOrTupleType(c, t) && !checker.Checker_isArrayType(c, t)
	if tuple {
		number(1)
	} else {
		number(0)
	}
	if tuple {
		arguments := checker.Checker_getTypeArguments(c, t)
		number(len(arguments))
		for _, argument := range arguments {
			typ(argument)
		}
	} else {
		number(0)
	}
	typ(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c)))
	typ(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_stringType(c)))
	signatures := checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)
	number(len(signatures))
	for _, signature := range signatures {
		typ(checker.Checker_getReturnTypeOfSignature(c, signature))
		parameters := signature.Parameters()
		number(len(parameters))
		for _, parameter := range parameters {
			typ(c.GetTypeOfSymbolAtLocation(parameter, node))
			declaration := parameter.ValueDeclaration
			if declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				number(1)
			} else {
				number(0)
			}
		}
	}
	var thenType, propertyType *checker.Type
	if then := checker.Checker_getPropertyOfType(c, t, "then"); then != nil {
		thenType = c.GetTypeOfSymbolAtLocation(then, node)
	}
	if propertyPresent {
		if member := checker.Checker_getPropertyOfType(c, t, property); member != nil {
			propertyType = checker.Checker_getTypeOfSymbol(c, member)
		}
	}
	typ(thenType)
	typ(propertyType)
	return out
}
