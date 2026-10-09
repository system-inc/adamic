package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) censusToBoolean(argument ir.Expression) string {
	value := e.snapshot(argument.Type(), e.value(argument))
	result := ""
	switch argument.Type() {
	case ir.Boolean:
		result = value
	case ir.Number:
		result = fmt.Sprintf("(%s != 0.0 && !isnan(%s))", value, value)
	case ir.MaybeNumber:
		result = fmt.Sprintf("(%s.present && %s.number != 0.0 && !isnan(%s.number))", value, value, value)
	case ir.MaybeBoolean:
		result = fmt.Sprintf("(%s.present && %s.boolean)", value, value)
	case ir.String:
		result = fmt.Sprintf("(%s != NULL && %s->length != 0)", value, value)
	case ir.Union:
		result = "adamic_census_to_boolean(" + value + ")"
	default:
		result = "(" + value + " != NULL)"
	}
	return e.snapshot(ir.Boolean, result)
}
