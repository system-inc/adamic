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
	type source struct {
		local   int
		payload bool
	}
	aliases := map[int]source{}
	var visit func([]ir.Statement, bool)
	visit = func(body []ir.Statement, loop bool) {
		for _, statement := range body {
			if declaration, ok := statement.(ir.Declare); ok {
				declarations[declaration.Local]++
				repeated[declaration.Local] = repeated[declaration.Local] || loop
				switch value := declaration.Value.(type) {
				case ir.Read:
					aliases[declaration.Local] = source{local: value.Local}
				case ir.Await:
					if read, ok := value.Value.(ir.Read); ok {
						aliases[declaration.Local] = source{local: read.Local, payload: true}
					}
				}
			}
			_, isLoop := statement.(ir.Loop)
			walkStatement(statement, func(ir.Expression) {}, func(inner []ir.Statement) { visit(inner, loop || isLoop) })
		}
	}
	visit(body, false)
	stable := func(local int) bool {
		binding := program.Locals[local]
		return binding.Function == function && !binding.Global && !binding.Captured && !binding.ExpressionAssigned && !assigned[local] && !repeated[local] && declarations[local] <= 1 && binding.Type.IsReference()
	}
	borrowed := map[int]bool{}
	for local, source := range aliases {
		if stable(local) && stable(source.local) && (source.payload && program.Locals[source.local].Type == ir.Promise || !source.payload && program.Locals[local].Type == program.Locals[source.local].Type) {
			borrowed[local] = true
		}
	}
	return borrowed
}
