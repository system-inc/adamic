package ir

// A descriptor alone does not enable global read checks. An array view must be demanded.
func HasArrayViews(program *Program) bool { return program.ArrayViewEnabled }

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
