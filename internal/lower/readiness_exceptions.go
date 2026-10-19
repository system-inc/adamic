package lower

import "github.com/system-inc/adamic/internal/ir"

// Readiness can add dead-zone checks after the initial exception census.
// Propagate their effects before borrowing and native ownership analysis.
func readinessExceptions(program *ir.Program) {
	l := lowering{result: program}
	dispatched := map[int]bool{}
	for _, class := range program.Classes {
		for _, method := range class.Methods {
			dispatched[method] = true
		}
		for _, accessor := range class.Accessors {
			if accessor.Getter >= 0 {
				dispatched[accessor.Getter] = true
			}
		}
	}
	for index, function := range program.Functions {
		if function.Receiver || function.MethodName != "" {
			dispatched[index] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for index := range program.Functions {
			function := &program.Functions[index]
			if !function.MayThrow && l.throwsOutReadiness(function.Body, true) {
				function.MayThrow = true
				changed = true
			}
			if (function.Closure || dispatched[index]) && function.MayThrow && !program.ClosuresMayThrow {
				program.ClosuresMayThrow = true
				changed = true
			}
		}
	}
}
