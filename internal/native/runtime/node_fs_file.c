// Linux is the gate of record. The filesystem owns descriptor offsets and paths;
// the host owns only temporary buffers and returned counted values.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <errno.h>
#include <dirent.h>
#include <fcntl.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>

static const char *const error_fields[] = {"name", "message", "code"};
static const bool error_refs[] = {true, true, true};
static const adamic_shape error_shape = {3, error_fields, error_refs, NULL};

static adamic_string *text(const char *bytes) {
    return adamic_decode_utf8((const unsigned char *)bytes, strlen(bytes));
}

static void raise_error(const char *name, const char *code, const char *message) {
    adamic_object *error = adamic_object_new(&error_shape);
    error->slots[0].reference = text(name);
    error->slots[1].reference = text(message);
    error->slots[2].reference = text(code);
    error->class = strcmp(name, "TypeError") == 0 ? &adamic_host_type_error_class :
        strcmp(name, "RangeError") == 0 ? &adamic_host_range_error_class : &adamic_host_error_class;
    adamic_thrown = &error->heap;
    adamic_exception_pending = true;
}

static void system_error(int error, const char *operation, const char *path) {
    const char *code = "UNKNOWN", *reason = "unknown error";
    switch (error) {
#define FS_ERROR(e, r) case e: code = #e; reason = r; break
    FS_ERROR(ENOENT, "no such file or directory");
    FS_ERROR(ENOTDIR, "not a directory");
    FS_ERROR(EEXIST, "file already exists");
    FS_ERROR(EACCES, "permission denied");
    FS_ERROR(EPERM, "operation not permitted");
    FS_ERROR(EISDIR, "illegal operation on a directory");
    FS_ERROR(EBADF, "bad file descriptor");
    FS_ERROR(EINVAL, "invalid argument");
    FS_ERROR(ELOOP, "too many symbolic links encountered");
    FS_ERROR(ENAMETOOLONG, "name too long");
    FS_ERROR(ENOSPC, "no space left on device");
    FS_ERROR(EROFS, "read-only file system");
    FS_ERROR(EMFILE, "too many open files");
    FS_ERROR(ENFILE, "file table overflow");
    FS_ERROR(EIO, "i/o error");
    FS_ERROR(ENXIO, "no such device or address");
    FS_ERROR(EAGAIN, "resource temporarily unavailable");
    FS_ERROR(EINTR, "interrupted system call");
    FS_ERROR(EFBIG, "file too large");
    FS_ERROR(ENOMEM, "not enough memory");
    FS_ERROR(ENOTEMPTY, "directory not empty");
    FS_ERROR(ENOSYS, "function not implemented");
#undef FS_ERROR
    }
    size_t length = strlen(code) + strlen(reason) + strlen(operation) + (path == NULL ? 8 : strlen(path) + 12);
    char *message = malloc(length);
    if (message == NULL) { adamic_panic("out of memory", 13); }
    if (path == NULL) { snprintf(message, length, "%s: %s, %s", code, reason, operation); }
    else { snprintf(message, length, "%s: %s, %s '%s'", code, reason, operation, path); }
    raise_error("Error", code, message);
    free(message);
}

// Node's UTF-8 encoder replaces WTF-8 lone surrogates. A NUL in a path
// is rejected before a syscall. existsSync deliberately answers false instead.
static char *bytes(const adamic_string *value) {
    char *result = malloc(value->length + 1);
    if (result == NULL) { adamic_panic("out of memory", 13); }
    memcpy(result, value->bytes, value->length);
    for (size_t i = 0; i + 2 < value->length; i++) {
        if ((unsigned char)result[i] == 0xed && (unsigned char)result[i + 1] >= 0xa0) {
            memcpy(result + i, "\xef\xbf\xbd", 3);
            i += 2;
        }
    }
    result[value->length] = 0;
    return result;
}

