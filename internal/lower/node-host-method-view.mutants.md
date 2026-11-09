# Mutants for the host method view rule (#sn788gx)

Each is applied alone to internal/lower/library_node.go, then `go test ./internal/oracle -run 'TestNodeHostMethodViewNotYet|TestNodeHostViewRegisteredAgrees' -count=1`.

(a) The clause taken out. Replace

	if host && programDeclared(property) && nodeHostMethodOffShape(own, member) {

with

	if false && host && programDeclared(property) && nodeHostMethodOffShape(own, member) {

All seven TestNodeHostMethodViewNotYet pins go red. Run with ADAMIC_HOST_METHOD_MUTANT=1 and each pin must lower and disagree with Node.

(b) performance's now dropped from the slot table. Replace

	"node:perf_hooks.Performance.now":           true,

with

	"node:perf_hooks.Performance.now":           false,

TestNodeHostViewRegisteredAgrees/now_related goes red: performance's now is a slot, and a method view of it lowers and agrees.
