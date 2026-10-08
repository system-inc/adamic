package load

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

type OptionalObjectContract struct {
	Kind     string
	Element  *OptionalObjectContract
	Result   *checker.Type
	Source   *checker.Type
	Nullable bool
	Fields   []string
	Children []OptionalObjectChild
}
type OptionalObjectChild struct {
	Name     string
	Optional bool
	Contract *OptionalObjectContract
}

func optionalContractCandidate(node *ast.Node, site OptionSite, file *ast.SourceFile) *ast.Node {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		if node.Name() != nil && site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
			return node.AsVariableDeclaration().Initializer
		}
	case ast.KindReturnStatement:
		if site.Position == scanner.GetTokenPosOfNode(node, file, false) {
			return node.AsReturnStatement().Expression
		}
	case ast.KindPropertyAssignment:
		if site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
			return node.AsPropertyAssignment().Initializer
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if (binary.OperatorToken.Kind == ast.KindEqualsToken || binary.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) && site.Position >= binary.Left.Pos() && site.Position < binary.Left.End() {
			return binary.Right
		}
	}
	return nil
}

func (p *Program) acceptOptionalNested(site OptionSite) bool {
	if len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" || (site.Code != 2322 && site.Code != 2375) {
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		checked, release := p.Checker(context.Background(), file)
		defer release()
		var found *ast.Node
		var contract *OptionalObjectContract
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			candidate := optionalContractCandidate(node, site, file)
			if candidate != nil {
				candidate = ast.SkipParentheses(candidate)
				target := checked.GetContextualType(candidate, checker.ContextFlagsNone)
				if target != nil {
					contract = optionalObjectContract(checked, checked.GetTypeAtLocation(candidate), target, map[[2]*checker.Type]bool{}, 0)
					if contract != nil && (len(contract.Children) > 0 || contract.Kind != "") {
						found = candidate
						return true
					}
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			if p.optionalViews == nil {
				p.optionalViews = map[string]OptionalViewContract{}
			}
			if p.checkedOptions == nil {
				p.checkedOptions = map[string]bool{}
			}
			p.optionalViews[p.Where(found)] = OptionalViewContract{Site: site, Tree: contract}
			return true
		}
	}
	return false
}

func optionalObjectContract(checked *checker.Checker, source, target *checker.Type, seen map[[2]*checker.Type]bool, depth int) *OptionalObjectContract {
	if source == nil || target == nil || source == target || depth > 16 {
		return nil
	}
	original := source
	source = checked.GetNonNullableType(source)
	target = checked.GetNonNullableType(target)
	pair := [2]*checker.Type{source, target}
	if seen[pair] {
		return nil
	}
	seen[pair] = true
	if checked.IsArrayType(source) && checked.IsArrayType(target) {
		sources, targets := checked.GetTypeArguments(source), checked.GetTypeArguments(target)
		if len(sources) != 1 || len(targets) != 1 {
			return nil
		}
		element := optionalObjectContract(checked, sources[0], targets[0], seen, depth+1)
		if element == nil {
			return nil
		}
		return &OptionalObjectContract{Kind: "array", Source: original, Nullable: original != source, Element: element}
	}
	calls, targets := checked.GetSignaturesOfType(source, checker.SignatureKindCall), checked.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(calls) == 1 && len(targets) > 0 {
		result := checked.GetReturnTypeOfSignature(calls[0])
		if !optionalObjectReturn(checked, result) {
			return nil
		}
		for _, signature := range targets {
			if !optionalObjectReturn(checked, checked.GetReturnTypeOfSignature(signature)) {
				return nil
			}
		}
		return &OptionalObjectContract{Kind: "function", Source: original, Result: result}
	}
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || checked.IsArrayType(source) || checked.IsArrayType(target) {
		return nil
	}
	result := &OptionalObjectContract{Source: original, Nullable: original != source}
	for _, field := range checked.GetPropertiesOfType(target) {
		own := checked.GetPropertyOfType(source, field.Name)
		if own == nil || own.Flags&ast.SymbolFlagsProperty == 0 {
			continue
		}
		optional := own.Flags&ast.SymbolFlagsOptional != 0
		if field.Flags&ast.SymbolFlagsOptional != 0 && !optional {
			result.Fields = append(result.Fields, field.Name)
		}
		child := optionalObjectContract(checked, checked.GetTypeOfPropertyOfType(source, field.Name), checked.GetTypeOfPropertyOfType(target, field.Name), seen, depth+1)
		if child != nil {
			result.Children = append(result.Children, OptionalObjectChild{Name: field.Name, Optional: optional, Contract: child})
		}
	}
	if len(result.Fields) == 0 && len(result.Children) == 0 {
		return nil
	}
	return result
}

// A compiled object-return callback uses the same presence-capable object ABI.
// Primitive or opaque return representations need a different contract.
func optionalObjectReturn(checked *checker.Checker, result *checker.Type) bool {
	result = checked.GetNonNullableType(result)
	if result.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range result.Types() {
			if !optionalObjectReturn(checked, member) {
				return false
			}
		}
		return true
	}
	return result.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 && !checked.IsArrayType(result)
}
