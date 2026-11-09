# Mutants for the host method view rule (#sn788gx)

Each is applied alone to internal/lower/library_node.go. Run the command each names.

(a) The clause taken out. Replace

	if host && programDeclared(property) && l.hostBindingViews()[view.Symbol()] == nil && l.nodeHostMethodOffShape(own, member) {

with

	if false && host && programDeclared(property) && l.hostBindingViews()[view.Symbol()] == nil && l.nodeHostMethodOffShape(own, member) {

`go test ./internal/oracle -run 'TestNodeHostMethodViewNotYet' -count=1`: all seven pins go red. With ADAMIC_HOST_METHOD_MUTANT=1 each lowers and disagrees with Node (six die with "a method the checker proved is there is missing", and the optional view prints true where Node prints false). `go test ./internal/lower -run 'TestNodeWatchInheritedOnMutant' -count=1` goes red too: under Library's mutant, ChildProcess's on view lowers again.

(b) performance's now dropped from the slot table. Replace

	"node:perf_hooks.Performance.now":           true,

with

	"node:perf_hooks.Performance.now":           false,

`go test ./internal/oracle -run 'TestNodeHostViewRegisteredAgrees' -count=1`: now_related goes red.

(c) FSWatcher's on dropped from the slot table. Replace

	"node:fs.FSWatcher.on":                      true,

with

	"node:fs.FSWatcher.on":                      false,

`go test ./internal/lower -run 'TestNodeWatchTscShape' -count=1` goes red: tsc's FsWatchWorkerWatcher view of an FSWatcher reads on, which the runtime holds as a slot.

(d) The routed host binding judged as any view. Replace

	l.hostBindingViews()[view.Symbol()] == nil && l.nodeHostMethodOffShape(own, member) {

with

	l.nodeHostMethodOffShape(own, member) {

`go test ./internal/lower -run 'TestHostBindingProcessReads' -count=1` goes red: core.ts's process binding names nextTick, whose reads route through node:process.
