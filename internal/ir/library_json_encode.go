package ir

// JSONEncode writes the declared data graph, never the value's hidden runtime fields.
type JSONEncode struct {
	Value  Expression
	Schema JSONDecodeSchema
}

func (JSONEncode) Type() Type { return String }
