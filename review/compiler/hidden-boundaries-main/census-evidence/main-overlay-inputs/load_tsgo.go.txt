package load

// EnableTSGo opts a native compilation into the external checker library. The
// driver must select its archive and use native.TSGoC and native.BuildTSGo.
func (p *Program) EnableTSGo()       { p.tsgo = true }
func (p *Program) TSGoEnabled() bool { return p.tsgo }