// Node's ERR_INVALID_ARG_VALUE uses inspect followed by a 128-unit preview.
static void invalid_value(const char *prefix, const adamic_string *value) {
    size_t capacity = value->length * 4 + 8;
    char *shown = malloc(capacity);
    if (shown == NULL) { adamic_panic("out of memory", 13); }
    char quote = '\'';
    if (memchr(value->bytes, '\'', value->length) != NULL) {
        quote = memchr(value->bytes, '"', value->length) == NULL ? '"' :
            memchr(value->bytes, '`', value->length) == NULL && memchr(value->bytes, '$', value->length) == NULL ? '`' : '\'';
    }
    size_t at = 0; shown[at++] = quote;
    for (size_t i = 0; i < value->length; i++) {
        unsigned char c = (unsigned char)value->bytes[i];
        if (c == '\\' || c == (unsigned char)quote) { shown[at++] = '\\'; shown[at++] = (char)c; }
        else if (c == '\n' || c == '\r' || c == '\t' || c == '\b' || c == '\f' || c == '\v') {
            shown[at++] = '\\'; shown[at++] = c == '\n' ? 'n' : c == '\r' ? 'r' : c == '\t' ? 't' : c == '\b' ? 'b' : c == '\f' ? 'f' : 'v';
        } else if (c < 32 || c == 127) { snprintf(shown + at, 5, "\\x%02x", c); at += 4; }
        else { shown[at++] = (char)c; }
    }
    shown[at++] = quote; shown[at] = 0;
    adamic_string *inspection = adamic_decode_utf8((const unsigned char *)shown, at);
    free(shown);
    adamic_string *preview = inspection;
    if (adamic_string_length(inspection) > 128) {
        adamic_string *slice = adamic_string_slice(inspection, 0, 128, true);
        static adamic_string dots = ADAMIC_STRING("...");
        preview = adamic_string_concat(2, (adamic_string *const[]){slice, &dots});
        adamic_release(slice); adamic_release(inspection);
    }
    size_t length = strlen(prefix) + preview->length + 1;
    char *message = malloc(length);
    if (message == NULL) { adamic_panic("out of memory", 13); }
    memcpy(message, prefix, strlen(prefix));
    memcpy(message + strlen(prefix), preview->bytes, preview->length);
    message[length - 1] = 0;
    raise_error("TypeError", "ERR_INVALID_ARG_VALUE", message);
    free(message); adamic_release(preview);
}

static char *path_bytes(const adamic_string *path, bool throwing) {
    if (memchr(path->bytes, 0, path->length) == NULL) { return bytes(path); }
    if (throwing) { invalid_value("The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received ", path); }
    return NULL;
}

static bool integer(double value, double maximum, const char *argument) {
    if (isfinite(value) && value == trunc(value) && value >= 0 && value <= maximum) { return true; }
    adamic_string *shown = adamic_string_from_number(value);
    char message[300];
    if (!isfinite(value) || value != trunc(value)) {
        snprintf(message, sizeof message, "The value of \"%s\" is out of range. It must be an integer. Received %.*s", argument, (int)shown->length, shown->bytes);
    } else {
        snprintf(message, sizeof message, "The value of \"%s\" is out of range. It must be >= 0 && <= %.0f. Received %.*s", argument, maximum, (int)shown->length, shown->bytes);
    }
    adamic_release(shown);
    raise_error("RangeError", "ERR_OUT_OF_RANGE", message);
    return false;
}

static void read_range(const char *argument, const char *range, double value) {
    adamic_string *shown = adamic_string_from_number(value);
    char message[300];
    snprintf(message, sizeof message, "The value of \"%s\" is out of range. It must be %s. Received %.*s", argument, range, (int)shown->length, shown->bytes);
    adamic_release(shown);
    raise_error("RangeError", "ERR_OUT_OF_RANGE", message);
}
static bool file_descriptor(double value) {
    if (isfinite(value) && (value < 0 || value > INT_MAX)) {
        read_range("fd", ">= 0 && <= 2147483647", value); return false;
    }
    return integer(value, INT_MAX, "fd");
}

