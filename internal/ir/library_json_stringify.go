package ir

// JSONSchema describes only values whose complete JSON behavior lowering proved. A literal's
// fields are complete; an ordinary structural object type does not establish that fact.
type JSONSchema struct {
	Null    bool // A missing reference is null rather than undefined.
	Kind    string
	Element *JSONSchema
	Fields  []JSONField
}
type JSONField struct {
	Name   int
	Slot   int
	Schema *JSONSchema
}
type JSONStringify struct {
	Value          Expression
	Schema         *JSONSchema
	Replacer       Expression
	ReplacerSchema *JSONSchema
	Space          Expression
	SpaceSchema    *JSONSchema
}

func (JSONStringify) Type() Type { return String }

// JSONNull occurs only inside stringify's input, until null has a general representation.
type JSONNull struct{}

func (JSONNull) Type() Type { return Object }
