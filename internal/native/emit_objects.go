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
func (nativeEmitter *emitter) fieldSlot(object string, name string, class int) string {
	// Uniform offsets describe own storage; static names may instead read live parent data.
	if nativeEmitter.staticFieldName(name) {
		return fmt.Sprintf("adamic_object_field(%s, %s, &%s)", object, cString(name), nativeEmitter.cache())
	}
	if slot := nativeEmitter.uniformFieldSlot(object, name); slot != "" {
		return slot
	}
	lookup := fmt.Sprintf("adamic_object_data_field(%s, %s, &%s)", object, cString(name), nativeEmitter.cache())
	if class == 0 || !cName.MatchString(object) {
		return lookup
	}
	body := nativeEmitter.program.Functions[class-1].Body
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
			return fmt.Sprintf(
				"(%s->shape == &%s ? &%s->slots[%d] : %s)",
				object,
				nativeEmitter.literalShape(literal),
				object,
				index,
				lookup,
			)
		}
	}
	return lookup
}

// staticFieldName is conservative across all constructor layouts, including inherited
// fields. A name in none of them cannot need live-parent reads or own-write flags.
// Class descriptors are emitted only from these IR layouts (class_inheritance.go).
func (nativeEmitter *emitter) staticFieldName(name string) bool {
	for _, layout := range nativeEmitter.program.Classes {
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
func (nativeEmitter *emitter) writeFieldSlot(object, name string, class int) string {
	lookup := fmt.Sprintf("adamic_object_write_field(%s, %s, &%s)", object, cString(name), nativeEmitter.cache())
	if !cName.MatchString(object) {
		return lookup
	}
	static := nativeEmitter.staticFieldName(name)
	fallback := lookup
	if !static {
		fallback = fmt.Sprintf("adamic_object_data_field(%s, %s, &%s)", object, cString(name), nativeEmitter.cache())
	}
	data := nativeEmitter.fieldSlot(object, name, class)
	if slot := nativeEmitter.uniformFieldSlot(object, name); slot != "" {
		seen := map[string]bool{}
		checks := []string{}
		walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
			literal, ok := expression.(ir.ObjectLiteral)
			if !ok || literal.Spread != nil {
				return
			}
			for _, field := range literal.Fields {
				if field.Name == name {
					shape := nativeEmitter.literalShape(literal)
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
func (nativeEmitter *emitter) objectLiteral(literal ir.ObjectLiteral) string {
	if literal.Record {
		return nativeEmitter.recordLiteral(literal)
	}
	if !constructionFields(literal.Fields) {
		if reused, ok := nativeEmitter.reused(literal); ok {
			return reused
		}
	}
	if literal.Spread != nil {
		// A copy of the source's object, whatever its shape, with the named fields replaced. The copy
		// is made the moment the spread is evaluated, before any field's value: JavaScript reads the
		// spread's fields first, so a field's expression that writes one of them (a call that sets
		// it) must not show in the result.
		source := nativeEmitter.value(literal.Spread)
		if literal.NoReuse {
			source = nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", source))
		}
		object := nativeEmitter.own(ir.Object, nativeEmitter.spreadCopy(literal, source))
		if nativeEmitter.program.ObjectSpreadMayThrow(literal) {
			nativeEmitter.checkThrown()
		}
		nativeEmitter.adoptGraphObject(object, literal)
		nativeEmitter.programObject(object, literal.ProgramRegion)
		nativeEmitter.emptySpread(literal, source, object)
		values := make([]string, 0, len(literal.Fields))
		for _, field := range literal.Fields {
			values = append(values, nativeEmitter.value(field.Value))
		}
		for index, field := range literal.Fields {
			slot := nativeEmitter.temporary()
			cache := nativeEmitter.cache()
			nativeEmitter.line(
				"adamic_value *%s = adamic_object_field(%s, %s, &%s);",
				slot,
				object,
				cString(field.Name),
				cache,
			)
			if nativeEmitter.fieldTypesNeeded() {
				nativeEmitter.line(
					"adamic_object_field_types(%s)[adamic_slot_index(%s, %s)] = %d;",
					object,
					object,
					slot,
					fieldInitialRepresentation(field),
				)
			}
			if nativeEmitter.fieldReadinessNeeded(field.Name) {
				nativeEmitter.line(
					"adamic_object_initialized(%s)[adamic_slot_index(%s, %s)] = %d;",
					object,
					object,
					slot,
					map[bool]int{true: 0, false: 1}[field.Uninitialized && !field.Unset],
				)
			}
			if field.Value.Type().IsReference() {
				nativeEmitter.dropIn(object, slot+"->reference")
				nativeEmitter.line("%s->reference = %s;", slot, nativeEmitter.keptIn(object, values[index]))
			} else {
				nativeEmitter.line(
					"%s->%s = %s;",
					slot,
					member(field.Value.Type()),
					slotted(field.Value.Type(), values[index]),
				)
			}
		}
		return object
	}
	// The literal a fresh function returns is made in the region it was handed, and so are the
	// fresh values its fields are (region.go).
	region := !literal.ProgramRegion &&
		nativeEmitter.regionLiteralDepth != 0 &&
		nativeEmitter.regionLiteralDepth == nativeEmitter.depth
	if !literal.ProgramRegion &&
		literal.Class != 0 &&
		nativeEmitter.inRegion &&
		nativeEmitter.program.Classes[literal.Class-1].Constructor == nativeEmitter.functionIndex {
		region = true
	}
	nativeEmitter.regionLiteralDepth = 0
	values := make([]string, 0, len(literal.Fields))
	for _, field := range literal.Fields {
		if region {
			values = append(values, nativeEmitter.handRegion(field.Value, "region"))
		} else {
			if field.Absent {
				values = append(values, "")
				continue
			}
			values = append(values, nativeEmitter.value(field.Value))
		}
	}
	object := ""
	if constructionFields(literal.Fields) {
		object = nativeEmitter.own(
			ir.Object,
			fmt.Sprintf("adamic_object_construct(&%s)", nativeEmitter.literalShape(literal)),
		)
	} else if region {
		allocator := "adamic_object_new_filled_in"
		if literal.Class != 0 {
			// Constructor writes can call or throw after allocation, so its untouched slots need zero.
			allocator = "adamic_object_new_in"
		}
		object = nativeEmitter.regionValue(
			fmt.Sprintf("%s(region, &%s)", allocator, nativeEmitter.literalShape(literal)),
		)
	} else {
		object = nativeEmitter.own(
			ir.Object,
			fmt.Sprintf("adamic_object_new(&%s)", nativeEmitter.literalShape(literal)),
		)
		nativeEmitter.adoptGraphObject(object, literal)
	}
	nativeEmitter.programObject(object, literal.ProgramRegion)
	if literal.Tuple {
		nativeEmitter.line("%s->tuple = true;", object)
	}
	if len(literal.Fields) > 0 && nativeEmitter.dynamicProperties() {
		nativeEmitter.line("adamic_register_shape_types(&%s_metadata);", nativeEmitter.literalShape(literal))
	}
	if literal.Class != 0 {
		nativeEmitter.line("%s->class = &%s;", object, nativeEmitter.className(literal.Class))
		if nativeEmitter.dynamicProperties() {
			class := nativeEmitter.program.Classes[literal.Class-1]
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
				nativeEmitter.line("adamic_register_shape_types(&%s_metadata);", nativeEmitter.publicClassShape(class))
			}
		}
	}
	for index, field := range literal.Fields {
		if nativeEmitter.fieldTypesNeeded() {
			nativeEmitter.line(
				"adamic_object_field_types(%s)[%d] = %d;",
				object,
				index,
				fieldInitialRepresentation(field),
			)
		}
		if field.Absent {
			continue
		}
		if constructionFields(literal.Fields) {
			nativeEmitter.line(
				"adamic_object_publish(%s, %d); adamic_object_initialized(%s)[%d] = 1;",
				object,
				index,
				object,
				index,
			)
		}
		if field.Uninitialized && !field.Unset {
			nativeEmitter.line("adamic_object_initialized(%s)[%d] = 0;", object, index)
		}
		value := values[index]
		if nativeEmitter.regionValues[value] {
			// A value in the region is immortal while the region lives: held without a count.
			nativeEmitter.line("%s->slots[%d].reference = %s;", object, index, value)
			continue
		}
		if field.Value.Type().IsReference() {
			if region && !constantUndefined.MatchString(value) {
				// Report both retained and moved heap references. Constants pass over at runtime.
				nativeEmitter.line("adamic_region_hold(region, %s);", value)
			}
			value = nativeEmitter.keptIn(object, value)
		}
		nativeEmitter.line(
			"%s->slots[%d].%s = %s;",
			object,
			index,
			member(field.Value.Type()),
			slotted(field.Value.Type(), value),
		)
	}
	return object
}

// shape declares an object literal's layout once, at file scope, and names it.
func (nativeEmitter *emitter) shape(fields []ir.Field) string {
	names, types := []string{}, []ir.Type{}
	for _, field := range fields {
		names = append(names, nativeEmitter.printedFieldName(field))
		types = append(types, field.Value.Type())
	}
	return nativeEmitter.shapeOf(names, types)
}

// literalShape is the layout an object literal makes: a class's constructor's has the class's methods
// too, so it's the class's own, never shared with a literal of the same fields.
func (nativeEmitter *emitter) literalShape(literal ir.ObjectLiteral) string {
	if len(literal.Methods) == 0 {
		return nativeEmitter.shape(literal.Fields)
	}
	names, types := []string{}, []ir.Type{}
	for _, field := range literal.Fields {
		names = append(names, nativeEmitter.printedFieldName(field))
		types = append(types, field.Value.Type())
	}
	return nativeEmitter.shapeWith(names, types, literal.Methods)
}

// shapeOf declares a layout by its field names and types.
func (nativeEmitter *emitter) shapeOf(fieldNames []string, fieldTypes []ir.Type) string {
	return nativeEmitter.shapeWith(fieldNames, fieldTypes, nil)
}

// shapeWith declares a layout by its field names and types, and a class's methods, each called
// through a thunk that takes what a call through an interface gives (adamic_method).
func (nativeEmitter *emitter) shapeWith(fieldNames []string, fieldTypes []ir.Type, methods []ir.Method) string {
	names, references, kinds, storage := []string{}, []string{}, []string{}, []string{}
	for index, name := range fieldNames {
		names = append(names, cString(name))
		kinds = append(kinds, strconv.Itoa(int(fieldTypes[index])))
		references = append(references, strconv.FormatBool(fieldTypes[index].IsReference()))
		storage = append(storage, fieldStorageKind(fieldTypes[index]))
	}
	fields := fieldNames
	layout := storage
	if nativeEmitter.dynamicProperties() {
		layout = append(append([]string{}, storage...), kinds...)
	}
	key := strings.Join(names, ",") + "|" + strings.Join(layout, ",")
	for _, method := range methods {
		key += fmt.Sprintf("|%s=%s", method.Name, nativeEmitter.functionName(method.Function))
	}
	if nativeEmitter.shapes == nil {
		nativeEmitter.shapes = map[string]string{}
	}
	if name, isDeclared := nativeEmitter.shapes[key]; isDeclared {
		return name
	}
	name := stableName("adamic_shape", "layout", key)
	nativeEmitter.shapes[key] = name
	table := "NULL"
	methodNames, thunks := []string{}, []string{}
	for _, method := range methods {
		if !nativeEmitter.dispatchable(method.Function) {
			continue
		}
		methodNames = append(methodNames, cString(method.Name))
		thunk := nativeEmitter.methodThunk(method.Function)
		if nativeEmitter.program.ClosureConventionNeeded() {
			field := "code"
			if nativeEmitter.program.PackedCountNeeded(method.Function) {
				field = "counted_code"
			}
			thunk = fmt.Sprintf(
				"{.counted = %t, .%s = %s}",
				nativeEmitter.program.PackedCountNeeded(method.Function),
				field,
				thunk,
			)
		}
		thunks = append(thunks, thunk)
	}
	if len(thunks) > 0 {
		nativeEmitter.declarations = append(nativeEmitter.declarations,
			fmt.Sprintf("static const char *const %s_method_names[] = {%s};", name, strings.Join(methodNames, ", ")),
			fmt.Sprintf(
				"static const %s %s_method_code[] = {%s};",
				nativeEmitter.methodEntryType(),
				name,
				strings.Join(thunks, ", "),
			),
			fmt.Sprintf(
				"static const adamic_methods %s_methods = {%d, %s_method_names, %s_method_code};",
				name,
				len(thunks),
				name,
				name,
			))
		table = "&" + name + "_methods"
	}
	if len(fields) == 0 {
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static const adamic_shape %s = {0, NULL, NULL, %s, NULL};", name, table),
		)
	} else {
		nativeEmitter.declarations = append(nativeEmitter.declarations,
			fmt.Sprintf("static const char *const %s_names[] = {%s};", name, strings.Join(names, ", ")),
			fmt.Sprintf("static const bool %s_references[] = {%s};", name, strings.Join(references, ", ")),
			fmt.Sprintf("static const adamic_field_kind %s_kinds[] = {%s};", name, strings.Join(storage, ", ")),
			fmt.Sprintf(
				"static const adamic_shape %s = {%d, %s_names, %s_references, %s, %s_kinds};",
				name,
				len(fields),
				name,
				name,
				table,
				name,
			))
	}
	if len(fields) > 0 && nativeEmitter.dynamicProperties() {
		nativeEmitter.declarations = append(nativeEmitter.declarations,
			fmt.Sprintf("static const int %s_types[] = {%s};", name, strings.Join(kinds, ", ")),
			fmt.Sprintf("static adamic_shape_types %s_metadata = {&%s, %s_types, NULL, false};", name, name, name))
	}
	return name
}

// dispatchable reports whether a class's method can be called through an interface: each value it
// takes and gives fits an adamic_value. One that doesn't (a union) can't be
// passed to a function value either (lower's callClosure says not yet), so no call through an
// interface reaches it with one, and it's left out of its class's table.
func (nativeEmitter *emitter) dispatchable(function int) bool {
	method := nativeEmitter.program.Functions[function]
	needed := nativeEmitter.program.StructuralMethodThunks[function]
	if nativeEmitter.program.StructuralMethodThunks == nil {
		walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
			call, ok := expression.(ir.CallClosure)
			if !ok {
				return
			}
			targets, resolved := nativeEmitter.program.FunctionTypeTargets[call.FunctionType]
			if !resolved {
				targets = nativeEmitter.program.ClosureTargets(call).Functions
			}
			for _, target := range targets {
				if nativeEmitter.program.Functions[target].Name == method.Name {
					needed = true
				}
			}
		})
	}
	slotless := func(valueType ir.Type) bool { return valueType == ir.Union || valueType == ir.MaybeBoolean && !needed }
	for index, parameter := range method.Parameters {
		if index > 0 && slotless(nativeEmitter.program.Locals[parameter].Type) {
			return false
		}
	}
	return !slotless(method.Returns)
}

// methodThunk declares, once, the adamic_method that calls a class's method: this from the object,
// each argument unpacked from its adamic_value, and the result packed into one. A closure's arguments
// are its caller's, so one the method takes over (reuse.go) is given a count of its own first; and
// its result, a reference, comes back owned, as a closure's does.
func (nativeEmitter *emitter) methodThunk(function int) string {
	name := stableName(
		"adamic_method",
		nativeEmitter.sourceReadable(function),
		nativeEmitter.functionKey(function, map[int]bool{}),
	)
	if nativeEmitter.thunks[name] {
		return name
	}
	if nativeEmitter.thunks == nil {
		nativeEmitter.thunks = map[string]bool{}
	}
	nativeEmitter.thunks[name] = true
	method := nativeEmitter.program.Functions[function]
	count := ""
	if nativeEmitter.program.PackedCountNeeded(function) {
		count = ", size_t argument_count"
	}
	shape := "adamic_method_function"
	if nativeEmitter.program.PackedCountNeeded(function) {
		shape = "adamic_counted_method_function"
	}
	lines := []string{
		fmt.Sprintf("static %s %s;", shape, name),
		fmt.Sprintf("static adamic_value %s(adamic_object *self, adamic_value *arguments%s) {", name, count),
		"\t(void)arguments;",
	}
	values := []string{}
	for index, parameter := range method.Parameters {
		local := nativeEmitter.program.Locals[parameter]
		value := "self"
		if method.RestElement != 0 && index == len(method.Parameters)-1 {
			slot := nativeEmitter.program.RestArgumentSlots[ir.FunctionRestArguments(method)]
			value = fmt.Sprintf("(adamic_array *)arguments[%d].reference", slot)
		} else if index > 0 {
			value = unslotted(local.Type, fmt.Sprintf("arguments[%d].%s", index-1, member(local.Type)))
			if nativeEmitter.program.PackedCountNeeded(function) {
				value = closureArgument(local.Type, index-1)
			}
			if local.Type.IsReference() {
				value = fmt.Sprintf("(%s)%s", cType(local.Type), value)
			}
			if nativeEmitter.program.PackedCountNeeded(function) && (local.Type.IsMaybe() || local.Type.IsReference()) {
				value = fmt.Sprintf("(argument_count > %d ? %s : %s)", index-1, value, absent(local.Type))
			}
		}
		if local.Type.IsReference() && nativeEmitter.reuse.consumed[parameter] {
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
	call := fmt.Sprintf("%s(%s)", nativeEmitter.functionName(function), strings.Join(values, ", "))
	if method.Returns == 0 {
		lines = append(lines, "\t"+call+";", "\treturn (adamic_value){.number = 0};")
	} else {
		lines = append(
			lines,
			fmt.Sprintf("\treturn (adamic_value){.%s = %s};", member(method.Returns), slotted(method.Returns, call)),
		)
	}
	lines = append(lines, "}")
	nativeEmitter.declarations = append(nativeEmitter.declarations, strings.Join(lines, "\n"))
	return name
}

// cache declares a field cache for one place in the program that finds a field.
func (nativeEmitter *emitter) cache() string {
	nativeEmitter.caches++
	owner := "main"
	if nativeEmitter.mainModule != "" {
		owner = nativeEmitter.mainModule
	}
	if nativeEmitter.function != nil {
		owner = nativeEmitter.functionName(nativeEmitter.functionIndex)
		if nativeEmitter.inRegion {
			owner = nativeEmitter.regionFunctionName(nativeEmitter.functionIndex)
		}
	}
	name := stableName("adamic_cache", "slot", owner) + fmt.Sprintf("_%d", nativeEmitter.caches)
	if nativeEmitter.generatedDeclarations == nil {
		nativeEmitter.generatedDeclarations = map[string]bool{}
	}
	if !nativeEmitter.generatedDeclarations[name] {
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static adamic_slot_cache %s;", name),
		)
		nativeEmitter.generatedDeclarations[name] = true
	}
	return name
}

// Metadata queries share one walk. Lowering is complete before an emitter is made,
// and emission does not change the expressions or field contracts these facts use.
// A new emitter recomputes them even when a caller edits and re-emits the same IR.
type objectMetadataNeeds struct {
	dynamic    bool
	fieldTypes bool

	// keyed is whether the program reads by a run-time key (keyedClassDeclarations).
	keyed bool
}

func (nativeEmitter *emitter) prepareObjectMetadata() {
	if nativeEmitter.objectMetadata != nil {
		return
	}
	needs := &objectMetadataNeeds{
		fieldTypes: len(nativeEmitter.program.CheckedFields) != 0 || nativeEmitter.constructionNeeded(),
	}
	walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
		switch expression := expression.(type) {
		case ir.ObjectCall:
			needs.dynamic = needs.dynamic || expression.Checked
		case ir.CheckedCast:
			needs.fieldTypes = needs.fieldTypes || expression.CheckedFields
		case ir.DynamicProperty:
			needs.dynamic = true
		case ir.DynamicKeyedProperty:
			needs.dynamic, needs.keyed = true, true
		case ir.Property:
			needs.fieldTypes = needs.fieldTypes || expression.View != ""
		}
	})
	needs.fieldTypes = needs.fieldTypes || needs.dynamic
	nativeEmitter.objectMetadata = needs
}

// Programs without reflection keep their original layouts and allocation code.
func (nativeEmitter *emitter) dynamicProperties() bool {
	nativeEmitter.prepareObjectMetadata()
	return nativeEmitter.objectMetadata.dynamic
}

func (nativeEmitter *emitter) keyedReads() bool {
	nativeEmitter.prepareObjectMetadata()
	return nativeEmitter.objectMetadata.keyed
}

func (nativeEmitter *emitter) methodEntryType() string {
	if nativeEmitter.program.ClosureConventionNeeded() {
		return "adamic_method_entry"
	}
	return "adamic_method"
}

func (nativeEmitter *emitter) fieldTypesNeeded() bool {
	nativeEmitter.prepareObjectMetadata()
	return nativeEmitter.objectMetadata.fieldTypes
}

// Hand-built IR can carry field contracts without the lowerer's summary maps.
func (nativeEmitter *emitter) fieldReadinessNeeded(name string) bool {
	if nativeEmitter.program.UninitializedFields[name] {
		return true
	}
	needed := false
	walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
		if property, ok := expression.(ir.Property); ok &&
			property.Name == name &&
			(property.Readiness != "" || property.View != "") {
			needed = true
		}
	})
	return needed
}

// Preserve null/undefined semantic tags before a checked read normalizes the
// physical slot. Uninitialized boxed slots retain their original layout tag.
func fieldRepresentation(value ir.Expression) int {
	switch value.(type) {
	case ir.Null:
		return int(ir.NullRepresentation)
	case ir.Undefined:
		if value.Type().IsReference() {
			return int(ir.UndefinedRepresentation)
		}
	}
	return int(value.Type())
}
func fieldInitialRepresentation(field ir.Field) int {
	if field.Uninitialized && field.Value.Type() == ir.Union {
		return int(ir.Union)
	}
	return fieldRepresentation(field.Value)
}

// Optional scalars and implementation words use the numeric slot representation.
func fieldStorageKind(of ir.Type) string {
	if of.IsReference() {
		return "adamic_field_reference"
	}
	if of == ir.Boolean {
		return "adamic_field_boolean"
	}
	return "adamic_field_number"
}
