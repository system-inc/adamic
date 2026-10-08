package ir

// Array reads and source writes share the existing producer signature checker.
func ArrayCallableContract(program *Program, id ViewContractID) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	c := program.ViewContracts[id-1]
	return c.Kind == ViewCallable && c.Of == Closure && c.Unsupported == "" && (c.Result != 0 || c.DiscardResult)
}
