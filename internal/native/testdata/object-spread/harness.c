// adamic_object_spread_with driven directly (object_spread.c, #5c0z219): no program lowers to it until
// Emit's half of #7nhzq7h, so each mode builds one witness's source and overwrites as C, the values and
// representation bytes the emitter writes for them, and prints what Node prints for the same program
// (oracle.js, its mode of the same name): each result's keys in Object.keys order, every key's value read
// by a run-time key (Reflect.get), an in test where the program has one, and the source after the spread.
// Lines starting "layout:" are native's alone: the result's allocation size and each slot's kind,
// representation byte, initialized byte and write order, which object_spread_test.go holds to its table.
// A mode named refuse_ or for a broken overwrite stops in the primitive, and the test reads its words.
#include "adamic.h"
#include "graph_regions.h"
#include "json_stringify.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *const k_names[] = {"k"};
static const bool number_references[] = {false};
static const adamic_field_kind number_kinds[] = {adamic_field_number};
static const bool reference_references[] = {true};
static const adamic_field_kind reference_kinds[] = {adamic_field_reference};
static const adamic_shape k_number = {1, k_names, number_references, NULL, number_kinds};
static const adamic_shape k_reference = {1, k_names, reference_references, NULL, reference_kinds};
static const adamic_shape k_host = {1, k_names, number_references, &adamic_host_methods, number_kinds};
static const adamic_shape k_groups = {1, k_names, number_references, &adamic_groups_methods, number_kinds};

// width.ts's actual source, wider than its static type.
static const char *const width_names[] = {"a", "k"};
static const bool width_references[] = {false, true};
static const adamic_field_kind width_kinds[] = {adamic_field_number, adamic_field_reference};
static const adamic_shape width_shape = {2, width_names, width_references, NULL, width_kinds};

// p01's source, { kind: 'Holder', value: 1234.5678 }.
static const char *const p01_names[] = {"kind", "value"};
static const bool p01_references[] = {true, false};
static const adamic_field_kind p01_kinds[] = {adamic_field_reference, adamic_field_number};
static const adamic_shape p01_shape = {2, p01_names, p01_references, NULL, p01_kinds};

// p01's reverse, a number over a reference slot: { kind: 'Holder', value: 'v' }.
static const char *const reverse_names[] = {"kind", "value"};
static const bool reverse_references[] = {true, true};
static const adamic_field_kind reverse_kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape reverse_shape = {2, reverse_names, reverse_references, NULL, reverse_kinds};

// A constructed source that wrote a and never wrote b.
static const char *const kept_names[] = {"a", "b"};
static const adamic_shape kept_shape = {2, kept_names, reverse_references, NULL, reverse_kinds};

// moduleNameResolver.ts 245:77's two sources of Resolved: :140's has no originalPath, and :2774's holds it
// in slot 3, written as undefined or never written.
static const char *const resolved_140_names[] = {"path", "extension", "packageId", "resolvedUsingTsExtension"};
static const bool resolved_140_references[] = {true, true, true, false};
static const adamic_field_kind resolved_140_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_number};
static const adamic_shape resolved_140 = {4, resolved_140_names, resolved_140_references, NULL, resolved_140_kinds};
static const char *const resolved_2774_names[] = {"path", "extension", "packageId", "originalPath", "resolvedUsingTsExtension"};
static const bool resolved_2774_references[] = {true, true, true, true, false};
static const adamic_field_kind resolved_2774_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_number};
static const adamic_shape resolved_2774 = {5, resolved_2774_names, resolved_2774_references, NULL, resolved_2774_kinds};

// 279:17's source, a cached resolution: its optional arrays reserved, written or not.
static const char *const cached_names[] = {"resolvedModule", "failedLookupLocations", "affectingLocations", "resolutionDiagnostics"};
static const bool cached_references[] = {true, true, true, true};
static const adamic_field_kind cached_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference};
static const adamic_shape cached_shape = {4, cached_names, cached_references, NULL, cached_kinds};
static const char *const module_names[] = {"resolvedFileName"};
static const adamic_shape module_shape = {1, module_names, reference_references, NULL, reference_kinds};

// A graph or Program member holding another and a label.
static const char *const node_names[] = {"held", "label"};
static const bool node_references[] = {true, true};
static const adamic_field_kind node_kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape node_shape = {2, node_names, node_references, NULL, node_kinds};

