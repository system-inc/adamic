#include "adamic.h"
#include "count.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *const names[] = {"text", "padding"};
static const bool references[] = {true, true};
static const adamic_shape shape = {2, names, references, NULL};
static const char *const reversed_names[] = {"padding", "text"};
static const bool reversed_references[] = {true, true};
static const adamic_shape reversed_shape = {2, reversed_names, reversed_references, NULL};
static adamic_string literal = ADAMIC_STRING("é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀é😀");
static adamic_object *objects[2];
static adamic_map *map;
static adamic_heap **remote_values;
static size_t remote_count = 300000;
static pthread_mutex_t gate = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t ready = PTHREAD_COND_INITIALIZER;
static int arrived;
static int remote_arrived;

// The owner and all remote releasers enter the allocator phase together.
static void remote_ready(void) {
	pthread_mutex_lock(&gate);
	remote_arrived++;
	pthread_cond_broadcast(&ready);
	while (remote_arrived != 5) { pthread_cond_wait(&ready, &gate); }
	pthread_mutex_unlock(&gate);
}

static void *run(void *given) {
	size_t worker = (size_t)given;
	pthread_mutex_lock(&gate);
	arrived++;
	pthread_cond_broadcast(&ready);
	while (arrived != 4) { pthread_cond_wait(&ready, &gate); }
	pthread_mutex_unlock(&gate);
	// Count contention precedes caches, iterators and Weak's mutex. Those can
	// accidentally order the count accesses that this fixture needs to expose.
	adamic_string *shared_text = objects[0]->slots[0].reference;
	for (size_t i = 0; i < 256; i++) { adamic_retain(shared_text); }
	for (size_t i = 0; i < 256; i++) { adamic_release(shared_text); }
	for (size_t i = 0; i < 10000; i++) {
		// One call site alternates different shapes: a torn cache can name the wrong slot.
		static _Thread_local adamic_slot_cache cache;
		adamic_object *object = objects[(i + worker) % 2];
		adamic_string *text = adamic_object_field(object, "text", &cache)->reference;
		adamic_retain(text);
		if (adamic_string_length(text) != 60 || adamic_string_char_code_at(text, 1) != 55357 ||
			adamic_string_code_point_at(text, 1).number != 128512 || adamic_string_length(&literal) != 60) { abort(); }
		// Scattered cursor reads as well as direct BMP views.
		bool low;
		(void)adamic_string_locate(text, i % 60, &low);
		adamic_release(text);
		adamic_map_iterator *iterator = adamic_map_iterate(map);
		adamic_value key, value;
		if (!adamic_map_iterator_next(iterator, &key, &value) || value.number != 17) { abort(); }
		adamic_release(iterator);
		// Weak handles are local but their global table overlaps frees in other workers.
		adamic_number_box *box = (adamic_number_box *)adamic_box_number(1);
		adamic_weak *weak = adamic_weak_of(box);
		if (!adamic_weak_held(box)) { abort(); }
		adamic_release(box);
		if (adamic_weak_target(weak) != NULL) { abort(); }
		adamic_release(weak);
	}
	remote_ready();
	for (size_t i = worker; i < remote_count; i += 4) { adamic_release(remote_values[i]); }
	adamic_heap_thread_end();
	return NULL;
}

int main(void) {
	objects[0] = adamic_object_new(&shape);
	objects[1] = adamic_object_new(&reversed_shape);
	adamic_string *text = adamic_string_concat(1, (adamic_string *const[]){&literal});
	objects[0]->slots[0].reference = text;
	// Both shapes have two valid reference slots. A torn cache must reach TSan rather than crash
	// on a garbage pointer before the detector can report the race we are proving.
	objects[0]->slots[1].reference = adamic_retain(text);
	objects[1]->slots[0].reference = adamic_retain(text);
	objects[1]->slots[1].reference = adamic_retain(text);
	adamic_share(objects[0]);
	adamic_share(objects[1]);
	map = adamic_map_new(false, false);
	adamic_map_set(map, (adamic_value){.number = 3}, (adamic_value){.number = 17});
	adamic_share(map);
	remote_values = malloc(remote_count * sizeof *remote_values);
	if (remote_values == NULL) { abort(); }
	for (size_t i = 0; i < remote_count; i++) {
		remote_values[i] = adamic_allocate(256, adamic_kind_number);
		adamic_share(remote_values[i]);
	}
	pthread_t workers[4];
	for (size_t i = 0; i < 4; i++) { if (pthread_create(&workers[i], NULL, run, (void *)i) != 0) { abort(); } }
	// Allocate/free while other threads return slots into this owner's chunks.
	remote_ready();
	for (size_t i = 0; i < remote_count; i++) {
		adamic_heap *value = adamic_allocate(256, adamic_kind_number);
		adamic_release(value);
	}
	for (size_t i = 0; i < 4; i++) { pthread_join(workers[i], NULL); }
	free(remote_values);
	adamic_release(objects[0]); adamic_release(objects[1]); adamic_release(map);
#ifdef ADAMIC_COUNT
	if (adamic_counted.allocations != adamic_counted.frees || adamic_counted.live != 0) { abort(); }
#endif
	puts("memory clean");
	return 0;
}