static int flags(const adamic_string *flag) {
    char *name = bytes(flag);
    int value = -1;
    if (strlen(name) != flag->length) { /* Embedded NUL is not a flag. */ }
    else if (strcmp(name, "r") == 0) { value = O_RDONLY; }
    else if (strcmp(name, "r+") == 0) { value = O_RDWR; }
    else if (strcmp(name, "rs") == 0 || strcmp(name, "sr") == 0) { value = O_RDONLY | O_SYNC; }
    else if (strcmp(name, "rs+") == 0 || strcmp(name, "sr+") == 0) { value = O_RDWR | O_SYNC; }
    else if (strcmp(name, "w") == 0) { value = O_WRONLY | O_CREAT | O_TRUNC; }
    else if (strcmp(name, "wx") == 0 || strcmp(name, "xw") == 0) { value = O_WRONLY | O_CREAT | O_TRUNC | O_EXCL; }
    else if (strcmp(name, "w+") == 0) { value = O_RDWR | O_CREAT | O_TRUNC; }
    else if (strcmp(name, "wx+") == 0 || strcmp(name, "xw+") == 0) { value = O_RDWR | O_CREAT | O_TRUNC | O_EXCL; }
    else if (strcmp(name, "a") == 0) { value = O_WRONLY | O_CREAT | O_APPEND; }
    else if (strcmp(name, "ax") == 0 || strcmp(name, "xa") == 0) { value = O_WRONLY | O_CREAT | O_APPEND | O_EXCL; }
    else if (strcmp(name, "a+") == 0) { value = O_RDWR | O_CREAT | O_APPEND; }
    else if (strcmp(name, "ax+") == 0 || strcmp(name, "xa+") == 0) { value = O_RDWR | O_CREAT | O_APPEND | O_EXCL; }
    else if (strcmp(name, "as") == 0 || strcmp(name, "sa") == 0) { value = O_WRONLY | O_CREAT | O_APPEND | O_SYNC; }
    else if (strcmp(name, "as+") == 0 || strcmp(name, "sa+") == 0) { value = O_RDWR | O_CREAT | O_APPEND | O_SYNC; }
    if (value < 0) { invalid_value("The argument 'flags' is invalid. Received ", flag); }
    free(name);
    return value;
}

static int open_file(const adamic_string *path, const adamic_string *flag, double mode) {
    char *name = path_bytes(path, true);
    if (name == NULL) { return -1; }
    int of = flags(flag);
    if (of < 0) { free(name); return -1; }
    if (!integer(mode, 4294967295., "mode")) { free(name); return -1; }
    int descriptor;
    do { descriptor = open(name, of | O_CLOEXEC, (mode_t)(uint32_t)mode); } while (descriptor < 0 && errno == EINTR);
    if (descriptor < 0) { system_error(errno, "open", name); }
    free(name);
    return descriptor;
}

double adamic_fs_file_open(const adamic_string *path, const adamic_string *flag, double mode) {
    adamic_output_flush();
    return (double)open_file(path, flag, mode);
}

int adamic_fs_file_read_bytes(int descriptor, unsigned char **out, size_t *length) {
    size_t used = 0, capacity = 4096;
    unsigned char *buffer = malloc(capacity);
    if (buffer == NULL) { return ENOMEM; }
    for (;;) {
        if (used == capacity) {
            if (capacity > SIZE_MAX / 2) { free(buffer); return ENOMEM; }
            capacity *= 2;
            unsigned char *grown = realloc(buffer, capacity);
            if (grown == NULL) { free(buffer); return ENOMEM; }
            buffer = grown;
        }
        ssize_t count = read(descriptor, buffer + used, capacity - used);
        if (count < 0 && errno == EINTR) { continue; }
        if (count < 0) { int saved = errno; free(buffer); return saved; }
        if (count == 0) { *out = buffer; *length = used; return 0; }
        used += (size_t)count;
    }
}

static adamic_string *read_file(int descriptor, bool owned) {
    unsigned char *buffer = NULL;
    size_t length = 0;
    int error = adamic_fs_file_read_bytes(descriptor, &buffer, &length);
    if (owned) { close(descriptor); }
    if (error != 0) { system_error(error, "read", NULL); return NULL; }
    if (length >= ADAMIC_STRING_MAX_UNITS) {
        free(buffer);
        char message[120];
        snprintf(message, sizeof message, "Cannot create a string longer than 0x%zx characters", (size_t)ADAMIC_STRING_MAX_UNITS);
        raise_error("Error", "ERR_STRING_TOO_LONG", message);
        return NULL;
    }
    adamic_string *result = adamic_decode_utf8(buffer, length);
    free(buffer);
    return result;
}

