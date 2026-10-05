package lower

import "github.com/system-inc/adamic/internal/ir"

// borrow marks which reference parameters are borrowed (docs/memory.md, "Borrowed parameters").
//
// The caller already keeps every argument alive to the end of its statement, and every place that
// keeps a reference past its statement takes a count of its own. So a parameter needs its own count
// only when something writes it, since a write releases the old value and that one is the caller's,
// or when it belongs to a closure: map's loop hands a callback its element unretained, and the
// callback may overwrite that element. Everything else is borrowed: a named function's or a
// method's reference parameters, this included, that no ir.Assign anywhere in the program writes.
//
// ir.Assign is the only statement that writes a local that already exists, and locals are numbered
// across the whole program, so a closure writing a parameter through its cell is found too.
func borrow(program *ir.Program) {
	assigned := map[int]bool{}
	for _, function := range program.Functions {
		findAssigned(function.Body, assigned)
	}
	findAssigned(program.Main, assigned)
	for _, function := range program.Functions {
		if function.Closure {
			continue
		}
		for _, parameter := range function.Parameters {
			local := &program.Locals[parameter]
			local.Borrowed = local.Type.IsReference() && !assigned[parameter]
		}
	}
}

// findAssigned adds every local an ir.Assign in statements writes, however deeply nested.
func findAssigned(statements []ir.Statement, assigned map[int]bool) {
	for _, statement := range statements {
		switch statement := statement.(type) {
		case ir.Assign:
			assigned[statement.Local] = true
		case ir.If:
			findAssigned(statement.Then, assigned)
			findAssigned(statement.Else, assigned)
		case ir.Loop:
			findAssigned(statement.Body, assigned)
			findAssigned(statement.Update, assigned)
		case ir.Block:
			findAssigned(statement.Body, assigned)
		case ir.ForOf:
			findAssigned(statement.Body, assigned)
		case ir.Switch:
			for _, switchCase := range statement.Cases {
				findAssigned(switchCase.Body, assigned)
			}
			findAssigned(statement.Default, assigned)
		}
	}
}
