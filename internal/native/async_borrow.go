package native

import "github.com/system-inc/adamic/internal/ir"

// A direct alias can borrow a stable slot in the same suspension frame. Both
// slots survive every await, and neither is writable by a closure. Repeated
// declarations in loops are writes too, even when the source binding is const.
func asyncBorrowedSlots(program *ir.Program, function int) map[int]bool {
	body := program.Functions[function].Body
	assigned := assignedLocals(body)
	declarations := map[int]int{}
	repeated := map[int]bool{}
	aliases := map[int]int{}
	var visit func([]ir.Statement, bool)
	visit = func(body []ir.Statement, loop bool) {
		for _, statement := range body {
			if declaration, ok := statement.(ir.Declare); ok {
				declarations[declaration.Local]++
				repeated[declaration.Local] = repeated[declaration.Local] || loop
				if read, ok := declaration.Value.(ir.Read); ok {
					aliases[declaration.Local] = read.Local
				}
			}
			_, isLoop := statement.(ir.Loop)
			walkStatement(statement, func(ir.Expression) {}, func(inner []ir.Statement) { visit(inner, loop || isLoop) })
		}
	}
	visit(body, false)
	stable := func(local int) bool {
		binding := program.Locals[local]
		return binding.Function == function && !binding.Global && !binding.Captured && !binding.IterationCell && !binding.ExpressionAssigned && !assigned[local] && !repeated[local] && declarations[local] <= 1 && binding.Type.IsReference()
	}
	borrowed := map[int]bool{}
	for local, source := range aliases {
		if stable(local) && stable(source) && program.Locals[local].Type == program.Locals[source].Type {
			borrowed[local] = true
		}
	}
	return borrowed
}
