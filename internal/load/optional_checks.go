package load

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"sort"
)

// acceptOptionalWrite admits only direct property assignments. Lowering must
// still select a presence-capable storage representation and emit the guard.
func (p *Program) acceptOptionalWrite(site OptionSite) bool {
	if site.Code != 2412 || len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" {
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		var found *ast.Node
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindBinaryExpression {
				binary := node.AsBinaryExpression()
				target := ast.SkipParentheses(binary.Left)
				if binary.OperatorToken.Kind == ast.KindEqualsToken && target.Kind == ast.KindPropertyAccessExpression && site.Position >= target.Pos() && site.Position < target.End() {
					found = target
					return true
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			if p.optionalChecks == nil {
				p.optionalChecks = map[string]OptionSite{}
				p.checkedOptions = map[string]bool{}
			}
			p.optionalChecks[p.Where(found)] = site
			return true
		}
	}
	return false
}

// OptionalWriteSite is a pending check contract, not an emitted check.
func (p *Program) OptionalWriteSite(target *ast.Node) (OptionSite, bool) {
	site, ok := p.optionalChecks[p.Where(target)]
	return site, ok
}

// RecordOptionalWrite is called only after the guarded store has been lowered.
func (p *Program) RecordOptionalWrite(target *ast.Node) { p.checkedOptions[p.Where(target)] = true }

// ExplainedOptionalChecks reports only guards actually emitted by lowering.
func (p *Program) ExplainedOptionalChecks() []string {
	result := []string{}
	for where := range p.checkedOptions {
		if view, ok := p.optionalViews[where]; ok {
			result = append(result, fmt.Sprintf("%s: checked TS%d exactOptionalPropertyTypes: required fields preserve own presence through optional view", where, view.Site.Code))
			continue
		}
		if literals, ok := p.optionalLiterals[where]; ok {
			for _, literal := range literals {
				result = append(result, fmt.Sprintf("%s: checked TS%d exactOptionalPropertyTypes: object construction preserves own presence", where, literal.Code))
			}
			continue
		}
		if relation, ok := p.optionalRelations[where]; ok {
			result = append(result, fmt.Sprintf("%s: checked TS%d exactOptionalPropertyTypes: implements relation uses own-presence representation", where, relation.Code))
			continue
		}
		site := p.optionalChecks[where]
		result = append(result, fmt.Sprintf("%s: checked TS%d exactOptionalPropertyTypes: optional write preserves own presence", where, site.Code))
	}
	sort.Strings(result)
	return result
}

// Exact-only implements diagnostics describe an optional-property relation
// already accepted by the owning project. The object's representation, rather
// than a value prohibition, supplies the required distinction from absence.
func (p *Program) acceptOptionalRelation(site OptionSite) bool {
	if site.Code != 2420 || len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" {
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		var found *ast.Node
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindClassDeclaration && node.Name() != nil && site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
				found = node
				return true
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			if p.optionalRelations == nil {
				p.optionalRelations = map[string]OptionSite{}
			}
			if p.checkedOptions == nil {
				p.checkedOptions = map[string]bool{}
			}
			p.optionalRelations[p.Where(found)] = site
			return true
		}
	}
	return false
}

// RecordOptionalRelation records a representation check only after the class
// instance and its storage have successfully lowered.
func (p *Program) RecordOptionalRelation(class *ast.Node) {
	where := p.Where(class)
	if _, ok := p.optionalRelations[where]; ok {
		p.checkedOptions[where] = true
	}
}
