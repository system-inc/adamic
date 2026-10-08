package ir

// ClosureOperands gives a function-value call's evaluation operands in source
// order. Lowering can rebind them without inferring which function will run;
// analyses of the possible callees still use ClosureTargets.
func ClosureOperands(call CallClosure) []Expression {
	return append([]Expression{call.Closure}, call.Arguments...)
}