// Two shapes of one content whose names are in storage of their own, so nothing but their content is the same.
static char first_k[] = "k";
static char second_k[] = "k";
static const char *const first_names[] = {first_k};
static const char *const second_names[] = {second_k};
static const adamic_shape first_shape = {1, first_names, number_references, NULL, number_kinds};
static const adamic_shape second_shape = {1, second_names, number_references, NULL, number_kinds};

static adamic_value getter(adamic_object *self) {
	(void)self;
	return (adamic_value){.number = 1};
}
// A class with a field k and a getter g on its prototype (oracle.js's Getter), which a spread passes over.
static const adamic_accessor accessors[] = {{"g", getter, NULL, adamic_ir_Number}};
static const adamic_class accessor_class = {.public_shape = &k_number, .accessors = accessors, .accessor_count = 1};
// One whose accessor is named as its public field k, which TypeScript refuses and the spread stops on.
static const adamic_accessor field_accessors[] = {{"k", getter, NULL, adamic_ir_Number}};
static const adamic_class field_accessor_class = {.public_shape = &k_number, .accessors = field_accessors, .accessor_count = 1};

// text is a string made at run time, so the counts see it and ASan sees a read of a freed one.
static adamic_string *text(const char *bytes) {
	size_t length = strlen(bytes);
	adamic_string *made = adamic_string_allocate(length);
	memcpy((char *)made->bytes, bytes, length);
	return made;
}

static adamic_array *strings(size_t count, const char *const *values) {
	adamic_array *array = adamic_array_make(count, adamic_ir_String);
	for (size_t index = 0; index < count; index++) adamic_array_push(array, (adamic_value){.reference = text(values[index])});
	return array;
}

static adamic_value five(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	(void)arguments;
	return (adamic_value){.number = 5};
}

// set_written stores a value in a construction object's slot as a program's write does: its byte, initialized,
// and published in write order.
static void set_written(adamic_object *object, size_t index, adamic_value value, unsigned char representation) {
	object->slots[index] = value;
	adamic_object_field_types(object)[index] = representation;
	adamic_object_initialized(object)[index] = 1;
	adamic_object_publish(object, index);
}

static void print_string(const adamic_string *string) {
	fwrite(string->bytes, 1, string->length, stdout);
}

// show prints a value as oracle.js's show does.
static void show(const adamic_heap *value) {
	if (value == NULL) {
		fputs("undefined", stdout);
		return;
	}
	if (value == &adamic_null) {
		fputs("null", stdout);
		return;
	}
	switch (value->kind) {
	case adamic_kind_number: {
		adamic_string *number = adamic_string_from_number(((const adamic_number_box *)value)->number);
		fputs("number:", stdout);
		print_string(number);
		adamic_release(number);
		return;
	}
	case adamic_kind_boolean:
		fputs(((const adamic_boolean_box *)value)->boolean ? "boolean:true" : "boolean:false", stdout);
		return;
	case adamic_kind_string:
		fputs("string:", stdout);
		print_string((const adamic_string *)value);
		return;
	case adamic_kind_closure: {
		// The one function here is five, whose result is a number.
		adamic_value result = adamic_closure_call((adamic_closure *)value, NULL, 0);
		adamic_string *number = adamic_string_from_number(result.number);
		fputs("function:number:", stdout);
		print_string(number);
		adamic_release(number);
		return;
	}
	case adamic_kind_array: {
		const adamic_array *array = (const adamic_array *)value;
		fputs("array:", stdout);
		for (size_t index = 0; index < array->length; index++) {
			if (index != 0) putchar(',');
			print_string(array->elements[index].reference);
		}
		return;
	}
	default:
		fputs("object", stdout);
	}
}

// print is an object's keys and each key's value read by the key at run time, as Reflect.get reads it.
static void print(const char *label, adamic_object *object) {
	adamic_array *keys = adamic_object_keys(object);
	fputs(label, stdout);
	for (size_t index = 0; index < keys->length; index++) {
		putchar(index == 0 ? ' ' : ',');
		print_string(keys->elements[index].reference);
	}
	putchar('\n');
	for (size_t index = 0; index < keys->length; index++) {
		const adamic_string *key = keys->elements[index].reference;
		adamic_heap *value = adamic_dynamic_property_keyed(&object->heap, key, NULL, 0);
		putchar(' ');
		print_string(key);
		putchar('=');
		show(value);
		putchar('\n');
		adamic_release(value);
	}
	adamic_release(keys);
}