adamic_string *adamic_fs_file_read_file(const adamic_string *path, const adamic_string *flag) {
    adamic_output_flush();
    int descriptor = open_file(path, flag, 0666);
    return descriptor < 0 ? NULL : read_file(descriptor, true);
}

adamic_string *adamic_fs_file_read_fd(double descriptor, const adamic_string *flag) {
    adamic_output_flush();
    (void)flag; // Node ignores the flag for an already-open descriptor.
    if (!file_descriptor(descriptor)) { return NULL; }
    return read_file((int)descriptor, false);
}

static adamic_array *read_buffer(int descriptor, bool owned) {
    unsigned char *bytes = NULL;
    size_t length = 0;
    int error = adamic_fs_file_read_bytes(descriptor, &bytes, &length);
    if (owned) { close(descriptor); }
    if (error != 0) { system_error(error, "read", NULL); return NULL; }
    adamic_array *result = adamic_array_new(length, false);
    for (size_t i = 0; i < length; i++) {
        adamic_value byte = {0};
        byte.number = bytes[i];
        adamic_array_push(result, byte);
    }
    free(bytes);
    return result;
}
adamic_array *adamic_fs_file_read_buffer(const adamic_string *path, const adamic_string *flag) {
    adamic_output_flush();
    int descriptor = open_file(path, flag, 0666);
    return descriptor < 0 ? NULL : read_buffer(descriptor, true);
}
adamic_array *adamic_fs_file_read_buffer_fd(double descriptor, const adamic_string *flag) {
    adamic_output_flush();
    (void)flag;
    if (!file_descriptor(descriptor)) { return NULL; }
    return read_buffer((int)descriptor, false);
}

double adamic_fs_file_close(double descriptor) {
    adamic_output_flush();
    if (!file_descriptor(descriptor)) { return 0; }
    if (close((int)descriptor) != 0 && errno != EINTR) { system_error(errno, "close", NULL); }
    return 0;
}

double adamic_fs_file_read_sync(double descriptor, adamic_array *buffer, double offset, double length, double position) {
    adamic_output_flush();
    if (!integer(offset, 9007199254740991., "offset")) { return 0; }
    length = adamic_bitwise_or(length, 0);
    if (!isfinite(position) || position != trunc(position)) { read_range("position", "an integer", position); return 0; }
    if (position < -1 || position > 9007199254740991.) { read_range("position", ">= -1 && <= 9007199254740991", position); return 0; }
    if (length == 0) { return 0; }
    if (buffer->length == 0) {
        raise_error("TypeError", "ERR_INVALID_ARG_VALUE", "The argument 'buffer' is empty and cannot be written. Received <Buffer >"); return 0;
    }
    if (length < 0) { read_range("length", ">= 0", length); return 0; }
    if (offset + length > (double)buffer->length) {
        char range[80]; snprintf(range, sizeof range, "<= %.0f", (double)buffer->length - offset);
        read_range("length", range, length); return 0;
    }
    if (!file_descriptor(descriptor)) { return 0; }
    unsigned char *bytes = malloc((size_t)length);
    if (bytes == NULL) { adamic_panic("out of memory", 13); }
    ssize_t count;
    do { count = position < 0 ? read((int)descriptor, bytes, (size_t)length) : pread((int)descriptor, bytes, (size_t)length, (off_t)position); } while (count < 0 && errno == EINTR);
    if (count < 0) { free(bytes); system_error(errno, "read", NULL); return 0; }
    for (size_t i = 0; i < (size_t)count; i++) { buffer->elements[(size_t)offset + i].number = bytes[i]; }
    free(bytes);
    return (double)count;
}

