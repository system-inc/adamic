package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

type checkedViewReadHook func(*emitter, ir.Property, string) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) checkedViewField(property ir.Property) string {
	return e.checkedViewFieldAt(property, "")
}

// checkedViewFieldAt is checkedViewField from object, the receiver already evaluated, or from
// property.Object when object is empty.
func (e *emitter) checkedViewFieldAt(property ir.Property, object string) string {
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property, object); handled {
			return value
		}
	}
	if object == "" {
		object = e.value(property.Object)
	}
	if property.Optional {
		return throughOptionalReceiver(e, property, object, func(e *emitter, property ir.Property, object string) (string, bool) {
			return e.checkedViewFieldAt(property, object), true
		})
	}
	expected := property.ViewType
	if expected == "" {
		expected = map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}[property.Of]
	}
	return fmt.Sprintf("adamicViewField(%s, %s, %s, %d, %s, [%s]%s)", object, quote(property.Name), quote(property.View), property.Of, quote(expected), e.values(property.ViewAllowed), e.viewStatics(property))
}

// viewStatics is adamicViewField's last arguments where a checked read expects an object outside a union
// whose contract admits a class constructor (ir.Program's ObjectReadContract, StaticsAdmitted): the fields
// it requires, which one has to hold as data statics, its own or a base's, to be the object read, then the
// optional ones, none of which may be a name the walk refuses, as native's staticViewRead admits it.
// Anywhere else it's empty, and adamicViewField finds a class constructor a function.
func (e *emitter) viewStatics(property ir.Property) string {
	contract, checked := e.program.ObjectReadContract(property)
	if !checked || !contract.StaticsAdmitted {
		return ""
	}
	required, optional := []string{}, []string{}
	for _, field := range contract.Fields {
		if field.Optional {
			optional = append(optional, quote(field.Name))
		} else {
			required = append(required, quote(field.Name))
		}
	}
	return ", [" + strings.Join(required, ", ") + "], [" + strings.Join(optional, ", ") + "]"
}

// throughOptionalReceiver is a ?. read through a checked view: undefined from an absent receiver
// before any field is looked at, and a present one checked as read does any other.
func throughOptionalReceiver(e *emitter, property ir.Property, object string, read checkedViewReadHook) string {
	property.Optional = false
	checked, _ := read(e, property, "adamicReceiver")
	return fmt.Sprintf("((adamicReceiver) => adamicReceiver === undefined || adamicReceiver === null ? undefined : %s)(%s)", checked, object)
}

// checkedViewWrite is object.name = value through a checked view, both already evaluated.
func (e *emitter) checkedViewWrite(statement ir.SetProperty, object, value string) string {
	write := func(object, value string) string {
		return fmt.Sprintf("adamicViewWrite(%s, %s, %s, %d, %s)", object, quote(statement.Name), value, viewFieldRepresentation(statement.Value), quote(statement.DeclaredType))
	}
	return e.staticViewWrite(statement, object, value, write)
}

// staticViewWrite is write(object, value), stopped first when a view a class constructor may stand for
// holds one (ir.Program's StaticsAdmittedReceiver), as native's adamic_view_static_write stops it: such a
// view reads a class's statics and never writes them. Native evaluates the receiver, then the value, then
// guards, so here the two are each evaluated once, in that order, before the guard.
func (e *emitter) staticViewWrite(statement ir.SetProperty, object, value string, write func(object, value string) string) string {
	if !e.program.StaticsAdmittedReceiver(statement.ViewReceiverTypeID) {
		return write(object, value)
	}
	held, assigned := e.temporary(), e.temporary()
	stop := quote("field write failed: a write to " + statement.Name + " through a checked view that holds a class constructor is not yet lowered")
	return fmt.Sprintf("((%s, %s) => { if (adamicViewConstructor(%s)) panic(%s); return %s; })(%s, %s)", held, assigned, held, stop, write(held, assigned), object, value)
}
