package ir

// ArrayRecord adds a fixed set of own data properties to a fresh array. The
// producer copies the properties, preserving array identity and source aliases.
type ArrayRecord struct {
	Array      Expression
	Properties Expression
	Where      string
}

func (ArrayRecord) Type() Type { return Array }

// ArrayProperties exposes the array's existing own-field storage to the shared
// readiness, kind, literal and source-slot checks. It never visits elements.
type ArrayProperties struct{ Array Expression }

func (ArrayProperties) Type() Type { return Object }
