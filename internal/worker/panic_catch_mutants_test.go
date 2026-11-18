package worker

import "testing"

var panicCatchMutants = []workerMutation{
	{"panic-catch-identity", "internal/javascript/javascript.go", `e.line("if (adamicIsPanic(%s)) throw %s;", caught, caught)`, `_ = caught`, "TestWorkerPanicIsNeverCaught/catch", "a panic reached the program catch"},
	{"panic-finally-guard", "internal/javascript/javascript.go", `e.line("if (!adamicProcessExiting() && !%s) {", panicking)`, `e.line("if (!adamicProcessExiting()) {"); _ = panicking`, "TestWorkerPanicIsNeverCaught/finally", "a finally ran while a panic unwound"},
	{"panic-finally-unset", "internal/javascript/javascript.go", `e.line("\tif (adamicIsPanic(%s)) %s = true;", unwinding, panicking)`, `e.line("\tif (adamicIsPanic(%s)) %s;", unwinding, panicking)`, "TestWorkerPanicIsNeverCaught/finally_return", "a finally return swallowed a panic"},
}

func TestProduct_WorkerPanicCatchMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, panicCatchMutants[0], false)
}
func TestProduct_WorkerPanicFinallyMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, panicCatchMutants[1], false)
}
func TestProduct_WorkerPanicUnsetMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, panicCatchMutants[2], false)
}
func TestWorkerPanicCatchMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, mutant := range panicCatchMutants {
		t.Run(mutant.name, func(t *testing.T) { t.Parallel(); runWorkerMutant(t, mutant) })
	}
}
