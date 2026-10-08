package ir

// RecordCall operates on own string data entries. Arguments are evaluated in order.
// Element is the representation stored in each slot; Returns includes missing reads.
type RecordCall struct {
	Method    string
	Arguments []Expression
	Element   Type
	Returns   Type
	Site      int
	// OwnOnly changes prototype handling, never the value provenance or ownership.
	OwnOnly bool
	// DictionaryKeys enumerates actual source keys without reading element values.
	DictionaryKeys bool
	// Demand metadata for operations not yet covered by the dictionary read adapter.
	ViewTypeID     int
	ViewWhere      string
	DictionaryRead *Property
}

func (c RecordCall) Type() Type { return c.Returns }

type RecordEntry struct{ Key, Value Expression }
type RecordLiteral struct {
	Element Type
	Spread  Expression
	Entries []RecordEntry
	Site    int
}

func (RecordLiteral) Type() Type { return Record }

// RecordCoalesce is r[k] ??= value: the receiver and key are evaluated once,
// and the fallback is evaluated and assigned only for a missing/undefined value.
type RecordCoalesce struct {
	Record, Key, Value Expression
	Element            Type
	Site               int
}

func (r RecordCoalesce) Type() Type { return r.Element }
