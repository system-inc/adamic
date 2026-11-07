package load

// EnableTSGo opts a native compilation into the external checker library. The
// driver must select its archive and use native.TSGoC and native.BuildTSGo.
func (program *Program) EnableTSGo()       { program.tsgo = true }
func (program *Program) TSGoEnabled() bool { return program.tsgo }
