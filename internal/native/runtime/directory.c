// directory.c: the file system a walk needs: readDirectory, fileStatus and realPath.
//
// Each answers as Node does on the same machine. readdirSync's names come from libuv's scandir,
// sorted by strcmp on the raw bytes, without . and .., and are decoded as UTF-8 after the sort, so
// the order is the bytes' even where a name isn't valid UTF-8. statSync follows a symbolic link and
// lstatSync doesn't; fileStatus gives the type and size of what a path names, following a link, and
// says whether the path itself is one.

// lstat, S_ISLNK and the directory calls are POSIX's, which strict C11 doesn't show without asking.
#define _XOPEN_SOURCE 700
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"

#include <dirent.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static const char *const ok_listing_names[] = {"kind", "names"};
static const char *const ok_status_names[] = {"kind", "type", "size", "symbolicLink"};
static const char *const error_names[] = {"kind", "message"};
static const bool two_references[] = {true, true};
static const bool status_references[] = {true, true, false, false};
static const adamic_shape listing_shape = {2, ok_listing_names, two_references, NULL};
static const adamic_shape status_shape = {4, ok_status_names, status_references, NULL};
static const adamic_shape error_shape = {2, error_names, two_references, NULL};
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

// fs.realpathSync resolves the input lexically before following links. In particular,
// an empty path names cwd, and dangling/.. does not inspect the dangling component.
static char *real_path_input(char *name) {
	if (name[0] != '/') {
		char *cwd = getcwd(NULL, 0);
		if (cwd == NULL) { int error = errno; free(name); errno = error; return NULL; }
		size_t length = strlen(cwd), rest = strlen(name);
		char *absolute = malloc(length + rest + 2);
		if (absolute == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		memcpy(absolute, cwd, length);
		absolute[length] = '/';
		memcpy(absolute + length + 1, name, rest + 1);
		free(cwd);
		free(name);
		name = absolute;
	}
	size_t output = 1;
	char *cursor = name;
	while (*cursor != '\0') {
		while (*cursor == '/') { cursor++; }
		char *segment = cursor;
		while (*cursor != '\0' && *cursor != '/') { cursor++; }
		size_t length = (size_t)(cursor - segment);
		if (length == 0 || (length == 1 && segment[0] == '.')) { continue; }
		if (length == 2 && segment[0] == '.' && segment[1] == '.') {
			while (output > 1 && name[output - 1] != '/') { output--; }
			if (output > 1) { output--; }
			continue;
		}
		if (output > 1) { name[output++] = '/'; }
		memmove(name + output, segment, length);
		output += length;
	}
	name[0] = '/';
	name[output] = '\0';
	return name;
}

// Ported from Node v24.14.1 lib/fs.js realpathSync, for POSIX paths.
// The root and ordinary components keep their input spelling. Links alone replace it.
static adamic_object *real_path(const adamic_string *path, bool node) {
	static const char prefix[] = "cannot resolve path ";
	static const char *const names[] = {"kind", "path"};
	static const adamic_shape shape = {2, names, two_references, NULL};
	adamic_output_flush();
	char *name = adamic_path_bytes(path);
	if (name == NULL) { return failure(prefix, sizeof prefix - 1, path, 0, false); }
	name = real_path_input(name);
	if (name == NULL) {
		if (node) { adamic_node_fs_raise(NULL, errno, "uv_cwd"); return NULL; }
		return failure(prefix, sizeof prefix - 1, path, errno, false);
	}
	size_t position = 1;
	int error = 0;
	const char *operation = "lstat";
	while (name[position] != '\0') {
		size_t previous = position;
		while (name[position] != '\0' && name[position] != '/') { position++; }
		char separator = name[position];
		name[position] = '\0';
		struct stat status;
		if (lstat(name, &status) != 0) { error = errno; break; }
		if (!S_ISLNK(status.st_mode)) {
			name[position] = separator;
			if (separator != '\0') { position++; }
			continue;
		}
		// Node stats the target first: the OS supplies ENOENT or ELOOP.
		if (stat(name, &status) != 0) { error = errno; operation = "stat"; break; }
		size_t capacity = 256;
		char *target = NULL;
		ssize_t length;
		for (;;) {
			char *grown = realloc(target, capacity + 1);
			if (grown == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			target = grown;
			length = readlink(name, target, capacity);
			if (length < 0 || (size_t)length < capacity) { break; }
			capacity *= 2;
		}
		if (length < 0) { error = errno; operation = "readlink"; free(target); break; }
		target[length] = '\0';
		name[position] = separator;
		const char *rest = name + position + (separator != '\0');
		size_t base_length = target[0] == '/' ? 0 : previous;
		size_t target_length = (size_t)length, rest_length = strlen(rest);
		char *replacement = malloc(base_length + target_length + rest_length + 2);
		if (replacement == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		memcpy(replacement, name, base_length);
		memcpy(replacement + base_length, target, target_length + 1);
		free(target);
		// Resolve the target against the link's directory, then resolve the rest.
		replacement = real_path_input(replacement);
		size_t resolved_length = strlen(replacement);
		replacement[resolved_length] = '/';
		memcpy(replacement + resolved_length + 1, rest, rest_length + 1);
		free(name);
		name = real_path_input(replacement);
		position = 1;
	}
	if (error != 0) {
		if (node) {
			adamic_string *failed = adamic_decode_utf8((const unsigned char *)name, strlen(name));
			free(name);
			adamic_node_fs_raise(failed, error, operation);
			adamic_release(failed);
			return NULL;
		}
		free(name);
		return failure(prefix, sizeof prefix - 1, path, error, false);
	}
	char *resolved = name;
	adamic_object *result = adamic_object_new(&shape);
	result->slots[0].reference = &ok_kind;
	result->slots[1].reference = adamic_decode_utf8((const unsigned char *)resolved, strlen(resolved));
	free(resolved);
	return result;
}

adamic_object *adamic_real_path(const adamic_string *path) { return real_path(path, false); }
adamic_object *adamic_real_path_node(const adamic_string *path) { return real_path(path, true); }
