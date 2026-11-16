package worker

import "testing"

// One observation mutant is run separately against every L3 pin: removing the
// common NotYet cannot be hidden by a later lowering stop or another pin.
var environmentShapeMutants = []workerMutation{
	{"environment-shape-observed", "internal/lower/worker_environment.go", `return r.l.notYet(node, "worker environment "+form+workerEnvironmentLift)`, `return nil`, "TestWorkerEnvironmentShapeStops/keys", "environment boundary keys"},
	{"environment-class-holder", "internal/lower/worker_environment.go", `if t != nil && !allowed && r.holds(t, map[*checker.Type]bool{}, 0) {`, `if false && t != nil && !allowed && r.holds(t, map[*checker.Type]bool{}, 0) {`, "TestWorkerEnvironmentShapeStops/class_field", "environment boundary class_field"},
	{"environment-parameter-write", "internal/lower/worker_environment.go", `if ast.IsAssignmentTarget(node) && r.parameterField(node) {`, `if false && ast.IsAssignmentTarget(node) && r.parameterField(node) {`, "TestWorkerEnvironmentShapeStops/parameter_write", "environment boundary parameter_write"},
	{"environment-local-destructure", "internal/lower/worker_environment.go", `if node.Name().Kind != ast.KindIdentifier {`, `if node.Name().Kind != ast.KindIdentifier { return r.refused(node,"a const destructuring");`, "TestWorkerEnvironmentFieldReadRoads/const_destructure", "environment field read build:"},
	{"environment-parameter-destructure", "internal/lower/worker_environment.go", "case ast.KindParameter:\n\t\tt =", "case ast.KindParameter:\n if node.Name().Kind == ast.KindObjectBindingPattern {return r.refused(node,\"a parameter destructuring\")};\n\t\tt =", "TestWorkerEnvironmentFieldReadRoads/parameter_destructure", "environment field read build:"},
	{"environment-optional-read", "internal/lower/worker_environment.go", `if r.declared(parent) {`, `if r.declared(parent) { if parent.Flags&ast.NodeFlagsOptionalChain!=0 {return r.observation(node,"optional read")};`, "TestWorkerEnvironmentFieldReadRoads/optional_read", "environment field read build:"},
}

func TestProduct_WorkerEnvironmentShapeMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[0], false)
}
func TestProduct_WorkerEnvironmentHolderMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[1], false)
}
func TestProduct_WorkerEnvironmentWriteMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[2], false)
}
func TestWorkerEnvironmentShapeMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range environmentShapeMutants {
		t.Run(m.name, func(t *testing.T) { t.Parallel(); runWorkerMutant(t, m) })
	}
	for _, name := range []string{"let", "array", "module_binding", "return", "optional_parameter", "this_parameter"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := environmentShapeMutants[1]
			m.check = "TestWorkerEnvironmentShapeStops/" + name
			m.witness = "environment boundary " + name
			runWorkerMutant(t, m)
		})
	}

	for _, name := range []string{"in", "values", "entries", "for_in", "spread", "json", "has_own", "reflect", "object", "unknown", "record", "structural", "union", "alias_keys", "rest_destructure", "generic_argument"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := environmentShapeMutants[0]
			m.check = "TestWorkerEnvironmentShapeStops/" + name
			m.witness = "environment boundary " + name
			runWorkerMutant(t, m)
		})
	}
}

func TestProduct_WorkerEnvironmentLocalDestructureMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[3], false)
}
func TestProduct_WorkerEnvironmentParameterDestructureMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[4], false)
}
func TestProduct_WorkerEnvironmentOptionalReadMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentShapeMutants[5], false)
}
