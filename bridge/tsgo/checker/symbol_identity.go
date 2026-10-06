package checker

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// symbolID is a borrowed identity scoped to this program, never a pointer.
func (p *Program) symbolID(symbol *ast.Symbol) uint64 {
	if symbol == nil {
		return 0
	}
	if p.symbolIDs == nil {
		p.symbolIDs = make(map[*ast.Symbol]uint64)
	}
	if p.symbolIDs[symbol] == 0 {
		id := uint64(len(p.symbolIDs)) + 1
		if id > (1<<53)-1 {
			panic("symbol identity space exhausted")
		}
		p.symbolIDs[symbol] = id
		p.symbolsByID = append(p.symbolsByID, symbol)
	}
	return p.symbolIDs[symbol]
}
