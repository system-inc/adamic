#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "parallel.h"
#include "count.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

enum { numbers, strings, objects, captured_map, fresh, nested, exceptions, million, panic_work, nested_exception };
static pthread_t caller;
static bool single;
static bool timed;
static pthread_mutex_t first_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t first_ready = PTHREAD_COND_INITIALIZER;
static size_t first_arrived;
static _Thread_local bool first_string;

// Force the first four callbacks to overlap before any result can mark its input owner. Without
// this rendezvous, a missing items share can be accidentally repaired by the first slice result.
static void first_string_reads(adamic_string *text) {
	if (single || timed || first_string) { return; }
	first_string = true;
	pthread_mutex_lock(&first_lock);
	first_arrived++;
	pthread_cond_broadcast(&first_ready);
	while (first_arrived != adamic_parallel_threads()) { pthread_cond_wait(&first_ready, &first_lock); }
	pthread_mutex_unlock(&first_lock);
	for (size_t i = 0; i < 10000; i++) { adamic_retain(text); adamic_release(text); }
}
static const char *const names[] = {"value", "children"};
static const bool fields[] = {false, true};
static const adamic_field_kind shape_kinds_0[] = {adamic_field_number, adamic_field_reference};
static const adamic_shape shape = {2, names, fields, NULL, shape_kinds_0, NULL, NULL};
static adamic_string ascii = ADAMIC_STRING("ASCII");

static adamic_value inner(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	return (adamic_value){.number = arguments[0].number + arguments[1].number};
}

static adamic_value inner_throw(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	if (arguments[1].number == 3) {
		adamic_string *message = adamic_string_from_number(3);
		adamic_thrown = adamic_error_new(message);
		adamic_release(message);
	}
	return (adamic_value){.number = arguments[0].number};
}

static adamic_value work(adamic_closure *self, adamic_value *arguments) {
	ADAMIC_CHECK_STACK();
	if (single && !pthread_equal(caller, pthread_self())) { abort(); }
	int mode = (int)self->cells[0]->value.number;
	size_t index = (size_t)arguments[1].number;
	if (mode == strings) {
		adamic_string *text = arguments[0].reference;
		first_string_reads(text);
		adamic_retain(text);
		// Both supplementary-point cursors and compact BMP views, in scattered read order.
		size_t length = (size_t)adamic_string_length(text);
		double sum = 0;
		for (size_t i = 0; i < 256; i++) { sum += adamic_string_char_code_at(text, (double)((i * 37 + index) % length)); }
		if (sum == 0) { abort(); }
		adamic_string *slice = adamic_string_slice(text, 0, (double)length, true);
		adamic_release(text);
		return (adamic_value){.reference = slice};
	}
	if (mode == objects) {
		adamic_object *object = arguments[0].reference;
		static _Thread_local adamic_slot_cache cache;
		adamic_array *children = adamic_object_field(object, "children", &cache)->reference;
		adamic_array *child = children->elements[0].reference;
		adamic_string *text = child->elements[0].reference;
		adamic_retain(text); adamic_release(text);
		return (adamic_value){.number = object->slots[0].number + (double)text->length};
	}
	if (mode == captured_map) {
		adamic_map *map = self->cells[1]->value.reference;
		adamic_value *slot = adamic_map_get(map, (adamic_value){.number = (double)(index % 16)});
		if (slot == NULL) { abort(); }
		adamic_string *text = slot->reference;
		adamic_retain(text); adamic_release(text);
		adamic_map_iterator *iterator = adamic_map_iterate(map);
		adamic_value key, value;
		if (!adamic_map_iterator_next(iterator, &key, &value)) { abort(); }
		adamic_release(iterator);
		return (adamic_value){.number = (double)text->length + arguments[0].number};
	}
	if (mode == fresh || mode == exceptions) {
		if (mode == exceptions && (index == 257 || index == 900)) {
			// The higher index can throw first. Completion order must not choose the exception.
			if (index == 257) { struct timespec delay = {0, 20000000}; nanosleep(&delay, NULL); }
			adamic_string *message = adamic_string_from_number((double)index);
			adamic_thrown = adamic_error_new(message);
			adamic_release(message);
			// The callback result is owned even when an exception is pending.
			return (adamic_value){.reference = adamic_object_new(&shape)};
		}
		adamic_object *result = adamic_object_new(&shape);
		result->slots[0].number = arguments[0].number * 2 + arguments[1].number;
		adamic_array *children = adamic_array_new(1, true);
		adamic_array_push(children, (adamic_value){.reference = adamic_string_from_number((double)index)});
		result->slots[1].reference = children;
		return (adamic_value){.reference = result};
	}
	if (mode == nested || mode == nested_exception) {
		adamic_array *items = adamic_array_new(16, false);
		for (size_t i = 0; i < 16; i++) { adamic_array_push(items, (adamic_value){.number = (double)i}); }
		adamic_closure *callback = adamic_closure_new(mode == nested ? inner : inner_throw, 0);
		adamic_array *results = adamic_parallel_map(items, callback, false);
		adamic_release(callback); adamic_release(items);
		if (adamic_thrown != NULL) { return (adamic_value){.number = 0}; }
		double sum = 0;
		for (size_t i = 0; i < results->length; i++) { sum += results->elements[i].number; }
		adamic_release(results);
		return (adamic_value){.number = sum + (double)index};
	}
	if (mode == panic_work) {
		if (!pthread_equal(caller, pthread_self())) { adamic_panic("worker panic", sizeof "worker panic" - 1); }
		struct timespec delay = {0, 1000000}; nanosleep(&delay, NULL);
	}
	return (adamic_value){.number = arguments[0].number * 2 + arguments[1].number};
}

