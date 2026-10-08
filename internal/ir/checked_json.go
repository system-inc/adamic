package ir

// JSONContract is a reifiable structural use, not a promise about an any value.
// Optional fields and undefined alternatives permit source recovery values.
type JSONContract struct {
	Kind           string
	Name           string
	LiteralText    string
	LiteralNumber  float64
	LiteralBoolean bool
	Element        *JSONContract
	Fields         []JSONContractField
	Alternatives   []*JSONContract
}
type JSONContractField struct {
	Name     string
	Optional bool
	Contract *JSONContract
}
type CheckedJSON struct {
	Value    Expression
	Contract *JSONContract
	Path     string
	Of       Type
}

func (v CheckedJSON) Type() Type { return v.Of }

// JSONReadType reports representations used by reifiable JSON contracts. Other
// fields with the same name keep their ordinary typed emitter path.
func JSONReadType(t Type) bool {
	if t.IsMaybe() {
		return JSONReadType(t.Present())
	}
	switch t {
	case Number, Boolean, String, Array, Object, Union:
		return true
	}
	return false
}
