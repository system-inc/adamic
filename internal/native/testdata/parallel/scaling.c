#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "parallel.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static adamic_value read_text(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	adamic_string *text = arguments[0].reference;
	size_t units = (size_t)adamic_string_length(text);
	for (size_t i = 0; i < units; i++) {
		double expected = text->bytes[0] == 'a' ? 'a' : 233;
		if (adamic_string_char_code_at(text, (double)i) != expected ||
			adamic_string_code_point_at(text, (double)i).number != expected) { abort(); }
	}
	return (adamic_value){.number = (double)units};
}

int main(int argc, char **argv) {
	adamic_start(argc, argv);
	adamic_slot_cache cache = {0};
	if (argc > 1) {
		adamic_slot_cache_store(&cache, (const adamic_shape *)(uintptr_t)(UINT64_C(1) << 48), 0);
		abort();
	}
	if (adamic_parallel_grain(4096, 16) != 32 || adamic_parallel_grain(7, 16) != 1 ||
		adamic_parallel_grain(1000000, 4) != 256) { abort(); }
	// A large shape must miss the cache, not truncate its slot into a different field.
	size_t count = 65537;
	const char **names = malloc(count * sizeof *names);
	bool *references = calloc(count, sizeof *references);
	if (names == NULL || references == NULL) { abort(); }
	for (size_t i = 0; i < count; i++) { names[i] = "padding"; }
	names[count - 1] = "value";
	adamic_shape shape = {count, names, references, NULL};
	adamic_object *object = adamic_object_new(&shape);
	object->slots[count - 1].number = 987;
	for (size_t i = 0; i < 2; i++) {
		if (adamic_object_field(object, "value", &cache)->number != 987 || cache.packed != 0) { abort(); }
	}
	adamic_release(object);
	free(names); free(references);
	adamic_array *items = adamic_array_new(1024, true);
	// Repeated aliases exercise racing builders. Empty, small, ASCII and indexed BMP
	// strings all publish a length; only indexed strings need checkpoints and a view.
	size_t lengths[] = {0, 2, 160, 512};
	for (size_t k = 0; k < 4; k++) {
		adamic_string *text = adamic_string_allocate(lengths[k]);
		for (size_t i = 0; i < lengths[k]; i += k == 3 ? 1 : 2) {
			memcpy((char *)text->bytes + i, k == 3 ? "a" : "é", k == 3 ? 1 : 2);
		}
		size_t known_units = text->units;
		adamic_share(text);
		if (text->index != NULL || text->units != known_units) { abort(); }
		for (size_t i = 0; i < 256; i++) { adamic_array_push(items, (adamic_value){.reference = adamic_retain(text)}); }
		adamic_release(text);
	}
	adamic_closure *work = adamic_closure_new(read_text, 0);
	adamic_array *results = adamic_parallel_map(items, work, false);
	for (size_t i = 0; i < results->length; i++) {
		size_t k = i / 256;
		double expected = (double)(k == 3 ? lengths[k] : lengths[k] / 2);
		if (results->elements[i].number != expected) { abort(); }
	}
	adamic_release(results); adamic_release(work); adamic_release(items);
	puts("scaling clean");
	return 0;
}
