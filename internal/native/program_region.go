package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) adoptProgram(value, size string, member bool) {
	if member {
		e.line("%s = adamic_program_adopt_owned(%s, %s);", value, value, size)
	}
}
func (e *emitter) programObjectMember(literal ir.ObjectLiteral) bool {
	return literal.ProgramRegion || literal.Class != 0 && e.program.Classes[literal.Class-1].ProgramRegion
}
func (e *emitter) adoptProgramObject(value string, literal ir.ObjectLiteral) {
	size := fmt.Sprintf("adamic_object_size(%s->shape->count)", value)
	e.adoptProgram(value, size, e.programObjectMember(literal))
}
func (e *emitter) programArray(code string, member bool) string {
	value := e.own(ir.Array, code)
	e.adoptProgram(value, "sizeof *"+value, member)
	return value
}

// A canonical closure must move before it becomes reachable through the cache.
// The ordinary canonical runtime helper is unchanged for counted closures.
func (e *emitter) programCanonical(function int) string {
	name := fmt.Sprintf("adamic_program_canonical_%d", function)
	for _, declaration := range e.declarations {
		if strings.HasPrefix(declaration, "static adamic_closure *"+name+"(") {
			return name
		}
	}
	code := e.functionName(function)
	constructor := "adamic_closure_new"
	match := "held->code == " + code
	if e.program.PackedCountNeeded(function) {
		constructor = "adamic_counted_closure_new"
		match = "held->counted && held->counted_code == " + code
	} else if e.program.ClosureConventionNeeded() {
		match = "!held->counted && " + match
	}
	e.declarations = append(e.declarations, fmt.Sprintf(`static adamic_closure *%s(adamic_cell *identity, size_t count, adamic_cell *const cells[]) {
 adamic_environment *owner = identity->owner;
 for (adamic_closure *held = owner->functions; held != NULL; held = held->canonical_next) {
  if (%s) return adamic_retain(held);
 }
 adamic_closure *closure = %s(%s, count);
 closure = adamic_program_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]);
 for (size_t i = 0; i < count; i++) closure->cells[i] = adamic_retain(cells[i]);
 closure->canonical_owner = owner;
 closure->canonical_next = owner->functions;
 if (owner->functions != NULL) owner->functions->canonical_previous = closure;
 owner->functions = closure;
 return closure;
}`, name, match, constructor, code))
	return name
}

func (e *emitter) programUnique(value string) string {
	if !e.program.ProgramRegion {
		return uniquelyHeld(value)
	}
	return fmt.Sprintf("(!adamic_program_is(%s) && %s)", value, uniquelyHeld(value))
}
