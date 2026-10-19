// node_path.c: POSIX lexical paths. No filesystem lookup, including for NUL
// bytes.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static void *allocate(size_t size) {
	void *result = malloc(size);
	if (result == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return result;
}
static adamic_string *text(const char *bytes, size_t length) {
	adamic_string *result =
		adamic_allocate(sizeof *result + length, adamic_kind_string);
	result->length = length;
	result->bytes = (const char *)(result + 1);
	result->units = 0;
	result->index = NULL;
	result->owner = NULL;
	result->capacity = length;
	memcpy((char *)result->bytes, bytes, length);
	return result;
}
// Each stack entry marks the beginning of a normal component or an unresolved
// '..'.
static adamic_string *normalize(const char *bytes, size_t length, bool absolute,
								bool trailing) {
	char *output = allocate(length + 3);
	size_t *starts = allocate((length + 1) * sizeof *starts);
	bool *parents = allocate((length + 1) * sizeof *parents);
	size_t used = absolute ? 1 : 0, count = 0;
	if (absolute) {
		output[0] = '/';
	}
	for (size_t at = 0; at < length;) {
		while (at < length && bytes[at] == '/') {
			at++;
		}
		size_t begin = at;
		while (at < length && bytes[at] != '/') {
			at++;
		}
		size_t size = at - begin;
		if (size == 0 || (size == 1 && bytes[begin] == '.')) {
			continue;
		}
		bool parent =
			size == 2 && bytes[begin] == '.' && bytes[begin + 1] == '.';
		if (parent && count > 0 && !parents[count - 1]) {
			used = starts[--count];
			continue;
		}
		if (parent && absolute) {
			continue;
		}
		starts[count] = used;
		parents[count++] = parent;
		if (used > 0 && output[used - 1] != '/') {
			output[used++] = '/';
		}
		memcpy(output + used, bytes + begin, size);
		used += size;
	}
	if (used == 0) {
		output[used++] = '.';
	}
	if (trailing && output[used - 1] != '/') {
		output[used++] = '/';
	}
	adamic_string *result = text(output, used);
	free(parents);
	free(starts);
	free(output);
	return result;
}
adamic_string *adamic_fs_file_host_join(size_t count,
									 adamic_string *const paths[]) {
	size_t length = 0;
	for (size_t i = 0; i < count; i++) {
		if (paths[i]->length > 0) {
			length += paths[i]->length + 1;
		}
	}
	if (length == 0) {
		return text(".", 1);
	}
	char *bytes = allocate(length);
	size_t used = 0;
	for (size_t i = 0; i < count; i++) {
		if (paths[i]->length == 0) {
			continue;
		}
		if (used > 0) {
			bytes[used++] = '/';
		}
		memcpy(bytes + used, paths[i]->bytes, paths[i]->length);
		used += paths[i]->length;
	}
	adamic_string *result =
		normalize(bytes, used, bytes[0] == '/', bytes[used - 1] == '/');
	free(bytes);
	return result;
}
// Compatibility entry points use the same Node policy and runtime providers.
adamic_string *adamic_fs_file_host_cwd(void) { return adamic_node_cwd(); }
double adamic_fs_file_host_chdir(const adamic_string *directory) {
    adamic_node_chdir(directory);
    return 0;
}
adamic_string *adamic_fs_file_host_tmpdir(void) { return adamic_node_tmpdir(); }
