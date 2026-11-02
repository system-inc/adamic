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
