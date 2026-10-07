package lower

import "github.com/system-inc/adamic/internal/ir"

// Node's filesystem errors can override toString to include an error code.
// Nominal identity and standard prefix reads are supported, but that host
// override must not silently dispatch to the plain built-in Error method.
func (l *lowering) checkHostErrorMethods() error {
	methods := map[int]bool{}
	for _, name := range errorNames {
		if instance := l.instances["builtin-error:"+name]; instance != nil {
			methods[instance.methods["toString"]] = true
		}
	}
	host, method := false, false
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.NodeFSFile:
			host = host || value.MayThrow()
		case ir.Call:
			if value.Virtual != 0 {
				for _, target := range l.result.CallTargets(value) {
					method = method || methods[target]
				}
			}
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	if host && method {
		return &NotYet{Where: l.result.Source, What: "virtual Error.toString with filesystem failures whose host override includes an error code; use Error.prototype.toString.call(error) for standard formatting (adamic/host-error-to-string)"}
	}
	return nil
}
