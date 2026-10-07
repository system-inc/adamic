package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Function control-flow and return-type facts, without lint judgments.
func (p *Program) functionReturnFlow(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) error {
	if question != "function-return-flow" || !ast.IsFunctionLike(node) {
		return fmt.Errorf("function-return-flow requires a function")
	}
	options := p.Compiler.Options()
	out.yes(options.NoImplicitReturns.IsTrue())
	out.yes(options.GetStrictOptionValue(options.StrictNullChecks))
	flags := ast.GetFunctionFlags(node)
	body := node.Body()
	out.yes(body != nil && !ast.NodeIsMissing(body) && ast.IsBlock(body))
	var declared *checker.Type
	if node.Kind == ast.KindGetAccessor {
		declared = checker.Checker_getTypeOfAccessors(c, node.Symbol())
	} else {
		declared = checker.Checker_getReturnTypeFromAnnotation(c, node)
	}
	var unwrapped *checker.Type
	if declared != nil {
		unwrapped = checker.Checker_unwrapReturnType(c, declared, flags)
	}
	out.yes(unwrapped != nil)
	if unwrapped == nil {
		out.number(0)
		out.yes(false)
		out.yes(false)
	} else {
		out.number(uint64(unwrapped.Flags()))
		out.yes(checker.Checker_maybeTypeOfKind(c, unwrapped, checker.TypeFlagsVoid))
		out.yes(checker.Checker_isTypeAssignableTo(c, checker.Checker_undefinedType(c), unwrapped))
	}
	implicit := false
	if body != nil && ast.IsBlock(body) {
		implicit = checker.Checker_functionHasImplicitReturn(c, node)
	}
	out.yes(implicit)
	out.yes(node.Flags&ast.NodeFlagsHasExplicitReturn != 0)
	signature := checker.Checker_getSignatureFromDeclaration(c, node)
	inferred := checker.Checker_getReturnTypeOfSignature(c, signature)
	out.number(uint64(inferred.Flags()))
	result := checker.Checker_unwrapReturnType(c, inferred, flags)
	if result == nil {
		out.number(0)
		out.yes(false)
	} else {
		out.number(uint64(result.Flags()))
		out.yes(checker.Checker_maybeTypeOfKind(c, result, checker.TypeFlagsVoid))
	}
	errorNode := node.Type()
	if errorNode == nil {
		if data := node.FunctionLikeData(); data != nil {
			errorNode = data.FullSignature
		}
	}
	if errorNode == nil {
		errorNode = node
	}
	span := scanner.GetErrorRangeForNode(source, errorNode)
	out.number(uint64(span.Pos()))
	out.number(uint64(span.End()))
	return nil
}

func init() { additionalQuestions["function-return-flow"] = (*Program).functionReturnFlow }
