// node_fs_directory.c: synchronous POSIX host operations, with catchable Node
// errors.
#define _DEFAULT_SOURCE
#define _XOPEN_SOURCE 700
#define _POSIX_C_SOURCE 200809L
// macOS hides d_type's DT_* constants once _POSIX_C_SOURCE is set, unless Darwin's
// own extensions are asked for too. glibc ignores this macro.
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <dirent.h>
#include <fcntl.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>
#include <time.h>
#include "node_fs_wasi.h"

static void *allocate(size_t size) {
	void *result = malloc(size);
	if (result == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return result;
}
static adamic_string *decode(const char *text) {
	return adamic_decode_utf8((const unsigned char *)text, strlen(text));
}
void adamic_node_fs_raise(const adamic_string *path, int error,
						  const char *operation) {
	const char *code, *reason;
	switch (error) {
	case ENOENT:
		code = "ENOENT";
		reason = "no such file or directory";
		break;
	case EEXIST:
		code = "EEXIST";
		reason = "file already exists";
		break;
	case ENOTDIR:
		code = "ENOTDIR";
		reason = "not a directory";
		break;
	case EACCES:
		code = "EACCES";
		reason = "permission denied";
		break;
	case EPERM:
		code = "EPERM";
		reason = "operation not permitted";
		break;
	case ELOOP:
		code = "ELOOP";
		reason = "too many symbolic links encountered";
		break;
	case EINVAL:
		code = "EINVAL";
		reason = "invalid argument";
		break;
	case ENAMETOOLONG:
		code = "ENAMETOOLONG";
		reason = "name too long";
		break;
	case EMFILE:
		code = "EMFILE";
		reason = "too many open files";
		break;
	case ENFILE:
		code = "ENFILE";
		reason = "file table overflow";
		break;
	case ENOMEM:
		code = "ENOMEM";
		reason = "not enough memory";
		break;
	case EOVERFLOW:
		code = "EOVERFLOW";
		reason = "value too large for defined data type";
		break;
	case EBADF:
		code = "EBADF";
		reason = "bad file descriptor";
		break;
	case EFAULT:
		code = "EFAULT";
		reason = "bad address in system call argument";
		break;
	case EINTR:
		code = "EINTR";
		reason = "interrupted system call";
		break;
	case ENOSYS:
		code = "ENOSYS";
		reason = "function not implemented";
		break;
	case EIO:
		code = "EIO";
		reason = "i/o error";
		break;
	default:
		code = "UNKNOWN";
		reason = "unknown error";
		break;
	}
	adamic_string *before = decode(code), *detail = decode(reason),
				  *syscall = decode(operation);
	static adamic_string colon = ADAMIC_STRING(": "),
						 comma = ADAMIC_STRING(", "),
						 quote = ADAMIC_STRING(" '"), end = ADAMIC_STRING("'");
	adamic_string *message = adamic_string_concat(
		path == NULL ? 5 : 8,
		(adamic_string *const[]){before, &colon, detail, &comma, syscall,
								 &quote, (adamic_string *)path, &end});
	adamic_object *thrown = adamic_builtin_error_new(0, message);
	adamic_release(message);
	thrown->slots[2].reference = before;
	adamic_release(detail);
	adamic_release(syscall);
	adamic_error_tag(thrown);
	adamic_thrown = thrown;
}
// NUL is rejected before any filesystem effect. The supported path contract is
// a string.
static bool valid(const adamic_string *path, bool target) {
	if (memchr(path->bytes, 0, path->length) == NULL) {
		return true;
	}
	static adamic_string code = ADAMIC_STRING("ERR_INVALID_ARG_VALUE");
	static adamic_string target_before =
		ADAMIC_STRING("The argument 'target' must be a string, Uint8Array, or "
					  "URL without null bytes. Received ");
	static adamic_string before =
		ADAMIC_STRING("The argument 'path' must be a string, Uint8Array, or "
					  "URL without null bytes. Received ");
	char quote = '\'';
	if (memchr(path->bytes, '\'', path->length) != NULL) {
		if (memchr(path->bytes, '"', path->length) == NULL) {
			quote = '"';
		} else if (memchr(path->bytes, '`', path->length) == NULL) {
			quote = '`';
			for (size_t i = 1; i < path->length; i++) {
				if (path->bytes[i - 1] == '$' && path->bytes[i] == '{') {
					quote = '\'';
					break;
				}
			}
		}
	}
	char *bytes = allocate(path->length * 6 + 3);
	size_t used = 0;
	bytes[used++] = quote;
	for (size_t i = 0; i < path->length; i++) {
		unsigned char ch = (unsigned char)path->bytes[i];
		const char *escape = ch == '\n'	  ? "\\n"
							 : ch == '\r' ? "\\r"
							 : ch == '\t' ? "\\t"
							 : ch == '\b' ? "\\b"
							 : ch == '\f' ? "\\f"
										  : NULL;
		if (escape != NULL) {
			memcpy(bytes + used, escape, 2);
			used += 2;
		} else if (ch < 32 || ch == 127) {
			static const char hex[] = "0123456789ABCDEF";
			bytes[used++] = '\\';
			bytes[used++] = 'x';
			bytes[used++] = hex[ch >> 4];
			bytes[used++] = hex[ch & 15];
		} else if (ch == 0xed && i + 2 < path->length &&
				   (unsigned char)path->bytes[i + 1] >= 0xa0 &&
				   (unsigned char)path->bytes[i + 1] <= 0xbf) {
			unsigned point = ((ch & 15) << 12) |
							 (((unsigned char)path->bytes[i + 1] & 63) << 6) |
							 ((unsigned char)path->bytes[i + 2] & 63);
			static const char hex[] = "0123456789abcdef";
			bytes[used++] = '\\';
			bytes[used++] = 'u';
			for (int shift = 12; shift >= 0; shift -= 4) {
				bytes[used++] = hex[(point >> (unsigned)shift) & 15];
			}
			i += 2;
		} else {
			if (ch == '\\' || ch == (unsigned char)quote) {
				bytes[used++] = '\\';
			}
			bytes[used++] = (char)ch;
		}
	}
	bytes[used++] = quote;
	adamic_string *shown = adamic_decode_utf8((unsigned char *)bytes, used);
	free(bytes);
	if (adamic_string_length(shown) > 128) {
		adamic_string *cut = adamic_string_slice(shown, 0, 128, true);
		adamic_release(shown);
		static adamic_string dots = ADAMIC_STRING("...");
		shown = adamic_string_concat(2, (adamic_string *const[]){cut, &dots});
		adamic_release(cut);
	}
	adamic_string *message = adamic_string_concat(2, (adamic_string *const[]){target ? &target_before : &before, shown});
	adamic_object *thrown = adamic_builtin_error_new(1, message);
	adamic_release(message);
	thrown->slots[2].reference = &code;
	adamic_release(shown);
	adamic_error_tag(thrown);
	adamic_thrown = thrown;
	return false;
}
static const char *const dirent_names[] = {"name", "type"};
static const bool dirent_refs[] = {true, false};
static const adamic_shape dirent_shape = {2, dirent_names, dirent_refs, NULL};
typedef struct {
	char *name;
	unsigned type;
} entry;
static int compare(const void *left, const void *right) {
	return strcmp(((const entry *)left)->name, ((const entry *)right)->name);
}
static unsigned kind(mode_t mode) {
	return S_ISREG(mode) ? 1 : S_ISDIR(mode) ? 2 : S_ISLNK(mode) ? 3
		: S_ISBLK(mode) ? 5 : S_ISCHR(mode) ? 6 : S_ISFIFO(mode) ? 7
		: S_ISSOCK(mode) ? 8 : 0;
}
bool adamic_node_fs_dirent_is(const adamic_object *entry_value,
							  const char *method) {
	// Shape identity is the host value's runtime kind. Union declarations
	// can begin with either StatsBase or Dirent.
	if (adamic_fs_file_is_stats(entry_value)) {
        return adamic_fs_file_stat_is(entry_value, method);
    }
    unsigned wanted = strcmp(method, "isFile") == 0 ? 1
        : strcmp(method, "isDirectory") == 0 ? 2
        : strcmp(method, "isSymbolicLink") == 0 ? 3
        : strcmp(method, "isBlockDevice") == 0 ? 5
        : strcmp(method, "isCharacterDevice") == 0 ? 6
        : strcmp(method, "isFIFO") == 0 ? 7 : 8;
	static adamic_slot_cache cache;
	double actual = adamic_object_field(entry_value, "type", &cache)->number;
#ifdef ADAMIC_TARGET_WASI
    // This type also represents FIFOs in Node's WASI host.
    if (actual == 8 && (wanted == 7 || wanted == 8)) {
        adamic_panic("wasm32-wasi: fs.isFIFO/isSocket cannot distinguish FIFOs from sockets", sizeof "wasm32-wasi: fs.isFIFO/isSocket cannot distinguish FIFOs from sockets" - 1);
    }
#endif
    return actual == (double)wanted;
}
adamic_array *adamic_node_fs_readdir(const adamic_string *path,
									 const adamic_object *options) {
	adamic_output_flush();
	if (!valid(path, false)) {
		return NULL;
	}
	bool typed = false;
	if (options != NULL &&
		adamic_object_has(options,
						  &(adamic_string)ADAMIC_STRING("withFileTypes"))) {
		static adamic_slot_cache cache;
		typed = adamic_object_field(options, "withFileTypes", &cache)->boolean;
	}
	#ifdef ADAMIC_TARGET_WASI
    // wasi-libc treats an empty directory path as cwd; Node rejects it.
    if (path->length == 0) { adamic_node_fs_raise(path, ENOENT, "scandir"); return NULL; }
#endif
	char *name = adamic_path_bytes(path);
	DIR *directory = opendir(name);
	if (directory == NULL) {
		int error = errno;
		free(name);
		adamic_node_fs_raise(path, error, "scandir");
		return NULL;
	}
	entry *entries = NULL;
	size_t count = 0, capacity = 0;
	int error = 0;
	for (;;) {
		errno = 0;
		struct dirent *found = readdir(directory);
		if (found == NULL) {
			error = errno;
			break;
		}
		if (strcmp(found->d_name, ".") == 0 ||
			strcmp(found->d_name, "..") == 0) {
			continue;
		}
		if (count == capacity) {
			capacity = capacity == 0 ? 16 : capacity * 2;
			entry *grown = realloc(entries, capacity * sizeof *entries);
			if (grown == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			entries = grown;
		}
		entries[count].name = strdup(found->d_name);
		if (entries[count].name == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		// lstat preserves the entry's identity; symlink targets are followed
		// only by sys.ts.
		entries[count].type = 0;
#if defined(_DIRENT_HAVE_D_TYPE) || defined(__APPLE__)
		entries[count].type = found->d_type == DT_REG		? 1
							  : found->d_type == DT_DIR		? 2
							  : found->d_type == DT_LNK		? 3
							  : found->d_type == DT_BLK ? 5
                              : found->d_type == DT_CHR ? 6
                              : found->d_type == DT_FIFO ? 7
                              : found->d_type == DT_SOCK ? 8
                              : found->d_type == DT_UNKNOWN ? 4
															: 0;
#else
		entries[count].type = 4;
#endif
		count++;
	}
	closedir(directory);
	if (count > 1) {
		qsort(entries, count, sizeof *entries, compare);
	}
	adamic_array *result = adamic_array_new(count, true);
	for (size_t i = 0; i < count; i++) {
		if (error == 0 && typed && entries[i].type == 4) {
			size_t base = strlen(name), length = strlen(entries[i].name);
			char *full = allocate(base + length + 2);
			memcpy(full, name, base);
			full[base] = '/';
			memcpy(full + base + 1, entries[i].name, length + 1);
			struct stat status;
			if (lstat(full, &status) != 0) {
				error = errno;
				adamic_string *failed = decode(full);
				adamic_node_fs_raise(failed, error, "lstat");
				adamic_release(failed);
			} else {
				entries[i].type = kind(status.st_mode);
			}
			free(full);
		}
		if (error == 0) {
			adamic_string *decoded = decode(entries[i].name);
			if (typed) {
				adamic_object *value = adamic_object_new(&dirent_shape);
				value->slots[0].reference = decoded;
				value->slots[1].number = (double)entries[i].type;
				adamic_array_push(result, (adamic_value){.reference = value});
			} else {
				adamic_array_push(result, (adamic_value){.reference = decoded});
			}
		}
		free(entries[i].name);
	}
	free(entries);
	free(name);
	if (error != 0) {
		adamic_release(result);
		if (adamic_thrown == NULL) {
			adamic_node_fs_raise(path, error, "scandir");
		}
		return NULL;
	}
	return result;
}
adamic_string *adamic_node_fs_realpath(const adamic_string *path, bool native) {
	adamic_output_flush();
	if (!valid(path, false)) {
		return NULL;
	}
	if (native) {
		char *name = adamic_path_bytes(path);
		char *resolved = realpath(name, NULL);
		int error = errno;
		free(name);
		if (resolved == NULL) {
			adamic_node_fs_raise(path, error, "realpath");
			return NULL;
		}
		adamic_string *result = decode(resolved);
		free(resolved);
		return result;
	}
	adamic_object *value = adamic_real_path_node(path);
	if (value == NULL) {
		return NULL;
	}
	adamic_string *result = adamic_retain(value->slots[1].reference);
	adamic_release(value);
	return result;
}

// Node reports the target before the link path and validates both before effects.
double adamic_node_fs_symlink(const adamic_string *target,
                              const adamic_string *path) {
    adamic_output_flush();
    if (!valid(target, true) || !valid(path, false)) return 0;
    char *destination = adamic_path_bytes(target);
    char *name = adamic_path_bytes(path);
    int result = symlink(destination, name);
    int error = errno;
    free(name);
    free(destination);
    if (result != 0) {
        adamic_node_fs_raise(target, error, "symlink");
        static adamic_string arrow = ADAMIC_STRING(" -> '"), quote = ADAMIC_STRING("'");
        adamic_string *before = adamic_thrown->slots[1].reference;
        adamic_thrown->slots[1].reference = adamic_string_concat(4,
            (adamic_string *const[]){before, &arrow, (adamic_string *)path, &quote});
        adamic_release(before);
    }
    return 0;
}
