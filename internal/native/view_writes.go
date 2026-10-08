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
		drop := fmt.Sprintf("adamic_release(%s->reference)", slot)
		if len(e.program.GraphTypes) != 0 {
			drop = fmt.Sprintf("adamic_graph_drop(%s,%s->reference)", object, slot)
		}
		e.line("if (adamic_object_field_types(%s)[%s.index] == 7) { %s->number = adamic_maybe_number_pack((adamic_maybe_number){false,0}); } else if (adamic_object_field_types(%s)[%s.index] == 9) { %s->maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){false,false}); } else { if (%s->shape->references[%s.index]) %s; %s->reference = NULL; adamic_object_field_types(%s)[%s.index] = 13; }", object, cache, slot, object, cache, slot, object, cache, drop, slot, object, cache)
	} else if of == ir.Number || of == ir.MaybeNumber {
		packed, number := fmt.Sprintf("adamic_maybe_number_pack((adamic_maybe_number){true,%s})", value), value
		if of == ir.MaybeNumber {
			packed, number = "adamic_maybe_number_pack("+value+")", "("+value+").number"
		}
		e.line("if (adamic_object_field_types(%s)[%s.index] == 7) { %s->number = %s; } else { %s->number = %s; adamic_object_field_types(%s)[%s.index] = 1; }", object, cache, slot, packed, slot, number, object, cache)
	} else if of == ir.Boolean || of == ir.MaybeBoolean {
		// Preserve a maybe-boolean slot's byte encoding for every alias, including
		// readers whose proof lets them load the slot without the view adapter.
		packed := fmt.Sprintf("adamic_maybe_boolean_pack((adamic_maybe_boolean){true,%s})", value)
		boolean, tag := value, "2"
		if of == ir.MaybeBoolean {
			packed = "adamic_maybe_boolean_pack(" + value + ")"
			boolean, tag = "("+value+").boolean", "("+value+").present ? 2 : 13"
		}
		e.line("if (adamic_object_field_types(%s)[%s.index] == 9) { %s->maybe_boolean = %s; } else { %s->boolean = %s; adamic_object_field_types(%s)[%s.index] = %s; }", object, cache, slot, packed, slot, boolean, object, cache, tag)
	} else if of == ir.String || of == ir.Object || of == ir.Union {
		old := e.temporary()
		e.line("void *%s = %s->reference;", old, slot)
		e.line("%s->reference = %s;", slot, e.keptIn(object, value))
		e.dropIn(object, old)
		e.line("adamic_object_field_types(%s)[%s.index] = %d;", object, cache, of)
	} else {
		panic("compiler bug: unsupported scalar checked write")
	}
	e.end()
}