// json prints JSON.stringify of a result, for the modes whose program prints one (with_json). A MaybeBoolean
// slot, p01's and Resolved's resolvedUsingTsExtension, isn't among them: json_objects.c stops on its
// representation whatever made the object ("a field with a two-word representation").
static bool with_json;
static void json(adamic_object *object) {
	static const adamic_json_schema any = {.kind = adamic_json_union};
	adamic_value empty = {.reference = NULL};
	adamic_string *printed = adamic_json_stringify_objects((adamic_value){.reference = object}, &any, empty, NULL, empty, NULL);
	fputs("json ", stdout);
	print_string(printed);
	putchar('\n');
	adamic_release(printed);
}

static void has(const adamic_object *object, const char *name) {
	printf("has %s %s\n", name, adamic_has_property(&object->heap, name) ? "true" : "false");
}

// layout is native's own view of a result: its allocation size, and each slot's kind, representation and
// initialized bytes and write order, "-" for a complete object's and "absent" for a slot never written.
static void layout(const adamic_object *object) {
	printf("layout: size %zu %s\n", adamic_object_allocation_size(object), object->write_order == NULL ? "complete" : "written");
	for (size_t index = 0; index < object->shape->count; index++) {
		printf("layout: %s kind %u byte %u initialized %u order ", object->shape->names[index], (unsigned)object->shape->kinds[index], (unsigned)adamic_object_field_types(object)[index], (unsigned)adamic_object_initialized(object)[index]);
		if (object->write_order == NULL) puts("-");
		else if (object->write_order[index] == SIZE_MAX) puts("absent");
		else printf("%zu\n", object->write_order[index]);
	}
}

static unsigned char *snapshot(const adamic_object *object) {
	size_t size = adamic_object_allocation_size(object);
	unsigned char *copy = malloc(size);
	if (copy == NULL) abort();
	memcpy(copy, object, size);
	return copy;
}

// witness spreads source with the overwrites and prints the result, its in test, its layout, then the
// source, which the spread must leave as it was, byte for byte.
static adamic_object *witness(const adamic_object *source, const adamic_spread_overwrite *overwrites, size_t count, const char *probe) {
	bool held = source != NULL && (const adamic_heap *)source != &adamic_null;
	unsigned char *before = held ? snapshot(source) : NULL;
	adamic_object *result = adamic_object_spread_with(source, overwrites, count);
	print("result", result);
	if (probe != NULL) has(result, probe);
	if (with_json) json(result);
	layout(result);
	if (held) {
		print("source", (adamic_object *)source);
		puts(memcmp(before, source, adamic_object_allocation_size(source)) == 0 ? "source unchanged" : "source changed");
		free(before);
	}
	return result;
}

// kept_source is a constructed source holding a, its b reserved and never written.
static adamic_object *kept_source(void) {
	adamic_object *source = adamic_object_construct(&kept_shape);
	set_written(source, 0, (adamic_value){.reference = text("x")}, adamic_ir_String);
	return source;
}

static const adamic_spread_overwrite appended_c = {"c", adamic_field_number, adamic_ir_Number, {.number = 1}};

static adamic_object *number_source(double number) {
	adamic_object *source = adamic_object_new(&k_number);
	source->slots[0].number = number;
	adamic_object_field_types(source)[0] = adamic_ir_Number;
	return source;
}

static adamic_object *string_source(const char *value) {
	adamic_object *source = adamic_object_new(&k_reference);
	source->slots[0].reference = text(value);
	adamic_object_field_types(source)[0] = adamic_ir_String;
	return source;
}

static adamic_object *resolver_source(bool original, bool written) {
	adamic_object *source = original && !written ? adamic_object_construct(&resolved_2774) : adamic_object_new(original ? &resolved_2774 : &resolved_140);
	size_t last = original ? 4 : 3;
	adamic_maybe_boolean extension = {true, false};
	if (source->write_order != NULL) {
		set_written(source, 0, (adamic_value){.reference = text("/a.ts")}, adamic_ir_String);
		set_written(source, 1, (adamic_value){.reference = text(".ts")}, adamic_ir_String);
		set_written(source, 2, (adamic_value){.reference = NULL}, adamic_ir_UndefinedRepresentation);
		set_written(source, last, (adamic_value){.maybe_boolean = adamic_maybe_boolean_pack(extension)}, adamic_ir_MaybeBoolean);
		return source;
	}
	source->slots[0].reference = text("/a.ts");
	adamic_object_field_types(source)[0] = adamic_ir_String;
	source->slots[1].reference = text(".ts");
	adamic_object_field_types(source)[1] = adamic_ir_String;
	adamic_object_field_types(source)[2] = adamic_ir_UndefinedRepresentation;
	if (original) adamic_object_field_types(source)[3] = adamic_ir_UndefinedRepresentation;
	source->slots[last].maybe_boolean = adamic_maybe_boolean_pack(extension);
	adamic_object_field_types(source)[last] = adamic_ir_MaybeBoolean;
	return source;
}

