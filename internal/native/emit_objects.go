// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
	"strconv"
	"strings"
)

// fieldSlot is the adamic_value * of an object's field, for object a C name the statement holds. A
// field of a class (class is its constructor plus one) is where the class's layout puts it, the
// layout the constructor's object literal has, for an object of that shape: a compare and a load,
// inline. Any other object is looked up by name through the field cache, as is every other field:
// an object literal can be seen through a class's type, since tsc lets one through and only cohere's
// adamic/nominal-class refuses it, so the shape is checked unless fields.go proves a uniform slot.
func (e *emitter) fieldSlot(object string, name string, class int) string {
	// Uniform offsets describe own storage; static names may instead read live parent data.
	if e.staticFieldName(name) {
		return fmt.Sprintf("adamic_object_field(%s, %s, &%s)", object, cString(name), e.cache())
	}
	if slot := e.uniformFieldSlot(object, name); slot != "" {
		return slot
	}
	lookup := fmt.Sprintf("adamic_object_data_field(%s, %s, &%s)", object, cString(name), e.cache())
	if class == 0 || !cName.MatchString(object) {
		return lookup
	}
	body := e.program.Functions[class-1].Body
	if len(body) == 0 {
		return lookup
	}
	declare, isDeclare := body[0].(ir.Declare)
	if !isDeclare {
		return lookup
	}
	literal, isLiteral := declare.Value.(ir.ObjectLiteral)
	if !isLiteral || literal.Spread != nil {
		return lookup
	}
	for index, field := range literal.Fields {
		if field.Name == name {
			return fmt.Sprintf("(%s->shape == &%s ? &%s->slots[%d] : %s)", object, e.literalShape(literal), object, index, lookup)
		}
	}
	return lookup
}

// staticFieldName is conservative across all constructor layouts, including inherited
// fields. A name in none of them cannot need live-parent reads or own-write flags.
// Class descriptors are emitted only from these IR layouts (class_inheritance.go).
func (e *emitter) staticFieldName(name string) bool {
	for _, layout := range e.program.Classes {
		if !layout.Static {
			continue
		}
		for _, field := range layout.Fields {
			if field.Name == name {
				return true
			}
		}
	}
	return false
}

// writeFieldSlot keeps static own-property bookkeeping on the runtime path. A write
// does not prove an optional field exists, and SetProperty carries no presence proof.
// Uniform offsets therefore need an exact literal-layout guard here; a class fallback
// already has that guard. Unknown, absent and conflicting layouts keep checked lookup.
// Frozen checks and value evaluation remain at the statement. Only C names may repeat.
func (e *emitter) writeFieldSlot(object, name string, class int) string {
	lookup := fmt.Sprintf("adamic_object_write_field(%s, %s, &%s)", object, cString(name), e.cache())
	if !cName.MatchString(object) {
		return lookup
	}
	static := e.staticFieldName(name)
	fallback := lookup
	if !static {
		fallback = fmt.Sprintf("adamic_object_data_field(%s, %s, &%s)", object, cString(name), e.cache())
	}
	data := e.fieldSlot(object, name, class)
	if slot := e.uniformFieldSlot(object, name); slot != "" {
		seen := map[string]bool{}
		checks := []string{}
		walkExpressions(e.program, func(expression ir.Expression) {
			literal, ok := expression.(ir.ObjectLiteral)
			if !ok || literal.Spread != nil {
				return
			}
			for _, field := range literal.Fields {
				if field.Name == name {
					shape := e.literalShape(literal)
					if !seen[shape] {
						seen[shape] = true
						checks = append(checks, fmt.Sprintf("%s->shape == &%s", object, shape))
					}
					break
				}
			}
		})
		data = fallback
		if len(checks) != 0 {
			data = fmt.Sprintf("(%s ? %s : %s)", strings.Join(checks, " || "), slot, fallback)
		}
	}
	if !static {
		return data
	}
	return fmt.Sprintf("(%s->class != NULL && %s->class->is_static ? %s : %s)", object, object, lookup, data)
}