double adamic_fs_file_write(double descriptor, const adamic_string *string, double position) {
    adamic_output_flush();
    if (!file_descriptor(descriptor)) { return 0; }
    char *buffer = bytes(string);
    ssize_t count;
    // libuv treats a negative offset as the descriptor's current position.
    // Node uses the current offset for fractional and non-finite positions.
    if (!isfinite(position) || position != trunc(position) || position > 9223372036854774784.) { position = -1; }
    do {
        count = position < 0 ? write((int)descriptor, buffer, string->length) : pwrite((int)descriptor, buffer, string->length, (off_t)position);
    } while (count < 0 && errno == EINTR);
    free(buffer);
    if (count < 0) { system_error(errno, "write", NULL); return 0; }
    return (double)count;
}

static double write_data(int descriptor, const char *buffer, size_t length, bool owned, bool flush) {
    size_t used = 0;
    int error = 0;
    while (used < length) {
        ssize_t count = write(descriptor, buffer + used, length - used);
        if (count < 0 && errno == EINTR) { continue; }
        if (count < 0) { error = errno; break; }
        if (count == 0) { error = EIO; break; }
        used += (size_t)count;
    }
    if (error != 0) { system_error(error, "write", NULL); }
    else if (flush && fsync(descriptor) < 0) { system_error(errno, "fsync", NULL); }
    // finally closes an internally opened fd, even after write/fsync failed.
    // A close failure overrides an earlier exception as Node's finally does.
    if (owned && close(descriptor) < 0 && errno != EINTR) {
        if (adamic_exception_pending) { adamic_release(adamic_thrown); adamic_thrown = NULL; adamic_exception_pending = false; }
        system_error(errno, "close", NULL);
    }
    return 0;
}

static double write_file(int descriptor, const adamic_string *string, bool owned, bool flush) {
    char *buffer = bytes(string);
    double result = write_data(descriptor, buffer, string->length, owned, flush);
    free(buffer);
    return result;
}
static double write_buffer(int descriptor, const adamic_array *buffer, bool owned, bool flush) {
    char *data = malloc(buffer->length == 0 ? 1 : buffer->length);
    if (data == NULL) { adamic_panic("out of memory", 13); }
    for (size_t i = 0; i < buffer->length; i++) { data[i] = (char)(unsigned char)buffer->elements[i].number; }
    double result = write_data(descriptor, data, buffer->length, owned, flush);
    free(data);
    return result;
}
double adamic_fs_file_write_buffer(const adamic_string *path, const adamic_array *buffer, const adamic_string *flag, double mode, bool flush) {
    adamic_output_flush();
    int descriptor = open_file(path, flag, mode);
    return descriptor < 0 ? 0 : write_buffer(descriptor, buffer, true, flush);
}
double adamic_fs_file_write_buffer_fd(double descriptor, const adamic_array *buffer, const adamic_string *flag, double mode, bool flush) {
    adamic_output_flush();
    (void)flag; (void)mode;
    if (!file_descriptor(descriptor)) { return 0; }
    return write_buffer((int)descriptor, buffer, false, flush);
}

double adamic_fs_file_write_file(const adamic_string *path, const adamic_string *string, const adamic_string *flag, double mode, bool flush) {
    adamic_output_flush();
    int descriptor = open_file(path, flag, mode);
    return descriptor < 0 ? 0 : write_file(descriptor, string, true, flush);
}

double adamic_fs_file_write_fd(double descriptor, const adamic_string *string, const adamic_string *flag, double mode, bool flush) {
    adamic_output_flush();
    (void)flag; (void)mode;
    if (!file_descriptor(descriptor)) { return 0; }
    return write_file((int)descriptor, string, false, flush);
}

bool adamic_fs_file_exists(const adamic_string *path) {
    adamic_output_flush();
    char *name = path_bytes(path, false);
    if (name == NULL) { return false; }
    struct stat information;
    bool exists = stat(name, &information) == 0;
    free(name);
    return exists;
}

static const char *const date_fields[] = {"_fsFileTime"};
static const bool date_refs[] = {false};
static const adamic_shape date_shape = {1, date_fields, date_refs, NULL};

