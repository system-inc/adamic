package ir

// MethodCallee is the field a call through an interface reads its callee by (Property's Method), and
// the check around it, if one is. A CallClosure's Closure is the Property itself, or one of two checks
// lowering's callClosure keeps around the call's own field read (#0bxc4jh): the Defined where the
// checker narrowed undefined out of an optional field (if (o.m) o.m()), or the non-null assertion's
// Coalesce, whose Panic is the assertion's failure (o.m!(), which JavaScript runs as o.m()). All three
// are the same call: the callee is looked up on the object and called with it as this, so a host's
// method or a literal method gets its receiver. The check is then made on what the lookup found,
// after the object and before the arguments, in its own words and kind. check is nil for a bare
// Property. Every pass that asks whether a call is a method call asks this, so none of them takes a
// checked one for a plain function value (TestEveryCalleeFieldReadGoesThroughMethodCallee).
func MethodCallee(closure Expression) (property Property, check Expression, ok bool) {
	switch closure := closure.(type) {
	case Property:
		return closure, nil, closure.Method
	case Defined:
		if property, isProperty := closure.Value.(Property); isProperty && property.Method {
			return property, closure, true
		}
	case Coalesce:
		if property, isProperty := closure.Value.(Property); isProperty && property.Method && closure.Panic != nil && closure.Fallback == nil {
			return property, closure, true
		}
	}
	return Property{}, nil, false
}

// MethodCalleeWith is check, as MethodCallee gave it, around property instead of the field it held: a
// pass that reads the method off a receiver of its own keeps the call's check (lower's
// checkedFunctionTypeCall). A nil check is the bare property.
func MethodCalleeWith(check Expression, property Property) Expression {
	switch check := check.(type) {
	case Defined:
		check.Value = property
		return check
	case Coalesce:
		check.Value = property
		return check
	}
	return property
}
