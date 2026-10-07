package ir

// JSONParse is a library operation, not an any-valued language expression. Scalar results are
// proven from literal text or a scalar reviver's return type. Canonical results are private
// carriers consumed immediately by JSONStringify; discarded results never escape.
type JSONParse struct {
	Text           Expression
	Reviver        Expression
	IgnoredReviver Expression
	ReviverKind    string
	Takes          int
	Mode           string
	Of             Type
}

func (p JSONParse) Type() Type { return p.Of }
