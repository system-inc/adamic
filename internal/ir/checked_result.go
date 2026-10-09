package ir

// ContractResult checks the resolved overload's promise without changing the
// contracts of the returned allocation. It evaluates the call exactly once.
type ContractResult struct {
	Value      Expression
	Fields     []ContractField
	Expression string
}

func (value ContractResult) Type() Type { return value.Value.Type() }