// cName is a C name alone, which a C expression can repeat without evaluating anything twice.
var cName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// objectLiteral makes an object. Its fields' values are evaluated in order first; making the object
// itself can't be observed, so it may come after them.
func (e *emitter) objectLiteral(literal ir.ObjectLiteral) string {
	if literal.Record {
		return e.recordLiteral(literal)
	}
	if reused, ok := e.reused(literal); ok {
		return reused
	}
	if literal.Spread != nil {
		// A copy of the source's object, whatever its shape, with the named fields replaced. The copy
		// is made the moment the spread is evaluated, before any field's value: JavaScript reads the
		// spread's fields first, so a field's expression that writes one of them (a call that sets
		// it) must not show in the result.
		source := e.value(literal.Spread)
		if literal.NoReuse {
			source = e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", source))
		}
		object := e.own(ir.Object, e.spreadCopy(literal, source))
		e.emptySpread(literal, source, object)
		values := make([]string, 0, len(literal.Fields))
		for _, field := range literal.Fields {
			values = append(values, e.value(field.Value))
		}
		for index, field := range literal.Fields {
			slot := e.temporary()
			cache := e.cache()
			e.line("adamic_value *%s = adamic_object_field(%s, %s, &%s);", slot, object, cString(field.Name), cache)
			if e.fieldTypesNeeded() {
				e.line("adamic_object_field_types(%s)[adamic_slot_index(%s, %s)] = %d;", object, object, slot, field.Value.Type())
			}
			if e.fieldReadinessNeeded(field.Name) {
				e.line("adamic_object_initialized(%s)[adamic_slot_index(%s, %s)] = %d;", object, object, slot, map[bool]int{true: 0, false: 1}[field.Uninitialized])
			}
			if field.Value.Type().IsReference() {
				e.line("adamic_release(%s->reference);", slot)
				e.line("%s->reference = %s;", slot, e.kept(values[index]))
			} else {
				e.line("%s->%s = %s;", slot, member(field.Value.Type()), slotted(field.Value.Type(), values[index]))
			}
		}
		return object
	}
	// The literal a fresh function returns is made in the region it was handed, and so are the
	// fresh values its fields are (region.go).
	region := e.regionLiteralDepth != 0 && e.regionLiteralDepth == e.depth
	if literal.Class != 0 && e.inRegion && e.program.Classes[literal.Class-1].Constructor == e.functionIndex {
		region = true
	}
	e.regionLiteralDepth = 0
	values := make([]string, 0, len(literal.Fields))
	for _, field := range literal.Fields {
		if region {
			values = append(values, e.handRegion(field.Value, "region"))
		} else {
			values = append(values, e.value(field.Value))
		}
	}
	object := ""
	if region {
		object = e.regionValue(fmt.Sprintf("adamic_object_new_in(region, &%s)", e.literalShape(literal)))
	} else {
		object = e.own(ir.Object, fmt.Sprintf("adamic_object_new(&%s)", e.literalShape(literal)))
	}
	if len(literal.Fields) > 0 && e.dynamicProperties() {
		e.line("adamic_register_shape_types(&%s_metadata);", e.literalShape(literal))
	}
	if literal.Class != 0 {
		e.line("%s->class = &adamic_class_%d;", object, literal.Class)
		if e.dynamicProperties() {
			class := e.program.Classes[literal.Class-1]
			public := len(class.PublicFields)
			if !class.Literal {
				public = 0
				for _, field := range class.Fields {
					if !field.Private {
						public++
					}
				}
			}
			if public > 0 {
				e.line("adamic_register_shape_types(&%s_metadata);", e.publicClassShape(class))
			}
		}
	}
	for index, field := range literal.Fields {
		if e.fieldTypesNeeded() {
			e.line("adamic_object_field_types(%s)[%d] = %d;", object, index, field.Value.Type())
		}
		if field.Uninitialized {
			e.line("adamic_object_initialized(%s)[%d] = 0;", object, index)
		}
		value := values[index]
		if e.regionValues[value] {
			// A value in the region is immortal while the region lives: held without a count.
			e.line("%s->slots[%d].reference = %s;", object, index, value)
			continue
		}
		if field.Value.Type().IsReference() {
			value = e.kept(value)
		}
		e.line("%s->slots[%d].%s = %s;", object, index, member(field.Value.Type()), slotted(field.Value.Type(), value))
	}
	return object
}

// shape declares an object literal's layout once, at file scope, and names it.
func (e *emitter) shape(fields []ir.Field) string {
	names, types := []string{}, []ir.Type{}
	for _, field := range fields {
		names = append(names, field.Name)
		types = append(types, field.Value.Type())
	}
	return e.shapeOf(names, types)
}

// literalShape is the layout an object literal makes: a class's constructor's has the class's methods
// too, so it's the class's own, never shared with a literal of the same fields.
func (e *emitter) literalShape(literal ir.ObjectLiteral) string {
	if len(literal.Methods) == 0 {
		return e.shape(literal.Fields)
	}
	names, types := []string{}, []ir.Type{}
	for _, field := range literal.Fields {
		names = append(names, field.Name)
		types = append(types, field.Value.Type())
	}
	return e.shapeWith(names, types, literal.Methods)
}