// resolved is 245:77's literal, { ...resolved, path: resolvedFileName, originalPath }.
static adamic_object *resolved(const adamic_object *source, bool shown) {
	adamic_spread_overwrite overwrites[] = {
		{"path", adamic_field_reference, adamic_ir_String, {.reference = text("/b.ts")}},
		{"originalPath", adamic_field_reference, adamic_ir_String, {.reference = text("/c.ts")}},
	};
	if (shown) return witness(source, overwrites, 2, NULL);
	return adamic_object_spread_with(source, overwrites, 2);
}

static const adamic_spread_overwrite undefined_k = {"k", adamic_field_reference, adamic_ir_UndefinedRepresentation, {.reference = NULL}};

// Seats' five (internal/oracle/testdata/spread_overwrite): { ...source, k: undefined } over a number, an
// optional string, an optional number and a string, the source also undefined or null.
static adamic_object *seat(const char *mode) {
	if (strcmp(mode, "number_source") == 0 || strcmp(mode, "number_present") == 0) return number_source(1);
	if (strcmp(mode, "optional_string_present") == 0 || strcmp(mode, "string_source") == 0) return string_source("one");
	if (strcmp(mode, "optional_string_absent") == 0) return adamic_object_construct(&k_reference);
	if (strcmp(mode, "optional_number_present") == 0) {
		adamic_object *source = adamic_object_new(&k_number);
		source->slots[0].number = adamic_maybe_number_pack((adamic_maybe_number){true, 1});
		adamic_object_field_types(source)[0] = adamic_ir_MaybeNumber;
		return source;
	}
	if (strcmp(mode, "optional_number_absent") == 0) return adamic_object_construct(&k_number);
	if (strcmp(mode, "number_source_null") == 0) return (adamic_object *)(void *)&adamic_null;
	return NULL;
}

static bool seats(const char *mode) {
	static const char *const modes[] = {"number_source", "number_source_undefined", "number_source_null", "number_present", "optional_string_present", "optional_string_absent", "optional_number_present", "optional_number_absent", "string_source", "string_source_undefined"};
	for (size_t index = 0; index < sizeof modes / sizeof modes[0]; index++) {
		if (strcmp(mode, modes[index]) == 0) return true;
	}
	return false;
}

static size_t interned(void) {
#ifdef ADAMIC_COUNT
	return adamic_counted.spread_shapes;
#else
	return 0;
#endif
}

static void counted_shapes(size_t before) {
#ifdef ADAMIC_COUNT
	printf("layout: interned %zu\n", interned() - before);
#else
	(void)before;
#endif
}

// region is graph_source and program_source: a source that's a member of a graph region or of the
// Program region, holding another member and a label, spread into a heap result that outlives the source.
static void region(bool program) {
	adamic_object *n = adamic_object_new(&node_shape);
	adamic_object *s = adamic_object_new(&node_shape);
	if (program) {
		n = adamic_program_adopt_owned(n, adamic_object_allocation_size(n));
		s = adamic_program_adopt_owned(s, adamic_object_allocation_size(s));
	} else {
		n = adamic_graph_adopt(n, adamic_object_allocation_size(n));
		s = adamic_graph_adopt(s, adamic_object_allocation_size(s));
	}
	n->slots[1].reference = text("n");
	adamic_object_field_types(n)[1] = adamic_ir_String;
	adamic_object_field_types(n)[0] = adamic_ir_UndefinedRepresentation;
	s->slots[1].reference = text("s");
	adamic_object_field_types(s)[1] = adamic_ir_String;
	// Held as the region harnesses link a member: hold before the slot is written.
	s->slots[0].reference = program ? n : adamic_graph_hold(s, n);
	adamic_object_field_types(s)[0] = adamic_ir_Object;
	if (!program) adamic_release(n);
	adamic_spread_overwrite count = {"count", adamic_field_number, adamic_ir_Number, {.number = 1}};
	adamic_object *result = witness(s, &count, 1, NULL);
	adamic_slot_cache cache = {0};
	const adamic_object *held = adamic_object_field(result, "held", &cache)->reference;
	adamic_string label = ADAMIC_STRING("label");
	adamic_heap *value = adamic_dynamic_property_keyed((adamic_heap *)&held->heap, &label, NULL, 0);
	fputs("held.label=", stdout);
	show(value);
	putchar('\n');
	adamic_release(value);
	// The source's region goes first; the result's retains keep what it holds alive.
	adamic_release(s);
	value = adamic_dynamic_property_keyed((adamic_heap *)&held->heap, &label, NULL, 0);
	fputs("held.label=", stdout);
	show(value);
	putchar('\n');
	adamic_release(value);
	adamic_release(result);
	if (program) adamic_program_region_end();
}

