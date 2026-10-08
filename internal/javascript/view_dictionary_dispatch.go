package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) dictionaryRead(property ir.Property) string {
	kinds, ok := ir.DictionaryReadKinds(e.program, property.ViewContract)
	if !ok {
		panic("compiler bug: incomplete dictionary read contract")
	}
	quoted := []string{}
	for _, kind := range kinds {
		quoted = append(quoted, quote(kind))
	}
	value := fmt.Sprintf("adamicViewDictionaryRead(%s, %s, [%s], %d, %s, %s).value", e.value(property.Object), e.value(property.DictionaryKey), strings.Join(quoted, ", "), property.ViewContract, quote(property.View), quote(property.ViewType))
	if len(property.ViewAllowed) != 0 {
		value = fmt.Sprintf("((value) => { if (value !== undefined && ![%s].includes(value)) panic('cast failed: field read failed: ' + %s + ' expected ' + %s + ', found ' + typeof value + ' ' + value); return value; })(%s)", e.values(property.ViewAllowed), quote(property.View), quote(property.ViewType), value)
	}
	if property.Of == ir.Object {
		return e.viewObjectUnion(property, value)
	}
	return value
}
