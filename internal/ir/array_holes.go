package ir

// ArrayHoles constructs a length-form Array. Present slots live separately from
// absent indices, so a present undefined can never stand for a hole.
type ArrayHoles struct {
	Length  Expression
	Element Type
}

func (ArrayHoles) Type() Type { return Array }

// ArrayRangeErrorIs observes constructor identity independently of .name.
type ArrayRangeErrorIs struct{ Value Expression }

func (ArrayRangeErrorIs) Type() Type { return Boolean }

// ArraySetLength applies ArraySetLength semantics and may throw RangeError.
type ArraySetLength struct{ Array, Length Expression }

func (ArraySetLength) Type() Type { return Number }