adamic_object *adamic_fs_file_date_new(double milliseconds) {
    adamic_object *date = adamic_object_new(&date_shape);
    double clipped = isfinite(milliseconds) && fabs(milliseconds) <= 8640000000000000. ? trunc(milliseconds) : NAN;
    date->slots[0].number = clipped == 0 ? 0. : clipped;
    return date;
}

double adamic_fs_file_date_time(const adamic_object *date) {
    static adamic_slot_cache cache;
    return adamic_object_field(date, "_fsFileTime", &cache)->number;
}

static const char *const stat_fields[] = {"size", "mtimeMs", "mtime", "_fsFileMode", "atime"};
static const bool stat_refs[] = {false, false, true, false, true};
static const adamic_shape stat_shape = {5, stat_fields, stat_refs, NULL};

adamic_object *adamic_fs_file_stat(const adamic_string *path, bool throw_if_missing) {
    adamic_output_flush();
    char *name = path_bytes(path, true);
    if (name == NULL) { return NULL; }
    struct stat information;
    if (stat(name, &information) != 0) {
        int error = errno;
        if (throw_if_missing || (error != ENOENT && error != ENOTDIR)) { system_error(error, "stat", name); }
        free(name);
        return NULL;
    }
    free(name);
#ifdef __APPLE__
    double atime = (double)information.st_atimespec.tv_sec * 1000 + (double)information.st_atimespec.tv_nsec / 1000000;
    double mtime = (double)information.st_mtimespec.tv_sec * 1000 + (double)information.st_mtimespec.tv_nsec / 1000000;
#else
    double atime = (double)information.st_atim.tv_sec * 1000 + (double)information.st_atim.tv_nsec / 1000000;
    double mtime = (double)information.st_mtim.tv_sec * 1000 + (double)information.st_mtim.tv_nsec / 1000000;
#endif
    adamic_object *result = adamic_object_new(&stat_shape);
    result->slots[0].number = (double)information.st_size;
    result->slots[1].number = mtime;
    result->slots[2].reference = adamic_fs_file_date_new(floor(mtime + 0.5));
    result->slots[3].number = (double)information.st_mode;
    result->slots[4].reference = adamic_fs_file_date_new(floor(atime + 0.5));
    return result;
}

static mode_t stat_mode(const adamic_object *information) {
    static adamic_slot_cache cache;
    return (mode_t)adamic_object_field(information, "_fsFileMode", &cache)->number;
}
bool adamic_fs_file_is_file(const adamic_object *information) { return S_ISREG(stat_mode(information)); }
bool adamic_fs_file_is_directory(const adamic_object *information) { return S_ISDIR(stat_mode(information)); }
bool adamic_fs_file_is_symbolic_link(const adamic_object *information) { return S_ISLNK(stat_mode(information)); }

// Start at the full path, as libuv does. Only ENOENT permits making a parent.
// Record the first successful mkdir, retaining the spelling the caller supplied.
static int make_directory(char *name, mode_t mode, bool recursive, adamic_string **first) {
    if (mkdir(name, mode) == 0) {
        if (recursive && *first == NULL) { *first = text(name); }
        return 0;
    }
    int error = errno;
    if (!recursive) { return error; }
    if (error == EEXIST || error == EACCES || error == EROFS) {
        struct stat information;
        if (stat(name, &information) == 0 && S_ISDIR(information.st_mode)) { return 0; }
        return error;
    }
    if (error != ENOENT) { return error; }
    size_t length = strlen(name);
    while (length > 1 && name[length - 1] == '/') { length--; }
    size_t slash = length;
    while (slash > 0 && name[slash - 1] != '/') { slash--; }
    if (slash == 0) { return error; }
    size_t parent_length = slash - 1;
    while (parent_length > 1 && name[parent_length - 1] == '/') { parent_length--; }
    if (parent_length == 0) { parent_length = 1; }
    char saved = name[parent_length];
    name[parent_length] = 0;
    int parent_error = make_directory(name, mode, true, first);
    name[parent_length] = saved;
    if (parent_error != 0) { return parent_error; }
    if (mkdir(name, mode) == 0) { if (*first == NULL) { *first = text(name); } return 0; }
    error = errno;
    if (error == EEXIST) { struct stat information; if (stat(name, &information) == 0 && S_ISDIR(information.st_mode)) { return 0; } }
    return error;
}

