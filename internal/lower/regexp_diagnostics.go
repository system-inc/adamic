package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Parser diagnostics are not V8 diagnostic text. Prove that their errors cannot
// leave main; caught-value/message observations are separately refused earlier.
// Track only parser-originated errors, so a validating assertion may still throw
// its own matching diagnostic after handling the parser error.
func (l *lowering) regexpDiagnostics() error {
	invalid := false
	walk(l.result.Main, func(node any) bool {
		if value, ok := node.(ir.RegExpNew); ok {
			invalid = invalid || value.Invalid
		}
		return !invalid
	})
	for _, function := range l.result.Functions {
		if invalid {
			break
		}
		walk(function.Body, func(node any) bool {
			if value, ok := node.(ir.RegExpNew); ok {
				invalid = invalid || value.Invalid
			}
			return !invalid
		})
	}
	if !invalid {
		return nil
	}
	dispatched := map[int]bool{}
	for _, instance := range l.instances {
		for _, method := range instance.methodList() {
			dispatched[method.Function] = true
		}
	}
	functions := make([]bool, len(l.result.Functions))
	closures := false
	for changed := true; changed; {
		changed = false
		for index, function := range l.result.Functions {
			if !functions[index] && l.regexpDiagnosticLeaves(function.Body, functions, closures) {
				functions[index], changed = true, true
			}
			if functions[index] && (function.Closure || dispatched[index]) && !closures {
				closures, changed = true, true
			}
		}
	}
	if l.regexpDiagnosticLeaves(l.result.Main, functions, closures) {
		return l.notYet(l.program.Files()[0].AsNode(), "a potentially unhandled RegExp SyntaxError (V8 diagnostic wording is not yet implemented)")
	}
	return nil
}

func (l *lowering) regexpDiagnosticLeaves(statements []ir.Statement, functions []bool, closures bool) bool {
	found := false
	walk(statements, func(node any) bool {
		switch value := node.(type) {
		case ir.Try:
			if value.HasCatch {
				found = found || l.regexpDiagnosticLeaves(value.Catch, functions, closures) || l.regexpDiagnosticLeaves(value.Finally, functions, closures)
				return false
			}
		case ir.RegExpNew:
			found = found || value.Invalid
		case ir.Call:
			for _, target := range l.result.CallTargets(value) {
				found = found || functions[target]
			}
		case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach:
			found = found || closures
		case ir.RegExpCall:
			found = found || (strings.HasSuffix(value.Method, "Callback") && closures)
		case ir.ArraySort:
			targets := l.result.ClosureTargets(value)
			found = found || (targets.Unknown && closures)
			for _, target := range targets.Functions {
				found = found || functions[target]
			}
		}
		return !found
	})
	return found
}
