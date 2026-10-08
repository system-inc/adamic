package load

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// RequiresIndexedSite identifies a use rejected only after an indexed read gained
// undefined. A joint exact-optional/index diagnostic can occur on the containing
// object literal or variable declaration, whose permissive contextual type would
// otherwise hide the required read guard. Guarding the read removes that absence.
func (p *Program) RequiresIndexedSite(node *ast.Node) bool {
	return p.requiresOptionSite(node, "noUncheckedIndexedAccess")
}

func (p *Program) requiresOptionSite(node *ast.Node, name string) bool {
	for parent := node; parent != nil && parent.Kind != ast.KindSourceFile; parent = parent.Parent {
		where := p.Where(parent)
		for _, site := range p.optionSites {
			if where != fmt.Sprintf("%s:%d:%d", site.File, site.Line, site.Column) {
				continue
			}
			for _, option := range site.Options {
				if option == name {
					return true
				}
			}
		}
	}
	return false
}
