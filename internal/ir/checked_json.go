package ir

// JSONContract is a reifiable structural use, not a promise about an any value.
// Optional fields and undefined alternatives permit source recovery values.
type JSONContract struct {
	ID             int
	Reference      int
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

// JSONContractGraph gives both backends a deterministic finite table, including
// numeric recursive references; IR pointers remain acyclic. Runtime data
// recursion is bounded by the check.
func JSONContractGraph(root *JSONContract) []*JSONContract {
	var nodes []*JSONContract
	seen := map[*JSONContract]bool{}
	var visit func(*JSONContract)
	visit = func(c *JSONContract) {
		if c == nil || seen[c] {
			return
		}
		seen[c] = true
		nodes = append(nodes, c)
		visit(c.Element)
		for _, f := range c.Fields {
			visit(f.Contract)
		}
		for _, a := range c.Alternatives {
			visit(a)
		}
	}
	visit(root)
	return nodes
}