// shapeOf declares a layout by its field names and types.
func (e *emitter) shapeOf(fieldNames []string, fieldTypes []ir.Type) string {
	return e.shapeWith(fieldNames, fieldTypes, nil)
}

// shapeWith declares a layout by its field names and types, and a class's methods, each called
// through a thunk that takes what a call through an interface gives (adamic_method).
func (e *emitter) shapeWith(fieldNames []string, fieldTypes []ir.Type, methods []ir.Method) string {
	names, references, kinds := []string{}, []string{}, []string{}
	for index, name := range fieldNames {
		names = append(names, cString(name))
		kinds = append(kinds, strconv.Itoa(int(fieldTypes[index])))
		references = append(references, strconv.FormatBool(fieldTypes[index].IsReference()))
	}
	fields := fieldNames
	layout := references
	if e.dynamicProperties() {
		layout = kinds
	}
	key := strings.Join(names, ",") + "|" + strings.Join(layout, ",")
	for _, method := range methods {
		key += fmt.Sprintf("|%s=%d", method.Name, method.Function)
	}
	if e.shapes == nil {
		e.shapes = map[string]string{}
	}
	if name, isDeclared := e.shapes[key]; isDeclared {
		return name
	}
	name := fmt.Sprintf("adamic_shape_%d", len(e.shapes))
	e.shapes[key] = name
	table := "NULL"
	methodNames, thunks := []string{}, []string{}
	for _, method := range methods {
		if !e.dispatchable(method.Function) {
			continue
		}
		methodNames = append(methodNames, cString(method.Name))
		thunk := e.methodThunk(method.Function)
		if e.program.ClosureConventionNeeded() {
			field := "code"
			if e.program.PackedCountNeeded(method.Function) {
				field = "counted_code"
			}
			thunk = fmt.Sprintf("{.counted = %t, .%s = %s}", e.program.PackedCountNeeded(method.Function), field, thunk)
		}
		thunks = append(thunks, thunk)
	}
	if len(thunks) > 0 {
		e.declarations = append(e.declarations,
			fmt.Sprintf("static const char *const %s_method_names[] = {%s};", name, strings.Join(methodNames, ", ")),
			fmt.Sprintf("static const %s %s_method_code[] = {%s};", e.methodEntryType(), name, strings.Join(thunks, ", ")),
			fmt.Sprintf("static const adamic_methods %s_methods = {%d, %s_method_names, %s_method_code};", name, len(thunks), name, name))
		table = "&" + name + "_methods"
	}
	if len(fields) == 0 {
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_shape %s = {0, NULL, NULL, %s};", name, table))
	} else {
		e.declarations = append(e.declarations,
			fmt.Sprintf("static const char *const %s_names[] = {%s};", name, strings.Join(names, ", ")),
			fmt.Sprintf("static const bool %s_references[] = {%s};", name, strings.Join(references, ", ")),
			fmt.Sprintf("static const adamic_shape %s = {%d, %s_names, %s_references, %s};", name, len(fields), name, name, table))
	}
	if len(fields) > 0 && e.dynamicProperties() {
		e.declarations = append(e.declarations,
			fmt.Sprintf("static const int %s_types[] = {%s};", name, strings.Join(kinds, ", ")),
			fmt.Sprintf("static adamic_shape_types %s_metadata = {&%s, %s_types, NULL, false};", name, name, name))
	}
	return name
}

// dispatchable reports whether a class's method can be called through an interface: each value it
// takes and gives fits an adamic_value. One that doesn't (a union) can't be
// passed to a function value either (lower's callClosure says not yet), so no call through an
// interface reaches it with one, and it's left out of its class's table.
func (e *emitter) dispatchable(function int) bool {
	method := e.program.Functions[function]
	needed := e.program.StructuralMethodThunks[function]
	if e.program.StructuralMethodThunks == nil {
		walkExpressions(e.program, func(expression ir.Expression) {
			call, ok := expression.(ir.CallClosure)
			if !ok {
				return
			}
			targets, resolved := e.program.FunctionTypeTargets[call.FunctionType]
			if !resolved {
				targets = e.program.ClosureTargets(call).Functions
			}
			for _, target := range targets {
				if e.program.Functions[target].Name == method.Name {
					needed = true
				}
			}
		})
	}
	slotless := func(valueType ir.Type) bool { return valueType == ir.Union || valueType == ir.MaybeBoolean && !needed }
	for index, parameter := range method.Parameters {
		if index > 0 && slotless(e.program.Locals[parameter].Type) {
			return false
		}
	}
	return !slotless(method.Returns)
}

