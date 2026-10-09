#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "json_stringify.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static const adamic_json_schema number_schema = {adamic_json_number, NULL, 0, NULL, false};
static const adamic_json_schema string_schema = {adamic_json_string, NULL, 0, NULL, false};
static const adamic_json_schema strings_schema = {adamic_json_array, &string_schema, 0, NULL, false};
static const adamic_json_schema numbers_schema = {adamic_json_array, &number_schema, 0, NULL, false};
static adamic_string *text(const char *bytes) {
	size_t length = strlen(bytes);
	adamic_string *result = adamic_string_allocate(length);
	memcpy((char *)result->bytes, bytes, length);
	return result;
}
static void put(adamic_record *record, const char *key, double value) {
	adamic_record_define(record, text(key), (adamic_value){.number = value});
}
static void json(adamic_value value, const adamic_json_schema *schema) {
	adamic_string *result = adamic_json_stringify(value, schema, (adamic_value){0}, NULL, (adamic_value){0}, NULL);
	(void)fwrite(result->bytes, 1, result->length, stdout);
	putchar('\n');
	adamic_release(result);
}
static void snapshot(adamic_record *record) {
	adamic_array *keys = adamic_record_keys(record);
	json((adamic_value){.reference = keys}, &strings_schema);
	adamic_array *values = adamic_array_new(keys->length, false);
	for (size_t index = 0; index < keys->length; index++) {
		adamic_value *value = adamic_record_get_own(record, keys->elements[index].reference);
		adamic_array_push(values, *value);
	}
	json((adamic_value){.reference = values}, &numbers_schema);
	// Dynamic record JSON is not integrated with compiler schemas yet. Adapt the ordered
	// own slots to the existing JSON writer here, to check the consumer interface against Node.
	adamic_json_field *fields = calloc(keys->length == 0 ? 1 : keys->length, sizeof *fields);
	if (fields == NULL) abort();
	for (size_t index = 0; index < keys->length; index++) {
		fields[index] = (adamic_json_field){keys->elements[index].reference, index, &number_schema};
	}
	const adamic_json_schema schema = {adamic_json_object, NULL, keys->length, fields, false};
	// A borrowed object view for JSON only, never counted or freed as a heap object.
	adamic_object *view = calloc(1, sizeof *view + keys->length * sizeof *view->slots);
	if (view == NULL) abort();
	for (size_t index = 0; index < keys->length; index++) view->slots[index] = values->elements[index];
	json((adamic_value){.reference = view}, &schema);
	free(view);
	free(fields);
	adamic_release(values);
	adamic_release(keys);
}
static bool erase(adamic_record *record, const char *key) {
	adamic_string *name = text(key);
	bool result = adamic_record_delete(record, name);
	adamic_release(name);
	return result;
}
static void pairs(adamic_record *record) {
	adamic_record_iterator *iterator = adamic_record_iterate(record);
	adamic_string *key;
	adamic_value value;
	while (adamic_record_iterator_next(iterator, &key, &value)) {
		json((adamic_value){.reference = key}, &string_schema);
		printf("%.0f\n", value.number);
	}
	if (adamic_record_iterator_next(iterator, &key, &value)) abort();
	adamic_release(iterator);
}
static void semantics(void) {
	adamic_record *record = adamic_record_new(false);
	snapshot(record);
	const char *names[] = {"10", "2", "-1", "01", "4294967295", "1.5", "-0", "0", "4294967294", "1e0", "", "toString", "constructor", "__proto__", "hasOwnProperty", "世界🌍", "é"};
	for (size_t index = 0; index < sizeof names / sizeof names[0]; index++) put(record, names[index], (double)index);
	char long_key[4097];
	memset(long_key, '9', sizeof long_key - 1);
	long_key[sizeof long_key - 1] = '\0';
	put(record, long_key, 99);
	snapshot(record);
	put(record, "2", 100); // overwrite preserves the original position
	printf("%d %d %d\n", erase(record, "01"), erase(record, "10"), erase(record, "absent"));
	put(record, "01", 101);
	put(record, "10", 102);
	snapshot(record);
	pairs(record);
	// Spread into another record defines __proto__ as a data property.
	adamic_record *copy = adamic_record_new(false);
	adamic_record_iterator *iterator = adamic_record_iterate(record);
	adamic_string *key;
	adamic_value value;
	while (adamic_record_iterator_next(iterator, &key, &value)) {
		adamic_record_define(copy, adamic_retain(key), value);
	}
	adamic_release(iterator);
	put(copy, "2", 103);
	snapshot(copy);
	printf("%zu %zu\n", adamic_record_size(record), adamic_record_size(copy));
	adamic_release(record);
	pairs(copy); // copy's keys outlive the source
	adamic_release(copy);
}
static void prototypes(void) {
	adamic_record *record = adamic_record_new(false);
	const char *names[] = {"constructor", "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__", "isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString", "missing", "", "toStringX"};
	for (size_t index = 0; index < sizeof names / sizeof names[0]; index++) {
		adamic_string *key = text(names[index]);
		printf("%d %d %d\n", adamic_record_has_own(record, key), adamic_record_has_own(record, key), adamic_record_get_own(record, key) != NULL);
		if (strcmp(names[index], "__proto__") == 0) {
			adamic_record_define(record, adamic_retain(key), (adamic_value){.number = 7});
		} else {
			adamic_record_set(record, adamic_retain(key), (adamic_value){.number = 7});
		}
		printf("%d %d %.0f\n", adamic_record_has_own(record, key), adamic_record_has(record, key), adamic_record_get(record, key)->number);
		(void)adamic_record_delete(record, key);
		printf("%d %d\n", adamic_record_has_own(record, key), adamic_record_get_own(record, key) != NULL);
		adamic_release(key);
	}
	snapshot(record);
	adamic_release(record);
}
static void reads(void) {
	adamic_record *record = adamic_record_new(false);
	put(record, "hit", 42);
	put(record, "toString", 7);
	const char *hits[] = {"hit", "toString"};
	for (size_t index = 0; index < sizeof hits / sizeof hits[0]; index++) {
		adamic_string *key = text(hits[index]);
		adamic_value *value = adamic_record_get(record, key);
		printf("%.0f %d %d\n", value->number, value == adamic_record_get_own(record, key), adamic_record_has(record, key));
		adamic_release(key);
	}
	// Include near-matches in every member-length group, plus long, empty and Unicode keys.
	const char *misses[] = {"missing", "valueOX", "toStrinX", "__proto_X", "constructoX",
		"isPrototypeOX", "hasOwnPropertX", "toLocaleStrinX", "__defineGetter_X",
		"__defineSetter_X", "__lookupGetter_X", "__lookupSetter_X", "propertyIsEnumerablX",
		"", "a-long-key-more-than-twenty-bytes", "世界🌍"};
	for (size_t index = 0; index < sizeof misses / sizeof misses[0]; index++) {
		adamic_string *key = text(misses[index]);
		printf("%d %d\n", adamic_record_get(record, key) == NULL, adamic_record_has(record, key));
		adamic_release(key);
	}
	adamic_release(record);
	// An own undefined is present, even under a prototype-member name.
	record = adamic_record_new(true);
	adamic_string *key = text("toString");
	adamic_record_define(record, adamic_retain(key), (adamic_value){.reference = NULL});
	adamic_value *value = adamic_record_get(record, key);
	printf("%d %d %d\n", value != NULL, value->reference == NULL, adamic_record_has(record, key));
	adamic_record_define(record, adamic_retain(key), (adamic_value){.reference = text("owned value")});
	adamic_string *held = adamic_retain(adamic_record_get(record, key)->reference);
	(void)adamic_record_delete(record, key);
	printf("%d %d\n", adamic_record_has_own(record, key), adamic_record_get_own(record, key) != NULL);
	adamic_release(key);
	adamic_release(record);
	json((adamic_value){.reference = held}, &string_schema);
	adamic_release(held);
}

