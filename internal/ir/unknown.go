package ir

// HasProperty observes presence, including inherited names, without reading a value.
type HasProperty struct {
	Object Expression
	Name   string
}

func (HasProperty) Type() Type { return Boolean }

// DynamicProperty reads a runtime-tagged value. Its type remains unknown until narrowed.
type DynamicProperty struct {
	Object Expression
	Name   string
}

func (DynamicProperty) Type() Type { return Union }

// UnknownKind is what String(Value) needs to know of an unknown to dispatch it (lower/to_primitive.go):
// one of the UnknownKind numbers below, numbered once here for lowering and both backends. It reads
// only the value's own tags, and never calls.
type UnknownKind struct{ Value Expression }

func (UnknownKind) Type() Type { return Number }

// The numbers UnknownKind answers, in the order native's adamic_union_kind lists them. A tuple is an
// array on the JavaScript backend, so the two share UnknownArray, and a record is a plain object
// there, so it shares UnknownObject.
const (
	UnknownUndefined = iota
	UnknownNull
	UnknownNumber
	UnknownBoolean
	UnknownString
	UnknownError
	UnknownInstance
	UnknownStaticClass
	UnknownObject
	UnknownArray
	UnknownFunction
	UnknownCollection
	UnknownTypedArray
	UnknownOther
)

// OwnFunction is whether Object, an object UnknownKind called UnknownObject, has an own property
// Name holding a function, as ToPrimitive asks of toString and valueOf before it calls one. A
// property that's there and holds anything else isn't one, as ToPrimitive skips it.
type OwnFunction struct {
	Object Expression
	Name   string
}

func (OwnFunction) Type() Type { return Boolean }

// ErrorToString is Error.prototype.toString of Value, an Error UnknownKind called UnknownError: its
// name and message joined by ": ", or whichever one isn't empty. The string is new, owned.
type ErrorToString struct{ Value Expression }

func (ErrorToString) Type() Type { return String }

// ArrayOfScalars is whether Array, an array UnknownKind called UnknownArray, has elements and every one
// is a number, a boolean, null or undefined. String's join stops on one as not yet: natively an array
// of numbers or booleans holds them unboxed, its elements saying nothing of their kind, so the one test
// both backends make alike is by the elements, and such an array always passes it when it has any.
type ArrayOfScalars struct{ Array Expression }

func (ArrayOfScalars) Type() Type { return Boolean }

// JoinEnter is V8's join stack, which String(Array) pushes Array onto before joining its elements:
// false, pushing nothing, when Array is already on it, a join of itself under way, which joins as "".
// Only identity is kept, never a count, since the join's own caller keeps Array alive throughout.
type JoinEnter struct{ Array Expression }

func (JoinEnter) Type() Type { return Boolean }

// JoinLeave pops the array the innermost JoinEnter pushed, however its join ends, a throw included.
type JoinLeave struct{}

func (JoinLeave) Type() Type { return 0 }
