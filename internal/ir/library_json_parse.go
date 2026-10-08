package ir

// JSONParseSchema is a boundary contract. Layout additionally describes the complete
// native storage, which structural object types alone cannot establish.
type JSONParseSchema struct {
	Kind       string
	Name       string
	Of         Type
	Optional   bool
	Literal    string
	HasLiteral bool
	Element    *JSONParseSchema
	Fields     []JSONParseField
	Members    []*JSONParseSchema
}
type JSONParseField struct {
	Name   string
	Schema *JSONParseSchema
}
type JSONParse struct {
	Text          Expression
	Check, Layout *JSONParseSchema
	Of            Type
}

func (p JSONParse) Type() Type { return p.Of }
