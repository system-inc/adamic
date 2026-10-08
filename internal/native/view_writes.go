package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The shared checked view uses a static type certificate at the write, then
// converts to the actual slot's storage. Readiness changes only after success.
func (e *emitter) checkedWrite(write ir.SetProperty) {
	object, value := e.value(write.Object), e.value(write.Value)
	e.mapEntryNominalCertificate(write.WriteContract, value, write.WriteWhere)
	allowed := ir.ScalarWriteContracts(e.program, write.WriteContract)
	ids := []string{}
	for _, id := range allowed {
		ids = append(ids, fmt.Sprint(id))
	}
	pointer := "NULL"
	if len(ids) != 0 {
		pointer = "(const unsigned int[]){" + strings.Join(ids, ",") + "}"
	}
	cache := e.cache()
	if !write.WriteProven {
		e.line("adamic_object_checked_write(%s, %s, &%s, %s, %d, %s);", object, cString(write.Name), cache, pointer, len(ids), cString(write.WriteWhere))
	}
	e.line("adamic_object_check_data_write(%s, %s);", object, cString(write.Name))
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_object_write_field(%s, %s, &%s);", slot, object, cString(write.Name), cache)
	of := write.Value.Type()
	if _, undefined := write.Value.(ir.Undefined); undefined {
		e.line("if (adamic_object_field_types(%s)[%s.index] == 7) { %s->number = adamic_maybe_number_pack((adamic_maybe_number){false,0}); } else { if (%s->shape->references[%s.index]) adamic_release(%s->reference); %s->reference = NULL; adamic_object_field_types(%s)[%s.index] = 13; }", object, cache, slot, object, cache, slot, slot, object, cache)
	} else if of == ir.Number || of == ir.MaybeNumber {
		packed, number := fmt.Sprintf("adamic_maybe_number_pack((adamic_maybe_number){true,%s})", value), value
		if of == ir.MaybeNumber {
			packed, number = "adamic_maybe_number_pack("+value+")", "("+value+").number"
		}
		e.line("if (adamic_object_field_types(%s)[%s.index] == 7) { %s->number = %s; } else { %s->number = %s; adamic_object_field_types(%s)[%s.index] = 1; }", object, cache, slot, packed, slot, number, object, cache)
	} else if of == ir.Boolean {
		e.line("%s->boolean = %s;", slot, value)
		e.line("adamic_object_field_types(%s)[%s.index] = 2;", object, cache)
	} else if of == ir.MaybeBoolean {
		e.line("%s->boolean = (%s).boolean; adamic_object_field_types(%s)[%s.index] = (%s).present ? 2 : 13;", slot, value, object, cache, value)
	} else if of == ir.String || of == ir.Object || of == ir.Union {
		old := e.temporary()
		e.line("void *%s = %s->reference;", old, slot)
		e.line("%s->reference = %s;", slot, e.kept(value))
		e.line("adamic_release(%s);", old)
		e.line("adamic_object_field_types(%s)[%s.index] = %d;", object, cache, of)
	} else {
		panic("compiler bug: unsupported scalar checked write")
	}
	e.end()
}
