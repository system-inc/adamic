package ir

// GeneratorTypes keeps the three independent promises at a suspension boundary.
// Factory parameters are owned even though the factory returns before its body starts.
type GeneratorTypes struct {
	Yield, Return, Next Type
	Lowering            bool
	Prologue            int
	NextAllowsUndefined bool
	Frame               int
	FrameLocals         []int
}

// GeneratorYield exists only while a generator body is lowered. Suspension lowering
// removes it before ordinary backend emission and ownership analysis.
type GeneratorYield struct {
	Value    Expression
	Where    string
	Next     Type
	Delegate bool
	Element  Type
	Result   Type
	Kind     string
}

func (y GeneratorYield) Type() Type {
	if y.Delegate {
		return y.Result
	}
	return y.Next
}

// GeneratorFrame identifies a compiler-created ordinary object whose slots own
// every value kept across a suspension. It has no special runtime representation.
type GeneratorFrame struct{ Value Expression }

func (GeneratorFrame) Type() Type { return Object }
