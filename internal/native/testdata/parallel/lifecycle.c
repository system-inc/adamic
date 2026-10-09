#include "adamic.h"
#include "parallel.h"
#include "count.h"
#include <stdio.h>
#include <stdlib.h>

static adamic_string text = ADAMIC_STRING("text");
static const char *const names[] = {"child"};
static const bool fields[] = {true};
static const adamic_field_kind shape_kinds_0[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, fields, NULL, shape_kinds_0};
static adamic_value scalar(adamic_closure *self, adamic_value *arguments) {
	(void)self; ADAMIC_CHECK_STACK();
	return arguments[0];
}
static adamic_value region_text(adamic_closure *self, adamic_value *arguments) {
	(void)self; ADAMIC_CHECK_STACK();
	adamic_object *object = arguments[0].reference;
	return (adamic_value){.reference = adamic_retain(object->slots[0].reference)};
}

int main(void) {
	adamic_array *empty = adamic_array_new(0, false);
	adamic_closure *callback = adamic_closure_new(scalar, 0);
	adamic_array *mapped = adamic_parallel_map(empty, callback, false);
	if (mapped->length != 0) { abort(); }
	adamic_release(mapped); adamic_release(empty); adamic_release(callback);
	adamic_array *one = adamic_array_new(1, false);
	adamic_array_push(one, (adamic_value){.number = 17});
	callback = adamic_closure_new(scalar, 0);
	mapped = adamic_parallel_map(one, callback, false);
	if (mapped->length != 1 || mapped->elements[0].number != 17) { abort(); }
	adamic_release(mapped); adamic_release(one); adamic_release(callback);

	// Zero-count region objects still have children requiring sharing. Their storage outlives join.
	adamic_region region = ADAMIC_REGION;
	adamic_object *object = adamic_object_new_in(&region, &shape);
	object->slots[0].reference = adamic_string_concat(1, (adamic_string *const[]){&text});
	adamic_array *items = adamic_array_new(1, true);
	adamic_array_push(items, (adamic_value){.reference = object});
	callback = adamic_closure_new(region_text, 0);
	mapped = adamic_parallel_map(items, callback, true);
	if (!adamic_is_shared(mapped->elements[0].reference) || adamic_reference_count(&object->heap) != 0) { abort(); }
	adamic_release(mapped); adamic_release(items); adamic_release(callback); adamic_region_end(&region);
	adamic_share(adamic_library_identity(4)); // An immortal identity has only a header, no object slots.

	// Once only this thread owns a formerly shared string, the acquire count permits reuse.
	adamic_string *unique = adamic_string_repeat(&text, 40);
	unique = adamic_string_append(unique, 1, (adamic_string *const[]){&text});
	adamic_share(unique);
	adamic_string *before = unique;
	unique = adamic_string_append(unique, 1, (adamic_string *const[]){&text});
	if (unique != before || adamic_reference_count(&unique->heap) != 1 || !adamic_is_shared(&unique->heap) || adamic_string_length(unique) != 168) { abort(); }
	adamic_release(unique);

	// Reused shared containers must publish fresh descendants on their next crossing.
	items = adamic_array_new(2, true);
	adamic_array_push(items, (adamic_value){.reference = adamic_string_concat(1, (adamic_string *const[]){&text})});
	adamic_share(items);
	adamic_string *fresh = adamic_string_repeat(&text, 40);
	adamic_array_push(items, (adamic_value){.reference = fresh});
	adamic_retain(items); // A new alias after reuse must not hide the fresh descendant.
	adamic_share(items);
	adamic_release(items);
	if (!adamic_is_shared(&fresh->heap)) { abort(); }
	adamic_release(items);

	void *chain = adamic_string_concat(1, (adamic_string *const[]){&text});
	for (size_t index = 0; index < 50000; index++) {
		object = adamic_object_new(&shape); object->slots[0].reference = chain; chain = object;
	}
	adamic_share(chain); adamic_release(chain); // Both graph marking and dropping are iterative.
#ifdef ADAMIC_COUNT
	if (adamic_counted.allocations != adamic_counted.frees + adamic_counted.regions || adamic_counted.live != 0) { abort(); }
#endif
	puts("lifecycle clean");
	return 0;
}
