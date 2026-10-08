package ir

// Array views currently use the same conservative program-wide read policy as
// object fields. All index reads retain their own declared element contract.
func HasArrayViews(program *Program) bool {
	// Constructor schemas can intern arrays without introducing a checked view.
	// Only admitted view origins enable the conservative program-wide policy.
	if len(program.ViewOrigins) == 0 {
		// Boxed primitive array consumers require the same adapters even without an
		// assertion: raw scalar slots and boxed references must never be confused.
		for _, contract := range program.ViewContracts {
			if contract.Kind == ViewArray && contract.Element != 0 && program.ViewContracts[contract.Element-1].Of == Union && PrimitiveArrayContract(program, contract.Element) {
				return true
			}
		}
		return false
	}
	for _, contract := range program.ViewContracts {
		if contract.Kind == ViewArray {
			return true
		}
	}
	return false
}

// Metadata has no executable operands. It must not masquerade as an index
// expression in CFG, borrowing, reuse, or ownership traversals.
type ArrayViewRead struct {
	View, ViewType   string
	ViewAllowed      []ViewLiteral
	ViewContract     ViewContractID
	ViewTypeID       int
	Element          Type
	Required         bool
	UndefinedAllowed bool
}

func (read ArrayViewRead) Index() ArrayIndex {
	return ArrayIndex{View: read.View, ViewType: read.ViewType, ViewAllowed: read.ViewAllowed, ViewContract: read.ViewContract, ViewTypeID: read.ViewTypeID, Element: read.Element, Required: read.Required, UndefinedAllowed: read.UndefinedAllowed}
}