adamic_string *adamic_fs_file_mkdir(const adamic_string *path, bool recursive, double mode) {
    adamic_output_flush();
    char *name = path_bytes(path, true);
    if (name == NULL) { return NULL; }
    if (!integer(mode, 4294967295., "options.mode")) { free(name); return NULL; }
    adamic_string *first = NULL;
    int error = make_directory(name, (mode_t)(uint32_t)mode, recursive, &first);
    if (error != 0) { adamic_release(first); first = NULL; system_error(error, "mkdir", name); }
    free(name);
    return first;
}

#ifdef ADAMIC_NODE_HOST
#ifdef ADAMIC_TARGET_WASI
#include <sys/random.h>
// wasi-libc leaves mkdtemp out, so do what musl's does: six random characters over
// the template's XXXXXX, then mkdir 0700, trying again only while the name is taken.
static char *mkdtemp(char *name) {
    static const char alphabet[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    char *suffix = name + strlen(name) - 6;
    for (int attempt = 0; attempt < 100; attempt++) {
        unsigned char random[6];
        if (getentropy(random, sizeof random) != 0) { return NULL; }
        for (int index = 0; index < 6; index++) { suffix[index] = alphabet[random[index] % (sizeof alphabet - 1)]; }
        if (mkdir(name, 0700) == 0) { return name; }
        if (errno != EEXIST) { return NULL; }
    }
    return NULL;
}
#endif

// libc mkdtemp atomically creates a private directory with a six-byte suffix.
adamic_string *adamic_fs_file_mkdtemp(const adamic_string *prefix) {
    adamic_output_flush();
    if (memchr(prefix->bytes, 0, prefix->length) != NULL) {
        invalid_value("The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received ", prefix);
        return NULL;
    }
    char *base = bytes(prefix);
    size_t length = strlen(base);
    char *name = malloc(length + 7);
    if (name == NULL) { free(base); adamic_panic("out of memory", 13); }
    memcpy(name, base, length); memcpy(name + length, "XXXXXX", 7); free(base);
    adamic_string *result = NULL;
    if (mkdtemp(name) != NULL) { result = text(name); }
    else { int error = errno; memcpy(name + length, "XXXXXX", 7); system_error(error, "mkdtemp", name); }
    free(name);
    return result;
}

// Node's native rmSync uses filesystem removal errors after its lstat check.
static void remove_error(int error, const char *operation, const char *name) {
    if (error == EACCES || error == EPERM) {
        const char *code = error == EACCES ? "EACCES" : "EPERM";
        const char *reason = error == EACCES ? "Permission denied" : "Operation not permitted";
        size_t length = strlen(name) * 2 + strlen(reason) + 32;
        char *message = malloc(length);
        if (message == NULL) { adamic_panic("out of memory", 13); }
        snprintf(message, length, "%s, %s: %s '%s'", code, reason, name, name);
        raise_error("Error", code, message); free(message);
    } else { system_error(error, operation, name); }
}

// lstat keeps symlinks as leaves: recursive removal never follows their targets.
static void remove_path(const char *name, bool recursive, bool force) {
    struct stat information;
    if (lstat(name, &information) != 0) {
        if (errno != ENOENT || !force) { system_error(errno, "lstat", name); }
        return;
    }
    if (!S_ISDIR(information.st_mode)) {
        if (unlink(name) != 0 && errno != ENOENT) { remove_error(errno, "unlink", name); }
        return;
    }
    if (!recursive) {
        size_t length = strlen(name) + 80;
        char *message = malloc(length);
        if (message == NULL) { adamic_panic("out of memory", 13); }
        snprintf(message, length, "Path is a directory: rm returned EISDIR (is a directory) %s", name);
        raise_error("SystemError", "ERR_FS_EISDIR", message); free(message); return;
    }
    if (rmdir(name) == 0) { return; }
    int error = errno;
    if (error != ENOTEMPTY && error != EEXIST && error != EPERM) {
        if (error != ENOENT) { remove_error(error, "rmdir", name); }
        return;
    }
    DIR *directory = opendir(name);
    if (directory == NULL) { if (errno != ENOENT) { remove_error(errno, "scandir", name); } return; }
    struct dirent *entry;
    while (!adamic_exception_pending) {
        errno = 0; entry = readdir(directory);
        if (entry == NULL) { if (errno != 0) { remove_error(errno, "scandir", name); } break; }
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) { continue; }
        size_t length = strlen(name) + strlen(entry->d_name) + 2;
        char *child = malloc(length);
        if (child == NULL) { closedir(directory); adamic_panic("out of memory", 13); }
        snprintf(child, length, "%s/%s", name, entry->d_name);
        remove_path(child, true, true); free(child);
    }
    closedir(directory);
    if (!adamic_exception_pending && rmdir(name) != 0 && errno != ENOENT) { remove_error(errno, "rmdir", name); }
}

