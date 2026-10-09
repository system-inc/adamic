package javascript

import (
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) callValues(arguments []ir.Expression, spread []bool) string {
	values := []string{}
	for index, argument := range arguments {
		value := e.value(argument)
		if len(spread) > index && spread[index] {
			value = "..." + value
		}
		values = append(values, value)
	}
	return strings.Join(values, ", ")
}

func tailSpreads(spread []bool) []bool {
	if len(spread) == 0 {
		return nil
	}
	return spread[1:]
}