static adamic_string *long_text(bool bmp) {
	size_t length = bmp ? 4096 : 6144;
	adamic_string *text = adamic_string_allocate(length);
	for (size_t i = 0; i < length; i += bmp ? 2 : 6) {
		memcpy((char *)text->bytes + i, bmp ? "é" : "é😀", bmp ? 2 : 6);
	}
	return text;
}

int main(int argc, char **argv) {
	adamic_start(argc, argv);
	if (argc < 2) { return 2; }
	const char *modes[] = {"numbers", "strings", "objects", "map", "fresh", "nested", "exception", "million", "panic", "nested_exception"};
	int mode = -1;
	for (int i = 0; i < 10; i++) { if (strcmp(argv[1], modes[i]) == 0) { mode = i; } }
	if (mode < 0) { return 2; }
	if (mode == panic_work) {
		static adamic_string preface = ADAMIC_STRING("before worker panic");
		adamic_write_line(adamic_stdout, &preface);
	}
	caller = pthread_self();
	timed = argc > 2;
	single = strcmp(getenv("ADAMIC_THREADS") == NULL ? "" : getenv("ADAMIC_THREADS"), "1") == 0;
	size_t count = mode == million ? 1000000 : mode == strings && argc > 2 ? 32768 : 2048;
	if (mode == nested_exception) { count = 32; }
	adamic_array *items = adamic_array_new(count, mode == strings || mode == objects);
	adamic_closure *callback = adamic_closure_new(work, mode == captured_map ? 2 : 1);
	callback->cells[0] = adamic_cell_new((adamic_value){.number = (double)mode}, false);
	adamic_string *texts[2] = {NULL, NULL};
	if (mode == strings) { texts[0] = long_text(true); texts[1] = long_text(false); }
	if (mode == captured_map) {
		adamic_map *map = adamic_map_new(false, true);
		for (size_t i = 0; i < 16; i++) {
			adamic_map_set(map, (adamic_value){.number = (double)i}, (adamic_value){.reference = adamic_string_concat(1, (adamic_string *const[]){&ascii})});
		}
		callback->cells[1] = adamic_cell_new((adamic_value){.reference = map}, true);
	}
	for (size_t i = 0; i < count; i++) {
		adamic_value item = {.number = (double)i};
		if (mode == strings) { item.reference = adamic_retain(texts[i % 2]); }
		if (mode == objects) {
			adamic_object *object = adamic_object_new(&shape);
			object->slots[0].number = (double)i;
			adamic_array *children = adamic_array_new(1, true), *child = adamic_array_new(1, true);
			adamic_array_push(child, (adamic_value){.reference = adamic_string_concat(1, (adamic_string *const[]){&ascii})});
			adamic_array_push(children, (adamic_value){.reference = child});
			object->slots[1].reference = children;
			item.reference = object;
		}
		adamic_array_push(items, item);
	}
	adamic_release(texts[0]); adamic_release(texts[1]);
	struct timespec before, after;
	clock_gettime(CLOCK_MONOTONIC, &before);
	adamic_array *results = adamic_parallel_map(items, callback, mode == strings || mode == fresh || mode == exceptions);
	clock_gettime(CLOCK_MONOTONIC, &after);
	if (single && adamic_parallel_workers() != 0) { abort(); }
	if (mode == exceptions || mode == nested_exception) {
		if (results != NULL || adamic_thrown == NULL) { abort(); }
		adamic_string *message = adamic_thrown->slots[1].reference;
		const char *expected = mode == exceptions ? "257" : "3";
		if (message->length != strlen(expected) || memcmp(message->bytes, expected, message->length) != 0) { abort(); }
		printf("exception %s\n", expected);
		adamic_release(adamic_thrown); adamic_thrown = NULL;
	} else {
		if (results == NULL || results->length != count) { abort(); }
		double sum = 0;
		for (size_t i = 0; i < count; i++) {
			double value = results->elements[i].number;
			double expected = (double)i * 3;
			if (mode == fresh) {
				adamic_object *object = results->elements[i].reference;
				adamic_array *children = object->slots[1].reference;
				if (!adamic_is_shared(&object->heap) || !adamic_is_shared(&children->heap) || !adamic_is_shared(children->elements[0].reference)) { abort(); }
				value = object->slots[0].number;
			}
			if (mode == strings) {
				adamic_string *text = results->elements[i].reference;
				if (!adamic_is_shared(&text->heap)) { abort(); }
				value = adamic_string_length(text); expected = i % 2 == 0 ? 2048 : 3072;
			}
			if (mode == objects || mode == captured_map) { expected = (double)i + 5; }
			if (mode == nested) { expected = (double)i + 240; }
			if (value != expected) { fprintf(stderr, "index %zu: got %.0f expected %.0f\n", i, value, expected); return 3; }
			sum += value;
		}
		printf("%s %zu %.0f\n", modes[mode], count, sum);
	}
	if (argc > 2) {
		double seconds = (double)(after.tv_sec - before.tv_sec) + (double)(after.tv_nsec - before.tv_nsec) / 1e9;
		fprintf(stderr, "map seconds %.9f\n", seconds);
	}
	adamic_release(results); adamic_release(callback); adamic_release(items);
#ifdef ADAMIC_COUNT
	if (adamic_counted.allocations != adamic_counted.frees || adamic_counted.live != 0) { abort(); }
#endif
	return 0;
}