double adamic_fs_file_rm(const adamic_string *path, bool recursive, bool force) {
    adamic_output_flush();
    char *name = path_bytes(path, true);
    if (name != NULL) { remove_path(name, recursive, force); free(name); }
    return 0;
}

#endif

double adamic_fs_file_unlink(const adamic_string *path) {
    adamic_output_flush();
    char *name = path_bytes(path, true);
    if (name == NULL) { return 0; }
    if (unlink(name) != 0) { system_error(errno, "unlink", name); }
    free(name);
    return 0;
}

static double fs_utimes(const adamic_string *path, double atime, double mtime, bool atime_date, bool mtime_date) {
    adamic_output_flush();
    char *name = path_bytes(path, true);
    if (name == NULL) { return 0; }
    double seconds[] = {atime, mtime};
    bool dates[] = {atime_date, mtime_date};
    struct timespec times[2];
    for (size_t i = 0; i < 2; i++) {
        double value = seconds[i];
        if (dates[i] && isnan(value)) { times[i].tv_sec=0; times[i].tv_nsec=UTIME_OMIT; continue; }
        if (!isfinite(value)) {
            adamic_string *shown = adamic_string_from_number(value);
            char message[250];
            snprintf(message, sizeof message, "The \"time\" argument must be an instance of Date or an Time in seconds. Received type number (%.*s)", (int)shown->length, shown->bytes);
            adamic_release(shown); free(name);
            raise_error("TypeError", "ERR_INVALID_ARG_TYPE", message);
            return 0;
        }
        if (value < 0 && !dates[i]) { clock_gettime(CLOCK_REALTIME, &times[i]); continue; }
        // libuv converts double seconds to timespec nanoseconds on Linux.
        if (value >= 9223372036854774784.) { times[i].tv_sec = (time_t)INT64_MAX; times[i].tv_nsec = 0; }
        else { times[i].tv_sec = (time_t)floor(value); times[i].tv_nsec = (long)((value - (double)times[i].tv_sec) * 1000000000); }
    }
    if (utimensat(AT_FDCWD, name, times, 0) != 0) { system_error(errno, "utime", name); }
    free(name);
    return 0;
}

double adamic_fs_file_utimes(const adamic_string *path,double atime,double mtime) {return fs_utimes(path,atime,mtime,false,false);}
double adamic_fs_file_utimes_dates(const adamic_string *path,const adamic_object *atime,const adamic_object *mtime) {return fs_utimes(path,adamic_fs_file_date_time(atime)/1000,adamic_fs_file_date_time(mtime)/1000,true,true);}
double adamic_fs_file_utimes_atime_date(const adamic_string *path,const adamic_object *atime,double mtime) {return fs_utimes(path,adamic_fs_file_date_time(atime)/1000,mtime,true,false);}
double adamic_fs_file_utimes_mtime_date(const adamic_string *path,double atime,const adamic_object *mtime) {return fs_utimes(path,atime,adamic_fs_file_date_time(mtime)/1000,false,true);}
