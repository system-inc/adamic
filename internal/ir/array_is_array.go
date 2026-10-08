package ir

// ArrayIsArray tests the runtime array brand, including through a boxed union.
type ArrayIsArray struct{ Value Expression }

func (ArrayIsArray) Type() Type { return Boolean }
