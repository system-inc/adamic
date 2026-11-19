#include "adamic.h"
#include "count.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Two workers loop over one shared record's for...in keys many times at once (record.c's
// adamic_record_for_in_keys), each list checked against the record's own-key order. An unshared record's
// table keeps one list for every loop; a shared record's keeps none, so each loop gets a list of its own.
// A lazy fill of a shared table's list would be two threads writing key_list, and each retaining with
// plain counts an array the other made, which TSan reports (the for_in_keys_shared mutant). After the
// join, two loops over the shared record must hold two lists, which every build checks.

static adamic_record *record;
// Inserted b, 10, a, 2, c: array indexes ascending first, then insertion order.
static const char *const expected[] = {"2", "10", "b", "a", "c"};
static pthread_mutex_t gate = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t ready = PTHREAD_COND_INITIALIZER;
static int arrived;

static adamic_string *text(const char *bytes) {
	size_t length = strlen(bytes);
	adamic_string *made = adamic_string_allocate(length);
	memcpy((char *)made->bytes, bytes, length);
	return made;
}

static void check(const adamic_array *keys) {
	if (keys->length != sizeof expected / sizeof expected[0]) {
		fprintf(stderr, "for...in keys: %zu, want %zu\n", (size_t)keys->length, sizeof expected / sizeof expected[0]);
		abort();
	}
	for (size_t index = 0; index < keys->length; index++) {
		const adamic_string *key = keys->elements[index].reference;
		size_t length = strlen(expected[index]);
		if (key->length != length || memcmp(key->bytes, expected[index], length) != 0 || !adamic_record_has_own(record, key)) {
			fprintf(stderr, "for...in key %zu: %.*s, want %s\n", index, (int)key->length, key->bytes, expected[index]);
			abort();
		}
	}
}

static void *loop(void *given) {
	(void)given;
	// Both workers start their loops together, so their first asks for a list meet.
	pthread_mutex_lock(&gate);
	arrived++;
	pthread_cond_broadcast(&ready);
	while (arrived != 2) { pthread_cond_wait(&ready, &gate); }
	pthread_mutex_unlock(&gate);
	for (size_t pass = 0; pass < 20000; pass++) {
		adamic_array *keys = adamic_record_for_in_keys(record);
		check(keys);
		adamic_release(keys);
	}
	adamic_heap_thread_end();
	return NULL;
}

int main(void) {
	record = adamic_record_new(adamic_field_number);
	adamic_record_set(record, text("b"), (adamic_value){.number = 1});
	adamic_record_set(record, text("10"), (adamic_value){.number = 2});
	adamic_record_set(record, text("a"), (adamic_value){.number = 3});
	adamic_record_set(record, text("2"), (adamic_value){.number = 4});
	adamic_record_set(record, text("c"), (adamic_value){.number = 5});
	// Unshared, two loops read one list, which the table keeps when sharing begins.
	adamic_array *first = adamic_record_for_in_keys(record);
	adamic_array *second = adamic_record_for_in_keys(record);
	if (first != second) {
		fputs("an unshared record's loops hold two lists\n", stderr);
		abort();
	}
	check(first);
	adamic_release(first);
	adamic_release(second);
	adamic_share(record);
	pthread_t workers[2];
	for (size_t i = 0; i < 2; i++) { if (pthread_create(&workers[i], NULL, loop, (void *)i) != 0) { abort(); } }
	for (size_t i = 0; i < 2; i++) { pthread_join(workers[i], NULL); }
	adamic_array *one = adamic_record_for_in_keys(record);
	adamic_array *two = adamic_record_for_in_keys(record);
	if (one == two) {
		fputs("a shared record's loops share a list\n", stderr);
		abort();
	}
	check(one);
	check(two);
	adamic_release(one);
	adamic_release(two);
	adamic_release(record);
#ifdef ADAMIC_COUNT
	if (adamic_counted.allocations != adamic_counted.frees || adamic_counted.live != 0) { abort(); }
#endif
	puts("for-in keys clean");
	return 0;
}
