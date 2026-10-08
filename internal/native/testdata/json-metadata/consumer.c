// Test-only contract consumer. It never links or calls the library JSON encoder.
// Its small writer covers this fixture's ASCII strings and number spelling.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
static const adamic_class hook_class = {.public_shape = &HOOK_SHAPE};
static const adamic_class returned_class = {.public_shape = &RETURNED_SHAPE};
static const adamic_class outer_class = {.public_shape = &OUTER_SHAPE};
static adamic_array *keys;
static adamic_array *reentrant_container;
static adamic_object *simple;
static adamic_array *numbers;
static adamic_string text = ADAMIC_STRING("str");
static adamic_value parallel_number(adamic_closure *self, adamic_value *args) {
	(void)self;
	return (adamic_value){.number = args[0].number + 1};
}
static adamic_value parallel_boolean(adamic_closure *self, adamic_value *args) {
	(void)self;
	return (adamic_value){.boolean = args[0].number > 1};
}
static adamic_value parallel_string(adamic_closure *self, adamic_value *args) {
	(void)self;
	(void)args;
	return (adamic_value){.reference = &text};
}
static adamic_value noop(adamic_closure *self, adamic_value *args) {
	(void)self;
	(void)args;
	return (adamic_value){0};
}
static adamic_heap *adamic_function_0_hook(adamic_object *self, adamic_string *key) {
	adamic_array_push(keys, (adamic_value){.reference = adamic_retain(key)});
	if (self->slots[0].number == 99) {
		adamic_thrown = adamic_error_new(&text);
		return NULL;
	}
	switch ((int)self->slots[0].number) {
	case 0:
		return adamic_box_number(7);
	case 1:
		return (adamic_heap *)&adamic_box_false;
	case 2:
		return (adamic_heap *)&text;
	case 3:
		return &adamic_null;
	case 4:
		return NULL;
	case 5:
		return (adamic_heap *)adamic_closure_new(noop, 0);
	case 6:
		return (adamic_heap *)adamic_map_new(false, false);
	case 7:
		return adamic_retain(numbers);
	default:
		return adamic_retain(simple);
	}
}
static double adamic_function_1_returned(adamic_object *self) {
	(void)self;
	adamic_panic("second hook", 11);
}
static adamic_object *adamic_function_2_outer(adamic_object *self) {
	(void)self;
	adamic_object *o = adamic_object_new(&RETURNED_SHAPE);
	o->class = &returned_class;
	o->slots[0].number = 7;
	return o;
}
static double adamic_function_3_reentrant(adamic_object *self) {
	adamic_array_remove_typed(reentrant_container, 0, 1, true, 0, NULL, &adamic_json_union_schema);
	return self->slots[0].number;
}
static void quote(const adamic_string *s) {
	putchar('"');
	for (size_t i = 0; i < s->length; i++) {
		unsigned char c = (unsigned char)s->bytes[i];
		if (c == '"' || c == '\\') {
			putchar('\\');
			putchar(c);
		} else if (c < 32)
			printf("\\u%04x", c);
		else
			putchar(c);
	}
	putchar('"');
}
static bool reference_kind(const adamic_json_schema *s) {
	return s->kind == adamic_json_union || s->kind == adamic_json_string || s->kind == adamic_json_map || s->kind == adamic_json_function || s->kind == adamic_json_array || s->kind == adamic_json_object || s->kind == adamic_json_toJSON;
}
typedef struct prepared {
	adamic_json_result original, resolved;
	bool owned;
} prepared;
static prepared prepare(adamic_json_result r, const adamic_string *key) {
	prepared p = {.original = r, .resolved = adamic_json_resolve(r.value, r.schema), .owned = false};
	if (p.resolved.schema == NULL)
		adamic_panic("missing descriptor", 18);
	if (p.resolved.schema->kind == adamic_json_toJSON) {
		p.original = adamic_json_runtime_providers.to_json(p.resolved.value.reference, key);
		p.owned = true;
		if (adamic_thrown != NULL)
			return p;
		p.resolved = adamic_json_resolve(p.original.value, p.original.schema);
		if (p.resolved.schema == NULL)
			adamic_panic("missing hook descriptor", 23);
	}
	return p;
}
static void cleanup(prepared p) {
	if (p.owned && reference_kind(p.original.schema))
		adamic_release(p.original.value.reference);
}
static bool omitted(prepared p) { return p.resolved.schema->kind == adamic_json_undefined || p.resolved.schema->kind == adamic_json_function; }
static void write_prepared(prepared p);
static void write_fields(const adamic_object *o, const adamic_json_schema *schema, bool tuple) {
	putchar(tuple ? '[' : '{');
	size_t written = 0;
	for (size_t i = 0; i < schema->count; i++) {
		const adamic_json_field *f = &schema->fields[i];
		prepared child = prepare((adamic_json_result){o->slots[f->slot], f->schema}, f->name);
		if (!tuple && omitted(child)) {
			cleanup(child);
			continue;
		}
		if (written++)
			putchar(',');
		if (!tuple) {
			quote(f->name);
			putchar(':');
		}
		if (omitted(child))
			fputs("null", stdout);
		else
			write_prepared(child);
		cleanup(child);
	}
	putchar(tuple ? ']' : '}');
}
static void write_prepared(prepared p) {
	adamic_value v = p.resolved.value;
	const adamic_json_schema *s = p.resolved.schema;
	switch (s->kind) {
	case adamic_json_null:
		fputs("null", stdout);
		break;
	case adamic_json_number: {
		char buffer[ADAMIC_NUMBER_FORMAT_MAX];
		if (!isfinite(v.number))
			fputs("null", stdout);
		else {
			size_t n = adamic_number_format(v.number, buffer);
			fwrite(buffer, 1, n, stdout);
		}
		break;
	}
	case adamic_json_boolean:
		fputs(v.boolean ? "true" : "false", stdout);
		break;
	case adamic_json_string:
		quote(v.reference);
		break;
	case adamic_json_map:
		fputs("{}", stdout);
		break;
	case adamic_json_object:
	case adamic_json_toJSON:
	case adamic_json_tuple:
		write_fields(v.reference, s, s->kind == adamic_json_tuple);
		break;
	case adamic_json_array: {
		adamic_array *a = v.reference;
		putchar('[');
		for (size_t i = 0; i < a->length; i++) {
			if (i)
				putchar(',');
			char index[32];
			size_t n = (size_t)snprintf(index, sizeof index, "%zu", i);
			adamic_string *key = adamic_string_allocate(n);
			memcpy((char *)key->bytes, index, n);
			prepared child = prepare(adamic_json_runtime_providers.array_element(a, i), key);
			if (omitted(child))
				fputs("null", stdout);
			else
				write_prepared(child);
			cleanup(child);
			adamic_release(key);
		}
		putchar(']');
		break;
	}
	default:
		adamic_panic("unsupported test descriptor", 27);
	}
}
static void show(void *v) {
	static adamic_string root = ADAMIC_STRING("");
	prepared p = prepare((adamic_json_result){{.reference = v}, &adamic_json_union_schema}, &root);
	if (omitted(p))
		fputs("undefined", stdout);
	else
		write_prepared(p);
	cleanup(p);
	putchar('\n');
}
static adamic_string *str(const char *s) {
	size_t n = strlen(s);
	adamic_string *r = adamic_string_allocate(n);
	memcpy((char *)r->bytes, s, n);
	return r;
}
static adamic_array *pair_numbers(void) {
	adamic_array *a = adamic_array_new_typed(2, false, &adamic_json_number_schema);
	adamic_array_push(a, (adamic_value){.number = 1});
	adamic_array_push(a, (adamic_value){.number = 2});
	return a;
}
static int compare(adamic_value left, adamic_value right, void *ctx) {
	(void)ctx;
	(void)left;
	(void)right;
	return -1;
}
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	if (argc > 1) {
		if (strcmp(argv[1], "throw") == 0) {
			static adamic_string root = ADAMIC_STRING("");
			keys = adamic_array_new_typed(1, true, &adamic_json_string_schema);
			adamic_object *hook = adamic_object_new(&HOOK_SHAPE);
			hook->slots[0].number = 99;
			adamic_json_result result = adamic_json_runtime_providers.to_json(hook, &root);
			if (adamic_thrown == NULL || result.schema != &adamic_json_undefined_schema || result.value.reference != NULL || keys->length != 1)
				return 1;
			adamic_release(adamic_thrown);
			adamic_thrown = NULL;
			adamic_release(hook);
			adamic_release(keys);
			puts("hook threw once");
			return 0;
		}
		adamic_array *untyped = adamic_array_new(1, false);
		if (strcmp(argv[1], "empty") != 0)
			adamic_array_push(untyped, (adamic_value){.reference = (void *)1});
		(void)adamic_json_runtime_providers.array_element(untyped, 0);
		return 1;
	}
	keys = adamic_array_new_typed(32, true, &adamic_json_string_schema);
	numbers = pair_numbers();
	adamic_object *plugin = adamic_object_new(&PLUGIN_SHAPE);
	plugin->slots[0].reference = str("x");
	adamic_array *plugins = adamic_array_new_typed(1, true, &adamic_json_union_schema);
	adamic_array_push(plugins, (adamic_value){.reference = plugin});
	adamic_object *options = adamic_object_new(&OPTIONS_SHAPE);
	options->slots[0].boolean = true;
	options->slots[1].number = 1;
	options->slots[2].reference = plugins;
	options->slots[3].reference = str("kept");
	adamic_object *ref = adamic_object_new(&REF_SHAPE);
	ref->slots[0].reference = str("../a");
	ref->slots[1].boolean = true;
	ref->slots[2].boolean = false;
	adamic_array *refs = adamic_array_new_typed(1, true, &adamic_json_union_schema);
	adamic_array_push(refs, (adamic_value){.reference = ref});
	static adamic_string files = ADAMIC_STRING("a.ts"), comma = ADAMIC_STRING(",");
	adamic_object *info = adamic_object_new(&INFO_SHAPE);
	info->slots[0].reference = str("5.8.0");
	info->slots[1].reference = adamic_retain(numbers);
	info->slots[2].reference = adamic_string_split(&files, &comma);
	info->slots[3].reference = adamic_retain(options);
	info->slots[4].reference = str("kept");
	simple = adamic_object_new(&SIMPLE_SHAPE);
	simple->slots[0].number = 7;
	adamic_map *map = adamic_map_new(true, false);
	adamic_map_set(map, (adamic_value){.reference = str("x")}, (adamic_value){.number = 1});
	show(info);
	show(options);
	show(refs);
	show(numbers);
	show(simple);
	show(map);
	adamic_array *filled = adamic_array_filled_typed(3, (adamic_value){.number = 7}, false, &adamic_json_number_schema);
	adamic_object *flags = adamic_object_new(&FLAGS_SHAPE);
	flags->slots[0].boolean = true;
	flags->slots[1].boolean = false;
	adamic_array *booleans = adamic_object_values(flags, false, false);
	static adamic_string abc = ADAMIC_STRING("a,b,c");
	adamic_array *strings = adamic_string_split(&abc, &comma);
	show(filled);
	show(booleans);
	show(strings);
	adamic_array *slice = adamic_array_slice(filled, 1, 0, false);
	show(slice);
	adamic_release(slice);
	adamic_array *concat = adamic_array_concat(2, (adamic_array *[]){booleans, booleans});
	show(concat);
	adamic_release(concat);
	slice = adamic_array_slice(strings, 0, 0, false);
	show(adamic_array_reverse(slice));
	adamic_release(slice);
	for (int mode = 0; mode < 9; mode++) {
		adamic_object *hook = adamic_object_new(&HOOK_SHAPE);
		hook->class = &hook_class;
		hook->slots[0].number = mode;
		show(hook);
		adamic_object *property = adamic_object_new(&PROPERTY_SHAPE);
		property->slots[0].reference = adamic_retain(hook);
		show(property);
		adamic_release(property);
		adamic_array *array = adamic_array_new_typed(1, true, &adamic_json_union_schema);
		adamic_array_push(array, (adamic_value){.reference = adamic_retain(hook)});
		show(array);
		adamic_release(array);
		adamic_release(hook);
	}
	show(keys);
	adamic_object *outer = adamic_object_new(&OUTER_SHAPE);
	outer->class = &outer_class;
	show(outer);
	adamic_release(outer);
	// Explicit absent/undefined tags, not false/zero/NaN inference.
	adamic_array *mixed = adamic_array_new_typed(5, false, &adamic_json_number_schema);
	adamic_array_push_typed(mixed, (adamic_value){.number = 0}, &adamic_json_number_schema);
	adamic_array_push_typed(mixed, (adamic_value){.boolean = false}, &adamic_json_boolean_schema);
	adamic_array_push_typed(mixed, (adamic_value){0}, &adamic_json_undefined_schema);
	adamic_array_push_typed(mixed, (adamic_value){.number = NAN}, &adamic_json_number_schema);
	adamic_array_push_typed(mixed, (adamic_value){0}, &adamic_json_null_schema);
	show(mixed);
	adamic_release(mixed);
	mixed = adamic_array_new_typed(3, false, &adamic_json_number_schema);
	adamic_array_push(mixed, (adamic_value){.number = 1});
	adamic_array_push_typed(mixed, (adamic_value){.boolean = false}, &adamic_json_boolean_schema);
	adamic_array_push(mixed, (adamic_value){.number = 2});
	slice = adamic_array_slice(mixed, 1, 0, false);
	show(adamic_array_reverse(slice));
	adamic_release(slice);
	adamic_array_sort(mixed, compare, NULL);
	show(mixed);
	adamic_release(mixed);
	adamic_array *empty = adamic_array_new_typed(0, false, &adamic_json_boolean_schema);
	show(empty);
	adamic_release(empty);
	adamic_array *bf = adamic_array_filled_typed(2, (adamic_value){.boolean = true}, false, &adamic_json_boolean_schema);
	show(bf);
	adamic_release(bf);
	adamic_array *sf = adamic_array_filled_typed(2, (adamic_value){.reference = &text}, true, &adamic_json_string_schema);
	show(sf);
	adamic_release(sf);
	adamic_array *entries = adamic_object_values(flags, false, true);
	show(entries);
	adamic_release(entries);
	adamic_map *bm = adamic_map_new_booleans(true);
	bm->json_value = &adamic_json_string_schema;
	adamic_map_set(bm, (adamic_value){.boolean = true}, (adamic_value){.reference = &text});
	adamic_map_set(bm, (adamic_value){.boolean = false}, (adamic_value){.reference = &text});
	adamic_array *listed = adamic_map_keys(bm);
	show(listed);
	adamic_release(listed);
	listed = adamic_map_values(bm);
	show(listed);
	adamic_release(listed);
	listed = adamic_map_entries(bm, &PAIR_SHAPE);
	show(listed);
	adamic_release(listed);
	listed = adamic_set_values(bm);
	show(listed);
	adamic_release(listed);
	adamic_release(bm);
	adamic_code callbacks[] = {parallel_number, parallel_boolean, parallel_string};
	const adamic_json_schema *descriptors[] = {&adamic_json_number_schema, &adamic_json_boolean_schema, &adamic_json_string_schema};
	for (size_t i = 0; i < 3; i++) {
		adamic_closure *work = adamic_closure_new(callbacks[i], 0);
		listed = adamic_parallel_map_typed(numbers, work, i == 2, descriptors[i]);
		show(listed);
		adamic_release(listed);
		adamic_release(work);
	}
	adamic_array *changed = adamic_array_slice(numbers, 0, 0, false);
	adamic_array_set_typed(changed, 0, (adamic_value){.number = 9}, &adamic_json_number_schema);
	show(changed);
	adamic_array *removed = adamic_array_splice_typed(changed, 0, 1, true, 1, (adamic_value[]){{.number = 8}}, &adamic_json_number_schema);
	show(removed);
	show(changed);
	adamic_release(removed);
	adamic_array_fill_typed(changed, (adamic_value){.number = 3}, 0, 0, false, false, &adamic_json_number_schema);
	show(changed);
	adamic_release(changed);
	static const adamic_json_schema optional_boolean = {.kind = adamic_json_maybe_boolean};
	adamic_array *optional = adamic_array_new_typed(3, true, &optional_boolean);
	adamic_array_push(optional, (adamic_value){.reference = &adamic_box_true});
	adamic_array_push(optional, (adamic_value){.reference = &adamic_box_false});
	adamic_array_push(optional, (adamic_value){.reference = NULL});
	show(optional);
	adamic_release(optional);
	optional = adamic_array_new_typed(2, false, &adamic_json_maybe_number_schema);
	adamic_array_push(optional, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){true, 0})});
	adamic_array_push(optional, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0})});
	show(optional);
	adamic_release(optional);
	reentrant_container = adamic_array_new_typed(1, true, &adamic_json_union_schema);
	adamic_object *reentrant = adamic_object_new(&REENTRANT_SHAPE);
	reentrant->slots[0].number = 7;
	adamic_array_push(reentrant_container, (adamic_value){.reference = reentrant});
	show(reentrant_container);
	adamic_release(reentrant_container);
	adamic_object *derived = adamic_object_new(&DERIVED_SHAPE);
	derived->slots[0].number = 0;
	derived->slots[1].reference = str("kept");
	show(derived);
	adamic_release(derived);
	adamic_object *ordered = adamic_object_new(&ORDERED_SHAPE);
	ordered->slots[0].reference = str("ten");
	ordered->slots[1].reference = str("two");
	ordered->slots[2].boolean = true;
	ordered->slots[3].reference = str("yes");
	show(ordered);
	adamic_release(ordered);
	adamic_object *private = adamic_object_new(&PRIVATE_SHAPE);
	private->slots[0].number = 9;
	private->slots[1].number = 7;
	show(private);
	adamic_release(private);
	adamic_release(filled);
	adamic_release(booleans);
	adamic_release(strings);
	adamic_release(flags);
	adamic_release(info);
	adamic_release(options);
	adamic_release(refs);
	adamic_release(map);
	adamic_release(simple);
	adamic_release(numbers);
	adamic_release(keys);
	return 0;
}
