package ir

// FunctionSurface is source reflection, not an argument representation or a
// producer contract. Defaults and rest stop length; optional parameters do not.
type FunctionSurface struct {
	Name   string
	Length int
}

// FunctionValueSurface follows implementation-only forwarders without assigning
// their generated names or argument slots to the original source function.
func (p *Program) FunctionValueSurface(index int) *FunctionSurface {
	for steps := 0; steps < len(p.Functions); steps++ {
		function := p.Functions[index]
		if function.Surface != nil {
			return function.Surface
		}
		if function.ForwardsArguments == 0 {
			return nil
		}
		index = function.ForwardsArguments - 1
	}
	return nil
}