int main(int argc, char **argv) {
	if (argc != 2) return 2;
	const char *mode = argv[1];
	static const char *const json_modes[] = {"optional_string_absent", "p01_reverse", "absent_kept"};
	for (size_t index = 0; index < sizeof json_modes / sizeof json_modes[0]; index++) {
		if (strcmp(mode, json_modes[index]) == 0) with_json = true;
	}
	if (seats(mode)) {
		adamic_object *source = seat(mode);
		adamic_object *result = witness(source, &undefined_k, 1, "k");
		adamic_release(result);
		if ((const adamic_heap *)source != &adamic_null) adamic_release(source);
	} else if (strcmp(mode, "width") == 0 || strcmp(mode, "width_method") == 0) {
		adamic_object *source = adamic_object_new(&width_shape);
		source->slots[0].number = 1;
		adamic_object_field_types(source)[0] = adamic_ir_Number;
		source->slots[1].reference = text("x");
		adamic_object_field_types(source)[1] = adamic_ir_String;
		adamic_spread_overwrite k = {"k", adamic_field_number, adamic_ir_Number, {.number = 5}};
		if (strcmp(mode, "width_method") == 0) k = (adamic_spread_overwrite){"k", adamic_field_reference, adamic_ir_Closure, {.reference = adamic_closure_new(five, 0)}};
		adamic_release(witness(source, &k, 1, NULL));
		adamic_release(source);
	} else if (strcmp(mode, "prototype_getter") == 0 || strcmp(mode, "prototype_getter_past") == 0) {
		// #vjvcnnj's written_after: { ...new Getter(), g: 21 } keeps the write, and { ...new Getter(), w: 2 }
		// holds no g. Getter's g is its prototype's, so neither copy runs it or holds it as its own.
		adamic_object *source = number_source(1);
		source->class = &accessor_class;
		adamic_spread_overwrite written = {"g", adamic_field_number, adamic_ir_Number, {.number = 21}};
		if (strcmp(mode, "prototype_getter_past") == 0) written = (adamic_spread_overwrite){"w", adamic_field_number, adamic_ir_Number, {.number = 2}};
		adamic_object *result = adamic_object_spread_with(source, &written, 1);
		print("result", result);
		has(result, "g");
		layout(result);
		adamic_release(result);
		adamic_release(source);
	} else if (strcmp(mode, "p01") == 0 || strcmp(mode, "p01_true") == 0) {
		adamic_object *source = adamic_object_new(&p01_shape);
		source->slots[0].reference = text("Holder");
		adamic_object_field_types(source)[0] = adamic_ir_String;
		source->slots[1].number = 1234.5678;
		adamic_object_field_types(source)[1] = adamic_ir_Number;
		adamic_maybe_boolean picked = {strcmp(mode, "p01_true") == 0, true};
		adamic_spread_overwrite value = {"value", adamic_field_number, adamic_ir_MaybeBoolean, {.maybe_boolean = adamic_maybe_boolean_pack(picked)}};
		adamic_release(witness(source, &value, 1, "value"));
		adamic_release(source);
	} else if (strcmp(mode, "p01_reverse") == 0) {
		// Memory's must 2: a number over a reference slot never retains the string it replaces.
		adamic_object *source = adamic_object_new(&reverse_shape);
		source->slots[0].reference = text("Holder");
		adamic_object_field_types(source)[0] = adamic_ir_String;
		source->slots[1].reference = text("v");
		adamic_object_field_types(source)[1] = adamic_ir_String;
		adamic_spread_overwrite value = {"value", adamic_field_number, adamic_ir_Number, {.number = 1234.5678}};
		adamic_release(witness(source, &value, 1, "value"));
		adamic_release(source);
	} else if (strcmp(mode, "absent_kept") == 0) {
		// Memory's must 1: a field the source never wrote stays absent in the result, and a name appended
		// after it comes after the source's own keys.
		adamic_object *source = kept_source();
		adamic_object *result = witness(source, &appended_c, 1, "b");
		adamic_release(result);
		adamic_release(source);
	} else if (strcmp(mode, "resolver_140") == 0 || strcmp(mode, "resolver_2774_undefined") == 0 || strcmp(mode, "resolver_2774_absent") == 0) {
		adamic_object *source = resolver_source(strcmp(mode, "resolver_140") != 0, strcmp(mode, "resolver_2774_undefined") == 0);
		adamic_release(resolved(source, true));
		adamic_release(source);
	} else if (strcmp(mode, "cache_279") == 0) {
		adamic_object *source = adamic_object_construct(&cached_shape);
		adamic_object *module = adamic_object_new(&module_shape);
		module->slots[0].reference = text("/m.ts");
		adamic_object_field_types(module)[0] = adamic_ir_String;
		set_written(source, 0, (adamic_value){.reference = module}, adamic_ir_Object);
		set_written(source, 2, (adamic_value){.reference = strings(1, (const char *const[]){"/x"})}, adamic_ir_Array);
		adamic_spread_overwrite overwrites[] = {
			{"failedLookupLocations", adamic_field_reference, adamic_ir_Array, {.reference = strings(1, (const char *const[]){"/f"})}},
			{"affectingLocations", adamic_field_reference, adamic_ir_Array, {.reference = strings(1, (const char *const[]){"/g"})}},
			{"resolutionDiagnostics", adamic_field_reference, adamic_ir_Array, {.reference = strings(0, NULL)}},
		};
		adamic_release(witness(source, overwrites, 3, NULL));
		adamic_release(source);
	} else if (strcmp(mode, "intern") == 0) {
		// Equal content from two shapes at two addresses is one interned shape, the same address both times.
		size_t before = interned();
		adamic_object *first = adamic_object_new(&first_shape), *second = adamic_object_new(&second_shape);
		first->slots[0].number = second->slots[0].number = 1;
		adamic_object_field_types(first)[0] = adamic_object_field_types(second)[0] = adamic_ir_Number;
		adamic_object *one = adamic_object_spread_with(first, &undefined_k, 1);
		adamic_object *two = adamic_object_spread_with(second, &undefined_k, 1);
		print("result", one);
		print("result", two);
		puts(one->shape == two->shape ? "layout: one shape" : "layout: two shapes");
		counted_shapes(before);
		adamic_release(one);
		adamic_release(two);
		adamic_release(first);
		adamic_release(second);
	} else if (strcmp(mode, "loop") == 0) {
		// Runtime's condition: a thousand spreads over one layout make exactly one shape.
		size_t before = interned();
		adamic_object *source = number_source(1);
		adamic_object *result = NULL;
		for (int index = 0; index < 1000; index++) {
			adamic_release(result);
			result = adamic_object_spread_with(source, &undefined_k, 1);
		}
		print("result", result);
		counted_shapes(before);
		adamic_release(result);
		adamic_release(source);
	} else if (strcmp(mode, "freed_source") == 0) {
		// A shape made at run time, its names in storage of its own, as a decoded object's is, freed with
		// the source before the result is read: the result's shape must own its names.
		char *k = malloc(2), *label = malloc(6);
		const char **names = malloc(2 * sizeof *names);
		adamic_shape *shape = malloc(sizeof *shape);
		if (k == NULL || label == NULL || names == NULL || shape == NULL) abort();
		memcpy(k, "k", 2);
		memcpy(label, "label", 6);
		names[0] = k;
		names[1] = label;
		*shape = (adamic_shape){2, names, width_references, NULL, width_kinds};
		adamic_object *source = adamic_object_new(shape);
		source->slots[0].number = 1;
		adamic_object_field_types(source)[0] = adamic_ir_Number;
		source->slots[1].reference = text("s");
		adamic_object_field_types(source)[1] = adamic_ir_String;
		adamic_spread_overwrite two = {"k", adamic_field_number, adamic_ir_Number, {.number = 2}};
		adamic_object *result = adamic_object_spread_with(source, &two, 1);
		adamic_release(source);
		free(shape);
		free(names);
		free(label);
		free(k);
		print("result", result);
		adamic_release(result);
	} else if (strcmp(mode, "graph_source") == 0) {
		region(false);
	} else if (strcmp(mode, "program_source") == 0) {
		region(true);
	} else if (strcmp(mode, "graph_result_complete") == 0 || strcmp(mode, "graph_result_written") == 0 || strcmp(mode, "graph_result_absent") == 0) {
		// Memory's seam: a result adopted into a graph region moves by its allocation size, its write
		// order with it, and frees with its region. Its keys are read after the move, so a write order
		// left aimed at the freed original reads freed memory.
		bool written = strcmp(mode, "graph_result_written") == 0, absent = strcmp(mode, "graph_result_absent") == 0;
		adamic_object *source = written ? resolver_source(true, false) : absent ? kept_source() : number_source(1);
		adamic_object *result = written ? resolved(source, false) : absent ? adamic_object_spread_with(source, &appended_c, 1) : adamic_object_spread_with(source, &undefined_k, 1);
		result = adamic_graph_adopt(result, adamic_object_allocation_size(result));
		print("result", result);
		layout(result);
		adamic_release(result);
		adamic_release(source);
	} else if (strncmp(mode, "refuse_", 7) == 0) {
		adamic_object *source = NULL;
		if (strcmp(mode, "refuse_record") == 0) source = adamic_record_new(adamic_field_number);
		else if (strcmp(mode, "refuse_host") == 0) source = adamic_object_new(&k_host);
		else if (strcmp(mode, "refuse_groups") == 0) source = adamic_object_new(&k_groups);
		else if (strcmp(mode, "refuse_error") == 0) source = adamic_error_new(text("e"));
		else if (strcmp(mode, "refuse_tuple") == 0) {
			source = adamic_object_new(&k_number);
			source->tuple = true;
		} else if (strcmp(mode, "refuse_accessor_field") == 0) {
			source = adamic_object_new(&k_number);
			source->class = &field_accessor_class;
		} else if (strcmp(mode, "refuse_value") == 0) source = (adamic_object *)(void *)text("s");
		// #k7g6f9h's storages, each one Node copies its own enumerable keys of; a Set is stored as a Map.
		else if (strcmp(mode, "refuse_array") == 0) source = (adamic_object *)(void *)strings(1, (const char *const[]){"a"});
		else if (strcmp(mode, "refuse_typed_array") == 0) source = (adamic_object *)(void *)adamic_typed_array_make(adamic_typed_array_uint8, 2);
		else if (strcmp(mode, "refuse_map") == 0) source = (adamic_object *)(void *)adamic_map_new(true, false);
		else if (strcmp(mode, "refuse_closure") == 0) source = (adamic_object *)(void *)adamic_closure_new(five, 0);
		else return 3;
		adamic_release(adamic_object_spread_with(source, &undefined_k, 1));
		adamic_release(source);
	} else if (strcmp(mode, "wrong_byte") == 0) {
		// Runtime's invariant: a number stated as a string's representation stops in the primitive.
		adamic_object *source = number_source(1);
		adamic_spread_overwrite wrong = {"k", adamic_field_number, adamic_ir_String, {.number = 2}};
		adamic_object *result = adamic_object_spread_with(source, &wrong, 1);
		print("result", result);
		adamic_release(result);
		adamic_release(source);
	} else if (strcmp(mode, "wrong_value") == 0) {
		// A string's representation over an array.
		adamic_object *source = number_source(1);
		adamic_spread_overwrite wrong = {"k", adamic_field_reference, adamic_ir_String, {.reference = strings(0, NULL)}};
		adamic_object *result = adamic_object_spread_with(source, &wrong, 1);
		print("result", result);
		adamic_release(result);
		adamic_release(source);
	} else if (strcmp(mode, "duplicate") == 0) {
		adamic_spread_overwrite twice[] = {undefined_k, undefined_k};
		adamic_release(adamic_object_spread_with(NULL, twice, 2));
	} else {
		return 3;
	}
	// Keep stdout observable when LeakSanitizer exits before stdio flushes it.
	fflush(stdout);
	adamic_heap_end();
	return 0;
}
