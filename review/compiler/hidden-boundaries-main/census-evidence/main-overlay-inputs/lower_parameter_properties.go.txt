package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func parameterProperty(node *ast.Node) bool {
	return node.Kind == ast.KindParameter && ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier)
}

// Node's transform declares parameter-property slots before the written fields. The parameters'
// defaults are not field initializers: stores run after defaults, and after super in a derived class.
func classMembersWithParameters(declaration *ast.Node) []*ast.Node {
	members := []*ast.Node{}
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindConstructor {
			for _, parameter := range member.Parameters() {
				if parameterProperty(parameter) {
					members = append(members, parameter)
				}
			}
		}
	}
	return append(members, declaration.Members()...)
}

func (l *lowering) parameterPropertyStores(declaration *ast.Node, this int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	for _, parameter := range classMembersWithParameters(declaration) {
		if !parameterProperty(parameter) {
			continue
		}
		local, found := l.locals[l.symbol(parameter.Name())]
		if !found {
			return nil, l.notYet(parameter, "a parameter property without a bound parameter")
		}
		of, err := l.typeOf(parameter.Name())
		if err != nil {
			return nil, err
		}
		statements = append(statements, ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: parameter.Name().Text(), Value: fit(ir.Read{Local: local, Of: l.result.Locals[local].Type}, of), Site: l.writeSite(declaration.Name())})
	}
	return statements, nil
}

// The checker permits a default to read a parameter property that has not been copied yet.
// Node reads undefined there. A declared number slot cannot claim that value, so require an
// already initialized written field instead. Calls and deferred captures are checked conservatively.
func (l *lowering) parameterPropertyDefault(declaration, initializer *ast.Node) error {
	hasProperties := false
	available := map[string]bool{}
	for _, member := range classMembersWithParameters(declaration.Parent) {
		hasProperties = hasProperties || parameterProperty(member)
		if member.Kind == ast.KindPropertyDeclaration && member.AsPropertyDeclaration().Initializer != nil && !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			available[l.fieldName(member.Name())] = true
		}
	}
	if !hasProperties {
		return nil
	}
	return l.initializerReads(initializer, available)
}

// A derived parameter-property declaration defines its own slot after super, resetting any
// inherited property of the same name before the written field initializers run.
func (l *lowering) parameterPropertyResets(declaration *ast.Node, this int, available map[string]bool) ([]ir.Statement, error) {
	if l.instance.base == nil {
		return nil, nil
	}
	statements := []ir.Statement{}
	for _, parameter := range classMembersWithParameters(declaration) {
		if !parameterProperty(parameter) {
			continue
		}
		delete(available, parameter.Name().Text())
		of, err := l.typeOf(parameter.Name())
		if err != nil {
			return nil, err
		}
		value := zeroValue(of)
		if of.IsReference() {
			value = ir.Undefined{Of: of}
		}
		statements = append(statements, ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: parameter.Name().Text(), Value: value, Site: l.writeSite(declaration.Name())})
	}
	return statements, nil
}
