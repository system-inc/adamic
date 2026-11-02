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

// DynamicKeyedProperty is Reflect.get(Object(v), k, v), a read of an unknown by a key known only at
// run time (lower/reflect_get.go). Lowering admits it only where the checker proved v neither null nor
// undefined, so Object(v) is v and the receiver is moot. Each backend answers an own data slot, a
// record's own key or a class getter, and stops loudly where the representation can't answer as Node
// does (DynamicKeyedStops). The result is owned, as DynamicProperty's is.
type DynamicKeyedProperty struct {
	Object Expression
	Key    Expression
}

func (DynamicKeyedProperty) Type() Type { return Union }

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
