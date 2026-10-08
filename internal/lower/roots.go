package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// rootOrder preserves each root's dependency order, evaluating shared modules once.
// moduleOrder remains the single-root walk owned by the import-cycle unit.
func (l *lowering) rootOrder(roots []*ast.SourceFile) ([]*ast.SourceFile, error) {
	var order []*ast.SourceFile
	seen := map[*ast.SourceFile]bool{}
	for _, root := range roots {
		modules, err := l.moduleOrder(root)
		if err != nil {
			return nil, err
		}
		for _, module := range modules {
			if !seen[module] {
				seen[module] = true
				order = append(order, module)
			}
		}
	}
	return order, nil
}
