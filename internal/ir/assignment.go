package ir

// AssignmentValue performs an existing store and yields its new value. The store
// holds the target representation; lowering projects the checked result afterward.
type AssignmentValue struct {
	Store Statement
	Value Expression
}

func (a AssignmentValue) Type() Type {
	if a.Value != nil {
		return a.Value.Type()
	}
	switch store := a.Store.(type) {
	case Assign:
		return store.Value.Type()
	case SetProperty:
		return store.Value.Type()
	case SetIndex:
		return store.Value.Type()
	default:
		panic("ir: assignment value without a represented store")
	}
}
