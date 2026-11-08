package ir

// MethodCallee is the field a call through an interface reads its callee by (Property's Method), and
// the Defined around it, if one is: a CallClosure's Closure is either the Property itself or, where the
// checker narrowed undefined out of an optional field (if (o.m) o.m()), the one Defined lowering's
// callClosure keeps around it (#0bxc4jh). Both are the same call: the callee is looked up on the object
// and called with it as this, so a host's method or a literal method gets its receiver. The Defined's
// check is then made on what the lookup found, after the object and before the arguments, in its own
// words. defined is nil for a bare Property. Every pass that asks whether a call is a method call asks
// this, so none of them takes the narrowed one for a plain function value
// (TestEveryCalleeFieldReadGoesThroughMethodCallee).
func MethodCallee(closure Expression) (property Property, defined *Defined, ok bool) {
	switch closure := closure.(type) {
	case Property:
		return closure, nil, closure.Method
	case Defined:
		if property, isProperty := closure.Value.(Property); isProperty && property.Method {
			return property, &closure, true
		}
	}
	return Property{}, nil, false
}
