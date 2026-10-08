// directory.c: the file system a walk needs, opened in 0.2: readDirectory and fileStatus.
//
// Each answers as Node does on the same machine. readdirSync's names come from libuv's scandir,
// sorted by strcmp on the raw bytes, without . and .., and are decoded as UTF-8 after the sort, so
// the order is the bytes' even where a name isn't valid UTF-8. statSync follows a symbolic link and
// lstatSync doesn't; fileStatus gives the type and size of what a path names, following a link, and
// says whether the path itself is one.

// lstat, S_ISLNK and the directory calls are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"

#include <dirent.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>

static const char *const ok_listing_names[] = {"kind", "names"};
static const char *const ok_status_names[] = {"kind", "type", "size", "symbolicLink"};
static const char *const error_names[] = {"kind", "message"};
static const bool two_references[] = {true, true};
static const bool status_references[] = {true, true, false, false};
static const adamic_shape listing_shape = {2, ok_listing_names, two_references, NULL, NULL};
static const adamic_shape status_shape = {4, ok_status_names, status_references, NULL, NULL};
static const adamic_shape error_shape = {2, error_names, two_references, NULL, NULL};
static adamic_string ok_kind = ADAMIC_STRING("Ok");
static adamic_string error_kind = ADAMIC_STRING("Error");
static adamic_string file_type = ADAMIC_STRING("file");
static adamic_string directory_type = ADAMIC_STRING("directory");
static adamic_string other_type = ADAMIC_STRING("other");

// failure is { kind: 'Error', message }: "<prefix><path>: <reason>", in Adamic's own words, the same
// on every platform. The JavaScript runtime says the same (oracle/adamic.mjs), from Node's error
// codes, which are these errno names.
static adamic_object *failure(const char *prefix, size_t prefix_length, const adamic_string *path, int error, bool listing) {
	adamic_string before = {{0, adamic_kind_string, 0}, prefix_length, prefix, 0, NULL, NULL, 0};
	static adamic_string missing_file = ADAMIC_STRING(": no such file");
	static adamic_string missing_directory = ADAMIC_STRING(": no such directory");
	static adamic_string not_directory = ADAMIC_STRING(": not a directory");
	static adamic_string denied = ADAMIC_STRING(": permission denied");
	static adamic_string failed = ADAMIC_STRING(": failed");
	adamic_string *reason = &failed;
	switch (error) {
	case ENOENT:
		reason = listing ? &missing_directory : &missing_file;
		break;
	case ENOTDIR:
		reason = listing ? &not_directory : &missing_file;
		break;
	case EACCES:
	case EPERM:
		reason = &denied;
		break;
	}
	adamic_object *result = adamic_object_new(&error_shape);
	result->slots[0].reference = &error_kind;
	result->slots[1].reference = adamic_string_concat(3, (adamic_string *const[]){&before, (adamic_string *)path, reason});
	return result;
}

static int compare_names(const void *left, const void *right) {
	return strcmp(*(char *const *)left, *(char *const *)right);
}

adamic_object *adamic_read_directory(const adamic_string *path) {
	static const char prefix[] = "cannot read directory ";
	// What was printed comes first, as on Node, before anything of the file system is looked at.
	adamic_output_flush();
#ifdef ADAMIC_TARGET_WASI
	// WASI libc treats the empty path as cwd; Node reports ENOENT.
	if (path->length == 0) {
		return failure(prefix, sizeof prefix - 1, path, ENOENT, true);
	}
#endif
	char *name = adamic_path_bytes(path);
	if (name == NULL) {
		return failure(prefix, sizeof prefix - 1, path, 0, true);
	}
	DIR *directory = opendir(name);
	free(name);
	if (directory == NULL) {
		return failure(prefix, sizeof prefix - 1, path, errno, true);
	}
	char **names = NULL;
	size_t count = 0, capacity = 0;
	int error = 0;
	for (;;) {
		errno = 0;
		struct dirent *entry = readdir(directory);
		if (entry == NULL) {
			error = errno;
			break;
		}
		if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
			continue;
		}
		if (count == capacity) {
			capacity = capacity == 0 ? 16 : capacity * 2;
			char **grown = realloc(names, capacity * sizeof *grown);
			if (grown == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			names = grown;
		}
		names[count] = strdup(entry->d_name);
		if (names[count] == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		count++;
	}
	closedir(directory);
	if (error != 0) {
		for (size_t index = 0; index < count; index++) {
			free(names[index]);
		}
		free(names);
		return failure(prefix, sizeof prefix - 1, path, error, true);
	}
	if (count > 1) {
		qsort(names, count, sizeof *names, compare_names);
	}
	adamic_array *listed = adamic_array_new(count, true);
	for (size_t index = 0; index < count; index++) {
		adamic_array_push(listed, (adamic_value){.reference = adamic_decode_utf8((const unsigned char *)names[index], strlen(names[index]))});
		free(names[index]);
	}
	free(names);
	adamic_object *result = adamic_object_new(&listing_shape);
	result->slots[0].reference = &ok_kind;
	result->slots[1].reference = listed;
	return result;
}

adamic_object *adamic_file_status(const adamic_string *path) {
	static const char prefix[] = "cannot read status of ";
	// What was printed comes first, as on Node: the path may be stdout itself, its size what's
	// been written.
	adamic_output_flush();
#ifdef ADAMIC_TARGET_WASI
	// WASI libc treats the empty path as cwd; Node reports ENOENT.
	if (path->length == 0) {
		return failure(prefix, sizeof prefix - 1, path, ENOENT, false);
	}
#endif
	char *name = adamic_path_bytes(path);
	if (name == NULL) {
		return failure(prefix, sizeof prefix - 1, path, 0, false);
	}
	struct stat link, status;
	int failed = lstat(name, &link);
	bool symbolic = failed == 0 && S_ISLNK(link.st_mode);
	if (failed == 0) {
		failed = symbolic ? stat(name, &status) : 0;
		if (!symbolic) {
			status = link;
		}
	}
	int error = errno;
	free(name);
	if (failed != 0) {
		return failure(prefix, sizeof prefix - 1, path, error, false);
	}
	adamic_object *result = adamic_object_new(&status_shape);
	result->slots[0].reference = &ok_kind;
	result->slots[1].reference = S_ISREG(status.st_mode) ? &file_type : S_ISDIR(status.st_mode) ? &directory_type : &other_type;
	result->slots[2].number = (double)status.st_size;
	result->slots[3].boolean = symbolic;
	return result;
}
