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

func (call RecordCall) Type() Type { return call.Returns }

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

func (coalesce RecordCoalesce) Type() Type { return coalesce.Element }

// RecordStorage is whether Value, an object (an UnknownObject, so neither a class's nor an Error),
// must be read by its keys rather than its slots: natively, whether it's record storage, a
// dictionary behind one slot. A weak field read asks it (lower's weakFieldRead), since tsc lets a
// record be seen as a type whose every field is optional. Native stops on a record of numbers or
// booleans, whose unboxed values no read by key takes. On the JavaScript backend a record and a
// fixed object are both plain objects, read by key alike, so it's always true there.
type RecordStorage struct{ Value Expression }

func (RecordStorage) Type() Type { return Boolean }

// RecordWriteMayThrow says a node can throw because it writes or deletes a record entry that may be
// a hidden key (HiddenRecordKeys): a record write, ??= or delete whose key is known only as it runs,
// or is a literal some define hides, and an object write under such a name, which native's
// writeProperty stores through a record when one arrives there.
func (program *Program) RecordWriteMayThrow(node any) bool {
	if len(program.HiddenRecordKeys) == 0 {
		return false
	}
	switch node := node.(type) {
	case RecordCall:
		return (node.Method == "set" || node.Method == "delete") && program.mayBeHiddenKey(node.Arguments[1])
	case RecordCoalesce:
		return program.mayBeHiddenKey(node.Key)
	case SetProperty:
		return program.HiddenRecordKeys[node.Name]
	}
	return false
}

// mayBeHiddenKey says a record key may be one some define hides: a literal one is exactly when a
// define names it, and any other may be.
func (program *Program) mayBeHiddenKey(key Expression) bool {
	if constant, literal := key.(StringConstant); literal {
		return program.HiddenRecordKeys[program.Strings[constant.Index]]
	}
	return true
}
