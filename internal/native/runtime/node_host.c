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
static adamic_string *current_directory;
static bool cleanup_registered;
static void finish_directory(void) { adamic_release(current_directory); current_directory = NULL; }

static void host_directory_error(int error, const char *operation, const adamic_string *from, const adamic_string *to) {
    const char *code, *description;
    switch (error) {
        case ENOENT: code = "ENOENT"; description = "no such file or directory"; break;
        case ENOTDIR: code = "ENOTDIR"; description = "not a directory"; break;
        case EACCES: code = "EACCES"; description = "permission denied"; break;
        case ELOOP: code = "ELOOP"; description = "too many symbolic links encountered"; break;
        case ENAMETOOLONG: code = "ENAMETOOLONG"; description = "name too long"; break;
        case ENOMEM: code = "ENOMEM"; description = "not enough memory"; break;
        case EIO: code = "EIO"; description = "i/o error"; break;
        default: { static const char message[] = "unsupported host directory errno"; adamic_panic(message, sizeof message - 1); }
    }
    if (error == ENOENT && strcmp(operation, "uv_cwd") == 0) {
        description = "process.cwd failed with error no such file or directory, the current working directory was likely removed without changing the working directory";
    }
    static const char *const directory_error_names[] = {"name", "message", "code"};
    static const bool references[] = {true, true, true};
    static const adamic_shape shape = {3, directory_error_names, references, NULL, NULL};
    static adamic_string error_name = ADAMIC_STRING("Error");
    adamic_string *prefix = adamic_decode_utf8((const unsigned char *)code, strlen(code));
    adamic_string *reason = adamic_decode_utf8((const unsigned char *)description, strlen(description));
    adamic_string *syscall = adamic_decode_utf8((const unsigned char *)operation, strlen(operation));
    adamic_string colon = ADAMIC_STRING(": "), comma = ADAMIC_STRING(", ");
    adamic_string quote = ADAMIC_STRING("'"), arrow = ADAMIC_STRING("' -> '");
    adamic_string *message;
    if (to == NULL) {
        message = adamic_string_concat(5, (adamic_string *const[]){prefix, &colon, reason, &comma, syscall});
    } else {
        adamic_string space = ADAMIC_STRING(" ");
        message = adamic_string_concat(10, (adamic_string *const[]){prefix, &colon, reason, &comma, syscall, &space, &quote, (adamic_string *)from, &arrow, (adamic_string *)to});
        adamic_string *closed = adamic_string_concat(2, (adamic_string *const[]){message, &quote});
        adamic_release(message);
        message = closed;
    }
    adamic_thrown = adamic_object_new(&shape);
    adamic_thrown->slots[0].reference = &error_name;
    adamic_thrown->slots[1].reference = message;
    adamic_thrown->slots[2].reference = prefix;
    adamic_release(reason);
    adamic_release(syscall);
}

adamic_string *adamic_fs_file_host_cwd(void) {
    if (!cleanup_registered) { atexit(finish_directory); cleanup_registered = true; }
    if (current_directory != NULL) { return adamic_retain(current_directory); }
    size_t capacity = 256;
    for (;;) {
        char *buffer = malloc(capacity);
        if (buffer == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
        if (getcwd(buffer, capacity) != NULL) {
            adamic_string *result = adamic_decode_utf8((const unsigned char *)buffer, strlen(buffer));
            free(buffer);
            current_directory = result;
            return adamic_retain(result);
        }
        int error = errno;
        free(buffer);
        if (error == ERANGE) { capacity *= 2; continue; }
        host_directory_error(error, "uv_cwd", NULL, NULL);
        return NULL;
    }
}

static char *environment_text(const adamic_string *text) {
    char *bytes = malloc(text->length + 1);
    if (bytes == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
    memcpy(bytes, text->bytes, text->length);
    for (size_t at = 0; at + 3 <= text->length; at++) {
        if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
            memcpy(bytes + at, "\xef\xbf\xbd", 3);
            at += 2;
        }
    }
    bytes[text->length] = 0;
    return bytes;
}

double adamic_fs_file_host_chdir(const adamic_string *directory) {
    char *path = environment_text(directory);
    if (chdir(path) == 0) {
        free(path);
        if (current_directory != NULL) { adamic_release(current_directory); current_directory = NULL; }
        return 0;
    }
    int error = errno;
    adamic_string *target = adamic_decode_utf8((const unsigned char *)path, strlen(path));
    free(path);
    adamic_string *from = adamic_fs_file_host_cwd();
    if (adamic_thrown != NULL) { adamic_release(target); return 0; }
    host_directory_error(error, "chdir", from, target);
    adamic_release(target);
    adamic_release(from);
    return 0;
}

adamic_string *adamic_fs_file_host_tmpdir(void) {
    const char *value = getenv("TMPDIR");
    if (value == NULL || value[0] == 0) { value = getenv("TMP"); }
    if (value == NULL || value[0] == 0) { value = getenv("TEMP"); }
    if (value == NULL || value[0] == 0) { value = "/tmp"; }
    size_t length = strlen(value);
    if (length > 1 && value[length - 1] == '/') { length--; }
    return adamic_decode_utf8((const unsigned char *)value, length);
}
