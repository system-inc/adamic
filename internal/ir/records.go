package ir

// RecordCall operates on own string data entries. Arguments are evaluated in order.
// Element is the representation stored in each slot; Returns includes missing reads.
// "defineHidden" is Object.defineProperty(record, key, { enumerable: false, writable: false, value })
// with record, key and value its arguments: it returns false, storing nothing, when the key is hidden
// already, for the caller to throw.
type RecordCall struct {
	Method    string
	Arguments []Expression
	Element   Type
	Returns   Type
	Site      int
	// OwnOnly changes prototype handling, never the value provenance or ownership.
	OwnOnly bool
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

// RecordWriteMayThrow says a node can throw because it writes or deletes a record entry that may be
// a hidden key (HiddenRecordKeys): a record write, ??= or delete whose key is known only as it runs,
// or is a literal some define hides. A write to a fixed object never meets one: a record seen as a
// fixed object is refused (lowering's recordStorageView), and Library's #6c4swvt, admitting those
// views, is the seam that must add the check to native's writeProperty.
func (p *Program) RecordWriteMayThrow(node any) bool {
	if len(p.HiddenRecordKeys) == 0 {
		return false
	}
	switch node := node.(type) {
	case RecordCall:
		return (node.Method == "set" || node.Method == "delete") && p.mayBeHiddenKey(node.Arguments[1])
	case RecordCoalesce:
		return p.mayBeHiddenKey(node.Key)
	}
	return false
}

// mayBeHiddenKey says a record key may be one some define hides: a literal one is exactly when a
// define names it, and any other may be.
func (p *Program) mayBeHiddenKey(key Expression) bool {
	if constant, literal := key.(StringConstant); literal {
		return p.HiddenRecordKeys[p.Strings[constant.Index]]
	}
	return true
}