static void missing_member(const char *operation, const char *name) {
	adamic_record *record = adamic_record_new(false);
	adamic_string *key = text(name);
	if (strcmp(operation, "missing-get") == 0) {
		printf("%d\n", adamic_record_get(record, key) != NULL);
	} else {
		printf("%d\n", adamic_record_has(record, key));
	}
	adamic_release(key);
	adamic_release(record);
}

static void references(void) {
	adamic_record *record = adamic_record_new(true);
	for (size_t index = 0; index < 10000; index++) {
		// Both key and value are fresh counted strings. Immortal literals would hide a leak.
		adamic_record_set(record, text("same"), (adamic_value){.reference = text("value")});
	}
	adamic_string *key = text("same");
	adamic_string *held = adamic_retain(adamic_record_get_own(record, key)->reference);
	adamic_array *keys = adamic_record_keys(record);
	(void)adamic_record_delete(record, key);
	adamic_release(key);
	adamic_release(record);
	json((adamic_value){.reference = held}, &string_schema);
	json((adamic_value){.reference = keys}, &strings_schema);
	adamic_release(held);
	adamic_release(keys);
	// Compaction and repeated rehashing, with a snapshot alive across deletions.
	record = adamic_record_new(true);
	for (int round = 0; round < 10; round++) {
		for (int index = 0; index < 1000; index++) {
			char bytes[32];
			(void)snprintf(bytes, sizeof bytes, "k%d", index);
			adamic_record_set(record, text(bytes), (adamic_value){.reference = text(bytes)});
		}
		keys = adamic_record_keys(record);
		for (int index = 0; index < 1000; index++) {
			char bytes[32];
			(void)snprintf(bytes, sizeof bytes, "k%d", index);
			(void)erase(record, bytes);
		}
		adamic_release(keys);
	}
	printf("%zu\n", adamic_record_size(record));
	adamic_release(record);
	// Presence is independent of a reference value being undefined.
	record = adamic_record_new(true);
	key = text("undefined");
	adamic_record_set(record, adamic_retain(key), (adamic_value){.reference = NULL});
	printf("%d %d %d\n", adamic_record_has_own(record, key), adamic_record_get_own(record, key) != NULL, adamic_record_get_own(record, key)->reference == NULL);
	adamic_release(key);
	adamic_release(record);
	// Scalar flags never read the unspecified bytes of a boolean as a number.
	record = adamic_record_new(false);
	key = text("flag");
	adamic_record_set(record, adamic_retain(key), (adamic_value){.boolean = true});
	printf("%d\n", adamic_record_get_own(record, key)->boolean);
	adamic_record_set(record, adamic_retain(key), (adamic_value){.boolean = false});
	printf("%d %d\n", adamic_record_has_own(record, key), adamic_record_get_own(record, key)->boolean);
	adamic_release(key);
	adamic_release(record);
	// Records nested as values must use iterative heap cleanup, not recursive destructors.
	adamic_record *chain = NULL;
	for (size_t index = 0; index < 100000; index++) {
		adamic_record *next = adamic_record_new(true);
		adamic_record_define(next, text("next"), (adamic_value){.reference = chain});
		chain = next;
	}
	adamic_release(chain);
	puts("chain released");
}
static void iteration(void) {
	adamic_record *record = adamic_record_new(false);
	put(record, "10", 10); put(record, "2", 2); put(record, "a", 1); put(record, "b", 2);
	adamic_record_iterator *iterator = adamic_record_iterate(record);
	adamic_record_iterator *second = adamic_record_iterate(record);
	(void)erase(record, "10");
	put(record, "new", 3); put(record, "a", 9);
	adamic_string *key;
	adamic_value value;
	while (adamic_record_iterator_next(iterator, &key, &value)) {
		json((adamic_value){.reference = key}, &string_schema);
		printf("%.0f\n", value.number);
	}
	adamic_release(iterator);
	// Iterators own the record and their key snapshot, even after the caller drops its count.
	adamic_release(record);
	while (adamic_record_iterator_next(second, &key, &value)) {
		json((adamic_value){.reference = key}, &string_schema);
		printf("%.0f\n", value.number);
	}
	adamic_release(second);
}
static void numeric(size_t count) {
	adamic_record *record = adamic_record_new(false);
	for (size_t index = count; index > 0; index--) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "%zu", index * 3);
		put(record, bytes, (double)index);
	}
	for (size_t index = 2; index <= count; index += 2) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "%zu", index * 3);
		(void)erase(record, bytes);
	}
	for (size_t index = count; index > 0; index--) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "%zu", index * 3);
		put(record, bytes, (double)index);
	}
	pairs(record);
	adamic_release(record);
}
static double now(void) {
	struct timespec time;
	if (clock_gettime(CLOCK_MONOTONIC, &time) != 0) abort();
	return (double)time.tv_sec + (double)time.tv_nsec / 1e9;
}
static void workload(size_t count, bool benchmark) {
	adamic_record *record = adamic_record_new(false);
	double start = now();
	for (size_t index = 0; index < count; index++) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "k%zu", index);
		adamic_record_set(record, text(bytes), (adamic_value){.number = (double)index});
	}
	double build = now() - start;
	start = now();
	double sum = 0;
	for (size_t index = 0; index < count; index++) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "k%zu", index);
		adamic_string *key = text(bytes);
		adamic_value *value = adamic_record_get(record, key);
		if (value == NULL) abort();
		sum += value->number;
		adamic_release(key);
	}
	double hit = now() - start;
	start = now();
	size_t misses = 0;
	for (size_t index = 0; index < count; index++) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "m%zu", index);
		adamic_string *key = text(bytes);
		misses += adamic_record_get(record, key) == NULL;
		adamic_release(key);
	}
	double miss = now() - start;
	start = now();
	for (size_t index = 0; index < count; index += 2) {
		char bytes[32];
		(void)snprintf(bytes, sizeof bytes, "k%zu", index);
		(void)erase(record, bytes);
	}
	double deletion = now() - start;
	start = now();
	adamic_record_iterator *iterator = adamic_record_iterate(record);
	adamic_string *key;
	adamic_value value;
	size_t visited = 0;
	while (adamic_record_iterator_next(iterator, &key, &value)) {
		visited++;
		sum += value.number;
		// The million-key correctness run compares every remaining key, not just a checksum.
		if (!benchmark) { (void)fwrite(key->bytes, 1, key->length, stdout); putchar('\n'); }
	}
	adamic_release(iterator);
	double iteration_time = now() - start;
	printf("%.0f %zu %zu %zu\n", sum, misses, visited, adamic_record_size(record));
	if (benchmark) printf("%.9f %.9f %.9f %.9f %.9f\n", build, hit, miss, deletion, iteration_time);
	adamic_release(record);
}
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	if (argc < 2) return 2;
	if (strcmp(argv[1], "semantics") == 0) semantics();
	else if (strcmp(argv[1], "prototypes") == 0) prototypes();
	else if (strcmp(argv[1], "references") == 0) references();
	else if (strcmp(argv[1], "reads") == 0) reads();
	else if (strcmp(argv[1], "missing-get") == 0 || strcmp(argv[1], "missing-has") == 0) {
		if (argc != 3) return 2;
		missing_member(argv[1], argv[2]);
	}
	else if (strcmp(argv[1], "iteration") == 0) iteration();
	else if (strcmp(argv[1], "numeric") == 0) numeric(100000);
	else if (strcmp(argv[1], "workload") == 0 || strcmp(argv[1], "bench") == 0) {
		if (argc != 3) return 2;
		workload((size_t)strtoull(argv[2], NULL, 10), strcmp(argv[1], "bench") == 0);
	} else if (strcmp(argv[1], "proto-assignment") == 0) {
		adamic_record *record = adamic_record_new(false);
		adamic_record_set(record, text("__proto__"), (adamic_value){.number = 42});
		adamic_release(record);
	} else return 2;
	return 0;
}
