All mutations were selected from production code before any mutant run. Fixed menu only. Probes are separate. Starting commit: 6d890e0129beecb2df9f65c89b7733a6f02b5248

M01 cmd/adamic-test262/cache.go:49 drop statement: fmt.Fprintf(hash, "%d:", len(part)) -> _ = part
M02 cmd/adamic-test262/cache.go:56 change constant: "test262-compiler-v1", program, compiler, command, context -> "test262-compiler-v1", "", compiler, command, context
M03 cmd/adamic-test262/cache.go:56 change constant: "test262-compiler-v1", program, compiler, command, context -> "test262-compiler-v1", program, "", command, context
M04 cmd/adamic-test262/cache.go:60 change constant: "test262-node-v1", program, version, adaptation, command, context -> "test262-node-v1", "", version, adaptation, command, context
M05 cmd/adamic-test262/cache.go:60 change constant: "test262-node-v1", program, version, adaptation, command, context -> "test262-node-v1", program, "", adaptation, command, context
M06 cmd/adamic-test262/cache.go:60 change constant: "test262-node-v1", program, version, adaptation, command, context -> "test262-node-v1", program, version, "", command, context
M07 cmd/adamic-test262/cache.go:63 change constant: "test262-native-v1", code, library, command, context -> "test262-native-v1", "", library, command, context
M08 cmd/adamic-test262/cache.go:63 change constant: "test262-native-v1", code, library, command, context -> "test262-native-v1", code, "", command, context
M09 cmd/adamic-test262/cache.go:35 change constant: []byte(value.Stdout), []byte(value.Stderr) -> []byte{}, []byte(value.Stderr)
M10 cmd/adamic-test262/cache.go:35 change constant: value.Exit, value.Signal, value.TimedOut -> value.Exit, "", value.TimedOut
M11 cmd/adamic-test262/cache.go:67 change constant: os.Getenv("ADAMIC_GATE_UNCACHED") == "1" -> os.Getenv("ADAMIC_GATE_UNCACHED") == "0"
M12 cmd/adamic-test262/cache.go:84 change condition option: envelope.Digest == cacheKey(string(envelope.Value)) -> true
M13 cmd/adamic-test262/cache.go:90 change condition option: !reusable || result.TimedOut || -> !reusable || false ||
M14 cmd/adamic-test262/cache.go:90 change constant: result.Exit == -1 && result.Signal == "" -> result.Exit == -2 && result.Signal == ""
M15 cmd/adamic-test262/cache.go:109 drop statement: _ = os.Rename(temporary.Name(), path) -> _ = path
M16 cmd/adamic-test262/run.go:358 flip condition: importedProgram.MatchString(codeOnly(program)) || referencedProgram.MatchString(program) -> importedProgram.MatchString(codeOnly(program)) && referencedProgram.MatchString(program)
M17 cmd/adamic-test262/run.go:154 off-by-one bound: attempted >= limit -> attempted > limit
M18 cmd/adamic-test262/run.go:213 change constant: (index+1)%50 == 0 -> (index+1)%49 == 0
M19 cmd/adamic-test262/compiler.go:120 change constant: worker.encoder.Encode(path) -> worker.encoder.Encode("")
M20 cmd/adamic-test262/compiler.go:146 change constant: return execution{Exit: -1, TimedOut: true} -> return execution{Exit: -1, TimedOut: false}
P01 cmd/adamic-test262/cache.go:75 empty-answer probe: func (cache *resultCache) reuse(key string, execute func() (execution, bool)) execution { -> func (cache *resultCache) reuse(key string, execute func() (execution, bool)) execution {
	return execution{}
P02 cmd/adamic-test262/cache.go:66 empty-answer probe: func (cache *resultCache) observe(key string, execute func() (execution, bool)) execution { -> func (cache *resultCache) observe(key string, execute func() (execution, bool)) execution {
	return execution{}
P03 cmd/adamic-test262/cache.go:46 empty-answer probe: func cacheKey(parts ...string) string { -> func cacheKey(parts ...string) string {
	return ""
P04 cmd/adamic-test262/cache.go:55 empty-answer probe: func compilerResultKey(program, compiler, command, context string) string { -> func compilerResultKey(program, compiler, command, context string) string {
	return ""
P05 cmd/adamic-test262/cache.go:59 empty-answer probe: func nodeResultKey(program, version, adaptation, command, context string) string { -> func nodeResultKey(program, version, adaptation, command, context string) string {
	return ""
P06 cmd/adamic-test262/cache.go:62 empty-answer probe: func nativeResultKey(code, library, command, context string) string { -> func nativeResultKey(code, library, command, context string) string {
	return ""
P07 cmd/adamic-test262/run.go:127 empty-answer probe: func (e *engine) runFilter(filter string, limit int, classifyOnly bool) (filterReport, error) { -> func (e *engine) runFilter(filter string, limit int, classifyOnly bool) (filterReport, error) {
	return filterReport{}, nil
P08 cmd/adamic-test262/run.go:232 empty-answer probe: func (e *engine) attempt(test classified) result { -> func (e *engine) attempt(test classified) result {
	return result{}
P09 cmd/adamic-test262/compiler.go:94 empty-answer probe: func (worker *compilerWorker) compile(path string) execution { -> func (worker *compilerWorker) compile(path string) execution {
	return execution{}
P10 cmd/adamic-test262/compiler.go:38 empty-answer probe: func compileInProcess(path string) (result execution) { -> func compileInProcess(path string) (result execution) {
	return execution{}
P11 cmd/adamic-test262/run.go:357 empty-answer probe: func dependentProgram(program string) bool { -> func dependentProgram(program string) bool {
	return false