// methodThunk declares, once, the adamic_method that calls a class's method: this from the object,
// each argument unpacked from its adamic_value, and the result packed into one. A closure's arguments
// are its caller's, so one the method takes over (reuse.go) is given a count of its own first; and
// its result, a reference, comes back owned, as a closure's does.
func (e *emitter) methodThunk(function int) string {
	name := fmt.Sprintf("adamic_method_%d", function)
	if e.thunks[function] {
		return name
	}
	if e.thunks == nil {
		e.thunks = map[int]bool{}
	}
	e.thunks[function] = true
	method := e.program.Functions[function]
	count := ""
	if e.program.PackedCountNeeded(function) {
		count = ", size_t argument_count"
	}
	shape := "adamic_method_function"
	if e.program.PackedCountNeeded(function) {
		shape = "adamic_counted_method_function"
	}
	lines := []string{fmt.Sprintf("static %s %s;", shape, name), fmt.Sprintf("static adamic_value %s(adamic_object *self, adamic_value *arguments%s) {", name, count), "\t(void)arguments;"}
	values := []string{}
	for index, parameter := range method.Parameters {
		local := e.program.Locals[parameter]
		value := "self"
		if method.RestElement != 0 && index == len(method.Parameters)-1 {
			slot := e.program.RestArgumentSlots[ir.FunctionRestArguments(method)]
			value = fmt.Sprintf("(adamic_array *)arguments[%d].reference", slot)
		} else if index > 0 {
			value = unslotted(local.Type, fmt.Sprintf("arguments[%d].%s", index-1, member(local.Type)))
			if e.program.PackedCountNeeded(function) {
				value = closureArgument(local.Type, index-1)
			}
			if local.Type.IsReference() {
				value = fmt.Sprintf("(%s)%s", cType(local.Type), value)
			}
			if e.program.PackedCountNeeded(function) && (local.Type.IsMaybe() || local.Type.IsReference()) {
				value = fmt.Sprintf("(argument_count > %d ? %s : %s)", index-1, value, absent(local.Type))
			}
		}
		if local.Type.IsReference() && e.reuse.consumed[parameter] {
			value = fmt.Sprintf("adamic_retain(%s)", value)
		}
		values = append(values, value)
	}
	if method.ArgumentsCount != 0 {
		count := "0"
		if method.ReadsArguments {
			count = "(double)argument_count"
		}
		values = append(values, count)
	}
	call := fmt.Sprintf("%s(%s)", e.functionName(function), strings.Join(values, ", "))
	if method.Returns == 0 {
		lines = append(lines, "\t"+call+";", "\treturn (adamic_value){.number = 0};")
	} else {
		lines = append(lines, fmt.Sprintf("\treturn (adamic_value){.%s = %s};", member(method.Returns), slotted(method.Returns, call)))
	}
	lines = append(lines, "}")
	e.declarations = append(e.declarations, strings.Join(lines, "\n"))
	return name
}

// cache declares a field cache for one place in the program that finds a field.
func (e *emitter) cache() string {
	e.temporaries++
	name := fmt.Sprintf("adamic_cache_%d", e.temporaries)
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_slot_cache %s;", name))
	return name
}

// Programs without reflection keep their original layouts and allocation code.
func (e *emitter) dynamicProperties() bool {
	found := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if call, ok := expression.(ir.ObjectCall); ok && call.Checked {
			found = true
		}
		if _, dynamic := expression.(ir.DynamicProperty); dynamic {
			found = true
		}
	})
	return found
}

func (e *emitter) methodEntryType() string {
	if e.program.ClosureConventionNeeded() {
		return "adamic_method_entry"
	}
	return "adamic_method"
}

func (e *emitter) fieldTypesNeeded() bool {
	if len(e.program.CheckedFields) != 0 || e.dynamicProperties() {
		return true
	}
	needed := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if property, ok := expression.(ir.Property); ok && property.View != "" {
			needed = true
		}
	})
	return needed
}

// Hand-built IR can carry field contracts without the lowerer's summary maps.
func (e *emitter) fieldReadinessNeeded(name string) bool {
	if e.program.UninitializedFields[name] {
		return true
	}
	needed := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if property, ok := expression.(ir.Property); ok && property.Name == name && (property.Readiness != "" || property.View != "") {
			needed = true
		}
	})
	return needed
}
