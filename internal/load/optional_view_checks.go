package load

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Required source fields are present even when their values are undefined.
// Optional source fields need a different contract: legitimate absence is allowed.
type OptionalViewContract struct {
	Site   OptionSite
	Fields []string
}

func (p *Program) acceptOptionalView(site OptionSite) bool {
	if len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" || (site.Code != 2379 && site.Code != 2375 && site.Code != 2345) {
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		checked, release := p.Checker(context.Background(), file)
		defer release()
		var found *ast.Node
		var names []string
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindIdentifier && site.Position == scanner.GetTokenPosOfNode(node, file, false) {
				source := checked.GetTypeAtLocation(node)
				target := checked.GetContextualType(node, checker.ContextFlagsNone)
				if target != nil {
					target = checked.GetNonNullableType(target)
				}
				if source.Flags()&checker.TypeFlagsObject == 0 || target == nil || target.Flags()&checker.TypeFlagsObject == 0 {
					return false
				}
				for _, field := range checked.GetPropertiesOfType(target) {
					if field.Flags&ast.SymbolFlagsOptional == 0 {
						continue
					}
					own := checked.GetPropertyOfType(source, field.Name)
					if own == nil || own.Flags&ast.SymbolFlagsOptional != 0 || own.Flags&ast.SymbolFlagsProperty == 0 {
						return false
					}
					names = append(names, field.Name)
				}
				if len(names) > 0 {
					found = node
					return true
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
			p.optionalViews[p.Where(found)] = OptionalViewContract{Site: site, Fields: names}
			return true
		}
	}
	return false
}

func (p *Program) OptionalViewSite(node *ast.Node) (OptionalViewContract, bool) {
	contract, ok := p.optionalViews[p.Where(node)]
	return contract, ok
}
